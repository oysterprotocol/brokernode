package jobs_test

import (
	"github.com/gobuffalo/pop/nulls"
	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
)

func (suite *JobsSuite) Test_PurgeCompletedSessions() {

	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	defer oyster_utils.ResetBrokerMode()

	fileBytesCount := uint64(2500)
	numChunks := 3
	privateKey := "1111111111111111111111111111111111111111111111111111111111111111"

	uploadSession1 := models.UploadSession{
		GenesisHash:   "abcdeff1",
		FileSizeBytes: fileBytesCount,
		NumChunks:     numChunks,
		Type:          models.SessionTypeBeta,
		ETHAddrAlpha:  nulls.String{string("SOME_ALPHA_ETH_ADDRESS1"), true},
		ETHAddrBeta:   nulls.String{string("SOME_BETA_ETH_ADDRESS1"), true},
		ETHPrivateKey: privateKey,
	}

	chunksReady, vErr, err := uploadSession1.StartSessionAndWaitForChunks(500)
	suite.True(chunksReady)

	suite.False(vErr.HasAny())
	suite.Nil(err)

	uploadSession2 := models.UploadSession{
		GenesisHash:   "abcdeff2",
		FileSizeBytes: fileBytesCount,
		NumChunks:     numChunks,
		Type:          models.SessionTypeAlpha,
		ETHAddrAlpha:  nulls.String{string("SOME_ALPHA_ETH_ADDRESS2"), true},
		ETHAddrBeta:   nulls.String{string("SOME_BETA_ETH_ADDRESS2"), true},
		ETHPrivateKey: privateKey,
	}

	chunksReady, vErr, err = uploadSession2.StartSessionAndWaitForChunks(500)
	suite.True(chunksReady)
	suite.False(vErr.HasAny())
	suite.Nil(err)

	uploadSession3 := models.UploadSession{
		GenesisHash:   "abcdeff3",
		FileSizeBytes: fileBytesCount,
		NumChunks:     numChunks,
		Type:          models.SessionTypeAlpha,
		ETHAddrAlpha:  nulls.String{string("SOME_ALPHA_ETH_ADDRESS3"), true},
		ETHAddrBeta:   nulls.String{string("SOME_BETA_ETH_ADDRESS3"), true},
		ETHPrivateKey: privateKey,
	}

	chunksReady, vErr, err = uploadSession3.StartSessionAndWaitForChunks(500)
	suite.True(chunksReady)
	suite.False(vErr.HasAny())
	suite.Nil(err)

	allDataMaps := []models.DataMap{}
	err = suite.DB.All(&allDataMaps)
	suite.Nil(err)

	completedDataMaps := []models.CompletedDataMap{}
	err = suite.DB.All(&completedDataMaps)
	suite.Nil(err)

	uploadSessions := []models.UploadSession{}
	err = suite.DB.All(&uploadSessions)
	suite.Nil(err)

	completedUploads := []models.CompletedUpload{}
	err = suite.DB.All(&completedUploads)
	suite.Nil(err)

	// verify initial lengths are what we expected
	suite.Equal(3*(numChunks+1), len(allDataMaps)) // 3 data maps so 3 extra chunks have been added
	suite.Equal(0, len(completedDataMaps))
	suite.Equal(3, len(uploadSessions))
	suite.Equal(0, len(completedUploads))

	// set all chunks of first data map to complete or confirmed
	allDone := []models.DataMap{}
	err = suite.DB.Where("genesis_hash = ?", "abcdeff1").All(&allDone)
	suite.Nil(err)

	for _, dataMap := range allDone {
		dataMap.Status = models.Complete
		suite.DB.ValidateAndSave(&dataMap)
	}

	// set one of them to "confirmed"
	allDone[1].Status = models.Confirmed
	suite.DB.ValidateAndSave(&allDone[1])

	// set one chunk of second data map to complete
	someDone := []models.DataMap{}
	err = suite.DB.Where("genesis_hash = ?", "abcdeff2").All(&someDone)
	suite.Nil(err)

	someDone[0].Status = models.Complete
	suite.DB.ValidateAndSave(&someDone[0])

	//call method under test
	jobs.PurgeCompletedSessions(jobs.PrometheusWrapper)

	allDataMaps = []models.DataMap{}
	err = suite.DB.All(&allDataMaps)
	suite.Nil(err)

	completedDataMaps = []models.CompletedDataMap{}
	err = suite.DB.All(&completedDataMaps)
	suite.Nil(err)

	if services.IsKvStoreEnabled() {
		var keys services.KVKeys
		for _, cDataMap := range completedDataMaps {
			keys = append(keys, cDataMap.MsgID)
		}
		kvPairs, err := services.BatchGet(&keys)
		suite.Nil(err)
		suite.Equal(len(completedDataMaps), len(*kvPairs))
	}

	uploadSessions = []models.UploadSession{}
	err = suite.DB.All(&uploadSessions)
	suite.Nil(err)

	completedUploads = []models.CompletedUpload{}
	err = suite.DB.All(&completedUploads)
	suite.Nil(err)

	// verify final lengths are what we expected
	suite.Equal(2*(numChunks+1), len(allDataMaps))   // 2 data maps so 2 extra chunks
	suite.Equal(numChunks+1, len(completedDataMaps)) // 1 data map so 1 extra chunk
	suite.Equal(2, len(uploadSessions))
	suite.Equal(1, len(completedUploads))

	// for good measure, verify that it's only "abcdeff1" in completed_data_maps
	// and that "abcdeff1" is not in data_maps at all
	genHash1InDataMaps := []models.DataMap{}
	err = suite.DB.Where("genesis_hash = ?", "abcdeff1").All(&genHash1InDataMaps)
	suite.Equal(0, len(genHash1InDataMaps))
	suite.Nil(err)

	genHash1Completed := []models.CompletedDataMap{}
	err = suite.DB.Where("genesis_hash = ?", "abcdeff1").All(&genHash1Completed)
	suite.Equal(numChunks+1, len(genHash1Completed))
	suite.Nil(err)

	suite.Equal("SOME_BETA_ETH_ADDRESS1", completedUploads[0].ETHAddr)
	suite.Equal("abcdeff1", completedUploads[0].GenesisHash)
}

package jobs_test

import (
	"github.com/gobuffalo/pop/nulls"
	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/models"
)

func (suite *JobsSuite) Test_PurgeCompletedSessions() {
	fileBytesCount := 2500
	numChunks := 3

	uploadSession1 := models.UploadSession{
		GenesisHash:   "genHash1",
		FileSizeBytes: fileBytesCount,
		NumChunks:     numChunks,
		Type:          models.SessionTypeBeta,
		ETHAddrAlpha:  nulls.String{string("SOME_ALPHA_ETH_ADDRESS"), true},
		ETHAddrBeta:   nulls.String{string("SOME_BETA_ETH_ADDRESS"), true},
		ETHPrivateKey: "SOME_PRIVATE_KEY",
	}

	vErr, err := uploadSession1.StartUploadSession()
	suite.Equal(0, len(vErr.Errors))
	suite.Equal(nil, err)

	uploadSession2 := models.UploadSession{
		GenesisHash:   "genHash2",
		FileSizeBytes: fileBytesCount,
		NumChunks:     numChunks,
		Type:          models.SessionTypeAlpha,
	}

	vErr, err = uploadSession2.StartUploadSession()
	suite.Equal(0, len(vErr.Errors))
	suite.Equal(nil, err)

	uploadSession3 := models.UploadSession{
		GenesisHash:   "genHash3",
		FileSizeBytes: fileBytesCount,
		NumChunks:     numChunks,
		Type:          models.SessionTypeAlpha,
	}

	vErr, err = uploadSession3.StartUploadSession()
	suite.Equal(0, len(vErr.Errors))
	suite.Equal(nil, err)

	allDataMaps := []models.DataMap{}
	err = suite.DB.All(&allDataMaps)
	suite.Equal(nil, err)

	completedDataMaps := []models.CompletedDataMap{}
	err = suite.DB.All(&completedDataMaps)
	suite.Equal(nil, err)

	uploadSessions := []models.UploadSession{}
	err = suite.DB.All(&uploadSessions)
	suite.Equal(nil, err)

	storedGenHashes := []models.StoredGenesisHash{}
	err = suite.DB.All(&storedGenHashes)
	suite.Equal(nil, err)

	completedUploads := []models.CompletedUpload{}
	err = suite.DB.All(&completedUploads)
	suite.Equal(nil, err)

	// verify initial lengths are what we expected
	suite.Equal(3*(numChunks+1), len(allDataMaps)) // 3 data maps so 3 extra chunks have been added
	suite.Equal(0, len(completedDataMaps))
	suite.Equal(3, len(uploadSessions))
	suite.Equal(0, len(storedGenHashes))
	suite.Equal(0, len(completedUploads))

	// set all chunks of first data map to complete or confirmed
	allDone := []models.DataMap{}
	err = suite.DB.Where("genesis_hash = ?", "genHash1").All(&allDone)
	suite.Equal(nil, err)

	for _, dataMap := range allDone {
		dataMap.Status = models.Complete
		suite.DB.ValidateAndSave(&dataMap)
	}

	// set one of them to "confirmed"
	allDone[1].Status = models.Confirmed
	suite.DB.ValidateAndSave(&allDone[1])

	// set one chunk of second data map to complete
	someDone := []models.DataMap{}
	err = suite.DB.Where("genesis_hash = ?", "genHash2").All(&someDone)
	suite.Equal(nil, err)

	someDone[0].Status = models.Complete
	suite.DB.ValidateAndSave(&someDone[0])

	//call method under test
	jobs.PurgeCompletedSessions()

	allDataMaps = []models.DataMap{}
	err = suite.DB.All(&allDataMaps)
	suite.Equal(nil, err)

	completedDataMaps = []models.CompletedDataMap{}
	err = suite.DB.All(&completedDataMaps)
	suite.Equal(nil, err)

	uploadSessions = []models.UploadSession{}
	err = suite.DB.All(&uploadSessions)
	suite.Equal(nil, err)

	storedGenHashes = []models.StoredGenesisHash{}
	err = suite.DB.All(&storedGenHashes)
	suite.Equal(nil, err)

	completedUploads = []models.CompletedUpload{}
	err = suite.DB.All(&completedUploads)
	suite.Equal(nil, err)

	// verify final lengths are what we expected
	suite.Equal(2*(numChunks+1), len(allDataMaps))   // 2 data maps so 2 extra chunks
	suite.Equal(numChunks+1, len(completedDataMaps)) // 1 data map so 1 extra chunk
	suite.Equal(2, len(uploadSessions))
	suite.Equal(1, len(storedGenHashes))
	suite.Equal(1, len(completedUploads))

	// for good measure, verify that it's only "genHash1" in completed_data_maps
	// and that "genHash1" is not in data_maps at all
	genHash1InDataMaps := []models.DataMap{}
	err = suite.DB.Where("genesis_hash = ?", "genHash1").All(&genHash1InDataMaps)
	suite.Equal(0, len(genHash1InDataMaps))
	suite.Equal(nil, err)

	genHash1Completed := []models.CompletedDataMap{}
	err = suite.DB.Where("genesis_hash = ?", "genHash1").All(&genHash1Completed)
	suite.Equal(numChunks+1, len(genHash1Completed))
	suite.Equal(nil, err)

	suite.Equal("SOME_BETA_ETH_ADDRESS", completedUploads[0].ETHAddr)
	suite.Equal("SOME_PRIVATE_KEY", completedUploads[0].ETHPrivateKey)
	suite.Equal("genHash1", completedUploads[0].GenesisHash)
}

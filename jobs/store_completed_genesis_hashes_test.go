package jobs_test

import (
	"github.com/gobuffalo/pop/nulls"
	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/utils"
)

func (suite *JobsSuite) Test_StoreCompletedGenesisHashes() {

	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	defer oyster_utils.ResetBrokerMode()

	fileBytesCount := uint64(2500)
	numChunks := 3
	privateKey := "1111111111111111111111111111111111111111111111111111111111111111"

	suite.DB.ValidateAndSave(&models.StoredGenesisHash{
		GenesisHash: "abcdeff4",
	})

	uploadSession1 := models.UploadSession{
		GenesisHash:   "abcdeff1",
		FileSizeBytes: fileBytesCount,
		NumChunks:     numChunks,
		Type:          models.SessionTypeBeta,
		ETHAddrAlpha:  nulls.String{string("SOME_ALPHA_ETH_ADDRESS1"), true},
		ETHAddrBeta:   nulls.String{string("SOME_BETA_ETH_ADDRESS1"), true},
		ETHPrivateKey: privateKey,
	}

	vErr, err := uploadSession1.StartUploadSession()
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

	vErr, err = uploadSession2.StartUploadSession()
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

	vErr, err = uploadSession3.StartUploadSession()
	suite.False(vErr.HasAny())
	suite.Nil(err)

	uploadSession4 := models.UploadSession{
		GenesisHash:   "abcdeff4",
		FileSizeBytes: fileBytesCount,
		NumChunks:     numChunks,
		Type:          models.SessionTypeAlpha,
		ETHAddrAlpha:  nulls.String{string("SOME_ALPHA_ETH_ADDRESS4"), true},
		ETHAddrBeta:   nulls.String{string("SOME_BETA_ETH_ADDRESS4"), true},
		ETHPrivateKey: privateKey,
	}

	vErr, err = uploadSession4.StartUploadSession()
	suite.False(vErr.HasAny())
	suite.Nil(err)

	storedGenHashes := []models.StoredGenesisHash{}
	err = suite.DB.All(&storedGenHashes)
	suite.Nil(err)

	// verify initial lengths are what we expected
	// we created one to start with
	suite.Equal(1, len(storedGenHashes))

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

	// set all chunks of fourth data map to complete or confirmed
	// this entry will already be in the stored genesis hashes table but it should not add it again
	allDoneButAlreadyInStoredGenesisHashesTable := []models.DataMap{}
	err = suite.DB.Where("genesis_hash = ?", "abcdeff4").All(&allDoneButAlreadyInStoredGenesisHashesTable)
	suite.Nil(err)

	for _, dataMap := range allDoneButAlreadyInStoredGenesisHashesTable {
		dataMap.Status = models.Complete
		suite.DB.ValidateAndSave(&dataMap)
	}

	//call method under test
	jobs.StoreCompletedGenesisHashes(jobs.PrometheusWrapper)

	storedGenHashes = []models.StoredGenesisHash{}
	err = suite.DB.All(&storedGenHashes)
	suite.Nil(err)

	// verify final lengths is what we expected
	// one new entry has been added
	suite.Equal(2, len(storedGenHashes))
}

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

	fileBytesCount := uint64(30000)
	numChunks := 30
	privateKey := "1111111111111111111111111111111111111111111111111111111111111111"

	// we will set an upload session to use this same genesis hash to make sure a new one
	// does not get created
	suite.DB.ValidateAndSave(&models.StoredGenesisHash{
		GenesisHash: "abcdef38",
	})

	uploadSession1 := models.UploadSession{
		GenesisHash:   "abcdef39",
		FileSizeBytes: fileBytesCount,
		NumChunks:     numChunks,
		Type:          models.SessionTypeBeta,
		ETHAddrAlpha:  nulls.String{string("SOME_ALPHA_ETH_ADDRESS1"), true},
		ETHAddrBeta:   nulls.String{string("SOME_BETA_ETH_ADDRESS1"), true},
		ETHPrivateKey: privateKey,
	}

	SessionSetUpForTest(&uploadSession1, []int{15}, uploadSession1.NumChunks)

	uploadSession2 := models.UploadSession{
		GenesisHash:   "abcdef40",
		FileSizeBytes: fileBytesCount,
		NumChunks:     numChunks,
		Type:          models.SessionTypeAlpha,
		ETHAddrAlpha:  nulls.String{string("SOME_ALPHA_ETH_ADDRESS2"), true},
		ETHAddrBeta:   nulls.String{string("SOME_BETA_ETH_ADDRESS2"), true},
		ETHPrivateKey: privateKey,
	}

	SessionSetUpForTest(&uploadSession2, []int{15}, uploadSession2.NumChunks)

	uploadSession3 := models.UploadSession{
		GenesisHash:   "abcdef41",
		FileSizeBytes: fileBytesCount,
		NumChunks:     numChunks,
		Type:          models.SessionTypeAlpha,
		ETHAddrAlpha:  nulls.String{string("SOME_ALPHA_ETH_ADDRESS3"), true},
		ETHAddrBeta:   nulls.String{string("SOME_BETA_ETH_ADDRESS3"), true},
		ETHPrivateKey: privateKey,
	}

	SessionSetUpForTest(&uploadSession3, []int{15}, uploadSession3.NumChunks)

	// setting upload session to the same genesis hash as the stored genesis hash that already
	// exists, to make sure a new one does not get created.
	uploadSession4 := models.UploadSession{
		GenesisHash:   "abcdef38",
		FileSizeBytes: fileBytesCount,
		NumChunks:     numChunks,
		Type:          models.SessionTypeAlpha,
		ETHAddrAlpha:  nulls.String{string("SOME_ALPHA_ETH_ADDRESS4"), true},
		ETHAddrBeta:   nulls.String{string("SOME_BETA_ETH_ADDRESS4"), true},
		ETHPrivateKey: privateKey,
	}

	SessionSetUpForTest(&uploadSession4, []int{15}, uploadSession4.NumChunks)

	storedGenHashes := []models.StoredGenesisHash{}
	err := suite.DB.All(&storedGenHashes)
	suite.Nil(err)

	// verify initial lengths are what we expected
	// we created one to start with
	suite.Equal(1, len(storedGenHashes))

	// set first session's indexes so that it will be regarded as complete
	firstSession := models.UploadSession{}
	suite.DB.Where("genesis_hash = ?", uploadSession1.GenesisHash).First(&firstSession)
	firstSession.NextIdxToAttach = -1
	firstSession.NextIdxToVerify = -1
	firstSession.AllDataReady = models.AllDataReady
	firstSession.PaymentStatus = models.PaymentStatusConfirmed
	firstSession.TreasureStatus = models.TreasureInDataMapComplete
	suite.DB.ValidateAndSave(&firstSession)

	// set second session's indexes to midway through the map
	secondSession := models.UploadSession{}
	suite.DB.Where("genesis_hash = ?", uploadSession2.GenesisHash).First(&secondSession)
	secondSession.NextIdxToAttach = int64(secondSession.NumChunks / 2)
	secondSession.NextIdxToVerify = int64(secondSession.NumChunks / 2)
	secondSession.AllDataReady = models.AllDataReady
	secondSession.PaymentStatus = models.PaymentStatusConfirmed
	secondSession.TreasureStatus = models.TreasureInDataMapComplete
	suite.DB.ValidateAndSave(&secondSession)

	// set fourth session's indexes so that it will be regarded as complete
	fourthSession := models.UploadSession{}
	suite.DB.Where("genesis_hash = ?", uploadSession4.GenesisHash).First(&fourthSession)
	fourthSession.NextIdxToAttach = int64(fourthSession.NumChunks)
	fourthSession.NextIdxToVerify = int64(fourthSession.NumChunks)
	fourthSession.AllDataReady = models.AllDataReady
	fourthSession.PaymentStatus = models.PaymentStatusConfirmed
	fourthSession.TreasureStatus = models.TreasureInDataMapComplete
	suite.DB.ValidateAndSave(&fourthSession)

	//call method under test
	jobs.StoreCompletedGenesisHashes(jobs.PrometheusWrapper)

	storedGenHashes = []models.StoredGenesisHash{}
	err = suite.DB.All(&storedGenHashes)
	suite.Nil(err)

	// verify final lengths is what we expected
	// one new entry has been added
	for _, hash := range storedGenHashes {
		suite.True(hash.GenesisHash == "abcdef38" || hash.GenesisHash == uploadSession1.GenesisHash)
	}

	suite.Equal(2, len(storedGenHashes))
}

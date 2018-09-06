package jobs_test

import (
	"github.com/gobuffalo/pop/nulls"
	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/utils"
)

func (suite *JobsSuite) Test_BuryTreasureInDataMaps() {

	fileBytesCount := uint64(300000)

	// This map seems pointless but it makes the testing
	// in the for loop later on a bit simpler
	treasureIndexes := map[int]int{}
	treasureIndexes[5] = 5
	treasureIndexes[78] = 78
	treasureIndexes[199] = 199

	// create another dummy TreasureIdxMap for the data maps
	// who already have treasure buried
	testMap2 := `[{
		"sector": 1,
		"idx": 15,
		"key": "firstKeySecondMap"
		},
		{
		"sector": 2,
		"idx": 112,
		"key": "secondKeySecondMap"
		},
		{
		"sector": 3,
		"idx": 275,
		"key": "thirdKeySecondMap"
		}]`

	// create and start the upload session for the data maps that need treasure buried
	uploadSession1 := models.UploadSession{
		GenesisHash:    oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		NumChunks:      300,
		FileSizeBytes:  fileBytesCount,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureInDataMapPending,
	}

	mergedIndexes := []int{treasureIndexes[5]}
	privateKeys := []string{"1000000001"}

	chunkReqs := GenerateChunkRequests(300, uploadSession1.GenesisHash)

	models.ProcessAndStoreChunkData(chunkReqs, uploadSession1.GenesisHash, mergedIndexes, oyster_utils.TestValueTimeToLive)

	_, err := uploadSession1.StartUploadSession()
	suite.Nil(err)

	uploadSession1.MakeTreasureIdxMap(mergedIndexes, privateKeys)

	// create and start the upload session for a data map that will not need a treasure (payment still invoiced)
	uploadSession2 := models.UploadSession{
		GenesisHash:    oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		NumChunks:      300,
		FileSizeBytes:  fileBytesCount,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  models.PaymentStatusInvoiced,
		TreasureStatus: models.TreasureInDataMapPending,
		TreasureIdxMap: nulls.String{string(testMap2), true},
	}

	mergedIndexes = []int{112}
	privateKeys = []string{"2000000002"}

	chunkReqs = GenerateChunkRequests(300, uploadSession2.GenesisHash)
	models.ProcessAndStoreChunkData(chunkReqs, uploadSession2.GenesisHash, mergedIndexes, oyster_utils.TestValueTimeToLive)

	_, err = uploadSession2.StartUploadSession()
	suite.Nil(err)

	for {
		jobs.BuryTreasureInDataMaps()
		session := models.UploadSession{}
		models.DB.Find(&session, uploadSession1.ID)
		finishedMessages, _ := session.WaitForAllMessages(3)
		finishedHashes, _ := session.WaitForAllHashes(3)
		if finishedMessages && finishedHashes {
			break
		}
	}

	session := []models.UploadSession{}

	suite.DB.Where("treasure_status = ?", models.TreasureInDataMapComplete).All(&session)
	suite.Equal(1, len(session))
	suite.Equal(uploadSession1.GenesisHash, session[0].GenesisHash)

	jobs.CheckSessionsWithIncompleteData()

	u := models.UploadSession{}
	models.DB.Find(&u, session[0].ID)

	suite.Equal(models.AllDataReady, u.AllDataReady)
}

func (suite *JobsSuite) Test_BuryTreasure() {

	fileBytesCount := uint64(300000)

	// This map seems pointless but it makes the testing
	// in the for loop later on a bit simpler
	treasureIndexes := map[int]int{}
	treasureIndexes[5] = 5
	treasureIndexes[78] = 78
	treasureIndexes[199] = 199

	// create and start the upload session for the data maps that need treasure buried
	uploadSession1 := models.UploadSession{
		GenesisHash:    oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		NumChunks:      300,
		FileSizeBytes:  fileBytesCount,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureInDataMapPending,
	}

	_, err := uploadSession1.StartUploadSession()
	suite.Nil(err)

	mergedIndexes := []int{treasureIndexes[5], treasureIndexes[78], treasureIndexes[199]}
	privateKeys := []string{"0000000001", "0000000002", "0000000003"}

	chunkReqs := GenerateChunkRequests(300, uploadSession1.GenesisHash)

	models.ProcessAndStoreChunkData(chunkReqs, uploadSession1.GenesisHash, mergedIndexes, oyster_utils.TestValueTimeToLive)

	uploadSession1.MakeTreasureIdxMap(mergedIndexes, privateKeys)

	session := []models.UploadSession{}
	suite.DB.Where("genesis_hash = ?", uploadSession1.GenesisHash).All(&session)

	for {
		ready := session[0].CheckIfAllHashesAreReady()
		if ready {
			break
		}
	}

	//Check that we have hash and address data for all 3 chunks,
	//but not message data
	chunkData1 := models.GetSingleChunkData(oyster_utils.InProgressDir, uploadSession1.GenesisHash,
		int64(treasureIndexes[5]))

	suite.NotEqual("", chunkData1.Hash)
	suite.NotEqual("", chunkData1.Address)
	suite.Equal("", chunkData1.Message)

	chunkData2 := models.GetSingleChunkData(oyster_utils.InProgressDir, uploadSession1.GenesisHash,
		int64(treasureIndexes[78]))

	suite.NotEqual("", chunkData2.Hash)
	suite.NotEqual("", chunkData2.Address)

	chunkData3 := models.GetSingleChunkData(oyster_utils.InProgressDir, uploadSession1.GenesisHash,
		int64(treasureIndexes[199]))

	suite.NotEqual("", chunkData3.Hash)
	suite.NotEqual("", chunkData3.Address)

	for {
		//Call BuryTreasureInDataMaps, which will in turn call BuryTreasure, the method under test
		jobs.BuryTreasureInDataMaps()
		finishedMessages, _ := session[0].WaitForAllMessages(3)
		if finishedMessages {
			break
		}
	}

	//Check that we now have message data for all 3 chunks
	chunkData1 = models.GetSingleChunkData(oyster_utils.InProgressDir, uploadSession1.GenesisHash,
		int64(treasureIndexes[5]))

	suite.NotEqual("", chunkData1.Message)

	chunkData2 = models.GetSingleChunkData(oyster_utils.InProgressDir, uploadSession1.GenesisHash,
		int64(treasureIndexes[78]))

	suite.NotEqual("", chunkData2.Message)

	chunkData3 = models.GetSingleChunkData(oyster_utils.InProgressDir, uploadSession1.GenesisHash,
		int64(treasureIndexes[199]))

	suite.NotEqual("", chunkData3.Message)

	//Check that the session's TreasureStatus has changed
	session = []models.UploadSession{}
	suite.DB.Where("genesis_hash = ?", uploadSession1.GenesisHash).All(&session)
	suite.Equal(models.TreasureInDataMapComplete, session[0].TreasureStatus)
	suite.Equal(models.AllDataReady, session[0].AllDataReady)
}

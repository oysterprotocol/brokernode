package jobs_test

import (
	"github.com/gobuffalo/pop/nulls"
	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
)

func (suite *JobsSuite) Test_ProcessPaidSessions() {
	fileBytesCount := uint64(500000)

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
		"idx": 155,
		"key": "firstKeySecondMap"
		},
		{
		"sector": 2,
		"idx": 204,
		"key": "secondKeySecondMap"
		},
		{
		"sector": 3,
		"idx": 599,
		"key": "thirdKeySecondMap"
		}]`

	// create and start the upload session for the data maps that need treasure buried
	uploadSession1 := models.UploadSession{
		GenesisHash:    "abcdeff1",
		NumChunks:      500,
		FileSizeBytes:  fileBytesCount,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureInDataMapPending,
	}

	uploadSession1.StartUploadSession()
	mergedIndexes := []int{treasureIndexes[5], treasureIndexes[78], treasureIndexes[199]}
	privateKeys := []string{"0000000001", "0000000002", "0000000003"}

	uploadSession1.MakeTreasureIdxMap(mergedIndexes, privateKeys)
	uploadSession1.EncryptTreasureIdxMapKeys()

	// create and start the upload session for the data maps that already have buried treasure
	uploadSession2 := models.UploadSession{
		GenesisHash:    "abcdeff2",
		NumChunks:      500,
		FileSizeBytes:  fileBytesCount,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureInDataMapComplete,
		TreasureIdxMap: nulls.String{string(testMap2), true},
	}

	uploadSession2.StartUploadSession()
	mergedIndexes = []int{treasureIndexes[5], treasureIndexes[78], treasureIndexes[199]}
	privateKeys = []string{"0000000001", "0000000002", "0000000003"}

	uploadSession2.MakeTreasureIdxMap(mergedIndexes, privateKeys)
	uploadSession2.EncryptTreasureIdxMapKeys()

	// verify that we have successfully created all the data maps
	paidButUnburied := []models.DataMap{}
	err := suite.DB.Where("genesis_hash = ?", "abcdeff1").All(&paidButUnburied)
	suite.Nil(err)

	paidAndBuried := []models.DataMap{}
	err = suite.DB.Where("genesis_hash = ?", "abcdeff2").All(&paidAndBuried)
	suite.Nil(err)

	suite.NotEqual(0, len(paidButUnburied))
	suite.NotEqual(0, len(paidAndBuried))

	// verify that the "Message" field for every chunk in paidButUnburied is ""
	for _, dMap := range paidButUnburied {
		if services.IsKvStoreEnabled() {
			suite.Nil(services.BatchSet(&services.KVPairs{dMap.MsgID: "NOTEMPTY"}))
		} else {
			dMap.Message = "NOTEMPTY"
		}
		dMap.MsgStatus = models.MsgStatusUploadedNoNeedEncode
		suite.DB.ValidateAndSave(&dMap)
	}

	// verify that the "Status" field for every chunk in paidAndBuried is NOT Unassigned
	for _, dMap := range paidAndBuried {
		suite.NotEqual(models.Unassigned, dMap.Status)
		if services.IsKvStoreEnabled() {
			suite.Nil(services.BatchSet(&services.KVPairs{dMap.MsgID: "NOTEMPTY"}))
		} else {
			dMap.Message = "NOTEMPTY"
		}
		dMap.MsgStatus = models.MsgStatusUploadedNoNeedEncode
		suite.DB.ValidateAndSave(&dMap)
	}

	// call method under test
	jobs.ProcessPaidSessions(jobs.PrometheusWrapper)

	paidButUnburied = []models.DataMap{}
	err = suite.DB.Where("genesis_hash = ?", "abcdeff1").All(&paidButUnburied)
	suite.Nil(err)

	/* Verify the following:
	1.  If a chunk in paidButUnburied was one of the treasure chunks, Message is no longer ""
	2.  Status of all data maps in paidButUnburied is now Unassigned (to get picked up by process_unassigned_chunks
	*/
	for _, dMap := range paidButUnburied {
		if _, ok := treasureIndexes[dMap.ChunkIdx]; ok {
			suite.NotEqual("", services.GetMessageFromDataMap(dMap))
		} else {
			suite.Equal("NOTEMPTY", services.GetMessageFromDataMap(dMap))
		}
		suite.Equal(models.Unassigned, dMap.Status)
	}

	paidAndBuried = []models.DataMap{}
	err = suite.DB.Where("genesis_hash = ?", "abcdeff2").All(&paidAndBuried)
	suite.Nil(err)

	// verify that all chunks in paidAndBuried have statuses changed to Unassigned
	for _, dMap := range paidAndBuried {
		suite.Equal(models.Unassigned, dMap.Status)
	}

	// get the session that was originally paid but unburied, and verify that all the
	// keys are now "" but that we still have a value for the Idx
	paidAndUnburiedSession := models.UploadSession{}
	err = suite.DB.Where("genesis_hash = ?", "abcdeff1").First(&paidAndUnburiedSession)
	suite.Nil(err)

	treasureIndex, err := paidAndUnburiedSession.GetTreasureMap()
	suite.Nil(err)

	suite.Equal(3, len(treasureIndex))

	for _, entry := range treasureIndex {
		_, ok := treasureIndexes[entry.Idx]
		suite.True(ok)
	}
}

func (suite *JobsSuite) Test_EncryptKeysInTreasureIdxMaps() {

	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	defer oyster_utils.ResetBrokerMode()

	fileBytesCount := uint64(500000)

	// This map seems pointless but it makes the testing
	// in the for loop later on a bit simpler
	treasureIndexes := map[int]int{}
	treasureIndexes[5] = 5
	treasureIndexes[78] = 78
	treasureIndexes[199] = 199

	// create and start the upload session for the data maps that need treasure buried
	uploadSession1 := models.UploadSession{
		GenesisHash:    "abcdeff111111111111111111111111111111111111111111111",
		NumChunks:      500,
		FileSizeBytes:  fileBytesCount,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureGeneratingKeys,
	}

	uploadSession1.StartUploadSession()
	mergedIndexes := []int{
		treasureIndexes[5],
		treasureIndexes[78],
		treasureIndexes[199],
	}

	privateKey1 := "0000000001"
	privateKey2 := "0000000002"
	privateKey3 := "0000000003"

	uploadSession1.MakeTreasureIdxMap(mergedIndexes, []string{
		privateKey1,
		privateKey2,
		privateKey3,
	})

	treasureMap, err := uploadSession1.GetTreasureMap()
	suite.Nil(err)

	suite.Equal(3, len(treasureMap))

	for _, entry := range treasureMap {
		suite.True(entry.Key == privateKey1 ||
			entry.Key == privateKey2 ||
			entry.Key == privateKey3)
	}

	jobs.EncryptKeysInTreasureIdxMaps()

	uploadSession := models.UploadSession{}
	suite.DB.Where("genesis_hash = ?",
		uploadSession1.GenesisHash).First(&uploadSession)

	treasureMap2, err := uploadSession.GetTreasureMap()
	suite.Nil(err)

	suite.Equal(3, len(treasureMap2))

	for _, entry := range treasureMap2 {
		suite.True(entry.Key != privateKey1 &&
			entry.Key != privateKey2 &&
			entry.Key != privateKey3)
	}

	for _, entry := range treasureMap2 {
		dm := models.DataMap{}
		err := suite.DB.Where("genesis_hash = ? && chunk_idx = ?",
			uploadSession1.GenesisHash,
			entry.Idx).First(&dm)
		suite.Nil(err)

		decryptedKey, err := dm.DecryptEthKey(entry.Key)
		suite.Nil(err)

		suite.True(decryptedKey == privateKey1 ||
			decryptedKey == privateKey2 ||
			decryptedKey == privateKey3)
	}
}

package jobs_test

import (
	"github.com/gobuffalo/pop/nulls"
	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/models"
)

func (suite *JobsSuite) Test_ProcessPaidSessions() {
	fileBytesCount := 500000

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
		GenesisHash:    "genHash1",
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

	// create and start the upload session for the data maps that already have buried treasure
	uploadSession2 := models.UploadSession{
		GenesisHash:    "genHash2",
		NumChunks:      500,
		FileSizeBytes:  fileBytesCount,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureInDataMapComplete,
		TreasureIdxMap: nulls.String{string(testMap2), true},
	}

	uploadSession2.StartUploadSession()

	// verify that we have successfully created all the data maps
	paidButUnburied := []models.DataMap{}
	err := suite.DB.Where("genesis_hash = ?", "genHash1").All(&paidButUnburied)
	suite.Equal(nil, err)

	paidAndBuried := []models.DataMap{}
	err = suite.DB.Where("genesis_hash = ?", "genHash2").All(&paidAndBuried)
	suite.Equal(nil, err)

	suite.NotEqual(0, len(paidButUnburied))
	suite.NotEqual(0, len(paidAndBuried))

	// verify that the "Message" field for every chunk in paidButUnburied is ""
	for _, dMap := range paidButUnburied {
		dMap.Message = "NOTEMPTY"
		suite.DB.ValidateAndSave(&dMap)
	}

	// verify that the "Status" field for every chunk in paidAndBuried is NOT Unassigned
	for _, dMap := range paidAndBuried {
		suite.NotEqual(models.Unassigned, dMap.Status)
		dMap.Message = "NOTEMPTY"
		suite.DB.ValidateAndSave(&dMap)
	}

	// call method under test
	jobs.ProcessPaidSessions()

	paidButUnburied = []models.DataMap{}
	err = suite.DB.Where("genesis_hash = ?", "genHash1").All(&paidButUnburied)
	suite.Equal(nil, err)

	/* Verify the following:
	1.  If a chunk in paidButUnburied was one of the treasure chunks, Message is no longer ""
	2.  Status of all data maps in paidButUnburied is now Unassigned (to get picked up by process_unassigned_chunks
	*/
	for _, dMap := range paidButUnburied {
		if _, ok := treasureIndexes[dMap.ChunkIdx]; ok {
			suite.NotEqual("", dMap.Message)
		} else {
			suite.Equal("NOTEMPTY", dMap.Message)
		}
		suite.Equal(models.Unassigned, dMap.Status)
	}

	paidAndBuried = []models.DataMap{}
	err = suite.DB.Where("genesis_hash = ?", "genHash2").All(&paidAndBuried)
	suite.Equal(nil, err)

	// verify that all chunks in paidAndBuried have statuses changed to Unassigned
	for _, dMap := range paidAndBuried {
		suite.Equal(models.Unassigned, dMap.Status)
	}

	// get the session that was originally paid but unburied, and verify that all the
	// keys are now "" but that we still have a value for the Idx
	paidAndUnburiedSession := models.UploadSession{}
	err = suite.DB.Where("genesis_hash = ?", "genHash1").First(&paidAndUnburiedSession)
	suite.Equal(nil, err)

	treasureIndex, err := paidAndUnburiedSession.GetTreasureMap()
	suite.Equal(nil, err)

	suite.Equal(3, len(treasureIndex))

	for _, entry := range treasureIndex {
		_, ok := treasureIndexes[entry.Idx]
		suite.Equal(true, ok)
	}
}

package jobs_test

import (
	"github.com/gobuffalo/pop/nulls"
	"github.com/iotaledger/giota"
	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"time"
)

var (
	sendChunksToChannelMockCalled_process_unassigned_chunks              = false
	verifyChunkMessagesMatchesRecordMockCalled_process_unassigned_chunks = false
	findTransactionsMockCalled_process_unassigned_chunks                 = false
	AllChunksCalled                                                      []models.DataMap
	fakeFindTransactionsAddress                                          = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
)

func (suite *JobsSuite) Test_ProcessUnassignedChunks() {

	numChunks := 31

	// reset back to generic mocks
	defer suite.SetupSuite()

	// make suite available inside mock methods
	Suite = *suite

	// assign the mock methods for this test
	makeMocks_process_unassigned_chunks(&IotaMock)

	// make 3 channels
	models.MakeChannels(3)

	uploadSession1 := models.UploadSession{
		GenesisHash:    "genHash1",
		NumChunks:      numChunks,
		FileSizeBytes:  3000,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureBuried,
	}

	uploadSession2 := models.UploadSession{
		GenesisHash:    "genHash2",
		NumChunks:      numChunks,
		FileSizeBytes:  3000,
		Type:           models.SessionTypeBeta,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureBuried,
	}

	uploadSession3 := models.UploadSession{
		GenesisHash:    "genHash3",
		NumChunks:      numChunks,
		FileSizeBytes:  3000,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureBuried,
	}

	uploadSession4 := models.UploadSession{
		GenesisHash:    "genHash4",
		NumChunks:      numChunks,
		FileSizeBytes:  3000,
		Type:           models.SessionTypeBeta,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureBuried,
	}

	uploadSession1.StartUploadSession()
	uploadSession2.StartUploadSession()
	uploadSession3.StartUploadSession()
	uploadSession4.StartUploadSession()

	// set uploadSession4 to be the oldest
	err := suite.DB.RawQuery("UPDATE upload_sessions SET created_at = ? WHERE genesis_hash = ?",
		time.Now().Add(-20*time.Second), "genHash4").All(&[]models.UploadSession{})
	suite.Nil(err)

	// set uploadSession2 to next oldest
	err = suite.DB.RawQuery("UPDATE upload_sessions SET created_at = ? WHERE genesis_hash = ?",
		time.Now().Add(-15*time.Second), "genHash2").All(&[]models.UploadSession{})
	suite.Nil(err)

	// set uploadSession1 to next oldest after uploadSession2
	err = suite.DB.RawQuery("UPDATE upload_sessions SET created_at = ? WHERE genesis_hash = ?",
		time.Now().Add(-10*time.Second), "genHash1").All(&[]models.UploadSession{})
	suite.Nil(err)

	// uploadSession3 will be the newest

	// set all data maps to unassigned
	err = suite.DB.RawQuery("UPDATE data_maps SET status = ?", models.Unassigned).All(&[]models.DataMap{})
	suite.Nil(err)

	// call method under test
	jobs.ProcessUnassignedChunks(IotaMock)

	suite.Equal(true, sendChunksToChannelMockCalled_process_unassigned_chunks)
	suite.Equal(true, verifyChunkMessagesMatchesRecordMockCalled_process_unassigned_chunks)
	suite.Equal(4*(numChunks+1), len(AllChunksCalled)) // 4 data maps so 4 chunks have been added

	/* This test is verifying that the chunks belonging to particular sessions were sent
	in the order we would expect and that the ordering of chunk ids within each data map was
	appropriate for alpha or beta.

		Session 4 was the oldest session and was of type beta
		Session 2 was next oldest and was of type beta
		Session 1 was next-to-last oldest and was of type alpha
		Session 3 was newest and was of type alpha


	If BundleSize is changed, these tests will need to be changed.
	*/

	genHashMapIdx := map[string][]int{}
	genHashMapOrder := map[string]int{}
	i := 0

	for _, chunk := range AllChunksCalled {
		genHashMapIdx[chunk.GenesisHash] = append(genHashMapIdx[chunk.GenesisHash], chunk.ChunkIdx)
		if _, ok := genHashMapOrder[chunk.GenesisHash]; !ok {
			genHashMapOrder[chunk.GenesisHash] = i
			i++
		}
	}

	suite.Equal(true, genHashMapIdx["genHash4"][0] > genHashMapIdx["genHash4"][len(genHashMapIdx["genHash4"])-1])
	suite.Equal(true, genHashMapIdx["genHash2"][0] > genHashMapIdx["genHash2"][len(genHashMapIdx["genHash2"])-1])
	suite.Equal(true, genHashMapIdx["genHash1"][0] < genHashMapIdx["genHash1"][len(genHashMapIdx["genHash1"])-1])
	suite.Equal(true, genHashMapIdx["genHash3"][0] < genHashMapIdx["genHash3"][len(genHashMapIdx["genHash3"])-1])

	suite.Equal(0, genHashMapOrder["genHash4"])
	suite.Equal(1, genHashMapOrder["genHash2"])
	suite.Equal(2, genHashMapOrder["genHash1"])
	suite.Equal(3, genHashMapOrder["genHash3"])
}

func (suite *JobsSuite) Test_HandleTreasureChunks() {

	numChunks := 25

	// reset back to generic mocks
	defer suite.SetupSuite()

	// make suite available inside mock methods
	Suite = *suite

	// assign the mock methods for this test
	makeMocks_process_unassigned_chunks(&IotaMock)

	// make 3 channels
	models.MakeChannels(3)

	treasureMap := `[{
		"sector": 1,
		"idx": 15,
		"key": "000000000000000000000000000000010000000000000000000000000000000001"
		},
		{
		"sector": 1,
		"idx": 23,
		"key": "000000000000000000000000000000020000000000000000000000000000000002"
		}]`

	uploadSession1 := models.UploadSession{
		GenesisHash:    "genHash1",
		NumChunks:      numChunks,
		FileSizeBytes:  3000,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureBuried,
		TreasureIdxMap: nulls.String{string(treasureMap), true},
	}

	uploadSession1.StartUploadSession()

	// set all data maps to unassigned
	err := suite.DB.RawQuery("UPDATE data_maps SET status = ?", models.Unassigned).All(&[]models.DataMap{})
	suite.Nil(err)

	treasureChunk := models.DataMap{}
	err = suite.DB.RawQuery("SELECT * from data_maps WHERE chunk_idx = ?", 23).First(&treasureChunk)
	suite.Nil(err)

	// setting the address to something that the findTransactions mock can check for
	treasureChunk.Address = fakeFindTransactionsAddress
	suite.DB.ValidateAndSave(&treasureChunk)

	dataMaps := []models.DataMap{}
	err = suite.DB.RawQuery("SELECT * from data_maps").All(&dataMaps)
	suite.Nil(err)

	// call method under test
	chunksToAttach, treasureChunks := jobs.HandleTreasureChunks(dataMaps, uploadSession1, IotaMock)

	for _, chunk := range chunksToAttach {
		suite.NotEqual(15, chunk.ChunkIdx)
	}

	suite.Equal(true, findTransactionsMockCalled_process_unassigned_chunks)
	suite.Equal(len(dataMaps)-2, len(chunksToAttach))
	suite.Equal(1, len(treasureChunks))
	suite.Equal(15, treasureChunks[0].ChunkIdx)
}

func (suite *JobsSuite) Test_InsertTreasureChunks_AlphaSession() {

	numChunks := 25

	uploadSession1 := models.UploadSession{
		GenesisHash:   "genHash1",
		NumChunks:     numChunks,
		FileSizeBytes: 3000,
		Type:          models.SessionTypeAlpha,
	}

	nonTreasureChunks := []models.DataMap{}
	treasureChunks := []models.DataMap{}

	for i := 2; i < 25; i++ {
		if i != 1 && i != 10 && i != 25 {
			nonTreasureChunks = append(nonTreasureChunks, models.DataMap{
				ChunkIdx: i,
			})
		}
	}

	treasureChunks = append(treasureChunks, models.DataMap{
		ChunkIdx: 1,
	})
	treasureChunks = append(treasureChunks, models.DataMap{
		ChunkIdx: 10,
	})
	treasureChunks = append(treasureChunks, models.DataMap{
		ChunkIdx: 25,
	})

	// call method under test
	allChunks := jobs.InsertTreasureChunks(nonTreasureChunks, treasureChunks, uploadSession1)

	suite.Equal(len(treasureChunks)+len(nonTreasureChunks), len(allChunks))

	// verify chunks are in the expected (ascending) order
	for i, chunk := range allChunks {
		if chunk.ChunkIdx == 1 {
			suite.Equal(2, allChunks[i+1].ChunkIdx)
		}
		if chunk.ChunkIdx == 10 {
			suite.Equal(9, allChunks[i-1].ChunkIdx)
			suite.Equal(11, allChunks[i+1].ChunkIdx)
		}
		if chunk.ChunkIdx == 25 {
			suite.Equal(24, allChunks[i-1].ChunkIdx)
		}
	}
}

func (suite *JobsSuite) Test_InsertTreasureChunks_BetaSession() {

	numChunks := 25

	uploadSession1 := models.UploadSession{
		GenesisHash:   "genHash1",
		NumChunks:     numChunks,
		FileSizeBytes: 3000,
		Type:          models.SessionTypeBeta,
	}

	nonTreasureChunks := []models.DataMap{}
	treasureChunks := []models.DataMap{}

	for i := 25; i > 0; i-- {
		if i != 1 && i != 10 && i != 25 {
			nonTreasureChunks = append(nonTreasureChunks, models.DataMap{
				ChunkIdx: i,
			})
		}
	}

	treasureChunks = append(treasureChunks, models.DataMap{
		ChunkIdx: 1,
	})
	treasureChunks = append(treasureChunks, models.DataMap{
		ChunkIdx: 10,
	})
	treasureChunks = append(treasureChunks, models.DataMap{
		ChunkIdx: 25,
	})

	// call method under test
	allChunks := jobs.InsertTreasureChunks(nonTreasureChunks, treasureChunks, uploadSession1)

	suite.Equal(len(treasureChunks)+len(nonTreasureChunks), len(allChunks))

	// verify chunks are in the expected (descending) order
	for i, chunk := range allChunks {
		if chunk.ChunkIdx == 1 {
			suite.Equal(2, allChunks[i-1].ChunkIdx)
		}
		if chunk.ChunkIdx == 10 {
			suite.Equal(9, allChunks[i+1].ChunkIdx)
			suite.Equal(11, allChunks[i-1].ChunkIdx)
		}
		if chunk.ChunkIdx == 25 {
			suite.Equal(24, allChunks[i+1].ChunkIdx)
		}
	}
}

func makeMocks_process_unassigned_chunks(iotaMock *services.IotaService) {
	iotaMock.VerifyChunkMessagesMatchRecord = verifyChunkMessagesMatchesRecordMock_process_unassigned_chunks
	iotaMock.SendChunksToChannel = sendChunksToChannelMock_process_unassigned_chunks
	iotaMock.FindTransactions = findTransactions_process_unassigned_chunks
}

func sendChunksToChannelMock_process_unassigned_chunks(chunks []models.DataMap, channel *models.ChunkChannel) {

	// our mock was called
	sendChunksToChannelMockCalled_process_unassigned_chunks = true

	// stored all the chunks that get sent to the mock so we can run tests on them.
	AllChunksCalled = append(AllChunksCalled, chunks...)
}

func verifyChunkMessagesMatchesRecordMock_process_unassigned_chunks(chunks []models.DataMap) (filteredChunks services.FilteredChunk, err error) {

	verifyChunkMessagesMatchesRecordMockCalled_process_unassigned_chunks = true

	matchesTangle := []models.DataMap{}
	doesNotMatchTangle := []models.DataMap{}
	notAttached := []models.DataMap{}

	// mark everything as unattached
	notAttached = append(notAttached, chunks...)

	return services.FilteredChunk{
		MatchesTangle:      matchesTangle,
		NotAttached:        notAttached,
		DoesNotMatchTangle: doesNotMatchTangle,
	}, err
}

func findTransactions_process_unassigned_chunks(addresses []giota.Address) (map[giota.Address][]giota.Transaction, error) {

	addrToTransactionMap := make(map[giota.Address][]giota.Transaction)

	if addresses[0] == giota.Address(fakeFindTransactionsAddress) {
		// only add to the map if the address is the address we decided to check for
		addrToTransactionMap[addresses[0]] = []giota.Transaction{}
	}

	findTransactionsMockCalled_process_unassigned_chunks = true

	return addrToTransactionMap, nil
}

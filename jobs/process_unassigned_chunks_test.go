package jobs_test

import (
	"fmt"
	"github.com/gobuffalo/pop/nulls"
	"github.com/iotaledger/giota"
	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"time"
)

var (
	sendChunksToChannelMockCalled_process_unassigned_chunks              = false
	verifyChunkMessagesMatchesRecordMockCalled_process_unassigned_chunks = false
	findTransactionsMockCalled_process_unassigned_chunks                 = false
	AllChunksCalled                                                      []models.DataMap
	fakeFindTransactionsAddress                                          = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
)

func (suite *JobsSuite) Test_ProcessUnassignedChunks() {

	oyster_utils.SetPoWMode(oyster_utils.PoWEnabled)
	defer oyster_utils.ResetPoWMode()

	numChunks := 31

	// make suite available inside mock methods
	Suite = *suite

	// assign the mock methods for this test
	makeMocks_process_unassigned_chunks(&IotaMock)

	// make 3 channels
	models.MakeChannels(3)

	uploadSession1 := models.UploadSession{
		GenesisHash:    "abcdeff1",
		NumChunks:      numChunks,
		FileSizeBytes:  3000,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureInDataMapComplete,
	}

	uploadSession2 := models.UploadSession{
		GenesisHash:    "abcdeff2",
		NumChunks:      numChunks,
		FileSizeBytes:  3000,
		Type:           models.SessionTypeBeta,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureInDataMapComplete,
	}

	uploadSession3 := models.UploadSession{
		GenesisHash:    "abcdeff3",
		NumChunks:      numChunks,
		FileSizeBytes:  3000,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureInDataMapComplete,
	}

	uploadSession4 := models.UploadSession{
		GenesisHash:    "abcdeff4",
		NumChunks:      numChunks,
		FileSizeBytes:  3000,
		Type:           models.SessionTypeBeta,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureInDataMapComplete,
	}

	chunksReady, _, err := uploadSession1.StartSessionAndWaitForChunks(500)
	suite.True(chunksReady)
	suite.Nil(err)

	chunksReady, _, err = uploadSession2.StartSessionAndWaitForChunks(500)
	suite.True(chunksReady)
	suite.Nil(err)

	chunksReady, _, err = uploadSession3.StartSessionAndWaitForChunks(500)
	suite.True(chunksReady)
	suite.Nil(err)

	chunksReady, _, err = uploadSession4.StartSessionAndWaitForChunks(500)
	suite.True(chunksReady)
	suite.Nil(err)

	// set uploadSession4 to be the oldest
	err = suite.DB.RawQuery("UPDATE upload_sessions SET created_at = ? WHERE genesis_hash = ?",
		time.Now().Add(-20*time.Second), "abcdeff4").All(&[]models.UploadSession{})
	suite.Nil(err)

	// set uploadSession2 to next oldest
	err = suite.DB.RawQuery("UPDATE upload_sessions SET created_at = ? WHERE genesis_hash = ?",
		time.Now().Add(-15*time.Second), "abcdeff2").All(&[]models.UploadSession{})
	suite.Nil(err)

	// set uploadSession1 to next oldest after uploadSession2
	err = suite.DB.RawQuery("UPDATE upload_sessions SET created_at = ? WHERE genesis_hash = ?",
		time.Now().Add(-10*time.Second), "abcdeff1").All(&[]models.UploadSession{})
	suite.Nil(err)

	// uploadSession3 will be the newest

	// set all data maps to unassigned
	err = suite.DB.RawQuery("UPDATE data_maps SET status = ?", models.Unassigned).All(&[]models.DataMap{})
	suite.Nil(err)

	// call method under test
	jobs.ProcessUnassignedChunks(IotaMock, jobs.PrometheusWrapper)

	suite.True(sendChunksToChannelMockCalled_process_unassigned_chunks)
	suite.True(verifyChunkMessagesMatchesRecordMockCalled_process_unassigned_chunks)
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

	suite.True(genHashMapIdx["abcdeff4"][0] > genHashMapIdx["abcdeff4"][len(genHashMapIdx["abcdeff4"])-1])
	suite.True(genHashMapIdx["abcdeff2"][0] > genHashMapIdx["abcdeff2"][len(genHashMapIdx["abcdeff2"])-1])
	suite.True(genHashMapIdx["abcdeff1"][0] < genHashMapIdx["abcdeff1"][len(genHashMapIdx["abcdeff1"])-1])
	suite.True(genHashMapIdx["abcdeff3"][0] < genHashMapIdx["abcdeff3"][len(genHashMapIdx["abcdeff3"])-1])

	suite.Equal(0, genHashMapOrder["abcdeff4"])
	suite.Equal(1, genHashMapOrder["abcdeff2"])
	suite.Equal(2, genHashMapOrder["abcdeff1"])
	suite.Equal(3, genHashMapOrder["abcdeff3"])
}

func (suite *JobsSuite) Test_HandleTreasureChunks() {

	numChunks := 25

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
		"idx": 20,
		"key": "000000000000000000000000000000020000000000000000000000000000000002"
		}]`

	uploadSession1 := models.UploadSession{
		GenesisHash:    "abcdeff1",
		NumChunks:      numChunks,
		FileSizeBytes:  3000,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureInDataMapComplete,
		TreasureIdxMap: nulls.String{string(treasureMap), true},
	}

	for i := 0; i < numChunks; i++ {
		suite.DB.ValidateAndSave(&models.DataMap{
			ChunkIdx:    i,
			GenesisHash: "abcdeff1",
			Hash:        "SOMEHASH",
			Address:     "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB",
			MsgID:       fmt.Sprintf("msg_id_%d", i),
			MsgStatus:   models.MsgStatusNotUploaded,
		})
	}

	// set all data maps to unassigned
	err := suite.DB.RawQuery("UPDATE data_maps SET status = ?", models.Unassigned).All(&[]models.DataMap{})
	suite.Nil(err)

	treasureChunk := models.DataMap{}
	err = suite.DB.RawQuery("SELECT * FROM data_maps WHERE chunk_idx = ?", 20).First(&treasureChunk)
	suite.Nil(err)

	// setting the address to something that the findTransactions mock can check for
	treasureChunk.Address = fakeFindTransactionsAddress
	suite.DB.ValidateAndSave(&treasureChunk)

	dataMaps := []models.DataMap{}
	err = suite.DB.RawQuery("SELECT * FROM data_maps ORDER BY chunk_idx ASC").All(&dataMaps)
	suite.Nil(err)
	suite.Equal(numChunks, len(dataMaps))

	// call method under test
	chunksToAttach, treasureChunks := jobs.HandleTreasureChunks(dataMaps, uploadSession1, IotaMock)

	for _, chunk := range chunksToAttach {
		suite.NotEqual(20, chunk.ChunkIdx)
	}

	suite.True(findTransactionsMockCalled_process_unassigned_chunks)
	suite.Equal(len(dataMaps)-2, len(chunksToAttach))
	suite.Equal(1, len(treasureChunks))
	suite.Equal(15, treasureChunks[0].ChunkIdx)
}

func (suite *JobsSuite) Test_InsertTreasureChunks_AlphaSession() {

	numChunks := 25

	uploadSession1 := models.UploadSession{
		GenesisHash:   "abcdeff1",
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
		GenesisHash:   "abcdeff1",
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

func (suite *JobsSuite) Test_SkipVerificationOfFirstChunks_Beta() {

	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	defer oyster_utils.ResetBrokerMode()

	numChunks := 29

	uploadSession := models.UploadSession{
		GenesisHash:   "abcdeff1",
		NumChunks:     numChunks,
		FileSizeBytes: 3000,
		Type:          models.SessionTypeBeta,
	}

	chunksReady, _, err := uploadSession.StartSessionAndWaitForChunks(500)
	suite.True(chunksReady)
	suite.Nil(err)

	dataMaps := []models.DataMap{}
	err = suite.DB.RawQuery("SELECT * FROM data_maps ORDER BY chunk_idx ASC").All(&dataMaps)
	suite.Nil(err)
	suite.Equal(numChunks+1, len(dataMaps))

	skipVerifyChunks, restOfChunks := jobs.SkipVerificationOfFirstChunks(dataMaps, uploadSession)

	var lenOfChunksToSkipVerifying int
	lenOfChunksToSkipVerifying = int((float64(numChunks + 1)) * (float64(jobs.PercentOfChunksToSkipVerification) /
		float64(100)))

	var lenOfRestOfChunks int
	lenOfRestOfChunks = numChunks + 1 - lenOfChunksToSkipVerifying

	suite.Equal(lenOfChunksToSkipVerifying,
		len(skipVerifyChunks))
	suite.Equal(numChunks+1-len(skipVerifyChunks), len(restOfChunks))

	var skipVerifyMinIdx int
	var skipVerifyMaxIdx int
	var restMinIdx int
	var restMaxIdx int

	skipVerifyMinIdx = numChunks - lenOfChunksToSkipVerifying
	skipVerifyMaxIdx = numChunks

	restMinIdx = 0
	restMaxIdx = lenOfRestOfChunks - 1

	for _, chunk := range skipVerifyChunks {
		suite.True(chunk.ChunkIdx >= skipVerifyMinIdx &&
			chunk.ChunkIdx <= skipVerifyMaxIdx)
	}
	for _, chunk := range restOfChunks {
		suite.True(chunk.ChunkIdx >= restMinIdx &&
			chunk.ChunkIdx <= restMaxIdx)
	}
}

func (suite *JobsSuite) Test_SkipVerificationOfFirstChunks_Alpha() {

	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	defer oyster_utils.ResetBrokerMode()

	numChunks := 29

	uploadSession := models.UploadSession{
		GenesisHash:   "abcdeff1",
		NumChunks:     numChunks,
		FileSizeBytes: 3000,
		Type:          models.SessionTypeAlpha,
	}

	chunksReady, _, err := uploadSession.StartSessionAndWaitForChunks(500)
	suite.True(chunksReady)
	suite.Nil(err)

	dataMaps := []models.DataMap{}
	err = suite.DB.RawQuery("SELECT * FROM data_maps ORDER BY chunk_idx ASC").All(&dataMaps)
	suite.Nil(err)
	suite.Equal(numChunks+1, len(dataMaps))

	skipVerifyChunks, restOfChunks := jobs.SkipVerificationOfFirstChunks(dataMaps, uploadSession)

	var lenOfChunksToSkipVerifying int
	lenOfChunksToSkipVerifying = int((float64(numChunks + 1)) * (float64(jobs.PercentOfChunksToSkipVerification) /
		float64(100)))

	suite.Equal(lenOfChunksToSkipVerifying,
		len(skipVerifyChunks))
	suite.Equal(numChunks+1-len(skipVerifyChunks), len(restOfChunks))

	var skipVerifyMinIdx int
	var skipVerifyMaxIdx int
	var restMinIdx int
	var restMaxIdx int

	skipVerifyMinIdx = 0
	skipVerifyMaxIdx = lenOfChunksToSkipVerifying - 1

	restMinIdx = lenOfChunksToSkipVerifying
	restMaxIdx = numChunks

	for _, chunk := range skipVerifyChunks {
		suite.True(chunk.ChunkIdx >= skipVerifyMinIdx &&
			chunk.ChunkIdx <= skipVerifyMaxIdx)
	}
	for _, chunk := range restOfChunks {
		suite.True(chunk.ChunkIdx >= restMinIdx &&
			chunk.ChunkIdx <= restMaxIdx)
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

	address, _ := giota.ToAddress(fakeFindTransactionsAddress)
	if addresses[0] == address {
		// only add to the map if the address is the address we decided to check for
		addrToTransactionMap[addresses[0]] = []giota.Transaction{}
	}

	findTransactionsMockCalled_process_unassigned_chunks = true

	return addrToTransactionMap, nil
}

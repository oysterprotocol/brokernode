package jobs_test

import (
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
	AllChunksCalled                                                      []oyster_utils.ChunkData
	fakeFindTransactionsAddress                                          = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
)

func (suite *JobsSuite) Done_ProcessUnassignedChunks() {

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
		GenesisHash:    oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		NumChunks:      numChunks,
		FileSizeBytes:  3000,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureInDataMapComplete,
		AllDataReady:   models.AllDataReady,
	}

	uploadSession2 := models.UploadSession{
		GenesisHash:    oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		NumChunks:      numChunks,
		FileSizeBytes:  3000,
		Type:           models.SessionTypeBeta,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureInDataMapComplete,
		AllDataReady:   models.AllDataReady,
	}

	uploadSession3 := models.UploadSession{
		GenesisHash:    oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		NumChunks:      numChunks,
		FileSizeBytes:  3000,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureInDataMapComplete,
		AllDataReady:   models.AllDataReady,
	}

	uploadSession4 := models.UploadSession{
		GenesisHash:    oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		NumChunks:      numChunks,
		FileSizeBytes:  3000,
		Type:           models.SessionTypeBeta,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureInDataMapComplete,
		AllDataReady:   models.AllDataReady,
	}

	SessionSetUpForTest(&uploadSession1, []int{15}, uploadSession1.NumChunks)
	SessionSetUpForTest(&uploadSession2, []int{15}, uploadSession2.NumChunks)
	SessionSetUpForTest(&uploadSession3, []int{15}, uploadSession3.NumChunks)
	SessionSetUpForTest(&uploadSession4, []int{15}, uploadSession4.NumChunks)

	// set uploadSession4 to be the oldest
	err := suite.DB.RawQuery("UPDATE upload_sessions SET created_at = ? WHERE genesis_hash = ?",
		time.Now().Add(-20*time.Second), uploadSession4.GenesisHash).All(&[]models.UploadSession{})
	suite.Nil(err)

	// set uploadSession2 to next oldest
	err = suite.DB.RawQuery("UPDATE upload_sessions SET created_at = ? WHERE genesis_hash = ?",
		time.Now().Add(-15*time.Second), uploadSession2.GenesisHash).All(&[]models.UploadSession{})
	suite.Nil(err)

	// set uploadSession1 to next oldest after uploadSession2
	err = suite.DB.RawQuery("UPDATE upload_sessions SET created_at = ? WHERE genesis_hash = ?",
		time.Now().Add(-10*time.Second), uploadSession1.GenesisHash).All(&[]models.UploadSession{})
	suite.Nil(err)

	// uploadSession3 will be the newest

	// call method under test
	jobs.ProcessUnassignedChunks(IotaMock, jobs.PrometheusWrapper)

	suite.True(sendChunksToChannelMockCalled_process_unassigned_chunks)
	suite.True(verifyChunkMessagesMatchesRecordMockCalled_process_unassigned_chunks)
	suite.Equal(4*(numChunks), len(AllChunksCalled))
	// Normally we'd expect this to be 4*(numChunks + 1), but we didn't bother to make the
	// message data for the treasure chunks

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
		genHashMapIdx[chunk.GenesisHash] = append(genHashMapIdx[chunk.GenesisHash], int(chunk.Idx))
		if _, ok := genHashMapOrder[chunk.GenesisHash]; !ok {
			genHashMapOrder[chunk.GenesisHash] = i
			i++
		}
	}

	suite.True(genHashMapIdx[uploadSession4.GenesisHash][0] >
		genHashMapIdx[uploadSession4.GenesisHash][len(genHashMapIdx[uploadSession4.GenesisHash])-1])
	suite.True(genHashMapIdx[uploadSession2.GenesisHash][0] >
		genHashMapIdx[uploadSession2.GenesisHash][len(genHashMapIdx[uploadSession2.GenesisHash])-1])
	suite.True(genHashMapIdx[uploadSession1.GenesisHash][0] <
		genHashMapIdx[uploadSession1.GenesisHash][len(genHashMapIdx[uploadSession1.GenesisHash])-1])
	suite.True(genHashMapIdx[uploadSession3.GenesisHash][0] <
		genHashMapIdx[uploadSession3.GenesisHash][len(genHashMapIdx[uploadSession3.GenesisHash])-1])

	suite.Equal(0, genHashMapOrder[uploadSession4.GenesisHash])
	suite.Equal(1, genHashMapOrder[uploadSession2.GenesisHash])
	suite.Equal(2, genHashMapOrder[uploadSession1.GenesisHash])
	suite.Equal(3, genHashMapOrder[uploadSession3.GenesisHash])
}

func (suite *JobsSuite) Done_HandleTreasureChunks_not_attached_yet() {

	numChunks := 25

	// make suite available inside mock methods
	Suite = *suite

	// assign the mock methods for this test
	makeMocks_process_unassigned_chunks(&IotaMock)

	// make 3 channels
	models.MakeChannels(3)

	uploadSession1 := models.UploadSession{
		GenesisHash:    oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		NumChunks:      numChunks,
		FileSizeBytes:  3000,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureInDataMapPending,
	}

	bulkChunkData := SessionSetUpForTest(&uploadSession1, []int{15}, uploadSession1.NumChunks)

	finishedHashes := uploadSession1.CheckIfAllHashesAreReady()
	finishedMessages := uploadSession1.CheckIfAllMessagesAreReady()
	suite.True(finishedHashes && finishedMessages)

	// call method under test
	chunksToAttach, treasureChunks := jobs.HandleTreasureChunks(bulkChunkData, uploadSession1, IotaMock)

	suite.True(findTransactionsMockCalled_process_unassigned_chunks)
	suite.Equal(len(bulkChunkData)-1, len(chunksToAttach))
	suite.Equal(1, len(treasureChunks))
	suite.Equal(int64(15), treasureChunks[0].Idx)
}

func (suite *JobsSuite) Done_HandleTreasureChunks_already_attached() {

	numChunks := 25

	// make suite available inside mock methods
	Suite = *suite

	// assign the mock methods for this test
	makeMocks_process_unassigned_chunks(&IotaMock)

	// make 3 channels
	models.MakeChannels(3)

	uploadSession1 := models.UploadSession{
		GenesisHash:    oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		NumChunks:      numChunks,
		FileSizeBytes:  3000,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureInDataMapPending,
	}

	bulkChunkData := SessionSetUpForTest(&uploadSession1, []int{15}, uploadSession1.NumChunks)

	chunkData20 := oyster_utils.GetChunkData(oyster_utils.InProgressDir, uploadSession1.GenesisHash,
		int64(20))

	suite.NotEqual("", chunkData20.Address)
	suite.NotEqual("", chunkData20.Message)

	// tell the findTransactions mock to check for this address
	fakeFindTransactionsAddress = chunkData20.Address

	// call method under test
	chunksToAttach, treasureChunks := jobs.HandleTreasureChunks(bulkChunkData, uploadSession1, IotaMock)

	for _, chunk := range chunksToAttach {
		suite.NotEqual(20, chunk.Idx)
	}

	suite.True(findTransactionsMockCalled_process_unassigned_chunks)
	suite.Equal(len(bulkChunkData)-1, len(chunksToAttach))
	suite.Equal(0, len(treasureChunks))
}

func (suite *JobsSuite) Done_InsertTreasureChunks_AlphaSession() {

	numChunks := 25

	uploadSession1 := models.UploadSession{
		GenesisHash:   oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		NumChunks:     numChunks,
		FileSizeBytes: 3000,
		Type:          models.SessionTypeAlpha,
	}

	bulkChunkData := SessionSetUpForTest(&uploadSession1, []int{15}, uploadSession1.NumChunks)

	nonTreasureChunks := []oyster_utils.ChunkData{}
	treasureChunks := []oyster_utils.ChunkData{}

	for i := 0; i < len(bulkChunkData); i++ {
		if bulkChunkData[i].Idx != 1 && bulkChunkData[i].Idx != 10 && bulkChunkData[i].Idx != 25 {
			nonTreasureChunks = append(nonTreasureChunks, bulkChunkData[i])
		} else {
			treasureChunks = append(treasureChunks, bulkChunkData[i])
		}
	}

	// call method under test
	allChunks := jobs.InsertTreasureChunks(nonTreasureChunks, treasureChunks, uploadSession1)

	suite.Equal(len(treasureChunks)+len(nonTreasureChunks), len(allChunks))

	// verify chunks are in the expected (ascending) order
	for i, chunk := range allChunks {
		if chunk.Idx == 1 {
			suite.Equal(int64(2), allChunks[i+1].Idx)
		}
		if chunk.Idx == 10 {
			suite.Equal(int64(9), allChunks[i-1].Idx)
			suite.Equal(int64(11), allChunks[i+1].Idx)
		}
		if chunk.Idx == 25 {
			suite.Equal(int64(24), allChunks[i-1].Idx)
		}
	}
}

func (suite *JobsSuite) Done_InsertTreasureChunks_BetaSession() {

	numChunks := 25

	uploadSession1 := models.UploadSession{
		GenesisHash:   oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		NumChunks:     numChunks,
		FileSizeBytes: 3000,
		Type:          models.SessionTypeBeta,
	}

	bulkChunkData := SessionSetUpForTest(&uploadSession1, []int{15}, uploadSession1.NumChunks)

	nonTreasureChunks := []oyster_utils.ChunkData{}
	treasureChunks := []oyster_utils.ChunkData{}

	for i := 0; i < len(bulkChunkData); i++ {
		if bulkChunkData[i].Idx != 1 && bulkChunkData[i].Idx != 10 && bulkChunkData[i].Idx != 25 {
			nonTreasureChunks = append(nonTreasureChunks, bulkChunkData[i])
		} else {
			treasureChunks = append(treasureChunks, bulkChunkData[i])
		}
	}

	// call method under test
	allChunks := jobs.InsertTreasureChunks(nonTreasureChunks, treasureChunks, uploadSession1)

	suite.Equal(len(treasureChunks)+len(nonTreasureChunks), len(allChunks))

	// verify chunks are in the expected (descending) order
	for i, chunk := range allChunks {
		if chunk.Idx == 1 {
			suite.Equal(int64(2), allChunks[i-1].Idx)
		}
		if chunk.Idx == 10 {
			suite.Equal(int64(9), allChunks[i+1].Idx)
			suite.Equal(int64(11), allChunks[i-1].Idx)
		}
		if chunk.Idx == 25 {
			suite.Equal(int64(24), allChunks[i+1].Idx)
		}
	}
}

func (suite *JobsSuite) Done_SkipVerificationOfFirstChunks_Beta() {

	// Running this in TestModeNoTreasure mode, so we will just expect numChunks
	// instead of numChunks + 1
	oyster_utils.SetBrokerMode(oyster_utils.TestModeNoTreasure)
	defer oyster_utils.ResetBrokerMode()

	numChunks := 29

	uploadSession := models.UploadSession{
		GenesisHash:   oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		NumChunks:     numChunks,
		FileSizeBytes: 3000,
		Type:          models.SessionTypeBeta,
	}

	bulkChunkData := SessionSetUpForTest(&uploadSession, []int{}, uploadSession.NumChunks)

	suite.Equal(numChunks, len(bulkChunkData))

	skipVerifyChunks, restOfChunks := jobs.SkipVerificationOfFirstChunks(bulkChunkData, uploadSession)

	var lenOfChunksToSkipVerifying int
	lenOfChunksToSkipVerifying = int((float64(numChunks + 1)) * (float64(jobs.PercentOfChunksToSkipVerification) /
		float64(100)))

	var lenOfRestOfChunks int
	lenOfRestOfChunks = numChunks + 1 - lenOfChunksToSkipVerifying

	suite.Equal(lenOfChunksToSkipVerifying,
		len(skipVerifyChunks))
	suite.Equal(numChunks-len(skipVerifyChunks), len(restOfChunks))

	var skipVerifyMinIdx int
	var skipVerifyMaxIdx int
	var restMinIdx int
	var restMaxIdx int

	skipVerifyMinIdx = numChunks - lenOfChunksToSkipVerifying
	skipVerifyMaxIdx = numChunks

	restMinIdx = 0
	restMaxIdx = lenOfRestOfChunks - 1

	for _, chunk := range skipVerifyChunks {
		suite.True(int(chunk.Idx) >= skipVerifyMinIdx &&
			int(chunk.Idx) <= skipVerifyMaxIdx)
	}
	for _, chunk := range restOfChunks {
		suite.True(int(chunk.Idx) >= restMinIdx &&
			int(chunk.Idx) <= restMaxIdx)
	}
}

func (suite *JobsSuite) Done_SkipVerificationOfFirstChunks_Alpha() {

	// Running this in TestModeNoTreasure mode, so we will just expect numChunks
	// instead of numChunks + 1
	oyster_utils.SetBrokerMode(oyster_utils.TestModeNoTreasure)
	defer oyster_utils.ResetBrokerMode()

	numChunks := 29

	uploadSession := models.UploadSession{
		GenesisHash:   oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		NumChunks:     numChunks,
		FileSizeBytes: 3000,
		Type:          models.SessionTypeAlpha,
	}

	bulkChunkData := SessionSetUpForTest(&uploadSession, []int{}, uploadSession.NumChunks)

	suite.Equal(numChunks, len(bulkChunkData))

	skipVerifyChunks, restOfChunks := jobs.SkipVerificationOfFirstChunks(bulkChunkData, uploadSession)

	var lenOfChunksToSkipVerifying int
	lenOfChunksToSkipVerifying = int((float64(numChunks + 1)) * (float64(jobs.PercentOfChunksToSkipVerification) /
		float64(100)))

	suite.Equal(lenOfChunksToSkipVerifying,
		len(skipVerifyChunks))
	suite.Equal(numChunks-len(skipVerifyChunks), len(restOfChunks))

	var skipVerifyMinIdx int
	var skipVerifyMaxIdx int
	var restMinIdx int
	var restMaxIdx int

	skipVerifyMinIdx = 0
	skipVerifyMaxIdx = lenOfChunksToSkipVerifying - 1

	restMinIdx = lenOfChunksToSkipVerifying
	restMaxIdx = numChunks

	for _, chunk := range skipVerifyChunks {
		suite.True(int(chunk.Idx) >= skipVerifyMinIdx &&
			int(chunk.Idx) <= skipVerifyMaxIdx)
	}
	for _, chunk := range restOfChunks {
		suite.True(int(chunk.Idx) >= restMinIdx &&
			int(chunk.Idx) <= restMaxIdx)
	}
}

func (suite *JobsSuite) Test_StageTreasures() {
	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	defer oyster_utils.ResetBrokerMode()

	u := models.UploadSession{
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        15000,
		NumChunks:            15,
		StorageLengthInYears: 1,
	}

	SessionSetUpForTest(&u, []int{5}, u.NumChunks)

	treasureIndexes, _ := u.GetTreasureIndexes()
	suite.Equal(1, len(treasureIndexes))

	chunkData := oyster_utils.GetChunkData(oyster_utils.InProgressDir, u.GenesisHash, int64(treasureIndexes[0]))

	jobs.StageTreasures([]oyster_utils.ChunkData{chunkData}, u)

	treasure := []models.Treasure{}

	suite.DB.Where("genesis_hash = ? AND address = ?", u.GenesisHash, chunkData.Address).All(&treasure)

	suite.Equal(1, len(treasure))
}

func makeMocks_process_unassigned_chunks(iotaMock *services.IotaService) {
	iotaMock.VerifyChunkMessagesMatchRecord = verifyChunkMessagesMatchesRecordMock_process_unassigned_chunks
	iotaMock.SendChunksToChannel = sendChunksToChannelMock_process_unassigned_chunks
	iotaMock.FindTransactions = findTransactions_process_unassigned_chunks
}

func sendChunksToChannelMock_process_unassigned_chunks(chunks []oyster_utils.ChunkData, channel *models.ChunkChannel) {

	// our mock was called
	sendChunksToChannelMockCalled_process_unassigned_chunks = true

	// stored all the chunks that get sent to the mock so we can run tests on them.
	AllChunksCalled = append(AllChunksCalled, chunks...)
}

func verifyChunkMessagesMatchesRecordMock_process_unassigned_chunks(chunks []oyster_utils.ChunkData) (filteredChunks services.FilteredChunk, err error) {

	verifyChunkMessagesMatchesRecordMockCalled_process_unassigned_chunks = true

	matchesTangle := []oyster_utils.ChunkData{}
	doesNotMatchTangle := []oyster_utils.ChunkData{}
	notAttached := []oyster_utils.ChunkData{}

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

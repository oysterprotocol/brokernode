package jobs_test

import (
	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"time"
)

var (
	sendChunksToChannelMockCalled_process_unassigned_chunks              = false
	verifyChunkMessagesMatchesRecordMockCalled_process_unassigned_chunks = false
	AllChunksCalled                                                      []models.DataMap
)

func (suite *JobsSuite) Test_ProcessUnassignedChunks() {

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
		NumChunks:      31,
		FileSizeBytes:  3000,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  models.PaymentStatusPaid,
		TreasureStatus: models.TreasureBuried,
	}

	uploadSession2 := models.UploadSession{
		GenesisHash:    "genHash2",
		NumChunks:      31,
		FileSizeBytes:  3000,
		Type:           models.SessionTypeBeta,
		PaymentStatus:  models.PaymentStatusPaid,
		TreasureStatus: models.TreasureBuried,
	}

	uploadSession3 := models.UploadSession{
		GenesisHash:    "genHash3",
		NumChunks:      31,
		FileSizeBytes:  3000,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  models.PaymentStatusPaid,
		TreasureStatus: models.TreasureBuried,
	}

	uploadSession4 := models.UploadSession{
		GenesisHash:    "genHash4",
		NumChunks:      31,
		FileSizeBytes:  3000,
		Type:           models.SessionTypeBeta,
		PaymentStatus:  models.PaymentStatusPaid,
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
	suite.Equal(124, len(AllChunksCalled))

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

func makeMocks_process_unassigned_chunks(iotaMock *services.IotaService) {
	iotaMock.VerifyChunkMessagesMatchRecord = verifyChunkMessagesMatchesRecordMock_process_unassigned_chunks
	iotaMock.SendChunksToChannel = sendChunksToChannelMock_process_unassigned_chunks
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

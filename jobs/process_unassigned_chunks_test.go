package jobs_test

import (
	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"time"
)

var (
	sendChunksToChannelMockCalled = false
	AllChunksCalled               []models.DataMap
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
		FileSizeBytes:  3000,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  models.PaymentStatusPaid,
		TreasureStatus: models.TreasureBuried,
	}

	uploadSession2 := models.UploadSession{
		GenesisHash:    "genHash2",
		FileSizeBytes:  3000,
		Type:           models.SessionTypeBeta,
		PaymentStatus:  models.PaymentStatusPaid,
		TreasureStatus: models.TreasureBuried,
	}

	uploadSession3 := models.UploadSession{
		GenesisHash:    "genHash3",
		FileSizeBytes:  25000,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  models.PaymentStatusPaid,
		TreasureStatus: models.TreasureBuried,
	}

	uploadSession4 := models.UploadSession{
		GenesisHash:    "genHash4",
		FileSizeBytes:  20000,
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

	suite.Equal(true, sendChunksToChannelMockCalled)
	suite.Equal(31, len(AllChunksCalled))

	/* This test is verifying that the chunks belonging to particular sessions were sent
	in the order we would expect and that the ordering of chunk ids within each data map was
	appropriate for alpha or beta.

		Session 4 was the oldest session and was of type beta
		Session 2 was next oldest and was of type beta
		Session 1 was next-to-last oldest and was of type alpha
		Session 3 was newest and was of type alpha


	If BundleSize is changed, these tests will need to be changed.
	*/
	for i, chunk := range AllChunksCalled {
		if i >= 0 && i < 10 {
			suite.Equal("genHash4", chunk.GenesisHash)
			suite.Equal(true, AllChunksCalled[i].ChunkIdx > AllChunksCalled[i+1].ChunkIdx)
		} else if i > 10 && i < 13 {
			suite.Equal("genHash2", chunk.GenesisHash)
			suite.Equal(true, AllChunksCalled[i].ChunkIdx > AllChunksCalled[i+1].ChunkIdx)
		} else if i > 13 && i < 16 {
			suite.Equal("genHash1", chunk.GenesisHash)
			suite.Equal(true, AllChunksCalled[i].ChunkIdx < AllChunksCalled[i+1].ChunkIdx)
		} else if i > 16 && i < 30 {
			suite.Equal("genHash3", chunk.GenesisHash)
			suite.Equal(true, AllChunksCalled[i].ChunkIdx < AllChunksCalled[i+1].ChunkIdx)
		}
	}
}

func makeMocks_process_unassigned_chunks(iotaMock *services.IotaService) {
	iotaMock.SendChunksToChannel = sendChunksToChannelMock_process_unassigned_chunks
}

func sendChunksToChannelMock_process_unassigned_chunks(chunks []models.DataMap, channel *models.ChunkChannel) {

	// our mock was called
	sendChunksToChannelMockCalled = true

	// stored all the chunks that get sent to the mock so we can run tests on them.
	AllChunksCalled = append(AllChunksCalled, chunks...)
}

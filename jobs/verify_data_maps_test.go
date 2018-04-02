package jobs_test

import (
	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
)

var (
	sendChunksToChannelMockCalled_verify              = false
	verifyChunkMessagesMatchesRecordMockCalled_verify = false
)

func (suite *JobsSuite) Test_VerifyDataMaps() {
	// reset back to generic mocks
	defer suite.SetupSuite()

	// make suite available inside mock methods
	Suite = *suite

	// assign the mock methods for this test
	makeMocks(&IotaMock)

	models.MakeChannels(3)

	// populate data_maps
	genHash := "someGenHash"
	fileBytesCount := 18000

	vErr, err := models.BuildDataMaps(genHash, fileBytesCount)
	suite.Nil(err)
	suite.Equal(0, len(vErr.Errors))

	// check that it is the length we expect
	allDataMaps := []models.DataMap{}
	err = suite.DB.All(&allDataMaps)
	suite.Equal(8, len(allDataMaps))

	// make first 6 data maps unverified
	for i := 0; i < 6; i++ {
		allDataMaps[i].Status = models.Unverified
		suite.DB.ValidateAndSave(&allDataMaps[i])
	}

	// call method under test, passing in our mock of our iota methods
	jobs.VerifyDataMaps(IotaMock)

	// verify the mock methods were called
	suite.Equal(true, sendChunksToChannelMockCalled_verify)
	suite.Equal(true, verifyChunkMessagesMatchesRecordMockCalled_verify)
}

func makeMocks(iotaMock *services.IotaService) {
	iotaMock.VerifyChunkMessagesMatchRecord = verifyChunkMessagesMatchesRecordMock
	iotaMock.SendChunksToChannel = sendChunksToChannelMock
}

func verifyChunkMessagesMatchesRecordMock(chunks []models.DataMap) (filteredChunks services.FilteredChunk, err error) {

	// our mock was called
	verifyChunkMessagesMatchesRecordMockCalled_verify = true

	allDataMaps := []models.DataMap{}
	err = Suite.DB.All(&allDataMaps)
	Suite.Nil(err)

	matchesTangle := []models.DataMap{}
	doesNotMatchTangle := []models.DataMap{}
	notAttached := []models.DataMap{}

	// assign some data_maps to different arrays to be returned when we filter the chunks
	matchesTangle = append(matchesTangle, allDataMaps[0], allDataMaps[3], allDataMaps[4], allDataMaps[5])
	notAttached = append(notAttached, allDataMaps[1])
	// the contents of 'doesNotMatchTangle' is what we will be checking for later
	doesNotMatchTangle = append(doesNotMatchTangle, allDataMaps[2])

	return services.FilteredChunk{
		MatchesTangle:      matchesTangle,
		NotAttached:        notAttached,
		DoesNotMatchTangle: doesNotMatchTangle,
	}, err
}

func sendChunksToChannelMock(chunks []models.DataMap, channel *models.ChunkChannel) {

	// our mock was called
	sendChunksToChannelMockCalled_verify = true

	allDataMaps := []models.DataMap{}
	var err = Suite.DB.All(&allDataMaps)
	Suite.Nil(err)

	// SendChunksToChannel should have gotten called with the index 2 data map from data_maps
	// make sure it's the same chunk
	Suite.Equal(chunks[0].ID, allDataMaps[2].ID)
	Suite.Equal(chunks[0].GenesisHash, allDataMaps[2].GenesisHash)
	Suite.Equal(chunks[0].GenesisHash, allDataMaps[2].GenesisHash)

	// status should have changed
	Suite.NotEqual(chunks[0].Status, allDataMaps[2].Status)
}

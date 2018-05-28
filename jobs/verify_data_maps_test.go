package jobs_test

import (
	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
)

var (
	verifyChunkMessagesMatchesRecordMockCalled_verify = false
)

func (suite *JobsSuite) Test_VerifyDataMaps() {
	// make suite available inside mock methods
	Suite = *suite

	// assign the mock methods for this test
	makeMocks(&IotaMock)

	models.MakeChannels(3)

	// populate data_maps
	genHash := "abcdef"
	numChunks := 10

	vErr, err := models.BuildDataMaps(genHash, numChunks)
	suite.Nil(err)
	suite.Equal(0, len(vErr.Errors))

	// check that it is the length we expect
	allDataMaps := []models.DataMap{}
	err = suite.DB.All(&allDataMaps)
	suite.Equal(numChunks+1, len(allDataMaps)) // 1  data map so 1 chunk has been added

	// make first 6 data maps unverified
	for i := 0; i < 6; i++ {
		allDataMaps[i].Status = models.Unverified
		suite.DB.ValidateAndSave(&allDataMaps[i])
	}

	// call method under test, passing in our mock of our iota methods
	jobs.VerifyDataMaps(IotaMock)

	// verify the mock methods were called
	suite.Equal(true, verifyChunkMessagesMatchesRecordMockCalled_verify)

	// verify that the data maps that didn't match the tangle were set to an error state
	allDataMaps = []models.DataMap{}
	err = suite.DB.Where("status = ?", models.Error).All(&allDataMaps)
	suite.Equal(1, len(allDataMaps))
}

func makeMocks(iotaMock *services.IotaService) {
	iotaMock.VerifyChunkMessagesMatchRecord = verifyChunkMessagesMatchesRecordMock
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

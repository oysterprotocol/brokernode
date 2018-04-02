package jobs_test

import (
	"github.com/gobuffalo/suite"
	"github.com/iotaledger/giota"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"testing"
)

var IotaMock services.IotaService

type JobsSuite struct {
	*suite.Model
}

var Suite JobsSuite
var sendChunksToChannelMockCalled = false
var verifyChunkMessagesMatchesRecordMockCalled = false

func (suite *JobsSuite) SetupSuite() {

	/*
		This creates a "generic" mock of our iota wrapper. we can assign
		different mocking functions in individual test files.
	*/

	IotaMock = services.IotaService{
		SendChunksToChannel: func(chunks []models.DataMap, channel *models.ChunkChannel) {

		},
		VerifyChunkMessagesMatchRecord: func(chunks []models.DataMap) (filteredChunks services.FilteredChunk, err error) {

			emptyChunkArray := []models.DataMap{}

			return services.FilteredChunk{
				MatchesTangle:      emptyChunkArray,
				NotAttached:        emptyChunkArray,
				DoesNotMatchTangle: emptyChunkArray,
			}, err
		},
		VerifyChunksMatchRecord: func(chunks []models.DataMap, checkChunkAndBranch bool) (filteredChunks services.FilteredChunk, err error) {
			emptyChunkArray := []models.DataMap{}

			return services.FilteredChunk{
				MatchesTangle:      emptyChunkArray,
				NotAttached:        emptyChunkArray,
				DoesNotMatchTangle: emptyChunkArray,
			}, err
		},
		ChunksMatch: func(chunkOnTangle giota.Transaction, chunkOnRecord models.DataMap, checkBranchAndTrunk bool) bool {
			return false
		},
	}
}

//
//func (suite *JobsSuite) TearDownSuite() {
//}
//
//func (suite *JobsSuite) SetupTest() {
//}
//
//func (suite *JobsSuite) TearDownTest() {
//}

func Test_JobsSuite(t *testing.T) {

	as := &JobsSuite{suite.NewModel()}
	suite.Run(t, as)
}

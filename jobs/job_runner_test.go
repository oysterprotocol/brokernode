package jobs_test

import (
	"testing"

	"github.com/gobuffalo/suite"
	"github.com/iotaledger/giota"
	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
)

var IotaMock services.IotaService

type JobsSuite struct {
	*suite.Model
}

var Suite JobsSuite

func (suite *JobsSuite) SetupTest() {
	suite.Model.SetupTest()

	if services.IsKvStoreEnabled() {
		suite.Nil(services.RemoveAllKvStoreData())
		suite.Nil(services.InitKvStore())
	}

	// Reset the jobs's IotaWrapper, EthWrapper before each test.
	// Some tests may override this value.
	jobs.IotaWrapper = services.IotaWrapper
	jobs.EthWrapper = services.EthWrapper
	jobs.PrometheusWrapper = services.PrometheusWrapper

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
		FindTransactions: func([]giota.Address) (map[giota.Address][]giota.Transaction, error) {
			return nil, nil
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
	oyster_utils.SetBrokerMode(oyster_utils.TestModeDummyTreasure)
	defer oyster_utils.ResetBrokerMode()
	as := &JobsSuite{suite.NewModel()}
	suite.Run(t, as)
}

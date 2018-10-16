package jobs_test

import (
	"math/rand"
	"strconv"
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
	suite.Nil(oyster_utils.InitKvStore())

	// Reset the jobs's IotaWrapper, EthWrapper before each test.
	// Some tests may override this value.
	jobs.IotaWrapper = services.IotaWrapper
	jobs.EthWrapper = oyster_utils.EthWrapper
	jobs.PrometheusWrapper = services.PrometheusWrapper

	/*
		This creates a "generic" mock of our iota wrapper. we can assign
		different mocking functions in individual test files.
	*/

	IotaMock = services.IotaService{
		SendChunksToChannel: func(chunks []oyster_utils.ChunkData, channel *models.ChunkChannel) {

		},
		VerifyChunkMessagesMatchRecord: func(chunks []oyster_utils.ChunkData) (filteredChunks services.FilteredChunk, err error) {

			emptyChunkArray := []oyster_utils.ChunkData{}

			return services.FilteredChunk{
				MatchesTangle:      emptyChunkArray,
				NotAttached:        emptyChunkArray,
				DoesNotMatchTangle: emptyChunkArray,
			}, err
		},
		VerifyChunksMatchRecord: func(chunks []oyster_utils.ChunkData, checkChunkAndBranch bool) (filteredChunks services.FilteredChunk, err error) {
			emptyChunkArray := []oyster_utils.ChunkData{}

			return services.FilteredChunk{
				MatchesTangle:      emptyChunkArray,
				NotAttached:        emptyChunkArray,
				DoesNotMatchTangle: emptyChunkArray,
			}, err
		},
		ChunksMatch: func(chunkOnTangle giota.Transaction, chunkOnRecord oyster_utils.ChunkData, checkBranchAndTrunk bool) bool {
			return false
		},
		FindTransactions: func([]giota.Address) (map[giota.Address][]giota.Transaction, error) {
			return nil, nil
		},
	}
}

func Test_JobsSuite(t *testing.T) {
	oyster_utils.SetBrokerMode(oyster_utils.TestModeDummyTreasure)
	defer oyster_utils.ResetBrokerMode()
	js := &JobsSuite{suite.NewModel()}
	suite.Run(t, js)
}

func GenerateChunkRequests(indexToStopAt int, genesisHash string) []models.ChunkReq {
	chunkReqs := []models.ChunkReq{}

	for i := 1; i <= indexToStopAt; i++ {

		asciiValue := ""
		for i := 0; i < 5; i++ {
			asciiValue += string(rand.Intn(255))
		}

		req := models.ChunkReq{
			Idx:  i,
			Hash: genesisHash,
			Data: asciiValue,
		}

		chunkReqs = append(chunkReqs, req)
	}
	return chunkReqs
}

func SessionSetUpForTest(session *models.UploadSession, mergedIndexes []int,
	indexToStopAt int) []oyster_utils.ChunkData {
	session.StartUploadSession()
	privateKeys := []string{}

	for i := 0; i < len(mergedIndexes); i++ {
		key := ""
		for j := 0; j < 9; j++ {
			key += strconv.Itoa(rand.Intn(8) + 1)
		}
		privateKeys = append(privateKeys, key+strconv.Itoa(i))
	}

	session.PaymentStatus = models.PaymentStatusConfirmed
	models.DB.ValidateAndUpdate(session)
	session.MakeTreasureIdxMap(mergedIndexes, privateKeys)

	chunkReqs := GenerateChunkRequests(indexToStopAt, session.GenesisHash)

	models.ProcessAndStoreChunkData(chunkReqs, session.GenesisHash, mergedIndexes, oyster_utils.TestValueTimeToLive)

	session.WaitForAllHashes(100)

	session.CreateTreasures()

	for {
		finishedMessages, _ := session.WaitForAllMessages(3)
		if finishedMessages {
			break
		}
	}

	bulkKeys := oyster_utils.GenerateBulkKeys(session.GenesisHash, models.MetaDataChunkIdx, int64(session.NumChunks))
	bulkChunkData, _ := models.GetMultiChunkData(oyster_utils.InProgressDir, session.GenesisHash,
		bulkKeys)

	return bulkChunkData
}

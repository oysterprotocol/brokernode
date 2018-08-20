package jobs_test

import (
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

	if oyster_utils.IsKvStoreEnabled() {
		suite.Nil(oyster_utils.RemoveAllKvStoreData())
		suite.Nil(oyster_utils.InitKvStore())
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
	as := &JobsSuite{suite.NewModel()}
	suite.Run(t, as)
}

func GenerateChunkRequests(numToGenerate int, genesisHash string) []models.ChunkReq {
	chunkReqs := []models.ChunkReq{}

	for i := 0; i < numToGenerate; i++ {

		trytes, _ := giota.ToTrytes(oyster_utils.RandSeq(10, oyster_utils.TrytesAlphabet))

		req := models.ChunkReq{
			Idx:  i,
			Hash: genesisHash,
			Data: string(trytes),
		}

		chunkReqs = append(chunkReqs, req)
	}
	return chunkReqs
}

func SessionSetUpForTest(session *models.UploadSession, mergedIndexes []int,
	numChunksToGenerate int) []oyster_utils.ChunkData {
	session.StartUploadSession()
	privateKeys := []string{}

	for i := 0; i < len(mergedIndexes); i++ {
		privateKeys = append(privateKeys, "100000000"+strconv.Itoa(i))
	}

	session.MakeTreasureIdxMap(mergedIndexes, privateKeys)

	chunkReqs := GenerateChunkRequests(numChunksToGenerate, session.GenesisHash)
	models.ProcessAndStoreChunkData(chunkReqs, session.GenesisHash, mergedIndexes, models.TestValueTimeToLive)

	for {
		jobs.BuryTreasureInDataMaps()
		finishedMessages, _ := session.WaitForAllMessages(3)
		if finishedMessages {
			break
		}
	}

	session.WaitForAllHashes(100)

	bulkKeys := oyster_utils.GenerateBulkKeys(session.GenesisHash, 0, int64(session.NumChunks-1))
	bulkChunkData, _ := oyster_utils.GetBulkChunkData(oyster_utils.InProgressDir, session.GenesisHash,
		bulkKeys)

	return bulkChunkData
}

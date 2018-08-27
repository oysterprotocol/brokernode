package models_test

import (
	"github.com/gobuffalo/suite"
	"github.com/iotaledger/giota"
	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/utils"
	"strconv"
	"testing"
)

type ModelSuite struct {
	*suite.Model
}

func Test_ModelSuite(t *testing.T) {
	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	defer oyster_utils.ResetBrokerMode()
	ms := &ModelSuite{suite.NewModel()}
	suite.Run(t, ms)
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

	session.PaymentStatus = models.PaymentStatusConfirmed
	models.DB.ValidateAndUpdate(session)
	session.MakeTreasureIdxMap(mergedIndexes, privateKeys)

	chunkReqs := GenerateChunkRequests(numChunksToGenerate, session.GenesisHash)

	models.ProcessAndStoreChunkData(chunkReqs, session.GenesisHash, mergedIndexes, oyster_utils.TestValueTimeToLive)

	for {
		jobs.BuryTreasureInDataMaps()
		finishedMessages, _ := session.WaitForAllMessages(3)
		if finishedMessages {
			break
		}
	}

	session.WaitForAllHashes(100)

	bulkKeys := oyster_utils.GenerateBulkKeys(session.GenesisHash, 0, int64(session.NumChunks-1))
	bulkChunkData, _ := models.GetMultiChunkData(oyster_utils.InProgressDir, session.GenesisHash,
		bulkKeys)

	return bulkChunkData
}

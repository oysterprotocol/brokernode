package models_test

import (
	"strconv"
	"testing"

	"github.com/gobuffalo/suite"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/utils"
	"math/rand"
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

func GenerateChunkRequestsForTests(indexToStopAt int, genesisHash string) []models.ChunkReq {
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

	chunkReqs := GenerateChunkRequestsForTests(indexToStopAt, session.GenesisHash)

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

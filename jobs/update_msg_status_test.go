package jobs_test

import (
	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
)

func (suite *JobsSuite) Test_GetActiveSessions_active_sessions() {
	u := models.UploadSession{
		GenesisHash:    "abcdeff1",
		NumChunks:      30,
		FileSizeBytes:  30000,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureInDataMapComplete,
	}

	u.StartUploadSession()

	sessions := jobs.GetActiveSessions()

	suite.Equal(1, len(sessions))
	suite.Equal("abcdeff1", sessions[0].GenesisHash)
}

func (suite *JobsSuite) Test_GetActiveSessions_no_active_sessions() {
	sessions := jobs.GetActiveSessions()

	suite.Equal(0, len(sessions))
}

func (suite *JobsSuite) Test_GetDataMapsToCheckForMessages_no_rows() {

	// Verify that our query will not pick up data maps with the wrong msg_status
	numChunks := 30

	u := models.UploadSession{
		GenesisHash:    "abcdeff1",
		NumChunks:      numChunks,
		FileSizeBytes:  30000,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureInDataMapComplete,
	}

	u.StartSessionAndWaitForChunks(500)
	mergedIndexes := []int{3, 12, 24}
	privateKeys := []string{"0000000001", "0000000002", "0000000003"}

	u.MakeTreasureIdxMap(mergedIndexes, privateKeys)
	u.EncryptTreasureIdxMapKeys()

	count, err := suite.DB.Where("genesis_hash = ?", u.GenesisHash).Count(&models.DataMap{})
	suite.Equal(numChunks+1, count)

	// set all data maps to a msg_status that will cause them to not get
	// picked up by GetDataMapsToCheckForMessages
	dataMaps := []models.DataMap{}

	err = suite.DB.RawQuery("UPDATE data_maps SET msg_status = ? "+
		"WHERE genesis_hash = ?",
		models.MsgStatusUploadedHaveNotEncoded, u.GenesisHash).All(&dataMaps)
	suite.Nil(err)

	sessions := jobs.GetActiveSessions()
	msgIdChunkMap := jobs.GetDataMapsToCheckForMessages(sessions[0])

	suite.Equal(0, len(msgIdChunkMap))
}

func (suite *JobsSuite) Test_GetDataMapsToCheckForMessages_skip_treasure_chunks() {

	// Verify that our query skips the treasure chunks
	defer oyster_utils.ResetBrokerMode()
	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)

	numChunks := 30

	u := models.UploadSession{
		GenesisHash:    "abcdeff1",
		NumChunks:      numChunks,
		FileSizeBytes:  30000,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureGeneratingKeys,
	}

	u.StartSessionAndWaitForChunks(500)
	mergedIndexes := []int{3, 12, 24}
	privateKeys := []string{"0000000001", "0000000002", "0000000003"}

	u.MakeTreasureIdxMap(mergedIndexes, privateKeys)
	u.EncryptTreasureIdxMapKeys()

	count, err := suite.DB.Where("genesis_hash = ?", u.GenesisHash).Count(&models.DataMap{})
	suite.Equal(numChunks+1, count)
	suite.Nil(err)

	sessions := jobs.GetActiveSessions()
	msgIdChunkMap := jobs.GetDataMapsToCheckForMessages(sessions[0])

	// check that treasure chunks have been omitted
	suite.Equal(numChunks-len(mergedIndexes)+1, len(msgIdChunkMap))

	for _, value := range msgIdChunkMap {
		suite.NotEqual(3, value.ChunkIdx)
		suite.NotEqual(12, value.ChunkIdx)
		suite.NotEqual(24, value.ChunkIdx)
	}
}

func (suite *JobsSuite) Test_GetDataMapsToCheckForMessages_match_session_genesis_hash() {
	// Verify that our query will only return data_maps from the session we have specified
	defer oyster_utils.ResetBrokerMode()
	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)

	numChunks := 30

	genesisHash := oyster_utils.RandSeq(6, []rune("abcde123456789"))
	u1 := models.UploadSession{
		GenesisHash:    genesisHash,
		NumChunks:      numChunks,
		FileSizeBytes:  30000,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureGeneratingKeys,
	}

	u1.StartSessionAndWaitForChunks(500)
	mergedIndexes := []int{3, 12, 24}
	privateKeys := []string{"0000000001", "0000000002", "0000000003"}

	u1.MakeTreasureIdxMap(mergedIndexes, privateKeys)
	u1.EncryptTreasureIdxMapKeys()

	genesisHash = oyster_utils.RandSeq(6, []rune("abcde123456789"))
	u2 := models.UploadSession{
		GenesisHash:    genesisHash,
		NumChunks:      numChunks,
		FileSizeBytes:  30000,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureGeneratingKeys,
	}

	u2.StartSessionAndWaitForChunks(500)
	mergedIndexes = []int{3, 12, 24}
	privateKeys = []string{"0000000001", "0000000002", "0000000003"}

	u2.MakeTreasureIdxMap(mergedIndexes, privateKeys)
	u2.EncryptTreasureIdxMapKeys()

	sessions := jobs.GetActiveSessions()
	msgIdChunkMap := jobs.GetDataMapsToCheckForMessages(sessions[0])

	// check that we only have one session's worth of data maps
	suite.Equal(numChunks-len(mergedIndexes)+1, len(msgIdChunkMap))

}

func (suite *JobsSuite) Test_CheckBadgerForKVPairs_no_kv_pairs() {

	// Verify that we don't return any K:V pairs when there aren't any
	defer oyster_utils.ResetBrokerMode()
	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)

	numChunks := 30

	u := models.UploadSession{
		GenesisHash:    "abcdeff1",
		NumChunks:      numChunks,
		FileSizeBytes:  30000,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureGeneratingKeys,
	}

	u.StartSessionAndWaitForChunks(500)
	mergedIndexes := []int{3, 12, 24}
	privateKeys := []string{"0000000001", "0000000002", "0000000003"}

	u.MakeTreasureIdxMap(mergedIndexes, privateKeys)
	u.EncryptTreasureIdxMapKeys()

	count, err := suite.DB.Where("genesis_hash = ?", u.GenesisHash).Count(&models.DataMap{})
	suite.Equal(numChunks+1, count)
	suite.Nil(err)

	sessions := jobs.GetActiveSessions()
	msgIdChunkMap := jobs.GetDataMapsToCheckForMessages(sessions[0])
	keyValuePairs, err := jobs.CheckBadgerForKVPairs(msgIdChunkMap)
	suite.Nil(err)

	// check that there are no K:V pairs
	suite.Equal(0, len(*keyValuePairs))
}

func (suite *JobsSuite) Test_CheckBadgerForKVPairs_some_kv_pairs() {
	// Verify that we can return some K:V pairs
	defer oyster_utils.ResetBrokerMode()
	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)

	numChunks := 30

	u := models.UploadSession{
		GenesisHash:    "abcdeff1",
		NumChunks:      numChunks,
		FileSizeBytes:  30000,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureGeneratingKeys,
	}

	u.StartSessionAndWaitForChunks(500)
	mergedIndexes := []int{3, 12, 24}
	privateKeys := []string{"0000000001", "0000000002", "0000000003"}

	u.MakeTreasureIdxMap(mergedIndexes, privateKeys)
	u.EncryptTreasureIdxMapKeys()

	dataMaps := []models.DataMap{}
	err := suite.DB.Where("genesis_hash = ?", u.GenesisHash).All(&dataMaps)
	suite.Equal(numChunks+1, len(dataMaps))
	suite.Nil(err)

	sessions := jobs.GetActiveSessions()
	msgIdChunkMap := jobs.GetDataMapsToCheckForMessages(sessions[0])

	// Give message data to some of the chunks
	countOfChunksWithMessages := 0
	batchSetKvMap := services.KVPairs{} // Store chunk.Data into KVStore
	for _, chunk := range msgIdChunkMap {
		if chunk.ChunkIdx < 15 {
			countOfChunksWithMessages++
			batchSetKvMap[chunk.MsgID] = "someData"
		}
	}
	services.BatchSet(&batchSetKvMap)

	keyValuePairs, err := jobs.CheckBadgerForKVPairs(msgIdChunkMap)
	suite.Nil(err)

	// check that there are an expected amount of K:V pairs
	suite.Equal(countOfChunksWithMessages, len(*keyValuePairs))
}

func (suite *JobsSuite) Test_UpdateMsgStatusForKVPairsFound() {
	// Verify that we can update the msg_status for the data_maps for which
	// we have found K:V pairs
	defer oyster_utils.ResetBrokerMode()
	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)

	numChunks := 30

	u := models.UploadSession{
		GenesisHash:    "abcdeff1",
		NumChunks:      numChunks,
		FileSizeBytes:  30000,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureGeneratingKeys,
	}

	u.StartSessionAndWaitForChunks(500)
	mergedIndexes := []int{3, 14, 18, 27}
	privateKeys := []string{"0000000001", "0000000002", "0000000003", "0000000004"}

	u.MakeTreasureIdxMap(mergedIndexes, privateKeys)
	u.EncryptTreasureIdxMapKeys()

	dataMaps := []models.DataMap{}
	err := suite.DB.Where("genesis_hash = ?", u.GenesisHash).All(&dataMaps)
	suite.Equal(numChunks+1, len(dataMaps))
	suite.Nil(err)

	sessions := jobs.GetActiveSessions()
	msgIdChunkMap := make(map[string]models.DataMap)
	msgIdChunkMap = jobs.GetDataMapsToCheckForMessages(sessions[0])
	// check that treasure chunks have been omitted
	suite.Equal(numChunks-len(mergedIndexes)+1, len(msgIdChunkMap))

	// Give message data to some of the chunks
	countOfChunksWithMessages := 0
	batchSetKvMap := services.KVPairs{} // Store chunk.Data into KVStore
	for key := range msgIdChunkMap {
		if msgIdChunkMap[key].ChunkIdx < 15 {
			countOfChunksWithMessages++
			batchSetKvMap[key] = oyster_utils.RandSeq(5, []rune("abcdefghijklmnopqrstuvwxyz"))
		}
	}

	err = services.BatchSet(&batchSetKvMap)
	suite.Nil(err)

	keyValuePairs, err := jobs.CheckBadgerForKVPairs(msgIdChunkMap)
	suite.Nil(err)
	jobs.UpdateMsgStatusForKVPairsFound(keyValuePairs, msgIdChunkMap)

	// verify that the data_maps that we added messages for now have their
	// msg_status updated
	dataMaps = []models.DataMap{}
	err = suite.DB.Where("genesis_hash = ?", u.GenesisHash).All(&dataMaps)
	suite.Equal(numChunks+1, len(dataMaps))
	suite.Nil(err)

	for _, dataMap := range dataMaps {
		if dataMap.ChunkIdx < 15 && dataMap.ChunkIdx != 3 &&
			dataMap.ChunkIdx != 14 && dataMap.ChunkIdx != 18 &&
			dataMap.ChunkIdx != 27 {
			suite.Equal(models.MsgStatusUploadedHaveNotEncoded, dataMap.MsgStatus)
		} else {
			suite.Equal(models.MsgStatusNotUploaded, dataMap.MsgStatus)
		}
	}
}

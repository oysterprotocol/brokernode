package jobs

import (
	"fmt"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
)

/*MsgIdChunkMap is a map of badger message Ids to their corresponding chunks*/
type MsgIdChunkMap map[string]models.DataMap

/*UpdateMsgStatus checks badger to verify that message data for particular chunks has arrived, and if so,
updates the msg_status field of those chunks*/
func UpdateMsgStatus(PrometheusWrapper services.PrometheusService) {

	activeSessions := GetActiveSessions()
	for _, session := range activeSessions {
		msgIdChunkMap := GetDataMapsWithToCheckForMessages(session)
		keyValuePairs, err := CheckBadgerForKVPairs(msgIdChunkMap)
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			continue
		}
		UpdateMsgStatusForKVPairsFound(keyValuePairs, msgIdChunkMap)
	}
}

func GetActiveSessions() []models.UploadSession {
	activeSessions := []models.UploadSession{}
	err := models.DB.RawQuery("SELECT * FROM upload_sessions " +
		"ORDER BY created_at ASC").All(&activeSessions)
	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return []models.UploadSession{}
	}
	return activeSessions
}

func GetDataMapsWithToCheckForMessages(session models.UploadSession) MsgIdChunkMap {

	dms := []models.DataMap{}
	treasureIndexes, _ := session.GetTreasureIndexes()
	treasureIndexInterface := make([]interface{}, len(treasureIndexes))
	for i, treasureIdx := range treasureIndexes {
		treasureIndexInterface[i] = treasureIdx
	}
	err := models.DB.Where(
		"chunk_idx NOT IN (?)", treasureIndexInterface...).Where(
		"genesis_hash = ?", session.GenesisHash).All(&dms)

	oyster_utils.LogIfError(err, nil)
	if err != nil || len(dms) <= 0 {
		return MsgIdChunkMap{}
	}
	return MakeMsgIdChunkMap(dms)
}

func MakeMsgIdChunkMap(chunks []models.DataMap) MsgIdChunkMap {
	msgIdChunkMap := make(map[string]models.DataMap)

	for _, chunk := range chunks {
		msgIdChunkMap[chunk.MsgID] = chunk
	}
	return msgIdChunkMap
}

func CheckBadgerForKVPairs(msgIdChunkMap MsgIdChunkMap) (kvs *services.KVPairs, err error) {
	var keys services.KVKeys

	for key := range msgIdChunkMap {
		keys = append(keys, key)
	}

	return services.BatchGet(&keys)
}

func UpdateMsgStatusForKVPairsFound(kvs *services.KVPairs, msgIdChunkMap MsgIdChunkMap) {
	if len(*kvs) <= 0 {
		return
	}
	readyChunks := []models.DataMap{}
	var updatedDms []string
	dbOperation, _ := oyster_utils.CreateDbUpdateOperation(&models.DataMap{})

	for key := range *kvs {
		chunk := msgIdChunkMap[key]
		chunk.MsgStatus = models.MsgStatusUploadedHaveNotEncoded
		readyChunks = append(readyChunks, chunk)
		updatedDms = append(updatedDms, fmt.Sprintf("(%s)", dbOperation.GetUpdatedValue(chunk)))
	}

	err := models.BatchUpsertDataMaps(updatedDms, dbOperation.GetColumns())
	oyster_utils.LogIfError(err, nil)
}

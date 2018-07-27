package jobs

import (
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"sort"
)

/*MsgIDChunkMap is a map of badger message Ids to their corresponding chunks*/
type MsgIDChunkMap map[string]models.DataMap

/*UpdateMsgStatus checks badger to verify that message data for particular chunks has arrived, and if so,
updates the msg_status field of those chunks*/
func UpdateMsgStatus(PrometheusWrapper services.PrometheusService) {

	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramUpdateMsgStatus, start)

	activeSessions := GetActiveSessions()
	for _, session := range activeSessions {
		msgIDChunkMap := GetDataMapsToCheckForMessages(session)
		keyValuePairs, err := CheckBadgerForKVPairs(msgIDChunkMap)
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			continue
		}
		UpdateMsgStatusForKVPairsFound(keyValuePairs, msgIDChunkMap)
	}
}

/*GetActiveSessions gets all the active sessions in order of creation*/
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

/*GetDataMapsToCheckForMessages gets all the data maps that have a status of MsgStatusNotUploaded
(excluding treasure chunks) so we can check if the message data is in badger*/
func GetDataMapsToCheckForMessages(session models.UploadSession) MsgIDChunkMap {

	dms := []models.DataMap{}
	var err error

	treasureIndexes, _ := session.GetTreasureIndexes()

	if len(treasureIndexes) > 0 {
		sort.Ints(treasureIndexes)
		dms, err = getDataMapsWithoutTreasures(session, treasureIndexes)

	} else {
		err = models.DB.Where(
			"genesis_hash = ? AND msg_status = ?",
			session.GenesisHash,
			models.MsgStatusNotUploaded).All(&dms)
	}

	oyster_utils.LogIfError(err, nil)
	if err != nil || len(dms) <= 0 {
		return MsgIDChunkMap{}
	}

	return MakeMsgIDChunkMap(dms)
}

/*MakeMsgIDChunkMap makes a map with MsgIDs as the keys and chunks as the values*/
func MakeMsgIDChunkMap(chunks []models.DataMap) MsgIDChunkMap {
	msgIDChunkMap := make(map[string]models.DataMap)

	for _, chunk := range chunks {
		msgIDChunkMap[chunk.MsgID] = chunk
	}
	return msgIDChunkMap
}

/*CheckBadgerForKVPairs checks badger for K:V pairs for the data_map rows*/
func CheckBadgerForKVPairs(msgIDChunkMap MsgIDChunkMap) (kvs *services.KVPairs, err error) {
	if len(msgIDChunkMap) <= 0 {
		return &services.KVPairs{}, nil
	}

	var keys services.KVKeys

	for key := range msgIDChunkMap {
		keys = append(keys, key)
	}

	return services.BatchGet(&keys)
}

/*UpdateMsgStatusForKVPairsFound will update the msg_status of data_maps that we found message data for*/
func UpdateMsgStatusForKVPairsFound(kvs *services.KVPairs, msgIDChunkMap MsgIDChunkMap) {
	if len(*kvs) <= 0 {
		return
	}
	readyChunks := []models.DataMap{}
	var updatedDms []string
	dbOperation, _ := oyster_utils.CreateDbUpdateOperation(&models.DataMap{})

	for key := range *kvs {
		chunk := msgIDChunkMap[key]
		chunk.MsgStatus = models.MsgStatusUploadedHaveNotEncoded
		readyChunks = append(readyChunks, chunk)
		updatedDms = append(updatedDms, dbOperation.GetUpdatedValue(chunk))
	}

	err := models.BatchUpsert(
		"data_maps",
		updatedDms,
		dbOperation.GetColumns(),
		[]string{"message", "status", "updated_at", "msg_status"})

	oyster_utils.LogIfError(err, nil)
}

func getDataMapsWithoutTreasures(session models.UploadSession, treasureIndexes []int) ([]models.DataMap, error) {

	dms := []models.DataMap{}
	tempDataMaps := []models.DataMap{}

	err := models.DB.RawQuery("SELECT * FROM data_maps WHERE "+
		"chunk_idx < ? AND genesis_hash = ? AND msg_status = ?",
		treasureIndexes[0],
		session.GenesisHash,
		models.MsgStatusNotUploaded).All(&tempDataMaps)
	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return []models.DataMap{}, err
	}
	dms = append(dms, tempDataMaps...)

	if len(treasureIndexes) > 1 {
		for i := 0; i < len(treasureIndexes)-1; i++ {
			tempDataMaps = []models.DataMap{}

			err = models.DB.RawQuery("SELECT * FROM data_maps WHERE "+
				"chunk_idx > ? AND chunk_idx < ? AND genesis_hash = ? AND msg_status = ?",
				treasureIndexes[i],
				treasureIndexes[i+1],
				session.GenesisHash,
				models.MsgStatusNotUploaded).All(&tempDataMaps)
			if err != nil {
				oyster_utils.LogIfError(err, nil)
				return []models.DataMap{}, err
			}
			dms = append(dms, tempDataMaps...)
		}
	}

	tempDataMaps = []models.DataMap{}
	err = models.DB.RawQuery("SELECT * FROM data_maps WHERE "+
		"chunk_idx > ? AND genesis_hash = ? AND msg_status = ?",
		treasureIndexes[len(treasureIndexes)-1],
		session.GenesisHash,
		models.MsgStatusNotUploaded).All(&tempDataMaps)
	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return []models.DataMap{}, err
	}
	dms = append(dms, tempDataMaps...)

	return dms, nil
}

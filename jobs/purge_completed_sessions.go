package jobs

import (
	"errors"
	"sync"

	"github.com/gobuffalo/pop"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"gopkg.in/segmentio/analytics-go.v3"
)

var purgeMutex = &sync.Mutex{}

func PurgeCompletedSessions(PrometheusWrapper services.PrometheusService) {

	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramPurgeCompletedSessions, start)

	purgeMutex.Lock()
	defer purgeMutex.Unlock()

	completeGenesisHashes, err := getAllCompletedGenesisHashes()
	oyster_utils.LogIfError(err, nil)

	for _, genesisHash := range completeGenesisHashes {
		var moveToCompleteDm = []models.DataMap{}

		err := models.DB.RawQuery("SELECT * FROM data_maps WHERE genesis_hash = ?", genesisHash).All(&moveToCompleteDm)
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			continue
		}

		err = models.DB.Transaction(func(tx *pop.Connection) error {
			// Passed in the connection
			if err := moveToComplete(tx, moveToCompleteDm); err != nil {
				return err
			}

			if err := tx.RawQuery("DELETE FROM data_maps WHERE genesis_hash = ?", genesisHash).All(&[]models.DataMap{}); err != nil {
				oyster_utils.LogIfError(err, nil)
				return err
			}

			session := []models.UploadSession{}
			if err := tx.RawQuery("SELECT * FROM upload_sessions WHERE genesis_hash = ?", genesisHash).All(&session); err != nil {
				oyster_utils.LogIfError(err, nil)
				return err
			}

			if len(session) > 0 {
				if err := models.NewCompletedUpload(session[0]); err != nil {
					oyster_utils.LogIfError(err, nil)
					return err
				}
			}

			if err := tx.RawQuery("DELETE FROM upload_sessions WHERE genesis_hash = ?", genesisHash).All(&[]models.UploadSession{}); err != nil {
				oyster_utils.LogIfError(err, nil)
				return err
			}

			oyster_utils.LogToSegment("purge_completed_sessions: completed_session_purged", analytics.NewProperties().
				Set("genesis_hash", genesisHash).
				Set("session_id", session[0].ID))

			return nil
		})
		if err == nil && services.IsKvStoreEnabled() {
			services.DeleteMsgDatas(moveToCompleteDm)
		}
	}
}

func getAllCompletedGenesisHashes() ([]string, error) {
	var genesisHashesNotComplete = []models.DataMap{}
	var allGenesisHashesStruct = []models.DataMap{}
	var completeGenesisHashes []string

	err := models.DB.RawQuery("SELECT distinct genesis_hash FROM data_maps").All(&allGenesisHashesStruct)
	oyster_utils.LogIfError(err, nil)
	if err != nil {
		return completeGenesisHashes, err
	}

	allGenesisHashes := make([]string, 0, len(allGenesisHashesStruct))

	for _, genesisHash := range allGenesisHashesStruct {
		allGenesisHashes = append(allGenesisHashes, genesisHash.GenesisHash)
	}

	err = models.DB.RawQuery("SELECT distinct genesis_hash FROM data_maps WHERE status != ? AND status != ?",
		models.Complete,
		models.Confirmed).All(&genesisHashesNotComplete)
	oyster_utils.LogIfError(err, nil)
	if err != nil {
		return completeGenesisHashes, err
	}

	notCompleteMap := map[string]bool{}

	for _, genesisHash := range genesisHashesNotComplete {
		notCompleteMap[genesisHash.GenesisHash] = true
	}

	for _, genesisHash := range allGenesisHashes {
		if _, hasKey := notCompleteMap[genesisHash]; !hasKey {
			completeGenesisHashes = append(completeGenesisHashes, genesisHash)
		}
	}

	return completeGenesisHashes, nil
}

func moveToComplete(tx *pop.Connection, dataMaps []models.DataMap) error {
	index := 0
	messagsKvPairs := services.KVPairs{}

	// TODO this is awful, do something different

	for _, dataMap := range dataMaps {
		completedDataMap := models.CompletedDataMap{
			Status:      dataMap.Status,
			NodeID:      dataMap.NodeID,
			NodeType:    dataMap.NodeType,
			TrunkTx:     dataMap.TrunkTx,
			BranchTx:    dataMap.BranchTx,
			GenesisHash: dataMap.GenesisHash,
			ChunkIdx:    dataMap.ChunkIdx,
			Hash:        dataMap.Hash,
			Address:     dataMap.Address,
			MsgStatus:   dataMap.MsgStatus,
		}
		if !services.IsKvStoreEnabled() {
			completedDataMap.Message = services.GetMessageFromDataMap(dataMap)
		}

		dupeCheck := []models.CompletedDataMap{}
		err := tx.RawQuery("SELECT * from completed_data_maps WHERE address = ? AND genesis_hash =?",
			dataMap.Address, dataMap.GenesisHash).All(&dupeCheck)

		if len(dupeCheck) > 0 {
			continue
		}

		vErr, err := tx.ValidateAndSave(&completedDataMap)
		if err == nil && !vErr.HasAny() {
			// Force GetMessageFromDataMap to return un-encoded msg.
			msgStatus := dataMap.MsgStatus
			if dataMap.MsgStatus == models.MsgStatusUploadedHaveNotEncoded {
				dataMap.MsgStatus = models.MsgStatusUploadedNoNeedEncode
			}
			messagsKvPairs[completedDataMap.MsgID] = services.GetMessageFromDataMap(dataMap)
			dataMap.MsgStatus = msgStatus
		} else {
			if vErr.HasAny() {
				oyster_utils.LogIfValidationError("CompletedDataMap validation failed", vErr, map[string]interface{}{
					"numOfDataMaps":  len(dataMaps),
					"proceededIndex": index,
				})
				return errors.New("Unable to create CompletedDataMap")
			}

			oyster_utils.LogIfError(err, map[string]interface{}{
				"numOfDataMaps":  len(dataMaps),
				"proceededIndex": index,
			})
			return err
		}
		index++
	}

	if services.IsKvStoreEnabled() {
		services.BatchSet(&messagsKvPairs, models.CompletedDataMapsTimeToLive)
	}
	return nil
}

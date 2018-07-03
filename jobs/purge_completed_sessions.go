package jobs

import (
	"errors"

	"github.com/gobuffalo/pop"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"gopkg.in/segmentio/analytics-go.v3"
)

func PurgeCompletedSessions(PrometheusWrapper services.PrometheusService) {

	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramPurgeCompletedSessions, start)

	var genesisHashesNotComplete = []models.DataMap{}
	var allGenesisHashesStruct = []models.DataMap{}

	err := models.DB.RawQuery("SELECT distinct genesis_hash FROM data_maps").All(&allGenesisHashesStruct)
	oyster_utils.LogIfError(err, nil)

	allGenesisHashes := make([]string, 0, len(allGenesisHashesStruct))

	for _, genesisHash := range allGenesisHashesStruct {
		allGenesisHashes = append(allGenesisHashes, genesisHash.GenesisHash)
	}

	err = models.DB.RawQuery("SELECT distinct genesis_hash FROM data_maps WHERE status != ? AND status != ?",
		models.Complete,
		models.Confirmed).All(&genesisHashesNotComplete)
	oyster_utils.LogIfError(err, nil)

	notComplete := map[string]bool{}

	for _, genesisHash := range genesisHashesNotComplete {
		notComplete[genesisHash.GenesisHash] = true
	}

	for _, genesisHash := range allGenesisHashes {
		if _, hasKey := notComplete[genesisHash]; !hasKey {
			var moveToCompleteDm = []models.DataMap{}

			err := models.DB.RawQuery("SELECT * from data_maps WHERE genesis_hash = ?", genesisHash).All(&moveToCompleteDm)
			if err != nil {
				oyster_utils.LogIfError(err, nil)
				continue
			}

			err = models.DB.Transaction(func(tx *pop.Connection) error {
				// Passed in the connection
				if err := moveToComplete(tx, moveToCompleteDm); err != nil {
					return err
				}

				if err := tx.RawQuery("DELETE from data_maps WHERE genesis_hash = ?", genesisHash).All(&[]models.DataMap{}); err != nil {
					oyster_utils.LogIfError(err, nil)
					return err
				}

				session := []models.UploadSession{}
				if err := tx.RawQuery("SELECT * from upload_sessions WHERE genesis_hash = ?", genesisHash).All(&session); err != nil {
					oyster_utils.LogIfError(err, nil)
					return err
				}

				if len(session) > 0 {
					vErr, err := tx.ValidateAndSave(&models.StoredGenesisHash{
						GenesisHash:   session[0].GenesisHash,
						NumChunks:     session[0].NumChunks,
						FileSizeBytes: session[0].FileSizeBytes,
					})
					if vErr.HasAny() {
						oyster_utils.LogIfValidationError("StoredGenesisHash validation failed.", vErr, nil)
						return errors.New("Unable to validate StoredGenesisHash")
					}
					if err != nil {
						oyster_utils.LogIfError(err, nil)
						return err
					}
					if err := models.NewCompletedUpload(session[0]); err != nil {
						return err
					}
				}

				if err := tx.RawQuery("DELETE from upload_sessions WHERE genesis_hash = ?", genesisHash).All(&[]models.UploadSession{}); err != nil {
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
}

func moveToComplete(tx *pop.Connection, dataMaps []models.DataMap) error {
	index := 0
	messagsKvPairs := services.KVPairs{}
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
		}
		if !services.IsKvStoreEnabled() {
			completedDataMap.Message = services.GetMessageFromDataMap(dataMap)
		}

		vErr, err := tx.ValidateAndSave(&completedDataMap)
		if err == nil && !vErr.HasAny() {
			messagsKvPairs[completedDataMap.MsgID] = services.GetMessageFromDataMap(dataMap)
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
		services.BatchSet(&messagsKvPairs)
	}
	return nil
}

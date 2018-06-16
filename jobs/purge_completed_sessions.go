package jobs

import (
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
			var moveToComplete = []models.DataMap{}

			err := models.DB.RawQuery("SELECT * from data_maps WHERE genesis_hash = ?", genesisHash).All(&moveToComplete)
			if err != nil {
				oyster_utils.LogIfError(err, nil)
				continue
			}

			models.DB.Transaction(func(tx *pop.Connection) error {
				MoveToComplete(tx, moveToComplete) // Passed in the connection

				err = tx.RawQuery("DELETE from data_maps WHERE genesis_hash = ?", genesisHash).All(&[]models.DataMap{})
				if err != nil {
					oyster_utils.LogIfError(err, nil)
					return err
				}

				session := []models.UploadSession{}

				err = tx.RawQuery("SELECT * from upload_sessions WHERE genesis_hash = ?", genesisHash).All(&session)
				if err != nil {
					oyster_utils.LogIfError(err, nil)
					return err
				}

				if len(session) > 0 {
					_, err = tx.ValidateAndSave(&models.StoredGenesisHash{
						GenesisHash:   session[0].GenesisHash,
						NumChunks:     session[0].NumChunks,
						FileSizeBytes: session[0].FileSizeBytes,
					})
					if err != nil {
						oyster_utils.LogIfError(err, nil)
						return err
					}
					err = models.NewCompletedUpload(session[0])
					if err != nil {
						return err
					}
				}

				err = tx.RawQuery("DELETE from upload_sessions WHERE genesis_hash = ?", genesisHash).All(&[]models.UploadSession{})
				if err != nil {
					oyster_utils.LogIfError(err, nil)
					return err
				}

				oyster_utils.LogToSegment("purge_completed_sessions: completed_session_purged", analytics.NewProperties().
					Set("genesis_hash", genesisHash).
					Set("session_id", session[0].ID))

				return nil
			})
			services.DeleteMsgDatas(moveToComplete)
		}
	}
}

func MoveToComplete(tx *pop.Connection, dataMaps []models.DataMap) {
	index := 0
	for _, dataMap := range dataMaps {
		completedDataMap := models.CompletedDataMap{
			Status:      dataMap.Status,
			Message:     services.GetMessageFromDataMap(dataMap),
			NodeID:      dataMap.NodeID,
			NodeType:    dataMap.NodeType,
			TrunkTx:     dataMap.TrunkTx,
			BranchTx:    dataMap.BranchTx,
			GenesisHash: dataMap.GenesisHash,
			ChunkIdx:    dataMap.ChunkIdx,
			Hash:        dataMap.Hash,
			Address:     dataMap.Address,
		}

		_, err := tx.ValidateAndSave(&completedDataMap)
		oyster_utils.LogIfError(err, map[string]interface{}{
			"numOfDataMaps":  len(dataMaps),
			"proceededIndex": index,
		})
		index++
	}
}

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
	if err != nil {
		oyster_utils.LogIfError(errors.New(err.Error()+" getting the completeGenesisHashes in "+
			"purge_completed_sessions"), nil)
	}

	for _, genesisHash := range completeGenesisHashes {
		if err := createCompletedDataMapIfNeeded(genesisHash); err != nil {
			continue
		}

		var moveToCompleteDm = []models.DataMap{}
		err := models.DB.RawQuery("SELECT * FROM data_maps WHERE genesis_hash = ?", genesisHash).All(&moveToCompleteDm)
		if err != nil {
			oyster_utils.LogIfError(errors.New(err.Error()+" getting the data_maps that match genesis_hash in "+
				"purge_completed_sessions"), nil)
			continue
		}

		if err := moveToComplete(moveToCompleteDm); err != nil {
			continue
		}

		err = models.DB.Transaction(func(tx *pop.Connection) error {
			if err := tx.RawQuery("DELETE FROM data_maps WHERE genesis_hash = ?", genesisHash).All(&[]models.DataMap{}); err != nil {
				oyster_utils.LogIfError(err, nil)
				return err
			}

			if err := tx.RawQuery("DELETE FROM upload_sessions WHERE genesis_hash = ?", genesisHash).All(&[]models.UploadSession{}); err != nil {
				oyster_utils.LogIfError(errors.New(err.Error()+" while deleting upload_sessions in "+
					"purge_completed_sessions"), nil)
				return err
			}
			return nil
		})
		oyster_utils.LogToSegment("purge_completed_sessions: completed_session_purged", analytics.NewProperties().
			Set("genesis_hash", genesisHash))

		if err != nil {
			continue
		}

		services.DeleteMsgDatas(moveToCompleteDm)
	}
}

func getAllCompletedGenesisHashes() ([]string, error) {
	var genesisHashesNotComplete = []models.DataMap{}
	var allGenesisHashesStruct = []models.DataMap{}
	var completeGenesisHashes []string

	err := models.DB.RawQuery("SELECT distinct genesis_hash FROM data_maps").All(&allGenesisHashesStruct)
	if err != nil {
		oyster_utils.LogIfError(errors.New(err.Error()+" while getting distinct genesis_hash from data_maps in "+
			"purge_completed_sessions"), nil)
		return completeGenesisHashes, err
	}

	allGenesisHashes := make([]string, 0, len(allGenesisHashesStruct))

	for _, genesisHash := range allGenesisHashesStruct {
		allGenesisHashes = append(allGenesisHashes, genesisHash.GenesisHash)
	}

	err = models.DB.RawQuery("SELECT distinct genesis_hash FROM data_maps WHERE status != ? AND status != ?",
		models.Complete,
		models.Confirmed).All(&genesisHashesNotComplete)
	if err != nil {
		oyster_utils.LogIfError(errors.New(err.Error()+" while getting distinct genesis_hash from data_maps "+
			"where status is Completed or Confirmed in purge_completed_sessions"), nil)
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

func createCompletedDataMapIfNeeded(genesisHash string) error {
	completedUploadedSession := []models.CompletedUpload{}
	if err := models.DB.RawQuery("SELECT * FROM completed_uploads WHERE genesis_hash = ?", genesisHash).All(&completedUploadedSession); err != nil {
		oyster_utils.LogIfError(err, nil)
		return err
	}
	if len(completedUploadedSession) > 0 {
		return nil
	}

	session := []models.UploadSession{}
	if err := models.DB.RawQuery("SELECT * FROM upload_sessions WHERE genesis_hash = ?", genesisHash).All(&session); err != nil {
		oyster_utils.LogIfError(err, nil)
		return err
	}

	if len(session) == 0 {
		return nil
	}

	err := models.NewCompletedUpload(session[0])
	oyster_utils.LogIfError(err, nil)
	return err
}

func moveToComplete(dataMaps []models.DataMap) error {
	if len(dataMaps) == 0 {
		return nil
	}

	existedDataMaps := []models.CompletedDataMap{}
	models.DB.RawQuery("SELECT address FROM completed_data_maps WHERE genesis_hash =?", dataMaps[0].GenesisHash).All(&existedDataMaps)
	existedMap := make(map[string]bool)
	for _, dm := range existedDataMaps {
		existedMap[dm.Address] = true
	}

	messagsKvPairs := services.KVPairs{}
	var upsertedValues []string
	dbOperation, _ := oyster_utils.CreateDbUpdateOperation(&models.CompletedDataMap{})
	hasValidationError := false

	for _, dataMap := range dataMaps {
		if _, hasKey := existedMap[dataMap.Address]; hasKey {
			continue
		}

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
			MsgID:       oyster_utils.GenerateBadgerKey(models.CompletedDataMapsMsgIDPrefix, dataMap.GenesisHash, dataMap.ChunkIdx),
		}
		completedDataMap.Message = services.GetMessageFromDataMap(dataMap)

		if vErr, err := completedDataMap.Validate(nil); err != nil || vErr.HasAny() {
			oyster_utils.LogIfValidationError("CompletedDataMap validation failed", vErr, nil)
			oyster_utils.LogIfError(err, nil)
			hasValidationError = true
			continue
		}

		// Force GetMessageFromDataMap to return un-encoded msg.
		msgStatus := dataMap.MsgStatus
		if dataMap.MsgStatus == models.MsgStatusUploadedHaveNotEncoded {
			dataMap.MsgStatus = models.MsgStatusUploadedNoNeedEncode
		}
		messagsKvPairs[completedDataMap.MsgID] = services.GetMessageFromDataMap(dataMap)
		dataMap.MsgStatus = msgStatus

		upsertedValues = append(upsertedValues, dbOperation.GetNewInsertedValue(completedDataMap))
	}

	err := models.BatchUpsert("completed_data_maps", upsertedValues, dbOperation.GetColumns(), nil)
	services.BatchSet(&messagsKvPairs, models.CompletedDataMapsTimeToLive)
	if hasValidationError {
		return errors.New("Partial update failed")
	}
	return err
}

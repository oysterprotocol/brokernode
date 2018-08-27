package jobs

import (
	"errors"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"gopkg.in/segmentio/analytics-go.v3"
)

func ProcessPaidSessions(PrometheusWrapper services.PrometheusService) {

	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramProcessPaidSessions, start)

	BuryTreasureInDataMaps()
}

func BuryTreasureInDataMaps() error {

	unburiedSessions, err := models.GetSessionsThatNeedTreasure()

	if err != nil {
		oyster_utils.LogIfError(err, nil)
	}

	for _, unburiedSession := range unburiedSessions {

		treasureIndex, err := unburiedSession.GetTreasureMap()

		if err != nil {
			oyster_utils.LogIfError(err, nil)
			return err
		}

		BuryTreasure(treasureIndex, &unburiedSession)
	}
	return nil
}

func BuryTreasure(treasureIndexMap []models.TreasureMap, unburiedSession *models.UploadSession) error {

	for _, entry := range treasureIndexMap {

		treasureChunk := models.GetSingleChunkData(oyster_utils.InProgressDir, unburiedSession.GenesisHash, int64(entry.Idx))

		if treasureChunk.Address == "" || treasureChunk.Hash == "" {
			errString := "did not find a chunk that matched genesis_hash and chunk_idx in process_paid_sessions, or " +
				"found duplicate chunks"
			err := errors.New(errString)
			oyster_utils.LogIfError(errors.New(err.Error()+" in BuryTreasure() in process_paid_sessions"),
				map[string]interface{}{
					"entry.Id":    entry.Idx,
					"genesisHash": unburiedSession.GenesisHash,
					"method":      "BuryTreasure() in process_paid_sessions",
				})
			return err
		}

		decryptedEthKey, err := unburiedSession.DecryptTreasureChunkEthKey(entry.Key)
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			return err
		}

		message, err := models.CreateTreasurePayload(decryptedEthKey, treasureChunk.Hash, models.MaxSideChainLength)
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			return err
		}

		err = unburiedSession.SetTreasureMessage(entry.Idx, message, models.DataMapsTimeToLive)
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			return err
		}

		oyster_utils.LogToSegment("process_paid_sessions: treasure_payload_buried_in_data_map", analytics.NewProperties().
			Set("genesis_hash", unburiedSession.GenesisHash).
			Set("sector", entry.Sector).
			Set("chunk_idx", entry.Idx).
			Set("message", message))
	}

	unburiedSession.TreasureStatus = models.TreasureInDataMapComplete

	if unburiedSession.CheckIfAllDataIsReady() && unburiedSession.AllDataReady != models.AllDataReady {
		unburiedSession.AllDataReady = models.AllDataReady
	}

	vErr, err := models.DB.ValidateAndUpdate(unburiedSession)
	oyster_utils.LogIfError(err, nil)
	oyster_utils.LogIfValidationError("", vErr, nil)

	return nil
}

package jobs

import (
	"errors"
	"github.com/getsentry/raven-go"
	"github.com/oysterprotocol/brokernode/models"
	"log"
)

func init() {
}

func ProcessPaidSessions() {

	BuryTreasureInDataMaps()
	MarkBuriedMapsAsUnassigned()
}

func BuryTreasureInDataMaps() error {

	unburiedSessions, err := models.GetSessionsThatNeedTreasure()

	if err != nil {
		raven.CaptureError(err, nil)
	}

	for _, unburiedSession := range unburiedSessions {

		treasureIndex, err := unburiedSession.GetTreasureMap()
		if err != nil {
			raven.CaptureError(err, nil)
			return err
		}

		BuryTreasure(treasureIndex, &unburiedSession)
	}
	return nil
}

func BuryTreasure(treasureIndexMap []models.TreasureMap, unburiedSession *models.UploadSession) error {

	for i, entry := range treasureIndexMap {
		treasureChunks, err := models.GetDataMapByGenesisHashAndChunkIdx(unburiedSession.GenesisHash, entry.Idx)
		if err != nil {
			raven.CaptureError(err, nil)
			return err
		}
		if len(treasureChunks) == 0 || len(treasureChunks) > 1 {
			errString := "did not find a chunk that matched genesis_hash and chunk_idx in process_paid_sessions, or " +
				"found duplicate chunks"
			log.Println(errString)
			err = errors.New(errString)
			raven.CaptureError(err, nil)
			return err
		}
		treasureChunks[0].Message, err = models.CreateTreasurePayload(entry.Key, treasureChunks[0].Hash, models.MaxSideChainLength)
		if err != nil {
			raven.CaptureError(err, nil)
			return err
		}
		models.DB.ValidateAndSave(&treasureChunks[0])
		// delete the keys now that they have been buried
		treasureIndexMap[i].Key = ""
	}
	unburiedSession.TreasureStatus = models.TreasureBuried
	unburiedSession.SetTreasureMap(treasureIndexMap)
	models.DB.ValidateAndSave(unburiedSession)
	return nil
}

// marking the maps as "Unassigned" will trigger them to get processed by the process_unassigned_chunks cron task.
func MarkBuriedMapsAsUnassigned() {
	readySessions, err := models.GetReadySessions()
	if err != nil {
		raven.CaptureError(err, nil)
	}

	for _, readySession := range readySessions {
		err = readySession.BulkMarkDataMapsAsUnassigned()
	}
}

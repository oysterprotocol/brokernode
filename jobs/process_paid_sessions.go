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

	BuryTreasureInPaidDataMaps()
	MarkBuriedMapsAsUnassigned()
}

func BuryTreasureInPaidDataMaps() error {

	unburiedSessions := []models.UploadSession{}

	err := models.DB.Where("payment_status = ? AND treasure_status = ?",
		models.PaymentStatusPaid, models.TreasureUnburied).All(&unburiedSessions)
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
		treasureChunks := []models.DataMap{}
		err := models.DB.Where("genesis_hash = ?",
			unburiedSession.GenesisHash).Where("chunk_idx = ?", entry.Idx).All(&treasureChunks)
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
	readySessions := []models.UploadSession{}

	err := models.DB.Where("payment_status = ? AND treasure_status = ?",
		models.PaymentStatusPaid, models.TreasureBuried).All(&readySessions)
	if err != nil {
		raven.CaptureError(err, nil)
	}

	var dataMaps = []models.DataMap{}

	for _, readySession := range readySessions {
		err = models.DB.RawQuery("UPDATE data_maps SET status = ? WHERE genesis_hash = ? AND status = ?",
			models.Unassigned,
			readySession.GenesisHash,
			models.Pending).All(&dataMaps)
	}
}

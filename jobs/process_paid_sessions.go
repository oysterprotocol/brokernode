package jobs

import (
	"errors"
	"fmt"
	"github.com/getsentry/raven-go"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/utils"
	"gopkg.in/segmentio/analytics-go.v3"
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
		fmt.Println(err)
		raven.CaptureError(err, nil)
	}

	for _, unburiedSession := range unburiedSessions {

		treasureIndex, err := unburiedSession.GetTreasureMap()
		if err != nil {
			fmt.Println(err)
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
			fmt.Println(err)
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
			fmt.Println(err)
			raven.CaptureError(err, nil)
			return err
		}
		models.DB.ValidateAndSave(&treasureChunks[0])
		// delete the keys now that they have been buried
		treasureIndexMap[i].Key = ""

		oyster_utils.LogToSegment("process_paid_sessions: treasure_payload_buried_in_data_map", analytics.NewProperties().
			Set("genesis_hash", unburiedSession.GenesisHash).
			Set("sector", entry.Sector).
			Set("chunk_idx", entry.Idx).
			Set("address", treasureChunks[0].Address).
			Set("message", treasureChunks[0].Message))
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
		fmt.Println(err)
		raven.CaptureError(err, nil)
	}

	for _, readySession := range readySessions {

		pendingChunks, err := models.GetPendingChunksBySession(readySession, 1)
		if err != nil {
			fmt.Println(err)
			raven.CaptureError(err, nil)
		}

		if len(pendingChunks) > 0 {
			oyster_utils.LogToSegment("process_paid_sessions: mark_data_maps_as_ready", analytics.NewProperties().
				Set("genesis_hash", readySession.GenesisHash))

			err = readySession.BulkMarkDataMapsAsUnassigned()
		}
	}
}

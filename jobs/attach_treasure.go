package jobs

import (
	"github.com/gobuffalo/validate"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"gopkg.in/segmentio/analytics-go.v3"
	"time"
)

func AttachTreasuresToTangle(iotaWrapper services.IotaService, PrometheusWrapper services.PrometheusService,
	thresholdTime time.Time) {

	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramAttachTreasuresToTangle, start)

	AttachTreasureTransactions(iotaWrapper)
	VerifyTreasureTransactions(iotaWrapper, thresholdTime)
}

func AttachTreasureTransactions(iotaWrapper services.IotaService) {
	var vErr *validate.Errors
	treasuresToAttach, err := models.GetTreasuresToBuryBySignedStatus([]models.SignedStatus{models.TreasureSigned,
		models.TreasureAttachError})
	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return
	}

	for _, treasure := range treasuresToAttach {
		chunk := oyster_utils.ChunkData{
			Address:     treasure.Address,
			Message:     treasure.Message,
			GenesisHash: treasure.GenesisHash,
			RawMessage:  treasure.RawMessage,
			Idx:         treasure.Idx,
		}

		err := iotaWrapper.DoPoW([]oyster_utils.ChunkData{chunk})
		if err == nil {
			treasure.SignedStatus = models.TreasureSignedAndAttached
			vErr, err = models.DB.ValidateAndUpdate(treasure)
			oyster_utils.LogIfValidationError("error updating treasure's SignedStatus to "+
				"TreasureSignedAndAttached in attach_treasure", vErr, nil)
			logTreasureAttachmentResult("attach_treasure: attachment attempted", treasure)
		} else {
			logTreasureAttachmentResult("attach_treasure: attachment attempted: ERROR", treasure)
		}
		oyster_utils.LogIfError(err, nil)
	}
}

func VerifyTreasureTransactions(iotaWrapper services.IotaService, thresholdTime time.Time) {
	treasuresToAttach, err := models.GetTreasuresToBuryBySignedStatus([]models.SignedStatus{
		models.TreasureSignedAndAttached})
	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return
	}

	for _, treasure := range treasuresToAttach {
		chunk := oyster_utils.ChunkData{
			Address:     treasure.Address,
			Message:     treasure.Message,
			GenesisHash: treasure.GenesisHash,
			RawMessage:  treasure.RawMessage,
			Idx:         treasure.Idx,
		}

		filteredChunks, err := iotaWrapper.VerifyChunkMessagesMatchRecord([]oyster_utils.ChunkData{chunk})
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			continue
		}

		handleAttachmentResults(filteredChunks, treasure, thresholdTime)
	}
}

func handleAttachmentResults(filteredChunks services.FilteredChunk, treasure models.Treasure,
	thresholdTime time.Time) {
	var err error
	var vErr *validate.Errors

	if len(filteredChunks.MatchesTangle) > 0 {
		treasure.SignedStatus = models.TreasureSignedAndAttachmentVerified
		vErr, err = models.DB.ValidateAndUpdate(treasure)
		logTreasureAttachmentResult("attach_treasure: attachment verified", treasure)
	} else if len(filteredChunks.DoesNotMatchTangle) > 0 {
		treasure.SignedStatus = models.TreasureAttachError
		vErr, err = models.DB.ValidateAndUpdate(treasure)
		logTreasureAttachmentResult("attach_treasure: attachment error", treasure)
	} else if len(filteredChunks.NotAttached) > 0 {
		if treasure.UpdatedAt.Before(thresholdTime) {
			treasure.SignedStatus = models.TreasureSigned
			vErr, err = models.DB.ValidateAndUpdate(treasure)
			logTreasureAttachmentResult("attach_treasure: attachment timed out", treasure)
		}
	}

	oyster_utils.LogIfValidationError("error updating treasure's SignedStatus to "+
		models.SignedStatusMap[treasure.SignedStatus]+
		" TreasureAttachError in attach_treasure", vErr, nil)
	oyster_utils.LogIfError(err, nil)
}

func logTreasureAttachmentResult(eventName string, treasure models.Treasure) {
	oyster_utils.LogToSegment(eventName, analytics.NewProperties().
		Set("tangle_address", treasure.Address).
		Set("genesis_hash", treasure.GenesisHash).
		Set("expected_message", treasure.Message).
		Set("encryption_index", treasure.EncryptionIndex).
		Set("index", treasure.Idx))
}

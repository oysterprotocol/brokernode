package jobs

import (
	raven "github.com/getsentry/raven-go"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
)

func init() {
}

func ProcessUnassignedChunks(iotaWrapper services.IotaService) {

	chunks, err := GetUnassignedChunks()
	if err != nil {
		raven.CaptureError(err, nil)
	}

	if len(chunks) > 0 {
		AssignChunksToChannels(chunks, iotaWrapper)
	}
}

func AssignChunksToChannels(chunks []models.DataMap, iotaWrapper services.IotaService) {

	/*
		TODO:  More sophisticated chunk grabbing.  I.e. only grab as many as
		we have ready channels for, and try to grab an equal number per unique
		genesis hash
	*/

	for i := 0; i < len(chunks); i += BundleSize {
		end := i + BundleSize

		if end > len(chunks) {
			end = len(chunks)
		}

		if i >= end {
			break
		}

		channel, err := models.GetOneReadyChannel()
		if err != nil {
			raven.CaptureError(err, nil)
		}
		if channel.ChannelID == "" {
			break
		}

		filteredChunks, err := IotaWrapper.VerifyChunkMessagesMatchRecord(chunks[i:end])

		if err != nil {
			raven.CaptureError(err, nil)
		}
		chunksToSend := append(filteredChunks.NotAttached, filteredChunks.DoesNotMatchTangle...)

		if len(chunksToSend) > 0 {
			iotaWrapper.SendChunksToChannel(chunksToSend, &channel)
		}
		if len(filteredChunks.MatchesTangle) > 0 {
			for _, chunk := range filteredChunks.MatchesTangle {
				chunk.Status = models.Complete
				models.DB.ValidateAndSave(&chunk)
			}
		}
	}
}

func GetUnassignedChunks() (dataMaps []models.DataMap, err error) {

	query := models.DB.Where("status = ?", models.Unassigned)
	dataMaps = []models.DataMap{}
	err = query.All(&dataMaps)
	if err != nil {
		raven.CaptureError(err, nil)
	}

	return dataMaps, err
}

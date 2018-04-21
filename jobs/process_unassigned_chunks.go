package jobs

import (
	"github.com/getsentry/raven-go"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
)

func init() {
}

func ProcessUnassignedChunks(iotaWrapper services.IotaService) {

	sessions, err := models.GetSessionsByAge()
	if err != nil {
		raven.CaptureError(err, nil)
	}

	if len(sessions) > 0 {
		GetSessionUnassignedChunks(sessions, iotaWrapper)
	}
}

func GetSessionUnassignedChunks(sessions []models.UploadSession, iotaWrapper services.IotaService) {
	for _, session := range sessions {
		channels, err := models.GetReadyChannels()
		if err != nil {
			raven.CaptureError(err, nil)
		}
		if len(channels) <= 0 {
			break
		}

		chunks, err := models.GetUnassignedChunksBySession(session, len(channels)*BundleSize)
		AssignChunksToChannels(chunks, channels, iotaWrapper)
		if len(chunks) == len(channels)*BundleSize {
			// we have used up all the channels, no point in doing the for loop again
			break
		}
	}
}

func AssignChunksToChannels(chunks []models.DataMap, channels []models.ChunkChannel, iotaWrapper services.IotaService) {

	// as long as there are still chunks and still channels this for loop continues
	for ok, i, j := true, 0, 0; ok; ok = i < len(chunks) && j < len(channels) {
		end := i + BundleSize

		if end > len(chunks) {
			end = len(chunks)
		}

		if i >= end {
			break
		}

		filteredChunks, err := iotaWrapper.VerifyChunkMessagesMatchRecord(chunks[i:end])

		if err != nil {
			raven.CaptureError(err, nil)
		}
		chunksToSend := append(filteredChunks.NotAttached, filteredChunks.DoesNotMatchTangle...)

		if len(chunksToSend) > 0 {
			iotaWrapper.SendChunksToChannel(chunksToSend, &channels[j])
		}
		if len(filteredChunks.MatchesTangle) > 0 {
			for _, chunk := range filteredChunks.MatchesTangle {
				chunk.Status = models.Complete
				models.DB.ValidateAndSave(&chunk)
			}
		}
		j++
		i += BundleSize
	}
}

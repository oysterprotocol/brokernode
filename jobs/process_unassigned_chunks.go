package jobs

import (
	"github.com/getsentry/raven-go"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"gopkg.in/segmentio/analytics-go.v3"
	"time"
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

		if len(chunks) > 0 {
			AssignChunksToChannels(chunks, channels, iotaWrapper)

			oyster_utils.LogToSegment("process_unassigned_chunks: processing_chunks_for_session", analytics.NewProperties().
				Set("genesis_hash", session.GenesisHash).
				Set("id", session.ID).
				Set("num_chunks_processing", len(chunks)).
				Set("num_ready_channels", len(channels)))
		}

		if len(chunks) == len(channels)*BundleSize {
			// we have used up all the channels, no point in doing the for loop again
			break
		}
	}
}

func AssignChunksToChannels(chunks []models.DataMap, channels []models.ChunkChannel, iotaWrapper services.IotaService) {

	defer oyster_utils.TimeTrack(time.Now(), "process_unassigned_chunks: assign_chunks_to_channel", analytics.NewProperties().
		Set("num_chunks", len(chunks)).
		Set("num_channels", len(channels)))

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

			addresses, indexes := models.MapChunkIndexesAndAddresses(chunksToSend)

			oyster_utils.LogToSegment("process_unassigned_chunks: sending_chunks_to_channel", analytics.NewProperties().
				Set("genesis_hash", chunksToSend[0].GenesisHash).
				Set("num_chunks", len(chunksToSend)).
				Set("channel_id", channels[j].ChannelID).
				Set("addresses", addresses).
				Set("chunk_indexes", indexes))

			iotaWrapper.SendChunksToChannel(chunksToSend, &channels[j])
		}
		if len(filteredChunks.MatchesTangle) > 0 {
			addresses, indexes := models.MapChunkIndexesAndAddresses(chunksToSend)

			oyster_utils.LogToSegment("process_unassigned_chunks: chunks_already_attached", analytics.NewProperties().
				Set("genesis_hash", filteredChunks.MatchesTangle[0].GenesisHash).
				Set("num_chunks", len(filteredChunks.MatchesTangle)).
				Set("addresses", addresses).
				Set("chunk_indexes", indexes))

			for _, chunk := range filteredChunks.MatchesTangle {
				chunk.Status = models.Complete
				models.DB.ValidateAndSave(&chunk)
			}
		}
		j++
		i += BundleSize
	}
}

package jobs

import (
	"errors"
	"math"
	"os"
	"time"

	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"gopkg.in/segmentio/analytics-go.v3"
)

const PercentOfChunksToSkipVerification = 45

func ProcessUnassignedChunks(iotaWrapper services.IotaService, PrometheusWrapper services.PrometheusService) {

	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramProcessUnassignedChunks, start)

	sessions, _ := models.GetReadySessions()

	if len(sessions) > 0 {
		GetSessionUnassignedChunks(sessions, iotaWrapper)
	}
}

func GetSessionUnassignedChunks(sessions []models.UploadSession, iotaWrapper services.IotaService) {
	for _, session := range sessions {
		channels, _ := models.GetReadyChannels()

		if len(channels) <= 0 {
			break
		}

		chunks, _ := session.GetUnassignedChunksBySession(len(channels) * BundleSize)

		if len(chunks) > 0 {

			FilterAndAssignChunksToChannels(chunks, channels, iotaWrapper, session)

			oyster_utils.LogToSegment("process_unassigned_chunks: processing_chunks_for_session", analytics.NewProperties().
				Set("genesis_hash", session.GenesisHash).
				Set("id", session.ID).
				Set("num_chunks_processing", len(chunks)).
				Set("num_ready_channels", len(channels)))
		}

		if len(chunks) >= len(channels)*BundleSize {
			// we have used up all the channels, no point in doing the for loop again
			break
		}
	}
}

/*
There are 3 "types" of chunks that we care about and we have to handle them differently:
	1.  A normal chunk which is not yet attached
	2.  A normal chunk which is already attached
	3.  A normal chunk which is already attached,
		but the message is wrong
We filter them in the iotaWrapper.VerifyChunkMessagesMatchRecord method.
*/
func FilterAndAssignChunksToChannels(chunksIn []oyster_utils.ChunkData, channels []models.ChunkChannel,
	iotaWrapper services.IotaService, sessionParam models.UploadSession) {

	session := &models.UploadSession{}
	models.DB.Find(session, sessionParam.ID)

	defer oyster_utils.TimeTrack(time.Now(), "process_unassigned_chunks: filter_and_assign_chunks_to_channel", analytics.NewProperties().
		Set("num_chunks", len(chunksIn)).
		Set("num_channels", len(channels)))

	for i := 0; i < len(chunksIn); i += services.MaxNumberOfAddressPerFindTransactionRequest {
		end := i + services.MaxNumberOfAddressPerFindTransactionRequest

		if end > len(chunksIn) {
			end = len(chunksIn)
		}

		if i >= end {
			break
		}

		chunks := chunksIn[i:end]

		skipVerifyOfChunks, restOfChunks := SkipVerificationOfFirstChunks(chunks, *session)

		filteredChunks, err := iotaWrapper.VerifyChunkMessagesMatchRecord(restOfChunks)

		if err != nil {
			oyster_utils.LogIfError(errors.New(err.Error()+" in FilterAndAssignChunksToChannels() in "+
				"process_unassigned_chunks"), map[string]interface{}{
				"forLoopIndex":   i,
				"totalLoopCount": len(chunksIn),
				"numOfChunk":     len(restOfChunks),
			})
		}

		if len(filteredChunks.MatchesTangle) > 0 {

			oyster_utils.LogToSegment("process_unassigned_chunks: chunks_already_attached", analytics.NewProperties().
				Set("num_chunks", len(filteredChunks.MatchesTangle)))

			session.UpdateIndexWithAttachedChunks(filteredChunks.MatchesTangle)
			session.UpdateIndexWithVerifiedChunks(filteredChunks.MatchesTangle)
		}

		chunksToSend := []oyster_utils.ChunkData{}
		chunksToSend = append(chunksToSend, skipVerifyOfChunks...)
		chunksToSend = append(chunksToSend, filteredChunks.NotAttached...)
		chunksToSend = append(chunksToSend, filteredChunks.DoesNotMatchTangle...)

		if oyster_utils.PoWMode == oyster_utils.PoWEnabled && len(chunksToSend) > 0 {
			session.UpdateIndexWithAttachedChunks(chunksToSend)
			if os.Getenv("ENABLE_LAMBDA") == "true" {
				iotaWrapper.SendChunksToLambda(&chunksToSend)
			} else {
				SendChunks(chunksToSend, channels, iotaWrapper, *session)
			}
		}
	}
}

/*SkipVerificationOfFirstChunks will skip verifying for the first PercentOfChunksToSkipVerification% of chunks of
the alpha session and the last PercentOfChunksToSkipVerification% of the beta session*/
func SkipVerificationOfFirstChunks(chunks []oyster_utils.ChunkData, session models.UploadSession) ([]oyster_utils.ChunkData,
	[]oyster_utils.ChunkData) {

	if len(chunks) == 0 {
		return []oyster_utils.ChunkData{},
			[]oyster_utils.ChunkData{}
	}

	numChunks := session.NumChunks

	var lenOfChunksToSkipVerifying int
	lenOfChunksToSkipVerifying = int(float64(numChunks) * float64(PercentOfChunksToSkipVerification) / float64(100))

	var lenOfChunksToVerify int
	lenOfChunksToVerify = numChunks - lenOfChunksToSkipVerifying

	var skipVerifyMinIdx int
	var skipVerifyMaxIdx int
	var verifyMinIdx int
	var verifyMaxIdx int

	if session.Type == models.SessionTypeAlpha {
		skipVerifyMinIdx = 0
		skipVerifyMaxIdx = lenOfChunksToSkipVerifying - 1
		verifyMinIdx = lenOfChunksToSkipVerifying
		verifyMaxIdx = numChunks - 1
	} else {
		skipVerifyMinIdx = numChunks - lenOfChunksToSkipVerifying
		skipVerifyMaxIdx = numChunks - 1
		verifyMinIdx = 0
		verifyMaxIdx = lenOfChunksToVerify - 1
	}

	if skipVerifyMinIdx == skipVerifyMaxIdx {
		// very small file, don't bother with filtering
		return []oyster_utils.ChunkData{}, chunks
	}

	// first check that any are in the first third before we bother with this
	maxIdx := int64(math.Max(float64(chunks[0].Idx), float64(chunks[len(chunks)-1].Idx)))
	minIdx := int64(math.Min(float64(chunks[0].Idx), float64(chunks[len(chunks)-1].Idx)))

	if minIdx >= int64(verifyMinIdx) && minIdx <= int64(verifyMaxIdx) &&
		maxIdx >= int64(verifyMinIdx) && maxIdx <= int64(verifyMaxIdx) {
		return []oyster_utils.ChunkData{}, chunks
	}

	skipVerifyOfChunks := []oyster_utils.ChunkData{}
	restOfChunks := []oyster_utils.ChunkData{}

	for i := 0; i < len(chunks); i++ {
		if chunks[i].Idx >= int64(skipVerifyMinIdx) && chunks[i].Idx <= int64(skipVerifyMaxIdx) {
			skipVerifyOfChunks = append(skipVerifyOfChunks, chunks[i])
		} else {
			restOfChunks = append(restOfChunks, chunks[i])
		}
	}

	return skipVerifyOfChunks, restOfChunks
}

// actually send the chunks
func SendChunks(chunks []oyster_utils.ChunkData, channels []models.ChunkChannel, iotaWrapper services.IotaService, session models.UploadSession) {
	// as long as there are still chunks and still channels, this for loop continues
	for ok, i, j := true, 0, 0; ok; ok = i < len(chunks) && j < len(channels) {
		end := i + BundleSize

		if end > len(chunks) {
			end = len(chunks)
		}

		if i >= end {
			break
		}

		if len(chunks[i:end]) > 0 {

			oyster_utils.LogToSegment("process_unassigned_chunks: sending_chunks_to_channel", analytics.NewProperties().
				Set("genesis_hash", session.GenesisHash).
				Set("num_chunks", len(chunks[i:end])).
				Set("channel_id", channels[j].ChannelID))

			iotaWrapper.SendChunksToChannel(chunks[i:end], &channels[j])
		}
		j++
		i += BundleSize
	}
}

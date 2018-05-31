package jobs

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/getsentry/raven-go"
	"github.com/iotaledger/giota"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"gopkg.in/segmentio/analytics-go.v3"
)

const PercentOfChunksToSkipVerification = 45

func ProcessUnassignedChunks(iotaWrapper services.IotaService) {
	sessions, _ := models.GetSessionsByAge()

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

		chunks, _ := models.GetUnassignedChunksBySession(session, len(channels)*BundleSize)

		if len(chunks) > 0 {
			FilterAndAssignChunksToChannels(chunks, channels, iotaWrapper, session)

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

/*
There are 5 "types" of chunks that we care about and we have to handle them differently:
	1.  A normal chunk which is not yet attached
	2.  A normal chunk which is already attached
	3.  A normal chunks which is already attached,
		but the message is wrong
	4.  A treasure chunk which is not yet attached
	5.  A treasure chunk which is already attached

We check for the first 3 and filter them in the iotaWrapper.VerifyChunkMessagesMatchRecord method.

For the treasure chunks it's different because the two brokers will have different messages for
the treasures, so we don't expect the messages to match.  It is sufficient only that a transaction
already exists at that address, and if so, we will not attempt to attach the treasure chunk because
the other broker has already done so.

We separate out the treasure chunks from the regular chunks and keep the ones we need to attach,
then we filter the other types of chunks, then we insert the treasure chunks that need attaching
back into the array based on their chunk_idx, then we send to the channels.
*/
func FilterAndAssignChunksToChannels(chunksIn []models.DataMap, channels []models.ChunkChannel, iotaWrapper services.IotaService, session models.UploadSession) {

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

		chunks, treasureChunksNeedAttaching := HandleTreasureChunks(chunksIn[i:end], session, iotaWrapper)

		skipVerifyOfChunks, restOfChunks := SkipVerificationOfFirstChunks(chunks, session)

		filteredChunks, err := iotaWrapper.VerifyChunkMessagesMatchRecord(restOfChunks)
		oyster_utils.LogIfError(err)

		if len(filteredChunks.MatchesTangle) > 0 {

			oyster_utils.LogToSegment("process_unassigned_chunks: chunks_already_attached", analytics.NewProperties().
				Set("genesis_hash", filteredChunks.MatchesTangle[0].GenesisHash).
				Set("num_chunks", len(filteredChunks.MatchesTangle)))

			for _, chunk := range filteredChunks.MatchesTangle {
				chunk.Status = models.Complete
				models.DB.ValidateAndSave(&chunk)
			}
		}

		nonTreasureChunksToSend := append(skipVerifyOfChunks, filteredChunks.NotAttached...)
		nonTreasureChunksToSend = append(nonTreasureChunksToSend, filteredChunks.DoesNotMatchTangle...)

		chunksIncludingTreasureChunks := InsertTreasureChunks(nonTreasureChunksToSend, treasureChunksNeedAttaching, session)

		StageTreasures(treasureChunksNeedAttaching, session)

		if os.Getenv("DISABLE_POW") == "" {
			SendChunks(chunksIncludingTreasureChunks, channels, iotaWrapper, session)
		}
	}
}

func SkipVerificationOfFirstChunks(chunks []models.DataMap, session models.UploadSession) ([]models.DataMap, []models.DataMap) {

	if len(chunks) == 0 {
		return []models.DataMap{}, []models.DataMap{}
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
		return []models.DataMap{}, chunks
	}

	// first check that any are in the first third before we bother with this
	firstIndex := chunks[0].ChunkIdx
	lastIndex := chunks[len(chunks)-1].ChunkIdx

	if firstIndex >= verifyMinIdx && firstIndex <= verifyMaxIdx &&
		lastIndex >= verifyMinIdx && lastIndex <= verifyMaxIdx {
		return []models.DataMap{}, chunks
	}

	skipVerifyOfChunks := []models.DataMap{}
	restOfChunks := []models.DataMap{}

	for i := 0; i < len(chunks); i++ {
		if chunks[i].ChunkIdx >= skipVerifyMinIdx && chunks[i].ChunkIdx <= skipVerifyMaxIdx {
			skipVerifyOfChunks = append(skipVerifyOfChunks, chunks[i])
		} else {
			restOfChunks = append(restOfChunks, chunks[i])
		}
	}

	return skipVerifyOfChunks, restOfChunks
}

func StageTreasures(treasureChunks []models.DataMap, session models.UploadSession) {
	/*TODO add tests for this method*/

	if len(treasureChunks) == 0 {
		return
	}
	treasureIdxMapArray, err := session.GetTreasureMap()
	if err != nil {
		fmt.Println("Cannot stage treasures to bury in process_unassigned_chunks: " + err.Error())
		// already captured error in upstream function
		return
	}
	if len(treasureIdxMapArray) == 0 {
		fmt.Println("Cannot stage treasures to bury in process_unassigned_chunks: " + "treasureIdxMapArray is empty")
		return
	}

	treasureIdxMap := make(map[int]models.TreasureMap)
	for _, treasureIdxEntry := range treasureIdxMapArray {
		treasureIdxMap[treasureIdxEntry.Idx] = treasureIdxEntry
	}

	prlPerTreasure, err := session.GetPRLsPerTreasure()
	if err != nil {
		fmt.Println("Cannot stage treasures to bury in process_unassigned_chunks: " + err.Error())
		// captured error in upstream method
		return
	}

	prlInWei := oyster_utils.ConvertToWeiUnit(prlPerTreasure)

	for _, treasureChunk := range treasureChunks {
		if _, ok := treasureIdxMap[treasureChunk.ChunkIdx]; ok {
			decryptedKey, err := treasureChunk.DecryptEthKey(treasureIdxMap[treasureChunk.ChunkIdx].Key)
			if err != nil {
				fmt.Println("Cannot stage treasures to bury in process_unassigned_chunks: " + err.Error())
				// already captured error in upstream function
				return
			}
			ethAddress := services.EthWrapper.GenerateEthAddrFromPrivateKey(decryptedKey)
			treasureToBury := models.Treasure{
				ETHAddr: ethAddress.Hex(),
				ETHKey:  decryptedKey,
				Address: treasureChunk.Address,
				Message: treasureChunk.Message,
			}

			treasureToBury.SetPRLAmount(prlInWei)

			models.DB.ValidateAndCreate(&treasureToBury)
		}
	}
}

// add the treasure chunks back into the array in the appropriate position
func InsertTreasureChunks(chunks []models.DataMap, treasureChunks []models.DataMap, session models.UploadSession) []models.DataMap {

	if len(chunks) == 0 && len(treasureChunks) == 0 {
		return []models.DataMap{}
	}
	if len(treasureChunks) == 0 {
		return chunks
	}
	if len(chunks) == 0 {
		return treasureChunks
	}

	var idxTarget int
	if session.Type == models.SessionTypeAlpha {
		idxTarget = 1
	} else {
		idxTarget = -1
	}

	treasureChunksMapped := make(map[int]models.DataMap)
	for _, treasureChunk := range treasureChunks {
		treasureChunksMapped[treasureChunk.ChunkIdx] = treasureChunk
	}

	defer oyster_utils.TimeTrack(time.Now(), "process_unassigned_chunks: reinsert_treasure_chunks", analytics.NewProperties().
		Set("num_chunks", len(chunks)).
		Set("num_treasure_chunks", len(treasureChunks)))

	treasureChunksInserted := 0

	// this puts the treasure chunks back into the array where they belong
	for ok, i := true, 0; ok; ok = treasureChunksInserted < len(treasureChunks) && i < len(chunks) {
		if _, ok := treasureChunksMapped[chunks[i].ChunkIdx-idxTarget]; ok && i == 0 {
			chunks = append([]models.DataMap{treasureChunksMapped[chunks[i].ChunkIdx-idxTarget]}, chunks...)
			treasureChunksInserted++
			i++ // skip an iteration
		} else if _, ok := treasureChunksMapped[chunks[i].ChunkIdx+idxTarget]; ok &&
			i == len(chunks)-1 {
			chunks = append(chunks, treasureChunksMapped[chunks[i].ChunkIdx+idxTarget])
			treasureChunksInserted++
			i++ // skip an iteration
		} else if _, ok := treasureChunksMapped[chunks[i].ChunkIdx+idxTarget]; ok {
			chunks = append(chunks[:i+2], chunks[i+1:]...)
			chunks[i+1] = treasureChunksMapped[chunks[i].ChunkIdx+idxTarget]
			treasureChunksInserted++
			i++ // skip an iteration
		}
		i++
	}
	return chunks
}

// actually send the chunks
func SendChunks(chunks []models.DataMap, channels []models.ChunkChannel, iotaWrapper services.IotaService, session models.UploadSession) {
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

			//addresses, indexes := models.MapChunkIndexesAndAddresses(chunks[i:end])

			oyster_utils.LogToSegment("process_unassigned_chunks: sending_chunks_to_channel", analytics.NewProperties().
				Set("genesis_hash", chunks[i:end][0].GenesisHash).
				Set("num_chunks", len(chunks[i:end])).
				Set("channel_id", channels[j].ChannelID))
			//Set("addresses", addresses).
			//Set("chunk_indexes", indexes))

			iotaWrapper.SendChunksToChannel(chunks[i:end], &channels[j])
		}
		j++
		i += BundleSize
	}
}

// check if a transaction exists for a treasure chunk's address.  If it does, mark it as complete.  If it doesn't,
// we need to separate it from the other chunks so it doesn't get filtered out when VerifyChunkMessagesMatchRecord is
// finding chunks that don't match the tangle.  Later we'll re-add it to the array for sending to the channels.
func HandleTreasureChunks(chunks []models.DataMap, session models.UploadSession, iotaWrapper services.IotaService) ([]models.DataMap, []models.DataMap) {

	var chunksToAttach []models.DataMap
	var treasureChunksToAttach []models.DataMap

	treasureIndexes, err := session.GetTreasureIndexes()
	oyster_utils.LogIfError(err)

	if len(chunks) == 0 {
		return chunks, []models.DataMap{}
	}

	maxIdx := int(math.Max(float64(chunks[0].ChunkIdx), float64(chunks[len(chunks)-1].ChunkIdx)))
	minIdx := int(math.Min(float64(chunks[0].ChunkIdx), float64(chunks[len(chunks)-1].ChunkIdx)))

	treasureMap := make(map[int]bool)
	for _, idx := range treasureIndexes {
		if idx >= minIdx && idx <= maxIdx {
			treasureMap[idx] = true
		}
	}

	if len(treasureMap) == 0 {
		return chunks, []models.DataMap{}
	}

	for i := 0; i < len(chunks); i++ {
		if _, ok := treasureMap[chunks[i].ChunkIdx]; ok {
			address := make([]giota.Address, 0, 1)
			address = append(address, giota.Address(chunks[i].Address))
			transactionsMap, err := iotaWrapper.FindTransactions(address)
			if err != nil {
				fmt.Println(err)
				raven.CaptureError(err, nil)
			}
			if _, ok := transactionsMap[giota.Address(chunks[i].Address)]; !ok || transactionsMap == nil {
				oyster_utils.LogToSegment("process_unassigned_chunks: "+
					"treasure_chunk_not_attached", analytics.NewProperties().
					Set("genesis_hash", chunks[i].GenesisHash).
					Set("address", chunks[i].Address).
					Set("chunk_index", chunks[i].ChunkIdx).
					Set("message", chunks[i].Message))

				treasureChunksToAttach = append(treasureChunksToAttach, chunks[i])
			} else {
				oyster_utils.LogToSegment("process_unassigned_chunks: "+
					"treasure_chunk_already_attached", analytics.NewProperties().
					Set("genesis_hash", chunks[i].GenesisHash).
					Set("address", chunks[i].Address).
					Set("chunk_index", chunks[i].ChunkIdx).
					Set("message", chunks[i].Message))

				chunks[i].Status = models.Complete
				models.DB.ValidateAndSave(&chunks[i])
			}
		} else {
			chunksToAttach = append(chunksToAttach, chunks[i])
		}
	}

	return chunksToAttach, treasureChunksToAttach
}

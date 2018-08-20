package jobs

import (
	"errors"
	"fmt"
	"github.com/iotaledger/giota"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"gopkg.in/segmentio/analytics-go.v3"
	"math"
	"os"
	"time"
)

const PercentOfChunksToSkipVerification = 45

func ProcessUnassignedChunks(iotaWrapper services.IotaService, PrometheusWrapper services.PrometheusService) {

	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramProcessUnassignedChunks, start)

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

		chunks, _ := session.GetUnassignedChunksBySession(len(channels) * BundleSize)

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
func FilterAndAssignChunksToChannels(chunksIn []oyster_utils.ChunkData, channels []models.ChunkChannel,
	iotaWrapper services.IotaService, session models.UploadSession) {

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

			session.MoveChunksToCompleted(filteredChunks.MatchesTangle)
			session.UpdateIndexWithVerifiedChunks(filteredChunks.MatchesTangle)
		}

		nonTreasureChunksToSend := append(skipVerifyOfChunks, filteredChunks.NotAttached...)
		nonTreasureChunksToSend = append(nonTreasureChunksToSend, filteredChunks.DoesNotMatchTangle...)
		chunksIncludingTreasureChunks := InsertTreasureChunks(nonTreasureChunksToSend, treasureChunksNeedAttaching, session)

		StageTreasures(treasureChunksNeedAttaching, session)

		if oyster_utils.PoWMode == oyster_utils.PoWEnabled {
			SendChunks(chunksIncludingTreasureChunks, channels, iotaWrapper, session)
		}
	}
}

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

func StageTreasures(treasureChunks []oyster_utils.ChunkData, session models.UploadSession) {
	/*TODO add tests for this method*/

	if len(treasureChunks) == 0 || oyster_utils.BrokerMode != oyster_utils.ProdMode {
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

	treasureIdxMap := make(map[int64]models.TreasureMap)
	for _, treasureIdxEntry := range treasureIdxMapArray {
		treasureIdxMap[int64(treasureIdxEntry.Idx)] = treasureIdxEntry
	}

	prlPerTreasure, err := session.GetPRLsPerTreasure()
	if err != nil {
		fmt.Println("Cannot stage treasures to bury in process_unassigned_chunks: " + err.Error())
		// captured error in upstream method
		return
	}

	prlInWei := oyster_utils.ConvertToWeiUnit(prlPerTreasure)

	for _, treasureChunk := range treasureChunks {
		if _, ok := treasureIdxMap[treasureChunk.Idx]; ok {
			decryptedKey, err := session.DecryptTreasureChunkEthKey(treasureIdxMap[treasureChunk.Idx].Key)
			if err != nil {
				fmt.Println("Cannot stage treasure to bury in process_unassigned_chunks: " + err.Error())
				// already captured error in upstream function
				continue
			}

			if decryptedKey == os.Getenv("TEST_MODE_WALLET_KEY") {
				continue
			}

			if oyster_utils.BrokerMode == oyster_utils.ProdMode {
				ethAddress := EthWrapper.GenerateEthAddrFromPrivateKey(decryptedKey)

				treasureToBury := models.Treasure{
					GenesisHash: session.GenesisHash,
					ETHAddr:     ethAddress.Hex(),
					ETHKey:      decryptedKey,
					Address:     treasureChunk.Address,
					Message:     treasureChunk.Message,
				}

				treasureToBury.SetPRLAmount(prlInWei)

				models.DB.ValidateAndCreate(&treasureToBury)
			}
		}
	}
}

// add the treasure chunks back into the array in the appropriate position
func InsertTreasureChunks(chunks []oyster_utils.ChunkData, treasureChunks []oyster_utils.ChunkData,
	session models.UploadSession) []oyster_utils.ChunkData {

	if len(chunks) == 0 && len(treasureChunks) == 0 {
		return []oyster_utils.ChunkData{}
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

	treasureChunksMapped := make(map[int64]oyster_utils.ChunkData)
	for _, treasureChunk := range treasureChunks {
		treasureChunksMapped[treasureChunk.Idx] = treasureChunk
	}

	defer oyster_utils.TimeTrack(time.Now(), "process_unassigned_chunks: reinsert_treasure_chunks", analytics.NewProperties().
		Set("num_chunks", len(chunks)).
		Set("num_treasure_chunks", len(treasureChunks)))

	treasureChunksInserted := 0

	// this puts the treasure chunks back into the array where they belong
	for ok, i := true, 0; ok; ok = treasureChunksInserted < len(treasureChunks) && i < len(chunks) {
		if _, ok := treasureChunksMapped[chunks[i].Idx-int64(idxTarget)]; ok && i == 0 {
			chunks = append([]oyster_utils.ChunkData{treasureChunksMapped[chunks[i].Idx-int64(idxTarget)]}, chunks...)
			treasureChunksInserted++
			i++ // skip an iteration
		} else if _, ok := treasureChunksMapped[chunks[i].Idx+int64(idxTarget)]; ok &&
			i == len(chunks)-1 {
			chunks = append(chunks, treasureChunksMapped[chunks[i].Idx+int64(idxTarget)])
			treasureChunksInserted++
			i++ // skip an iteration
		} else if _, ok := treasureChunksMapped[chunks[i].Idx+int64(idxTarget)]; ok {
			// LOOK INTO THIS
			chunks = append(chunks[:i+2], chunks[i+1:]...)
			chunks[i+1] = treasureChunksMapped[chunks[i].Idx+int64(idxTarget)]
			treasureChunksInserted++
			i++ // skip an iteration
		}
		i++
	}
	return chunks
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

			//addresses, indexes := models.MapChunkIndexesAndAddresses(chunks[i:end])

			oyster_utils.LogToSegment("process_unassigned_chunks: sending_chunks_to_channel", analytics.NewProperties().
				Set("genesis_hash", session.GenesisHash).
				Set("num_chunks", len(chunks[i:end])).
				Set("channel_id", channels[j].ChannelID))
			//Set("addresses", addresses).
			//Set("chunk_indexes", indexes))

			iotaWrapper.SendChunksToChannel(chunks[i:end], &channels[j])
			session.UpdateIndexWithAttachedChunks(chunks[i:end])
		}
		j++
		i += BundleSize
	}
}

// check if a transaction exists for a treasure chunk's address.  If it does, mark it as complete.  If it doesn't,
// we need to separate it from the other chunks so it doesn't get filtered out when VerifyChunkMessagesMatchRecord is
// finding chunks that don't match the tangle.  Later we'll re-add it to the array for sending to the channels.
func HandleTreasureChunks(chunks []oyster_utils.ChunkData, session models.UploadSession,
	iotaWrapper services.IotaService) ([]oyster_utils.ChunkData, []oyster_utils.ChunkData) {

	var chunksToAttach []oyster_utils.ChunkData
	var treasureChunksToAttach []oyster_utils.ChunkData

	treasureIndexes, err := session.GetTreasureIndexes()
	if err != nil {
		oyster_utils.LogIfError(errors.New(err.Error()+" in HandleTreasureChunks() in "+
			"process_unassigned_chunks"), nil)
	}
	if len(chunks) == 0 {
		return chunks, []oyster_utils.ChunkData{}
	}

	maxIdx := int64(math.Max(float64(chunks[0].Idx), float64(chunks[len(chunks)-1].Idx)))
	minIdx := int64(math.Min(float64(chunks[0].Idx), float64(chunks[len(chunks)-1].Idx)))

	treasureMap := make(map[int]bool)
	for _, idx := range treasureIndexes {
		if int64(idx) >= minIdx && int64(idx) <= maxIdx {
			treasureMap[idx] = true
		}
	}

	if len(treasureMap) == 0 {
		return chunks, []oyster_utils.ChunkData{}
	}

	for i := 0; i < len(chunks); i++ {
		if _, ok := treasureMap[int(chunks[i].Idx)]; ok {
			address := make([]giota.Address, 0, 1)
			chunkAddress, err := giota.ToAddress(chunks[i].Address)
			if err != nil {
				oyster_utils.LogIfError(errors.New(err.Error()+" in HandleTreasureChunks() in "+
					"process_unassigned_chunks"), nil)
				return chunks, []oyster_utils.ChunkData{}
			}
			address = append(address, chunkAddress)
			transactionsMap, err := iotaWrapper.FindTransactions(address)
			if err != nil {
				oyster_utils.LogIfError(errors.New(err.Error()+" in HandleTreasureChunks() in "+
					"process_unassigned_chunks"), nil)
				return chunks, []oyster_utils.ChunkData{}
			}
			if _, ok := transactionsMap[chunkAddress]; !ok || transactionsMap == nil {
				oyster_utils.LogToSegment("process_unassigned_chunks: "+
					"treasure_chunk_not_attached", analytics.NewProperties().
					Set("address", chunks[i].Address).
					Set("chunk_index", chunks[i].Idx).
					Set("message", chunks[i].Message))

				treasureChunksToAttach = append(treasureChunksToAttach, chunks[i])
			} else {
				oyster_utils.LogToSegment("process_unassigned_chunks: "+
					"treasure_chunk_already_attached", analytics.NewProperties().
					Set("address", chunks[i].Address).
					Set("chunk_index", chunks[i].Idx).
					Set("message", chunks[i].Message))

				session.MoveChunksToCompleted([]oyster_utils.ChunkData{chunks[i]})
			}
		} else {
			chunksToAttach = append(chunksToAttach, chunks[i])
		}
	}

	return chunksToAttach, treasureChunksToAttach
}

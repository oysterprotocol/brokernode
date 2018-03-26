package services

import (
	"fmt"
	"github.com/iotaledger/giota"
	"github.com/oysterprotocol/brokernode/models"
	"strings"
	"sync"
	"runtime"
	"github.com/getsentry/raven-go"
	"time"
	"math/rand"
)

type ChunkTracker struct {
	ChunkCount  int
	ElapsedTime time.Duration
}

type PowJob struct {
	Chunks         []models.DataMap
	BroadcastNodes []string
}

type PowChannel struct {
	DBChannel     models.ChunkChannel
	ChannelID     string
	ChunkTrackers []ChunkTracker
	Channel       chan PowJob
}

type IotaService struct {
	ProcessChunks                  ProcessChunks
	VerifyChunkMessagesMatchRecord VerifyChunkMessagesMatchRecord
	VerifyChunksMatchRecord        VerifyChunksMatchRecord
	ChunksMatch                    ChunksMatch
}

type ProcessChunks func(chunks []models.DataMap, attachIfAlreadyAttached bool)
type VerifyChunkMessagesMatchRecord func(chunks []models.DataMap) (filteredChunks FilteredChunk, err error)
type VerifyChunksMatchRecord func(chunks []models.DataMap, checkChunkAndBranch bool) (filteredChunks FilteredChunk, err error)
type ChunksMatch func(chunkOnTangle giota.Transaction, chunkOnRecord models.DataMap, checkBranchAndTrunk bool) bool

type FilteredChunk struct {
	MatchesTangle      []models.DataMap
	DoesNotMatchTangle []models.DataMap
	NotAttached        []models.DataMap
}

// Things below are copied from the giota lib since they are not public.
// https://github.com/iotaledger/giota/blob/master/transfer.go#L322
const (
	maxTimestampTrytes = "MMMMMMMMM"
)

var (
	// PowProcs is number of concurrent processes (default is NumCPU()-1)
	PowProcs    int
	IotaWrapper IotaService
	//This mutex was added by us.
	mutex        = &sync.Mutex{}
	seed         giota.Trytes
	provider     = "http://172.21.0.1:14265"
	minDepth     = int64(giota.DefaultNumberOfWalks)
	minWeightMag = int64(14)
	api          = giota.NewAPI(provider, nil)
	bestPow      giota.PowFunc
	powName      string
	letters      = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	Channel      = map[string]PowChannel{}
)

func init() {
	seed = "OYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRL"

	powName, bestPow = giota.GetBestPoW()

	IotaWrapper = IotaService{
		ProcessChunks:                  processChunks,
		VerifyChunkMessagesMatchRecord: verifyChunkMessagesMatchRecord,
		VerifyChunksMatchRecord:        verifyChunksMatchRecord,
		ChunksMatch:                    chunksMatch,
	}

	PowProcs = runtime.NumCPU()
	if PowProcs != 1 {
		PowProcs--
	}

	//makeFakeChunks()

	makeChannels(PowProcs)
}

func makeFakeChunks() {

	dataMaps := []models.DataMap{}

	models.BuildDataMaps("GENHASH2", 100000)

	_ = models.DB.RawQuery("SELECT * from data_maps").All(&dataMaps)

	for i := 0; i < len(dataMaps); i++ {
		dataMaps[i].Address = randSeq(81)
		dataMaps[i].Message = "TESTMESSAGE"
		dataMaps[i].Status = models.Unassigned

		models.DB.ValidateAndSave(&dataMaps[i])
	}
}

func makeChannels(powProcs int) {

	err := models.DB.RawQuery("DELETE from chunk_channels;").All(&[]models.ChunkChannel{})

	if err != nil {
		raven.CaptureError(err, nil)
	}

	for i := 0; i < powProcs; i++ {

		jobQueue := make(chan PowJob)

		var err error;
		newID := randSeq(10)

		channel := models.ChunkChannel{}
		channel.ChannelID = newID
		channel.EstReadyTime = time.Now()
		channel.ChunksProcessed = 0

		_, err = models.DB.ValidateAndSave(&channel)
		if err != nil {
			raven.CaptureError(err, nil)
		}
		models.DB.RawQuery("SELECT * from chunk_channels where channel_id = ?", newID).First(&channel)

		Channel[newID] = PowChannel{
			DBChannel: channel,
			ChannelID: channel.ChannelID,
			Channel:   jobQueue,
		}

		// start the worker
		go PowWorker(Channel[newID].Channel, newID, err)
	}
}

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func PowWorker(jobQueue <-chan PowJob, channelID string, err error) {
	for powJobRequest := range jobQueue {
		// this is where we would call methods to deal with each job request
		fmt.Println("PowWorker: Starting")

		// do filtering/finding BEFORE chunks ever get send to these channels

		transfersArray := make([]giota.Transfer, len(powJobRequest.Chunks))

		tag := "OYSTERGOLANG"

		for i, chunk := range powJobRequest.Chunks {
			transfersArray[i].Address = giota.Address(chunk.Address)
			transfersArray[i].Value = int64(0)
			transfersArray[i].Message = giota.Trytes(chunk.Message)
			transfersArray[i].Tag = giota.Trytes(tag)
		}

		bdl, err := giota.PrepareTransfers(api, seed, transfersArray, nil, "", 1)

		if err != nil {
			raven.CaptureError(err, nil)
			return
		}

		transactions := []giota.Transaction(bdl)

		transactionsToApprove, err := api.GetTransactionsToApprove(minDepth, giota.DefaultNumberOfWalks, "")
		if err != nil {
			raven.CaptureError(err, nil)
			return
		}

		err = doPowAndBroadcast(
			transactionsToApprove.BranchTransaction,
			transactionsToApprove.TrunkTransaction,
			minDepth,
			transactions,
			minWeightMag,
			bestPow,
			powJobRequest.BroadcastNodes)

		fmt.Println("PowWorker: Leaving")
	}
}

func doPowAndBroadcast(branch giota.Trytes, trunk giota.Trytes, depth int64,
	trytes []giota.Transaction, mwm int64, bestPow giota.PowFunc, broadcastNodes []string) error {

	//defer oysterUtils.TimeTrack(time.Now(), "doPow_using_" + powName, analytics.NewProperties().
	//	Set("addresses", oysterUtils.MapTransactionsToAddrs(trytes)))

	var prev giota.Trytes
	var err error

	for i := len(trytes) - 1; i >= 0; i-- {
		switch {
		case i == len(trytes)-1:
			trytes[i].TrunkTransaction = trunk
			trytes[i].BranchTransaction = branch
		default:
			trytes[i].TrunkTransaction = prev
			trytes[i].BranchTransaction = trunk
		}

		timestamp := giota.Int2Trits(time.Now().UnixNano()/1000000, giota.TimestampTrinarySize).Trytes()
		trytes[i].AttachmentTimestamp = timestamp
		trytes[i].AttachmentTimestampLowerBound = ""
		trytes[i].AttachmentTimestampUpperBound = maxTimestampTrytes

		// We customized this to lock here.
		mutex.Lock()
		trytes[i].Nonce, err = bestPow(trytes[i].Trytes(), int(mwm))
		mutex.Unlock()

		if err != nil {
			raven.CaptureError(err, nil)
			return err
		}

		prev = trytes[i].Hash()
	}

	go func(branch giota.Trytes, trunk giota.Trytes, depth int64,
		trytes []giota.Transaction, mwm int64, bestPow giota.PowFunc, broadcastNodes []string) {

		err = api.BroadcastTransactions(trytes)

		if err != nil {

			// Async log
			//go oysterUtils.SegmentClient.Enqueue(analytics.Track{
			//	Event:  "broadcast_fail_redoing_pow",
			//	UserId: oysterUtils.GetLocalIP(),
			//	Properties: analytics.NewProperties().
			//		Set("addresses", oysterUtils.MapTransactionsToAddrs(trytes)),
			//})

			raven.CaptureError(err, nil)
		} else {

			/*
			TODO do we need this??
			 */
			//go BroadcastTxs(&trytes, broadcastNodes)

			//go oysterUtils.SegmentClient.Enqueue(analytics.Track{
			//	Event:  "broadcast_success",
			//	UserId: oysterUtils.GetLocalIP(),
			//	Properties: analytics.NewProperties().
			//		Set("addresses", oysterUtils.MapTransactionsToAddrs(trytes)),
			//})
		}
	}(branch, trunk, depth, trytes, mwm, bestPow, broadcastNodes)

	return nil
}

func processChunks(chunks []models.DataMap, attachIfAlreadyAttached bool) {
	channel := models.ChunkChannel{}

	err := models.DB.RawQuery("SELECT * from chunk_channels WHERE est_ready_time <= ?;", time.Now()).First(&channel)
	if err != nil {
		raven.CaptureError(err, nil)
	}

	powJob := PowJob{
		Chunks: chunks,
		BroadcastNodes: make([]string, 1),
	}

	Channel[channel.ChannelID].Channel <- powJob
}

func verifyChunkMessagesMatchRecord(chunks []models.DataMap) (filteredChunks FilteredChunk, err error) {
	filteredChunks, err = verifyChunksMatchRecord(chunks, false)
	return filteredChunks, err
}

func verifyChunksMatchRecord(chunks []models.DataMap, checkChunkAndBranch bool) (filteredChunks FilteredChunk, err error) {

	addresses := make([]giota.Address, 0, len(chunks))

	for _, chunk := range chunks {
		addresses = append(addresses, giota.Address(chunk.Address))
	}

	request := giota.FindTransactionsRequest{
		Command:   "findTransactions",
		Addresses: addresses,
	}

	response, err := api.FindTransactions(&request)

	if err != nil {
		raven.CaptureError(err, nil)
		return filteredChunks, err
	}

	filteredChunks = FilteredChunk{}

	if response != nil && len(response.Hashes) > 0 {
		trytesArray, err := api.GetTrytes(response.Hashes)
		if err != nil {
			raven.CaptureError(err, nil)
			return filteredChunks, err
		}

		transactionObjects := map[giota.Address][]giota.Transaction{}

		for _, txObject := range trytesArray.Trytes {
			transactionObjects[txObject.Address] = append(transactionObjects[txObject.Address], txObject)
		}

		for _, chunk := range chunks {

			if _, ok := transactionObjects[giota.Address(chunk.Address)]; ok {
				matchFound := false
				for _, txObject := range transactionObjects[giota.Address(chunk.Address)] {
					if chunksMatch(txObject, chunk, checkChunkAndBranch) {
						matchFound = true
						break
					}
				}
				if matchFound {
					filteredChunks.MatchesTangle = append(filteredChunks.MatchesTangle, chunk)
				} else {
					filteredChunks.DoesNotMatchTangle = append(filteredChunks.DoesNotMatchTangle, chunk)
				}
			} else {
				filteredChunks.NotAttached = append(filteredChunks.NotAttached, chunk)
			}
		}
	} else if len(response.Hashes) == 0 {
		filteredChunks.NotAttached = chunks
	}
	return filteredChunks, err
}

func chunksMatch(chunkOnTangle giota.Transaction, chunkOnRecord models.DataMap, checkBranchAndTrunk bool) bool {

	if checkBranchAndTrunk == false &&
		strings.Contains(fmt.Sprint(chunkOnTangle.SignatureMessageFragment), chunkOnRecord.Message) == true {

		return true

	} else if strings.Contains(fmt.Sprint(chunkOnTangle.SignatureMessageFragment), chunkOnRecord.Message) == true &&
		strings.Contains(fmt.Sprint(chunkOnTangle.TrunkTransaction), chunkOnRecord.TrunkTx) &&
		strings.Contains(fmt.Sprint(chunkOnTangle.BranchTransaction), chunkOnRecord.BranchTx) {

		return true

	} else {

		return false
	}
}

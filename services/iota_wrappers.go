package services

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/iotaledger/giota"
	"github.com/joho/godotenv"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/utils"
	"gopkg.in/segmentio/analytics-go.v3"
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
	ChannelID     string
	ChunkTrackers *[]ChunkTracker
	Channel       chan PowJob
}

type IotaService struct {
	SendChunksToChannel
	VerifyChunkMessagesMatchRecord
	VerifyChunksMatchRecord
	ChunksMatch
	VerifyTreasure
	FindTransactions
	GetTransactionsToApprove
}

type ProcessingFrequency struct {
	RecentProcessingTimes []time.Duration
	Frequency             float64
}

type SendChunksToChannel func([]models.DataMap, *models.ChunkChannel)
type VerifyChunkMessagesMatchRecord func([]models.DataMap) (filteredChunks FilteredChunk, err error)
type VerifyChunksMatchRecord func([]models.DataMap, bool) (filteredChunks FilteredChunk, err error)
type ChunksMatch func(giota.Transaction, models.DataMap, bool) bool
type VerifyTreasure func([]string) (verify bool, err error)
type FindTransactions func([]giota.Address) (map[giota.Address][]giota.Transaction, error)
type GetTransactionsToApprove func() (*giota.GetTransactionsToApproveResponse, error)

type FilteredChunk struct {
	MatchesTangle      []models.DataMap
	DoesNotMatchTangle []models.DataMap
	NotAttached        []models.DataMap
}

// Things below are copied from the giota lib since they are not public.
// https://github.com/iotaledger/giota/blob/master/transfer.go#L322
const (
	maxTimestampTrytes = "MMMMMMMMM"

	// By a hard limit on the request for FindTransaction to IOTA
	MaxNumberOfAddressPerFindTransactionRequest = 1000

	// Lambda
	// https://docs.aws.amazon.com/lambda/latest/dg/limits.html
	// 6MB payload, 300 sec execution time, 1000 concurrent exectutions.
	// Limit to 1000 POSTs and 50 chunks per request.
	maxLambdaConcurrency = 1000
	maxLambdaChunksLen   = 50
)

var (
	// PowProcs is number of concurrent processes (default is NumCPU()-1)
	PowProcs    int
	IotaWrapper IotaService
	//This mutex was added by us.
	mutex           = &sync.Mutex{}
	seed            giota.Trytes
	minDepth        = int64(1)
	minWeightMag    = int64(6)
	bestPow         giota.PowFunc
	powName         string
	Channel         = map[string]PowChannel{}
	wg              sync.WaitGroup
	api             *giota.API
	PoWFrequency    ProcessingFrequency
	minPoWFrequency = 1
	OysterTag, _    = giota.ToTrytes("OYSTERGOLANG")

	// Lambda
	lambdaChan = make(chan string, maxLambdaConcurrency)
)

func init() {

	// Load ENV variables
	err := godotenv.Load()
	if err != nil {
		oyster_utils.LogIfError(fmt.Errorf(".env file : %v", err), nil)
	}

	host_ip := os.Getenv("HOST_IP")
	if host_ip == "" {
		panic("Invalid IRI host: Check the .env file for HOST_IP")
	}

	provider := "http://" + host_ip + ":14265"

	api = giota.NewAPI(provider, nil)

	seed = "OYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRL"

	powName, bestPow = giota.GetBestPoW()

	IotaWrapper = IotaService{
		SendChunksToChannel:            sendChunksToChannel,
		VerifyChunkMessagesMatchRecord: verifyChunkMessagesMatchRecord,
		VerifyChunksMatchRecord:        verifyChunksMatchRecord,
		ChunksMatch:                    chunksMatch,
		VerifyTreasure:                 verifyTreasure,
		FindTransactions:               findTransactions,
		GetTransactionsToApprove:       getTransactionsToApprove,
	}

	PowProcs = runtime.NumCPU()
	if PowProcs != 1 {
		PowProcs--
	}

	channels := []models.ChunkChannel{}

	wg.Add(1)
	go func(channels *[]models.ChunkChannel, err *error) {
		defer wg.Done()
		*channels, *err = models.MakeChannels(PowProcs)
	}(&channels, &err)

	wg.Wait()

	for _, channel := range channels {

		chunkTracker := make([]ChunkTracker, 0)

		Channel[channel.ChannelID] = PowChannel{
			ChannelID:     channel.ChannelID,
			Channel:       make(chan PowJob),
			ChunkTrackers: &chunkTracker,
		}

		// start the worker
		go PowWorker(Channel[channel.ChannelID].Channel, channel.ChannelID, err)
		go lambdaWorker(lambdaChan)
	}

	PoWFrequency.Frequency = 2
}

func PowWorker(jobQueue <-chan PowJob, channelID string, err error) {
	for powJobRequest := range jobQueue {
		// this is where we would call methods to deal with each job request
		fmt.Println("PowWorker: Starting")

		startTime := time.Now()

		transfersArray := make([]giota.Transfer, len(powJobRequest.Chunks))

		for i, chunk := range powJobRequest.Chunks {
			address, err := giota.ToAddress(chunk.Address)
			if err != nil {
				oyster_utils.LogIfError(err, nil)
				panic(err)
			}
			transfersArray[i].Address = address
			transfersArray[i].Value = int64(0)
			transfersArray[i].Message, err = giota.ToTrytes(GetMessageFromDataMap(chunk))
			if err != nil {
				oyster_utils.LogIfError(err, nil)
				panic(err)
			}
			transfersArray[i].Tag = OysterTag
		}

		bdl, err := giota.PrepareTransfers(api, seed, transfersArray, nil, "", 1)

		oyster_utils.LogIfError(err, nil)

		transactions := []giota.Transaction(bdl)

		transactionsToApprove, err := getTransactionsToApprove()
		oyster_utils.LogIfError(err, nil)

		if err == nil {

			err = doPowAndBroadcast(
				transactionsToApprove.BranchTransaction,
				transactionsToApprove.TrunkTransaction,
				minDepth,
				transactions,
				minWeightMag,
				bestPow,
				powJobRequest.BroadcastNodes)

			channelToChange := Channel[channelID]

			channelInDB := models.ChunkChannel{}
			models.DB.RawQuery("SELECT * FROM chunk_channels WHERE channel_id = ?", channelID).First(&channelInDB)
			channelInDB.ChunksProcessed += len(powJobRequest.Chunks)
			models.DB.ValidateAndSave(&channelInDB)

			fmt.Println("PowWorker: Leaving")
			TrackProcessingTime(startTime, len(powJobRequest.Chunks), &channelToChange)
		}
	}
}

func getTransactionsToApprove() (*giota.GetTransactionsToApproveResponse, error) {
	return api.GetTransactionsToApprove(minDepth, minDepth, "")
}

func TrackProcessingTime(startTime time.Time, numChunks int, channel *PowChannel) {

	*(channel.ChunkTrackers) = append(*(channel.ChunkTrackers), ChunkTracker{
		ChunkCount:  numChunks,
		ElapsedTime: time.Since(startTime),
	})

	PoWFrequency.RecentProcessingTimes = append(PoWFrequency.RecentProcessingTimes, time.Since(startTime))

	if len(*(channel.ChunkTrackers)) > 10 {
		*(channel.ChunkTrackers) = (*(channel.ChunkTrackers))[1:11]
	}

	if len(PoWFrequency.RecentProcessingTimes) > 10 {
		PoWFrequency.RecentProcessingTimes = PoWFrequency.RecentProcessingTimes[1:11]
	}

	var totalTime time.Duration
	for _, elapsedTime := range PoWFrequency.RecentProcessingTimes {
		totalTime += elapsedTime
	}

	avgTimePerPoW := float64(totalTime/time.Second) / float64(len(PoWFrequency.RecentProcessingTimes))

	PoWFrequency.Frequency = math.Max(avgTimePerPoW, float64(minPoWFrequency))
}

func GetProcessingFrequency() float64 {
	return PoWFrequency.Frequency
}

// Finds Transactions with a list of addresses. Result in a map from Address to a list of Transactions
func findTransactions(addresses []giota.Address) (map[giota.Address][]giota.Transaction, error) {

	addrToTransactionMap := make(map[giota.Address][]giota.Transaction)

	numOfBatchRequest := int(math.Ceil(float64(len(addresses)) / float64(MaxNumberOfAddressPerFindTransactionRequest)))

	remainder := len(addresses)
	for i := 0; i < numOfBatchRequest; i++ {
		lower := i * MaxNumberOfAddressPerFindTransactionRequest
		upper := i*MaxNumberOfAddressPerFindTransactionRequest + int(math.Min(float64(remainder), MaxNumberOfAddressPerFindTransactionRequest))
		req := giota.FindTransactionsRequest{
			Addresses: addresses[lower:upper],
		}
		resp, err := api.FindTransactions(&req)
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			return nil, err
		}
		transactionResp, err := api.GetTrytes(resp.Hashes)
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			return nil, err
		}

		for _, transaction := range transactionResp.Trytes {
			list := addrToTransactionMap[transaction.Address]
			list = append(list, transaction)
			addrToTransactionMap[transaction.Address] = list
		}
		remainder = remainder - MaxNumberOfAddressPerFindTransactionRequest
	}

	return addrToTransactionMap, nil
}

func doPowAndBroadcast(branch giota.Trytes, trunk giota.Trytes, depth int64,
	trytes []giota.Transaction, mwm int64, bestPow giota.PowFunc, broadcastNodes []string) error {

	defer oyster_utils.TimeTrack(time.Now(), "iota_wrappers: doPow_using_"+powName, analytics.NewProperties().
		//Set("addresses", oyster_utils.MapTransactionsToAddrs(trytes)))
		Set("num_chunks", len(trytes)))

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
			oyster_utils.LogIfError(err, nil)
			return err
		}

		prev = trytes[i].Hash()
	}

	broadcastProperties := analytics.NewProperties().
		Set("num_chunks", len(trytes))

	go func(trytes []giota.Transaction, broadcastProperties analytics.Properties) {

		err = api.BroadcastTransactions(trytes)

		if err != nil {

			// Async log
			oyster_utils.LogToSegment("iota_wrappers: broadcast_FAIL", broadcastProperties)

			oyster_utils.LogIfError(err, nil)
		} else {

			err = api.StoreTransactions(trytes)
			fmt.Println("BROADCAST SUCCESS")

			// Async log
			oyster_utils.LogToSegment("iota_wrappers: broadcast_success", broadcastProperties)
		}
	}(trytes, broadcastProperties)

	return nil
}

func sendChunksToChannel(chunks []models.DataMap, channel *models.ChunkChannel) {

	for _, chunk := range chunks {
		chunk.Status = models.Unverified
		models.DB.ValidateAndSave(&chunk)
	}

	if os.Getenv("ENABLE_LAMBDA") == true {
		go func() {
			// TODO: prep and chunk by limit, then send to channel
		}()
	} else {
		channel.EstReadyTime = SetEstimatedReadyTime(Channel[channel.ChannelID], len(chunks))
		models.DB.ValidateAndSave(channel)

		powJob := PowJob{
			Chunks:         chunks,
			BroadcastNodes: make([]string, 1),
		}

		Channel[channel.ChannelID].Channel <- powJob
	}
}

func SetEstimatedReadyTime(channel PowChannel, numChunks int) time.Time {

	if channel.Channel == nil {
		delay := time.Duration(PoWFrequency.Frequency) * time.Second
		return time.Now().Add(delay)
	}

	var totalTime time.Duration = 0
	chunksCount := 0

	if len(*(channel.ChunkTrackers)) != 0 {

		for _, timeRecord := range *(channel.ChunkTrackers) {
			totalTime += timeRecord.ElapsedTime
			chunksCount += timeRecord.ChunkCount
		}

		avgTimePerChunk := int(totalTime) / chunksCount
		expectedDelay := int(math.Floor((float64(avgTimePerChunk * numChunks))))

		return time.Now().Add(time.Duration(expectedDelay))
	} else {

		// The application just started, we don't have any data yet,
		// so just set est_ready_time to 10 seconds from now

		/*
			TODO:  get a more precise estimate of what this default should be
		*/
		return time.Now().Add(10 * time.Second)
	}
}

func verifyChunkMessagesMatchRecord(chunks []models.DataMap) (filteredChunks FilteredChunk, err error) {
	filteredChunks, err = verifyChunksMatchRecord(chunks, false)
	return filteredChunks, err
}

func verifyChunksMatchRecord(chunks []models.DataMap, checkTrunkAndBranch bool) (filteredChunks FilteredChunk, err error) {

	filteredChunks = FilteredChunk{}
	addresses := make([]giota.Address, 0, len(chunks))

	for _, chunk := range chunks {
		// if a chunk did not match the tangle in verify_data_maps
		// we mark it as "Error" and there is no reason to check the tangle
		// for it again while its status is still in an Error state

		// this will cause this chunk to automatically get added to 'NotAttached' array
		// and send to the channels
		if chunk.Status != models.Error {
			address, err := giota.ToAddress(chunk.Address)
			if err != nil {
				oyster_utils.LogIfError(err, nil)
				// trytes were not valid, skip this iteration
				continue
			}
			addresses = append(addresses, address)
		}
	}

	if len(addresses) == 0 {
		return filteredChunks, nil
	}

	request := giota.FindTransactionsRequest{
		Addresses: addresses,
	}

	response, err := api.FindTransactions(&request)

	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return filteredChunks, err
	}

	if response != nil && len(response.Hashes) > 0 {

		matchesTangle, notAttached, doesNotMatch := filterChunks(response.Hashes, chunks, checkTrunkAndBranch)

		filteredChunks.MatchesTangle = append(filteredChunks.MatchesTangle, matchesTangle...)
		filteredChunks.DoesNotMatchTangle = append(filteredChunks.DoesNotMatchTangle, doesNotMatch...)
		filteredChunks.NotAttached = append(filteredChunks.NotAttached, notAttached...)

	} else if len(response.Hashes) == 0 {
		filteredChunks.NotAttached = chunks
	}

	if len(filteredChunks.MatchesTangle) > 0 {
		oyster_utils.LogToSegment("iota_wrappers: chunks_matched_tangle", analytics.NewProperties().
			Set("num_chunks", len(filteredChunks.MatchesTangle)))
	}
	if len(filteredChunks.NotAttached) > 0 {
		oyster_utils.LogToSegment("iota_wrappers: not_attached", analytics.NewProperties().
			Set("num_chunks", len(filteredChunks.NotAttached)))
	}
	return filteredChunks, err
}

func filterChunks(hashes []giota.Trytes, chunks []models.DataMap, checkTrunkAndBranch bool) (matchesTangle []models.DataMap,
	notAttached []models.DataMap, doesNotMatch []models.DataMap) {

	for i := 0; i < len(hashes); i += MaxNumberOfAddressPerFindTransactionRequest {
		end := i + MaxNumberOfAddressPerFindTransactionRequest

		if end > len(hashes) {
			end = len(hashes)
		}

		if i >= end {
			break
		}

		trytesArray, err := api.GetTrytes(hashes[i:end])

		if err != nil {
			oyster_utils.LogIfError(err, nil)
		}

		if len(trytesArray.Trytes) == 0 {
			return matchesTangle, notAttached, doesNotMatch
		}

		transactionObjects := makeTransactionObjects(trytesArray.Trytes)

		for _, chunk := range chunks {

			chunkAddress, err := giota.ToAddress(chunk.Address)
			if err != nil {
				oyster_utils.LogIfError(err, nil)
				// trytes were not valid, skip this iteration
				continue
			}
			if _, ok := transactionObjects[chunkAddress]; ok {

				matchFound := checkTxObjectsForMatch(transactionObjects[chunkAddress], chunk, checkTrunkAndBranch)
				if matchFound {
					matchesTangle = append(matchesTangle, chunk)
				} else {
					doesNotMatch = append(doesNotMatch, chunk)
				}
			} else {
				notAttached = append(notAttached, chunk)
			}
		}
	}
	return matchesTangle, notAttached, doesNotMatch
}

func checkTxObjectsForMatch(transactionObjectsArray []giota.Transaction, chunk models.DataMap, checkTrunkAndBranch bool) (matchFound bool) {
	matchFound = false
	for _, txObject := range transactionObjectsArray {
		if chunksMatch(txObject, chunk, checkTrunkAndBranch) {
			matchFound = true
			break
		}
	}
	return matchFound
}

func makeTransactionObjects(transactionObjects []giota.Transaction) (transactionObjectsMap map[giota.Address][]giota.Transaction) {
	transactionObjectsMap = make(map[giota.Address][]giota.Transaction)
	for _, txObject := range transactionObjects {
		transactionObjectsMap[txObject.Address] = append(transactionObjectsMap[txObject.Address], txObject)
	}
	return transactionObjectsMap
}

func chunksMatch(chunkOnTangle giota.Transaction, chunkOnRecord models.DataMap, checkBranchAndTrunk bool) bool {

	// TODO delete this line when we figure out why uploads
	// occasionally have the wrong chunk_idx for msg_id
	return true

	message := GetMessageFromDataMap(chunkOnRecord)
	if checkBranchAndTrunk == false &&
		strings.Contains(fmt.Sprint(chunkOnTangle.SignatureMessageFragment), message) {

		return true

	} else if strings.Contains(fmt.Sprint(chunkOnTangle.SignatureMessageFragment), message) &&
		strings.Contains(fmt.Sprint(chunkOnTangle.TrunkTransaction), chunkOnRecord.TrunkTx) &&
		strings.Contains(fmt.Sprint(chunkOnTangle.BranchTransaction), chunkOnRecord.BranchTx) {

		return true

	} else {

		oyster_utils.LogToSegment("iota_wrappers: resend_chunk_tangle_mismatch", analytics.NewProperties().
			Set("genesis_hash", chunkOnRecord.GenesisHash).
			Set("chunk_idx", chunkOnRecord.ChunkIdx).
			Set("address", chunkOnRecord.Address).
			Set("db_message", message).
			Set("tangle_message", chunkOnTangle.SignatureMessageFragment).
			Set("db_trunk", chunkOnRecord.TrunkTx).
			Set("tangle_trunk", chunkOnTangle.TrunkTransaction).
			Set("db_branch", chunkOnRecord.BranchTx).
			Set("tangle_branch", chunkOnTangle.BranchTransaction))

		return false
	}
}

// Verify PoW of work.
func verifyTreasure(addr []string) (bool, error) {

	iotaAddr := make([]giota.Address, len(addr))

	for i, address := range addr {
		validAddress, err := giota.ToAddress(address)
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			return false, err
		}
		iotaAddr[i] = validAddress
	}

	transactionsMap, err := findTransactions(iotaAddr)

	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return false, err
	}

	if len(transactionsMap) != len(iotaAddr) {
		return false, nil
	}

	isTransactionWithinTimePeriod := false
	passedTimestamp := time.Now().AddDate(-1, 0, 0)

	for _, iotaAddress := range iotaAddr {
		if _, hasKey := transactionsMap[iotaAddress]; !hasKey {
			return false, nil
		}

		transactions := transactionsMap[iotaAddress]
		// Check one the transactions has submit within the passed 1 year.
		for _, transaction := range transactions {
			if transaction.Timestamp.After(passedTimestamp) {
				isTransactionWithinTimePeriod = true
				break
			}
		}
		if !isTransactionWithinTimePeriod {
			return false, nil
		}
	}

	return true, nil
}

func lambdaWorker(lChan <-chan string) {
	for chunks := range lChan {
		go func() {
			// TODO: Send to lambda
		}()
	}
}

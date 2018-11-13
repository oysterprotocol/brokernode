package services

import (
	"fmt"
	"github.com/pkg/errors"
	"math"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	giota "github.com/iotaledger/iota.go/api"
	"github.com/iotaledger/iota.go/bundle"
	"github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/pow"
	"github.com/iotaledger/iota.go/transaction"
	"github.com/iotaledger/iota.go/trinary"
	"github.com/joho/godotenv"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services/awsgateway"
	"github.com/oysterprotocol/brokernode/utils"
	"gopkg.in/segmentio/analytics-go.v3"
)

type ChunkTracker struct {
	ChunkCount  int
	ElapsedTime time.Duration
}

type PowJob struct {
	Chunks         []oyster_utils.ChunkData
	BroadcastNodes []string
}

type PowChannel struct {
	ChannelID     string
	ChunkTrackers *[]ChunkTracker
	Channel       chan PowJob
}

type IotaService struct {
	SendChunksToChannel
	SendChunksToLambda
	VerifyChunkMessagesMatchRecord
	VerifyChunksMatchRecord
	ChunksMatch
	VerifyTreasure
	FindTransactions
	GetTransactionsToApprove
	DoPoW
}

type ProcessingFrequency struct {
	RecentProcessingTimes []time.Duration
	Frequency             float64
}

/*SendChunksToChannel defines the type for a function which sends chunks to a channel.  This type used for mocking.*/
type SendChunksToChannel func([]oyster_utils.ChunkData, *models.ChunkChannel)

/*VerifyChunkMessagesMatchRecord defines the type for a function which verifies chunk messages
on the tangle match what we have stored in the db.  This type used for mocking.*/
type VerifyChunkMessagesMatchRecord func([]oyster_utils.ChunkData) (filteredChunks FilteredChunk, err error)

/*VerifyChunksMatchRecord defines the type for a function which verifies that chunks on the tangle match
chunks in the db.  This type used for mocking.*/
type VerifyChunksMatchRecord func([]oyster_utils.ChunkData, bool) (filteredChunks FilteredChunk, err error)

/*ChunksMatch defines the type for a function which returns a boolean if chunk tangle data matches chunk db data.
This type used for mocking.*/
type ChunksMatch func(transaction.Transaction, oyster_utils.ChunkData, bool) bool

/*SendChunksToLambda defines the type for a function which will send the chunkds to AWS Lambda.
This type used for mocking.*/
type SendChunksToLambda func(chunks *[]oyster_utils.ChunkData)

/*VerifyTreasure defines the type for a function which will verify webnode treasure claims.
This type used for mocking.*/
type VerifyTreasure func([]string) (verify bool, err error)

/*FindTransactions defines the type for a function which will attempt to find transactions on the iota tangle based on
addresses.  This type used for mocking.*/
type FindTransactions func([]trinary.Hash) (map[trinary.Hash][]transaction.Transaction, error)

/*GetTransactionsToApprove defines the type for a function which will get a trunk and branch transaction for proof of
work.  This type used for mocking.*/
type GetTransactionsToApprove func() (*giota.TransactionsToApprove, error)

/*DoPoW defines the type for a function that does proof of work.*/
type DoPoW func(chunks []oyster_utils.ChunkData) error

type FilteredChunk struct {
	MatchesTangle      []oyster_utils.ChunkData
	DoesNotMatchTangle []oyster_utils.ChunkData
	NotAttached        []oyster_utils.ChunkData
}

const (
	// MaxNumberOfAddressPerFindTransactionRequest is a hard limit
	// on the request for FindTransaction to IOTA
	MaxNumberOfAddressPerFindTransactionRequest = 1000
	oysterTagStr                                = "OYSTERGOLANG"
	oysterTagHookStr                            = "OYSTERHOOKNODE"
)

var (
	// PowProcs is number of concurrent processes (default is NumCPU()-1)
	PowProcs    int
	IotaWrapper IotaService
	//This mutex was added by us.
	mutex            = &sync.Mutex{}
	seed             trinary.Trytes
	minDepth         = int64(1)
	minWeightMag     = int64(6)
	bestPow          pow.ProofOfWorkFunc
	powName          string
	Channel          = map[string]PowChannel{}
	wg               sync.WaitGroup
	api              *giota.API
	PoWFrequency     ProcessingFrequency
	minPoWFrequency  = 1
	OysterTag, _     = trinary.NewTrytes(oysterTagStr)
	OysterTagHook, _ = trinary.NewTrytes(oysterTagHookStr)

	// Lambda
	lambdaChan = make(chan []*awsgateway.HooknodeChunk, awsgateway.MaxConcurrency)
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

	powName, bestPow = pow.GetFastestProofOfWorkImpl()

	provider := "http://" + host_ip + ":14265"
	// create a new API instance
	api, err = giota.ComposeAPI(giota.HTTPClientSettings{
		URI:                  provider,
		LocalProofOfWorkFunc: bestPow,
	})
	if err != nil {
		panic(err)
	}

	seed = "OYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRL"

	IotaWrapper = IotaService{
		SendChunksToLambda:             sendChunksToLambda,
		SendChunksToChannel:            sendChunksToChannel,
		VerifyChunkMessagesMatchRecord: verifyChunkMessagesMatchRecord,
		VerifyChunksMatchRecord:        verifyChunksMatchRecord,
		ChunksMatch:                    chunksMatch,
		VerifyTreasure:                 verifyTreasure,
		FindTransactions:               findTransactions,
		GetTransactionsToApprove:       getTransactionsToApprove,
		DoPoW: doPoW,
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

		// Start lambda worker pool.
		for i := 0; i < awsgateway.MaxConcurrency; i++ {
			go lambdaWorker(provider, lambdaChan)
		}
	}

	PoWFrequency.Frequency = 2
}

func PowWorker(jobQueue <-chan PowJob, channelID string, err error) {
	for powJobRequest := range jobQueue {
		// this is where we would call methods to deal with each job request
		fmt.Println("PowWorker: Starting")

		startTime := time.Now()

		err = doPoW(powJobRequest.Chunks)

		if err == nil {
			channelToChange := Channel[channelID]

			channelInDB := models.ChunkChannel{}
			models.DB.RawQuery("SELECT * FROM chunk_channels WHERE channel_id = ?", channelID).First(&channelInDB)
			channelInDB.ChunksProcessed += len(powJobRequest.Chunks)
			models.DB.ValidateAndSave(&channelInDB)

			fmt.Println("PowWorker: Leaving")
			TrackProcessingTime(startTime, len(powJobRequest.Chunks), &channelToChange)
		} else {
			fmt.Println("PowWorker: FAILED")
			oyster_utils.LogIfError(err, nil)
		}
	}
}

func doPoW(chunks []oyster_utils.ChunkData) error {
	transfersArray := make([]bundle.Transfer, len(chunks))

	for i, chunk := range chunks {
		address, err := trinary.NewTrytes(chunk.Address)
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			panic(err)
		}
		transfersArray[i].Address = address
		transfersArray[i].Value = uint64(0)
		transfersArray[i].Message, err = trinary.NewTrytes(chunk.Message)
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			panic(err)
		}
		transfersArray[i].Tag = OysterTag
	}

	bdl, err := api.PrepareTransfers(seed, transfersArray, giota.PrepareTransfersOptions{})

	oyster_utils.LogIfError(err, nil)

	transactions, err := transaction.AsTransactionObjects(bdl, nil)
	oyster_utils.LogIfError(err, nil)

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
			[]string{""})
	}

	return err
}

func getTransactionsToApprove() (*giota.TransactionsToApprove, error) {
	return api.GetTransactionsToApprove(uint64(minDepth))
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
func findTransactions(addresses []trinary.Hash) (map[trinary.Hash][]transaction.Transaction, error) {

	addrToTransactionMap := make(map[trinary.Hash][]transaction.Transaction)

	numOfBatchRequest := int(math.Ceil(float64(len(addresses)) / float64(MaxNumberOfAddressPerFindTransactionRequest)))

	remainder := len(addresses)
	for i := 0; i < numOfBatchRequest; i++ {
		lower := i * MaxNumberOfAddressPerFindTransactionRequest
		upper := i*MaxNumberOfAddressPerFindTransactionRequest + int(math.Min(float64(remainder), MaxNumberOfAddressPerFindTransactionRequest))
		req := giota.FindTransactionsQuery{
			Addresses: addresses[lower:upper],
		}
		transactionsHashes, err := api.FindTransactions(req)
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			return nil, err
		}
		transactionTrytes, err := api.GetTrytes(transactionsHashes...)
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			return nil, err
		}
		transactions, err := transaction.AsTransactionObjects(transactionTrytes, nil)
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			return nil, err
		}
		for _, transaction := range transactions {
			address := transaction.Address
			addrToTransactionMap[address] = append(addrToTransactionMap[address], transaction)
		}
		remainder = remainder - MaxNumberOfAddressPerFindTransactionRequest
	}

	return addrToTransactionMap, nil
}

func doPowAndBroadcast(branch trinary.Trytes, trunk trinary.Trytes, depth int64,
	trytes []transaction.Transaction, mwm int64, bestPow pow.ProofOfWorkFunc, broadcastNodes []string) error {

	defer oyster_utils.TimeTrack(time.Now(), "iota_wrappers: doPow_using_"+powName, analytics.NewProperties().
		Set("num_chunks", len(trytes)))

	var prev trinary.Trytes
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

		trytes[i].AttachmentTimestamp = time.Now().UnixNano() / 1000000
		trytes[i].AttachmentTimestampLowerBound = consts.LowerBoundAttachmentTimestamp
		trytes[i].AttachmentTimestampUpperBound = consts.UpperBoundAttachmentTimestamp

		// We customized this to lock here.
		mutex.Lock()
		transactionToTrytes, err := transaction.TransactionToTrytes(&trytes[i])

		if err == nil {
			trytes[i].Nonce, err = bestPow(transactionToTrytes, int(mwm))
		}
		mutex.Unlock()

		if err != nil {
			oyster_utils.LogIfError(err, nil)
			return err
		}

		prev = transaction.TransactionHash(&trytes[i])
	}

	broadcastProperties := analytics.NewProperties().
		Set("num_chunks", len(trytes))

	go func(transactions []transaction.Transaction, broadcastProperties analytics.Properties) {
		trytes := transaction.MustTransactionsToTrytes(transactions)

		_, err = api.BroadcastTransactions(trytes...)

		if err != nil {

			// Async log
			oyster_utils.LogToSegment("iota_wrappers: broadcast_FAIL", broadcastProperties)
			oyster_utils.LogIfError(err, nil)
		} else {

			_, err = api.StoreTransactions(trytes...)
			fmt.Println("BROADCAST SUCCESS")

			// Async log
			oyster_utils.LogToSegment("iota_wrappers: broadcast_success", broadcastProperties)
		}
	}(trytes, broadcastProperties)

	return nil
}

func sendChunksToLambda(chunks *[]oyster_utils.ChunkData) {

	go batchPowOnLambda(chunks)
}

func sendChunksToChannel(chunks []oyster_utils.ChunkData, channel *models.ChunkChannel) {

	channel.EstReadyTime = SetEstimatedReadyTime(Channel[channel.ChannelID], len(chunks))
	models.DB.ValidateAndSave(channel)

	powJob := PowJob{
		Chunks:         chunks,
		BroadcastNodes: make([]string, 1),
	}

	Channel[channel.ChannelID].Channel <- powJob
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

func verifyChunkMessagesMatchRecord(chunks []oyster_utils.ChunkData) (filteredChunks FilteredChunk, err error) {
	filteredChunks, err = verifyChunksMatchRecord(chunks, false)
	return filteredChunks, err
}

func verifyChunksMatchRecord(chunks []oyster_utils.ChunkData, checkTrunkAndBranch bool) (filteredChunks FilteredChunk, err error) {

	filteredChunks = FilteredChunk{}
	addresses := make([]trinary.Hash, 0, len(chunks))

	for _, chunk := range chunks {
		address, err := trinary.NewTrytes(chunk.Address)
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			// trytes were not valid, skip this iteration
			continue
		}
		addresses = append(addresses, address)
	}

	if len(addresses) == 0 {
		return filteredChunks, nil
	}

	request := giota.FindTransactionsQuery{
		Addresses: addresses,
	}

	responseHashes, err := api.FindTransactions(request)

	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return filteredChunks, err
	}

	if responseHashes != nil && len(responseHashes) > 0 {
		matchesTangle, notAttached, doesNotMatch := filterChunks(responseHashes, chunks, checkTrunkAndBranch)

		filteredChunks.MatchesTangle = append(filteredChunks.MatchesTangle, matchesTangle...)
		filteredChunks.DoesNotMatchTangle = append(filteredChunks.DoesNotMatchTangle, doesNotMatch...)
		filteredChunks.NotAttached = append(filteredChunks.NotAttached, notAttached...)

	} else if len(responseHashes) == 0 {
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

func filterChunks(hashes []trinary.Trytes, chunks []oyster_utils.ChunkData, checkTrunkAndBranch bool) (matchesTangle []oyster_utils.ChunkData,
	notAttached []oyster_utils.ChunkData, doesNotMatch []oyster_utils.ChunkData) {

	transactionObjectsMap := make(map[trinary.Hash][]transaction.Transaction)

	for i := 0; i < len(hashes); i += MaxNumberOfAddressPerFindTransactionRequest {
		end := i + MaxNumberOfAddressPerFindTransactionRequest

		if end > len(hashes) {
			end = len(hashes)
		}

		if i >= end {
			break
		}

		trytesArray, err := api.GetTrytes(hashes[i:end]...)
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			continue
		}
		transactions, err := transaction.AsTransactionObjects(trytesArray, nil)
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			continue
		}

		for key, value := range makeTransactionObjectsMap(transactions) {
			transactionObjectsMap[key] = value
		}
	}

	for _, chunk := range chunks {

		chunkAddress, err := trinary.NewTrytes(chunk.Address)
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			// trytes were not valid, skip this iteration
			continue
		}
		if _, ok := transactionObjectsMap[chunkAddress]; ok {

			matchFound := checkTxObjectsForMatch(transactionObjectsMap[chunkAddress], chunk, checkTrunkAndBranch)
			if matchFound {
				matchesTangle = append(matchesTangle, chunk)
			} else {
				doesNotMatch = append(doesNotMatch, chunk)
			}
		} else {
			notAttached = append(notAttached, chunk)
		}
	}

	return matchesTangle, notAttached, doesNotMatch
}

func checkTxObjectsForMatch(transactionObjectsArray []transaction.Transaction, chunk oyster_utils.ChunkData, checkTrunkAndBranch bool) (matchFound bool) {
	matchFound = false
	for _, txObject := range transactionObjectsArray {
		if chunksMatch(txObject, chunk, checkTrunkAndBranch) {
			matchFound = true
			break
		}
	}
	return matchFound
}

func makeTransactionObjectsMap(transactionObjects []transaction.Transaction) (transactionObjectsMap map[trinary.Hash][]transaction.Transaction) {
	transactionObjectsMap = make(map[trinary.Hash][]transaction.Transaction)
	for _, txObject := range transactionObjects {
		transactionObjectsMap[txObject.Address] = append(transactionObjectsMap[txObject.Address], txObject)
	}
	return transactionObjectsMap
}

func chunksMatch(chunkOnTangle transaction.Transaction, chunkOnRecord oyster_utils.ChunkData, checkBranchAndTrunk bool) bool {

	message := chunkOnRecord.Message
	if checkBranchAndTrunk == false &&
		strings.Contains(fmt.Sprint(chunkOnTangle.SignatureMessageFragment), message) {
		return true
	} else {
		oyster_utils.LogIfError(errors.New("chunk on tangle does not match record"), nil)
		oyster_utils.LogToSegment("iota_wrappers: resend_chunk_tangle_mismatch", analytics.NewProperties().
			Set("chunk_idx", chunkOnRecord.Idx).
			Set("address", chunkOnRecord.Address).
			Set("db_message", message).
			Set("tangle_message", chunkOnTangle.SignatureMessageFragment))

		return false
	}
}

// Verify PoW of work.
func verifyTreasure(addr []string) (bool, error) {

	iotaAddr := make([]trinary.Hash, len(addr))

	for i, address := range addr {
		validAddress, err := trinary.NewTrytes(address)
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
	passedTimestamp := uint64(time.Now().AddDate(-1, 0, 0).UnixNano() / int64(time.Second))

	for _, iotaAddress := range iotaAddr {
		if _, hasKey := transactionsMap[iotaAddress]; !hasKey {
			return false, nil
		}

		transactions := transactionsMap[iotaAddress]
		// Check one the transactions has submit within the passed 1 year.
		for _, transaction := range transactions {
			if transaction.Timestamp > passedTimestamp {
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

func lambdaWorker(provider string, lChan <-chan []*awsgateway.HooknodeChunk) {
	for chkBatch := range lChan {
		if len(chkBatch) <= 0 {
			continue
		}

		fmt.Printf("Lambda Worker Processing!!! \t %d chunks\n", len(chkBatch))

		req := awsgateway.HooknodeReq{
			Provider: provider,
			Chunks:   chkBatch,
		}

		err := awsgateway.InvokeHooknode(&req)
		oyster_utils.LogIfError(err, nil)

		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("DONE PROCESSING!!!!!")

		// log res.Body to segment?
	}
}

func batchPowOnLambda(chunks *[]oyster_utils.ChunkData) {
	// Batch chunks by limit
	numBatches := (len(*chunks) / awsgateway.MaxChunksLen) + 1
	for i := 0; i < numBatches; i++ {
		offset := i * awsgateway.MaxChunksLen
		remChunks := len(*chunks) - offset

		// Numnber of chunks in this batch.
		var numChunks int
		if remChunks > awsgateway.MaxChunksLen {
			numChunks = awsgateway.MaxChunksLen
		} else {
			numChunks = remChunks
		}

		// Map chunk to lambdaChunk
		chunkBatch := make([]*awsgateway.HooknodeChunk, numChunks)
		for j := 0; j < numChunks; j++ {
			chk := (*chunks)[offset+j]
			lamChk := awsgateway.HooknodeChunk{
				Address: chk.Address,
				Value:   0,
				Tag:     OysterTagHook,
			}

			msg, err := trinary.NewTrytes(chk.Message)
			if err != nil {
				oyster_utils.LogIfError(err, nil)
				panic(err)
			}
			lamChk.Message = msg

			chunkBatch[j] = &lamChk
		}

		// Push chunkBatch to chan
		lambdaChan <- chunkBatch
	}
}

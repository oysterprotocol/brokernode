package services

import (
	"fmt"
	"github.com/iotaledger/giota"
	"github.com/oysterprotocol/brokernode/models"
	"os"
	"strings"
	"sync"
)

type PowJob struct {
	Transactions      []giota.Transaction
	TrunkTransaction  giota.Trytes
	BranchTransaction giota.Trytes
	BroadcastNodes    []string
}

type ProcessChunks func(chunks []models.DataMap, attachIfAlreadyAttached bool)
type VerifyChunkMessagesMatchRecord func(chunks []models.DataMap) (filteredChunks FilteredChunk, err error)
type VerifyChunksMatchRecord func(chunks []models.DataMap, checkChunkAndBranch bool) (filteredChunks FilteredChunk, err error)
type ChunksMatch func(chunkOnTangle giota.Transaction, chunkOnRecord models.DataMap, checkBranchAndTrunk bool) bool

type IotaService struct {
	ProcessChunks                  ProcessChunks
	VerifyChunkMessagesMatchRecord VerifyChunkMessagesMatchRecord
	VerifyChunksMatchRecord        VerifyChunksMatchRecord
	ChunksMatch                    ChunksMatch
}

type FilteredChunk struct {
	MatchesTangle      []models.DataMap
	DoesNotMatchTangle []models.DataMap
	NotAttached        []models.DataMap
}

var seed giota.Trytes

var provider = os.Getenv("PROVIDER")
var minDepth = int64(giota.DefaultNumberOfWalks)
var minWeightMag = int64(1)

var api = giota.NewAPI(provider, nil)
var bestPow giota.PowFunc
var powName string

// Things below are copied from the giota lib since they are not public.
// https://github.com/iotaledger/giota/blob/master/transfer.go#L322

// (3^27-1)/2
const maxTimestampTrytes = "MMMMMMMMM"

// This mutex was added by us.
var mutex = &sync.Mutex{}
var IotaWrapper IotaService

func init() {
	seed = "OYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRL"

	powName, bestPow = giota.GetBestPoW()

	IotaWrapper = IotaService{
		ProcessChunks:                  processChunks,
		VerifyChunkMessagesMatchRecord: verifyChunkMessagesMatchRecord,
		VerifyChunksMatchRecord:        verifyChunksMatchRecord,
		ChunksMatch:                    chunksMatch,
	}
}

func processChunks(chunks []models.DataMap, attachIfAlreadyAttached bool) {
	fmt.Println(chunks)
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
		fmt.Println(err)
		return filteredChunks, err
	}

	filteredChunks = FilteredChunk{}

	if response != nil && len(response.Hashes) > 0 {
		trytesArray, err := api.GetTrytes(response.Hashes)
		if err != nil {
			fmt.Println(err)
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

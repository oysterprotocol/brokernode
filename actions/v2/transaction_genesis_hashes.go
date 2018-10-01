package actions

import (
	"fmt"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/iotaledger/giota"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/utils"
	"math"
	"os"
	"strconv"
	"strings"
)

type TransactionGenesisHashResource struct {
	buffalo.Resource
}

// Request Response structs
type GenesisHashPow struct {
	Address  string `json:"address"`
	Message  string `json:"message"`
	BranchTx string `json:"branchTx"`
	TrunkTx  string `json:"trunkTx"`
}

type transactionGenesisHashCreateReq struct {
	CurrentList []string `json:"currentList"`
}

type transactionGenesisHashCreateRes struct {
	ID  uuid.UUID      `json:"id"`
	Pow GenesisHashPow `json:"pow"`
}

type transactionGenesisHashUpdateReq struct {
	Trytes string `json:"trytes"`
}

type transactionGenesisHashUpdateRes struct {
	Purchase       string `json:"purchase"`
	NumberOfChunks int    `json:"numberOfChunks"`
}

// Creates a transaction.
func (usr *TransactionGenesisHashResource) Create(c buffalo.Context) error {

	if os.Getenv("TANGLE_MAINTENANCE") == "true" {
		return c.Render(403, r.JSON(map[string]string{"error": "This broker is undergoing tangle maintenance"}))
	}

	if os.Getenv("DEPLOY_IN_PROGRESS") == "true" {
		return c.Render(403, r.JSON(map[string]string{"error": "Deployment in progress.  Try again later"}))
	}

	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramTransactionGenesisHashResourceCreate, start)

	req := transactionGenesisHashCreateReq{}
	if err := oyster_utils.ParseReqBody(c.Request(), &req); err != nil {
		err = fmt.Errorf("Invalid request, unable to parse request body  %v", err)
		c.Error(400, err)
		return err
	}

	storedGenesisHash, genesisHashNotFound := models.GetGenesisHashForWebnode(req.CurrentList)

	if genesisHashNotFound != nil {
		return c.Render(403, r.JSON(map[string]string{"error": "No genesis hash available"}))
	}

	dataMap, dataMapNotFoundErr := models.GetChunkForWebnodePoW()

	if dataMapNotFoundErr != nil {
		return c.Render(403, r.JSON(map[string]string{"error": "Cannot give proof of work because: " +
			dataMapNotFoundErr.Error()}))
	}

	tips, err := IotaWrapper.GetTransactionsToApprove()
	if err != nil {
		oyster_utils.LogIfError(err, nil)
		c.Error(400, err)
	}

	dataMapKey := oyster_utils.GetBadgerKey([]string{dataMap.GenesisHash, strconv.FormatInt(dataMap.Idx, 10)})

	t := models.Transaction{}
	models.DB.Transaction(func(tx *pop.Connection) error {

		storedGenesisHash.WebnodeCount++
		if storedGenesisHash.WebnodeCount >= models.WebnodeCountLimit {
			storedGenesisHash.Status = models.StoredGenesisHashAssigned
		}
		vErr, err := tx.ValidateAndSave(&storedGenesisHash)
		oyster_utils.LogIfError(err, nil)
		oyster_utils.LogIfValidationError("validation errors in transaction_genesis_hashes.", vErr, nil)

		t = models.Transaction{
			Type:        models.TransactionTypeGenesisHash,
			Status:      models.TransactionStatusPending,
			DataMapID:   dataMapKey,
			GenesisHash: dataMap.GenesisHash,
			Idx:         dataMap.Idx,
			Purchase:    storedGenesisHash.GenesisHash,
		}
		tx.ValidateAndSave(&t)
		return nil
	})

	res := transactionGenesisHashCreateRes{
		ID: t.ID,
		Pow: GenesisHashPow{
			Address:  dataMap.Address,
			Message:  dataMap.Message,
			BranchTx: string(tips.BranchTransaction),
			TrunkTx:  string(tips.TrunkTransaction),
		},
	}

	return c.Render(200, r.JSON(res))
}

func (usr *TransactionGenesisHashResource) Update(c buffalo.Context) error {
	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramTransactionGenesisHashResourceUpdate, start)

	req := transactionGenesisHashUpdateReq{}
	if err := oyster_utils.ParseReqBody(c.Request(), &req); err != nil {
		err = fmt.Errorf("Invalid request, unable to parse request body  %v", err)
		c.Error(400, err)
		return err
	}

	// Get transaction
	t := &models.Transaction{}
	transactionError := models.DB.Find(t, c.Param("id"))
	if transactionError != nil {
		return c.Render(400, r.JSON(map[string]string{"error": "No transaction found"}))
	}

	trytes, err := giota.ToTrytes(req.Trytes)
	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return c.Render(400, r.JSON(map[string]string{"error": err.Error()}))
	}
	iotaTransaction, iotaError := giota.NewTransaction(trytes)

	if iotaError != nil {
		return c.Render(400, r.JSON(map[string]string{"error": "Could not generate transaction object from trytes"}))
	}

	chunkDataInProgress := models.GetSingleChunkData(oyster_utils.InProgressDir, t.GenesisHash, t.Idx)
	chunkDataComplete := models.GetSingleChunkData(oyster_utils.InProgressDir, t.GenesisHash, t.Idx)

	chunkToUse := chunkDataInProgress
	if !oyster_utils.AllChunkDataHasArrived(chunkDataInProgress) &&
		oyster_utils.AllChunkDataHasArrived(chunkDataComplete) {
		chunkToUse = chunkDataComplete

	} else if !oyster_utils.AllChunkDataHasArrived(chunkDataInProgress) && !oyster_utils.AllChunkDataHasArrived(chunkDataComplete) {
		return c.Render(400, r.JSON(map[string]string{"error": "Could not find data for specified chunk"}))
	}

	address, addError := giota.ToAddress(chunkToUse.Address)

	validAddress := addError == nil && address == iotaTransaction.Address
	if !validAddress {
		return c.Render(400, r.JSON(map[string]string{"error": "Address is invalid"}))
	}

	_, messageErr := giota.ToTrytes(chunkToUse.Message)
	validMessage := messageErr == nil && strings.Contains(fmt.Sprint(iotaTransaction.SignatureMessageFragment),
		chunkToUse.Message)
	if !validMessage {
		return c.Render(400, r.JSON(map[string]string{"error": "Message is invalid"}))
	}

	host_ip := os.Getenv("HOST_IP")
	provider := "http://" + host_ip + ":14265"
	iotaAPI := giota.NewAPI(provider, nil)

	iotaTransactions := []giota.Transaction{*iotaTransaction}
	broadcastErr := iotaAPI.BroadcastTransactions(iotaTransactions)

	if broadcastErr != nil {
		return c.Render(400, r.JSON(map[string]string{"error": "Broadcast to Tangle failed"}))
	}

	storedGenesisHash := models.StoredGenesisHash{}
	genesisHashNotFound := models.DB.Limit(1).Where("genesis_hash = ?", t.Purchase).First(&storedGenesisHash)

	if genesisHashNotFound != nil {
		return c.Render(403, r.JSON(map[string]string{"error": "Stored genesis hash was not found"}))
	}

	models.DB.Transaction(func(tx *pop.Connection) error {
		t.Status = models.TransactionStatusComplete
		tx.ValidateAndSave(t)

		storedGenesisHash.Status = models.StoredGenesisHashUnassigned
		storedGenesisHash.WebnodeCount = storedGenesisHash.WebnodeCount + 1
		tx.ValidateAndSave(&storedGenesisHash)

		return nil
	})

	res := transactionGenesisHashUpdateRes{
		Purchase:       t.Purchase,
		NumberOfChunks: int(math.Ceil(float64(storedGenesisHash.FileSizeBytes) / models.FileBytesChunkSize)),
	}

	return c.Render(202, r.JSON(res))
}

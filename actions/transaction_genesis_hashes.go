package actions

import (
	"fmt"
	"github.com/pkg/errors"
	"math"
	"os"
	"strings"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/iotaledger/giota"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
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
	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramTransactionGenesisHashResourceCreate, start)

	req := transactionGenesisHashCreateReq{}
	oyster_utils.ParseReqBody(c.Request(), &req)

	storedGenesisHash, genesisHashNotFound := models.GetGenesisHashForWebnode(req.CurrentList)

	if genesisHashNotFound != nil {
		return c.Render(403, r.JSON(map[string]string{"error": "No genesis hash available"}))
	}

	dataMap := models.DataMap{}
	// TODO:  Would be better if this got a chunk from the session with the oldest "last chunk attached" time
	dataMapNotFound := models.DB.Limit(1).Where("status = ? ORDER BY updated_at asc",
		models.Unassigned).First(&dataMap)

	if dataMapNotFound != nil {
		return c.Render(403, r.JSON(map[string]string{"error": "No proof of work available"}))
	}

	tips, err := IotaWrapper.GetTransactionsToApprove()
	if err != nil {
		oyster_utils.LogIfError(err, nil)
		c.Error(400, err)
	}

	t := models.Transaction{}
	models.DB.Transaction(func(tx *pop.Connection) error {
		dataMap.Status = models.Unverified
		dataMap.BranchTx = string(tips.BranchTransaction)
		dataMap.TrunkTx = string(tips.TrunkTransaction)
		tx.ValidateAndSave(&dataMap)

		storedGenesisHash.WebnodeCount++
		if storedGenesisHash.WebnodeCount >= models.WebnodeCountLimit {
			storedGenesisHash.Status = models.StoredGenesisHashAssigned
		}
		vErr, err := tx.ValidateAndSave(&storedGenesisHash)

		if err != nil {
			oyster_utils.LogIfError(err, nil)
		}
		if len(vErr.Error()) > 0 {
			err = errors.New("validation errors in transaction_genesis_hashes: " + fmt.Sprint(vErr.Errors))
			oyster_utils.LogIfError(err, nil)
		}

		t = models.Transaction{
			Type:      models.TransactionTypeGenesisHash,
			Status:    models.TransactionStatusPending,
			DataMapID: dataMap.ID,
			Purchase:  storedGenesisHash.GenesisHash,
		}
		tx.ValidateAndSave(&t)
		return nil
	})

	res := transactionGenesisHashCreateRes{
		ID: t.ID,
		Pow: GenesisHashPow{
			Address:  dataMap.Address,
			Message:  services.GetMessageFromDataMap(dataMap),
			BranchTx: dataMap.BranchTx,
			TrunkTx:  dataMap.TrunkTx,
		},
	}

	return c.Render(200, r.JSON(res))
}

func (usr *TransactionGenesisHashResource) Update(c buffalo.Context) error {
	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramTransactionGenesisHashResourceUpdate, start)

	req := transactionGenesisHashUpdateReq{}
	oyster_utils.ParseReqBody(c.Request(), &req)

	// Get transaction
	t := &models.Transaction{}
	transactionError := models.DB.Eager("DataMap").Find(t, c.Param("id"))
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

	address, addError := giota.ToAddress(t.DataMap.Address)

	validAddress := addError == nil && address == iotaTransaction.Address
	if !validAddress {
		return c.Render(400, r.JSON(map[string]string{"error": "Address is invalid"}))
	}

	validMessage := strings.Contains(fmt.Sprint(iotaTransaction.SignatureMessageFragment), services.GetMessageFromDataMap(t.DataMap))
	if !validMessage {
		return c.Render(400, r.JSON(map[string]string{"error": "Message is invalid"}))
	}

	branchTxTrytes, err := giota.ToTrytes(t.DataMap.BranchTx)
	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return c.Render(400, r.JSON(map[string]string{"error": err.Error()}))
	}
	validBranch := branchTxTrytes == iotaTransaction.BranchTransaction
	if !validBranch {
		return c.Render(400, r.JSON(map[string]string{"error": "Branch is invalid"}))
	}

	trunkTxTrytes, err := giota.ToTrytes(t.DataMap.TrunkTx)
	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return c.Render(400, r.JSON(map[string]string{"error": err.Error()}))
	}
	validTrunk := trunkTxTrytes == iotaTransaction.TrunkTransaction
	if !validTrunk {
		return c.Render(400, r.JSON(map[string]string{"error": "Trunk is invalid"}))
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

		dataMap := t.DataMap
		dataMap.Status = models.Complete
		tx.ValidateAndSave(&dataMap)

		return nil
	})

	res := transactionGenesisHashUpdateRes{
		Purchase:       t.Purchase,
		NumberOfChunks: int(math.Ceil(float64(storedGenesisHash.FileSizeBytes) / models.FileBytesChunkSize)),
	}

	return c.Render(202, r.JSON(res))
}

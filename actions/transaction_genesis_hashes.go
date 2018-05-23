package actions

import (
	"fmt"
	// "os"
	"math"
	"strings"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/iotaledger/giota"
	"github.com/oysterprotocol/brokernode/models"
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
	req := transactionGenesisHashCreateReq{}
	oyster_utils.ParseReqBody(c.Request(), &req)

	existingGenesisHashes := oyster_utils.StringsJoin(req.CurrentList, ", ")
	storedGenesisHash := models.StoredGenesisHash{}
	genesisHashNotFound := models.DB.Limit(1).Where("genesis_hash NOT IN (?) AND webnode_count < ? AND status = ?", existingGenesisHashes, models.WebnodeCountLimit, models.StoredGenesisHashUnassigned).First(&storedGenesisHash)

	if genesisHashNotFound != nil {
		return c.Render(403, r.JSON(map[string]string{"error": "No genesis hash available"}))
	}

	dataMap := models.DataMap{}
	dataMapNotFound := models.DB.Limit(1).Where("status = ? AND genesis_hash = ?", models.Unassigned, storedGenesisHash.GenesisHash).First(&dataMap)

	if dataMapNotFound != nil {
		return c.Render(403, r.JSON(map[string]string{"error": "No proof of work available"}))
	}

	t := models.Transaction{}
	models.DB.Transaction(func(tx *pop.Connection) error {
		dataMap.Status = models.Unverified
		tx.ValidateAndSave(&dataMap)

		storedGenesisHash.Status = models.StoredGenesisHashAssigned
		tx.ValidateAndSave(&storedGenesisHash)

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
			Message:  dataMap.Message,
			BranchTx: dataMap.BranchTx,
			TrunkTx:  dataMap.TrunkTx,
		},
	}

	return c.Render(200, r.JSON(res))
}

func (usr *TransactionGenesisHashResource) Update(c buffalo.Context) error {
	req := transactionGenesisHashUpdateReq{}
	oyster_utils.ParseReqBody(c.Request(), &req)

	// Get transaction
	t := &models.Transaction{}
	transactionError := models.DB.Eager("DataMap").Find(t, c.Param("id"))

	trytes := giota.Trytes(req.Trytes)
	iotaTransaction, iotaError := giota.NewTransaction(trytes)

	if transactionError != nil || iotaError != nil {
		return c.Render(400, r.JSON(map[string]string{"error": "No transaction found"}))
	}

	address, addError := giota.ToAddress(t.DataMap.Address)
	validAddress := addError == nil && address == iotaTransaction.Address
	validMessage := strings.Contains(fmt.Sprint(iotaTransaction.SignatureMessageFragment), t.DataMap.Message)
	validBranch := giota.Trytes(t.DataMap.BranchTx) == iotaTransaction.BranchTransaction
	validTrunk := giota.Trytes(t.DataMap.TrunkTx) == iotaTransaction.TrunkTransaction

	if !(validAddress && validMessage && validBranch && validTrunk) {
		return c.Render(400, r.JSON(map[string]string{"error": "Transaction is invalid"}))
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
	genesisHashNotFound := models.DB.Limit(1).Where("genesis_hash = ?", t.DataMap.GenesisHash).First(&storedGenesisHash)

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

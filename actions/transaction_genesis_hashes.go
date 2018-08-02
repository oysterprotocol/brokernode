package actions

import (
	"errors"
	"fmt"
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

	if os.Getenv("TANGLE_MAINTENANCE") == "true" {
		return c.Render(403, r.JSON(map[string]string{"error": "This broker is undergoing tangle maintenance"}))
	}

	if os.Getenv("DEPLOY_IN_PROGRESS") == "true" {
		return c.Render(403, r.JSON(map[string]string{"error": "Deployment in progress.  Try again later"}))
	}

	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramTransactionGenesisHashResourceCreate, start)

	req := transactionGenesisHashCreateReq{}
	oyster_utils.ParseReqBody(c.Request(), &req)

	genesisHash := []models.StoredGenesisHash{}

	models.DB.All(&genesisHash)

	if len(genesisHash) == 0 {
		return c.Render(403, r.JSON(map[string]string{"error": "No genesis hash available"}))
	}

	dataMap := models.DataMap{}
	// TODO:  Would be better if this got a chunk from the session with the oldest "last chunk attached" time
	dataMapNotFound := models.DB.Limit(1).Where("status = ? ORDER BY updated_at ASC",
		models.Unassigned).First(&dataMap)

	if dataMapNotFound != nil {
		dataMap = models.DataMap{
			Address:        "OYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRL",
			GenesisHash:    "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			ChunkIdx:       0,
			Hash:           "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			ObfuscatedHash: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			Status:         models.Pending,
			MsgStatus:      1,
			MsgID: oyster_utils.GenerateMsgID("", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				0),
		}
	}

	tips, err := IotaWrapper.GetTransactionsToApprove()
	if err != nil {
		oyster_utils.LogIfError(err, nil)
		c.Error(400, err)
	}

	idToUse := dataMap.ID

	t := models.Transaction{}
	err = models.DB.Transaction(func(tx *pop.Connection) error {
		if dataMap.Address != "OYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRL" {
			dataMap.Status = models.Unverified
		}
		dataMap.BranchTx = string(tips.BranchTransaction)
		dataMap.TrunkTx = string(tips.TrunkTransaction)
		dm := []models.DataMap{}
		_ = models.DB.Where("address = ?", dataMap.Address).All(&dm)
		if len(dm) == 0 {
			vErr, err := tx.ValidateAndCreate(&dataMap)
			if vErr.HasAny() || err != nil {
				fmt.Println(vErr.Error())
				fmt.Println(err.Error())
				return errors.New("some error occurred while creating the data map")
			}
		} else {
			idToUse = dm[0].ID
			vErr, err := tx.ValidateAndUpdate(&dataMap)
			if vErr.HasAny() || err != nil {
				fmt.Println(vErr.Error())
				fmt.Println(err.Error())
				return errors.New("some error occurred while updating the data map")
			}
		}

		genesisHash[0].WebnodeCount++
		if genesisHash[0].WebnodeCount >= models.WebnodeCountLimit {
			//storedGenesisHash.Status = models.StoredGenesisHashAssigned
		}
		vErr, err := tx.ValidateAndSave(&genesisHash[0])
		if vErr.HasAny() || err != nil {
			fmt.Println(vErr.Error())
			fmt.Println(err.Error())
			return errors.New("some error occurred while updating the stored genesis hash")
		}
		oyster_utils.LogIfError(err, nil)
		oyster_utils.LogIfValidationError("validation errors in transaction_genesis_hashes.", vErr, nil)

		t = models.Transaction{
			Type:      models.TransactionTypeGenesisHash,
			Status:    models.TransactionStatusPending,
			DataMapID: idToUse,
			Purchase:  genesisHash[0].GenesisHash,
		}
		tx.ValidateAndSave(&t)
		return nil
	})

	if err != nil {
		return c.Render(403, r.JSON(map[string]string{"error": err.Error()}))
	}

	res := transactionGenesisHashCreateRes{
		ID: t.ID,
		Pow: GenesisHashPow{
			Address:  dataMap.Address,
			Message:  "THISISAWEBNODEDEMO",
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

	fmt.Println(t)
	if addError != nil {
		fmt.Println("addErr: " + addError.Error())
	}
	if address != iotaTransaction.Address {
		fmt.Println("did not match iota address")
	}

	validAddress := addError == nil && address == iotaTransaction.Address
	if !validAddress {
		return c.Render(400, r.JSON(map[string]string{"error": "Address is invalid"}))
	}

	validMessage := strings.Contains(fmt.Sprint(iotaTransaction.SignatureMessageFragment), services.GetMessageFromDataMap(t.DataMap))
	if !validMessage && address != "OYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRL" {
		return c.Render(400, r.JSON(map[string]string{"error": "Message is invalid"}))
	}

	_, err = giota.ToTrytes(t.DataMap.BranchTx)
	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return c.Render(400, r.JSON(map[string]string{"error": err.Error()}))
	}

	_, err = giota.ToTrytes(t.DataMap.TrunkTx)
	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return c.Render(400, r.JSON(map[string]string{"error": err.Error()}))
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
		tx.ValidateAndSave(&storedGenesisHash)

		dataMap := t.DataMap
		if dataMap.Address != "OYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRL" {
			dataMap.Status = models.Complete
		}
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

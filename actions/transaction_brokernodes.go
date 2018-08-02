package actions

import (
	"fmt"
	"github.com/pkg/errors"
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

type TransactionBrokernodeResource struct {
	buffalo.Resource
}

// Request Response structs
type BrokernodeAddressPow struct {
	Address  string `json:"address"`
	Message  string `json:"message"`
	BranchTx string `json:"branchTx"`
	TrunkTx  string `json:"trunkTx"`
}

type transactionBrokernodeCreateReq struct {
	CurrentList []string `json:"currentList"`
}

type transactionBrokernodeCreateRes struct {
	ID  uuid.UUID            `json:"id"`
	Pow BrokernodeAddressPow `json:"pow"`
}

type transactionBrokernodeUpdateReq struct {
	Trytes string `json:"trytes"`
}

type transactionBrokernodeUpdateRes struct {
	Purchase string `json:"purchase"`
}

// Creates a transaction.

func (usr *TransactionBrokernodeResource) Create(c buffalo.Context) error {

	if os.Getenv("TANGLE_MAINTENANCE") == "true" {
		return c.Render(403, r.JSON(map[string]string{"error": "This broker is undergoing tangle maintenance"}))
	}

	if os.Getenv("DEPLOY_IN_PROGRESS") == "true" {
		return c.Render(403, r.JSON(map[string]string{"error": "Deployment in progress.  Try again later"}))
	}

	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramTransactionBrokernodeResourceCreate, start)

	req := transactionBrokernodeCreateReq{}
	oyster_utils.ParseReqBody(c.Request(), &req)

	dataMap := models.DataMap{}
	brokernode := models.Brokernode{}
	t := models.Transaction{}

	// TODO:  Would be better if this got a chunk from the session with the oldest "last chunk attached" time
	dataMapNotFoundErr := models.DB.Limit(1).Where("status = ? ORDER BY updated_at ASC",
		models.Unassigned).First(&dataMap)

	brokernodeNotFoundErr := models.DB.First(&brokernode)

	// DB results error if First() does not return any error.
	if dataMapNotFoundErr != nil {
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

	// DB results error if First() does not return any error.
	if brokernodeNotFoundErr != nil {
		return c.Render(403, r.JSON(map[string]string{"error": "No brokernode addresses to sell"}))
	}

	tips, err := IotaWrapper.GetTransactionsToApprove()
	if err != nil {
		oyster_utils.LogIfError(err, nil)
		c.Error(400, err)
	}

	idToUse := dataMap.ID

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

		t = models.Transaction{
			Type:      models.TransactionTypeBrokernode,
			Status:    models.TransactionStatusPending,
			DataMapID: idToUse,
			Purchase:  brokernode.Address,
		}
		tx.ValidateAndSave(&t)
		return nil
	})

	if err != nil {
		return c.Render(403, r.JSON(map[string]string{"error": err.Error()}))
	}

	res := transactionBrokernodeCreateRes{
		ID: t.ID,
		Pow: BrokernodeAddressPow{
			Address:  dataMap.Address,
			Message:  "THISISAWEBNODEDEMO",
			BranchTx: dataMap.BranchTx,
			TrunkTx:  dataMap.TrunkTx,
		},
	}

	return c.Render(200, r.JSON(res))
}

func (usr *TransactionBrokernodeResource) Update(c buffalo.Context) error {
	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramTransactionBrokernodeResourceUpdate, start)

	req := transactionBrokernodeUpdateReq{}
	oyster_utils.ParseReqBody(c.Request(), &req)

	fmt.Println(c.Param("id"))

	// Get transaction
	t := &models.Transaction{}
	transactionError := models.DB.Eager("DataMap").Find(t, c.Param("id"))

	trytes, err := giota.ToTrytes(req.Trytes)
	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return c.Render(400, r.JSON(map[string]string{"error": err.Error()}))
	}
	iotaTransaction, iotaError := giota.NewTransaction(trytes)

	if transactionError != nil || iotaError != nil {
		return c.Render(400, r.JSON(map[string]string{"error": "No transaction found"}))
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

	models.DB.Transaction(func(tx *pop.Connection) error {
		t.Status = models.TransactionStatusComplete
		tx.ValidateAndSave(t)

		dataMap := t.DataMap
		if dataMap.Address != "OYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRL" {
			dataMap.Status = models.Complete
		}

		tx.ValidateAndSave(&dataMap)

		return nil
	})

	res := transactionBrokernodeUpdateRes{Purchase: t.Purchase}

	return c.Render(202, r.JSON(res))
}

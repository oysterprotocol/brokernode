package actions

import (
	"fmt"
	"os"
	"strings"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/iotaledger/giota"
	"github.com/oysterprotocol/brokernode/models"
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
	req := transactionBrokernodeCreateReq{}
	oyster_utils.ParseReqBody(c.Request(), &req)

	dataMap := models.DataMap{}
	brokernode := models.Brokernode{}
	t := models.Transaction{}

	dataMapNotFound := models.DB.Limit(1).Where("status = ?", models.Unassigned).First(&dataMap)

	existingAddresses := oyster_utils.StringsJoin(req.CurrentList, oyster_utils.StringsJoinDelim)
	brokernodeNotFound := models.DB.Limit(1).Where("address NOT IN (?)", existingAddresses).First(&brokernode)

	if dataMapNotFound != nil || brokernodeNotFound != nil {
		return c.Render(403, r.JSON(map[string]string{"error": "No proof of work available"}))
	}

	models.DB.Transaction(func(tx *pop.Connection) error {
		dataMap.Status = models.Unverified
		tx.ValidateAndSave(&dataMap)

		t = models.Transaction{
			Type:      models.TransactionTypeBrokernode,
			Status:    models.TransactionStatusPending,
			DataMapID: dataMap.ID,
			Purchase:  brokernode.Address,
		}
		tx.ValidateAndSave(&t)
		return nil
	})

	res := transactionBrokernodeCreateRes{
		ID: t.ID,
		Pow: BrokernodeAddressPow{
			Address:  dataMap.Address,
			Message:  dataMap.Message,
			BranchTx: dataMap.BranchTx,
			TrunkTx:  dataMap.TrunkTx,
		},
	}

	return c.Render(200, r.JSON(res))
}

func (usr *TransactionBrokernodeResource) Update(c buffalo.Context) error {
	req := transactionBrokernodeUpdateReq{}
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

	models.DB.Transaction(func(tx *pop.Connection) error {
		t.Status = models.TransactionStatusComplete
		tx.ValidateAndSave(t)

		dataMap := t.DataMap
		dataMap.Status = models.Complete
		tx.ValidateAndSave(&dataMap)

		return nil
	})

	res := transactionBrokernodeUpdateRes{Purchase: t.Purchase}

	return c.Render(202, r.JSON(res))
}

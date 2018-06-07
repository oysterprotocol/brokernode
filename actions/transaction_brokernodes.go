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

	dataMapNotFoundErr := models.DB.Where("status = ?", models.Unassigned).First(&dataMap)

	existingAddresses := oyster_utils.StringsJoin(req.CurrentList, oyster_utils.StringsJoinDelim)
	brokernodeNotFoundErr := models.DB.Where("address NOT IN (?)", existingAddresses).First(&brokernode)

	// DB results error if First() does not return any error.
	if dataMapNotFoundErr != nil || brokernodeNotFoundErr != nil {
		return c.Render(403, r.JSON(map[string]string{"error": "No proof of work available"}))
	}

	tips, err := IotaWrapper.GetTransactionsToApprove()
	if err != nil {
		oyster_utils.LogIfError(err, nil)
		c.Error(400, err)
	}

	err = models.DB.Transaction(func(tx *pop.Connection) error {
		dataMap.Status = models.Unverified
		dataMap.BranchTx = string(tips.BranchTransaction)
		dataMap.TrunkTx = string(tips.TrunkTransaction)
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
	if err != nil {
		oyster_utils.LogIfError(err, nil)
	}

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
	validAddress := addError == nil && address == iotaTransaction.Address
	if !validAddress {
		return c.Render(400, r.JSON(map[string]string{"error": "Address is invalid"}))
	}

	validMessage := strings.Contains(fmt.Sprint(iotaTransaction.SignatureMessageFragment), t.DataMap.Message)
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

package actions

import (
	"os"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/iotaledger/giota"
	"github.com/oysterprotocol/brokernode/models"
)

type TransactionBrokernodeResource struct {
	buffalo.Resource
}

// Request Response structs
type Pow struct {
	Address  string `json:"address"`
	Message  string `json:"message"`
	BranchTx string `json:"branchTx"`
	TrunkTx  string `json:"trunkTx"`
}

type transactionCreateReq struct {
	CurrentList []string `json:"currentList"`
}

type transactionCreateRes struct {
	ID  uuid.UUID `json:"id"`
	Pow Pow       `json:"pow"`
}

type transactionUpdateReq struct {
	Trytes string `json:"trytes"`
}

type transactionUpdateRes struct {
	Purchase string `json:"purchase"`
}

// Creates a transaction.
func (usr *TransactionBrokernodeResource) Create(c buffalo.Context) error {
	req := transactionCreateReq{}
	parseReqBody(c.Request(), &req)

	dataMap := models.DataMap{}
	brokernode := models.Brokernode{}
	t := models.Transaction{}

	dataMapNotFound := models.DB.Limit(1).Where("status = ?", models.Unassigned).First(&dataMap)

	existingAddresses := join(req.CurrentList, ", ")
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

	res := transactionCreateRes{
		ID: t.ID,
		Pow: Pow{
			Address:  dataMap.Address,
			Message:  dataMap.Message,
			BranchTx: dataMap.BranchTx,
			TrunkTx:  dataMap.TrunkTx,
		},
	}

	return c.Render(200, r.JSON(res))
}

func (usr *TransactionBrokernodeResource) Update(c buffalo.Context) error {
	req := transactionUpdateReq{}
	parseReqBody(c.Request(), &req)

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
	validMessage := giota.Trytes(t.DataMap.Message) == iotaTransaction.SignatureMessageFragment
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

	t.Status = models.TransactionStatusComplete
	models.DB.ValidateAndUpdate(t)

	res := transactionUpdateRes{Purchase: t.Purchase}

	return c.Render(202, r.JSON(res))
}

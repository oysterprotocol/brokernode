package actions

import (
	"bytes"
	// "fmt"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
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

// Creates a transaction.
func (usr *TransactionBrokernodeResource) Create(c buffalo.Context) error {
	req := transactionCreateReq{}
	parseReqBody(c.Request(), &req)

	dataMap := models.DataMap{}
	brokernode := models.Brokernode{}
	t := models.Transaction{}

	models.DB.Limit(1).Where("status = ?", models.Unassigned).First(&dataMap)

	existingAddresses := join(req.CurrentList, ", ")
	models.DB.Limit(1).Where("address NOT IN (?)", existingAddresses).First(&brokernode)

	models.DB.Transaction(func(tx *pop.Connection) error {
		dataMap.Status = models.Unverified
		tx.ValidateAndSave(&dataMap)

		t = models.Transaction{
			Type:      "BROKERNODE",
			Status:    "PAYMENT_PENDING",
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

// TODO: put this in a helper
func join(A []string, delim string) string {
	var buffer bytes.Buffer
	for i := 0; i < len(A); i++ {
		buffer.WriteString(A[i])
		if i != len(A)-1 {
			buffer.WriteString(delim)
		}
	}

	return buffer.String()
}

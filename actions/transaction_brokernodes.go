package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/oysterprotocol/brokernode/models"
)

type TransactionBrokernodeResource struct {
	buffalo.Resource
}

// Request Response structs

type transactionCreateRes struct {
	ID models.Transaction `json:"id"`
}

// Creates a transaction.
func (usr *TransactionBrokernodeResource) Create(c buffalo.Context) error {
	req := transactionCreateReq{}
	parseReqBody(c.Request(), &req)

	t := models.Transaction{
		Type:      "BROKERNODE",
		Status:    "PAYMENT_PENDING",
		DataMapID: "xxxxx",
		Purchase:  "xxxxx",
	}

	res := transactionCreateRes{
		ID: t.id,
	}

	return c.Render(200, r.JSON(res))
}

package actions

import (
	"fmt"
	"github.com/gobuffalo/buffalo"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/pkg/errors"
)

type TransactionBrokernodeResource struct {
	buffalo.Resource
}

// Request Response structs

type transactionCreateReq struct {
	CurrentList []string `json:"currentList"`
}

type transactionCreateRes struct {
	ID models.Transaction `json:"id"`
}

// Creates a transaction.
func (usr *TransactionBrokernodeResource) Create(c buffalo.Context) error {
	req := transactionCreateReq{}
	parseReqBody(c.Request(), &req)

	fmt.Println("xxxxxxxxxxxxxxxxxxxx")
	tx := models.DB
	query := tx.Limit(1).Where("status = 'unassigned' LIMIT 1")
	dataMap := models.DataMap{}
	err := query.All(&dataMap)

	if err != nil {
		c.Render(400, r.JSON(map[string]string{"Error finding session": errors.WithStack(err).Error()}))
		return err
	}

	fmt.Println(dataMap)

	t := models.Transaction{
		Type:      "BROKERNODE",
		Status:    "PAYMENT_PENDING",
		DataMapID: dataMap.ID,
		Purchase:  "xxxxx",
	}

	res := transactionCreateRes{
		ID: t.id,
	}

	return c.Render(200, r.JSON(res))
}

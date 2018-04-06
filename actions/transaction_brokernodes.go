package actions

import (
	"bytes"
	"fmt"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/uuid"
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
	ID uuid.UUID `json:"id"`
}

// Creates a transaction.
func (usr *TransactionBrokernodeResource) Create(c buffalo.Context) error {
	req := transactionCreateReq{}
	parseReqBody(c.Request(), &req)

	fmt.Println("DATAMAPPPPPPPPPPPPPPPP")
	tx := models.DB
	dataMap := models.DataMap{}
	dataMapError := tx.Limit(1).Where("status = 'unassigned'").First(&dataMap)

	fmt.Println(dataMap)

	if dataMapError != nil {
		c.Render(400, r.JSON(map[string]string{"Error finding session": errors.WithStack(dataMapError).Error()}))
		return dataMapError
	}

	fmt.Println("BROKERNODEEEEEEEEEEEE")
	brokernode := models.Brokernode{}
	existingAddresses := join(req.CurrentList, ", ")
	brokernodeError := tx.Limit(1).Where("address NOT IN (?)", existingAddresses).First(&brokernode)

	fmt.Println(brokernode)

	if brokernodeError != nil {
		c.Render(400, r.JSON(map[string]string{"Error finding session": errors.WithStack(brokernodeError).Error()}))
		return brokernodeError
	}

	t := models.Transaction{
		Type:      "BROKERNODE",
		Status:    "PAYMENT_PENDING",
		DataMapID: dataMap.ID,
		Purchase:  "xxxxx",
	}

	res := transactionCreateRes{
		ID: t.ID,
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

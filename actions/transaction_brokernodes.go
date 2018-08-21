package actions

import (
	"fmt"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/iotaledger/giota"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/utils"
	"os"
	"strconv"
	"strings"
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
	if err := oyster_utils.ParseReqBody(c.Request(), &req); err != nil {
		err = fmt.Errorf("Invalid request, unable to parse request body  %v", err)
		c.Error(400, err)
		return err
	}

	brokernode := models.Brokernode{}

	dataMap, dataMapNotFoundErr := models.GetChunkForWebnodePoW()

	existingAddresses := oyster_utils.StringsJoin(req.CurrentList, oyster_utils.StringsJoinDelim)
	brokernodeNotFoundErr := models.DB.Where("address NOT IN (?)", existingAddresses).First(&brokernode)

	// DB results error if First() does not return any error.
	if dataMapNotFoundErr != nil {
		return c.Render(403, r.JSON(map[string]string{"error": "Cannot give proof of work because: " +
			dataMapNotFoundErr.Error()}))
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

	dataMapKey := oyster_utils.GetBadgerKey([]string{dataMap.GenesisHash, strconv.FormatInt(dataMap.Idx, 10)})

	transaction := models.Transaction{
		Type:        models.TransactionTypeBrokernode,
		Status:      models.TransactionStatusPending,
		DataMapID:   dataMapKey,
		GenesisHash: dataMap.GenesisHash,
		Idx:         dataMap.Idx,
		Purchase:    brokernode.Address,
	}

	err = models.DB.Transaction(func(tx *pop.Connection) error {

		vErr, err := tx.ValidateAndCreate(&transaction)
		if err != nil || vErr.HasAny() {
			return fmt.Errorf("Unable to Save Transaction: %v, %v", vErr, err)
		}
		return nil
	})
	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return c.Error(400, err)
	}

	res := transactionBrokernodeCreateRes{
		ID: transaction.ID,
		Pow: BrokernodeAddressPow{
			Address:  dataMap.Address,
			Message:  dataMap.Message,
			BranchTx: string(tips.BranchTransaction),
			TrunkTx:  string(tips.TrunkTransaction),
		},
	}

	return c.Render(200, r.JSON(res))
}

func (usr *TransactionBrokernodeResource) Update(c buffalo.Context) error {
	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramTransactionBrokernodeResourceUpdate, start)

	req := transactionBrokernodeUpdateReq{}
	if err := oyster_utils.ParseReqBody(c.Request(), &req); err != nil {
		err = fmt.Errorf("Invalid request, unable to parse request body  %v", err)
		c.Error(400, err)
		return err
	}

	// Get transaction
	t := &models.Transaction{}
	transactionError := models.DB.Find(t, c.Param("id"))

	trytes, err := giota.ToTrytes(req.Trytes)
	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return c.Render(400, r.JSON(map[string]string{"error": err.Error()}))
	}
	iotaTransaction, iotaError := giota.NewTransaction(trytes)

	if transactionError != nil || iotaError != nil {
		return c.Render(400, r.JSON(map[string]string{"error": "No transaction found"}))
	}

	chunkDataInProgress := oyster_utils.GetChunkData(oyster_utils.InProgressDir, t.GenesisHash, t.Idx)
	chunkDataComplete := oyster_utils.GetChunkData(oyster_utils.InProgressDir, t.GenesisHash, t.Idx)

	chunkToUse := chunkDataInProgress
	if !oyster_utils.AllChunkDataHasArrived(chunkDataInProgress) &&
		oyster_utils.AllChunkDataHasArrived(chunkDataComplete) {
		chunkToUse = chunkDataComplete
	} else if !oyster_utils.AllChunkDataHasArrived(chunkDataInProgress) && !oyster_utils.AllChunkDataHasArrived(chunkDataComplete) {
		return c.Render(400, r.JSON(map[string]string{"error": "Could not find data for specified chunk"}))
	}

	address, addError := giota.ToAddress(chunkToUse.Address)
	validAddress := addError == nil && address == iotaTransaction.Address
	if !validAddress {
		return c.Render(400, r.JSON(map[string]string{"error": "Address is invalid"}))
	}

	_, messageErr := giota.ToTrytes(chunkToUse.Message)
	validMessage := messageErr == nil && strings.Contains(fmt.Sprint(iotaTransaction.SignatureMessageFragment),
		chunkToUse.Message)
	if !validMessage {
		return c.Render(400, r.JSON(map[string]string{"error": "Message is invalid"}))
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

		return nil
	})

	res := transactionBrokernodeUpdateRes{Purchase: t.Purchase}

	return c.Render(202, r.JSON(res))
}

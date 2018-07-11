package models_test

import (
	"github.com/gobuffalo/pop/nulls"
	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/utils"
	"github.com/shopspring/decimal"
	"time"
)

var (
	totalCost               = decimal.NewFromFloat(float64(0.015625))
	expectedTotalCostString = "15625000000000000"
)

func (suite *ModelSuite) Test_InitialBrokerBrokerTransactionCreation() {

	privateKey := "abcdef1234567890"
	startingSessionType := 5 // this type does not exist
	startingPaymentStatus := 0
	startingEthAddr := "0000000000"
	startingGenHash := "AAAAAAAAAAAAAAAAA"

	brokerTx := models.BrokerBrokerTransaction{
		GenesisHash:   startingGenHash,
		Type:          startingSessionType,
		ETHAddrAlpha:  startingEthAddr,
		ETHAddrBeta:   startingEthAddr,
		ETHPrivateKey: privateKey,
		TotalCost:     totalCost,
		PaymentStatus: models.PaymentStatus(startingPaymentStatus),
	}

	vErr, err := suite.DB.ValidateAndCreate(&brokerTx)
	suite.Nil(err)
	suite.False(vErr.HasAny())

	brokerTxs := returnAllBrokerBrokerTxs(suite)
	suite.Equal(1, len(brokerTxs))

	// verify this is the same brokerTx that we created
	suite.Equal(startingGenHash, brokerTxs[0].GenesisHash)

	// verify the eth key has been encrypted
	suite.NotEqual(privateKey, brokerTxs[0].ETHPrivateKey)

	// verify the session type has changed
	suite.NotEqual(startingSessionType, brokerTxs[0].Type)

	// verify the payment status has changed
	suite.NotEqual(startingPaymentStatus, brokerTxs[0].PaymentStatus)
}

func (suite *ModelSuite) Test_NewBrokerBrokerTransaction() {

	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	defer oyster_utils.ResetBrokerMode()

	genHash := "abcdef"
	fileSizeBytes := 123
	numChunks := 2
	storageLengthInYears := 2
	privateKey := "abcdef1234567890"
	startingEthAddr := "0000000000"

	u := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          genHash,
		FileSizeBytes:        uint64(fileSizeBytes),
		NumChunks:            numChunks,
		StorageLengthInYears: storageLengthInYears,
		ETHPrivateKey:        privateKey,
		ETHAddrAlpha:         nulls.String{string(startingEthAddr), true},
		ETHAddrBeta:          nulls.String{string(startingEthAddr), true},
		TotalCost:            totalCost,
	}

	vErr, err := u.StartUploadSession()
	suite.Nil(err)
	suite.False(vErr.HasAny())

	uSession := models.UploadSession{}
	suite.DB.Where("genesis_hash = ?", genHash).First(&uSession)

	models.NewBrokerBrokerTransaction(&uSession)

	brokerTxs := returnAllBrokerBrokerTxs(suite)
	suite.Equal(1, len(brokerTxs))

	suite.Equal(uSession.GenesisHash, brokerTxs[0].GenesisHash)
	suite.Equal(uSession.ETHAddrAlpha.String, brokerTxs[0].ETHAddrAlpha)
	suite.Equal(uSession.ETHAddrBeta.String, brokerTxs[0].ETHAddrBeta)
	suite.Equal(uSession.TotalCost, brokerTxs[0].TotalCost)
	suite.NotEqual(uSession.ETHPrivateKey, brokerTxs[0].ETHPrivateKey)

	ethKeyOfSession := uSession.DecryptSessionEthKey()
	ethKeyOfBrokerTx := brokerTxs[0].DecryptEthKey()

	suite.Equal(ethKeyOfSession, ethKeyOfBrokerTx)
}

func (suite *ModelSuite) Test_GetTotalCostInWei() {

	generateBrokerBrokerTransactions(
		suite,
		models.SessionTypeAlpha,
		models.BrokerTxBetaPaymentPending,
		1)

	brokerTxs := returnAllBrokerBrokerTxs(suite)
	suite.Equal(1, len(brokerTxs))

	totalCostInWei := brokerTxs[0].GetTotalCostInWei()

	suite.Equal(expectedTotalCostString, totalCostInWei.String())
}

func (suite *ModelSuite) Test_GetTransactionsBySessionTypesAndPaymentStatuses_with_session_type() {

	generateBrokerBrokerTransactions(
		suite,
		models.SessionTypeAlpha,
		models.BrokerTxBetaPaymentPending,
		1)

	generateBrokerBrokerTransactions(
		suite,
		models.SessionTypeBeta,
		models.BrokerTxBetaPaymentPending,
		1)

	generateBrokerBrokerTransactions(
		suite,
		models.SessionTypeAlpha,
		models.BrokerTxBetaPaymentConfirmed,
		1)

	allBrokerTxs := returnAllBrokerBrokerTxs(suite)
	suite.Equal(3, len(allBrokerTxs))

	brokerTxs, err := models.GetTransactionsBySessionTypesAndPaymentStatuses(
		[]int{models.SessionTypeAlpha},
		[]models.PaymentStatus{models.BrokerTxBetaPaymentPending})
	suite.Nil(err)

	suite.Equal(1, len(brokerTxs))

	suite.Equal(models.SessionTypeAlpha, brokerTxs[0].Type)
	suite.Equal(models.BrokerTxBetaPaymentPending, brokerTxs[0].PaymentStatus)
}

func (suite *ModelSuite) Test_GetTransactionsBySessionTypesAndPaymentStatuses_no_session() {

	generateBrokerBrokerTransactions(
		suite,
		models.SessionTypeAlpha,
		models.BrokerTxBetaPaymentPending,
		1)

	generateBrokerBrokerTransactions(
		suite,
		models.SessionTypeBeta,
		models.BrokerTxBetaPaymentPending,
		1)

	generateBrokerBrokerTransactions(
		suite,
		models.SessionTypeAlpha,
		models.BrokerTxBetaPaymentConfirmed,
		1)

	allBrokerTxs := returnAllBrokerBrokerTxs(suite)
	suite.Equal(3, len(allBrokerTxs))

	brokerTxs, err := models.GetTransactionsBySessionTypesAndPaymentStatuses(
		[]int{},
		[]models.PaymentStatus{models.BrokerTxBetaPaymentPending})
	suite.Nil(err)

	suite.Equal(2, len(brokerTxs))

	suite.Equal(models.BrokerTxBetaPaymentPending, brokerTxs[0].PaymentStatus)
	suite.Equal(models.BrokerTxBetaPaymentPending, brokerTxs[1].PaymentStatus)
}

func (suite *ModelSuite) Test_GetTransactionsBySessionTypesAndPaymentStatuses_both_session_types() {

	generateBrokerBrokerTransactions(
		suite,
		models.SessionTypeAlpha,
		models.BrokerTxBetaPaymentPending,
		1)

	generateBrokerBrokerTransactions(
		suite,
		models.SessionTypeBeta,
		models.BrokerTxBetaPaymentPending,
		1)

	generateBrokerBrokerTransactions(
		suite,
		models.SessionTypeAlpha,
		models.BrokerTxBetaPaymentConfirmed,
		1)

	allBrokerTxs := returnAllBrokerBrokerTxs(suite)
	suite.Equal(3, len(allBrokerTxs))

	brokerTxs, err := models.GetTransactionsBySessionTypesAndPaymentStatuses(
		[]int{models.SessionTypeAlpha, models.SessionTypeBeta},
		[]models.PaymentStatus{models.BrokerTxBetaPaymentPending})
	suite.Nil(err)

	suite.Equal(2, len(brokerTxs))

	suite.Equal(models.BrokerTxBetaPaymentPending, brokerTxs[0].PaymentStatus)
	suite.Equal(models.BrokerTxBetaPaymentPending, brokerTxs[1].PaymentStatus)
}

func (suite *ModelSuite) Test_GetTransactionsBySessionTypesPaymentStatusesAndTime() {

	generateBrokerBrokerTransactions(
		suite,
		models.SessionTypeAlpha,
		models.BrokerTxBetaPaymentPending,
		1)

	generateBrokerBrokerTransactions(
		suite,
		models.SessionTypeBeta,
		models.BrokerTxBetaPaymentPending,
		1)

	generateBrokerBrokerTransactions(
		suite,
		models.SessionTypeAlpha,
		models.BrokerTxBetaPaymentConfirmed,
		1)

	allBrokerTxs := returnAllBrokerBrokerTxs(suite)
	suite.Equal(3, len(allBrokerTxs))

	brokerTxs1, err := models.GetTransactionsBySessionTypesPaymentStatusesAndTime(
		[]int{models.SessionTypeAlpha},
		[]models.PaymentStatus{models.BrokerTxBetaPaymentPending},
		time.Now().Add(-5*time.Minute))
	suite.Nil(err)

	suite.Equal(0, len(brokerTxs1))

	brokerTxs2, err := models.GetTransactionsBySessionTypesPaymentStatusesAndTime(
		[]int{models.SessionTypeAlpha},
		[]models.PaymentStatus{models.BrokerTxBetaPaymentPending},
		time.Now().Add(5*time.Minute))
	suite.Nil(err)

	suite.Equal(1, len(brokerTxs2))

	suite.Equal(models.SessionTypeAlpha, brokerTxs2[0].Type)
	suite.Equal(models.BrokerTxBetaPaymentPending, brokerTxs2[0].PaymentStatus)
}

func (suite *ModelSuite) Test_SetUploadSessionToPaid() {

	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	defer oyster_utils.ResetBrokerMode()

	genHash := "abcdef"
	fileSizeBytes := 123
	numChunks := 2
	storageLengthInYears := 2
	privateKey := "abcdef1234567890"
	startingEthAddr := "0000000000"

	u := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          genHash,
		FileSizeBytes:        uint64(fileSizeBytes),
		NumChunks:            numChunks,
		StorageLengthInYears: storageLengthInYears,
		ETHPrivateKey:        privateKey,
		ETHAddrAlpha:         nulls.String{string(startingEthAddr), true},
		ETHAddrBeta:          nulls.String{string(startingEthAddr), true},
		TotalCost:            totalCost,
	}

	vErr, err := u.StartUploadSession()
	suite.Nil(err)
	suite.False(vErr.HasAny())

	uSession := models.UploadSession{}
	suite.DB.Where("genesis_hash = ?", genHash).First(&uSession)

	models.NewBrokerBrokerTransaction(&uSession)

	brokerTxs := returnAllBrokerBrokerTxs(suite)
	suite.Equal(1, len(brokerTxs))

	suite.Equal(models.BrokerTxAlphaPaymentPending, brokerTxs[0].PaymentStatus)
	suite.Equal(models.PaymentStatusInvoiced, uSession.PaymentStatus)

	models.SetUploadSessionToPaid(brokerTxs[0])

	suite.DB.Where("genesis_hash = ?", genHash).First(&uSession)
	suite.Equal(models.PaymentStatusConfirmed, uSession.PaymentStatus)
}

func (suite *ModelSuite) Test_DeleteCompletedBrokerTransactions() {

	generateBrokerBrokerTransactions(
		suite,
		models.SessionTypeAlpha,
		models.BrokerTxBetaPaymentConfirmed,
		2)

	generateBrokerBrokerTransactions(
		suite,
		models.SessionTypeBeta,
		models.BrokerTxBetaPaymentConfirmed,
		1)

	allBrokerTxs := returnAllBrokerBrokerTxs(suite)
	suite.Equal(3, len(allBrokerTxs))

	models.DeleteCompletedBrokerTransactions()

	allBrokerTxs = returnAllBrokerBrokerTxs(suite)
	suite.Equal(0, len(allBrokerTxs))
}

func generateBrokerBrokerTransactions(suite *ModelSuite,
	sessionType int,
	paymentStatus models.PaymentStatus,
	numToGenerate int) {

	for i := 0; i < numToGenerate; i++ {
		alphaAddr, key, _ := jobs.EthWrapper.GenerateEthAddr()
		betaAddr, _, _ := jobs.EthWrapper.GenerateEthAddr()

		validChars := []rune("abcde123456789")
		genesisHash := oyster_utils.RandSeq(64, validChars)

		brokerTx := models.BrokerBrokerTransaction{
			GenesisHash:   genesisHash,
			Type:          sessionType,
			ETHAddrAlpha:  alphaAddr.Hex(),
			ETHAddrBeta:   betaAddr.Hex(),
			ETHPrivateKey: key,
			TotalCost:     totalCost,
			PaymentStatus: paymentStatus,
		}

		vErr, err := suite.DB.ValidateAndCreate(&brokerTx)
		suite.Nil(err)
		suite.False(vErr.HasAny())
	}
}

func returnAllBrokerBrokerTxs(suite *ModelSuite) []models.BrokerBrokerTransaction {
	brokerTxs := []models.BrokerBrokerTransaction{}
	suite.DB.RawQuery("SELECT * FROM broker_broker_transactions").All(&brokerTxs)
	return brokerTxs
}

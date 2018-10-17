package jobs_test

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/utils"
	"github.com/oysterprotocol/brokernode/utils/eth_gateway"
	"github.com/shopspring/decimal"
	"math/big"
)

var (
	totalCost                                   = decimal.NewFromFloat(float64(0.015625))
	hasCalledCheckPRLBalance_checkAlphaPayments = false
	hasCalledCheckETHBalance_checkAlphaPayments = false
	hasCalledCalculateGas_checkAlphaPayments    = false
	hasCalledSendETH_checkAlphaPayments         = false
	hasCalledSendPRL_checkAlphaPayments         = false
)

func resetTestVariables_checkAlphaPayments(suite *JobsSuite) {

	suite.DB.RawQuery("DELETE FROM broker_broker_transactions").All(&[]models.BrokerBrokerTransaction{})

	hasCalledCheckPRLBalance_checkAlphaPayments = false
	hasCalledCheckETHBalance_checkAlphaPayments = false
	hasCalledCalculateGas_checkAlphaPayments = false
	hasCalledSendETH_checkAlphaPayments = false
	hasCalledSendPRL_checkAlphaPayments = false

	jobs.EthWrapper = eth_gateway.EthWrapper
}

func (suite *JobsSuite) Test_CheckPaymentToAlpha_no_prl_balance() {
	resetTestVariables_checkAlphaPayments(suite)
	jobs.EthWrapper.CheckPRLBalance = func(address common.Address) *big.Int {
		hasCalledCheckPRLBalance_checkAlphaPayments = true
		return big.NewInt(0)
	}

	generateBrokerBrokerTransactions(suite,
		models.SessionTypeAlpha,
		models.BrokerTxAlphaPaymentPending,
		1)

	generateBrokerBrokerTransactions(suite,
		models.SessionTypeBeta,
		models.BrokerTxAlphaPaymentPending,
		1)

	jobs.CheckPaymentToAlpha()

	brokerTxs := returnAllBrokerBrokerTxs(suite)
	suite.Equal(2, len(brokerTxs))

	for _, brokerTx := range brokerTxs {
		suite.Equal(models.BrokerTxAlphaPaymentPending, brokerTx.PaymentStatus)
	}

	suite.True(hasCalledCheckPRLBalance_checkAlphaPayments)
}

func (suite *JobsSuite) Test_CheckPaymentToAlpha_prl_arrived_alpha_session() {
	resetTestVariables_checkAlphaPayments(suite)
	jobs.EthWrapper.CheckPRLBalance = func(address common.Address) *big.Int {
		hasCalledCheckPRLBalance_checkAlphaPayments = true
		float64Cost, _ := totalCost.Float64()
		bigFloatCost := big.NewFloat(float64Cost)
		totalCostInWei := oyster_utils.ConvertToWeiUnit(bigFloatCost)
		return totalCostInWei
	}

	generateBrokerBrokerTransactions(suite,
		models.SessionTypeAlpha,
		models.BrokerTxAlphaPaymentPending,
		2)

	jobs.CheckPaymentToAlpha()

	brokerTxs := returnAllBrokerBrokerTxs(suite)
	suite.Equal(2, len(brokerTxs))

	for _, brokerTx := range brokerTxs {
		suite.Equal(models.BrokerTxAlphaPaymentConfirmed, brokerTx.PaymentStatus)
	}

	suite.True(hasCalledCheckPRLBalance_checkAlphaPayments)
}

func (suite *JobsSuite) Test_CheckPaymentToAlpha_prl_arrived_beta_session() {
	resetTestVariables_checkAlphaPayments(suite)
	jobs.EthWrapper.CheckPRLBalance = func(address common.Address) *big.Int {
		hasCalledCheckPRLBalance_checkAlphaPayments = true
		float64Cost, _ := totalCost.Float64()
		bigFloatCost := big.NewFloat(float64Cost)
		totalCostInWei := oyster_utils.ConvertToWeiUnit(bigFloatCost)
		return totalCostInWei
	}

	generateBrokerBrokerTransactions(suite,
		models.SessionTypeBeta,
		models.BrokerTxAlphaPaymentPending,
		2)

	jobs.CheckPaymentToAlpha()

	brokerTxs := returnAllBrokerBrokerTxs(suite)
	suite.Equal(2, len(brokerTxs))

	for _, brokerTx := range brokerTxs {

		suite.Equal(models.BrokerTxBetaPaymentPending, brokerTx.PaymentStatus)

	}

	suite.True(hasCalledCheckPRLBalance_checkAlphaPayments)
}

func (suite *JobsSuite) Test_CheckPaymentToAlpha_prl_arrived_update_upload_session() {
	resetTestVariables_checkAlphaPayments(suite)
	jobs.EthWrapper.CheckPRLBalance = func(address common.Address) *big.Int {
		hasCalledCheckPRLBalance_checkAlphaPayments = true
		float64Cost, _ := totalCost.Float64()
		bigFloatCost := big.NewFloat(float64Cost)
		totalCostInWei := oyster_utils.ConvertToWeiUnit(bigFloatCost)
		return totalCostInWei
	}

	generateBrokerBrokerTransactions(suite,
		models.SessionTypeBeta,
		models.BrokerTxAlphaPaymentPending,
		1)

	brokerTxs := returnAllBrokerBrokerTxs(suite)
	suite.Equal(1, len(brokerTxs))

	uploadSession := models.UploadSession{
		GenesisHash:    brokerTxs[0].GenesisHash,
		PaymentStatus:  models.PaymentStatusInvoiced,
		NumChunks:      10,
		FileSizeBytes:  3000,
		Type:           models.SessionTypeBeta,
		TreasureStatus: models.TreasureInDataMapComplete,
	}

	uploadSession.StartUploadSession()

	suite.DB.Where("genesis_hash = ?", brokerTxs[0].GenesisHash).First(&uploadSession)
	suite.Equal(models.PaymentStatusInvoiced, uploadSession.PaymentStatus)

	jobs.CheckPaymentToAlpha()

	uploadSession = models.UploadSession{}
	suite.DB.Where("genesis_hash = ?", brokerTxs[0].GenesisHash).First(&uploadSession)
	suite.Equal(models.PaymentStatusConfirmed, uploadSession.PaymentStatus)

	suite.True(hasCalledCheckPRLBalance_checkAlphaPayments)
}

func (suite *JobsSuite) Test_SendGasToAlphaTransactionAddress_do_not_need_more_gas() {
	resetTestVariables_checkAlphaPayments(suite)
	jobs.EthWrapper.CheckETHBalance = func(addr common.Address) *big.Int {
		hasCalledCheckETHBalance_checkAlphaPayments = true
		// give a large balance so we won't need to send more gas
		return big.NewInt(999999999999999)
	}
	jobs.EthWrapper.CalculateGasNeeded = func(desiredGasLimit uint64) (*big.Int, error) {
		hasCalledCalculateGas_checkAlphaPayments = true
		gasPrice := oyster_utils.ConvertGweiToWei(big.NewInt(1))
		gasToSend := new(big.Int).Mul(gasPrice, big.NewInt(int64(desiredGasLimit)))
		return gasToSend, nil
	}
	jobs.EthWrapper.SendETH = func(fromAddress common.Address, fromPrivKey *ecdsa.PrivateKey, toAddress common.Address,
		gas *big.Int) (types.Transactions, string, int64, error) {
		hasCalledSendETH_checkAlphaPayments = true
		return types.Transactions{}, "111111", 1, nil
	}

	generateBrokerBrokerTransactions(suite,
		models.SessionTypeAlpha,
		models.BrokerTxAlphaPaymentConfirmed,
		1)

	jobs.SendGasToAlphaTransactionAddress()

	brokerTxs := returnAllBrokerBrokerTxs(suite)
	suite.Equal(1, len(brokerTxs))

	suite.Equal(models.BrokerTxGasPaymentConfirmed, brokerTxs[0].PaymentStatus)

	suite.True(hasCalledCheckETHBalance_checkAlphaPayments)
	suite.True(hasCalledCalculateGas_checkAlphaPayments)
	suite.False(hasCalledSendETH_checkAlphaPayments)
}

func (suite *JobsSuite) Test_SendGasToAlphaTransactionAddress_gas_needed() {
	resetTestVariables_checkAlphaPayments(suite)
	jobs.EthWrapper.CheckETHBalance = func(addr common.Address) *big.Int {
		hasCalledCheckETHBalance_checkAlphaPayments = true
		// give a 0 balance so we will need gas
		return big.NewInt(0)
	}
	jobs.EthWrapper.CalculateGasNeeded = func(desiredGasLimit uint64) (*big.Int, error) {
		hasCalledCalculateGas_checkAlphaPayments = true
		gasPrice := oyster_utils.ConvertGweiToWei(big.NewInt(1))
		gasToSend := new(big.Int).Mul(gasPrice, big.NewInt(int64(desiredGasLimit)))
		return gasToSend, nil
	}
	jobs.EthWrapper.SendETH = func(fromAddress common.Address, fromPrivKey *ecdsa.PrivateKey, toAddress common.Address,
		gas *big.Int) (types.Transactions, string, int64, error) {
		hasCalledSendETH_checkAlphaPayments = true
		return types.Transactions{}, "111111", 1, nil
	}

	generateBrokerBrokerTransactions(suite,
		models.SessionTypeAlpha,
		models.BrokerTxAlphaPaymentConfirmed,
		1)

	jobs.SendGasToAlphaTransactionAddress()

	brokerTxs := returnAllBrokerBrokerTxs(suite)
	suite.Equal(1, len(brokerTxs))

	suite.Equal(models.BrokerTxGasPaymentPending, brokerTxs[0].PaymentStatus)

	suite.True(hasCalledCheckETHBalance_checkAlphaPayments)
	suite.True(hasCalledCalculateGas_checkAlphaPayments)
	suite.True(hasCalledSendETH_checkAlphaPayments)
}

func (suite *JobsSuite) Test_CheckGasPayments_gas_not_arrived() {
	resetTestVariables_checkAlphaPayments(suite)
	jobs.EthWrapper.CheckETHBalance = func(addr common.Address) *big.Int {
		hasCalledCheckETHBalance_checkAlphaPayments = true
		// give a 0 balance so we will need gas
		return big.NewInt(0)
	}
	jobs.EthWrapper.CalculateGasNeeded = func(desiredGasLimit uint64) (*big.Int, error) {
		hasCalledCalculateGas_checkAlphaPayments = true
		gasPrice := oyster_utils.ConvertGweiToWei(big.NewInt(1))
		gasToSend := new(big.Int).Mul(gasPrice, big.NewInt(int64(desiredGasLimit)))
		return gasToSend, nil
	}

	generateBrokerBrokerTransactions(suite,
		models.SessionTypeAlpha,
		models.BrokerTxGasPaymentPending,
		1)

	jobs.CheckGasPayments()

	brokerTxs := returnAllBrokerBrokerTxs(suite)
	suite.Equal(1, len(brokerTxs))

	suite.Equal(models.BrokerTxGasPaymentPending, brokerTxs[0].PaymentStatus)

	suite.True(hasCalledCheckETHBalance_checkAlphaPayments)
	suite.True(hasCalledCalculateGas_checkAlphaPayments)
}

func (suite *JobsSuite) Test_CheckGasPayments_gas_arrived() {
	resetTestVariables_checkAlphaPayments(suite)
	jobs.EthWrapper.CheckETHBalance = func(addr common.Address) *big.Int {
		hasCalledCheckETHBalance_checkAlphaPayments = true
		// give a large balance so we will not need gas
		return big.NewInt(999999999999)
	}
	jobs.EthWrapper.CalculateGasNeeded = func(desiredGasLimit uint64) (*big.Int, error) {
		hasCalledCalculateGas_checkAlphaPayments = true
		gasPrice := oyster_utils.ConvertGweiToWei(big.NewInt(1))
		gasToSend := new(big.Int).Mul(gasPrice, big.NewInt(int64(desiredGasLimit)))
		return gasToSend, nil
	}

	generateBrokerBrokerTransactions(suite,
		models.SessionTypeAlpha,
		models.BrokerTxGasPaymentPending,
		1)

	jobs.CheckGasPayments()

	brokerTxs := returnAllBrokerBrokerTxs(suite)
	suite.Equal(1, len(brokerTxs))

	suite.Equal(models.BrokerTxGasPaymentPending, brokerTxs[0].PaymentStatus)

	suite.True(hasCalledCheckETHBalance_checkAlphaPayments)
	suite.True(hasCalledCalculateGas_checkAlphaPayments)
}

func (suite *JobsSuite) Test_SendPaymentToBeta_payment_already_arrived() {
	resetTestVariables_checkAlphaPayments(suite)
	jobs.EthWrapper.CreateSendPRLMessage = eth_gateway.EthWrapper.CreateSendPRLMessage
	jobs.EthWrapper.CheckPRLBalance = func(addr common.Address) *big.Int {
		hasCalledCheckPRLBalance_checkAlphaPayments = true
		// give a large balance so we will see the payment as already arrived
		return big.NewInt(999999999999)
	}
	jobs.EthWrapper.SendPRLFromOyster = func(msg eth_gateway.OysterCallMsg) (bool, string, int64) {
		hasCalledSendPRL_checkAlphaPayments = true
		return true, "some__transaction_hash", 0
	}

	generateBrokerBrokerTransactions(suite,
		models.SessionTypeAlpha,
		models.BrokerTxGasPaymentConfirmed,
		1)

	jobs.SendPaymentToBeta()

	brokerTxs := returnAllBrokerBrokerTxs(suite)
	suite.Equal(1, len(brokerTxs))

	suite.Equal(models.BrokerTxBetaPaymentConfirmed, brokerTxs[0].PaymentStatus)

	suite.True(hasCalledCheckPRLBalance_checkAlphaPayments)
	suite.False(hasCalledSendPRL_checkAlphaPayments)
}

func (suite *JobsSuite) Test_SendPaymentToBeta_send_payment() {
	resetTestVariables_checkAlphaPayments(suite)
	jobs.EthWrapper.CreateSendPRLMessage = eth_gateway.EthWrapper.CreateSendPRLMessage
	jobs.EthWrapper.CheckPRLBalance = func(addr common.Address) *big.Int {
		hasCalledCheckPRLBalance_checkAlphaPayments = true
		// give a 0 balance so we will send the PRL
		return big.NewInt(0)
	}
	jobs.EthWrapper.SendPRLFromOyster = func(msg eth_gateway.OysterCallMsg) (bool, string, int64) {
		hasCalledSendPRL_checkAlphaPayments = true
		return true, "some__transaction_hash", 0
	}

	generateBrokerBrokerTransactions(suite,
		models.SessionTypeAlpha,
		models.BrokerTxGasPaymentConfirmed,
		1)

	jobs.SendPaymentToBeta()

	brokerTxs := returnAllBrokerBrokerTxs(suite)
	suite.Equal(1, len(brokerTxs))

	suite.Equal(models.BrokerTxBetaPaymentPending, brokerTxs[0].PaymentStatus)

	suite.True(hasCalledCheckPRLBalance_checkAlphaPayments)
	suite.True(hasCalledSendPRL_checkAlphaPayments)
}

func generateBrokerBrokerTransactions(suite *JobsSuite,
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

func returnAllBrokerBrokerTxs(suite *JobsSuite) []models.BrokerBrokerTransaction {
	brokerTxs := []models.BrokerBrokerTransaction{}
	suite.DB.RawQuery("SELECT * FROM broker_broker_transactions").All(&brokerTxs)
	return brokerTxs
}

package jobs_test

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"math/big"
	"time"
)

var (
	hasCalledCheckPRLBalance_checkBetaPayments = false
	hasCalledCheckETHBalance_checkBetaPayments = false
	hasCalledCalculateGas_checkBetaPayments    = false
	hasCalledSendETH_checkBetaPayments         = false
)

func resetTestVariables_checkBetaPayments(suite *JobsSuite) {

	suite.DB.RawQuery("DELETE FROM broker_broker_transactions").All(&[]models.BrokerBrokerTransaction{})

	hasCalledCheckPRLBalance_checkBetaPayments = false
	hasCalledCheckETHBalance_checkBetaPayments = false
	hasCalledCalculateGas_checkBetaPayments = false
	hasCalledSendETH_checkBetaPayments = false

	jobs.EthWrapper = services.EthWrapper
}

func (suite *JobsSuite) Test_CheckPaymentToBeta_payment_arrived() {
	resetTestVariables_checkBetaPayments(suite)

	jobs.EthWrapper.CheckPRLBalance = func(addr common.Address) *big.Int {
		hasCalledCheckPRLBalance_checkBetaPayments = true
		// give a balance that is half the total cost
		float64Cost, _ := totalCost.Float64()
		totalCostInWei := oyster_utils.ConvertToWeiUnit(big.NewFloat(float64Cost))
		halfOfTotalCost := new(big.Int).Quo(totalCostInWei, big.NewInt(2))
		return halfOfTotalCost
	}

	generateBrokerBrokerTransactions(suite,
		models.SessionTypeAlpha,
		models.BrokerTxBetaPaymentPending,
		1)

	generateBrokerBrokerTransactions(suite,
		models.SessionTypeBeta,
		models.BrokerTxBetaPaymentPending,
		1)

	jobs.CheckPaymentToBeta()

	brokerTxs := returnAllBrokerBrokerTxs(suite)
	suite.Equal(2, len(brokerTxs))

	for _, brokerTx := range brokerTxs {
		suite.Equal(models.BrokerTxBetaPaymentConfirmed, brokerTx.PaymentStatus)

	}

	suite.True(hasCalledCheckPRLBalance_checkBetaPayments)
}

func (suite *JobsSuite) Test_CheckPaymentToBeta_payment_still_pending() {
	resetTestVariables_checkBetaPayments(suite)

	jobs.EthWrapper.CheckPRLBalance = func(addr common.Address) *big.Int {
		hasCalledCheckPRLBalance_checkBetaPayments = true
		// give a 0 balance so it will see the transaction as still pending
		return big.NewInt(0)
	}

	generateBrokerBrokerTransactions(suite,
		models.SessionTypeAlpha,
		models.BrokerTxBetaPaymentPending,
		1)

	generateBrokerBrokerTransactions(suite,
		models.SessionTypeBeta,
		models.BrokerTxBetaPaymentPending,
		1)

	jobs.CheckPaymentToBeta()

	brokerTxs := returnAllBrokerBrokerTxs(suite)
	suite.Equal(2, len(brokerTxs))

	for _, brokerTx := range brokerTxs {
		suite.Equal(models.BrokerTxBetaPaymentPending, brokerTx.PaymentStatus)
	}

	suite.True(hasCalledCheckPRLBalance_checkBetaPayments)
}

func (suite *JobsSuite) Test_HandleTimedOutBetaPaymentIfBeta() {
	resetTestVariables_checkBetaPayments(suite)

	generateBrokerBrokerTransactions(suite,
		models.SessionTypeBeta,
		models.BrokerTxBetaPaymentPending,
		1)

	jobs.HandleTimedOutBetaPaymentIfBeta(time.Duration(6 * time.Hour))

	brokerTxs := returnAllBrokerBrokerTxs(suite)
	suite.Equal(0, len(brokerTxs))
}

func (suite *JobsSuite) HandleTimedOutTransactionsIfAlpha() {
	resetTestVariables_checkBetaPayments(suite)

	generateBrokerBrokerTransactions(suite,
		models.SessionTypeAlpha,
		models.BrokerTxGasPaymentPending,
		1)

	generateBrokerBrokerTransactions(suite,
		models.SessionTypeAlpha,
		models.BrokerTxBetaPaymentPending,
		1)

	jobs.HandleTimedOutTransactionsIfAlpha(time.Duration(6 * time.Hour))

	brokerTxs := returnAllBrokerBrokerTxs(suite)
	suite.Equal(2, len(brokerTxs))

	oneEqualToAlphaConfirmed := false
	oneEqualToGasConfirmed := false

	for _, brokerTx := range brokerTxs {
		if brokerTx.PaymentStatus == models.BrokerTxAlphaPaymentConfirmed {
			oneEqualToAlphaConfirmed = true
		}
		if brokerTx.PaymentStatus == models.BrokerTxGasPaymentConfirmed {
			oneEqualToGasConfirmed = true
		}
	}

	suite.True(oneEqualToAlphaConfirmed)
	suite.True(oneEqualToGasConfirmed)
}

func (suite *JobsSuite) HandleErrorTransactionsIfAlpha() {
	resetTestVariables_checkBetaPayments(suite)

	generateBrokerBrokerTransactions(suite,
		models.SessionTypeAlpha,
		models.BrokerTxGasPaymentError,
		1)

	generateBrokerBrokerTransactions(suite,
		models.SessionTypeAlpha,
		models.BrokerTxBetaPaymentError,
		1)

	jobs.HandleErrorTransactionsIfAlpha()

	brokerTxs := returnAllBrokerBrokerTxs(suite)
	suite.Equal(2, len(brokerTxs))

	oneEqualToAlphaConfirmed := false
	oneEqualToGasConfirmed := false

	for _, brokerTx := range brokerTxs {
		if brokerTx.PaymentStatus == models.BrokerTxAlphaPaymentConfirmed {
			oneEqualToAlphaConfirmed = true
		}
		if brokerTx.PaymentStatus == models.BrokerTxGasPaymentConfirmed {
			oneEqualToGasConfirmed = true
		}
	}

	suite.True(oneEqualToAlphaConfirmed)
	suite.True(oneEqualToGasConfirmed)
}

func (suite *JobsSuite) PurgeCompletedTransactions() {
	resetTestVariables_checkBetaPayments(suite)

	generateBrokerBrokerTransactions(suite,
		models.SessionTypeAlpha,
		models.BrokerTxBetaPaymentConfirmed,
		1)

	generateBrokerBrokerTransactions(suite,
		models.SessionTypeBeta,
		models.BrokerTxBetaPaymentConfirmed,
		1)

	jobs.PurgeCompletedTransactions()

	brokerTxs := returnAllBrokerBrokerTxs(suite)
	suite.Equal(0, len(brokerTxs))
}

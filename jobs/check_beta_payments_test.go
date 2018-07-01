package jobs_test

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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

	suite.DB.RawQuery("DELETE from broker_broker_transactions").All(&[]models.BrokerBrokerTransaction{})

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

func (suite *JobsSuite) Test_StartAndCheckGasReclaims_no_gas_to_reclaim() {
	resetTestVariables_checkBetaPayments(suite)

	jobs.EthWrapper.CheckETHBalance = func(addr common.Address) *big.Int {
		hasCalledCheckETHBalance_checkBetaPayments = true
		// give a 0 balance so it will not try to reclaim gas
		return big.NewInt(0)
	}
	jobs.EthWrapper.CalculateGasNeeded = func(desiredGasLimit uint64) (*big.Int, error) {
		hasCalledCalculateGas_checkBetaPayments = true
		gasPrice := oyster_utils.ConvertGweiToWei(big.NewInt(1))
		gasToSend := new(big.Int).Mul(gasPrice, big.NewInt(int64(desiredGasLimit)))
		return gasToSend, nil
	}
	jobs.EthWrapper.SendETH = func(fromAddress common.Address, fromPrivKey *ecdsa.PrivateKey, toAddress common.Address,
		gas *big.Int) (types.Transactions, string, int64, error) {
		hasCalledSendETH_checkBetaPayments = true
		return types.Transactions{}, "111111", 1, nil
	}

	generateBrokerBrokerTransactions(suite,
		models.SessionTypeAlpha,
		models.BrokerTxBetaPaymentConfirmed,
		1)

	generateBrokerBrokerTransactions(suite,
		models.SessionTypeAlpha,
		models.BrokerTxGasReclaimPending,
		1)

	jobs.StartAndCheckGasReclaims()

	brokerTxs := returnAllBrokerBrokerTxs(suite)
	suite.Equal(2, len(brokerTxs))

	for _, brokerTx := range brokerTxs {
		suite.Equal(models.BrokerTxGasReclaimConfirmed, brokerTx.PaymentStatus)
	}

	suite.True(hasCalledCheckETHBalance_checkBetaPayments)
	suite.False(hasCalledCalculateGas_checkBetaPayments)
	suite.False(hasCalledSendETH_checkBetaPayments)
}

func (suite *JobsSuite) Test_StartAndCheckGasReclaims_not_enough_gas_to_reclaim() {
	resetTestVariables_checkBetaPayments(suite)

	jobs.EthWrapper.CheckETHBalance = func(addr common.Address) *big.Int {
		hasCalledCheckETHBalance_checkBetaPayments = true
		// give a small balance so it will not try to reclaim gas
		return big.NewInt(1)
	}
	jobs.EthWrapper.CalculateGasNeeded = func(desiredGasLimit uint64) (*big.Int, error) {
		hasCalledCalculateGas_checkBetaPayments = true
		gasPrice := oyster_utils.ConvertGweiToWei(big.NewInt(1))
		gasToSend := new(big.Int).Mul(gasPrice, big.NewInt(int64(desiredGasLimit)))
		return gasToSend, nil
	}
	jobs.EthWrapper.SendETH = func(fromAddress common.Address, fromPrivKey *ecdsa.PrivateKey, toAddress common.Address,
		gas *big.Int) (types.Transactions, string, int64, error) {
		hasCalledSendETH_checkBetaPayments = true
		return types.Transactions{}, "111111", 1, nil
	}

	generateBrokerBrokerTransactions(suite,
		models.SessionTypeAlpha,
		models.BrokerTxBetaPaymentConfirmed,
		1)

	generateBrokerBrokerTransactions(suite,
		models.SessionTypeAlpha,
		models.BrokerTxGasReclaimPending,
		1)

	jobs.StartAndCheckGasReclaims()

	brokerTxs := returnAllBrokerBrokerTxs(suite)
	suite.Equal(2, len(brokerTxs))

	for _, brokerTx := range brokerTxs {
		suite.Equal(models.BrokerTxGasReclaimConfirmed, brokerTx.PaymentStatus)
	}

	suite.True(hasCalledCheckETHBalance_checkBetaPayments)
	suite.True(hasCalledCalculateGas_checkBetaPayments)
	suite.False(hasCalledSendETH_checkBetaPayments)
}

func (suite *JobsSuite) Test_StartAndCheckGasReclaims_enough_gas_to_reclaim() {
	resetTestVariables_checkBetaPayments(suite)

	jobs.EthWrapper.CheckETHBalance = func(addr common.Address) *big.Int {
		hasCalledCheckETHBalance_checkBetaPayments = true
		// give a large balance so it will try to reclaim the gas
		return big.NewInt(999999999)
	}
	jobs.EthWrapper.CalculateGasNeeded = func(desiredGasLimit uint64) (*big.Int, error) {
		// return a tiny amount of gas needed for the transaction so it will try to
		// reclaim the gas
		hasCalledCalculateGas_checkBetaPayments = true
		return big.NewInt(1), nil
	}
	jobs.EthWrapper.SendETH = func(fromAddress common.Address, fromPrivKey *ecdsa.PrivateKey, toAddress common.Address,
		gas *big.Int) (types.Transactions, string, int64, error) {
		hasCalledSendETH_checkBetaPayments = true
		return types.Transactions{}, "111111", 1, nil
	}

	generateBrokerBrokerTransactions(suite,
		models.SessionTypeAlpha,
		models.BrokerTxBetaPaymentConfirmed,
		1)

	generateBrokerBrokerTransactions(suite,
		models.SessionTypeAlpha,
		models.BrokerTxGasReclaimPending,
		1)

	jobs.StartAndCheckGasReclaims()

	brokerTxs := returnAllBrokerBrokerTxs(suite)
	suite.Equal(2, len(brokerTxs))

	for _, brokerTx := range brokerTxs {
		suite.Equal(models.BrokerTxGasReclaimPending, brokerTx.PaymentStatus)
	}

	suite.True(hasCalledCheckETHBalance_checkBetaPayments)
	suite.True(hasCalledCalculateGas_checkBetaPayments)
	suite.True(hasCalledSendETH_checkBetaPayments)
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

	generateBrokerBrokerTransactions(suite,
		models.SessionTypeAlpha,
		models.BrokerTxGasReclaimPending,
		1)

	jobs.HandleTimedOutTransactionsIfAlpha(time.Duration(6 * time.Hour))

	brokerTxs := returnAllBrokerBrokerTxs(suite)
	suite.Equal(3, len(brokerTxs))

	oneEqualToAlphaConfirmed := false
	oneEqualToGasConfirmed := false
	oneEqualToBetaConfirmed := false

	for _, brokerTx := range brokerTxs {
		if brokerTx.PaymentStatus == models.BrokerTxAlphaPaymentConfirmed {
			oneEqualToAlphaConfirmed = true
		}
		if brokerTx.PaymentStatus == models.BrokerTxGasPaymentConfirmed {
			oneEqualToGasConfirmed = true
		}
		if brokerTx.PaymentStatus == models.BrokerTxBetaPaymentConfirmed {
			oneEqualToBetaConfirmed = true
		}
	}

	suite.True(oneEqualToAlphaConfirmed)
	suite.True(oneEqualToGasConfirmed)
	suite.True(oneEqualToBetaConfirmed)
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

	generateBrokerBrokerTransactions(suite,
		models.SessionTypeAlpha,
		models.BrokerTxGasReclaimError,
		1)

	jobs.HandleErrorTransactionsIfAlpha()

	brokerTxs := returnAllBrokerBrokerTxs(suite)
	suite.Equal(3, len(brokerTxs))

	oneEqualToAlphaConfirmed := false
	oneEqualToGasConfirmed := false
	oneEqualToBetaConfirmed := false

	for _, brokerTx := range brokerTxs {
		if brokerTx.PaymentStatus == models.BrokerTxAlphaPaymentConfirmed {
			oneEqualToAlphaConfirmed = true
		}
		if brokerTx.PaymentStatus == models.BrokerTxGasPaymentConfirmed {
			oneEqualToGasConfirmed = true
		}
		if brokerTx.PaymentStatus == models.BrokerTxBetaPaymentConfirmed {
			oneEqualToBetaConfirmed = true
		}
	}

	suite.True(oneEqualToAlphaConfirmed)
	suite.True(oneEqualToGasConfirmed)
	suite.True(oneEqualToBetaConfirmed)
}

func (suite *JobsSuite) PurgeCompletedTransactions() {
	resetTestVariables_checkBetaPayments(suite)

	generateBrokerBrokerTransactions(suite,
		models.SessionTypeAlpha,
		models.BrokerTxGasReclaimConfirmed,
		1)

	jobs.PurgeCompletedTransactions()

	brokerTxs := returnAllBrokerBrokerTxs(suite)
	suite.Equal(0, len(brokerTxs))
}

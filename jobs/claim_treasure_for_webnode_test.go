package jobs_test

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"github.com/pkg/errors"
	"math/big"
	"time"
)

var (
	hasCalledCheckPRLBalance_claimTreasureForWebnode           = false
	hasCalledClaimPRL_claimTreasureForWebnode                  = false
	hasCalledCheckClaimClock_claimTreasureForWebnode           = false
	hasCalledCheckETHBalance_claimTreasureForWebnode           = false
	hasCalledCalculateGas_claimTreasureForWebnode              = false
	hasCalledSendETH_claimTreasureForWebnode                   = false
	hasCalledCheckIfWorthReclaimingGas_claimTreasureForWebnode = false
	calculateGasNeededCalls_claimTreasureForWebnode            = 0
	chechPRLBalanceCalls_claimTreasureForWebnode               = 0
)

func resetTestVariables_claimTreasureForWebnode(suite *JobsSuite) {
	suite.DB.RawQuery("DELETE FROM webnode_treasure_claims").All(
		&[]models.WebnodeTreasureClaim{})

	hasCalledCheckPRLBalance_claimTreasureForWebnode = false
	hasCalledClaimPRL_claimTreasureForWebnode = false
	hasCalledCheckClaimClock_claimTreasureForWebnode = false
	hasCalledCheckETHBalance_claimTreasureForWebnode = false
	hasCalledCalculateGas_claimTreasureForWebnode = false
	hasCalledSendETH_claimTreasureForWebnode = false
	hasCalledCheckIfWorthReclaimingGas_claimTreasureForWebnode = false

	calculateGasNeededCalls_claimTreasureForWebnode = 0
	chechPRLBalanceCalls_claimTreasureForWebnode = 0

	jobs.EthWrapper = services.EthWrapper
}

func (suite *JobsSuite) Test_CheckOngoingGasTransactions_gas_has_arrived() {
	resetTestVariables_claimTreasureForWebnode(suite)

	jobs.EthWrapper.CheckETHBalance = func(addr common.Address) *big.Int {
		hasCalledCheckETHBalance_claimTreasureForWebnode = true
		/* give a large balance so it will think the gas has arrived */
		return big.NewInt(999999999)
	}
	jobs.EthWrapper.CalculateGasNeeded = func(desiredGasLimit uint64) (*big.Int, error) {
		/* return a small calculation of gas needed for PRL claim transaction so it will
		think the sent gas has arrived */
		hasCalledCalculateGas_claimTreasureForWebnode = true
		return big.NewInt(1), nil
	}

	generateWebnodeTreasureClaims(
		suite,
		models.PRLClaimNotStarted,
		models.GasTransferProcessing,
		1,
		4)

	jobs.CheckOngoingGasTransactions()

	treasureClaims := getAllWebnodeTreasureClaims(suite)
	suite.Equal(4, len(treasureClaims))

	for _, treasureClaim := range treasureClaims {
		suite.Equal(models.GasTransferSuccess, treasureClaim.GasStatus)
	}

	suite.True(hasCalledCheckETHBalance_claimTreasureForWebnode)
	suite.True(hasCalledCalculateGas_claimTreasureForWebnode)
}

func (suite *JobsSuite) Test_CheckOngoingGasTransactions_gas_has_not_arrived() {
	resetTestVariables_claimTreasureForWebnode(suite)

	jobs.EthWrapper.CheckETHBalance = func(addr common.Address) *big.Int {
		hasCalledCheckETHBalance_claimTreasureForWebnode = true
		/* return a small balance so it will think the gas has not arrived */
		return big.NewInt(1)
	}
	jobs.EthWrapper.CalculateGasNeeded = func(desiredGasLimit uint64) (*big.Int, error) {
		/* return a large calculation of gas needed for PRL claim so it will
		think the gas has not arrived */
		hasCalledCalculateGas_claimTreasureForWebnode = true
		return big.NewInt(9999999), nil
	}

	generateWebnodeTreasureClaims(
		suite,
		models.PRLClaimNotStarted,
		models.GasTransferProcessing,
		1,
		4)

	jobs.CheckOngoingGasTransactions()

	treasureClaims := getAllWebnodeTreasureClaims(suite)
	suite.Equal(4, len(treasureClaims))

	for _, treasureClaim := range treasureClaims {
		suite.Equal(models.GasTransferProcessing, treasureClaim.GasStatus)
	}

	suite.True(hasCalledCheckETHBalance_claimTreasureForWebnode)
	suite.True(hasCalledCalculateGas_claimTreasureForWebnode)
}

func (suite *JobsSuite) Test_CheckOngoingPRLClaims_claim_has_succeeded() {
	resetTestVariables_claimTreasureForWebnode(suite)

	jobs.EthWrapper.CheckClaimClock = func(address common.Address) (*big.Int, error) {
		/* return a claim clock value different than starting claim clock value
		so it will think the claim has succeeded */
		hasCalledCheckClaimClock_claimTreasureForWebnode = true
		return big.NewInt(9999999), nil
	}

	generateWebnodeTreasureClaims(
		suite,
		models.PRLClaimProcessing,
		models.GasTransferSuccess,
		1,
		4)

	jobs.CheckOngoingPRLClaims()

	treasureClaims := getAllWebnodeTreasureClaims(suite)
	suite.Equal(4, len(treasureClaims))

	for _, treasureClaim := range treasureClaims {
		suite.Equal(models.PRLClaimSuccess, treasureClaim.ClaimPRLStatus)
	}

	suite.True(hasCalledCheckClaimClock_claimTreasureForWebnode)
}

func (suite *JobsSuite) Test_CheckOngoingPRLClaims_claim_still_processing() {
	resetTestVariables_claimTreasureForWebnode(suite)

	jobs.EthWrapper.CheckClaimClock = func(address common.Address) (*big.Int, error) {
		/* return a claim clock value same as starting value so it will
		think the claim is still in process */
		hasCalledCheckClaimClock_claimTreasureForWebnode = true
		return big.NewInt(1), nil
	}

	generateWebnodeTreasureClaims(
		suite,
		models.PRLClaimProcessing,
		models.GasTransferSuccess,
		1,
		4)

	jobs.CheckOngoingPRLClaims()

	treasureClaims := getAllWebnodeTreasureClaims(suite)
	suite.Equal(4, len(treasureClaims))

	for _, treasureClaim := range treasureClaims {
		suite.Equal(models.PRLClaimProcessing, treasureClaim.ClaimPRLStatus)
	}

	suite.True(hasCalledCheckClaimClock_claimTreasureForWebnode)
}

func (suite *JobsSuite) Test_CheckOngoingGasReclaims_all_gas_reclaimed() {
	resetTestVariables_claimTreasureForWebnode(suite)

	jobs.EthWrapper.CheckETHBalance = func(addr common.Address) *big.Int {
		hasCalledCheckETHBalance_claimTreasureForWebnode = true
		/* return a 0 balance so it will think all gas has been reclaimed */
		return big.NewInt(0)
	}
	jobs.EthWrapper.CalculateGasNeeded = func(desiredGasLimit uint64) (*big.Int, error) {
		/* this should not get called because eth balance was 0 */
		hasCalledCalculateGas_claimTreasureForWebnode = true
		return big.NewInt(9999999), nil
	}

	generateWebnodeTreasureClaims(
		suite,
		models.PRLClaimSuccess,
		models.GasTransferLeftoversReclaimProcessing,
		1,
		4)

	jobs.CheckOngoingGasReclaims()

	treasureClaims := getAllWebnodeTreasureClaims(suite)
	suite.Equal(4, len(treasureClaims))

	for _, treasureClaim := range treasureClaims {
		suite.Equal(models.GasTransferLeftoversReclaimSuccess, treasureClaim.GasStatus)
	}

	suite.True(hasCalledCheckETHBalance_claimTreasureForWebnode)
	suite.False(hasCalledCalculateGas_claimTreasureForWebnode)
}

func (suite *JobsSuite) Test_CheckOngoingGasReclaims_not_enough_gas_to_reclaim() {
	resetTestVariables_claimTreasureForWebnode(suite)

	jobs.EthWrapper.CheckETHBalance = func(addr common.Address) *big.Int {
		hasCalledCheckETHBalance_claimTreasureForWebnode = true
		/* return a small balance so it will decide it is not worth reclaiming */
		return big.NewInt(1)
	}
	jobs.EthWrapper.CalculateGasNeeded = func(desiredGasLimit uint64) (*big.Int, error) {
		/* return a calculation much larger than eth balance so it will think the eth in the
		balance is not worth reclaiming*/
		hasCalledCalculateGas_claimTreasureForWebnode = true
		return big.NewInt(9999999), nil
	}

	generateWebnodeTreasureClaims(
		suite,
		models.PRLClaimSuccess,
		models.GasTransferLeftoversReclaimProcessing,
		1,
		4)

	jobs.CheckOngoingGasReclaims()

	treasureClaims := getAllWebnodeTreasureClaims(suite)
	suite.Equal(4, len(treasureClaims))

	for _, treasureClaim := range treasureClaims {
		suite.Equal(models.GasTransferLeftoversReclaimSuccess, treasureClaim.GasStatus)
	}

	suite.True(hasCalledCheckETHBalance_claimTreasureForWebnode)
	suite.True(hasCalledCalculateGas_claimTreasureForWebnode)
}

func (suite *JobsSuite) Test_CheckOngoingGasReclaims_reclaim_still_pending() {
	resetTestVariables_claimTreasureForWebnode(suite)

	jobs.EthWrapper.CheckETHBalance = func(addr common.Address) *big.Int {
		hasCalledCheckETHBalance_claimTreasureForWebnode = true
		/* return a large balance so it will think the reclaim is still processing */
		return big.NewInt(9999999)
	}
	jobs.EthWrapper.CalculateGasNeeded = func(desiredGasLimit uint64) (*big.Int, error) {
		/* return a small calculation so it will not declare the gas reclaim successful,
		it will think the reclaim is still processing */
		hasCalledCalculateGas_claimTreasureForWebnode = true
		return big.NewInt(1), nil
	}

	generateWebnodeTreasureClaims(
		suite,
		models.PRLClaimSuccess,
		models.GasTransferLeftoversReclaimProcessing,
		1,
		4)

	jobs.CheckOngoingGasReclaims()

	treasureClaims := getAllWebnodeTreasureClaims(suite)
	suite.Equal(4, len(treasureClaims))

	for _, treasureClaim := range treasureClaims {
		suite.Equal(models.GasTransferLeftoversReclaimProcessing, treasureClaim.GasStatus)
	}

	suite.True(hasCalledCheckETHBalance_claimTreasureForWebnode)
	suite.True(hasCalledCalculateGas_claimTreasureForWebnode)
}

func (suite *JobsSuite) Test_ResendOldETHTransfers_some_timed_out() {
	resetTestVariables_claimTreasureForWebnode(suite)

	jobs.EthWrapper.CalculateGasNeeded = func(desiredGasLimit uint64) (*big.Int, error) {
		/* track number of times this is called */
		calculateGasNeededCalls_claimTreasureForWebnode++
		hasCalledCalculateGas_claimTreasureForWebnode = true
		return big.NewInt(1), errors.New("making this fail")
	}

	generateWebnodeTreasureClaims(
		suite,
		models.PRLClaimNotStarted,
		models.GasTransferProcessing,
		1,
		4)

	jobs.ResendOldETHTransfers(time.Now().Add(5 * time.Minute))

	// if this got called it means "SendGas" got called
	suite.True(hasCalledCalculateGas_claimTreasureForWebnode)
	suite.Equal(1, calculateGasNeededCalls_claimTreasureForWebnode)
}

func (suite *JobsSuite) Test_ResendOldETHTransfers_none_timed_out() {
	resetTestVariables_claimTreasureForWebnode(suite)

	jobs.EthWrapper.CalculateGasNeeded = func(desiredGasLimit uint64) (*big.Int, error) {
		/* track number of times this is called */
		calculateGasNeededCalls_claimTreasureForWebnode++
		hasCalledCalculateGas_claimTreasureForWebnode = true
		return big.NewInt(1), errors.New("making this fail")
	}

	generateWebnodeTreasureClaims(
		suite,
		models.PRLClaimNotStarted,
		models.GasTransferProcessing,
		1,
		4)

	jobs.ResendOldETHTransfers(time.Now().Add(-5 * time.Minute))

	// if this got called it means "SendGas" got called
	suite.False(hasCalledCalculateGas_claimTreasureForWebnode)
	suite.Equal(0, calculateGasNeededCalls_claimTreasureForWebnode)
}

func (suite *JobsSuite) Test_ResendOldPRLClaims_some_timed_out() {
	resetTestVariables_claimTreasureForWebnode(suite)

	jobs.EthWrapper.CheckPRLBalance = func(address common.Address) *big.Int {
		hasCalledCheckPRLBalance_claimTreasureForWebnode = true
		chechPRLBalanceCalls_claimTreasureForWebnode++
		/* return a 0 balance so execution will exit jobs.ClaimPRL early */
		return big.NewInt(0)
	}

	generateWebnodeTreasureClaims(
		suite,
		models.PRLClaimProcessing,
		models.GasTransferSuccess,
		1,
		4)

	/* Using a time in the future so claims will appear timed out */
	jobs.ResendOldPRLClaims(time.Now().Add(5 * time.Minute))

	/* if this got called it means "ClaimPRL" got called */
	suite.True(hasCalledCheckPRLBalance_claimTreasureForWebnode)
	suite.Equal(4, chechPRLBalanceCalls_claimTreasureForWebnode)
}

func (suite *JobsSuite) Test_ResendOldPRLClaims_none_timed_out() {
	resetTestVariables_claimTreasureForWebnode(suite)

	jobs.EthWrapper.CheckPRLBalance = func(address common.Address) *big.Int {
		hasCalledCheckPRLBalance_claimTreasureForWebnode = true
		chechPRLBalanceCalls_claimTreasureForWebnode++
		/* return a 0 balance so execution will exit jobs.ClaimPRL early */
		return big.NewInt(0)
	}

	generateWebnodeTreasureClaims(
		suite,
		models.PRLClaimProcessing,
		models.GasTransferSuccess,
		1,
		4)

	/* Using a time in the past so claims will not appear timed out */
	jobs.ResendOldPRLClaims(time.Now().Add(-5 * time.Minute))

	/* if this got called it means "ClaimPRL" got called */
	suite.False(hasCalledCheckPRLBalance_claimTreasureForWebnode)
	suite.Equal(0, chechPRLBalanceCalls_claimTreasureForWebnode)
}

func (suite *JobsSuite) Test_ResendOldGasReclaims_some_timed_out() {
	resetTestVariables_claimTreasureForWebnode(suite)

	generateWebnodeTreasureClaims(
		suite,
		models.PRLClaimSuccess,
		models.GasTransferLeftoversReclaimProcessing,
		1,
		4)

	/* Using a time in the future so claims will appear timed out */
	jobs.ResendOldGasReclaims(time.Now().Add(5 * time.Minute))

	treasureClaims := []models.WebnodeTreasureClaim{}
	treasureClaims = getAllWebnodeTreasureClaims(suite)

	suite.Equal(4, len(treasureClaims))

	for _, treasureClaim := range treasureClaims {
		/* verify that the claim's gas status got set back to a previous state so it
		will be attempted again */
		suite.Equal(models.GasTransferSuccess, treasureClaim.GasStatus)
	}
}

func (suite *JobsSuite) Test_ResendOldGasReclaims_none_timed_out() {
	resetTestVariables_claimTreasureForWebnode(suite)

	generateWebnodeTreasureClaims(
		suite,
		models.PRLClaimSuccess,
		models.GasTransferLeftoversReclaimProcessing,
		1,
		4)

	/* Using a time in the past so claims will not appear timed out */
	jobs.ResendOldGasReclaims(time.Now().Add(-5 * time.Minute))

	treasureClaims := []models.WebnodeTreasureClaim{}
	treasureClaims = getAllWebnodeTreasureClaims(suite)

	suite.Equal(4, len(treasureClaims))

	for _, treasureClaim := range treasureClaims {
		/* verify that it has not changed the gas status of any of the claims because
		none are timed out */
		suite.Equal(models.GasTransferLeftoversReclaimProcessing, treasureClaim.GasStatus)
	}
}

func (suite *JobsSuite) Test_ResendErroredETHTransfers_some_have_errors() {
	resetTestVariables_claimTreasureForWebnode(suite)

	jobs.EthWrapper.CalculateGasNeeded = func(desiredGasLimit uint64) (*big.Int, error) {
		/* track number of times this is called */
		calculateGasNeededCalls_claimTreasureForWebnode++
		hasCalledCalculateGas_claimTreasureForWebnode = true
		return big.NewInt(1), errors.New("making this fail")
	}

	generateWebnodeTreasureClaims(
		suite,
		models.PRLClaimNotStarted,
		models.GasTransferError,
		1,
		4)

	jobs.ResendErroredETHTransfers()

	// if this got called it means "SendGas" got called
	suite.True(hasCalledCalculateGas_claimTreasureForWebnode)
	suite.Equal(1, calculateGasNeededCalls_claimTreasureForWebnode)
}

func (suite *JobsSuite) Test_ResendErroredETHTransfers_no_errors() {
	resetTestVariables_claimTreasureForWebnode(suite)

	jobs.EthWrapper.CalculateGasNeeded = func(desiredGasLimit uint64) (*big.Int, error) {
		/* track number of times this is called */
		calculateGasNeededCalls_claimTreasureForWebnode++
		hasCalledCalculateGas_claimTreasureForWebnode = true
		return big.NewInt(1), errors.New("making this fail")
	}

	generateWebnodeTreasureClaims(
		suite,
		models.PRLClaimNotStarted,
		models.GasTransferSuccess,
		1,
		4)

	jobs.ResendErroredETHTransfers()

	// if this got called it means "SendGas" got called
	suite.False(hasCalledCalculateGas_claimTreasureForWebnode)
	suite.Equal(0, calculateGasNeededCalls_claimTreasureForWebnode)
}

func (suite *JobsSuite) Test_ResendErroredPRLClaims_some_have_errors() {
	resetTestVariables_claimTreasureForWebnode(suite)

	jobs.EthWrapper.CheckPRLBalance = func(address common.Address) *big.Int {
		hasCalledCheckPRLBalance_claimTreasureForWebnode = true
		chechPRLBalanceCalls_claimTreasureForWebnode++
		/* return a 0 balance so execution will exit jobs.ClaimPRL early */
		return big.NewInt(0)
	}

	generateWebnodeTreasureClaims(
		suite,
		models.PRLClaimError,
		models.GasTransferSuccess,
		1,
		4)

	jobs.ResendErroredPRLClaims()

	/* if this got called it means "ClaimPRL" got called */
	suite.True(hasCalledCheckPRLBalance_claimTreasureForWebnode)
	suite.Equal(4, chechPRLBalanceCalls_claimTreasureForWebnode)
}

func (suite *JobsSuite) Test_ResendErroredPRLClaims_no_errors() {
	resetTestVariables_claimTreasureForWebnode(suite)

	jobs.EthWrapper.CheckPRLBalance = func(address common.Address) *big.Int {
		hasCalledCheckPRLBalance_claimTreasureForWebnode = true
		chechPRLBalanceCalls_claimTreasureForWebnode++
		/* return a 0 balance so execution will execute jobs.ClaimPRL early */
		return big.NewInt(0)
	}

	generateWebnodeTreasureClaims(
		suite,
		models.PRLClaimSuccess,
		models.GasTransferSuccess,
		1,
		4)

	jobs.ResendErroredPRLClaims()

	/* if this got called it means "ClaimPRL" got called */
	suite.False(hasCalledCheckPRLBalance_claimTreasureForWebnode)
	suite.Equal(0, chechPRLBalanceCalls_claimTreasureForWebnode)
}

func (suite *JobsSuite) Test_ResendErroredGasReclaims_some_have_errors() {
	resetTestVariables_claimTreasureForWebnode(suite)

	generateWebnodeTreasureClaims(
		suite,
		models.PRLClaimSuccess,
		models.GasTransferLeftoversReclaimError,
		1,
		4)

	jobs.ResendErroredGasReclaims()

	treasureClaims := []models.WebnodeTreasureClaim{}
	treasureClaims = getAllWebnodeTreasureClaims(suite)

	suite.Equal(4, len(treasureClaims))

	for _, treasureClaim := range treasureClaims {
		/* verify that the claim's gas status got set back to a previous state so it
		will be attempted again */
		suite.Equal(models.GasTransferSuccess, treasureClaim.GasStatus)
	}
}

func (suite *JobsSuite) Test_ResendErroredGasReclaims_no_errors() {
	resetTestVariables_claimTreasureForWebnode(suite)

	generateWebnodeTreasureClaims(
		suite,
		models.PRLClaimSuccess,
		models.GasTransferLeftoversReclaimProcessing,
		1,
		4)

	jobs.ResendErroredGasReclaims()

	treasureClaims := []models.WebnodeTreasureClaim{}
	treasureClaims = getAllWebnodeTreasureClaims(suite)

	suite.Equal(4, len(treasureClaims))

	for _, treasureClaim := range treasureClaims {
		/* verify that the claim's gas status was not changed */
		suite.Equal(models.GasTransferLeftoversReclaimProcessing,
			treasureClaim.GasStatus)
	}
}

func (suite *JobsSuite) Test_SendGasForNewTreasureClaims() {
	resetTestVariables_claimTreasureForWebnode(suite)

	jobs.EthWrapper.CalculateGasNeeded = func(desiredGasLimit uint64) (*big.Int, error) {
		/* track number of times this is called */
		calculateGasNeededCalls_claimTreasureForWebnode++
		hasCalledCalculateGas_claimTreasureForWebnode = true
		return big.NewInt(1), errors.New("making this fail")
	}

	generateWebnodeTreasureClaims(
		suite,
		models.PRLClaimNotStarted,
		models.GasTransferNotStarted,
		1,
		4)

	jobs.SendGasForNewTreasureClaims()

	// if this got called it means "SendGas" got called
	suite.True(hasCalledCalculateGas_claimTreasureForWebnode)
	suite.Equal(1, calculateGasNeededCalls_claimTreasureForWebnode)
}

func (suite *JobsSuite) Test_StartNewTreasureClaims() {
	resetTestVariables_claimTreasureForWebnode(suite)

	jobs.EthWrapper.CheckPRLBalance = func(address common.Address) *big.Int {
		hasCalledCheckPRLBalance_claimTreasureForWebnode = true
		chechPRLBalanceCalls_claimTreasureForWebnode++
		/* return a 0 balance so execution will exit jobs.ClaimPRL early */
		return big.NewInt(0)
	}

	generateWebnodeTreasureClaims(
		suite,
		models.PRLClaimNotStarted,
		models.GasTransferSuccess,
		1,
		4)

	jobs.StartNewTreasureClaims()

	/* if this got called it means "ClaimPRL" got called */
	suite.True(hasCalledCheckPRLBalance_claimTreasureForWebnode)
	suite.Equal(4, chechPRLBalanceCalls_claimTreasureForWebnode)
}

func (suite *JobsSuite) Test_RetrieveLeftoverETHFromTreasureClaiming_worth_reclaiming() {
	resetTestVariables_claimTreasureForWebnode(suite)

	jobs.EthWrapper.SendETH = func(fromAddress common.Address, fromPrivKey *ecdsa.PrivateKey, toAddress common.Address,
		gas *big.Int) (types.Transactions, string, int64, error) {
		hasCalledSendETH_claimTreasureForWebnode = true
		return types.Transactions{}, "111111", 1, nil
	}
	jobs.EthWrapper.CheckIfWorthReclaimingGas = func(address common.Address, desiredGasLimit uint64) (bool, *big.Int, error) {
		hasCalledCheckIfWorthReclaimingGas_claimTreasureForWebnode = true
		return true, big.NewInt(5555), nil
	}

	generateWebnodeTreasureClaims(
		suite,
		models.PRLClaimSuccess,
		models.GasTransferSuccess,
		1,
		4)

	jobs.RetrieveLeftoverETHFromTreasureClaiming()

	suite.True(hasCalledCheckIfWorthReclaimingGas_claimTreasureForWebnode)
	suite.True(hasCalledSendETH_claimTreasureForWebnode)
}

func (suite *JobsSuite) Test_RetrieveLeftoverETHFromTreasureClaiming_not_worth_reclaiming() {
	resetTestVariables_claimTreasureForWebnode(suite)

	jobs.EthWrapper.CheckIfWorthReclaimingGas = func(address common.Address, desiredGasLimit uint64) (bool, *big.Int, error) {
		hasCalledCheckIfWorthReclaimingGas_claimTreasureForWebnode = true
		return false, big.NewInt(0), nil
	}

	generateWebnodeTreasureClaims(
		suite,
		models.PRLClaimSuccess,
		models.GasTransferSuccess,
		1,
		4)

	jobs.RetrieveLeftoverETHFromTreasureClaiming()

	suite.True(hasCalledCheckIfWorthReclaimingGas_claimTreasureForWebnode)
	suite.False(hasCalledSendETH_claimTreasureForWebnode)

	treasureClaims := getAllWebnodeTreasureClaims(suite)

	suite.Equal(4, len(treasureClaims))

	for _, treasureClaim := range treasureClaims {
		/* verify that we set the reclaim status to success */
		suite.Equal(models.GasTransferLeftoversReclaimSuccess,
			treasureClaim.GasStatus)
	}
}

func (suite *JobsSuite) Test_RetrieveLeftoverETHFromTreasureClaiming_error() {
	resetTestVariables_claimTreasureForWebnode(suite)

	jobs.EthWrapper.CheckIfWorthReclaimingGas = func(address common.Address, desiredGasLimit uint64) (bool, *big.Int, error) {
		hasCalledCheckIfWorthReclaimingGas_claimTreasureForWebnode = true
		return false, big.NewInt(0), errors.New("some error response")
	}

	generateWebnodeTreasureClaims(
		suite,
		models.PRLClaimSuccess,
		models.GasTransferSuccess,
		1,
		4)

	jobs.RetrieveLeftoverETHFromTreasureClaiming()

	suite.True(hasCalledCheckIfWorthReclaimingGas_claimTreasureForWebnode)
	suite.False(hasCalledSendETH_claimTreasureForWebnode)

	treasureClaims := getAllWebnodeTreasureClaims(suite)

	suite.Equal(4, len(treasureClaims))

	for _, treasureClaim := range treasureClaims {
		/* verify that the status has not changed */
		suite.Equal(models.GasTransferSuccess,
			treasureClaim.GasStatus)
	}
}

func (suite *JobsSuite) Test_SendGas_already_has_enough_gas() {
	resetTestVariables_claimTreasureForWebnode(suite)

	jobs.EthWrapper.CheckETHBalance = func(addr common.Address) *big.Int {
		hasCalledCheckETHBalance_claimTreasureForWebnode = true
		/* return a large balance so it will think it already has enough gas */
		return big.NewInt(9999999)
	}
	jobs.EthWrapper.CalculateGasNeeded = func(desiredGasLimit uint64) (*big.Int, error) {
		/* return a small calculation so it will think it does not need much gas */
		hasCalledCalculateGas_claimTreasureForWebnode = true
		return big.NewInt(1), nil
	}

	generateWebnodeTreasureClaims(
		suite,
		models.PRLClaimNotStarted,
		models.GasTransferNotStarted,
		1,
		4)

	treasureClaims := getAllWebnodeTreasureClaims(suite)
	suite.Equal(4, len(treasureClaims))

	jobs.SendGas(treasureClaims)

	suite.True(hasCalledCheckETHBalance_claimTreasureForWebnode)
	suite.True(hasCalledCheckETHBalance_claimTreasureForWebnode)
	suite.False(hasCalledSendETH_claimTreasureForWebnode)

	treasureClaims = getAllWebnodeTreasureClaims(suite)
	suite.Equal(4, len(treasureClaims))

	for _, treasureClaim := range treasureClaims {
		/* verify that we have changed the status to success */
		suite.Equal(models.GasTransferSuccess,
			treasureClaim.GasStatus)
	}
}

func (suite *JobsSuite) Test_SendGas_needs_more_gas() {
	resetTestVariables_claimTreasureForWebnode(suite)

	jobs.EthWrapper.SendETH = func(fromAddress common.Address, fromPrivKey *ecdsa.PrivateKey, toAddress common.Address,
		gas *big.Int) (types.Transactions, string, int64, error) {
		hasCalledSendETH_claimTreasureForWebnode = true
		return types.Transactions{}, "111111", 1, nil
	}
	jobs.EthWrapper.CheckETHBalance = func(addr common.Address) *big.Int {
		hasCalledCheckETHBalance_claimTreasureForWebnode = true
		/* return a small balance so it will think it needs more gas */
		return big.NewInt(1)
	}
	jobs.EthWrapper.CalculateGasNeeded = func(desiredGasLimit uint64) (*big.Int, error) {
		/* return a large gas needed calculation so it will think it needs more gas */
		hasCalledCalculateGas_claimTreasureForWebnode = true
		return big.NewInt(999999), nil
	}

	generateWebnodeTreasureClaims(
		suite,
		models.PRLClaimNotStarted,
		models.GasTransferNotStarted,
		1,
		4)

	treasureClaims := getAllWebnodeTreasureClaims(suite)
	suite.Equal(4, len(treasureClaims))

	jobs.SendGas(treasureClaims)

	suite.True(hasCalledCheckETHBalance_claimTreasureForWebnode)
	suite.True(hasCalledCheckETHBalance_claimTreasureForWebnode)
	suite.True(hasCalledSendETH_claimTreasureForWebnode)

	treasureClaims = getAllWebnodeTreasureClaims(suite)
	suite.Equal(4, len(treasureClaims))

	for _, treasureClaim := range treasureClaims {
		/* verify that we have changed the status to processing */
		suite.Equal(models.GasTransferProcessing,
			treasureClaim.GasStatus)
	}
}

func (suite *JobsSuite) SetClaimClockIfUnset() {
	resetTestVariables_claimTreasureForWebnode(suite)

	newClaimClockValue := big.NewInt(9999999)

	jobs.EthWrapper.CheckClaimClock = func(address common.Address) (*big.Int, error) {
		hasCalledCheckClaimClock_claimTreasureForWebnode = true
		return newClaimClockValue, nil
	}

	generateWebnodeTreasureClaims(
		suite,
		models.PRLClaimNotStarted,
		models.GasTransferNotStarted,
		models.UnsetClaimClockValue.Int64(),
		1)

	treasureClaims := getAllWebnodeTreasureClaims(suite)

	err := jobs.SetClaimClockIfUnset(&treasureClaims[0])
	suite.Nil(err)

	suite.Equal(newClaimClockValue.Int64(),
		treasureClaims[0].StartingClaimClock)

	suite.True(hasCalledCheckClaimClock_claimTreasureForWebnode)
}

func (suite *JobsSuite) ClaimPRL_success() {
	resetTestVariables_claimTreasureForWebnode(suite)

	jobs.EthWrapper.CheckPRLBalance = func(address common.Address) *big.Int {
		hasCalledCheckPRLBalance_claimTreasureForWebnode = true
		/* return a balance so it will try to claim the PRL */
		return big.NewInt(5)
	}
	jobs.EthWrapper.ClaimPRL = func(receiverAddress common.Address,
		treasureAddress common.Address, privateKey *ecdsa.PrivateKey) bool {
		hasCalledClaimPRL_claimTreasureForWebnode = true
		/* return true for success */
		return true
	}

	generateWebnodeTreasureClaims(
		suite,
		models.PRLClaimNotStarted,
		models.GasTransferSuccess,
		1,
		1)

	treasureClaims := getAllWebnodeTreasureClaims(suite)

	jobs.ClaimPRL(treasureClaims)

	suite.True(hasCalledCheckPRLBalance_claimTreasureForWebnode)
	suite.True(hasCalledClaimPRL_claimTreasureForWebnode)

	treasureClaims = getAllWebnodeTreasureClaims(suite)
	suite.Equal(1, len(treasureClaims))

	for _, treasureClaim := range treasureClaims {
		suite.Equal(models.PRLClaimProcessing, treasureClaim.ClaimPRLStatus)
	}
}

func (suite *JobsSuite) ClaimPRL_failure() {
	resetTestVariables_claimTreasureForWebnode(suite)

	jobs.EthWrapper.CheckPRLBalance = func(address common.Address) *big.Int {
		hasCalledCheckPRLBalance_claimTreasureForWebnode = true
		/* return a balance so it will try to claim the PRL */
		return big.NewInt(5)
	}
	jobs.EthWrapper.ClaimPRL = func(receiverAddress common.Address,
		treasureAddress common.Address, privateKey *ecdsa.PrivateKey) bool {
		hasCalledClaimPRL_claimTreasureForWebnode = true
		/* return true for failure */
		return false
	}

	generateWebnodeTreasureClaims(
		suite,
		models.PRLClaimNotStarted,
		models.GasTransferSuccess,
		1,
		1)

	treasureClaims := getAllWebnodeTreasureClaims(suite)

	jobs.ClaimPRL(treasureClaims)

	suite.True(hasCalledCheckPRLBalance_claimTreasureForWebnode)
	suite.True(hasCalledClaimPRL_claimTreasureForWebnode)

	treasureClaims = getAllWebnodeTreasureClaims(suite)
	suite.Equal(1, len(treasureClaims))

	for _, treasureClaim := range treasureClaims {
		suite.Equal(models.PRLClaimNotStarted, treasureClaim.ClaimPRLStatus)
	}
}

func (suite *JobsSuite) ClaimPRL_no_prl() {
	resetTestVariables_claimTreasureForWebnode(suite)

	jobs.EthWrapper.CheckPRLBalance = func(address common.Address) *big.Int {
		hasCalledCheckPRLBalance_claimTreasureForWebnode = true
		/* return a 0 balance so it will not try to claim the PRL */
		return big.NewInt(0)
	}
	jobs.EthWrapper.ClaimPRL = func(receiverAddress common.Address,
		treasureAddress common.Address, privateKey *ecdsa.PrivateKey) bool {
		hasCalledClaimPRL_claimTreasureForWebnode = true
		return false
	}

	generateWebnodeTreasureClaims(
		suite,
		models.PRLClaimNotStarted,
		models.GasTransferSuccess,
		1,
		1)

	treasureClaims := getAllWebnodeTreasureClaims(suite)

	jobs.ClaimPRL(treasureClaims)

	suite.True(hasCalledCheckPRLBalance_claimTreasureForWebnode)
	suite.False(hasCalledClaimPRL_claimTreasureForWebnode)

	treasureClaims = getAllWebnodeTreasureClaims(suite)
	suite.Equal(1, len(treasureClaims))

	for _, treasureClaim := range treasureClaims {
		suite.Equal(models.PRLClaimSuccess, treasureClaim.ClaimPRLStatus)
	}
}

func (suite *JobsSuite) PurgeCompletedTreasureClaims() {
	resetTestVariables_claimTreasureForWebnode(suite)

	generateWebnodeTreasureClaims(
		suite,
		models.PRLClaimSuccess,
		models.GasTransferLeftoversReclaimSuccess,
		1,
		4)

	treasureClaims := getAllWebnodeTreasureClaims(suite)
	suite.Equal(4, len(treasureClaims))

	jobs.PurgeCompletedTreasureClaims()

	treasureClaims = getAllWebnodeTreasureClaims(suite)
	suite.Equal(0, len(treasureClaims))
}

func generateWebnodeTreasureClaims(suite *JobsSuite,
	claimPRLStatus models.PRLClaimStatus,
	gasTransferStatus models.GasTransferStatus,
	startingClaimClock int64,
	numToGenerate int) {

	for i := 0; i < numToGenerate; i++ {
		treasureAddr, key, _ := services.EthWrapper.GenerateEthAddr()
		receiverAddr, _, _ := services.EthWrapper.GenerateEthAddr()

		validChars := []rune("abcde123456789")
		genesisHash := oyster_utils.RandSeq(64, validChars)

		webnodeTreasureClaim := models.WebnodeTreasureClaim{
			GenesisHash:           genesisHash,
			TreasureETHAddr:       treasureAddr.Hex(),
			TreasureETHPrivateKey: key,
			ReceiverETHAddr:       receiverAddr.Hex(),
			SectorIdx:             0,
			NumChunks:             100,
			ClaimPRLStatus:        claimPRLStatus,
			GasStatus:             gasTransferStatus,
			StartingClaimClock:    startingClaimClock,
		}

		vErr, err := suite.DB.ValidateAndCreate(&webnodeTreasureClaim)

		suite.False(vErr.HasAny())
		suite.Nil(err)
	}
}

func getAllWebnodeTreasureClaims(suite *JobsSuite) []models.WebnodeTreasureClaim {
	treasureClaims := []models.WebnodeTreasureClaim{}
	err := suite.DB.RawQuery("SELECT * FROM " +
		"webnode_treasure_claims").All(&treasureClaims)
	suite.Nil(err)
	return treasureClaims
}

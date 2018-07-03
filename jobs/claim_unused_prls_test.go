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
	hasCalledCheckPRLBalance   = false
	hasCalledCheckETHBalance   = false
	hasCalledSendPRLFromOyster = false
	hasCalledSendETH           = false
	SentToSendPRLFromOyster    = 0
	SentToSendETH              = 0
)

var (
	RowWithGasTransferNotStarted                 = models.CompletedUpload{}
	RowWithGasTransferProcessing                 = models.CompletedUpload{}
	RowWithGasTransferSuccess                    = models.CompletedUpload{}
	RowWithGasTransferError                      = models.CompletedUpload{}
	RowWithPRLClaimProcessing                    = models.CompletedUpload{}
	RowWithPRLClaimSuccess                       = models.CompletedUpload{}
	RowWithPRLClaimError                         = models.CompletedUpload{}
	RowWithGasTransferLeftoversReclaimProcessing = models.CompletedUpload{}
	RowWithGasTransferLeftoversReclaimSuccess    = models.CompletedUpload{}
)

func (suite *JobsSuite) Test_ClaimUnusedPRLs() {

	testResendTimedOutGasTransfers(suite)
	testResendTimedOutPRLTransfers(suite)
	testResendErroredGasTransfers(suite)
	testResendErroredPRLTransfers(suite)
	testSendGasForNewClaims(suite)
	testStartNewClaims(suite)
	testRetrieveLeftoverETH(suite)
	testInitiateGasTransfer(suite)
	testInitiatePRLClaim(suite)
	testPurgeCompletedClaims(suite)

	testCheckProcessingGasTransactions_transaction_succeeded(suite)
	testCheckProcessingGasTransactions_still_pending(suite)
	testCheckProcessingPRLTransactions_transaction_succeeded(suite)
	testCheckProcessingPRLTransactions_still_pending(suite)
	testCheckProcessingGasReclaims_reclaim_complete(suite)
	testCheckProcessingGasReclaims_not_enough_to_reclaim(suite)
	testCheckProcessingGasReclaims_still_pending(suite)

	testResendTimedOutGasReclaims(suite)
}

func testResendTimedOutGasTransfers(suite *JobsSuite) {
	testSetup(suite)

	suite.Equal(0, SentToSendETH)
	suite.Equal(false, hasCalledSendETH)

	jobs.ResendTimedOutGasTransfers(time.Now().Add(30 * time.Minute))
	// should be one timed out
	suite.Equal(1, SentToSendETH)
	suite.Equal(true, hasCalledSendETH)

	resetTestVariables()

	jobs.ResendTimedOutGasTransfers(time.Now().Add(-20 * time.Minute))
	// should be none, nothing timed out yet
	suite.Equal(0, SentToSendETH)
	suite.Equal(false, hasCalledSendETH)
}

func testResendTimedOutPRLTransfers(suite *JobsSuite) {
	testSetup(suite)

	suite.Equal(0, SentToSendPRLFromOyster)
	suite.Equal(false, hasCalledSendPRLFromOyster)

	jobs.ResendTimedOutPRLTransfers(time.Now().Add(30 * time.Minute))
	// should be one timed out
	suite.Equal(1, SentToSendPRLFromOyster)
	suite.Equal(true, hasCalledSendPRLFromOyster)

	resetTestVariables()

	jobs.ResendTimedOutPRLTransfers(time.Now().Add(-20 * time.Minute))
	// should be none, nothing timed out yet
	suite.Equal(0, SentToSendPRLFromOyster)
	suite.Equal(false, hasCalledSendPRLFromOyster)
}

func testResendErroredGasTransfers(suite *JobsSuite) {
	testSetup(suite)

	suite.Equal(0, SentToSendETH)
	suite.Equal(false, hasCalledSendETH)

	jobs.ResendErroredGasTransfers()

	// should be one error'd gas transfer that gets resent
	suite.Equal(1, SentToSendETH)
	suite.Equal(true, hasCalledSendETH)
}

func testResendErroredPRLTransfers(suite *JobsSuite) {
	testSetup(suite)

	suite.Equal(0, SentToSendPRLFromOyster)
	suite.Equal(false, hasCalledSendPRLFromOyster)

	jobs.ResendErroredPRLTransfers()

	// should be one error'd prl transfer that gets resent
	suite.Equal(1, SentToSendPRLFromOyster)
	suite.Equal(true, hasCalledSendPRLFromOyster)
}

func testSendGasForNewClaims(suite *JobsSuite) {
	testSetup(suite)

	suite.Equal(0, SentToSendETH)
	suite.Equal(false, hasCalledSendETH)

	jobs.SendGasForNewClaims()

	// should be one new gas transfer that gets sent
	suite.Equal(1, SentToSendETH)
	suite.Equal(true, hasCalledSendETH)
}

func testStartNewClaims(suite *JobsSuite) {
	testSetup(suite)

	suite.Equal(0, SentToSendPRLFromOyster)
	suite.Equal(false, hasCalledSendPRLFromOyster)

	jobs.StartNewClaims()

	// should be one new prl claim that gets sent
	suite.Equal(1, SentToSendPRLFromOyster)
	suite.Equal(true, hasCalledSendPRLFromOyster)
}

func testRetrieveLeftoverETH(suite *JobsSuite) {
	testSetup(suite)

	suite.Equal(0, SentToSendETH)
	suite.Equal(false, hasCalledSendETH)

	completedUploads := []models.CompletedUpload{}
	err := suite.DB.All(&completedUploads)
	suite.Nil(err)
	suite.NotEqual(0, len(completedUploads))

	jobs.RetrieveLeftoverETH()

	// should be 1 call to retrieve leftover gas
	suite.Equal(1, SentToSendETH)
	suite.Equal(true, hasCalledSendETH)
	suite.Equal(true, hasCalledCheckETHBalance)
}

func testInitiateGasTransfer(suite *JobsSuite) {
	testSetup(suite)

	suite.Equal(0, SentToSendETH)
	suite.Equal(false, hasCalledSendETH)

	completedUploads := []models.CompletedUpload{}
	err := suite.DB.RawQuery("SELECT * from completed_uploads WHERE gas_status != ?",
		models.GasTransferLeftoversReclaimSuccess).All(&completedUploads)
	suite.Nil(err)
	suite.NotEqual(0, len(completedUploads))

	jobs.InitiateGasTransfer(completedUploads)

	suite.Equal(len(completedUploads), SentToSendETH)
	suite.Equal(true, hasCalledSendETH)
}

func testInitiatePRLClaim(suite *JobsSuite) {
	testSetup(suite)

	suite.Equal(0, SentToSendPRLFromOyster)
	suite.Equal(false, hasCalledSendPRLFromOyster)

	completedUploads := []models.CompletedUpload{}
	err := suite.DB.All(&completedUploads)
	suite.Nil(err)
	suite.NotEqual(0, len(completedUploads))

	jobs.InitiatePRLClaim(completedUploads)

	suite.Equal(len(completedUploads), SentToSendPRLFromOyster)
	suite.Equal(true, hasCalledSendPRLFromOyster)
	suite.Equal(true, hasCalledCheckPRLBalance)
}

func testPurgeCompletedClaims(suite *JobsSuite) {
	testSetup(suite)

	completedUploadsStarting := []models.CompletedUpload{}
	err := suite.DB.All(&completedUploadsStarting)
	suite.Nil(err)
	suite.NotEqual(0, len(completedUploadsStarting))

	jobs.PurgeCompletedClaims()

	completedUploadsCurrent := []models.CompletedUpload{}
	err = suite.DB.All(&completedUploadsCurrent)
	suite.Nil(err)
	suite.NotEqual(0, len(completedUploadsCurrent))

	suite.Equal(len(completedUploadsStarting)-1, len(completedUploadsCurrent))
}

func testCheckProcessingGasTransactions_transaction_succeeded(suite *JobsSuite) {
	testSetup(suite)

	gasTransfersProcessing := []models.CompletedUpload{}

	err := suite.DB.RawQuery("SELECT * from "+
		"completed_uploads WHERE gas_status = ?", models.GasTransferProcessing).All(&gasTransfersProcessing)
	suite.Nil(err)
	suite.Equal(1, len(gasTransfersProcessing))

	// call method under test
	// should no longer be any transfers processing
	jobs.CheckProcessingGasTransactions()

	gasTransfersProcessing = []models.CompletedUpload{}

	err = suite.DB.RawQuery("SELECT * from "+
		"completed_uploads WHERE gas_status = ?", models.GasTransferProcessing).All(&gasTransfersProcessing)
	suite.Nil(err)
	suite.Equal(0, len(gasTransfersProcessing))
	suite.Equal(true, hasCalledCheckETHBalance)
}

func testCheckProcessingGasTransactions_still_pending(suite *JobsSuite) {
	testSetup(suite)

	jobs.EthWrapper.CheckETHBalance = func(addr common.Address) *big.Int {
		hasCalledCheckETHBalance = true
		// return a 0 balance so the transaction will remain processing
		return big.NewInt(0)
	}

	gasTransfersProcessing := []models.CompletedUpload{}

	err := suite.DB.RawQuery("SELECT * from "+
		"completed_uploads WHERE gas_status = ?", models.GasTransferProcessing).All(&gasTransfersProcessing)
	suite.Nil(err)
	suite.Equal(1, len(gasTransfersProcessing))

	// call method under test
	// should still be a transfer processing
	jobs.CheckProcessingGasTransactions()

	gasTransfersProcessing = []models.CompletedUpload{}

	err = suite.DB.RawQuery("SELECT * from "+
		"completed_uploads WHERE gas_status = ?", models.GasTransferProcessing).All(&gasTransfersProcessing)
	suite.Nil(err)
	suite.Equal(1, len(gasTransfersProcessing))
	suite.Equal(true, hasCalledCheckETHBalance)
}

func testCheckProcessingPRLTransactions_transaction_succeeded(suite *JobsSuite) {
	testSetup(suite)

	jobs.EthWrapper.CheckPRLBalance = func(addr common.Address) *big.Int {
		hasCalledCheckPRLBalance = true
		// return a 0 balance so the transaction will be seen as complete
		return big.NewInt(0)
	}

	transfersProcessing := []models.CompletedUpload{}

	err := suite.DB.RawQuery("SELECT * from "+
		"completed_uploads WHERE prl_status = ?", models.PRLClaimProcessing).All(&transfersProcessing)
	suite.Nil(err)
	suite.Equal(1, len(transfersProcessing))

	// call method under test
	// should no longer be any transfers processing
	jobs.CheckProcessingPRLTransactions()

	transfersProcessing = []models.CompletedUpload{}

	err = suite.DB.RawQuery("SELECT * from "+
		"completed_uploads WHERE prl_status = ?", models.PRLClaimProcessing).All(&transfersProcessing)
	suite.Nil(err)
	suite.Equal(0, len(transfersProcessing))
	suite.Equal(true, hasCalledCheckPRLBalance)
}

func testCheckProcessingPRLTransactions_still_pending(suite *JobsSuite) {
	testSetup(suite)

	transfersProcessing := []models.CompletedUpload{}

	err := suite.DB.RawQuery("SELECT * from "+
		"completed_uploads WHERE prl_status = ?", models.PRLClaimProcessing).All(&transfersProcessing)
	suite.Nil(err)
	suite.Equal(1, len(transfersProcessing))

	// call method under test
	// should still be a transfer processing
	jobs.CheckProcessingPRLTransactions()

	transfersProcessing = []models.CompletedUpload{}

	err = suite.DB.RawQuery("SELECT * from "+
		"completed_uploads WHERE prl_status = ?", models.PRLClaimProcessing).All(&transfersProcessing)
	suite.Nil(err)
	suite.Equal(1, len(transfersProcessing))
	suite.Equal(true, hasCalledCheckPRLBalance)
}

func testCheckProcessingGasReclaims_reclaim_complete(suite *JobsSuite) {
	testSetup(suite)

	jobs.EthWrapper.CheckETHBalance = func(addr common.Address) *big.Int {
		hasCalledCheckETHBalance = true
		// return a 0 balance so the transaction will be seen as complete
		return big.NewInt(0)
	}

	transfersProcessing := []models.CompletedUpload{}

	err := suite.DB.RawQuery("SELECT * from "+
		"completed_uploads WHERE gas_status = ?",
		models.GasTransferLeftoversReclaimProcessing).All(&transfersProcessing)
	suite.Nil(err)
	suite.Equal(1, len(transfersProcessing))

	// call method under test
	// should set the transfer to success
	jobs.CheckProcessingGasReclaims()

	transfersProcessing = []models.CompletedUpload{}

	err = suite.DB.RawQuery("SELECT * from "+
		"completed_uploads WHERE gas_status = ?", models.GasTransferLeftoversReclaimProcessing).All(&transfersProcessing)
	suite.Nil(err)
	suite.Equal(0, len(transfersProcessing))
	suite.Equal(true, hasCalledCheckETHBalance)
}

func testCheckProcessingGasReclaims_not_enough_to_reclaim(suite *JobsSuite) {
	testSetup(suite)

	jobs.EthWrapper.CheckETHBalance = func(addr common.Address) *big.Int {
		hasCalledCheckETHBalance = true
		// return a very tiny balance so the transfer will get set to
		// complete since there isn't enough to reclaim
		return big.NewInt(1)
	}

	transfersProcessing := []models.CompletedUpload{}

	err := suite.DB.RawQuery("SELECT * from "+
		"completed_uploads WHERE gas_status = ?",
		models.GasTransferLeftoversReclaimProcessing).All(&transfersProcessing)
	suite.Nil(err)
	suite.Equal(1, len(transfersProcessing))

	// call method under test
	// should set the transfer to success
	// because there is not enough ETH to
	// justify reclaiming it
	jobs.CheckProcessingGasReclaims()

	transfersProcessing = []models.CompletedUpload{}

	err = suite.DB.RawQuery("SELECT * from "+
		"completed_uploads WHERE gas_status = ?", models.GasTransferLeftoversReclaimProcessing).All(&transfersProcessing)
	suite.Nil(err)
	suite.Equal(0, len(transfersProcessing))
	suite.Equal(true, hasCalledCheckETHBalance)
}

func testCheckProcessingGasReclaims_still_pending(suite *JobsSuite) {
	testSetup(suite)

	jobs.EthWrapper.CheckETHBalance = func(addr common.Address) *big.Int {
		hasCalledCheckETHBalance = true
		ethRemaining := oyster_utils.ConvertGweiToWei(big.NewInt(42000))

		// return a balance that is worth trying to reclaim
		return ethRemaining
	}
	jobs.EthWrapper.CalculateGasToSend = func(desiredGasLimit uint64) (*big.Int, error) {
		gasPrice := oyster_utils.ConvertGweiToWei(big.NewInt(1))
		gasToSend := new(big.Int).Mul(gasPrice, big.NewInt(int64(desiredGasLimit)))
		return gasToSend, nil
	}

	transfersProcessing := []models.CompletedUpload{}

	err := suite.DB.RawQuery("SELECT * from "+
		"completed_uploads WHERE gas_status = ?",
		models.GasTransferLeftoversReclaimProcessing).All(&transfersProcessing)
	suite.Nil(err)
	suite.Equal(1, len(transfersProcessing))

	// call method under test
	// should not do anything to the transaction because it will still see it as pending
	jobs.CheckProcessingGasReclaims()

	transfersProcessing = []models.CompletedUpload{}

	err = suite.DB.RawQuery("SELECT * from "+
		"completed_uploads WHERE gas_status = ?", models.GasTransferLeftoversReclaimProcessing).All(&transfersProcessing)
	suite.Nil(err)
	suite.Equal(1, len(transfersProcessing))
	suite.Equal(true, hasCalledCheckETHBalance)
}

func testResendTimedOutGasReclaims(suite *JobsSuite) {
	testSetup(suite)

	transfersProcessing := []models.CompletedUpload{}

	err := suite.DB.RawQuery("SELECT * from "+
		"completed_uploads WHERE gas_status = ?",
		models.GasTransferLeftoversReclaimProcessing).All(&transfersProcessing)
	suite.Nil(err)
	suite.Equal(1, len(transfersProcessing))

	// call method under test
	// should set the transfer back to the previous state
	// so it will be attempted again
	jobs.ResendTimedOutGasReclaims(time.Now().Add(5 * time.Minute))

	transfersProcessing = []models.CompletedUpload{}

	err = suite.DB.RawQuery("SELECT * from "+
		"completed_uploads WHERE gas_status = ?",
		models.GasTransferLeftoversReclaimProcessing).All(&transfersProcessing)
	suite.Nil(err)
	suite.Equal(0, len(transfersProcessing))
}

func testSetup(suite *JobsSuite) {

	RowWithGasTransferNotStarted = models.CompletedUpload{}
	RowWithGasTransferProcessing = models.CompletedUpload{}
	RowWithGasTransferSuccess = models.CompletedUpload{}
	RowWithGasTransferError = models.CompletedUpload{}
	RowWithPRLClaimProcessing = models.CompletedUpload{}
	RowWithPRLClaimSuccess = models.CompletedUpload{}
	RowWithPRLClaimError = models.CompletedUpload{}
	RowWithGasTransferLeftoversReclaimProcessing = models.CompletedUpload{}
	RowWithGasTransferLeftoversReclaimSuccess = models.CompletedUpload{}

	addr, key, _ := jobs.EthWrapper.GenerateEthAddr()
	RowWithGasTransferNotStarted = models.CompletedUpload{
		GenesisHash:   "RowWithGasTransferNotStarted",
		ETHAddr:       addr.Hex(),
		ETHPrivateKey: key,
		PRLStatus:     models.PRLClaimNotStarted,
		GasStatus:     models.GasTransferNotStarted,
	}
	addr, key, _ = jobs.EthWrapper.GenerateEthAddr()
	RowWithGasTransferProcessing = models.CompletedUpload{
		GenesisHash:   "RowWithGasTransferProcessing",
		ETHAddr:       addr.Hex(),
		ETHPrivateKey: key,
		PRLStatus:     models.PRLClaimNotStarted,
		GasStatus:     models.GasTransferProcessing,
	}
	addr, key, _ = jobs.EthWrapper.GenerateEthAddr()
	RowWithGasTransferSuccess = models.CompletedUpload{
		GenesisHash:   "RowWithGasTransferSuccess",
		ETHAddr:       addr.Hex(),
		ETHPrivateKey: key,
		PRLStatus:     models.PRLClaimNotStarted,
		GasStatus:     models.GasTransferSuccess,
	}
	addr, key, _ = jobs.EthWrapper.GenerateEthAddr()
	RowWithGasTransferError = models.CompletedUpload{
		GenesisHash:   "RowWithGasTransferError",
		ETHAddr:       addr.Hex(),
		ETHPrivateKey: key,
		PRLStatus:     models.PRLClaimNotStarted,
		GasStatus:     models.GasTransferError,
	}
	addr, key, _ = jobs.EthWrapper.GenerateEthAddr()
	RowWithPRLClaimProcessing = models.CompletedUpload{
		GenesisHash:   "RowWithPRLClaimProcessing",
		ETHAddr:       addr.Hex(),
		ETHPrivateKey: key,
		PRLStatus:     models.PRLClaimProcessing,
		GasStatus:     models.GasTransferSuccess,
	}
	addr, key, _ = jobs.EthWrapper.GenerateEthAddr()
	RowWithPRLClaimSuccess = models.CompletedUpload{
		GenesisHash:   "RowWithPRLClaimSuccess",
		ETHAddr:       addr.Hex(),
		ETHPrivateKey: key,
		PRLStatus:     models.PRLClaimSuccess,
		GasStatus:     models.GasTransferSuccess,
	}
	addr, key, _ = jobs.EthWrapper.GenerateEthAddr()
	RowWithPRLClaimError = models.CompletedUpload{
		GenesisHash:   "RowWithPRLClaimError",
		ETHAddr:       addr.Hex(),
		ETHPrivateKey: key,
		PRLStatus:     models.PRLClaimError,
		GasStatus:     models.GasTransferSuccess,
	}

	addr, key, _ = jobs.EthWrapper.GenerateEthAddr()
	RowWithGasTransferLeftoversReclaimProcessing = models.CompletedUpload{
		GenesisHash:   "RowWithGasTransferLeftoversReclaimProcessing",
		ETHAddr:       addr.Hex(),
		ETHPrivateKey: key,
		PRLStatus:     models.PRLClaimSuccess,
		GasStatus:     models.GasTransferLeftoversReclaimProcessing,
	}

	addr, key, _ = jobs.EthWrapper.GenerateEthAddr()
	RowWithGasTransferLeftoversReclaimSuccess = models.CompletedUpload{
		GenesisHash:   "RowWithGasTransferLeftoversReclaimSuccess",
		ETHAddr:       addr.Hex(),
		ETHPrivateKey: key,
		PRLStatus:     models.PRLClaimSuccess,
		GasStatus:     models.GasTransferLeftoversReclaimSuccess,
	}

	err := suite.DB.RawQuery("DELETE from completed_uploads").All(&[]models.CompletedUpload{})
	suite.Nil(err)

	_, err = suite.DB.ValidateAndSave(&RowWithGasTransferNotStarted)
	RowWithGasTransferNotStarted.EncryptSessionEthKey()
	suite.Nil(err)

	_, err = suite.DB.ValidateAndSave(&RowWithGasTransferProcessing)
	RowWithGasTransferProcessing.EncryptSessionEthKey()
	suite.Nil(err)

	_, err = suite.DB.ValidateAndSave(&RowWithGasTransferSuccess)
	RowWithGasTransferSuccess.EncryptSessionEthKey()
	suite.Nil(err)

	_, err = suite.DB.ValidateAndSave(&RowWithGasTransferError)
	RowWithGasTransferError.EncryptSessionEthKey()
	suite.Nil(err)

	_, err = suite.DB.ValidateAndSave(&RowWithPRLClaimProcessing)
	RowWithPRLClaimProcessing.EncryptSessionEthKey()
	suite.Nil(err)

	_, err = suite.DB.ValidateAndSave(&RowWithPRLClaimSuccess)
	RowWithPRLClaimSuccess.EncryptSessionEthKey()
	suite.Nil(err)

	_, err = suite.DB.ValidateAndSave(&RowWithPRLClaimError)
	RowWithPRLClaimError.EncryptSessionEthKey()
	suite.Nil(err)

	_, err = suite.DB.ValidateAndSave(&RowWithGasTransferLeftoversReclaimProcessing)
	RowWithGasTransferLeftoversReclaimProcessing.EncryptSessionEthKey()
	suite.Nil(err)

	_, err = suite.DB.ValidateAndSave(&RowWithGasTransferLeftoversReclaimSuccess)
	RowWithGasTransferLeftoversReclaimSuccess.EncryptSessionEthKey()
	suite.Nil(err)

	resetTestVariables()
}

func resetTestVariables() {
	SentToSendPRLFromOyster = 0

	SentToSendETH = 0

	hasCalledCheckPRLBalance = false
	hasCalledCheckETHBalance = false
	hasCalledSendPRLFromOyster = false
	hasCalledSendETH = false

	jobs.EthWrapper = services.Eth{
		GenerateEthAddr:      services.EthWrapper.GenerateEthAddr,
		CreateSendPRLMessage: services.EthWrapper.CreateSendPRLMessage,
		CheckPRLBalance: func(addr common.Address) *big.Int {
			hasCalledCheckPRLBalance = true
			return big.NewInt(600000000000000000)
		},
		CheckETHBalance: func(addr common.Address) *big.Int {
			hasCalledCheckETHBalance = true
			return big.NewInt(600000000000000000)
		},
		SendPRLFromOyster: func(msg services.OysterCallMsg) (bool, string, int64) {
			SentToSendPRLFromOyster++
			hasCalledSendPRLFromOyster = true
			return false, "some__transaction_hash", 0
		},
		CalculateGasToSend: func(desiredGasLimit uint64) (*big.Int, error) {
			gasPrice := oyster_utils.ConvertGweiToWei(big.NewInt(1))
			gasToSend := new(big.Int).Mul(gasPrice, big.NewInt(int64(desiredGasLimit)))
			return gasToSend, nil
		},
		SendETH: func(fromAddress common.Address, fromPrivKey *ecdsa.PrivateKey, toAddress common.Address,
			gas *big.Int) (types.Transactions, string, int64, error) {
			SentToSendETH++
			hasCalledSendETH = true
			// make one of the transfers unsuccessful
			return types.Transactions{}, "111111", 1, nil
		},
	}
}

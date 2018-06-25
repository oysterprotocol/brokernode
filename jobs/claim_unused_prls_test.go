package jobs_test

import (
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
	hasCalledSendPRLFromOyster = false
	hasCalledSendETH           = false
	SentToSendPRLFromOyster    = 0
	SentToSendETH              = 0
)

var (
	RowWithGasTransferNotStarted = models.CompletedUpload{}
	RowWithGasTransferProcessing = models.CompletedUpload{}
	RowWithGasTransferSuccess    = models.CompletedUpload{}
	RowWithGasTransferError      = models.CompletedUpload{}
	RowWithPRLClaimProcessing    = models.CompletedUpload{}
	RowWithPRLClaimSuccess       = models.CompletedUpload{}
	RowWithPRLClaimError         = models.CompletedUpload{}
)

func (suite *JobsSuite) Test_ClaimUnusedPRLs() {

	testSetup(suite)

	testResendTimedOutGasTransfers(suite)
	testResendTimedOutPRLTransfers(suite)
	testResendErroredGasTransfers(suite)
	testResendErroredPRLTransfers(suite)
	testSendGasForNewClaims(suite)
	testStartNewClaims(suite)
	testInitiateGasTransfer(suite)
	testInitiatePRLClaim(suite)
	testPurgeCompletedClaims(suite)
}

func testResendTimedOutGasTransfers(suite *JobsSuite) {
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
	defer resetTestVariables()

	suite.Equal(0, SentToSendETH)
	suite.Equal(false, hasCalledSendETH)

	jobs.ResendErroredGasTransfers()

	// should be one error'd gas transfer that gets resent
	suite.Equal(1, SentToSendETH)
	suite.Equal(true, hasCalledSendETH)
}

func testResendErroredPRLTransfers(suite *JobsSuite) {
	defer resetTestVariables()

	suite.Equal(0, SentToSendPRLFromOyster)
	suite.Equal(false, hasCalledSendPRLFromOyster)

	jobs.ResendErroredPRLTransfers()

	// should be one error'd prl transfer that gets resent
	suite.Equal(1, SentToSendPRLFromOyster)
	suite.Equal(true, hasCalledSendPRLFromOyster)
}

func testSendGasForNewClaims(suite *JobsSuite) {
	defer resetTestVariables()

	suite.Equal(0, SentToSendETH)
	suite.Equal(false, hasCalledSendETH)

	jobs.SendGasForNewClaims()

	// should be one new gas transfer that gets sent
	suite.Equal(1, SentToSendETH)
	suite.Equal(true, hasCalledSendETH)
}

func testStartNewClaims(suite *JobsSuite) {
	defer resetTestVariables()

	suite.Equal(0, SentToSendPRLFromOyster)
	suite.Equal(false, hasCalledSendPRLFromOyster)

	jobs.StartNewClaims()

	// should be one new prl claim that gets sent
	suite.Equal(1, SentToSendPRLFromOyster)
	suite.Equal(true, hasCalledSendPRLFromOyster)
}

func testInitiateGasTransfer(suite *JobsSuite) {
	defer resetTestVariables()

	suite.Equal(0, SentToSendETH)
	suite.Equal(false, hasCalledSendETH)

	completedUploads := []models.CompletedUpload{}
	err := suite.DB.All(&completedUploads)
	suite.Nil(err)
	suite.NotEqual(0, len(completedUploads))

	jobs.InitiateGasTransfer(completedUploads)

	suite.Equal(len(completedUploads), SentToSendETH)
	suite.Equal(true, hasCalledSendETH)
}

func testInitiatePRLClaim(suite *JobsSuite) {
	defer resetTestVariables()

	suite.Equal(0, SentToSendPRLFromOyster)
	suite.Equal(false, hasCalledSendPRLFromOyster)

	completedUploads := []models.CompletedUpload{}
	err := suite.DB.All(&completedUploads)
	suite.Nil(err)
	suite.NotEqual(0, len(completedUploads))

	jobs.InitiatePRLClaim(completedUploads)

	suite.Equal(len(completedUploads), SentToSendPRLFromOyster)
	suite.Equal(true, hasCalledSendPRLFromOyster)
}

func testPurgeCompletedClaims(suite *JobsSuite) {
	defer resetTestVariables()

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

func testSetup(suite *JobsSuite) {

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

	resetTestVariables()
}

func resetTestVariables() {
	SentToSendPRLFromOyster = 0

	SentToSendETH = 0

	hasCalledCheckPRLBalance = false
	hasCalledSendPRLFromOyster = false
	hasCalledSendETH = false

	jobs.EthWrapper = services.Eth{
		CreateSendPRLMessage: services.EthWrapper.CreateSendPRLMessage,
		CheckPRLBalance: func(addr common.Address) *big.Int {
			hasCalledCheckPRLBalance = true
			return big.NewInt(600000000000000000)
		},
		SendPRLFromOyster: func(msg services.OysterCallMsg) (bool, string, int64) {
			SentToSendPRLFromOyster++
			hasCalledSendPRLFromOyster = true
			// make one of the transfers unsuccessful
			return false, "some__transaction_hash", 0
		},
		CalculateGasToSend: func(desiredGasLimit uint64) (*big.Int, error) {
			gasPrice := oyster_utils.ConvertGweiToWei(big.NewInt(1))
			gasToSend := new(big.Int).Mul(gasPrice, big.NewInt(int64(desiredGasLimit)))
			return gasToSend, nil
		},
		SendETH: func(address common.Address, gas *big.Int) (types.Transactions, string, int64, error) {
			SentToSendETH++
			hasCalledSendETH = true
			// make one of the transfers unsuccessful
			return types.Transactions{}, "111111", 1, nil
		},
	}
}

package jobs_test

import (
	"encoding/hex"
	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"time"
)

var (
	sendGasMockCalled_claim_unusued_prls         = false
	claimUnusedPRLsMockCalled_claim_unusued_prls = false
	EthMock                                      = services.EthMock
	SentToClaimUnusedPRLs                        []models.CompletedUpload
	SentToSendGas                                []models.CompletedUpload
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

	defer services.SetUpMock()

	makeEthMocks_claim_unused_prls(&EthMock)

	jobs.EthWrapper = EthMock

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
	suite.Equal(0, len(SentToSendGas))
	suite.Equal(false, sendGasMockCalled_claim_unusued_prls)

	jobs.ResendTimedOutGasTransfers(time.Now().Add(20 * time.Minute))
	// should be one timed out
	suite.Equal(1, len(SentToSendGas))
	suite.Equal(true, sendGasMockCalled_claim_unusued_prls)

	resetTestVariables()
	SentToSendGas = SentToSendGas[:0]

	jobs.ResendTimedOutGasTransfers(time.Now().Add(-20 * time.Minute))
	// should be none, nothing timed out yet
	suite.Equal(0, len(SentToSendGas))
	suite.Equal(false, sendGasMockCalled_claim_unusued_prls)
}

func testResendTimedOutPRLTransfers(suite *JobsSuite) {
	suite.Equal(0, len(SentToClaimUnusedPRLs))
	suite.Equal(false, claimUnusedPRLsMockCalled_claim_unusued_prls)

	jobs.ResendTimedOutPRLTransfers(time.Now().Add(20 * time.Minute))
	// should be one timed out
	suite.Equal(1, len(SentToClaimUnusedPRLs))
	suite.Equal(true, claimUnusedPRLsMockCalled_claim_unusued_prls)

	resetTestVariables()

	jobs.ResendTimedOutPRLTransfers(time.Now().Add(-20 * time.Minute))
	// should be none, nothing timed out yet
	suite.Equal(0, len(SentToClaimUnusedPRLs))
	suite.Equal(false, claimUnusedPRLsMockCalled_claim_unusued_prls)
}

func testResendErroredGasTransfers(suite *JobsSuite) {
	defer resetTestVariables()

	suite.Equal(0, len(SentToSendGas))
	suite.Equal(false, sendGasMockCalled_claim_unusued_prls)

	jobs.ResendErroredGasTransfers()

	// should be one error'd gas transfer that gets resent
	suite.Equal(1, len(SentToSendGas))
	suite.Equal(true, sendGasMockCalled_claim_unusued_prls)
}

func testResendErroredPRLTransfers(suite *JobsSuite) {
	defer resetTestVariables()

	suite.Equal(0, len(SentToClaimUnusedPRLs))
	suite.Equal(false, claimUnusedPRLsMockCalled_claim_unusued_prls)

	jobs.ResendErroredPRLTransfers()

	// should be one error'd prl transfer that gets resent
	suite.Equal(1, len(SentToClaimUnusedPRLs))
	suite.Equal(true, claimUnusedPRLsMockCalled_claim_unusued_prls)
}

func testSendGasForNewClaims(suite *JobsSuite) {
	defer resetTestVariables()

	suite.Equal(0, len(SentToSendGas))
	suite.Equal(false, sendGasMockCalled_claim_unusued_prls)

	jobs.SendGasForNewClaims()

	// should be one new gas transfer that gets sent
	suite.Equal(1, len(SentToSendGas))
	suite.Equal(true, sendGasMockCalled_claim_unusued_prls)
}

func testStartNewClaims(suite *JobsSuite) {
	defer resetTestVariables()

	suite.Equal(0, len(SentToClaimUnusedPRLs))
	suite.Equal(false, claimUnusedPRLsMockCalled_claim_unusued_prls)

	jobs.StartNewClaims()

	// should be one new prl claim that gets sent
	suite.Equal(1, len(SentToClaimUnusedPRLs))
	suite.Equal(true, claimUnusedPRLsMockCalled_claim_unusued_prls)
}

func testInitiateGasTransfer(suite *JobsSuite) {
	defer resetTestVariables()

	suite.Equal(0, len(SentToSendGas))
	suite.Equal(false, sendGasMockCalled_claim_unusued_prls)

	completedUploads := []models.CompletedUpload{}
	err := suite.DB.All(&completedUploads)
	suite.Nil(err)
	suite.NotEqual(0, len(completedUploads))

	jobs.InitiateGasTransfer(completedUploads)

	suite.Equal(len(completedUploads), len(SentToSendGas))
	suite.Equal(true, sendGasMockCalled_claim_unusued_prls)
}

func testInitiatePRLClaim(suite *JobsSuite) {
	defer resetTestVariables()

	suite.Equal(0, len(SentToClaimUnusedPRLs))
	suite.Equal(false, claimUnusedPRLsMockCalled_claim_unusued_prls)

	completedUploads := []models.CompletedUpload{}
	err := suite.DB.All(&completedUploads)
	suite.Nil(err)
	suite.NotEqual(0, len(completedUploads))

	jobs.InitiatePRLClaim(completedUploads)

	suite.Equal(len(completedUploads), len(SentToClaimUnusedPRLs))
	suite.Equal(true, claimUnusedPRLsMockCalled_claim_unusued_prls)
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

	RowWithGasTransferNotStarted = models.CompletedUpload{
		GenesisHash:   "RowWithGasTransferNotStarted",
		ETHAddr:       "SOME_ETH_ADDR_RowWithGasTransferNotStarted",
		ETHPrivateKey: hex.EncodeToString([]byte("SOME_PRIVATE_KEY_RowWithGasTransferNotStarted")),
		PRLStatus:     models.PRLClaimNotStarted,
		GasStatus:     models.GasTransferNotStarted,
	}
	RowWithGasTransferProcessing = models.CompletedUpload{
		GenesisHash:   "RowWithGasTransferProcessing",
		ETHAddr:       "SOME_ETH_ADDR_RowWithGasTransferProcessing",
		ETHPrivateKey: hex.EncodeToString([]byte("SOME_PRIVATE_KEY_RowWithGasTransferProcessing")),
		PRLStatus:     models.PRLClaimNotStarted,
		GasStatus:     models.GasTransferProcessing,
	}
	RowWithGasTransferSuccess = models.CompletedUpload{
		GenesisHash:   "RowWithGasTransferSuccess",
		ETHAddr:       "SOME_ETH_ADDR_RowWithGasTransferSuccess",
		ETHPrivateKey: hex.EncodeToString([]byte("SOME_PRIVATE_KEY_RowWithGasTransferSuccess")),
		PRLStatus:     models.PRLClaimNotStarted,
		GasStatus:     models.GasTransferSuccess,
	}
	RowWithGasTransferError = models.CompletedUpload{
		GenesisHash:   "RowWithGasTransferError",
		ETHAddr:       "SOME_ETH_ADDR_RowWithGasTransferError",
		ETHPrivateKey: hex.EncodeToString([]byte("SOME_PRIVATE_KEY_RowWithGasTransferError")),
		PRLStatus:     models.PRLClaimNotStarted,
		GasStatus:     models.GasTransferError,
	}
	RowWithPRLClaimProcessing = models.CompletedUpload{
		GenesisHash:   "RowWithPRLClaimProcessing",
		ETHAddr:       "SOME_ETH_ADDR_RowWithPRLClaimProcessing",
		ETHPrivateKey: hex.EncodeToString([]byte("SOME_PRIVATE_KEY_RowWithPRLClaimProcessing")),
		PRLStatus:     models.PRLClaimProcessing,
		GasStatus:     models.GasTransferSuccess,
	}
	RowWithPRLClaimSuccess = models.CompletedUpload{
		GenesisHash:   "RowWithPRLClaimSuccess",
		ETHAddr:       "SOME_ETH_ADDR_RowWithPRLClaimSuccess",
		ETHPrivateKey: hex.EncodeToString([]byte("SOME_PRIVATE_KEY_RowWithPRLClaimSuccess")),
		PRLStatus:     models.PRLClaimSuccess,
		GasStatus:     models.GasTransferSuccess,
	}
	RowWithPRLClaimError = models.CompletedUpload{
		GenesisHash:   "RowWithPRLClaimError",
		ETHAddr:       "SOME_ETH_ADDR_RowWithPRLClaimError",
		ETHPrivateKey: hex.EncodeToString([]byte("SOME_PRIVATE_KEY_RowWithPRLClaimError")),
		PRLStatus:     models.PRLClaimError,
		GasStatus:     models.GasTransferSuccess,
	}

	err := suite.DB.RawQuery("DELETE from completed_uploads").All(&[]models.CompletedUpload{})
	suite.Nil(err)

	_, err = suite.DB.ValidateAndSave(&RowWithGasTransferNotStarted)
	suite.Nil(err)

	_, err = suite.DB.ValidateAndSave(&RowWithGasTransferProcessing)
	suite.Nil(err)

	_, err = suite.DB.ValidateAndSave(&RowWithGasTransferSuccess)
	suite.Nil(err)

	_, err = suite.DB.ValidateAndSave(&RowWithGasTransferError)
	suite.Nil(err)

	_, err = suite.DB.ValidateAndSave(&RowWithPRLClaimProcessing)
	suite.Nil(err)

	_, err = suite.DB.ValidateAndSave(&RowWithPRLClaimSuccess)
	suite.Nil(err)

	_, err = suite.DB.ValidateAndSave(&RowWithPRLClaimError)
	suite.Nil(err)

	resetTestVariables()

	makeEthMocks_claim_unused_prls(&EthMock)
}

func resetTestVariables() {
	SentToClaimUnusedPRLs = nil
	SentToClaimUnusedPRLs = []models.CompletedUpload{}

	SentToSendGas = nil
	SentToSendGas = []models.CompletedUpload{}

	sendGasMockCalled_claim_unusued_prls = false
	claimUnusedPRLsMockCalled_claim_unusued_prls = false
}

func makeEthMocks_claim_unused_prls(ethMock *services.Eth) {
	ethMock.ClaimUnusedPRLs = claimUnusedPRLsMock_claim_unusued_prls
	ethMock.SendGas = sendGasMock_claim_unusued_prls
}

func claimUnusedPRLsMock_claim_unusued_prls(uploads []models.CompletedUpload) error {
	SentToClaimUnusedPRLs = append(SentToClaimUnusedPRLs, uploads...)
	claimUnusedPRLsMockCalled_claim_unusued_prls = true
	return nil
}

func sendGasMock_claim_unusued_prls(uploads []models.CompletedUpload) error {
	SentToSendGas = append(SentToSendGas, uploads...)
	sendGasMockCalled_claim_unusued_prls = true
	return nil
}

package models_test

import (
	"github.com/gobuffalo/pop/nulls"
	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/models"
	"time"
)

var (
	RowWithGasTransferNotStarted                 = models.CompletedUpload{}
	RowWithGasTransferProcessing                 = models.CompletedUpload{}
	RowWithGasTransferSuccess                    = models.CompletedUpload{}
	RowWithGasTransferError                      = models.CompletedUpload{}
	RowWithPRLClaimNotStarted                    = models.CompletedUpload{}
	RowWithPRLClaimProcessing                    = models.CompletedUpload{}
	RowWithPRLClaimSuccess                       = models.CompletedUpload{}
	RowWithPRLClaimError                         = models.CompletedUpload{}
	RowWithGasTransferLeftoversReclaimProcessing = models.CompletedUpload{}
	RowWithGasTransferLeftoversReclaimSuccess    = models.CompletedUpload{}
)

func testSetup(suite *ModelSuite) {

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
	RowWithPRLClaimNotStarted = models.CompletedUpload{
		GenesisHash:   "RowWithPRLClaimNotStarted",
		ETHAddr:       addr.Hex(),
		ETHPrivateKey: key,
		PRLStatus:     models.PRLClaimNotStarted,
		GasStatus:     models.GasTransferSuccess,
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

	err := suite.DB.RawQuery("DELETE FROM completed_uploads").All(&[]models.CompletedUpload{})
	suite.Nil(err)

	_, err = suite.DB.ValidateAndSave(&RowWithGasTransferNotStarted)
	RowWithGasTransferNotStarted.EncryptSessionEthKey()
	suite.Nil(err)

	_, err = suite.DB.ValidateAndSave(&RowWithGasTransferProcessing)
	RowWithGasTransferNotStarted.EncryptSessionEthKey()
	suite.Nil(err)

	_, err = suite.DB.ValidateAndSave(&RowWithGasTransferSuccess)
	RowWithGasTransferSuccess.EncryptSessionEthKey()
	suite.Nil(err)

	_, err = suite.DB.ValidateAndSave(&RowWithGasTransferError)
	RowWithGasTransferError.EncryptSessionEthKey()
	suite.Nil(err)

	_, err = suite.DB.ValidateAndSave(&RowWithPRLClaimNotStarted)
	RowWithPRLClaimNotStarted.EncryptSessionEthKey()
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

	_, err = suite.DB.ValidateAndSave(&RowWithPRLClaimError)
	RowWithGasTransferLeftoversReclaimProcessing.EncryptSessionEthKey()
	suite.Nil(err)

	_, err = suite.DB.ValidateAndSave(&RowWithPRLClaimError)
	RowWithGasTransferLeftoversReclaimSuccess.EncryptSessionEthKey()
	suite.Nil(err)
}

func (suite *ModelSuite) Test_CompletedUploads() {

	testNewCompletedUpload(suite)

	// don't need this for the first test
	testSetup(suite)

	testGetRowsByGasAndPRLStatus(suite)
	testGetRowsByGasStatus(suite)
	testGetRowsByPRLStatus(suite)

	testSetGasStatus(suite)

	testSetup(suite)
	testSetPRLStatus(suite)

	testSetup(suite)
	testGetTimedOutGasTransfers(suite)
	testGetTimedOutPRLTransfers(suite)

	testSetGasStatusByAddress(suite)

	testSetup(suite)
	testSetPRLStatusByAddress(suite)

	testSetup(suite)
	testDeleteCompletedClaims(suite)
}

func testNewCompletedUpload(suite *ModelSuite) {

	fileBytesCount := uint64(2500)
	privateKey := "1111111111111111111111111111111111111111111111111111111111111111"

	session1 := models.UploadSession{
		GenesisHash:   "abcdeff1",
		FileSizeBytes: fileBytesCount,
		Type:          models.SessionTypeBeta,
		NumChunks:     1000,
		ETHAddrAlpha:  nulls.String{string("SOME_ALPHA_ETH_ADDRESS1"), true},
		ETHAddrBeta:   nulls.String{string("SOME_BETA_ETH_ADDRESS1"), true},
		ETHPrivateKey: privateKey,
	}
	vErr, err := session1.StartUploadSession()
	suite.Nil(err)
	suite.False(vErr.HasAny())

	session2 := models.UploadSession{
		GenesisHash:   "abcdeff2",
		FileSizeBytes: fileBytesCount,
		Type:          models.SessionTypeAlpha,
		NumChunks:     1000,
		ETHAddrAlpha:  nulls.String{string("SOME_ALPHA_ETH_ADDRESS2"), true},
		ETHAddrBeta:   nulls.String{string("SOME_BETA_ETH_ADDRESS2"), true},
		ETHPrivateKey: privateKey,
	}
	vErr, err = session2.StartUploadSession()
	suite.Nil(err)
	suite.False(vErr.HasAny())

	err = models.NewCompletedUpload(session1)
	suite.Nil(err)
	err = models.NewCompletedUpload(session2)
	suite.Nil(err)

	completedUploads := []models.CompletedUpload{}
	err = suite.DB.All(&completedUploads)
	suite.Nil(err)

	suite.Equal(2, len(completedUploads))

	for _, completedUpload := range completedUploads {
		suite.True(completedUpload.GenesisHash == "abcdeff1" ||
			completedUpload.GenesisHash == "abcdeff2")
		if completedUpload.GenesisHash == "abcdeff1" {
			suite.Equal("SOME_BETA_ETH_ADDRESS1", completedUpload.ETHAddr)
		}
		if completedUpload.GenesisHash == "abcdeff2" {
			suite.Equal("SOME_ALPHA_ETH_ADDRESS2", completedUpload.ETHAddr)
		}
	}
}

func testGetRowsByGasAndPRLStatus(suite *ModelSuite) {
	// this is one of the rows we created
	results, err := models.GetRowsByGasAndPRLStatus(models.GasTransferError, models.PRLClaimNotStarted)
	suite.Nil(err)
	suite.Equal(1, len(results))
	suite.Equal("RowWithGasTransferError", results[0].GenesisHash)

	// this is not one of the rows we created
	results, err = models.GetRowsByGasAndPRLStatus(models.GasTransferError, models.PRLClaimSuccess)
	suite.Nil(err)
	suite.Equal(0, len(results))
}

func testGetRowsByGasStatus(suite *ModelSuite) {
	// this is one of the rows we created
	results, err := models.GetRowsByGasStatus(models.GasTransferError)
	suite.Nil(err)
	suite.Equal(1, len(results))
	suite.Equal("RowWithGasTransferError", results[0].GenesisHash)

	// we created several rows with this gas status
	results, err = models.GetRowsByGasStatus(models.GasTransferSuccess)
	suite.Nil(err)
	suite.Equal(5, len(results))
}

func testSetGasStatus(suite *ModelSuite) {
	// should only be 1 of these
	startingResults, err := models.GetRowsByGasStatus(models.GasTransferNotStarted)
	suite.Nil(err)
	suite.Equal(1, len(startingResults))

	// should only be 1 of these as well
	entriesBeingChanged, err := models.GetRowsByGasStatus(models.GasTransferError)
	suite.Nil(err)
	suite.Equal(1, len(entriesBeingChanged))

	models.SetGasStatus(entriesBeingChanged, models.GasTransferNotStarted)

	// should now be 2 of these
	currentResults, err := models.GetRowsByGasStatus(models.GasTransferNotStarted)
	suite.Nil(err)
	suite.Equal(2, len(currentResults))

	// we changed a row so change it back
	testSetup(suite)
}

func testGetRowsByPRLStatus(suite *ModelSuite) {
	// this is one of the rows we created
	results, err := models.GetRowsByPRLStatus(models.PRLClaimError)
	suite.Nil(err)
	suite.Equal(1, len(results))
	suite.Equal("RowWithPRLClaimError", results[0].GenesisHash)
}

func testSetPRLStatus(suite *ModelSuite) {
	// should be 3 of these
	startingResults, err := models.GetRowsByPRLStatus(models.PRLClaimSuccess)
	suite.Nil(err)
	suite.Equal(3, len(startingResults))

	// should only be 1 of these as well
	entriesBeingChanged, err := models.GetRowsByPRLStatus(models.PRLClaimError)
	suite.Nil(err)
	suite.Equal(1, len(entriesBeingChanged))

	models.SetPRLStatus(entriesBeingChanged, models.PRLClaimSuccess)

	// should now be 4 of these
	currentResults, err := models.GetRowsByPRLStatus(models.PRLClaimSuccess)
	suite.Nil(err)
	suite.Equal(4, len(currentResults))

	// we changed a row so change it back
	testSetup(suite)
}

func testGetTimedOutGasTransfers(suite *ModelSuite) {
	// should be none
	shouldBeNone, err := models.GetTimedOutGasTransfers(time.Now().Add(-5 * time.Minute))
	suite.Nil(err)
	suite.Equal(0, len(shouldBeNone))

	// should be 1
	shouldBeOne, err := models.GetTimedOutGasTransfers(time.Now().Add(5 * time.Minute))
	suite.Nil(err)
	suite.Equal(1, len(shouldBeOne))
}

func testGetTimedOutPRLTransfers(suite *ModelSuite) {
	// should be none
	shouldBeNone, err := models.GetTimedOutPRLTransfers(time.Now().Add(-5 * time.Minute))
	suite.Nil(err)
	suite.Equal(0, len(shouldBeNone))

	// should be 1
	shouldBeOne, err := models.GetTimedOutPRLTransfers(time.Now().Add(5 * time.Minute))
	suite.Nil(err)
	suite.Equal(1, len(shouldBeOne))
}

func testSetGasStatusByAddress(suite *ModelSuite) {
	// should only be 1 of these
	startingResultsNotStarted, err := models.GetRowsByGasStatus(models.GasTransferNotStarted)
	suite.Nil(err)
	suite.Equal(1, len(startingResultsNotStarted))

	// should only be 1 of these
	startingResultsError, err := models.GetRowsByGasStatus(models.GasTransferError)
	suite.Nil(err)
	suite.Equal(1, len(startingResultsError))

	models.SetGasStatusByAddress(startingResultsNotStarted[0].ETHAddr, models.GasTransferError)

	// should be none left
	currentResultsNotStarted, err := models.GetRowsByGasStatus(models.GasTransferNotStarted)
	suite.Nil(err)
	suite.Equal(0, len(currentResultsNotStarted))

	// should only be 2 of these
	currentResultsError, err := models.GetRowsByGasStatus(models.GasTransferError)
	suite.Nil(err)
	suite.Equal(2, len(currentResultsError))
}

func testSetPRLStatusByAddress(suite *ModelSuite) {
	// should be 3 of these
	startingResultsSuccess, err := models.GetRowsByPRLStatus(models.PRLClaimSuccess)
	suite.Nil(err)
	suite.Equal(3, len(startingResultsSuccess))

	// should only be 1 of these
	startingResultsError, err := models.GetRowsByPRLStatus(models.PRLClaimError)
	suite.Nil(err)
	suite.Equal(1, len(startingResultsError))

	models.SetPRLStatusByAddress(startingResultsSuccess[0].ETHAddr, models.PRLClaimError)

	// should be 2 left
	currentResultsSuccess, err := models.GetRowsByPRLStatus(models.PRLClaimSuccess)
	suite.Nil(err)
	suite.Equal(2, len(currentResultsSuccess))

	// should only be 2 of these
	currentResultsError, err := models.GetRowsByPRLStatus(models.PRLClaimError)
	suite.Nil(err)
	suite.Equal(2, len(currentResultsError))
}

func testDeleteCompletedClaims(suite *ModelSuite) {
	//should be 10 of these
	completedUploads := []models.CompletedUpload{}
	err := suite.DB.All(&completedUploads)
	suite.Nil(err)
	suite.Equal(10, len(completedUploads))

	models.DeleteCompletedClaims()

	//should be 9 now
	completedUploads = []models.CompletedUpload{}
	err = suite.DB.All(&completedUploads)
	suite.Nil(err)
	suite.Equal(9, len(completedUploads))
}

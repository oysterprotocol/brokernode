package models_test

import (
	"github.com/gobuffalo/pop/nulls"
	"github.com/oysterprotocol/brokernode/models"
	"time"
)

var (
	RowWithGasTransferNotStarted = models.CompletedUpload{}
	RowWithGasTransferProcessing = models.CompletedUpload{}
	RowWithGasTransferSuccess    = models.CompletedUpload{}
	RowWithGasTransferError      = models.CompletedUpload{}
	RowWithPRLClaimNotStarted    = models.CompletedUpload{}
	RowWithPRLClaimProcessing    = models.CompletedUpload{}
	RowWithPRLClaimSuccess       = models.CompletedUpload{}
	RowWithPRLClaimError         = models.CompletedUpload{}
)

func testSetup(ms *ModelSuite) {

	RowWithGasTransferNotStarted = models.CompletedUpload{
		GenesisHash:   "RowWithGasTransferNotStarted",
		ETHAddr:       "SOME_ETH_ADDR_RowWithGasTransferNotStarted",
		ETHPrivateKey: "SOME_PRIVATE_KEY_RowWithGasTransferNotStarted",
		PRLStatus:     models.PRLClaimNotStarted,
		GasStatus:     models.GasTransferNotStarted,
	}
	RowWithGasTransferProcessing = models.CompletedUpload{
		GenesisHash:   "RowWithGasTransferProcessing",
		ETHAddr:       "SOME_ETH_ADDR_RowWithGasTransferProcessing",
		ETHPrivateKey: "SOME_PRIVATE_KEY_RowWithGasTransferProcessing",
		PRLStatus:     models.PRLClaimNotStarted,
		GasStatus:     models.GasTransferProcessing,
	}
	RowWithGasTransferSuccess = models.CompletedUpload{
		GenesisHash:   "RowWithGasTransferSuccess",
		ETHAddr:       "SOME_ETH_ADDR_RowWithGasTransferSuccess",
		ETHPrivateKey: "SOME_PRIVATE_KEY_RowWithGasTransferSuccess",
		PRLStatus:     models.PRLClaimNotStarted,
		GasStatus:     models.GasTransferSuccess,
	}
	RowWithGasTransferError = models.CompletedUpload{
		GenesisHash:   "RowWithGasTransferError",
		ETHAddr:       "SOME_ETH_ADDR_RowWithGasTransferError",
		ETHPrivateKey: "SOME_PRIVATE_KEY_RowWithGasTransferError",
		PRLStatus:     models.PRLClaimNotStarted,
		GasStatus:     models.GasTransferError,
	}
	RowWithPRLClaimNotStarted = models.CompletedUpload{
		GenesisHash:   "RowWithPRLClaimNotStarted",
		ETHAddr:       "SOME_ETH_ADDR_RowWithPRLClaimNotStarted",
		ETHPrivateKey: "SOME_PRIVATE_KEY_RowWithPRLClaimNotStarted",
		PRLStatus:     models.PRLClaimNotStarted,
		GasStatus:     models.GasTransferSuccess,
	}
	RowWithPRLClaimProcessing = models.CompletedUpload{
		GenesisHash:   "RowWithPRLClaimProcessing",
		ETHAddr:       "SOME_ETH_ADDR_RowWithPRLClaimProcessing",
		ETHPrivateKey: "SOME_PRIVATE_KEY_RowWithPRLClaimProcessing",
		PRLStatus:     models.PRLClaimProcessing,
		GasStatus:     models.GasTransferSuccess,
	}
	RowWithPRLClaimSuccess = models.CompletedUpload{
		GenesisHash:   "RowWithPRLClaimSuccess",
		ETHAddr:       "SOME_ETH_ADDR_RowWithPRLClaimSuccess",
		ETHPrivateKey: "SOME_PRIVATE_KEY_RowWithPRLClaimSuccess",
		PRLStatus:     models.PRLClaimSuccess,
		GasStatus:     models.GasTransferSuccess,
	}
	RowWithPRLClaimError = models.CompletedUpload{
		GenesisHash:   "RowWithPRLClaimError",
		ETHAddr:       "SOME_ETH_ADDR_RowWithPRLClaimError",
		ETHPrivateKey: "SOME_PRIVATE_KEY_RowWithPRLClaimError",
		PRLStatus:     models.PRLClaimError,
		GasStatus:     models.GasTransferSuccess,
	}

	err := ms.DB.RawQuery("DELETE from completed_uploads").All(&[]models.CompletedUpload{})
	ms.Nil(err)

	_, err = ms.DB.ValidateAndSave(&RowWithGasTransferNotStarted)
	ms.Nil(err)

	_, err = ms.DB.ValidateAndSave(&RowWithGasTransferProcessing)
	ms.Nil(err)

	_, err = ms.DB.ValidateAndSave(&RowWithGasTransferSuccess)
	ms.Nil(err)

	_, err = ms.DB.ValidateAndSave(&RowWithGasTransferError)
	ms.Nil(err)

	_, err = ms.DB.ValidateAndSave(&RowWithPRLClaimNotStarted)
	ms.Nil(err)

	_, err = ms.DB.ValidateAndSave(&RowWithPRLClaimProcessing)
	ms.Nil(err)

	_, err = ms.DB.ValidateAndSave(&RowWithPRLClaimSuccess)
	ms.Nil(err)

	_, err = ms.DB.ValidateAndSave(&RowWithPRLClaimError)
	ms.Nil(err)
}

func (ms *ModelSuite) Test_CompletedUploads() {

	testNewCompletedUpload(ms)

	// don't need this for the first test
	testSetup(ms)

	testGetRowsByGasAndPRLStatus(ms)
	testGetRowsByGasStatus(ms)
	testGetRowsByPRLStatus(ms)

	testSetGasStatus(ms)

	testSetup(ms)
	testSetPRLStatus(ms)

	testSetup(ms)
	testGetTimedOutGasTransfers(ms)
	testGetTimedOutPRLTransfers(ms)

	testSetGasStatusByAddress(ms)

	testSetup(ms)
	testSetPRLStatusByAddress(ms)

	testSetup(ms)
	testDeleteCompletedClaims(ms)
}

func testNewCompletedUpload(ms *ModelSuite) {

	fileBytesCount := 2500

	err := models.NewCompletedUpload(models.UploadSession{
		GenesisHash:   "abcdeff1",
		FileSizeBytes: fileBytesCount,
		Type:          models.SessionTypeBeta,
		ETHAddrAlpha:  nulls.String{string("SOME_ALPHA_ETH_ADDRESS"), true},
		ETHAddrBeta:   nulls.String{string("SOME_BETA_ETH_ADDRESS"), true},
		ETHPrivateKey: "SOME_PRIVATE_KEY",
	})
	ms.Equal(nil, err)

	err = models.NewCompletedUpload(models.UploadSession{
		GenesisHash:   "abcdeff2",
		FileSizeBytes: fileBytesCount,
		Type:          models.SessionTypeAlpha,
		ETHAddrAlpha:  nulls.String{string("SOME_ALPHA_ETH_ADDRESS"), true},
		ETHAddrBeta:   nulls.String{string("SOME_BETA_ETH_ADDRESS"), true},
		ETHPrivateKey: "SOME_PRIVATE_KEY",
	})
	ms.Equal(nil, err)

	err = models.NewCompletedUpload(models.UploadSession{ // no session type
		GenesisHash:   "abcdeff3",
		FileSizeBytes: fileBytesCount,
		ETHAddrAlpha:  nulls.String{string("SOME_ALPHA_ETH_ADDRESS"), true},
		ETHAddrBeta:   nulls.String{string("SOME_BETA_ETH_ADDRESS"), true},
		ETHPrivateKey: "SOME_PRIVATE_KEY",
	})
	ms.NotEqual(err, nil)
	ms.Equal("no session type provided for session in method models.NewCompletedUpload", err.Error())

	completedUploads := []models.CompletedUpload{}
	err = ms.DB.All(&completedUploads)
	ms.Equal(nil, err)

	ms.Equal(2, len(completedUploads))

	for _, completedUpload := range completedUploads {
		ms.Equal(true, completedUpload.GenesisHash == "abcdeff1" ||
			completedUpload.GenesisHash == "abcdeff2")
		if completedUpload.GenesisHash == "abcdeff1" {
			ms.Equal("SOME_BETA_ETH_ADDRESS", completedUpload.ETHAddr)
		}
		if completedUpload.GenesisHash == "abcdeff2" {
			ms.Equal("SOME_ALPHA_ETH_ADDRESS", completedUpload.ETHAddr)
		}
	}
}

func testGetRowsByGasAndPRLStatus(ms *ModelSuite) {
	// this is one of the rows we created
	results, err := models.GetRowsByGasAndPRLStatus(models.GasTransferError, models.PRLClaimNotStarted)
	ms.Nil(err)
	ms.Equal(1, len(results))
	ms.Equal("RowWithGasTransferError", results[0].GenesisHash)

	// this is not one of the rows we created
	results, err = models.GetRowsByGasAndPRLStatus(models.GasTransferError, models.PRLClaimSuccess)
	ms.Nil(err)
	ms.Equal(0, len(results))
}

func testGetRowsByGasStatus(ms *ModelSuite) {
	// this is one of the rows we created
	results, err := models.GetRowsByGasStatus(models.GasTransferError)
	ms.Nil(err)
	ms.Equal(1, len(results))
	ms.Equal("RowWithGasTransferError", results[0].GenesisHash)

	// we created several rows with this gas status
	results, err = models.GetRowsByGasStatus(models.GasTransferSuccess)
	ms.Nil(err)
	ms.Equal(5, len(results))
}

func testSetGasStatus(ms *ModelSuite) {
	// should only be 1 of these
	startingResults, err := models.GetRowsByGasStatus(models.GasTransferNotStarted)
	ms.Nil(err)
	ms.Equal(1, len(startingResults))

	// should only be 1 of these as well
	entriesBeingChanged, err := models.GetRowsByGasStatus(models.GasTransferError)
	ms.Nil(err)
	ms.Equal(1, len(entriesBeingChanged))

	models.SetGasStatus(entriesBeingChanged, models.GasTransferNotStarted)

	// should now be 2 of these
	currentResults, err := models.GetRowsByGasStatus(models.GasTransferNotStarted)
	ms.Nil(err)
	ms.Equal(2, len(currentResults))

	// we changed a row so change it back
	testSetup(ms)
}

func testGetRowsByPRLStatus(ms *ModelSuite) {
	// this is one of the rows we created
	results, err := models.GetRowsByPRLStatus(models.PRLClaimError)
	ms.Nil(err)
	ms.Equal(1, len(results))
	ms.Equal("RowWithPRLClaimError", results[0].GenesisHash)
}

func testSetPRLStatus(ms *ModelSuite) {
	// should only be 1 of these
	startingResults, err := models.GetRowsByPRLStatus(models.PRLClaimSuccess)
	ms.Nil(err)
	ms.Equal(1, len(startingResults))

	// should only be 1 of these as well
	entriesBeingChanged, err := models.GetRowsByPRLStatus(models.PRLClaimError)
	ms.Nil(err)
	ms.Equal(1, len(entriesBeingChanged))

	models.SetPRLStatus(entriesBeingChanged, models.PRLClaimSuccess)

	// should now be 2 of these
	currentResults, err := models.GetRowsByPRLStatus(models.PRLClaimSuccess)
	ms.Nil(err)
	ms.Equal(2, len(currentResults))

	// we changed a row so change it back
	testSetup(ms)
}

func testGetTimedOutGasTransfers(ms *ModelSuite) {
	// should be none
	shouldBeNone, err := models.GetTimedOutGasTransfers(time.Now().Add(-5 * time.Minute))
	ms.Nil(err)
	ms.Equal(0, len(shouldBeNone))

	// should be 1
	shouldBeOne, err := models.GetTimedOutGasTransfers(time.Now().Add(5 * time.Minute))
	ms.Nil(err)
	ms.Equal(1, len(shouldBeOne))
}

func testGetTimedOutPRLTransfers(ms *ModelSuite) {
	// should be none
	shouldBeNone, err := models.GetTimedOutPRLTransfers(time.Now().Add(-5 * time.Minute))
	ms.Nil(err)
	ms.Equal(0, len(shouldBeNone))

	// should be 1
	shouldBeOne, err := models.GetTimedOutPRLTransfers(time.Now().Add(5 * time.Minute))
	ms.Nil(err)
	ms.Equal(1, len(shouldBeOne))
}

func testSetGasStatusByAddress(ms *ModelSuite) {
	// should only be 1 of these
	startingResultsNotStarted, err := models.GetRowsByGasStatus(models.GasTransferNotStarted)
	ms.Nil(err)
	ms.Equal(1, len(startingResultsNotStarted))

	// should only be 1 of these
	startingResultsError, err := models.GetRowsByGasStatus(models.GasTransferError)
	ms.Nil(err)
	ms.Equal(1, len(startingResultsError))

	models.SetGasStatusByAddress(startingResultsNotStarted[0].ETHAddr, models.GasTransferError)

	// should be none left
	currentResultsNotStarted, err := models.GetRowsByGasStatus(models.GasTransferNotStarted)
	ms.Nil(err)
	ms.Equal(0, len(currentResultsNotStarted))

	// should only be 2 of these
	currentResultsError, err := models.GetRowsByGasStatus(models.GasTransferError)
	ms.Nil(err)
	ms.Equal(2, len(currentResultsError))
}

func testSetPRLStatusByAddress(ms *ModelSuite) {
	// should only be 1 of these
	startingResultsSuccess, err := models.GetRowsByPRLStatus(models.PRLClaimSuccess)
	ms.Nil(err)
	ms.Equal(1, len(startingResultsSuccess))

	// should only be 1 of these
	startingResultsError, err := models.GetRowsByPRLStatus(models.PRLClaimError)
	ms.Nil(err)
	ms.Equal(1, len(startingResultsError))

	models.SetPRLStatusByAddress(startingResultsSuccess[0].ETHAddr, models.PRLClaimError)

	// should be none left
	currentResultsSuccess, err := models.GetRowsByPRLStatus(models.PRLClaimSuccess)
	ms.Nil(err)
	ms.Equal(0, len(currentResultsSuccess))

	// should only be 2 of these
	currentResultsError, err := models.GetRowsByPRLStatus(models.PRLClaimError)
	ms.Nil(err)
	ms.Equal(2, len(currentResultsError))
}

func testDeleteCompletedClaims(ms *ModelSuite) {
	//should be 8 of these
	completedUploads := []models.CompletedUpload{}
	err := ms.DB.All(&completedUploads)
	ms.Equal(nil, err)
	ms.Equal(8, len(completedUploads))

	models.DeleteCompletedClaims()

	//should be 7 now
	completedUploads = []models.CompletedUpload{}
	err = ms.DB.All(&completedUploads)
	ms.Equal(nil, err)
	ms.Equal(7, len(completedUploads))
}

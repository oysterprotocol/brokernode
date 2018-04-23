package jobs

import (
	"fmt"
	"github.com/getsentry/raven-go"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"log"
	"time"
)

func init() {
}

func ClaimUnusedPRLs(thresholdTime time.Time) {

	ResendTimedOutGasTransfers(thresholdTime)
	ResendTimedOutPRLTransfers(thresholdTime)

	ResendErroredGasTransfers()
	ResendErroredPRLTransfers()

	SendGasForNewClaims()
	StartNewClaims()

	PurgeCompletedClaims()
}

// for gas transfers that are still processing by the time of the threshold
func ResendTimedOutGasTransfers(thresholdTime time.Time) {
	timedOutGasTransfers, err := models.GetTimedOutGasTransfers(thresholdTime)
	if err != nil {
		fmt.Println(err)
		log.Println("Error getting timed out gas transfers")
		raven.CaptureError(err, nil)
		return
	}
	if len(timedOutGasTransfers) > 0 {
		InitiateGasTransfer(timedOutGasTransfers)
	}
}

// for prl transfers that are still processing by the time of the threshold
func ResendTimedOutPRLTransfers(thresholdTime time.Time) {
	timedOutPRLTransfers, err := models.GetTimedOutPRLTransfers(thresholdTime)
	if err != nil {
		fmt.Println(err)
		log.Println("Error getting timed out gas transfers")
		raven.CaptureError(err, nil)
		return
	}
	if len(timedOutPRLTransfers) > 0 {
		InitiatePRLClaim(timedOutPRLTransfers)
	}
}

// for gas transfers that are in an error state
func ResendErroredGasTransfers() {
	gasTransferErrors, err := models.GetRowsByGasStatus(models.GasTransferError)
	if err != nil {
		fmt.Println(err)
		log.Println("Error getting completed uploads whose gas transfers errored")
		raven.CaptureError(err, nil)
		return
	}
	if len(gasTransferErrors) > 0 {
		InitiateGasTransfer(gasTransferErrors)
	}
}

// for prl transfers that are in an error state
func ResendErroredPRLTransfers() {
	prlTransferErrors, err := models.GetRowsByPRLStatus(models.PRLClaimError)
	if err != nil {
		fmt.Println(err)
		log.Println("Error getting completed uploads whose PRL transfers errored")
		raven.CaptureError(err, nil)
		return
	}
	if len(prlTransferErrors) > 0 {
		InitiatePRLClaim(prlTransferErrors)
	}
}

// for new claims with no gas
func SendGasForNewClaims() {
	needGas, err := models.GetRowsByGasStatus(models.GasTransferNotStarted)
	if err != nil {
		fmt.Println(err)
		log.Println("Error getting completed uploads whose addresses need gas.")
		raven.CaptureError(err, nil)
		return
	}
	if len(needGas) > 0 {
		InitiateGasTransfer(needGas)
	}
}

// for claims whose gas transfers succeeded but there is still unclaimed PRL
func StartNewClaims() {
	readyClaims, err := models.GetRowsByGasAndPRLStatus(models.GasTransferSuccess, models.PRLClaimNotStarted)
	if err != nil {
		fmt.Println(err)
		log.Println("Error getting ready claims.")
		raven.CaptureError(err, nil)
		return
	}
	if len(readyClaims) > 0 {
		InitiatePRLClaim(readyClaims)
	}
}

// wraps calls eth_gatway's SendGas method and sets GasStatus to GasTransferProcessing
func InitiateGasTransfer(uploadsThatNeedGas []models.CompletedUpload) {
	err := services.SendGas(uploadsThatNeedGas)
	if err != nil {
		fmt.Println(err)
		log.Println("Error sending gas.")
		raven.CaptureError(err, nil)
		return
	}
	//TODO un-comment this out when ETH stuff works
	//models.SetGasStatus(needGas, models.GasTransferProcessing)
}

// wraps calls eth_gatway's ClaimPRLs method and sets GasStatus to GasTransferProcessing
func InitiatePRLClaim(uploadsWithUnclaimedPRLs []models.CompletedUpload) {
	err := services.ClaimPRLs(uploadsWithUnclaimedPRLs)
	if err != nil {
		fmt.Println(err)
		log.Println("Error claiming PRL.")
		raven.CaptureError(err, nil)
		return
	}
	//TODO un-comment this out when ETH stuff works
	//models.SetPRLStatus(uploadsWithUnclaimedPRLs, models.PRLClaimProcessing)
}

// purge claims whose PRLStatus is PRLClaimSuccess
func PurgeCompletedClaims() {
	err := models.DeleteCompletedClaims()
	if err != nil {
		fmt.Println(err)
		log.Println("Error purging completed claims.")
		raven.CaptureError(err, nil)
		return
	}
}

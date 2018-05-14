package jobs

import (
	"github.com/getsentry/raven-go"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"gopkg.in/segmentio/analytics-go.v3"
	"log"
	"time"
)

var ethWrapper = services.EthWrapper

func init() {
}

func ClaimUnusedPRLs(ethService services.Eth, thresholdTime time.Time) {

	if oyster_utils.BrokerMode == oyster_utils.ProdMode {
		SetEthWrapper(ethService)

		ResendTimedOutGasTransfers(thresholdTime)
		ResendTimedOutPRLTransfers(thresholdTime)

		ResendErroredGasTransfers()
		ResendErroredPRLTransfers()

		SendGasForNewClaims()
		StartNewClaims()

		PurgeCompletedClaims()
	}
}

func SetEthWrapper(ethService services.Eth) {
	ethWrapper = ethService
}

// for gas transfers that are still processing by the time of the threshold
func ResendTimedOutGasTransfers(thresholdTime time.Time) {
	timedOutGasTransfers, err := models.GetTimedOutGasTransfers(thresholdTime)

	if err != nil {
		log.Println("Error getting timed out gas transfers")
		raven.CaptureError(err, nil)
		return
	}
	if len(timedOutGasTransfers) > 0 {

		for _, transfer := range timedOutGasTransfers {
			oyster_utils.LogToSegment("claim_unused_prls: gas_transfer_timed_out", analytics.NewProperties().
				Set("eth_address_to", transfer.ETHAddr).
				Set("genesis_hash", transfer.GenesisHash))
		}

		InitiateGasTransfer(timedOutGasTransfers)
	}
}

// for prl transfers that are still processing by the time of the threshold
func ResendTimedOutPRLTransfers(thresholdTime time.Time) {
	timedOutPRLTransfers, err := models.GetTimedOutPRLTransfers(thresholdTime)
	if err != nil {
		log.Println("Error getting timed out gas transfers")
		raven.CaptureError(err, nil)
		return
	}
	if len(timedOutPRLTransfers) > 0 {

		for _, transfer := range timedOutPRLTransfers {
			oyster_utils.LogToSegment("claim_unused_prls: unclaimed_prl_transfer_timed_out", analytics.NewProperties().
				Set("eth_address_from", transfer.ETHAddr).
				Set("genesis_hash", transfer.GenesisHash))
		}

		InitiatePRLClaim(timedOutPRLTransfers)
	}
}

// for gas transfers that are in an error state
func ResendErroredGasTransfers() {
	gasTransferErrors, err := models.GetRowsByGasStatus(models.GasTransferError)
	if err != nil {
		log.Println("Error getting completed uploads whose gas transfers errored")
		raven.CaptureError(err, nil)
		return
	}
	if len(gasTransferErrors) > 0 {

		for _, transfer := range gasTransferErrors {
			oyster_utils.LogToSegment("claim_unused_prls: gas_transfer_error", analytics.NewProperties().
				Set("eth_address_to", transfer.ETHAddr).
				Set("genesis_hash", transfer.GenesisHash))
		}

		InitiateGasTransfer(gasTransferErrors)
	}
}

// for prl transfers that are in an error state
func ResendErroredPRLTransfers() {
	prlTransferErrors, err := models.GetRowsByPRLStatus(models.PRLClaimError)
	if err != nil {
		log.Println("Error getting completed uploads whose PRL transfers errored")
		raven.CaptureError(err, nil)
		return
	}
	if len(prlTransferErrors) > 0 {

		for _, transfer := range prlTransferErrors {
			oyster_utils.LogToSegment("claim_unused_prls: unclaimed_prl_transfer_error", analytics.NewProperties().
				Set("eth_address_from", transfer.ETHAddr).
				Set("genesis_hash", transfer.GenesisHash))
		}

		InitiatePRLClaim(prlTransferErrors)
	}
}

// for new claims with no gas
func SendGasForNewClaims() {
	needGas, err := models.GetRowsByGasStatus(models.GasTransferNotStarted)
	if err != nil {
		log.Println("Error getting completed uploads whose addresses need gas.")
		raven.CaptureError(err, nil)
		return
	}
	if len(needGas) > 0 {

		for _, transfer := range needGas {
			oyster_utils.LogToSegment("claim_unused_prls: send_gas_for_new_claim", analytics.NewProperties().
				Set("eth_address_to", transfer.ETHAddr).
				Set("genesis_hash", transfer.GenesisHash))
		}

		InitiateGasTransfer(needGas)
	}
}

// for claims whose gas transfers succeeded but there is still unclaimed PRL
func StartNewClaims() {
	readyClaims, err := models.GetRowsByGasAndPRLStatus(models.GasTransferSuccess, models.PRLClaimNotStarted)
	if err != nil {
		log.Println("Error getting ready claims.")
		raven.CaptureError(err, nil)
		return
	}
	if len(readyClaims) > 0 {

		for _, transfer := range readyClaims {
			oyster_utils.LogToSegment("claim_unused_prls: unclaimed_prl_new_claim", analytics.NewProperties().
				Set("eth_address_from", transfer.ETHAddr).
				Set("genesis_hash", transfer.GenesisHash))
		}

		InitiatePRLClaim(readyClaims)
	}
}

// wraps calls eth_gatway's SendGas method and sets GasStatus to GasTransferProcessing
func InitiateGasTransfer(uploadsThatNeedGas []models.CompletedUpload) {
	err := ethWrapper.SendGas(uploadsThatNeedGas)
	if err != nil {
		log.Println("Error sending gas.")
		raven.CaptureError(err, nil)
		return
	}
	//TODO un-comment this out when ETH stuff works
	//models.SetGasStatus(needGas, models.GasTransferProcessing)
}

// wraps calls eth_gatway's ClaimUnusedPRLs method and sets GasStatus to GasTransferProcessing
func InitiatePRLClaim(uploadsWithUnclaimedPRLs []models.CompletedUpload) {
	err := ethWrapper.ClaimUnusedPRLs(uploadsWithUnclaimedPRLs)
	if err != nil {
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
		log.Println("Error purging completed claims.")
		raven.CaptureError(err, nil)
		return
	}
}

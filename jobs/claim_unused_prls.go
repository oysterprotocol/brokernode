package jobs

import (
	"errors"
	"fmt"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"gopkg.in/segmentio/analytics-go.v3"
	"time"
)

func ClaimUnusedPRLs(thresholdTime time.Time, PrometheusWrapper services.PrometheusService) {

	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramClaimUnusedPRLs, start)

	if oyster_utils.BrokerMode == oyster_utils.ProdMode {

		ResendTimedOutGasTransfers(thresholdTime)
		ResendTimedOutPRLTransfers(thresholdTime)

		ResendErroredGasTransfers()
		ResendErroredPRLTransfers()

		SendGasForNewClaims()
		StartNewClaims()

		PurgeCompletedClaims()
	}
}

// for gas transfers that are still processing by the time of the threshold
func ResendTimedOutGasTransfers(thresholdTime time.Time) {
	timedOutGasTransfers, err := models.GetTimedOutGasTransfers(thresholdTime)

	if err != nil {
		oyster_utils.LogIfError(fmt.Errorf("Error getting timed out gas transfers: %v", err), nil)
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
		oyster_utils.LogIfError(fmt.Errorf("Error getting timed out gas transfers: %v", err), nil)
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
		oyster_utils.LogIfError(fmt.Errorf("Error getting completed uploads whose gas transfers errored: %v", err), nil)
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
		oyster_utils.LogIfError(fmt.Errorf("Error getting completed uploads whose PRL transfers errored: %v", err), nil)
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
		oyster_utils.LogIfError(fmt.Errorf("Error getting completed uploads whose addresses need gas: %v", err), nil)
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
		oyster_utils.LogIfError(fmt.Errorf("Error getting ready claims: %v", err), nil)
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

// wraps calls eth_gatway's SendETH method and sets GasStatus to GasTransferProcessing
func InitiateGasTransfer(uploadsThatNeedGas []models.CompletedUpload) {
	gasToSend, err := EthWrapper.CalculateGasToSend(services.GasLimitPRLSend)
	if err != nil {
		oyster_utils.LogIfError(fmt.Errorf("Error determining gas to send: %v", err), nil)
		return
	}
	for _, upload := range uploadsThatNeedGas {
		_, txHash, nonce, err := EthWrapper.SendETH(services.StringToAddress(upload.ETHAddr), gasToSend)
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			return
		}
		upload.GasStatus = models.GasTransferProcessing
		upload.GasTxHash = txHash
		upload.GasTxNonce = nonce
		models.DB.ValidateAndUpdate(&upload)
	}
}

// wraps calls eth_gatway's ClaimUnusedPRLs method and sets PRLStatus to PRLClaimProcessing
func InitiatePRLClaim(uploadsWithUnclaimedPRLs []models.CompletedUpload) {
	for _, upload := range uploadsWithUnclaimedPRLs {
		balance := EthWrapper.CheckPRLBalance(services.StringToAddress(upload.ETHAddr))

		privateKey := upload.DecryptSessionEthKey()
		ecdsaPrivateKey, err := services.StringToPrivateKey(privateKey)

		if err != nil {
			oyster_utils.LogIfError(err, nil)
			return
		}
		callMsg, err := EthWrapper.CreateSendPRLMessage(services.StringToAddress(upload.ETHAddr),
			ecdsaPrivateKey, services.MainWalletAddress, *balance)
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			return
		}
		sendSuccess, txHash, nonce := EthWrapper.SendPRLFromOyster(callMsg)
		if sendSuccess {
			upload.PRLStatus = models.PRLClaimProcessing
			upload.PRLTxHash = txHash
			upload.PRLTxNonce = nonce
			models.DB.ValidateAndUpdate(&upload)
		} else {
			err := errors.New("error claiming unused prls from addresss: " + upload.ETHAddr)
			oyster_utils.LogIfError(err, nil)
		}
	}
}

// purge claims whose PRLStatus is PRLClaimSuccess
func PurgeCompletedClaims() {
	err := models.DeleteCompletedClaims()
	if err != nil {
		oyster_utils.LogIfError(fmt.Errorf("Error purging completed claims: %v", err), nil)
		return
	}
}

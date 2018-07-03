package jobs

import (
	"errors"
	"fmt"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"gopkg.in/segmentio/analytics-go.v3"
	"math/big"
	"time"
)

func ClaimUnusedPRLs(thresholdTime time.Time, PrometheusWrapper services.PrometheusService) {

	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramClaimUnusedPRLs, start)

	if oyster_utils.BrokerMode == oyster_utils.ProdMode {

		CheckProcessingGasTransactions()
		CheckProcessingPRLTransactions()
		CheckProcessingGasReclaims()

		ResendTimedOutGasTransfers(thresholdTime)
		ResendTimedOutPRLTransfers(thresholdTime)
		ResendTimedOutGasReclaims(thresholdTime)

		ResendErroredGasTransfers()
		ResendErroredPRLTransfers()

		SendGasForNewClaims()
		StartNewClaims()
		RetrieveLeftoverETH()

		PurgeCompletedClaims()
	}
}

func CheckProcessingGasTransactions() {
	gasPending, err := models.GetRowsByGasStatus(models.GasTransferProcessing)
	if err != nil {
		fmt.Println("Cannot get completed_uploads with pending gas transfers: " + err.Error())
		// already captured error in upstream function
		return
	}

	for _, pending := range gasPending {
		ethBalance := EthWrapper.CheckETHBalance(services.StringToAddress(pending.ETHAddr))
		if ethBalance.Int64() > 0 {
			fmt.Println("ETH (gas) transaction sent to " + pending.ETHAddr + " in CheckProcessingGasTransactions() " +
				"in claim_unused_prls")
			pending.GasStatus = models.GasTransferSuccess
			vErr, err := models.DB.ValidateAndUpdate(&pending)
			if err != nil {
				oyster_utils.LogIfError(err, nil)
				continue
			}
			if len(vErr.Errors) > 0 {
				oyster_utils.LogIfValidationError(
					"validation errors in claim_unused_prls in CheckProcessingGasTransaction", vErr, nil)
				continue
			}
			oyster_utils.LogToSegment("claim_unused_prls: CheckProcessingGasTransactions", analytics.NewProperties().
				Set("new_status", models.GasTransferStatusMap[pending.GasStatus]).
				Set("eth_address", pending.ETHAddr))
		}
	}
}

func CheckProcessingPRLTransactions() {
	prlsPending, err := models.GetRowsByPRLStatus(models.PRLClaimProcessing)
	if err != nil {
		fmt.Println("Cannot get completed_uploads with pending PRL retrieval: " + err.Error())
		// already captured error in upstream function
		return
	}

	for _, pending := range prlsPending {
		prlBalance := EthWrapper.CheckPRLBalance(services.StringToAddress(pending.ETHAddr))
		if prlBalance.Int64() == int64(0) {
			fmt.Println("Unused PRLs retrieved from " + pending.ETHAddr + " in CheckProcessingPRLTransactions() " +
				"in claim_unused_prls")
			pending.PRLStatus = models.PRLClaimSuccess
			vErr, err := models.DB.ValidateAndUpdate(&pending)
			if err != nil {
				oyster_utils.LogIfError(err, nil)
				continue
			}
			if len(vErr.Errors) > 0 {
				oyster_utils.LogIfValidationError(
					"validation errors in claim_unused_prls in CheckProcessingPRLTransactions", vErr, nil)
				continue
			}
			oyster_utils.LogToSegment("claim_unused_prls: CheckProcessingPRLTransactions", analytics.NewProperties().
				Set("new_status", models.PRLClaimStatusMap[pending.PRLStatus]).
				Set("eth_address", pending.ETHAddr))
		}
	}
}

func CheckProcessingGasReclaims() {
	gasReclaimPending, err := models.GetRowsByGasStatus(models.GasTransferLeftoversReclaimProcessing)
	if err != nil {
		fmt.Println("Cannot get completed_uploads with pending gas transfers: " + err.Error())
		// already captured error in upstream function
		return
	}

	for _, pending := range gasReclaimPending {
		ethBalance := EthWrapper.CheckETHBalance(services.StringToAddress(pending.ETHAddr))
		if ethBalance.Int64() > 0 {
			gasNeededToReclaimETH, err := EthWrapper.CalculateGasToSend(services.GasLimitETHSend)
			if err != nil {
				fmt.Println("Could not calculate gas needed to retrieve ETH from " + pending.ETHAddr +
					" in CheckProcessingGasReclaims() in claim_unused_prls")
				continue
			}
			if gasNeededToReclaimETH.Int64() > ethBalance.Int64() {
				fmt.Println("Not enough ETH to retrieve leftover ETH from " + pending.ETHAddr +
					" in CheckProcessingGasReclaims() in claim_unused_prls, setting to success")
				// won't be able to reclaim whatever is left, just set to success
				pending.GasStatus = models.GasTransferLeftoversReclaimSuccess
				models.DB.ValidateAndUpdate(&pending)
			}
		} else {
			fmt.Println("Finished reclaiming gas from  " + pending.ETHAddr + " in CheckProcessingGasReclaims " +
				"in claim_unused_prls")
			pending.GasStatus = models.GasTransferLeftoversReclaimSuccess
			models.DB.ValidateAndUpdate(&pending)
		}
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

// for leftover gas reclaims that are still processing by the time of the threshold
func ResendTimedOutGasReclaims(thresholdTime time.Time) {
	timedOutGasReclaims, err := models.GetTimedOutGasReclaims(thresholdTime)
	if err != nil {
		oyster_utils.LogIfError(fmt.Errorf("Error getting timed out gas reclaims: %v", err), nil)
		return
	}
	for _, reclaim := range timedOutGasReclaims {
		// reset it back to a prior state so we will try again
		reclaim.GasStatus = models.GasTransferSuccess
		models.DB.ValidateAndUpdate(&reclaim)
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

// for claims whose gas transfers succeeded but there is still unclaimed PRL
func RetrieveLeftoverETH() {
	completedClaims, err := models.GetRowsByGasAndPRLStatus(models.GasTransferSuccess, models.PRLClaimSuccess)
	if err != nil {
		oyster_utils.LogIfError(fmt.Errorf("Error getting completed claims: %v", err), nil)
		return
	}
	for _, completedClaim := range completedClaims {
		ethBalance := EthWrapper.CheckETHBalance(services.StringToAddress(completedClaim.ETHAddr))
		if ethBalance.Int64() > 0 {
			gasNeededToReclaimETH, err := EthWrapper.CalculateGasToSend(services.GasLimitETHSend)
			if err != nil {
				fmt.Println("Could not calculate gas needed to retrieve ETH from " + completedClaim.ETHAddr +
					" in RetrieveLeftoverETH() in claim_unused_prls")
				continue
			}
			if gasNeededToReclaimETH.Int64() > ethBalance.Int64() {
				fmt.Println("Not enough ETH to retrieve leftover ETH from " + completedClaim.ETHAddr +
					" in RetrieveLeftoverETH() in claim_unused_prls, setting to success")
				// won't be able to reclaim whatever is left, just set to success
				completedClaim.GasStatus = models.GasTransferLeftoversReclaimSuccess
				models.DB.ValidateAndUpdate(&completedClaim)
				continue
			}

			gasToReclaim := new(big.Int).Sub(ethBalance, gasNeededToReclaimETH)

			privateKey, err := services.StringToPrivateKey(completedClaim.DecryptSessionEthKey())

			_, _, _, err = EthWrapper.SendETH(
				services.StringToAddress(completedClaim.ETHAddr),
				privateKey,
				services.MainWalletAddress,
				gasToReclaim)
			if err != nil {
				fmt.Println("Could not reclaim leftover ETH from " + completedClaim.ETHAddr +
					" in RetrieveLeftoverETH in claim_unused_prls")
			} else {
				fmt.Println("Reclaiming leftover ETH from " + completedClaim.ETHAddr + " in RetrieveLeftoverETH " +
					"in claim_unused_prls")
				completedClaim.GasStatus = models.GasTransferLeftoversReclaimProcessing
				models.DB.ValidateAndUpdate(&completedClaim)
			}
		} else {
			fmt.Println("No extra gas to reclaim from " + completedClaim.ETHAddr + " in RetrieveLeftoverETH " +
				"in claim_unused_prls")
			completedClaim.GasStatus = models.GasTransferLeftoversReclaimSuccess
			models.DB.ValidateAndUpdate(&completedClaim)
		}
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
		_, txHash, nonce, err := EthWrapper.SendETH(
			services.MainWalletAddress,
			services.MainWalletPrivateKey,
			services.StringToAddress(upload.ETHAddr),
			gasToSend)
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			continue
		}
		fmt.Println("InitiateGasTransfer processing to " + upload.ETHAddr + " claim_unused_prls")
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
			continue
		}
		callMsg, err := EthWrapper.CreateSendPRLMessage(services.StringToAddress(upload.ETHAddr),
			ecdsaPrivateKey, services.MainWalletAddress, *balance)

		if err != nil {
			oyster_utils.LogIfError(err, nil)
			continue
		}
		sendSuccess, txHash, nonce := EthWrapper.SendPRLFromOyster(callMsg)
		if sendSuccess {
			fmt.Println("InitiatePRLClaim processing from " + upload.ETHAddr + " claim_unused_prls")
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

// purge claims whose GasStatus is GasTransferLeftoversReclaimSuccess
func PurgeCompletedClaims() {
	models.DeleteCompletedClaims()
}

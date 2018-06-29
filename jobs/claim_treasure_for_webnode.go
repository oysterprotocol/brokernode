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

func ClaimTreasureForWebnode(thresholdTime time.Time, PrometheusWrapper services.PrometheusService) {

	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramClaimTreasureForWebnode, start)

	if oyster_utils.BrokerMode == oyster_utils.ProdMode {

		CheckOngoingGasTransactions()
		CheckOngoingPRLClaims()
		CheckOngoingGasReclaims()

		ResendOldETHTransfers(thresholdTime)
		ResendOldPRLClaims(thresholdTime)
		ResendOldGasReclaims(thresholdTime)

		ResendErroredETHTransfers()
		ResendErroredPRLClaims()

		SendGasForNewTreasureClaims()
		StartNewTreasureClaims()
		RetrieveLeftoverETHFromTreasureClaiming()

		PurgeCompletedTreasureClaims()
	}
}

func CheckOngoingGasTransactions() {
	gasPending, err := models.GetTreasureClaimsByGasStatus(models.GasTransferProcessing)
	if err != nil {
		fmt.Println("Cannot get webnode_treasure_claims with pending gas transfers: " + err.Error())
		// already captured error in upstream function
		return
	}

	if len(gasPending) <= 0 {
		return
	}
	gasToProcessTransaction, err := EthWrapper.CalculateGasToSend(services.GasLimitPRLClaim)
	if err != nil {
		fmt.Println("Cannot calculate gas to send to webnode_treasure_claims with pending " +
			"gas transfers: " + err.Error())
		// already captured error in upstream function
		return
	}

	for _, pending := range gasPending {
		ethBalance := EthWrapper.CheckETHBalance(services.StringToAddress(pending.TreasureETHAddr))
		if ethBalance.Int64() >= gasToProcessTransaction.Int64() {
			fmt.Println("ETH (gas) transaction sent to " + pending.TreasureETHAddr + " in CheckOngoingGasTransactions() " +
				"in claim_treasure_for_webnode")
			pending.GasStatus = models.GasTransferSuccess
			vErr, err := models.DB.ValidateAndUpdate(&pending)
			if err != nil {
				oyster_utils.LogIfError(err, nil)
				continue
			}
			if len(vErr.Errors) > 0 {
				errString := "validation errors in claim_treasure_for_webnode in CheckOngoingGasTransactions: " + fmt.Sprint(vErr.Errors)
				err = errors.New(errString)
				oyster_utils.LogIfError(err, nil)
				continue
			}
			oyster_utils.LogToSegment("claim_treasure_for_webnode: CheckOngoingGasTransactions", analytics.NewProperties().
				Set("new_status", models.GasTransferStatusMap[pending.GasStatus]).
				Set("eth_address_to", pending.TreasureETHAddr))
		}
	}
}

func CheckOngoingPRLClaims() {
	prlsPending, err := models.GetTreasureClaimsByPRLStatus(models.PRLClaimProcessing)
	if err != nil {
		fmt.Println("Cannot get webnode_treasure_claims with pending PRL retrieval: " + err.Error())
		// already captured error in upstream function
		return
	}

	for _, pending := range prlsPending {

		claimClock, err := EthWrapper.CheckClaimClock(services.StringToAddress(pending.TreasureETHAddr))
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			continue
		}

		if claimClock.Int64() != pending.StartingClaimClock && pending.StartingClaimClock != int64(-1) {
			fmt.Println("PRLs claimed from " + pending.TreasureETHAddr + " to " + pending.ReceiverETHAddr + " in CheckOngoingPRLClaims() " +
				"in claim_treasure_for_webnode")
			pending.ClaimPRLStatus = models.PRLClaimSuccess
			vErr, err := models.DB.ValidateAndUpdate(&pending)
			if err != nil {
				oyster_utils.LogIfError(err, nil)
				continue
			}
			if len(vErr.Errors) > 0 {
				errString := "validation errors in claim_treasure_for_webnode in CheckOngoingPRLClaims: " +
					fmt.Sprint(vErr.Errors)
				err = errors.New(errString)
				oyster_utils.LogIfError(err, nil)
				continue
			}
			oyster_utils.LogToSegment("claim_treasure_for_webnode: CheckOngoingPRLClaims", analytics.NewProperties().
				Set("new_status", models.PRLClaimStatusMap[pending.ClaimPRLStatus]).
				Set("eth_address_from", pending.TreasureETHAddr))
		}
	}
}

func CheckOngoingGasReclaims() {
	gasReclaimPending, err := models.GetTreasureClaimsByGasStatus(models.GasTransferLeftoversReclaimProcessing)
	if err != nil {
		fmt.Println("Cannot get webnode_treasure_claims with pending gas transfers: " + err.Error())
		// already captured error in upstream function
		return
	}

	for _, pending := range gasReclaimPending {
		ethBalance := EthWrapper.CheckETHBalance(services.StringToAddress(pending.TreasureETHAddr))
		if ethBalance.Int64() > 0 {
			gasNeededToReclaimETH, err := EthWrapper.CalculateGasToSend(services.GasLimitETHSend)
			if err != nil {
				fmt.Println("Could not calculate gas needed to retrieve ETH from " + pending.TreasureETHAddr +
					" in CheckProcessingGasReclaims() in claim_treasure_for_webnode")
				continue
			}
			if gasNeededToReclaimETH.Int64() > ethBalance.Int64() {
				fmt.Println("Not enough ETH to retrieve leftover ETH from " + pending.TreasureETHAddr +
					" in CheckOngoingGasReclaims() in claim_treasure_for_webnode, setting to success")
				// won't be able to reclaim whatever is left, just set to success
				pending.GasStatus = models.GasTransferLeftoversReclaimSuccess
				models.DB.ValidateAndUpdate(&pending)
			}
		} else {
			fmt.Println("Finished reclaiming gas from  " + pending.TreasureETHAddr + " in CheckOngoingGasReclaims() " +
				"in claim_treasure_for_webnode")
			pending.GasStatus = models.GasTransferLeftoversReclaimSuccess
			models.DB.ValidateAndUpdate(&pending)
		}
	}
}

// for gas transfers that are still processing by the time of the threshold
func ResendOldETHTransfers(thresholdTime time.Time) {
	oldGasTransfers, err := models.GetTreasureClaimsWithTimedOutGasTransfers(thresholdTime)

	if err != nil {
		oyster_utils.LogIfError(fmt.Errorf("Error getting timed out gas transfers: %v", err), nil)
		return
	}
	if len(oldGasTransfers) > 0 {

		for _, transfer := range oldGasTransfers {
			oyster_utils.LogToSegment("claim_treasure_for_webnode: gas_transfer_timed_out", analytics.NewProperties().
				Set("eth_address_to", transfer.TreasureETHAddr).
				Set("genesis_hash", transfer.GenesisHash))
		}

		SendGas(oldGasTransfers)
	}
}

// for prl transfers that are still processing by the time of the threshold
func ResendOldPRLClaims(thresholdTime time.Time) {
	oldPRLClaims, err := models.GetTreasureClaimsWithTimedOutPRLClaims(thresholdTime)
	if err != nil {
		oyster_utils.LogIfError(fmt.Errorf("Error getting timed out gas transfers: %v", err), nil)
		return
	}
	if len(oldPRLClaims) > 0 {
		for _, claim := range oldPRLClaims {
			oyster_utils.LogToSegment("claim_treasure_for_webnode: unclaimed_prl_transfer_timed_out", analytics.NewProperties().
				Set("eth_address_from", claim.TreasureETHAddr).
				Set("genesis_hash", claim.GenesisHash))
		}

		ClaimPRL(oldPRLClaims)
	}
}

// for leftover gas reclaims that are still processing by the time of the threshold
func ResendOldGasReclaims(thresholdTime time.Time) {
	oldGasReclaims, err := models.GetTreasureClaimsWithTimedOutGasReclaims(thresholdTime)
	if err != nil {
		oyster_utils.LogIfError(fmt.Errorf("Error getting timed out gas reclaims: %v", err), nil)
		return
	}
	for _, reclaim := range oldGasReclaims {
		// reset it back to a prior state so we will try again
		reclaim.GasStatus = models.GasTransferSuccess
		models.DB.ValidateAndUpdate(&reclaim)
	}
}

// for gas transfers that are in an error state
func ResendErroredETHTransfers() {
	gasTransferErrors, err := models.GetTreasureClaimsByGasStatus(models.GasTransferError)
	if err != nil {
		oyster_utils.LogIfError(fmt.Errorf("Error getting webnode treasure claims whose gas transfers errored: %v", err), nil)
		return
	}
	if len(gasTransferErrors) > 0 {

		for _, transfer := range gasTransferErrors {
			oyster_utils.LogToSegment("claim_treasure_for_webnode: gas_transfer_error", analytics.NewProperties().
				Set("eth_address_to", transfer.TreasureETHAddr).
				Set("genesis_hash", transfer.GenesisHash))
		}

		SendGas(gasTransferErrors)
	}
}

// for prl transfers that are in an error state
func ResendErroredPRLClaims() {
	prlTransferErrors, err := models.GetTreasureClaimsByPRLStatus(models.PRLClaimError)
	if err != nil {
		oyster_utils.LogIfError(fmt.Errorf("Error getting webnode treasure claims whose PRL transfers errored: %v", err), nil)
		return
	}
	if len(prlTransferErrors) > 0 {

		for _, transfer := range prlTransferErrors {
			oyster_utils.LogToSegment("claim_treasure_for_webnode: prl_claim_error", analytics.NewProperties().
				Set("eth_address_to", transfer.ReceiverETHAddr).
				Set("eth_address_from", transfer.TreasureETHAddr).
				Set("genesis_hash", transfer.GenesisHash))
		}

		ClaimPRL(prlTransferErrors)
	}
}

// for new claims with no gas
func SendGasForNewTreasureClaims() {
	needGas, err := models.GetTreasureClaimsByGasStatus(models.GasTransferNotStarted)
	if err != nil {
		oyster_utils.LogIfError(fmt.Errorf("Error getting webnode treasure claims whose addresses need gas: %v", err), nil)
		return
	}
	if len(needGas) > 0 {

		for _, transfer := range needGas {
			oyster_utils.LogToSegment("claim_treasure_for_webnode: send_gas_for_new_treasure_claim", analytics.NewProperties().
				Set("eth_address_to", transfer.TreasureETHAddr).
				Set("genesis_hash", transfer.GenesisHash))
		}

		SendGas(needGas)
	}
}

// for claims whose gas transfers succeeded but there is still unclaimed PRL
func StartNewTreasureClaims() {
	readyClaims, err := models.GetTreasureClaimsByGasAndPRLStatus(models.GasTransferSuccess, models.PRLClaimNotStarted)
	if err != nil {
		oyster_utils.LogIfError(fmt.Errorf("Error getting ready claims: %v", err), nil)
		return
	}
	if len(readyClaims) > 0 {

		for _, transfer := range readyClaims {

			oyster_utils.LogToSegment("claim_treasure_for_webnode: new_treasure_claim", analytics.NewProperties().
				Set("eth_address_from", transfer.TreasureETHAddr).
				Set("genesis_hash", transfer.GenesisHash))
		}

		ClaimPRL(readyClaims)
	}
}

// for claims whose gas transfers succeeded but there is still unclaimed PRL
func RetrieveLeftoverETHFromTreasureClaiming() {
	completedClaims, err := models.GetTreasureClaimsByGasAndPRLStatus(models.GasTransferSuccess, models.PRLClaimSuccess)
	if err != nil {
		oyster_utils.LogIfError(fmt.Errorf("Error getting completed claims: %v", err), nil)
		return
	}
	for _, completedClaim := range completedClaims {
		ethBalance := EthWrapper.CheckETHBalance(services.StringToAddress(completedClaim.TreasureETHAddr))
		if ethBalance.Int64() > 0 {
			gasNeededToReclaimETH, err := EthWrapper.CalculateGasToSend(services.GasLimitETHSend)
			if err != nil {
				fmt.Println("Could not calculate gas needed to retrieve ETH from " + completedClaim.TreasureETHAddr +
					" in RetrieveLeftoverETHFromTreasureClaiming() in claim_treasure_for_webnode")
				continue
			}
			if gasNeededToReclaimETH.Int64() > ethBalance.Int64() {
				fmt.Println("Not enough ETH to retrieve leftover ETH from " + completedClaim.TreasureETHAddr +
					" in RetrieveLeftoverETHFromTreasureClaiming() in claim_treasure_for_webnode, setting to success")
				// won't be able to reclaim whatever is left, just set to success
				completedClaim.GasStatus = models.GasTransferLeftoversReclaimSuccess
				models.DB.ValidateAndUpdate(&completedClaim)
				continue
			}

			gasToReclaim := new(big.Int).Sub(ethBalance, gasNeededToReclaimETH)

			privateKey, err := services.StringToPrivateKey(completedClaim.DecryptTreasureEthKey())

			_, _, _, err = EthWrapper.SendETH(
				services.StringToAddress(completedClaim.TreasureETHAddr),
				privateKey,
				services.MainWalletAddress,
				gasToReclaim)
			if err != nil {
				fmt.Println("Could not reclaim leftover ETH from " + completedClaim.TreasureETHAddr +
					" in RetrieveLeftoverETH in claim_treasure_for_webnode")
			} else {
				fmt.Println("Reclaiming leftover ETH from " + completedClaim.TreasureETHAddr + " in RetrieveLeftoverETHFromTreasureClaiming " +
					"in claim_treasure_for_webnode")
				completedClaim.GasStatus = models.GasTransferLeftoversReclaimProcessing
				models.DB.ValidateAndUpdate(&completedClaim)
			}
		} else {
			fmt.Println("No extra gas to reclaim from " + completedClaim.TreasureETHAddr + " in RetrieveLeftoverETHFromTreasureClaiming " +
				"in claim_treasure_for_webnode")
			completedClaim.GasStatus = models.GasTransferLeftoversReclaimSuccess
			models.DB.ValidateAndUpdate(&completedClaim)
		}
	}
}

// wraps call to eth_gatway's SendETH method and sets GasStatus to GasTransferProcessing
func SendGas(treasuresThatNeedGas []models.WebnodeTreasureClaim) {
	gasToClaim, err := EthWrapper.CalculateGasToSend(services.GasLimitPRLClaim)
	if err != nil {
		oyster_utils.LogIfError(fmt.Errorf("Error determining gas to send: %v", err), nil)
		return
	}
	for _, treasureClaim := range treasuresThatNeedGas {

		ethBalance := EthWrapper.CheckETHBalance(services.StringToAddress(treasureClaim.TreasureETHAddr))
		if ethBalance.Int64() > gasToClaim.Int64() {
			// already have enough gas
			treasureClaim.GasStatus = models.GasTransferSuccess
			models.DB.ValidateAndUpdate(&treasureClaim)
			continue
		}

		gasNeeded := new(big.Int).Sub(gasToClaim, ethBalance)

		_, txHash, nonce, err := EthWrapper.SendETH(
			services.MainWalletAddress,
			services.MainWalletPrivateKey,
			services.StringToAddress(treasureClaim.TreasureETHAddr),
			gasNeeded)
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			continue
		}
		fmt.Println("SendGas processing to " + treasureClaim.TreasureETHAddr + " claim_treasure_for_webnode")
		treasureClaim.GasStatus = models.GasTransferProcessing
		treasureClaim.GasTxHash = txHash
		treasureClaim.GasTxNonce = nonce
		models.DB.ValidateAndUpdate(&treasureClaim)
	}
}

// wraps calls eth_gatway's ClaimPRLs method and sets PRLStatus to PRLClaimProcessing
func ClaimPRL(treasuresWithPRLsToBeClaimed []models.WebnodeTreasureClaim) {
	for _, treasureClaim := range treasuresWithPRLsToBeClaimed {

		if treasureClaim.StartingClaimClock == -1 {

			// set the claim clock so we can check later that we were successful
			claimClock, err := EthWrapper.CheckClaimClock(services.StringToAddress(treasureClaim.TreasureETHAddr))
			if err != nil {
				oyster_utils.LogIfError(err, nil)
				continue
			}

			treasureClaim.StartingClaimClock = claimClock.Int64()
			vErr, err := models.DB.ValidateAndUpdate(&treasureClaim)

			if len(vErr.Errors) > 0 {
				oyster_utils.LogIfError(errors.New(vErr.Error()), nil)
				continue
			}
			if err != nil {
				oyster_utils.LogIfError(err, nil)
				continue
			}
		}

		balance := EthWrapper.CheckPRLBalance(services.StringToAddress(treasureClaim.TreasureETHAddr))

		if balance.Int64() == 0 {
			// no claimable treasure
			err := errors.New("expected treasure but got balance of 0 at " + treasureClaim.TreasureETHAddr)
			oyster_utils.LogIfError(err, nil)
			treasureClaim.ClaimPRLStatus = models.PRLClaimSuccess
			models.DB.ValidateAndUpdate(&treasureClaim)
			continue
		}

		privateKey := treasureClaim.DecryptTreasureEthKey()

		ecdsaPrivateKey, err := services.StringToPrivateKey(privateKey)

		if err != nil {
			oyster_utils.LogIfError(err, nil)
			continue
		}

		claimSuccess := services.EthWrapper.ClaimPRL(
			services.StringToAddress(treasureClaim.ReceiverETHAddr),
			services.StringToAddress(treasureClaim.TreasureETHAddr),
			ecdsaPrivateKey)

		if claimSuccess {
			fmt.Println("ClaimPRL processing from " + treasureClaim.TreasureETHAddr + " claim_treasure_for_webnode")
			treasureClaim.ClaimPRLStatus = models.PRLClaimProcessing
			models.DB.ValidateAndUpdate(&treasureClaim)
		} else {
			err := errors.New("error claiming prls from treasure addresss: " + treasureClaim.TreasureETHAddr)
			oyster_utils.LogIfError(err, nil)
		}
	}
}

// purge claims whose GasStatus is GasTransferLeftoversReclaimSuccess
func PurgeCompletedTreasureClaims() {
	err := models.DeleteCompletedTreasureClaims()
	if err != nil {
		oyster_utils.LogIfError(fmt.Errorf("Error purging completed treasure claims: %v", err), nil)
		return
	}
}

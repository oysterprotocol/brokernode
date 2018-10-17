package jobs

import (
	"errors"
	"fmt"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"github.com/oysterprotocol/brokernode/utils/eth_gateway"
	"gopkg.in/segmentio/analytics-go.v3"
	"time"
)

func BuryTreasureAddresses(thresholdTime time.Time, PrometheusWrapper services.PrometheusService) {

	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramBuryTreasureAddresses, start)

	if oyster_utils.BrokerMode == oyster_utils.ProdMode &&
		oyster_utils.PaymentMode == oyster_utils.UserIsPaying {

		CheckPRLTransactions()
		CheckGasTransactions()
		CheckBuryTransactions()

		// TODO: we may want to re-use the same nonce when a transaction
		// is considered timed-out, so we can replace the existing transaction
		// with more gas
		SetTimedOutTransactionsToError(thresholdTime)
		StageTransactionsWithErrorsForRetry()

		SendPRLsToWaitingTreasureAddresses()
		SendGasToTreasureAddresses()
		InvokeBury()
		CheckForReclaimableGas(thresholdTime.Add(-1 * time.Since(thresholdTime)))
		/* The network could be extremely congested and if we wait a while, it may be worth reclaiming gas later.
		So passing in a time for us to wait before we give up on attempting to reclaim gas if we have not already
		started the reclaim attempt*/

		/* TODO:  Enable this method after we have sent
		the PRLs to all the treasure addresses.  For now keep disabled so we don't lose them.
		PurgeFinishedTreasure()
		*/
	}
}

func CheckPRLTransactions() {
	prlsPending, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.PRLPending})
	if err != nil {
		fmt.Println("Cannot get treasures with prls pending in bury_treasure_addresses: " + err.Error())
		// already captured error in upstream function
		return
	}

	for _, pending := range prlsPending {
		prlBalance := EthWrapper.CheckPRLBalance(eth_gateway.StringToAddress(pending.ETHAddr))
		expectedPRLBalance := pending.GetPRLAmount()
		if prlBalance.Int64() > 0 && prlBalance.String() == expectedPRLBalance.String() ||
			prlBalance.Int64() >= expectedPRLBalance.Int64() {
			fmt.Println("PRL transaction confirmed in CheckPRLTransactions()")
			pending.PRLStatus = models.PRLConfirmed
			vErr, err := models.DB.ValidateAndUpdate(&pending)
			if err != nil {
				oyster_utils.LogIfError(err, nil)
				continue
			}
			if len(vErr.Errors) > 0 {
				oyster_utils.LogIfValidationError(
					"validation errors in bury_treasure_addresses in CheckPRLTransactions.", vErr, nil)
				continue
			}
			oyster_utils.LogToSegment("bury_treasure_addresses: CheckPRLTransactions", analytics.NewProperties().
				Set("new_status", models.PRLStatusMap[pending.PRLStatus]).
				Set("eth_address", pending.ETHAddr))
		}
	}
}

func CheckGasTransactions() {
	gasPending, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.GasPending})
	if err != nil {
		fmt.Println("Cannot get treasures with gas pending in bury_treasure_addresses: " + err.Error())
		// already captured error in upstream function
		return
	}

	for _, pending := range gasPending {
		ethBalance := EthWrapper.CheckETHBalance(eth_gateway.StringToAddress(pending.ETHAddr))
		if ethBalance.Int64() > 0 {
			fmt.Println("ETH (gas) transaction confirmed in CheckGasTransactions()")
			pending.PRLStatus = models.GasConfirmed
			vErr, err := models.DB.ValidateAndUpdate(&pending)
			if err != nil {
				oyster_utils.LogIfError(err, nil)
				continue
			}
			if len(vErr.Errors) > 0 {
				oyster_utils.LogIfValidationError(
					"validation errors in bury_treasure_addresses in CheckGasTransactions.", vErr, nil)
				continue
			}
			oyster_utils.LogToSegment("bury_treasure_addresses: CheckGasTransactions", analytics.NewProperties().
				Set("new_status", models.PRLStatusMap[pending.PRLStatus]).
				Set("eth_address", pending.ETHAddr))
		}
	}
}

func CheckBuryTransactions() {
	buryPending, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.BuryPending})
	if err != nil {
		fmt.Println("Cannot get treasures with burial pending in bury_treasure_addresses: " + err.Error())
		// already captured error in upstream function
		return
	}

	for _, pending := range buryPending {
		buried, err := EthWrapper.CheckBuriedState(eth_gateway.StringToAddress(pending.ETHAddr))
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			continue
		} else if buried {
			err := models.SetToTreasureBuriedByGenesisHash(pending.GenesisHash)
			if err != nil {
				oyster_utils.LogIfError(err, nil)
				continue
			}
			fmt.Println("Bury transaction confirmed in CheckBuryTransactions()")
			pending.PRLStatus = models.BuryConfirmed
			vErr, err := models.DB.ValidateAndUpdate(&pending)
			if err != nil {
				oyster_utils.LogIfError(err, nil)
				continue
			}
			if len(vErr.Errors) > 0 {
				oyster_utils.LogIfValidationError(
					"validation errors in bury_treasure_addresses in CheckBuryTransactions.", vErr, nil)
				continue
			}
			oyster_utils.LogToSegment("bury_treasure_addresses: CheckBuryTransactions", analytics.NewProperties().
				Set("new_status", models.PRLStatusMap[pending.PRLStatus]).
				Set("eth_address", pending.ETHAddr))
		}
	}
}

/*CheckForReclaimableGas - after the treasure has been buried, this method will check if there is
leftover eth at the treasure address and if it's enough to try to reclaim it.  If so it will
start the reclaim.  This method will also check existing gas reclaims--if the gas is no longer worth
reclaiming (because it has succeeded, or network congestion makes it impractical) it will set it
to success.*/
func CheckForReclaimableGas(thresholdTime time.Time) {
	reclaimableAddresses, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{
		models.GasReclaimPending,
		models.BuryConfirmed})
	if err != nil {
		fmt.Println("Cannot get treasures with gas reclaim pending or burials confirmed " +
			"in bury_treasure_addresses: " + err.Error())
		// already captured error in upstream function
		return
	}

	for _, reclaimable := range reclaimableAddresses {
		worthReclaimingGas, gasToReclaim, err := EthWrapper.CheckIfWorthReclaimingGas(
			eth_gateway.StringToAddress(reclaimable.ETHAddr), eth_gateway.GasLimitETHSend)
		if err != nil {
			fmt.Println("Error determining if it's worth it to retrieve leftover ETH from " +
				reclaimable.ETHAddr +
				" in CheckForReclaimableGas() in bury_treasure_addresses.")
			continue
		} else if !worthReclaimingGas {
			if reclaimable.UpdatedAt.Before(thresholdTime) {
				fmt.Println("Not enough ETH to retrieve leftover ETH from " + reclaimable.ETHAddr +
					" in CheckForReclaimableGas() in bury_treasure_addresses, setting to success")
				/* won't be able to reclaim whatever is left, just set to success */
				reclaimable.PRLStatus = models.GasReclaimConfirmed
				models.DB.ValidateAndUpdate(&reclaimable)
			} else {
				fmt.Println("Not enough ETH to retrieve leftover ETH from " + reclaimable.ETHAddr +
					" in CheckForReclaimableGas() in bury_treasure_addresses, wait for network congestion to decrease")
			}
			continue
		}

		if reclaimable.PRLStatus == models.GasReclaimPending {
			/* gas reclaim is still in progress, do not send again */
			continue
		}

		privateKey, err := eth_gateway.StringToPrivateKey(reclaimable.DecryptTreasureEthKey())

		reclaimingSuccess := EthWrapper.ReclaimGas(eth_gateway.StringToAddress(reclaimable.ETHAddr),
			privateKey, gasToReclaim)

		if reclaimingSuccess {
			reclaimable.PRLStatus = models.GasReclaimPending
		} else {
			reclaimable.PRLStatus = models.GasReclaimError
		}
		models.DB.ValidateAndUpdate(&reclaimable)
	}
}

func SetTimedOutTransactionsToError(thresholdTime time.Time) {
	timedOutTransactions, err := models.GetTreasuresToBuryByPRLStatusAndUpdateTime([]models.PRLStatus{
		models.PRLPending,
		models.GasPending,
		models.BuryPending,
		models.GasReclaimPending,
	}, thresholdTime)

	if err != nil {
		fmt.Println("Cannot get timed out treasures in bury_treasure_addresses: " + err.Error())
		// already captured error in upstream function
		return
	}

	for _, timedOutTransaction := range timedOutTransactions {
		oldStatus := timedOutTransaction.PRLStatus
		timedOutTransaction.PRLStatus = models.PRLStatus(int((timedOutTransaction.PRLStatus)-1) * -1)
		vErr, err := models.DB.ValidateAndUpdate(&timedOutTransaction)
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			continue
		}
		if len(vErr.Errors) > 0 {
			oyster_utils.LogIfValidationError(
				"validation errors in bury_treasure_addresses in SetTimedOutTransactionsToError.", vErr, nil)
			continue
		}
		oyster_utils.LogToSegment("bury_treasure_addresses: SetTimedOutTransactionsToError", analytics.NewProperties().
			Set("old_status", models.PRLStatusMap[oldStatus]).
			Set("new_status", models.PRLStatusMap[timedOutTransaction.PRLStatus]).
			Set("eth_address", timedOutTransaction.ETHAddr))
	}
}

func StageTransactionsWithErrorsForRetry() {
	erroredTransactions, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{
		models.PRLError,
		models.GasError,
		models.BuryError,
		models.GasReclaimError,
	})
	if err != nil {
		fmt.Println("Cannot get error'd treasures in bury_treasure_addresses: " + err.Error())
		// already captured error in upstream function
		return
	}

	for _, errorTransaction := range erroredTransactions {
		errorTransaction.PRLStatus = models.PRLStatus(int(errorTransaction.PRLStatus) * -1)
		vErr, err := models.DB.ValidateAndUpdate(&errorTransaction)
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			continue
		}
		if len(vErr.Errors) > 0 {
			oyster_utils.LogIfValidationError(
				"validation errors in bury_treasure_addresses in StageTransactionsWithErrorsForRetry.", vErr, nil)
			continue
		}
	}

	if len(erroredTransactions) > 0 {
		oyster_utils.LogToSegment("bury_treasure_addresses: StageTransactionsWithErrorsForRetry", analytics.NewProperties().
			Set("num_error'd_transactions", fmt.Sprint(len(erroredTransactions))))
	}
}

func SendPRLsToWaitingTreasureAddresses() {

	waitingForPRLS, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.PRLWaiting})
	if err != nil {
		fmt.Println("Cannot get treasures awaiting PRLs in bury_treasure_addresses: " + err.Error())
		// already captured error in upstream function
		return
	}

	for _, waitingAddress := range waitingForPRLS {
		sendPRL(waitingAddress)
	}
}

func SendGasToTreasureAddresses() {
	waitingForGas, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.PRLConfirmed})
	if err != nil {
		fmt.Println("Cannot get treasures awaiting gas in bury_treasure_addresses: " + err.Error())
		// already captured error in upstream function
		return
	}

	for _, waitingAddress := range waitingForGas {
		sendGas(waitingAddress)
	}
}

func InvokeBury() {
	readyToInvokeBury, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.GasConfirmed})
	if err != nil {
		fmt.Println("Cannot get treasures awaiting bury() in bury_treasure_addresses: " + err.Error())
		// already captured error in upstream function
		return
	}

	for _, buryAddress := range readyToInvokeBury {
		buryPRL(buryAddress)
	}
}

func PurgeFinishedTreasure() {
	completeTreasures, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.GasReclaimConfirmed})
	if err != nil {
		fmt.Println("Cannot get completed treasures in bury_treasure_addresses: " + err.Error())
		// already captured error in upstream function
		return
	}

	for _, completeTreasure := range completeTreasures {
		ethAddr := completeTreasure.ETHAddr

		genesisHashExists, genesisHashIsBuried, err :=
			models.CheckIfGenesisHashExistsAndIsBuried(completeTreasure.GenesisHash)
		oyster_utils.LogIfError(err, nil)
		if genesisHashExists && !genesisHashIsBuried {
			err := models.SetToTreasureBuriedByGenesisHash(completeTreasure.GenesisHash)
			if err != nil {
				oyster_utils.LogIfError(err, nil)
				continue
			}
		} else if !genesisHashExists {
			// skip until another process creates the genesis hash row
			continue
		}

		err = models.DB.Destroy(&completeTreasure)
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			continue
		}
		oyster_utils.LogToSegment("bury_treasure_addresses: PurgeFinishedTreasure",
			analytics.NewProperties().
				Set("eth_address", ethAddr))
	}
}

func sendPRL(treasureToBury models.Treasure) {

	balance := EthWrapper.CheckPRLBalance(eth_gateway.MainWalletAddress)
	if balance.Int64() <= 0 || balance.Int64() < treasureToBury.GetPRLAmount().Int64() {
		errorString := "Cannot send PRL to treasure address due to insufficient balance in wallet.  balance: " +
			fmt.Sprint(balance.Int64()) + "; amount_to_send: " + fmt.Sprint(treasureToBury.GetPRLAmount().Int64())
		err := errors.New(errorString)
		oyster_utils.LogIfError(err, nil)
		return
	}

	amount := *treasureToBury.GetPRLAmount()

	callMsg, _ := EthWrapper.CreateSendPRLMessage(eth_gateway.MainWalletAddress,
		eth_gateway.MainWalletPrivateKey,
		eth_gateway.StringToAddress(treasureToBury.ETHAddr), amount)

	sendSuccess, txHash, nonce := EthWrapper.SendPRLFromOyster(callMsg)
	if !sendSuccess {
		errorString := "\nFailure sending " + fmt.Sprint(treasureToBury.GetPRLAmount().Int64()) + " PRL to " +
			treasureToBury.ETHAddr
		err := errors.New(errorString)
		oyster_utils.LogIfError(err, nil)
		treasureToBury.PRLStatus = models.PRLError
		// TODO add method for logging vErrs
		//vErr, err := models.DB.ValidateAndUpdate(&treasureToBury)
		models.DB.ValidateAndUpdate(&treasureToBury)
	} else {
		treasureToBury.PRLStatus = models.PRLPending
		treasureToBury.PRLTxHash = txHash
		treasureToBury.PRLTxNonce = nonce
		vErr, err := models.DB.ValidateAndUpdate(&treasureToBury)
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			return
		}
		if len(vErr.Errors) > 0 {
			oyster_utils.LogIfValidationError(
				"validation errors in bury_treasure_addresses in sendPRL.", vErr, nil)
			return
		}
		oyster_utils.LogToSegment("bury_treasure_addresses: sendPRL", analytics.NewProperties().
			Set("new_status", models.PRLStatusMap[treasureToBury.PRLStatus]).
			Set("eth_address", treasureToBury.ETHAddr))
		//go waitForPRL(treasureToBury)
	}
}

func sendGas(treasureToBury models.Treasure) {

	gasToSend, err := EthWrapper.CalculateGasNeeded(eth_gateway.GasLimitPRLBury)
	if err != nil {
		fmt.Println("Cannot send Gas to treasure address: " + err.Error())
		// already captured error in upstream function
		return
	}

	balance := EthWrapper.CheckETHBalance(eth_gateway.MainWalletAddress)
	if balance.Int64() < gasToSend.Int64() {
		errorString := "Cannot send Gas to treasure address due to insufficient balance in wallet.  balance: " +
			fmt.Sprint(balance.Int64()) + "; amount_to_send: " + fmt.Sprint(gasToSend.Int64())
		err := errors.New(errorString)
		oyster_utils.LogIfError(err, nil)
		return
	}

	_, txHash, nonce, err := EthWrapper.SendETH(eth_gateway.MainWalletAddress, eth_gateway.MainWalletPrivateKey, eth_gateway.StringToAddress(treasureToBury.ETHAddr), gasToSend)
	if err != nil {
		errorString := "\nFailure sending " + fmt.Sprint(gasToSend.Int64()) + " Gas to " + treasureToBury.ETHAddr
		err := errors.New(errorString)
		oyster_utils.LogIfError(err, nil)
		treasureToBury.PRLStatus = models.GasError
		// TODO add method for logging vErrs
		//vErr, err := models.DB.ValidateAndUpdate(&treasureToBury)
		models.DB.ValidateAndUpdate(&treasureToBury)
	} else {
		treasureToBury.PRLStatus = models.GasPending
		treasureToBury.GasTxHash = txHash
		treasureToBury.GasTxNonce = nonce
		vErr, err := models.DB.ValidateAndUpdate(&treasureToBury)
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			return
		}
		if len(vErr.Errors) > 0 {
			oyster_utils.LogIfValidationError(
				"validation errors in bury_treasure_addresses in sendGas.", vErr, nil)
			return
		}
		oyster_utils.LogToSegment("bury_treasure_addresses: sendGas", analytics.NewProperties().
			Set("new_status", models.PRLStatusMap[treasureToBury.PRLStatus]).
			Set("eth_address", treasureToBury.ETHAddr))
		//go waitForGas(treasureToBury)
	}
}

func buryPRL(treasureToBury models.Treasure) {

	balanceOfPRL := EthWrapper.CheckPRLBalance(eth_gateway.StringToAddress(treasureToBury.ETHAddr))
	balanceOfETH := EthWrapper.CheckETHBalance(eth_gateway.StringToAddress(treasureToBury.ETHAddr))

	if balanceOfPRL.Int64() <= 0 || balanceOfETH.Int64() <= 0 {
		errorString := "Cannot bury treasure address due to insufficient balance in treasure wallet (" +
			treasureToBury.ETHAddr + ").  balance of PRL: " +
			fmt.Sprint(balanceOfPRL.Int64()) + "; balance of ETH: " + fmt.Sprint(balanceOfETH.Int64())
		err := errors.New(errorString)
		oyster_utils.LogIfError(err, nil)
		return
	}

	privateKey, err := eth_gateway.StringToPrivateKey(treasureToBury.DecryptTreasureEthKey())
	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return
	}

	callMsg := eth_gateway.OysterCallMsg{
		From:       eth_gateway.StringToAddress(treasureToBury.ETHAddr),
		PrivateKey: *privateKey,
	}

	success, txHash, nonce := EthWrapper.BuryPrl(callMsg)

	if !success {
		errorString := "\nFailure to bury  " + treasureToBury.ETHAddr
		err := errors.New(errorString)
		oyster_utils.LogIfError(err, nil)
		treasureToBury.PRLStatus = models.BuryError
		// TODO add method for logging vErrs
		//vErr, err := models.DB.ValidateAndUpdate(&treasureToBury)
		models.DB.ValidateAndUpdate(&treasureToBury)
	} else {
		treasureToBury.PRLStatus = models.BuryPending
		treasureToBury.BuryTxHash = txHash
		treasureToBury.BuryTxNonce = nonce
		vErr, err := models.DB.ValidateAndUpdate(&treasureToBury)
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			return
		}
		if len(vErr.Errors) > 0 {
			oyster_utils.LogIfValidationError(
				"validation errors in bury_treasure_addresses in buryPRL.", vErr, nil)
			return
		}
		oyster_utils.LogToSegment("bury_treasure_addresses: buryPRL", analytics.NewProperties().
			Set("new_status", models.PRLStatusMap[treasureToBury.PRLStatus]).
			Set("eth_address", treasureToBury.ETHAddr))
		//go waitForBury(treasureToBury)
	}
}

func waitForPRL(treasureToBury models.Treasure) {
	waitForConfirmation(treasureToBury, treasureToBury.PRLTxHash, treasureToBury.PRLTxNonce, eth_gateway.PRLTransfer)
}

func waitForGas(treasureToBury models.Treasure) {
	waitForConfirmation(treasureToBury, treasureToBury.GasTxHash, treasureToBury.GasTxNonce, eth_gateway.EthTransfer)
}

func waitForBury(treasureToBury models.Treasure) {
	waitForConfirmation(treasureToBury, treasureToBury.BuryTxHash, treasureToBury.BuryTxNonce, eth_gateway.PRLBury)
}

// TODO: get this to work and un-comment out the calls to waitForPRL, waitForGas, and waitForBury
func waitForConfirmation(treasureToBury models.Treasure, txHash string, txNonce int64, txType eth_gateway.TxType) {

	success := EthWrapper.WaitForConfirmation(eth_gateway.StringToTxHash(txHash), SecondsDelayForETHPolling)

	// we passed the row by value, get it again in case it has changed
	treasureRow := models.Treasure{}
	err := models.DB.Find(&treasureRow, treasureToBury.ID)
	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return
	}

	var newStatus models.PRLStatus

	if success == 1 {
		newStatus = updateStatusSuccess(txType, treasureRow)
	} else if success == 0 {
		newStatus = updateStatusFailed(txType, treasureRow)
	}

	if _, ok := models.PRLStatusMap[newStatus]; ok {
		treasureRow.PRLStatus = newStatus
		_, err = models.DB.ValidateAndUpdate(&treasureToBury)
		oyster_utils.LogIfError(err, nil)
	}
}

func updateStatusSuccess(txType eth_gateway.TxType, treasureRow models.Treasure) models.PRLStatus {
	var newStatus models.PRLStatus
	switch txType {
	case eth_gateway.PRLTransfer:
		if treasureRow.PRLStatus == models.PRLPending || treasureRow.PRLStatus == models.PRLError {
			newStatus = models.PRLConfirmed
		}
	case eth_gateway.EthTransfer:
		if treasureRow.PRLStatus == models.GasPending || treasureRow.PRLStatus == models.GasError {
			newStatus = models.GasConfirmed
		}
	case eth_gateway.PRLBury:
		if treasureRow.PRLStatus == models.BuryPending || treasureRow.PRLStatus == models.BuryError {
			newStatus = models.BuryConfirmed
		}
	default:
		logInvalidTxType(txType, treasureRow.PRLStatus)
	}
	return newStatus
}

func updateStatusFailed(txType eth_gateway.TxType, treasureRow models.Treasure) models.PRLStatus {
	var newStatus models.PRLStatus
	switch txType {
	case eth_gateway.PRLTransfer:
		if treasureRow.PRLStatus == models.PRLPending {
			newStatus = models.PRLError
		}
	case eth_gateway.EthTransfer:
		if treasureRow.PRLStatus == models.GasPending {
			newStatus = models.GasError
		}
	case eth_gateway.PRLBury:
		if treasureRow.PRLStatus == models.BuryPending {
			newStatus = models.BuryError
		}
	default:
		logInvalidTxType(txType, treasureRow.PRLStatus)
	}
	return newStatus
}

// logInvalidTxType Utility to log txType errors and prlStatus
func logInvalidTxType(txType eth_gateway.TxType, status models.PRLStatus) {
	txString := txToString(txType)
	errorMsg := fmt.Sprintf("not a valid tx type (%v) in bury_treasure_addresses waitForConfirmation  status : %v", txString, status)
	oyster_utils.LogIfError(errors.New(errorMsg), nil)
}

// txToString Utility to return the transaction type
func txToString(value eth_gateway.TxType) string {
	status := "Not Found"
	switch value {
	case eth_gateway.PRLTransfer:
		status = "PRL Transfer"
	case eth_gateway.EthTransfer:
		status = "Ether Transfer"
	case eth_gateway.PRLBury:
		status = "PRL Bury"
	case eth_gateway.PRLClaim:
		status = "PRL Claim"
	}
	return status
}

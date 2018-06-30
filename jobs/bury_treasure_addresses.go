package jobs

import (
	"errors"
	"fmt"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"gopkg.in/segmentio/analytics-go.v3"
	"os"
	"time"
)

func BuryTreasureAddresses(thresholdTime time.Time, PrometheusWrapper services.PrometheusService) {

	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramBuryTreasureAddresses, start)

	if oyster_utils.BrokerMode == oyster_utils.ProdMode &&
		os.Getenv("OYSTER_PAYS") == "" {

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

		PurgeFinishedTreasure()
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
		prlBalance := EthWrapper.CheckPRLBalance(services.StringToAddress(pending.ETHAddr))
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
		ethBalance := EthWrapper.CheckETHBalance(services.StringToAddress(pending.ETHAddr))
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
		buried, err := EthWrapper.CheckBuriedState(services.StringToAddress(pending.ETHAddr))
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			continue
		} else if buried {
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

func SetTimedOutTransactionsToError(thresholdTime time.Time) {
	timedOutTransactions, err := models.GetTreasuresToBuryByPRLStatusAndUpdateTime([]models.PRLStatus{
		models.PRLPending,
		models.GasPending,
		models.BuryPending,
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
	errorPRLTransactions, errPRL := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{
		models.PRLError,
	})
	if errPRL != nil {
		fmt.Println("Cannot get error'd treasures (for prl transactions) in bury_treasure_addresses: " + errPRL.Error())
		// already captured error in upstream function
	}
	errorGasTransactions, errGas := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{
		models.GasError,
	})
	if errGas != nil {
		fmt.Println("Cannot get error'd treasures (for gas transactions) in bury_treasure_addresses: " + errGas.Error())
		// already captured error in upstream function
	}
	errorBuryTransactions, errBury := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{
		models.BuryError,
	})
	if errBury != nil {
		fmt.Println("Cannot get error'd treasures (for bury transactions) in bury_treasure_addresses: " + errBury.Error())
		// already captured error in upstream function
	}

	if (len(errorPRLTransactions) == 0 && len(errorGasTransactions) == 0 && len(errorBuryTransactions) == 0) ||
		(errBury != nil && errGas != nil && errPRL != nil) {
		return
	}

	for _, errorTransaction := range errorPRLTransactions {
		errorTransaction.PRLStatus = models.PRLWaiting
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
	for _, errorTransaction := range errorGasTransactions {
		errorTransaction.PRLStatus = models.PRLConfirmed
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
	for _, errorTransaction := range errorBuryTransactions {
		errorTransaction.PRLStatus = models.GasConfirmed
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
	oyster_utils.LogToSegment("bury_treasure_addresses: StageTransactionsWithErrorsForRetry", analytics.NewProperties().
		Set("prl_num_error'd_transactions", fmt.Sprint(len(errorPRLTransactions))).
		Set("gas_num_error'd_transactions", fmt.Sprint(len(errorGasTransactions))).
		Set("bury_num_error'd_transactions", fmt.Sprint(len(errorBuryTransactions))))
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
	completeTreasures, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.BuryConfirmed})
	if err != nil {
		fmt.Println("Cannot get completed treasures in bury_treasure_addresses: " + err.Error())
		// already captured error in upstream function
		return
	}

	if len(completeTreasures) == 0 {
		return
	}

	for _, completeTreasure := range completeTreasures {
		ethAddr := completeTreasure.ETHAddr
		err := models.SetToTreasureBuriedByGenesisHash(completeTreasure.GenesisHash)
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			continue
		}

		// TODO:  Don't want to enable this for now
		// since we don't want to lose access to the
		// treasure addresses

		//err = models.DB.Destroy(&completeTreasure)
		//if err != nil {
		//	oyster_utils.LogIfError(err, nil)
		//	continue
		//}
		oyster_utils.LogToSegment("bury_treasure_addresses: PurgeFinishedTreasure", analytics.NewProperties().
			Set("eth_address", ethAddr))
	}
}

func sendPRL(treasureToBury models.Treasure) {

	balance := EthWrapper.CheckPRLBalance(services.MainWalletAddress)
	if balance.Int64() <= 0 || balance.Int64() < treasureToBury.GetPRLAmount().Int64() {
		errorString := "Cannot send PRL to treasure address due to insufficient balance in wallet.  balance: " +
			fmt.Sprint(balance.Int64()) + "; amount_to_send: " + fmt.Sprint(treasureToBury.GetPRLAmount().Int64())
		err := errors.New(errorString)
		oyster_utils.LogIfError(err, nil)
		return
	}

	amount := *treasureToBury.GetPRLAmount()

	callMsg, _ := EthWrapper.CreateSendPRLMessage(services.MainWalletAddress,
		services.MainWalletPrivateKey,
		services.StringToAddress(treasureToBury.ETHAddr), amount)

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

	gasToSend, err := EthWrapper.CalculateGasToSend(services.GasLimitPRLBury)
	if err != nil {
		fmt.Println("Cannot send Gas to treasure address: " + err.Error())
		// already captured error in upstream function
		return
	}

	balance := EthWrapper.CheckETHBalance(services.MainWalletAddress)
	if balance.Int64() < gasToSend.Int64() {
		errorString := "Cannot send Gas to treasure address due to insufficient balance in wallet.  balance: " +
			fmt.Sprint(balance.Int64()) + "; amount_to_send: " + fmt.Sprint(gasToSend.Int64())
		err := errors.New(errorString)
		oyster_utils.LogIfError(err, nil)
		return
	}

	_, txHash, nonce, err := EthWrapper.SendETH(
		services.MainWalletAddress,
		services.MainWalletPrivateKey,
		services.StringToAddress(treasureToBury.ETHAddr),
		gasToSend)
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

	balanceOfPRL := EthWrapper.CheckPRLBalance(services.StringToAddress(treasureToBury.ETHAddr))
	balanceOfETH := EthWrapper.CheckETHBalance(services.StringToAddress(treasureToBury.ETHAddr))

	if balanceOfPRL.Int64() <= 0 || balanceOfETH.Int64() <= 0 {
		errorString := "Cannot bury treasure address due to insufficient balance in treasure wallet (" +
			treasureToBury.ETHAddr + ").  balance of PRL: " +
			fmt.Sprint(balanceOfPRL.Int64()) + "; balance of ETH: " + fmt.Sprint(balanceOfETH.Int64())
		err := errors.New(errorString)
		oyster_utils.LogIfError(err, nil)
		return
	}

	privateKey, err := services.StringToPrivateKey(treasureToBury.DecryptTreasureEthKey())
	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return
	}

	callMsg := services.OysterCallMsg{
		From:       services.StringToAddress(treasureToBury.ETHAddr),
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
	waitForConfirmation(treasureToBury, treasureToBury.PRLTxHash, treasureToBury.PRLTxNonce, services.PRLTransfer)
}

func waitForGas(treasureToBury models.Treasure) {
	waitForConfirmation(treasureToBury, treasureToBury.GasTxHash, treasureToBury.GasTxNonce, services.EthTransfer)
}

func waitForBury(treasureToBury models.Treasure) {
	waitForConfirmation(treasureToBury, treasureToBury.BuryTxHash, treasureToBury.BuryTxNonce, services.PRLBury)
}

func waitForConfirmation(treasureToBury models.Treasure, txHash string, txNonce int64, txType services.TxType) {

	// TODO: get this to work and un-comment out the calls to waitForPRL, waitForGas, and waitForBury

	success := EthWrapper.WaitForConfirmation(services.StringToTxHash(txHash),
		SecondsDelayForETHPolling)

	// we passed the row by value, get it again in case it has changed
	treasureRow := models.Treasure{}
	err := models.DB.Find(treasureRow, treasureToBury.ID)
	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return
	}

	var newStatus models.PRLStatus

	if success == 1 {
		switch txType {
		case services.PRLTransfer:
			if treasureRow.PRLStatus == models.PRLPending || treasureRow.PRLStatus == models.PRLError {
				newStatus = models.PRLConfirmed
			}
		case services.EthTransfer:
			if treasureRow.PRLStatus == models.GasPending || treasureRow.PRLStatus == models.GasError {
				newStatus = models.GasConfirmed
			}
		case services.PRLBury:
			if treasureRow.PRLStatus == models.BuryPending || treasureRow.PRLStatus == models.BuryError {
				newStatus = models.BuryConfirmed
			}
		default:
			invalidTxTypeErr := errors.New("not a valid tx type in bury_treasure_addresses waitForConfirmation")
			oyster_utils.LogIfError(invalidTxTypeErr, nil)
			return
		}
	} else if success == 0 {
		switch txType {
		case services.PRLTransfer:
			if treasureRow.PRLStatus == models.PRLPending {
				newStatus = models.PRLError
			}
		case services.EthTransfer:
			if treasureRow.PRLStatus == models.GasPending {
				newStatus = models.GasError
			}
		case services.PRLBury:
			if treasureRow.PRLStatus == models.BuryPending {
				newStatus = models.BuryError
			}
		default:
			invalidTxTypeErr := errors.New("not a valid tx type in bury_treasure_addresses waitForConfirmation")
			oyster_utils.LogIfError(invalidTxTypeErr, nil)
			return
		}
	}

	if _, ok := models.PRLStatusMap[newStatus]; ok {
		treasureRow.PRLStatus = newStatus
		_, err = models.DB.ValidateAndUpdate(&treasureToBury)
		oyster_utils.LogIfError(err, nil)
	}
}

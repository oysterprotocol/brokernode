package jobs

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/getsentry/raven-go"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"gopkg.in/segmentio/analytics-go.v3"
	"time"
)

func ProcessPaidSessions(thresholdTime time.Time) {

	BuryTreasureInDataMaps()
	MarkBuriedMapsAsUnassigned()

	if oyster_utils.BrokerMode == oyster_utils.ProdMode {

		CheckPRLTransactions()
		CheckGasTransactions()
		SetTimedOutTransactionsToError(thresholdTime)
		StageTransactionsWithErrorsForRetry()

		SendPRLsToWaitingTreasureAddresses()
		SendGasToTreasureAddresses()
		InvokeBury()

		PurgeFinishedTreasure()
		// TODO: unit tests
	}
}

func BuryTreasureInDataMaps() error {

	unburiedSessions, err := models.GetSessionsThatNeedTreasure()

	if err != nil {
		fmt.Println(err)
	}

	for _, unburiedSession := range unburiedSessions {

		treasureIndex, err := unburiedSession.GetTreasureMap()
		if err != nil {
			fmt.Println(err)
			return err
		}

		BuryTreasure(treasureIndex, &unburiedSession)
	}
	return nil
}

func BuryTreasure(treasureIndexMap []models.TreasureMap, unburiedSession *models.UploadSession) error {

	for _, entry := range treasureIndexMap {
		treasureChunks, err := models.GetDataMapByGenesisHashAndChunkIdx(unburiedSession.GenesisHash, entry.Idx)
		if err != nil {
			fmt.Println(err)
			return err
		}
		if len(treasureChunks) == 0 || len(treasureChunks) > 1 {
			errString := "did not find a chunk that matched genesis_hash and chunk_idx in process_paid_sessions, or " +
				"found duplicate chunks"
			err = errors.New(errString)
			oyster_utils.LogIfError(err)
			return err
		}

		decryptedKey, err := treasureChunks[0].DecryptEthKey(entry.Key)
		if err != nil {
			fmt.Println(err)
			return err
		}

		treasureChunks[0].Message, err = models.CreateTreasurePayload(decryptedKey, treasureChunks[0].Hash, models.MaxSideChainLength)
		if err != nil {
			fmt.Println(err)
			return err
		}
		models.DB.ValidateAndSave(&treasureChunks[0])

		oyster_utils.LogToSegment("process_paid_sessions: treasure_payload_buried_in_data_map", analytics.NewProperties().
			Set("genesis_hash", unburiedSession.GenesisHash).
			Set("sector", entry.Sector).
			Set("chunk_idx", entry.Idx).
			Set("address", treasureChunks[0].Address).
			Set("message", treasureChunks[0].Message))
	}
	unburiedSession.TreasureStatus = models.TreasureInDataMapComplete
	models.DB.ValidateAndSave(unburiedSession)
	return nil
}

// marking the maps as "Unassigned" will trigger them to get processed by the process_unassigned_chunks cron task.
func MarkBuriedMapsAsUnassigned() {
	readySessions, err := models.GetReadySessions()
	if err != nil {
		fmt.Println(err)
	}

	for _, readySession := range readySessions {

		pendingChunks, err := models.GetPendingChunksBySession(readySession, 1)
		if err != nil {
			fmt.Println(err)
		}

		if len(pendingChunks) > 0 {
			oyster_utils.LogToSegment("process_paid_sessions: mark_data_maps_as_ready", analytics.NewProperties().
				Set("genesis_hash", readySession.GenesisHash))

			err = readySession.BulkMarkDataMapsAsUnassigned()
		}
	}
}

func CheckPRLTransactions() {

	prlsPending, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.PRLPending})
	if err != nil {
		fmt.Println("Cannot get treasures with prls pending in process_paid_sessions: " + err.Error())
		// already captured error in upstream function
		return
	}
	if len(prlsPending) == 0 {
		return
	}

	for _, pending := range prlsPending {
		prlBalance := EthWrapper.CheckPRLBalance(services.StringToAddress(pending.ETHAddr))
		expectedPRLBalance := pending.GetPRLAmount()
		if prlBalance.Int64() > 0 && prlBalance.Int64() > expectedPRLBalance.Int64() {
			pending.PRLStatus = models.PRLConfirmed
			vErr, err := models.DB.ValidateAndUpdate(&pending)
			if err != nil {
				oyster_utils.LogIfError(err)
				return
			}
			if len(vErr.Errors) > 0 {
				errString := "validation errors in process_paid_sessions in CheckPRLTransactions: " + fmt.Sprint(vErr.Errors)
				err = errors.New(errString)
				oyster_utils.LogIfError(err)
				return
			}
			oyster_utils.LogToSegment("process_paid_sessions: CheckPRLTransactions", analytics.NewProperties().
				Set("new_status", models.PRLStatusMap[pending.PRLStatus]).
				Set("eth_address", pending.ETHAddr))
		}
	}
}

func CheckGasTransactions() {

	gasPending, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.GasPending})
	if err != nil {
		fmt.Println("Cannot get treasures with prls pending in process_paid_sessions: " + err.Error())
		// already captured error in upstream function
		return
	}
	if len(gasPending) == 0 {
		return
	}

	for _, pending := range gasPending {
		ethBalance := EthWrapper.CheckETHBalance(services.StringToAddress(pending.ETHAddr))
		if ethBalance.Int64() > 0 {
			pending.PRLStatus = models.GasConfirmed
			vErr, err := models.DB.ValidateAndUpdate(&pending)
			if err != nil {
				oyster_utils.LogIfError(err)
				return
			}
			if len(vErr.Errors) > 0 {
				errString := "validation errors in process_paid_sessions in CheckGasTransactions: " + fmt.Sprint(vErr.Errors)
				err = errors.New(errString)
				oyster_utils.LogIfError(err)
				return
			}
			oyster_utils.LogToSegment("process_paid_sessions: CheckGasTransactions", analytics.NewProperties().
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
		fmt.Println("Cannot get timed out treasures in process_paid_sessions: " + err.Error())
		// already captured error in upstream function
		return
	}

	if len(timedOutTransactions) == 0 {
		return
	}

	for _, timedOutTransaction := range timedOutTransactions {
		oldStatus := timedOutTransaction.PRLStatus
		timedOutTransaction.PRLStatus = models.PRLStatus(int((timedOutTransaction.PRLStatus)-1) * -1)
		vErr, err := models.DB.ValidateAndUpdate(&timedOutTransaction)
		if err != nil {
			oyster_utils.LogIfError(err)
			continue
		}
		if len(vErr.Errors) > 0 {
			errString := "validation errors in process_paid_sessions in SetTimedOutTransactionsToError: " + fmt.Sprint(vErr.Errors)
			err = errors.New(errString)
			oyster_utils.LogIfError(err)
			continue
		}
		oyster_utils.LogToSegment("process_paid_sessions: SetTimedOutTransactionsToError", analytics.NewProperties().
			Set("old_status", models.PRLStatusMap[oldStatus]).
			Set("new_status", models.PRLStatusMap[timedOutTransaction.PRLStatus]).
			Set("eth_address", timedOutTransaction.ETHAddr))
	}
}

func StageTransactionsWithErrorsForRetry() {
	errorTransactions, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{
		models.PRLError,
		models.GasError,
		models.BuryError,
	})
	if err != nil {
		fmt.Println("Cannot get error'd treasures in process_paid_sessions: " + err.Error())
		// already captured error in upstream function
		return
	}

	if len(errorTransactions) == 0 {
		return
	}

	for _, errorTransaction := range errorTransactions {
		oldStatus := errorTransaction.PRLStatus
		errorTransaction.PRLStatus = models.PRLStatus(int(errorTransaction.PRLStatus) * -1)
		vErr, err := models.DB.ValidateAndUpdate(&errorTransaction)
		if err != nil {
			fmt.Println(err)
			raven.CaptureError(err, nil)
			continue
		}
		if len(vErr.Errors) > 0 {
			errString := "validation errors in process_paid_sessions in StageTransactionsWithErrorsForRetry: " + fmt.Sprint(vErr.Errors)
			err = errors.New(errString)
			fmt.Println(err)
			raven.CaptureError(err, nil)
			continue
		}
		oyster_utils.LogToSegment("process_paid_sessions: StageTransactionsWithErrorsForRetry", analytics.NewProperties().
			Set("old_status", models.PRLStatusMap[oldStatus]).
			Set("new_status", models.PRLStatusMap[errorTransaction.PRLStatus]).
			Set("eth_address", errorTransaction.ETHAddr))
	}
}

func SendPRLsToWaitingTreasureAddresses() {

	waitingForPRLS, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.PRLWaiting})
	if err != nil {
		fmt.Println("Cannot get treasures awaiting PRLs in process_paid_sessions: " + err.Error())
		// already captured error in upstream function
		return
	}

	if len(waitingForPRLS) == 0 {
		return
	}

	for _, waitingAddress := range waitingForPRLS {
		sendPRL(waitingAddress)
	}
}

func SendGasToTreasureAddresses() {
	waitingForGas, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.PRLConfirmed})
	if err != nil {
		fmt.Println("Cannot get treasures awaiting gas in process_paid_sessions: " + err.Error())
		// already captured error in upstream function
		return
	}

	if len(waitingForGas) == 0 {
		return
	}

	for _, waitingAddress := range waitingForGas {
		sendGas(waitingAddress)
	}
}

func InvokeBury() {
	readyToInvokeBury, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.GasConfirmed})
	if err != nil {
		fmt.Println("Cannot get treasures awaiting bury() in process_paid_sessions: " + err.Error())
		// already captured error in upstream function
		return
	}

	if len(readyToInvokeBury) == 0 {
		return
	}

	for _, buryAddress := range readyToInvokeBury {
		buryPRL(buryAddress)
	}
}

func PurgeFinishedTreasure() {
	completeTreasures, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.BuryConfirmed})
	if err != nil {
		fmt.Println("Cannot get completed treasures in process_paid_sessions: " + err.Error())
		// already captured error in upstream function
		return
	}

	if len(completeTreasures) == 0 {
		return
	}

	for _, completeTreasure := range completeTreasures {
		ethAddr := completeTreasure.ETHAddr
		err := models.DB.Destroy(&completeTreasure)
		if err != nil {
			oyster_utils.LogIfError(err)
			continue
		}
		oyster_utils.LogToSegment("process_paid_sessions: PurgeFinishedTreasure", analytics.NewProperties().
			Set("eth_address", ethAddr))
	}
}

func sendPRL(treasureToBury models.Treasure) {

	gas, err := EthWrapper.GetGasPrice()
	if err != nil {
		fmt.Println("Cannot send PRL to treasure address: " + err.Error())
		// already captured error in upstream function
		return
	}

	// TODO:  Need balance of PRL, need to have at least enough ETH for gas for transaction
	balance := EthWrapper.CheckPRLBalance(services.MainWalletAddress)
	if balance.Int64() <= 0 || balance.Int64() < treasureToBury.GetPRLAmount().Int64() {
		errorString := "Cannot send PRL to treasure address due to insufficient balance in wallet.  balance: " +
			fmt.Sprint(balance.Int64()) + "; amount_to_send: " + fmt.Sprint(treasureToBury.GetPRLAmount().Int64())
		err := errors.New(errorString)
		raven.CaptureError(err, nil)
		return
	}

	// TODO:  pull the lines below this out if keystore stuff gets fixed
	privateKeyString := services.MainWalletKey

	if privateKeyString[0:2] != "0x" && privateKeyString[0:2] != "0X" {
		privateKeyString = "0x" + privateKeyString
	}

	privateKeyBigInt := hexutil.MustDecodeBig(privateKeyString)

	privateKey := EthWrapper.GeneratePublicKeyFromPrivateKey(crypto.S256(), privateKeyBigInt)
	// TODO:  pull out the lines above this if keystore stuff gets fixed

	// TODO:  What else do I need here?
	callMsg := services.OysterCallMsg{
		From:       services.MainWalletAddress,
		To:         services.StringToAddress(treasureToBury.ETHAddr),
		Amount:     *treasureToBury.GetPRLAmount(),
		Gas:        gas.Uint64(),
		PrivateKey: *privateKey,
	}

	sendSuccess := EthWrapper.SendPRL(callMsg)
	if !sendSuccess {
		errorString := "Failure sending " + fmt.Sprint(treasureToBury.GetPRLAmount().Int64()) + " PRL to " +
			treasureToBury.ETHAddr
		err := errors.New(errorString)
		oyster_utils.LogIfError(err)
		treasureToBury.PRLStatus = models.PRLError
		//vErr, err := models.DB.ValidateAndUpdate(&treasureToBury)
		// TODO add method for logging vErrs
		models.DB.Save(&treasureToBury)
	} else {
		treasureToBury.PRLStatus = models.PRLPending
		vErr, err := models.DB.ValidateAndUpdate(&treasureToBury)
		if err != nil {
			oyster_utils.LogIfError(err)
			return
		}
		if len(vErr.Errors) > 0 {
			errString := "validation errors in process_paid_sessions in sendPRL: " + fmt.Sprint(vErr.Errors)
			err = errors.New(errString)
			oyster_utils.LogIfError(err)
			return
		}
		oyster_utils.LogToSegment("process_paid_sessions: sendPRL", analytics.NewProperties().
			Set("new_status", models.PRLStatusMap[treasureToBury.PRLStatus]).
			Set("eth_address", treasureToBury.ETHAddr))
		go waitForPRLs(treasureToBury)
	}
}

func sendGas(treasureToBury models.Treasure) {

	gas, err := EthWrapper.GetGasPrice()
	if err != nil {
		fmt.Println("Cannot send Gas to treasure address: " + err.Error())
		// already captured error in upstream function
		return
	}

	balance := EthWrapper.CheckETHBalance(services.MainWalletAddress)
	if balance.Int64() <= 0 || balance.Int64() < gas.Int64() {
		errorString := "Cannot send Gas to treasure address due to insufficient balance in wallet.  balance: " +
			fmt.Sprint(balance.Int64()) + "; amount_to_send: " + fmt.Sprint(gas.Int64())
		err := errors.New(errorString)
		oyster_utils.LogIfError(err)
		return
	}

	_, err = EthWrapper.SendETH(services.StringToAddress(treasureToBury.ETHAddr), gas)

	if err != nil {
		errorString := "Failure sending " + fmt.Sprint(gas.Int64()) + " Gas to " + treasureToBury.ETHAddr
		err := errors.New(errorString)
		oyster_utils.LogIfError(err)
		treasureToBury.PRLStatus = models.GasError
		//vErr, err := models.DB.ValidateAndUpdate(&treasureToBury)
		// TODO add method for logging vErrs
		models.DB.Save(&treasureToBury)
	} else {
		treasureToBury.PRLStatus = models.GasPending
		vErr, err := models.DB.ValidateAndUpdate(&treasureToBury)
		if err != nil {
			oyster_utils.LogIfError(err)
			return
		}
		if len(vErr.Errors) > 0 {
			errString := "validation errors in process_paid_sessions in sendGas: " + fmt.Sprint(vErr.Errors)
			err = errors.New(errString)
			oyster_utils.LogIfError(err)
			return
		}
		oyster_utils.LogToSegment("process_paid_sessions: sendGas", analytics.NewProperties().
			Set("new_status", models.PRLStatusMap[treasureToBury.PRLStatus]).
			Set("eth_address", treasureToBury.ETHAddr))
		go waitForGas(treasureToBury)
	}
}

func buryPRL(treasureToBury models.Treasure) {

	balanceOfPRL := EthWrapper.CheckPRLBalance(services.StringToAddress(treasureToBury.ETHAddr))
	balanceOfETH := EthWrapper.CheckETHBalance(services.StringToAddress(treasureToBury.ETHAddr))

	if balanceOfPRL.Int64() <= 0 || balanceOfETH.Int64() <= 0 {
		errorString := "Cannot bury treasure address due to insufficient balance in treasure wallet.  balance of PRL: " +
			fmt.Sprint(balanceOfPRL.Int64()) + "; balance of ETH: " + fmt.Sprint(balanceOfETH.Int64())
		err := errors.New(errorString)
		oyster_utils.LogIfError(err)
		return
	}

	// TODO:  What else do I need here?
	callMsg := services.OysterCallMsg{
		To: services.StringToAddress(treasureToBury.ETHAddr),
	}

	success := EthWrapper.BuryPrl(callMsg)
	if !success {
		errorString := "Failure to bury  " + treasureToBury.ETHAddr
		err := errors.New(errorString)
		oyster_utils.LogIfError(err)
		treasureToBury.PRLStatus = models.BuryError
		//vErr, err := models.DB.ValidateAndUpdate(&treasureToBury)
		// TODO add method for logging vErrs
		models.DB.Save(&treasureToBury)
	} else {
		treasureToBury.PRLStatus = models.BuryPending
		vErr, err := models.DB.ValidateAndUpdate(&treasureToBury)
		if err != nil {
			oyster_utils.LogIfError(err)
			return
		}
		if len(vErr.Errors) > 0 {
			errString := "validation errors in process_paid_sessions in buryPRL: " + fmt.Sprint(vErr.Errors)
			err = errors.New(errString)
			oyster_utils.LogIfError(err)
			return
		}
		oyster_utils.LogToSegment("process_paid_sessions: buryPRL", analytics.NewProperties().
			Set("new_status", models.PRLStatusMap[treasureToBury.PRLStatus]).
			Set("eth_address", treasureToBury.ETHAddr))
		go waitForBury(treasureToBury)
	}
}

func waitForPRLs(treasureToBury models.Treasure) {
	treasureAddress := services.StringToAddress(treasureToBury.ETHAddr)
	_, err := EthWrapper.WaitForTransfer(treasureAddress)

	prlStatus := models.PRLConfirmed
	if err != nil {
		oyster_utils.LogToSegment("process_paid_sessions: waitForPRLs_error", analytics.NewProperties().
			Set("new_status", models.PRLStatusMap[treasureToBury.PRLStatus]).
			Set("eth_address", treasureToBury.ETHAddr))
		prlStatus = models.PRLError
	} else {
		oyster_utils.LogToSegment("process_paid_sessions: waitForPRLs_confirmed", analytics.NewProperties().
			Set("new_status", models.PRLStatusMap[treasureToBury.PRLStatus]).
			Set("eth_address", treasureToBury.ETHAddr))
	}

	treasureToBury.PRLStatus = prlStatus

	_, err = models.DB.ValidateAndUpdate(&treasureToBury)
	oyster_utils.LogIfError(err)
}

func waitForGas(treasureToBury models.Treasure) {

	treasureAddress := services.StringToAddress(treasureToBury.ETHAddr)
	//TODO need to wait for transfer of ETH, not PRL
	_, err := EthWrapper.WaitForTransfer(treasureAddress)

	prlStatus := models.GasConfirmed
	if err != nil {
		oyster_utils.LogToSegment("process_paid_sessions: waitForGas_error", analytics.NewProperties().
			Set("new_status", models.PRLStatusMap[treasureToBury.PRLStatus]).
			Set("eth_address", treasureToBury.ETHAddr))
		prlStatus = models.GasError
	} else {
		oyster_utils.LogToSegment("process_paid_sessions: waitForGas_confirmed", analytics.NewProperties().
			Set("new_status", models.PRLStatusMap[treasureToBury.PRLStatus]).
			Set("eth_address", treasureToBury.ETHAddr))
	}

	treasureToBury.PRLStatus = prlStatus

	_, err = models.DB.ValidateAndUpdate(&treasureToBury)
	oyster_utils.LogIfError(err)
}

func waitForBury(treasureToBury models.Treasure) {
	treasureAddress := services.StringToAddress(treasureToBury.ETHAddr)
	// TODO find some blocking call for which we can wait and get
	// the bury() status
	_, err := EthWrapper.WaitForTransfer(treasureAddress)

	prlStatus := models.BuryConfirmed
	if err != nil {
		oyster_utils.LogToSegment("process_paid_sessions: waitForBury_error", analytics.NewProperties().
			Set("new_status", models.PRLStatusMap[treasureToBury.PRLStatus]).
			Set("eth_address", treasureToBury.ETHAddr))
		prlStatus = models.BuryError
	} else {
		oyster_utils.LogToSegment("process_paid_sessions: waitForBury_confirmed", analytics.NewProperties().
			Set("new_status", models.PRLStatusMap[treasureToBury.PRLStatus]).
			Set("eth_address", treasureToBury.ETHAddr))
	}

	treasureToBury.PRLStatus = prlStatus

	_, err = models.DB.ValidateAndUpdate(&treasureToBury)
	oyster_utils.LogIfError(err)
}

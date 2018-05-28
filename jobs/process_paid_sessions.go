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
)

func ProcessPaidSessions() {

	BuryTreasureInDataMaps()
	MarkBuriedMapsAsUnassigned()

	if oyster_utils.BrokerMode == oyster_utils.ProdMode {
		SendPRLsToWaitingTreasureAddresses()
		SendGasToTreasureAddresses()
		InvokeBury()

		// TODO:  add methods to check status of pending transactions, or subscribe to them and update
		// PRLStatus in the row in the DB, and add tests for these methods once we know they work.
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

func SendPRLsToWaitingTreasureAddresses() {

	waitingForPRLS, err := models.GetTreasuresToBuryByPRLStatus(models.PRLWaiting)
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
	waitingForGas, err := models.GetTreasuresToBuryByPRLStatus(models.PRLConfirmed)
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
	readyToInvokeBury, err := models.GetTreasuresToBuryByPRLStatus(models.GasConfirmed)
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

func sendPRL(treasureToBury models.Treasure) {

	gas, err := services.EthWrapper.GetGasPrice()
	if err != nil {
		fmt.Println("Cannot send PRL to treasure address: " + err.Error())
		// already captured error in upstream function
		return
	}

	// TODO:  Need balance of PRL, need to have at least enough ETH for gas for transaction
	balance := services.EthWrapper.CheckPRLBalance(services.MainWalletAddress)
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
	privateKey := services.EthWrapper.GeneratePublicKeyFromPrivateKey(crypto.S256(), privateKeyBigInt)
	// TODO:  pull out the lines above this if keystore stuff gets fixed

	fmt.Println("treasureToBury.Ethaddr")
	fmt.Println(treasureToBury.ETHAddr)

	// TODO:  What else do I need here?
	callMsg := services.OysterCallMsg{
		From:       services.MainWalletAddress,
		To:         services.StringToAddress(treasureToBury.ETHAddr),
		Amount:     *treasureToBury.GetPRLAmount(),
		Gas:        gas.Uint64(),
		PrivateKey: *privateKey,
	}

	sendSuccess := services.EthWrapper.SendPRL(callMsg)
	if !sendSuccess {
		errorString := "Failure sending " + fmt.Sprint(treasureToBury.GetPRLAmount().Int64()) + " PRL to " +
			treasureToBury.ETHAddr
		err := errors.New(errorString)
		oyster_utils.LogIfError(err)
	} else {
		treasureToBury.PRLStatus = models.PRLPending
		models.DB.ValidateAndUpdate(&treasureToBury)
	}
}

func sendGas(treasureToBury models.Treasure) {

	gas, err := services.EthWrapper.GetGasPrice()
	if err != nil {
		fmt.Println("Cannot send Gas to treasure address: " + err.Error())
		// already captured error in upstream function
		return
	}

	// TODO:  Need balance of ETH, need to have at least enough ETH for gas for transaction along with the gas
	// we are sending to the treasure address
	balance := services.EthWrapper.CheckETHBalance(services.MainWalletAddress)
	if balance.Int64() <= 0 || balance.Int64() < gas.Int64() {
		errorString := "Cannot send Gas to treasure address due to insufficient balance in wallet.  balance: " +
			fmt.Sprint(balance.Int64()) + "; amount_to_send: " + fmt.Sprint(gas.Int64())
		err := errors.New(errorString)
		oyster_utils.LogIfError(err)
		return
	}

	_, err = services.EthWrapper.SendETH(services.StringToAddress(treasureToBury.ETHAddr), gas)
	if err != nil {
		errorString := "Failure sending " + fmt.Sprint(gas.Int64()) + " Gas to " + treasureToBury.ETHAddr
		err := errors.New(errorString)
		oyster_utils.LogIfError(err)
	} else {
		treasureToBury.PRLStatus = models.GasPending
		models.DB.ValidateAndUpdate(&treasureToBury)
	}
}

func buryPRL(treasureToBury models.Treasure) {

	// TODO:  Need balance of PRL, it should be more than 0
	// TODO:  Need balance of ETH, it should be more than 0

	balanceOfPRL := services.EthWrapper.CheckPRLBalance(services.StringToAddress(treasureToBury.ETHAddr))
	balanceOfETH := services.EthWrapper.CheckETHBalance(services.StringToAddress(treasureToBury.ETHAddr))

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

	success := services.EthWrapper.BuryPrl(callMsg)
	if !success {
		errorString := "Failure bury  " + treasureToBury.ETHAddr
		err := errors.New(errorString)
		oyster_utils.LogIfError(err)
	} else {
		treasureToBury.PRLStatus = models.BuryPending
		models.DB.ValidateAndUpdate(&treasureToBury)
	}
}

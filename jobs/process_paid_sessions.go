package jobs

import (
	"errors"
	"fmt"
	"github.com/getsentry/raven-go"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"gopkg.in/segmentio/analytics-go.v3"
	"log"
)

func init() {
}

func ProcessPaidSessions() {

	BuryTreasureInDataMaps()
	MarkBuriedMapsAsUnassigned()

	if oyster_utils.BrokerMode == oyster_utils.ProdMode {
		SendPRLsToWaitingTreasureAddresses()
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
			log.Println(errString)
			err = errors.New(errString)
			raven.CaptureError(err, nil)
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

func InvokeBury() {
	readyToInvokeBury, err := models.GetTreasuresToBuryByPRLStatus(models.PRLConfirmed)
	if err != nil {
		fmt.Println("Cannot get treasures awaiting bury() in process_paid_sessions: " + err.Error())
		// already captured error in upstream function
		return
	}

	fmt.Println(readyToInvokeBury)

	// TODO:  do whatever we need to do to invoke bury() on these treasures
}

func sendPRL(treasureToBury models.Treasure) {

	gas, err := services.EthWrapper.GetGasPrice()
	if err != nil {
		fmt.Println("Cannot send PRL to treasure address: " + err.Error())
		// already captured error in upstream function
		return
	}

	balance := services.EthWrapper.CheckBalance(services.MainWalletAddress)
	if balance.Int64() <= 0 || balance.Int64() < treasureToBury.GetPRLAmount().Int64() {
		errorString := "Cannot send PRL to treasure address due to insufficient balance in wallet.  balance: " + fmt.Sprint(balance.Int64()) + " amount_to_send: " + fmt.Sprint(treasureToBury.GetPRLAmount().Int64())
		err := errors.New(errorString)
		raven.CaptureError(err, nil)
		return
	}

	// TODO:  What else do I need here?
	callMsg := services.OysterCallMsg{
		From:   services.MainWalletAddress,
		To:     services.StringToAddress(treasureToBury.ETHAddr),
		Amount: *treasureToBury.GetPRLAmount(),
		Gas:    gas.Uint64(),
	}

	sendSuccess := services.EthWrapper.SendPRL(callMsg)
	if !sendSuccess {
		errorString := "Failure sending " + fmt.Sprint(treasureToBury.GetPRLAmount().Int64()) + " PRL to " + treasureToBury.ETHAddr
		err := errors.New(errorString)
		raven.CaptureError(err, nil)
	} else {
		treasureToBury.PRLStatus = models.PRLPending
		models.DB.ValidateAndUpdate(&treasureToBury)
	}
}

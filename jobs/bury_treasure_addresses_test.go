package jobs_test

import (
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"math/big"
	"time"
)

func (suite *JobsSuite) Test_CheckPRLTransactions() {

	hasCalledCheckPRLBalance := false

	jobs.EthWrapper = services.Eth{
		CheckPRLBalance: func(addr common.Address) *big.Int {
			hasCalledCheckPRLBalance = true
			return big.NewInt(600000000000000000)
		},
	}

	generateTreasuresToBury(suite, 1, models.PRLPending)
	pendingPRL, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.PRLPending})
	suite.Nil(err)
	suite.Equal(1, len(pendingPRL))

	suite.Equal(false, hasCalledCheckPRLBalance)

	jobs.CheckPRLTransactions()

	pendingPRL, err = models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.PRLPending})
	suite.Nil(err)
	suite.Equal(0, len(pendingPRL))
	suite.Equal(true, hasCalledCheckPRLBalance)
}

func (suite *JobsSuite) Test_CheckGasTransactions() {

	hasCalledCheckETHBalance := false

	jobs.EthWrapper = services.Eth{
		CheckETHBalance: func(addr common.Address) *big.Int {
			hasCalledCheckETHBalance = true
			return big.NewInt(600000000000000000)
		},
	}

	generateTreasuresToBury(suite, 1, models.GasPending)
	pendingGas, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.GasPending})
	suite.Nil(err)
	suite.Equal(1, len(pendingGas))

	suite.Equal(false, hasCalledCheckETHBalance)

	jobs.CheckGasTransactions()

	pendingGas, err = models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.GasPending})
	suite.Nil(err)
	suite.Equal(0, len(pendingGas))
	suite.Equal(true, hasCalledCheckETHBalance)
}

func (suite *JobsSuite) Test_SetTimedOutTransactionsToError() {

	generateTreasuresToBury(suite, 2, models.GasPending)
	generateTreasuresToBury(suite, 2, models.PRLPending)
	generateTreasuresToBury(suite, 2, models.BuryPending)

	pending, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{
		models.GasPending,
		models.PRLPending,
		models.BuryPending})
	suite.Nil(err)
	suite.Equal(6, len(pending))

	errord, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{
		models.GasError,
		models.PRLError,
		models.BuryError})
	suite.Nil(err)
	suite.Equal(0, len(errord))

	// set some of the transactions to be old
	err = suite.DB.RawQuery("UPDATE treasures SET updated_at = ? WHERE eth_addr = ?",
		time.Now().Add(-24*time.Hour), pending[0].ETHAddr).All(&[]models.UploadSession{})
	suite.Nil(err)

	err = suite.DB.RawQuery("UPDATE treasures SET updated_at = ? WHERE eth_addr = ?",
		time.Now().Add(-24*time.Hour), pending[2].ETHAddr).All(&[]models.UploadSession{})
	suite.Nil(err)

	err = suite.DB.RawQuery("UPDATE treasures SET updated_at = ? WHERE eth_addr = ?",
		time.Now().Add(-24*time.Hour), pending[4].ETHAddr).All(&[]models.UploadSession{})
	suite.Nil(err)

	jobs.SetTimedOutTransactionsToError(time.Now().Add(-1 * time.Hour))

	pending, err = models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{
		models.GasPending,
		models.PRLPending,
		models.BuryPending})
	suite.Nil(err)
	suite.Equal(3, len(pending))

	errord, err = models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{
		models.GasError,
		models.PRLError,
		models.BuryError})
	suite.Nil(err)
	suite.Equal(3, len(errord))
}

func (suite *JobsSuite) Test_StageTransactionsWithErrorsForRetry() {
	generateTreasuresToBury(suite, 1, models.GasError)
	generateTreasuresToBury(suite, 1, models.PRLError)
	generateTreasuresToBury(suite, 1, models.BuryError)

	waiting, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{
		models.PRLWaiting,
		models.PRLConfirmed,
		models.GasConfirmed})
	suite.Nil(err)
	suite.Equal(0, len(waiting))

	errord, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{
		models.GasError,
		models.PRLError,
		models.BuryError})
	suite.Nil(err)
	suite.Equal(3, len(errord))

	jobs.StageTransactionsWithErrorsForRetry()

	waiting, err = models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{
		models.PRLWaiting,
		models.PRLConfirmed,
		models.GasConfirmed})
	suite.Nil(err)
	suite.Equal(3, len(waiting))

	errord, err = models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{
		models.GasError,
		models.PRLError,
		models.BuryError})
	suite.Nil(err)
	suite.Equal(0, len(errord))
}

func (suite *JobsSuite) Test_SendPRLsToWaitingTreasureAddresses() {
	generateTreasuresToBury(suite, 4, models.PRLWaiting)
	waiting, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.PRLWaiting})
	suite.Nil(err)
	suite.Equal(4, len(waiting))

	// set one of the treasures to a higher PRL amount than what we have in our
	// balance
	waiting[0].SetPRLAmount(big.NewInt(700000000000000000))
	suite.DB.Save(&waiting)

	// make one of the addresses fail in SendPRLFromOyster
	sendPRLFromOysterFailedAddress := waiting[1].ETHAddr

	// make 1 remain pending since the belance has not arrived
	noBalanceAddress := waiting[2].ETHAddr

	hasCalledCheckPRLBalance := false
	hasCalledSendPRL := false

	jobs.EthWrapper = services.Eth{
		CheckPRLBalance: func(addr common.Address) *big.Int {
			hasCalledCheckPRLBalance = true
			// set one of the transactions to 0 balance so it will remain pending
			if addr == services.StringToAddress(noBalanceAddress) {
				return big.NewInt(0)
			}
			return big.NewInt(600000000000000000)
		},
		SendPRLFromOyster: func(msg services.OysterCallMsg) (bool, string, int64) {
			hasCalledSendPRL = true
			// make one of the transfers unsuccessful
			if msg.To == services.StringToAddress(sendPRLFromOysterFailedAddress) {
				return false, "", -1
			}
			return true, "some__transaction_hash", 0
		},
		GeneratePublicKeyFromPrivateKey: services.EthWrapper.GeneratePublicKeyFromPrivateKey,
		CreateSendPRLMessage:            services.EthWrapper.CreateSendPRLMessage,
	}

	// 1 transfer shouldn't go through due to insufficient balance and should remain waiting
	// 1 transfer shouldn't go through to to SendPRLFromOyster failure and should be set to an error
	// 1 should remain pending since the balance has not arrived
	// 1 should succeed.  CheckPRLTransactions will cause this transaction to come back
	// as a success, and the status will change to confirmed

	// call method under test
	jobs.SendPRLsToWaitingTreasureAddresses()

	// call a method to check the status of the transactions, so they will
	// get transitioned to their next statuses
	jobs.CheckPRLTransactions()

	waiting, err = models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.PRLWaiting})
	suite.Nil(err)
	suite.Equal(1, len(waiting))

	confirmed, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.PRLConfirmed})
	suite.Nil(err)
	suite.Equal(1, len(confirmed))

	pending, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.PRLPending})
	suite.Nil(err)
	suite.Equal(1, len(pending))

	errored, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.PRLError})
	suite.Nil(err)
	suite.Equal(1, len(errored))

	suite.Equal(true, hasCalledCheckPRLBalance)
	suite.Equal(true, hasCalledSendPRL)
}

func (suite *JobsSuite) Test_SendGasToTreasureAddresses() {
	generateTreasuresToBury(suite, 3, models.PRLConfirmed)
	waitingForGas, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.PRLConfirmed})
	suite.Nil(err)
	suite.Equal(3, len(waitingForGas))

	hasCalledCalculateGasToSend := false
	hasCalledCheckETHBalance := false
	hasCalledSendGas := false

	failSendEthAddress := waitingForGas[0].ETHAddr
	noBalanceAddress := waitingForGas[1].ETHAddr

	jobs.EthWrapper = services.Eth{
		CheckETHBalance: func(addr common.Address) *big.Int {
			hasCalledCheckETHBalance = true
			// cause one to remain pending due to no balance
			if addr == services.StringToAddress(noBalanceAddress) {
				return big.NewInt(0)
			}
			return big.NewInt(600000000000000000)
		},
		SendETH: func(address common.Address, gas *big.Int) (types.Transactions, string, int64, error) {
			hasCalledSendGas = true
			// make one of the transfers unsuccessful
			if address == services.StringToAddress(failSendEthAddress) {
				return types.Transactions{}, "", -1, errors.New("FAIL")
			}
			return types.Transactions{}, "111111", 1, nil
		},
		GeneratePublicKeyFromPrivateKey: services.EthWrapper.GeneratePublicKeyFromPrivateKey,
		CalculateGasToSend: func(desiredGasLimit uint64) (*big.Int, error) {
			hasCalledCalculateGasToSend = true
			gasPrice := oyster_utils.ConvertGweiToWei(big.NewInt(1))
			gasToSend := new(big.Int).Mul(gasPrice, big.NewInt(int64(desiredGasLimit)))
			return gasToSend, nil
		},
	}

	// 1 transfer shouldn't go through to to SendETH failure and should be set to an error
	// 1 should return a 0 balance and remain pending
	// 1 should succeed.  CheckGasTransactions will cause this transaction to come back
	// as a success, and the status will change to confirmed

	// call method under test
	jobs.SendGasToTreasureAddresses()

	// call a method to check the status of the transactions, so they will
	// get transitioned to their next statuses
	jobs.CheckGasTransactions()

	pendingGas, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.GasPending})
	suite.Nil(err)
	suite.Equal(1, len(pendingGas))

	confirmed, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.GasConfirmed})
	suite.Nil(err)
	suite.Equal(1, len(confirmed))

	errored, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.GasError})
	suite.Nil(err)
	suite.Equal(1, len(errored))

	suite.Equal(true, hasCalledCalculateGasToSend)
	suite.Equal(true, hasCalledCheckETHBalance)
	suite.Equal(true, hasCalledSendGas)
}

func (suite *JobsSuite) Test_InvokeBury() {
	generateTreasuresToBury(suite, 3, models.GasConfirmed)
	waitingForBury, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.GasConfirmed})
	suite.Nil(err)
	suite.Equal(3, len(waitingForBury))

	hasCalledCheckETHBalance := false
	hasCalledCheckPRLBalance := false
	hasCalledBuryPRL := false
	hasCalledCheckBuryStatus := false

	failBuryPRLAddress := waitingForBury[0].ETHAddr
	notYetBuriedAddress := waitingForBury[1].ETHAddr

	jobs.EthWrapper = services.Eth{
		CheckETHBalance: func(addr common.Address) *big.Int {
			hasCalledCheckETHBalance = true
			return big.NewInt(600000000000000000)
		},
		CheckPRLBalance: func(addr common.Address) *big.Int {
			hasCalledCheckPRLBalance = true
			return big.NewInt(600000000000000000)
		},
		BuryPrl: func(msg services.OysterCallMsg) (bool, string, int64) {
			hasCalledBuryPRL = true
			// make one of the transfers unsuccessful
			if msg.From == services.StringToAddress(failBuryPRLAddress) {
				return false, "", -1
			}
			return true, "111111", 1
		},
		CheckBuriedState: func(address common.Address) (bool, error) {
			hasCalledCheckBuryStatus = true
			if address == services.StringToAddress(notYetBuriedAddress) {
				return false, nil
			}
			return true, nil
		},
		CreateSendPRLMessage:            services.EthWrapper.CreateSendPRLMessage,
		GeneratePublicKeyFromPrivateKey: services.EthWrapper.GeneratePublicKeyFromPrivateKey,
	}

	// 1 transfer shouldn't go through to to BuryPRL failure and should be set to an error
	// 1 shouldn't go through due to WaitForTransfer failure and should be set to an error
	// 1 should succeed.  CheckBuryTransactions will cause this transaction to come back
	// as a success, and the status will change to confirmed

	// call method under test
	jobs.InvokeBury()

	// call to transition transactions to their next statuses
	jobs.CheckBuryTransactions()

	waitingForBury, err = models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.GasConfirmed})
	suite.Nil(err)
	suite.Equal(0, len(waitingForBury))

	pending, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.BuryPending})
	suite.Nil(err)
	suite.Equal(1, len(pending))

	confirmed, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.BuryConfirmed})
	suite.Nil(err)
	suite.Equal(1, len(confirmed))

	errored, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.BuryError})
	suite.Nil(err)
	suite.Equal(1, len(errored))

	suite.Equal(true, hasCalledCheckPRLBalance)
	suite.Equal(true, hasCalledCheckETHBalance)
	suite.Equal(true, hasCalledBuryPRL)
	suite.Equal(true, hasCalledCheckBuryStatus)
}

func (suite *JobsSuite) Test_PurgeFinishedTreasure() {
	generateTreasuresToBury(suite, 3, models.BuryConfirmed)

	allTreasures, err := models.GetAllTreasuresToBury()
	suite.Nil(err)
	suite.Equal(3, len(allTreasures))

	for _, treasure := range allTreasures {
		suite.DB.ValidateAndCreate(&models.StoredGenesisHash{
			GenesisHash:    treasure.GenesisHash,
			FileSizeBytes:  10000,
			NumChunks:      10,
			WebnodeCount:   0,
			Status:         models.StoredGenesisHashUnassigned,
			TreasureStatus: models.TreasurePending,
		})
	}

	genHashesWithTreasureStatusPending := []models.StoredGenesisHash{}

	// verify there are 3 genesisHashes with treasure status TreasurePending
	err = suite.DB.RawQuery("SELECT * from stored_genesis_hashes "+
		"WHERE treasure_status = ?", models.TreasurePending).All(&genHashesWithTreasureStatusPending)
	suite.Nil(err)
	suite.Equal(3, len(genHashesWithTreasureStatusPending))

	jobs.PurgeFinishedTreasure()

	//allTreasures, err = models.GetAllTreasuresToBury()
	//suite.Nil(err)
	//suite.Equal(0, len(allTreasures))

	genHashesWithTreasureStatusPending = []models.StoredGenesisHash{}

	// verify that PurgeFinishedTreasure set these genesis hashes to treasure
	// status TreasureBuried
	err = suite.DB.RawQuery("SELECT * from stored_genesis_hashes "+
		"WHERE treasure_status = ?", models.TreasurePending).All(&genHashesWithTreasureStatusPending)
	suite.Nil(err)
	suite.Equal(0, len(genHashesWithTreasureStatusPending))
}

func generateTreasuresToBury(suite *JobsSuite, numToCreateOfEachStatus int, status models.PRLStatus) {
	prlAmount := big.NewInt(100000000000000000)
	for i := 0; i < numToCreateOfEachStatus; i++ {
		ethAddr, key, _ := services.EthWrapper.GenerateEthAddr()
		iotaAddr := oyster_utils.RandSeq(81, oyster_utils.TrytesAlphabet)
		iotaMessage := oyster_utils.RandSeq(10, oyster_utils.TrytesAlphabet)
		genesisHash := oyster_utils.RandSeq(64, []rune("abcdef123456789"))

		treasureToBury := models.Treasure{
			GenesisHash: genesisHash,
			PRLStatus:   status,
			ETHAddr:     ethAddr.Hex(),
			ETHKey:      key,
			Message:     iotaMessage,
			Address:     iotaAddr,
		}

		treasureToBury.SetPRLAmount(prlAmount)

		suite.DB.ValidateAndCreate(&treasureToBury)
	}
}

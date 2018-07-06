package jobs_test

import (
	"crypto/ecdsa"
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

	suite.False(hasCalledCheckPRLBalance)

	jobs.CheckPRLTransactions()

	pendingPRL, err = models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.PRLPending})
	suite.Nil(err)
	suite.Equal(0, len(pendingPRL))
	suite.True(hasCalledCheckPRLBalance)
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

	suite.False(hasCalledCheckETHBalance)

	jobs.CheckGasTransactions()

	pendingGas, err = models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.GasPending})
	suite.Nil(err)
	suite.Equal(0, len(pendingGas))
	suite.True(hasCalledCheckETHBalance)
}

func (suite *JobsSuite) Test_CheckBuryTransactions() {

	hasCalledCheckBuriedState := false

	jobs.EthWrapper = services.Eth{
		CheckBuriedState: func(addr common.Address) (bool, error) {
			hasCalledCheckBuriedState = true
			return true, nil
		},
	}

	generateTreasuresToBury(suite, 1, models.BuryPending)
	pending, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.BuryPending})
	suite.Nil(err)
	suite.Equal(1, len(pending))
	suite.False(hasCalledCheckBuriedState)

	suite.DB.ValidateAndCreate(&models.StoredGenesisHash{
		GenesisHash:    pending[0].GenesisHash,
		FileSizeBytes:  10000,
		NumChunks:      10,
		WebnodeCount:   0,
		Status:         models.StoredGenesisHashUnassigned,
		TreasureStatus: models.TreasurePending,
	})

	genHashesWithTreasureBurialPending := []models.StoredGenesisHash{}
	// verify there is 1 genesis hash with treasure status TreasurePending
	err = suite.DB.RawQuery("SELECT * from stored_genesis_hashes "+
		"WHERE treasure_status = ?", models.TreasurePending).All(&genHashesWithTreasureBurialPending)
	suite.Nil(err)
	suite.Equal(1, len(genHashesWithTreasureBurialPending))

	jobs.CheckBuryTransactions()

	pending, err = models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.BuryPending})
	suite.Nil(err)
	suite.Equal(0, len(pending))
	suite.True(hasCalledCheckBuriedState)

	genHashesWithTreasureBurialPending = []models.StoredGenesisHash{}
	err = suite.DB.RawQuery("SELECT * from stored_genesis_hashes "+
		"WHERE treasure_status = ?", models.TreasurePending).All(&genHashesWithTreasureBurialPending)
	suite.Nil(err)
	suite.Equal(0, len(genHashesWithTreasureBurialPending))
}

func (suite *JobsSuite) Test_SetTimedOutTransactionsToError() {

	generateTreasuresToBury(suite, 2, models.GasPending)
	generateTreasuresToBury(suite, 2, models.PRLPending)
	generateTreasuresToBury(suite, 2, models.BuryPending)
	generateTreasuresToBury(suite, 2, models.GasReclaimPending)

	pending, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{
		models.GasPending,
		models.PRLPending,
		models.BuryPending,
		models.GasReclaimPending})
	suite.Nil(err)
	suite.Equal(8, len(pending))

	errord, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{
		models.GasError,
		models.PRLError,
		models.BuryError,
		models.GasReclaimError})
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

	err = suite.DB.RawQuery("UPDATE treasures SET updated_at = ? WHERE eth_addr = ?",
		time.Now().Add(-24*time.Hour), pending[6].ETHAddr).All(&[]models.UploadSession{})
	suite.Nil(err)

	jobs.SetTimedOutTransactionsToError(time.Now().Add(-1 * time.Hour))

	pending, err = models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{
		models.GasPending,
		models.PRLPending,
		models.BuryPending,
		models.GasReclaimPending})
	suite.Nil(err)
	suite.Equal(4, len(pending))

	errord, err = models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{
		models.GasError,
		models.PRLError,
		models.BuryError,
		models.GasReclaimError})
	suite.Nil(err)
	suite.Equal(4, len(errord))
}

func (suite *JobsSuite) Test_StageTransactionsWithErrorsForRetry() {
	generateTreasuresToBury(suite, 1, models.GasError)
	generateTreasuresToBury(suite, 1, models.PRLError)
	generateTreasuresToBury(suite, 1, models.BuryError)
	generateTreasuresToBury(suite, 1, models.GasReclaimError)

	waiting, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{
		models.PRLWaiting,
		models.PRLConfirmed,
		models.GasConfirmed,
		models.BuryConfirmed})
	suite.Nil(err)
	suite.Equal(0, len(waiting))

	errord, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{
		models.GasError,
		models.PRLError,
		models.BuryError,
		models.GasReclaimError})
	suite.Nil(err)
	suite.Equal(4, len(errord))

	jobs.StageTransactionsWithErrorsForRetry()

	waiting, err = models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{
		models.PRLWaiting,
		models.PRLConfirmed,
		models.GasConfirmed,
		models.BuryConfirmed})
	suite.Nil(err)
	suite.Equal(4, len(waiting))

	errord, err = models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{
		models.GasError,
		models.PRLError,
		models.BuryError,
		models.GasReclaimError})
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

	suite.True(hasCalledCheckPRLBalance)
	suite.True(hasCalledSendPRL)
}

func (suite *JobsSuite) Test_SendGasToTreasureAddresses() {
	generateTreasuresToBury(suite, 3, models.PRLConfirmed)
	waitingForGas, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.PRLConfirmed})
	suite.Nil(err)
	suite.Equal(3, len(waitingForGas))

	hasCalledCalculateGasNeeded := false
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
		SendETH: func(fromAddress common.Address, fromPrivKey *ecdsa.PrivateKey, toAddress common.Address,
			gas *big.Int) (types.Transactions, string, int64, error) {
			hasCalledSendGas = true
			// make one of the transfers unsuccessful
			if toAddress == services.StringToAddress(failSendEthAddress) {
				return types.Transactions{}, "", -1, errors.New("FAIL")
			}
			return types.Transactions{}, "111111", 1, nil
		},
		GeneratePublicKeyFromPrivateKey: services.EthWrapper.GeneratePublicKeyFromPrivateKey,
		CalculateGasNeeded: func(desiredGasLimit uint64) (*big.Int, error) {
			hasCalledCalculateGasNeeded = true
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

	suite.True(hasCalledCalculateGasNeeded)
	suite.True(hasCalledCheckETHBalance)
	suite.True(hasCalledSendGas)
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

	suite.True(hasCalledCheckPRLBalance)
	suite.True(hasCalledCheckETHBalance)
	suite.True(hasCalledBuryPRL)
	suite.True(hasCalledCheckBuryStatus)
}

func (suite *JobsSuite) Test_CheckForReclaimableGas_not_worth_it_keep_waiting() {
	generateTreasuresToBury(suite, 1, models.BuryConfirmed)
	generateTreasuresToBury(suite, 1, models.GasReclaimPending)
	treasures, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.GasReclaimPending,
		models.BuryConfirmed})
	suite.Nil(err)
	suite.Equal(2, len(treasures))

	hasCalledCheckIfWorthReclaimingGas := false
	hasCalledReclaimGas := false

	jobs.EthWrapper = services.Eth{
		CheckIfWorthReclaimingGas: func(address common.Address,
			desiredGasLimit uint64) (bool, *big.Int, error) {
			hasCalledCheckIfWorthReclaimingGas = true
			return false, big.NewInt(5000), nil
		},
		ReclaimGas: func(address common.Address, privateKey *ecdsa.PrivateKey,
			gasToReclaim *big.Int) bool {
			SentToReclaimGas++
			hasCalledReclaimGas = true
			return true
		},
	}

	// we are setting it to not be worth it to reclaim gas, but we will still wait
	// in case network congestion dies down

	// call method under test
	jobs.CheckForReclaimableGas(time.Now().Add(-5 * time.Minute))

	waitingForGasReclaim, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.BuryConfirmed})
	suite.Nil(err)
	suite.Equal(1, len(waitingForGasReclaim))

	pendingGasReclaim, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.GasReclaimPending})
	suite.Nil(err)
	suite.Equal(1, len(pendingGasReclaim))

	gasReclaimSuccess, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.GasReclaimConfirmed})
	suite.Nil(err)
	suite.Equal(0, len(gasReclaimSuccess))

	suite.True(hasCalledCheckIfWorthReclaimingGas)
	suite.False(hasCalledReclaimGas)
}

func (suite *JobsSuite) Test_CheckForReclaimableGas_not_worth_it_stop_waiting() {
	generateTreasuresToBury(suite, 1, models.BuryConfirmed)
	generateTreasuresToBury(suite, 1, models.GasReclaimPending)
	treasures, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.GasReclaimPending,
		models.BuryConfirmed})
	suite.Nil(err)
	suite.Equal(2, len(treasures))

	hasCalledCheckIfWorthReclaimingGas := false
	hasCalledReclaimGas := false

	jobs.EthWrapper = services.Eth{
		CheckIfWorthReclaimingGas: func(address common.Address,
			desiredGasLimit uint64) (bool, *big.Int, error) {
			hasCalledCheckIfWorthReclaimingGas = true
			return false, big.NewInt(5000), nil
		},
		ReclaimGas: func(address common.Address, privateKey *ecdsa.PrivateKey,
			gasToReclaim *big.Int) bool {
			SentToReclaimGas++
			hasCalledReclaimGas = true
			return true
		},
	}

	// we are setting it to not be worth it to reclaim gas, and we have waited long enough for
	// network congestion to die down so we will just give up and set to success

	// call method under test
	jobs.CheckForReclaimableGas(time.Now().Add(5 * time.Minute))

	waitingForGasReclaim, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.BuryConfirmed})
	suite.Nil(err)
	suite.Equal(0, len(waitingForGasReclaim))

	reclaimPending, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.GasReclaimPending})
	suite.Nil(err)
	suite.Equal(0, len(reclaimPending))

	gasReclaimSuccess, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.GasReclaimConfirmed})
	suite.Nil(err)
	suite.Equal(2, len(gasReclaimSuccess))

	suite.True(hasCalledCheckIfWorthReclaimingGas)
	suite.False(hasCalledReclaimGas)
}

func (suite *JobsSuite) Test_CheckForReclaimableGas_worth_it() {
	generateTreasuresToBury(suite, 1, models.BuryConfirmed)
	generateTreasuresToBury(suite, 1, models.GasReclaimPending)
	treasures, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.GasReclaimPending,
		models.BuryConfirmed})
	suite.Nil(err)
	suite.Equal(2, len(treasures))

	hasCalledCheckIfWorthReclaimingGas := false
	hasCalledReclaimGas := false

	jobs.EthWrapper = services.Eth{
		CheckIfWorthReclaimingGas: func(address common.Address,
			desiredGasLimit uint64) (bool, *big.Int, error) {
			hasCalledCheckIfWorthReclaimingGas = true
			return true, big.NewInt(5000), nil
		},
		ReclaimGas: func(address common.Address, privateKey *ecdsa.PrivateKey,
			gasToReclaim *big.Int) bool {
			SentToReclaimGas++
			hasCalledReclaimGas = true
			return true
		},
	}

	// we are setting it to be worth it to reclaim gas.  One treasure still needs a reclaim started, so we will
	// start the reclaim on that one, and on the other one we won't do anything because it's still pending

	// call method under test
	jobs.CheckForReclaimableGas(time.Now().Add(-5 * time.Minute))

	waitingForGasReclaim, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.BuryConfirmed})
	suite.Nil(err)
	suite.Equal(0, len(waitingForGasReclaim))

	reclaimPending, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.GasReclaimPending})
	suite.Nil(err)
	suite.Equal(2, len(reclaimPending))

	gasReclaimSuccess, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.GasReclaimConfirmed})
	suite.Nil(err)
	suite.Equal(0, len(gasReclaimSuccess))

	suite.True(hasCalledCheckIfWorthReclaimingGas)
	suite.True(hasCalledReclaimGas)
}

func (suite *JobsSuite) Test_PurgeFinishedTreasure() {
	generateTreasuresToBury(suite, 3, models.GasReclaimConfirmed)

	allTreasures, err := models.GetAllTreasuresToBury()
	suite.Nil(err)
	suite.Equal(3, len(allTreasures))

	jobs.PurgeFinishedTreasure()

	//enable this part of the test when we are confident we don't need to
	//hold on to the treasures anymore
	//allTreasures, err = models.GetAllTreasuresToBury()
	//suite.Nil(err)
	//suite.Equal(0, len(allTreasures))
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

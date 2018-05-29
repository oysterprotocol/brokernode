package jobs_test

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gobuffalo/pop/nulls"
	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"github.com/pkg/errors"
	"math/big"
	"time"
)

func (suite *JobsSuite) Test_ProcessPaidSessions() {
	fileBytesCount := 500000

	// This map seems pointless but it makes the testing
	// in the for loop later on a bit simpler
	treasureIndexes := map[int]int{}
	treasureIndexes[5] = 5
	treasureIndexes[78] = 78
	treasureIndexes[199] = 199

	// create another dummy TreasureIdxMap for the data maps
	// who already have treasure buried
	testMap2 := `[{
		"sector": 1,
		"idx": 155,
		"key": "firstKeySecondMap"
		},
		{
		"sector": 2,
		"idx": 204,
		"key": "secondKeySecondMap"
		},
		{
		"sector": 3,
		"idx": 599,
		"key": "thirdKeySecondMap"
		}]`

	// create and start the upload session for the data maps that need treasure buried
	uploadSession1 := models.UploadSession{
		GenesisHash:    "abcdeff1",
		NumChunks:      500,
		FileSizeBytes:  fileBytesCount,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureInDataMapPending,
	}

	uploadSession1.StartUploadSession()
	mergedIndexes := []int{treasureIndexes[5], treasureIndexes[78], treasureIndexes[199]}
	privateKeys := []string{"0000000001", "0000000002", "0000000003"}

	uploadSession1.MakeTreasureIdxMap(mergedIndexes, privateKeys)

	// create and start the upload session for the data maps that already have buried treasure
	uploadSession2 := models.UploadSession{
		GenesisHash:    "abcdeff2",
		NumChunks:      500,
		FileSizeBytes:  fileBytesCount,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureInDataMapComplete,
		TreasureIdxMap: nulls.String{string(testMap2), true},
	}

	uploadSession2.StartUploadSession()

	// verify that we have successfully created all the data maps
	paidButUnburied := []models.DataMap{}
	err := suite.DB.Where("genesis_hash = ?", "abcdeff1").All(&paidButUnburied)
	suite.Equal(nil, err)

	paidAndBuried := []models.DataMap{}
	err = suite.DB.Where("genesis_hash = ?", "abcdeff2").All(&paidAndBuried)
	suite.Equal(nil, err)

	suite.NotEqual(0, len(paidButUnburied))
	suite.NotEqual(0, len(paidAndBuried))

	// verify that the "Message" field for every chunk in paidButUnburied is ""
	for _, dMap := range paidButUnburied {
		dMap.Message = "NOTEMPTY"
		suite.DB.ValidateAndSave(&dMap)
	}

	// verify that the "Status" field for every chunk in paidAndBuried is NOT Unassigned
	for _, dMap := range paidAndBuried {
		suite.NotEqual(models.Unassigned, dMap.Status)
		dMap.Message = "NOTEMPTY"
		suite.DB.ValidateAndSave(&dMap)
	}

	// call method under test
	jobs.ProcessPaidSessions(time.Now())

	paidButUnburied = []models.DataMap{}
	err = suite.DB.Where("genesis_hash = ?", "abcdeff1").All(&paidButUnburied)
	suite.Equal(nil, err)

	/* Verify the following:
	1.  If a chunk in paidButUnburied was one of the treasure chunks, Message is no longer ""
	2.  Status of all data maps in paidButUnburied is now Unassigned (to get picked up by process_unassigned_chunks
	*/
	for _, dMap := range paidButUnburied {
		if _, ok := treasureIndexes[dMap.ChunkIdx]; ok {
			suite.NotEqual("", dMap.Message)
		} else {
			suite.Equal("NOTEMPTY", dMap.Message)
		}
		suite.Equal(models.Unassigned, dMap.Status)
	}

	paidAndBuried = []models.DataMap{}
	err = suite.DB.Where("genesis_hash = ?", "abcdeff2").All(&paidAndBuried)
	suite.Equal(nil, err)

	// verify that all chunks in paidAndBuried have statuses changed to Unassigned
	for _, dMap := range paidAndBuried {
		suite.Equal(models.Unassigned, dMap.Status)
	}

	// get the session that was originally paid but unburied, and verify that all the
	// keys are now "" but that we still have a value for the Idx
	paidAndUnburiedSession := models.UploadSession{}
	err = suite.DB.Where("genesis_hash = ?", "abcdeff1").First(&paidAndUnburiedSession)
	suite.Equal(nil, err)

	treasureIndex, err := paidAndUnburiedSession.GetTreasureMap()
	suite.Equal(nil, err)

	suite.Equal(3, len(treasureIndex))

	for _, entry := range treasureIndex {
		_, ok := treasureIndexes[entry.Idx]
		suite.Equal(true, ok)
	}
}

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
	generateTreasuresToBury(suite, 3, models.PRLWaiting)
	waiting, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.PRLWaiting})
	suite.Nil(err)
	suite.Equal(3, len(waiting))

	// set one of the treasures to a higher PRL amount than what we have in our
	// balance
	waiting[0].SetPRLAmount(big.NewInt(700000000000000000))
	suite.DB.Save(&waiting)

	hasCalledGetGasPrice := false
	hasCalledCheckPRLBalance := false
	hasCalledSendPRL := false
	hasCalledWaitForTransfer := false

	jobs.EthWrapper = services.Eth{
		GetGasPrice: func() (*big.Int, error) {
			hasCalledGetGasPrice = true
			return big.NewInt(1), nil
		},
		CheckPRLBalance: func(addr common.Address) *big.Int {
			hasCalledCheckPRLBalance = true
			return big.NewInt(600000000000000000)
		},
		SendPRL: func(msg services.OysterCallMsg) bool {
			hasCalledSendPRL = true
			// make one of the transfers unsuccessful
			if msg.To == services.StringToAddress(waiting[1].ETHAddr) {
				return false
			}
			return true
		},
		GeneratePublicKeyFromPrivateKey: services.EthWrapper.GeneratePublicKeyFromPrivateKey,
		WaitForTransfer: func(addr common.Address) (result *big.Int, err error) {
			hasCalledWaitForTransfer = true
			return big.NewInt(200000000000000000), nil
		},
	}

	// 1 transfer shouldn't go through due to insufficient balance and should remain waiting
	// 1 transfer shouldn't go through to to SendPRL failure and should be set to an error
	// 1 should succeed.  Our WaitForTransfer mock will cause this transaction to come back
	// as a success, and the status will change to confirmed

	jobs.SendPRLsToWaitingTreasureAddresses()

	waiting, err = models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.PRLWaiting})
	suite.Nil(err)
	suite.Equal(1, len(waiting))

	confirmed, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.PRLConfirmed})
	suite.Nil(err)
	suite.Equal(1, len(confirmed))

	errored, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.PRLError})
	suite.Nil(err)
	suite.Equal(1, len(errored))

	suite.Equal(true, hasCalledGetGasPrice)
	suite.Equal(true, hasCalledCheckPRLBalance)
	suite.Equal(true, hasCalledSendPRL)
	suite.Equal(true, hasCalledWaitForTransfer)
}

func (suite *JobsSuite) Test_SendGasToTreasureAddresses() {
	generateTreasuresToBury(suite, 3, models.PRLConfirmed)
	waitingForGas, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.PRLConfirmed})
	suite.Nil(err)
	suite.Equal(3, len(waitingForGas))

	hasCalledGetGasPrice := false
	hasCalledCheckETHBalance := false
	hasCalledSendGas := false
	hasCalledWaitForTransfer := false

	failSendEthAddress := waitingForGas[0].ETHAddr
	failWaitForTransferAddress := waitingForGas[1].ETHAddr

	jobs.EthWrapper = services.Eth{
		GetGasPrice: func() (*big.Int, error) {
			hasCalledGetGasPrice = true
			return big.NewInt(1), nil
		},
		CheckETHBalance: func(addr common.Address) *big.Int {
			hasCalledCheckETHBalance = true
			return big.NewInt(600000000000000000)
		},
		SendETH: func(address common.Address, gas *big.Int) (types.Transactions, error) {
			hasCalledSendGas = true
			// make one of the transfers unsuccessful
			if address == services.StringToAddress(failSendEthAddress) {
				return types.Transactions{}, errors.New("FAIL")
			}
			return types.Transactions{}, nil
		},
		GeneratePublicKeyFromPrivateKey: services.EthWrapper.GeneratePublicKeyFromPrivateKey,
		WaitForTransfer: func(addr common.Address) (result *big.Int, err error) {
			hasCalledWaitForTransfer = true
			// make one of the wait for transfer calls unsuccessful
			if addr == services.StringToAddress(failWaitForTransferAddress) {
				return big.NewInt(0), errors.New("FAIL")
			}
			return big.NewInt(200000000000000000), nil
		},
	}

	// 1 transfer shouldn't go through to to SendETH failure and should be set to an error
	// 1 shouldn't go through due to WaitForTransfer failure and should be set to an error
	// 1 should succeed.  Our WaitForTransfer mock will cause this transaction to come back
	// as a success, and the status will change to confirmed

	jobs.SendGasToTreasureAddresses()

	waitingForGas, err = models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.PRLConfirmed})
	suite.Nil(err)
	suite.Equal(0, len(waitingForGas))

	confirmed, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.GasConfirmed})
	suite.Nil(err)
	suite.Equal(1, len(confirmed))

	errored, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.GasError})
	suite.Nil(err)
	suite.Equal(2, len(errored))

	suite.Equal(true, hasCalledGetGasPrice)
	suite.Equal(true, hasCalledCheckETHBalance)
	suite.Equal(true, hasCalledSendGas)
	suite.Equal(true, hasCalledWaitForTransfer)
}

func (suite *JobsSuite) Test_InvokeBury() {
	generateTreasuresToBury(suite, 3, models.GasConfirmed)
	waitingForBury, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.GasConfirmed})
	suite.Nil(err)
	suite.Equal(3, len(waitingForBury))

	hasCalledCheckETHBalance := false
	hasCalledCheckPRLBalance := false
	hasCalledBuryPRL := false
	hasCalledWaitForTransfer := false

	failBuryPRLAddress := waitingForBury[0].ETHAddr
	failWaitForTransferAddress := waitingForBury[1].ETHAddr

	jobs.EthWrapper = services.Eth{
		CheckETHBalance: func(addr common.Address) *big.Int {
			hasCalledCheckETHBalance = true
			return big.NewInt(600000000000000000)
		},
		CheckPRLBalance: func(addr common.Address) *big.Int {
			hasCalledCheckPRLBalance = true
			return big.NewInt(600000000000000000)
		},
		BuryPrl: func(msg services.OysterCallMsg) bool {
			hasCalledBuryPRL = true
			// make one of the transfers unsuccessful
			if msg.To == services.StringToAddress(failBuryPRLAddress) {
				return false
			}
			return true
		},
		GeneratePublicKeyFromPrivateKey: services.EthWrapper.GeneratePublicKeyFromPrivateKey,
		WaitForTransfer: func(addr common.Address) (result *big.Int, err error) {
			hasCalledWaitForTransfer = true
			// make one of the wait for transfer calls unsuccessful
			if addr == services.StringToAddress(failWaitForTransferAddress) {
				return big.NewInt(0), errors.New("FAIL")
			}
			return big.NewInt(200000000000000000), nil
		},
	}

	// 1 transfer shouldn't go through to to BuryPRL failure and should be set to an error
	// 1 shouldn't go through due to WaitForTransfer failure and should be set to an error
	// 1 should succeed.  Our WaitForTransfer mock will cause this transaction to come back
	// as a success, and the status will change to confirmed

	jobs.InvokeBury()

	waitingForBury, err = models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.GasConfirmed})
	suite.Nil(err)
	suite.Equal(0, len(waitingForBury))

	confirmed, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.BuryConfirmed})
	suite.Nil(err)
	suite.Equal(1, len(confirmed))

	errored, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.BuryError})
	suite.Nil(err)
	suite.Equal(2, len(errored))

	suite.Equal(true, hasCalledCheckPRLBalance)
	suite.Equal(true, hasCalledCheckETHBalance)
	suite.Equal(true, hasCalledBuryPRL)
	suite.Equal(true, hasCalledWaitForTransfer)
}

func (suite *JobsSuite) Test_PurgeFinishedTreasure() {
	generateTreasuresToBury(suite, 3, models.BuryConfirmed)

	allTreasures, err := models.GetAllTreasuresToBury()
	suite.Nil(err)
	suite.Equal(3, len(allTreasures))

	jobs.PurgeFinishedTreasure()

	allTreasures, err = models.GetAllTreasuresToBury()
	suite.Nil(err)
	suite.Equal(0, len(allTreasures))
}

func generateTreasuresToBury(suite *JobsSuite, numToCreateOfEachStatus int, status models.PRLStatus) {
	prlAmount := big.NewInt(100000000000000000)
	for i := 0; i < numToCreateOfEachStatus; i++ {
		ethAddr, key, _ := services.EthWrapper.GenerateEthAddr()
		iotaAddr := oyster_utils.RandSeq(81, oyster_utils.TrytesAlphabet)
		iotaMessage := oyster_utils.RandSeq(10, oyster_utils.TrytesAlphabet)

		treasureToBury := models.Treasure{
			PRLStatus: status,
			ETHAddr:   ethAddr.Hex(),
			ETHKey:    key,
			Message:   iotaMessage,
			Address:   iotaAddr,
		}

		treasureToBury.SetPRLAmount(prlAmount)

		suite.DB.ValidateAndCreate(&treasureToBury)
	}
}

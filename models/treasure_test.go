package models_test

import (
	"encoding/hex"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/utils"
	"github.com/oysterprotocol/brokernode/utils/eth_gateway"
	"math/big"
	"time"
)

func (suite *ModelSuite) Test_GetTreasuresToBuryByPRLStatus() {

	numToCreate := 2

	generateTreasuresToBuryOfEachStatus(suite, numToCreate)

	waitingForPRL, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.PRLWaiting})
	suite.Nil(err)

	suite.Equal(numToCreate, len(waitingForPRL))

	allTreasures, err := models.GetAllTreasuresToBury()
	suite.Nil(err)

	suite.True(len(allTreasures) > len(waitingForPRL))
}

func (suite *ModelSuite) Test_GetTreasuresToBuryByPRLStatusAndUpdateTime() {

	numToCreate := 2

	generateTreasuresToBuryOfEachStatus(suite, numToCreate)

	waitingForPRL, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.PRLWaiting})
	suite.Nil(err)
	suite.Equal(numToCreate, len(waitingForPRL))

	// set first treasureToBury to be old
	err = suite.DB.RawQuery("UPDATE treasures SET updated_at = ? WHERE eth_addr = ?",
		time.Now().Add(-24*time.Hour), waitingForPRL[0].ETHAddr).All(&[]models.Treasure{})
	suite.Nil(err)

	result, err := models.GetTreasuresToBuryByPRLStatusAndUpdateTime([]models.PRLStatus{models.PRLWaiting},
		time.Now().Add(-1*time.Hour))
	suite.Nil(err)

	suite.Equal(numToCreate-1, len(result))
}

func (suite *ModelSuite) Test_Get_And_Set_PRL_Amount() {

	prlAmount := big.NewInt(5000000000000000000)

	ethAddr, key, _ := eth_gateway.EthWrapper.GenerateEthAddr()
	iotaAddr := oyster_utils.RandSeq(81, oyster_utils.TrytesAlphabet)
	iotaMessage := oyster_utils.RandSeq(10, oyster_utils.TrytesAlphabet)

	treasureToBury := models.Treasure{
		ETHAddr: ethAddr.Hex(),
		ETHKey:  key,
		Message: iotaMessage,
		Address: iotaAddr,
	}

	treasureToBury.SetPRLAmount(prlAmount)
	returnedPrlAmount := treasureToBury.GetPRLAmount()

	suite.Equal(prlAmount, returnedPrlAmount)
}

func (suite *ModelSuite) Test_EncryptAndDecryptEthPrivateKey() {

	ethKey := hex.EncodeToString([]byte("SOME_PRIVATE_KEY"))
	ethAddr, _, _ := eth_gateway.EthWrapper.GenerateEthAddr()
	iotaAddr := oyster_utils.RandSeq(81, oyster_utils.TrytesAlphabet)
	iotaMessage := oyster_utils.RandSeq(10, oyster_utils.TrytesAlphabet)

	treasureToBury := models.Treasure{
		ETHAddr: ethAddr.Hex(),
		ETHKey:  ethKey,
		Message: iotaMessage,
		Address: iotaAddr,
	}

	suite.DB.ValidateAndCreate(&treasureToBury)
	suite.NotEqual(ethKey, treasureToBury.ETHKey)

	decryptedKey := treasureToBury.DecryptTreasureEthKey()
	suite.Equal(ethKey, decryptedKey)
}

func (suite *ModelSuite) Test_GetTreasuresToBuryBySignedStatus() {
	makeOneTreasureOfEachSignedStatus(suite)

	allSignedStatuses := []models.SignedStatus{}

	for key := range models.SignedStatusMap {
		allSignedStatuses = append(allSignedStatuses, key)
	}

	suite.NotEqual(0, len(allSignedStatuses))
	suite.Equal(len(models.SignedStatusMap), len(allSignedStatuses))

	treasures, err := models.GetTreasuresToBuryBySignedStatus(allSignedStatuses)
	suite.Nil(err)

	// we created one of each signed status so make sure the number of
	// treasures matches the length of allSignedStatuses
	suite.Equal(len(allSignedStatuses), len(treasures))

	// test that it can pluck out all treasures of just one type
	treasures, err = models.GetTreasuresToBuryBySignedStatus([]models.SignedStatus{
		models.TreasureSigned})
	suite.Nil(err)

	suite.Equal(1, len(treasures))

	// change all treasures to the same type
	for _, treasure := range treasures {
		treasure.SignedStatus = models.TreasureSigned
		suite.DB.ValidateAndSave(&treasure)
	}

	startingTreasureLength := len(treasures)

	// for good measure, make the same call again and make sure the length of the
	// returned array is the length of all the treasures
	treasures, err = models.GetTreasuresToBuryBySignedStatus([]models.SignedStatus{
		models.TreasureSigned})
	suite.Nil(err)

	suite.Equal(startingTreasureLength, len(treasures))
}

func (suite *ModelSuite) Test_GetTreasuresByGenesisHashAndIndexes() {
	// make two sets of treasures so they will have different genesis hashes
	makeOneTreasureOfEachSignedStatus(suite)
	makeOneTreasureOfEachSignedStatus(suite)

	treasures := []models.Treasure{}
	genHashes := make(map[string]string)

	suite.DB.All(&treasures)

	for _, treasure := range treasures {
		genHashes[treasure.GenesisHash] = treasure.GenesisHash
		if len(genHashes) == 2 {
			break
		}
	}

	suite.Equal(2, len(genHashes))

	for genHash := range genHashes {
		// Check that we can retrieve 1 treasure and that the genesis hash and index is what we expect
		treasures, err := models.GetTreasuresByGenesisHashAndIndexes(genHash, []int{0})
		suite.Nil(err)
		suite.Equal(1, len(treasures))
		suite.True(treasures[0].GenesisHash == genHash && treasures[0].Idx == 0)

		// Check that we can retrieve multiple treasures and that the genesis hashes and indexes are
		// what we expect
		treasures, err = models.GetTreasuresByGenesisHashAndIndexes(genHash, []int{
			0, 2000000})
		suite.Nil(err)
		suite.Equal(2, len(treasures))
		suite.True(treasures[0].GenesisHash == genHash &&
			treasures[0].Idx == 0 || treasures[0].Idx == 2000000)
		suite.True(treasures[1].GenesisHash == genHash &&
			treasures[1].Idx == 0 || treasures[1].Idx == 2000000)
		suite.True(treasures[1].Idx != treasures[0].Idx)
	}
}

func generateTreasuresToBuryOfEachStatus(suite *ModelSuite, numToCreateOfEachStatus int) {
	allStatuses := []models.PRLStatus{
		models.PRLWaiting,
		models.PRLPending,
		models.PRLConfirmed,
		models.GasPending,
		models.GasConfirmed,
		models.BuryPending,
		models.BuryConfirmed,
		models.PRLError,
		models.GasError,
		models.BuryError,
	}

	for _, status := range allStatuses {
		generateTreasuresToBury(suite, numToCreateOfEachStatus, status)
	}
}

func generateTreasuresToBury(suite *ModelSuite, numToCreateOfEachStatus int, status models.PRLStatus) {
	prlAmount := big.NewInt(500000000000000000)

	genesisHash := oyster_utils.RandSeq(6, []rune("abcdef0123456789"))

	for i := 0; i < numToCreateOfEachStatus; i++ {
		ethAddr, key, _ := eth_gateway.EthWrapper.GenerateEthAddr()
		iotaAddr := oyster_utils.RandSeq(81, oyster_utils.TrytesAlphabet)
		iotaMessage := oyster_utils.RandSeq(10, oyster_utils.TrytesAlphabet)

		treasureToBury := models.Treasure{
			GenesisHash: genesisHash,
			PRLStatus:   status,
			ETHAddr:     ethAddr.Hex(),
			ETHKey:      key,
			Message:     iotaMessage,
			Address:     iotaAddr,
			Idx:         int64(i * 1000000),
		}

		treasureToBury.SetPRLAmount(prlAmount)

		vErr, err := suite.DB.ValidateAndCreate(&treasureToBury)
		suite.Nil(err)
		suite.False(vErr.HasAny())
	}
}

func makeOneTreasureOfEachSignedStatus(suite *ModelSuite) {
	generateTreasuresToBury(suite, len(models.SignedStatusMap), models.PRLWaiting)

	treasures := []models.Treasure{}
	suite.DB.All(&treasures)

	var i = 0
	for status := range models.SignedStatusMap {
		treasures[i].SignedStatus = status
		vErr, err := suite.DB.ValidateAndSave(&treasures[i])
		suite.Nil(err)
		suite.False(vErr.HasAny())
		i++
	}
}

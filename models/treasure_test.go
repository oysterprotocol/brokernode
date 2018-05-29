package models_test

import (
	"encoding/hex"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"math/big"
	"time"
)

func (ms *ModelSuite) Test_GetTreasuresToBuryByPRLStatus() {

	numToCreate := 2

	generateTreasuresToBuryOfEachStatus(ms, numToCreate)

	waitingForPRL, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.PRLWaiting})
	ms.Nil(err)

	ms.Equal(numToCreate, len(waitingForPRL))

	allTreasures, err := models.GetAllTreasuresToBury()
	ms.Nil(err)

	ms.Equal(true, len(allTreasures) > len(waitingForPRL))

	//ms.DB.RawQuery("DELETE * from treasures").All(&[]models.Treasure{})
}

func (ms *ModelSuite) Test_GetTreasuresToBuryByPRLStatusAndUpdateTime() {

	numToCreate := 2

	generateTreasuresToBuryOfEachStatus(ms, numToCreate)

	waitingForPRL, err := models.GetTreasuresToBuryByPRLStatus([]models.PRLStatus{models.PRLWaiting})
	ms.Nil(err)
	ms.Equal(numToCreate, len(waitingForPRL))

	// set first treasureToBury to be old
	err = ms.DB.RawQuery("UPDATE treasures SET updated_at = ? WHERE eth_addr = ?",
		time.Now().Add(-24*time.Hour), waitingForPRL[0].ETHAddr).All(&[]models.Treasure{})
	ms.Nil(err)

	result, err := models.GetTreasuresToBuryByPRLStatusAndUpdateTime([]models.PRLStatus{models.PRLWaiting},
		time.Now().Add(-1*time.Hour))
	ms.Nil(err)

	ms.Equal(numToCreate-1, len(result))
}

func (ms *ModelSuite) Test_Get_And_Set_PRL_Amount() {

	prlAmount := big.NewInt(5000000000000000000)

	ethAddr, key, _ := services.EthWrapper.GenerateEthAddr()
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

	ms.Equal(prlAmount, returnedPrlAmount)
}

func (ms *ModelSuite) Test_EncryptAndDecryptEthPrivateKey() {

	ethKey := hex.EncodeToString([]byte("SOME_PRIVATE_KEY"))
	ethAddr, _, _ := services.EthWrapper.GenerateEthAddr()
	iotaAddr := oyster_utils.RandSeq(81, oyster_utils.TrytesAlphabet)
	iotaMessage := oyster_utils.RandSeq(10, oyster_utils.TrytesAlphabet)

	treasureToBury := models.Treasure{
		ETHAddr: ethAddr.Hex(),
		ETHKey:  ethKey,
		Message: iotaMessage,
		Address: iotaAddr,
	}

	ms.DB.ValidateAndCreate(&treasureToBury)
	ms.Equal(false, ethKey == treasureToBury.ETHKey)

	decryptedKey := treasureToBury.DecryptTreasureEthKey()
	ms.Equal(true, ethKey == decryptedKey)
}

func generateTreasuresToBuryOfEachStatus(ms *ModelSuite, numToCreateOfEachStatus int) {
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
		generateTreasuresToBury(ms, numToCreateOfEachStatus, status)
	}
}

func generateTreasuresToBury(ms *ModelSuite, numToCreateOfEachStatus int, status models.PRLStatus) {
	prlAmount := big.NewInt(500000000000000000)
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

		ms.DB.ValidateAndCreate(&treasureToBury)
	}
}

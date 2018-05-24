package models_test

import (
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"math/big"
)

func (ms *ModelSuite) Test_GetTreasuresToBuryByPRLStatus() {

	numToCreate := 2

	generateTreasuresToBuryOfEachStatus(ms, numToCreate)

	waitingForPRL, err := models.GetTreasuresToBuryByPRLStatus(models.PRLWaiting)
	ms.Nil(err)

	ms.Equal(numToCreate, len(waitingForPRL))

	allTreasures, err := models.GetAllTreasuresToBury()
	ms.Nil(err)

	ms.Equal(true, len(allTreasures) > len(waitingForPRL))
}

func (ms *ModelSuite) Test_Get_And_Set_PRL_Amount() {

	prlAmount := big.NewInt(5000000000000000000)

	ethAddr, _, _ := services.EthWrapper.GenerateEthAddr()
	iotaAddr := oyster_utils.RandSeq(81, oyster_utils.TrytesAlphabet)
	iotaMessage := oyster_utils.RandSeq(10, oyster_utils.TrytesAlphabet)

	treasureToBury := models.Treasure{
		ETHAddr: ethAddr.Hex(),
		Message: iotaMessage,
		Address: iotaAddr,
	}

	treasureToBury.SetPRLAmount(prlAmount)
	returnedPrlAmount := treasureToBury.GetPRLAmount()

	ms.Equal(prlAmount, returnedPrlAmount)
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
		ethAddr, _, _ := services.EthWrapper.GenerateEthAddr()
		iotaAddr := oyster_utils.RandSeq(81, oyster_utils.TrytesAlphabet)
		iotaMessage := oyster_utils.RandSeq(10, oyster_utils.TrytesAlphabet)

		treasureToBury := models.Treasure{
			PRLStatus: status,
			ETHAddr:   ethAddr.Hex(),
			Message:   iotaMessage,
			Address:   iotaAddr,
		}

		treasureToBury.SetPRLAmount(prlAmount)

		ms.DB.ValidateAndSave(&treasureToBury)
	}
}

package jobs_test

import (
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
)

// TODO WRITE ALL THESE TESTS

func (suite *JobsSuite) Test_CheckOngoingGasTransactions() {

}

func generateTreasureClaims(suite *JobsSuite, numToCreateOfEachStatus int, prlClaimStatus models.PRLClaimStatus,
	gasStatus models.GasTransferStatus) {
	for i := 0; i < numToCreateOfEachStatus; i++ {
		ethAddrTreasure, keyTreasure, _ := services.EthWrapper.GenerateEthAddr()
		ethAddrReceiver, _, _ := services.EthWrapper.GenerateEthAddr()
		genesisHash := oyster_utils.RandSeq(64, []rune("abcdef123456789"))

		treasureToClaim := models.WebnodeTreasureClaim{
			GenesisHash:           genesisHash,
			ReceiverETHAddr:       ethAddrReceiver.Hex(),
			TreasureETHAddr:       ethAddrTreasure.Hex(),
			TreasureETHPrivateKey: keyTreasure,
			SectorIdx:             0,
			NumChunks:             100,
			StartingClaimClock:    0,
			ClaimPRLStatus:        prlClaimStatus,
			GasStatus:             gasStatus,
		}

		suite.DB.ValidateAndCreate(&treasureToClaim)
	}
}

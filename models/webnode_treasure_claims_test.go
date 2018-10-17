package models_test

import (
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/utils"
	"github.com/oysterprotocol/brokernode/utils/eth_gateway"
	"time"
)

func (ms *ModelSuite) Test_GetTreasureClaimsByGasAndPRLStatus() {
	generateWebnodeTreasureClaims(ms,
		models.PRLClaimSuccess,
		models.GasTransferSuccess,
		1,
		4)

	generateWebnodeTreasureClaims(ms,
		models.PRLClaimNotStarted,
		models.GasTransferProcessing,
		1,
		3)

	treasureClaimsTest1, err := models.GetTreasureClaimsByGasAndPRLStatus(
		models.GasTransferSuccess,
		models.PRLClaimSuccess)
	ms.Nil(err)
	ms.Equal(4, len(treasureClaimsTest1))

	treasureClaimsTest2, err := models.GetTreasureClaimsByGasAndPRLStatus(
		models.GasTransferProcessing,
		models.PRLClaimNotStarted)
	ms.Nil(err)
	ms.Equal(3, len(treasureClaimsTest2))

	treasureClaimsTest3, err := models.GetTreasureClaimsByGasAndPRLStatus(
		models.GasTransferProcessing,
		models.PRLClaimError)
	ms.Nil(err)
	ms.Equal(0, len(treasureClaimsTest3))
}

func (ms *ModelSuite) Test_GetTreasureClaimsByPRLStatus() {
	generateWebnodeTreasureClaims(ms,
		models.PRLClaimSuccess,
		models.GasTransferSuccess,
		1,
		4)

	generateWebnodeTreasureClaims(ms,
		models.PRLClaimNotStarted,
		models.GasTransferProcessing,
		1,
		3)

	treasureClaimsTest1, err := models.GetTreasureClaimsByPRLStatus(
		models.PRLClaimSuccess)
	ms.Nil(err)
	ms.Equal(4, len(treasureClaimsTest1))

	treasureClaimsTest2, err := models.GetTreasureClaimsByPRLStatus(
		models.PRLClaimNotStarted)
	ms.Nil(err)
	ms.Equal(3, len(treasureClaimsTest2))

	treasureClaimsTest3, err := models.GetTreasureClaimsByPRLStatus(
		models.PRLClaimError)
	ms.Nil(err)
	ms.Equal(0, len(treasureClaimsTest3))
}

func (ms *ModelSuite) Test_GetTreasureClaimsByGasStatus() {
	generateWebnodeTreasureClaims(ms,
		models.PRLClaimSuccess,
		models.GasTransferSuccess,
		1,
		4)

	generateWebnodeTreasureClaims(ms,
		models.PRLClaimNotStarted,
		models.GasTransferProcessing,
		1,
		3)

	treasureClaimsTest1, err := models.GetTreasureClaimsByGasStatus(
		models.GasTransferSuccess)
	ms.Nil(err)
	ms.Equal(4, len(treasureClaimsTest1))

	treasureClaimsTest2, err := models.GetTreasureClaimsByGasStatus(
		models.GasTransferProcessing)
	ms.Nil(err)
	ms.Equal(3, len(treasureClaimsTest2))

	treasureClaimsTest3, err := models.GetTreasureClaimsByGasStatus(
		models.GasTransferError)
	ms.Nil(err)
	ms.Equal(0, len(treasureClaimsTest3))
}

func (ms *ModelSuite) Test_GetTreasureClaimsWithTimedOutGasTransfers() {
	generateWebnodeTreasureClaims(ms,
		models.PRLClaimNotStarted,
		models.GasTransferProcessing,
		1,
		4)

	/* Pass in a time in the past, so none will have an updated_at time which is
	<= to the time passed in */
	treasureClaimsTest1, err := models.GetTreasureClaimsWithTimedOutGasTransfers(
		time.Now().Add(-1 * time.Minute))
	ms.Nil(err)
	ms.Equal(0, len(treasureClaimsTest1))

	/* Pass in a time in the future, so all of them will have an updated_at
	time which is <= to the time passed in */
	treasureClaimsTest2, err := models.GetTreasureClaimsWithTimedOutGasTransfers(
		time.Now().Add(1 * time.Minute))
	ms.Nil(err)
	ms.Equal(4, len(treasureClaimsTest2))
}

func (ms *ModelSuite) Test_GetTreasureClaimsWithTimedOutPRLClaims() {
	generateWebnodeTreasureClaims(ms,
		models.PRLClaimProcessing,
		models.GasTransferSuccess,
		1,
		4)

	/* Pass in a time in the past, so none will have an updated_at time which is
	<= to the time passed in */
	treasureClaimsTest1, err := models.GetTreasureClaimsWithTimedOutPRLClaims(
		time.Now().Add(-1 * time.Minute))
	ms.Nil(err)
	ms.Equal(0, len(treasureClaimsTest1))

	/* Pass in a time in the future, so all of them will have an updated_at
	time which is <= to the time passed in */
	treasureClaimsTest2, err := models.GetTreasureClaimsWithTimedOutPRLClaims(
		time.Now().Add(1 * time.Minute))
	ms.Nil(err)
	ms.Equal(4, len(treasureClaimsTest2))
}

func (ms *ModelSuite) Test_GetTreasureClaimsWithTimedOutGasReclaims() {
	generateWebnodeTreasureClaims(ms,
		models.PRLClaimSuccess,
		models.GasTransferLeftoversReclaimProcessing,
		1,
		4)

	/* Pass in a time in the past, so none will have an updated_at time which is
	<= to the time passed in */
	treasureClaimsTest1, err := models.GetTreasureClaimsWithTimedOutGasReclaims(
		time.Now().Add(-1 * time.Minute))
	ms.Nil(err)
	ms.Equal(0, len(treasureClaimsTest1))

	/* Pass in a time in the future, so all of them will have an updated_at
	time which is <= to the time passed in */
	treasureClaimsTest2, err := models.GetTreasureClaimsWithTimedOutGasReclaims(
		time.Now().Add(1 * time.Minute))
	ms.Nil(err)
	ms.Equal(4, len(treasureClaimsTest2))
}

func (ms *ModelSuite) Test_DeleteCompletedTreasureClaims() {
	generateWebnodeTreasureClaims(ms,
		models.PRLClaimSuccess,
		models.GasTransferLeftoversReclaimSuccess,
		1,
		4)

	models.DeleteCompletedTreasureClaims()

	treasureClaims := []models.WebnodeTreasureClaim{}

	err := ms.DB.RawQuery("SELECT * FROM " +
		"webnode_treasure_claims").All(&treasureClaims)
	ms.Nil(err)
	ms.Equal(0, len(treasureClaims))
}

func generateWebnodeTreasureClaims(ms *ModelSuite,
	claimPRLStatus models.PRLClaimStatus,
	gasTransferStatus models.GasTransferStatus,
	startingClaimClock int64,
	numToGenerate int) {

	for i := 0; i < numToGenerate; i++ {
		treasureAddr, key, _ := eth_gateway.EthWrapper.GenerateEthAddr()
		receiverAddr, _, _ := eth_gateway.EthWrapper.GenerateEthAddr()

		validChars := []rune("abcde123456789")
		genesisHash := oyster_utils.RandSeq(64, validChars)

		webnodeTreasureClaim := models.WebnodeTreasureClaim{
			GenesisHash:           genesisHash,
			TreasureETHAddr:       treasureAddr.Hex(),
			TreasureETHPrivateKey: key,
			ReceiverETHAddr:       receiverAddr.Hex(),
			SectorIdx:             0,
			NumChunks:             100,
			ClaimPRLStatus:        claimPRLStatus,
			GasStatus:             gasTransferStatus,
			StartingClaimClock:    startingClaimClock,
		}

		vErr, err := ms.DB.ValidateAndCreate(&webnodeTreasureClaim)

		ms.False(vErr.HasAny())
		ms.Nil(err)
	}
}

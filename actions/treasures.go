package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/iotaledger/giota"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"time"
)

type TreasuresResource struct {
	buffalo.Resource
}

type treasureReq struct {
	ReceiverEthAddr string `json:"receiverEthAddr"`
	GenesisHash     string `json:"genesisHash"`
	SectorIdx       int    `json:"sectorIdx"`
	NumChunks       int    `json:"numChunks"`
	EthAddr         string `json:"ethAddr"`
	EthKey          string `json:"ethKey"`
}

type treasureRes struct {
	Success string `json:"success"`
}

const (
	POW_FAILURE = "PoW Failed"
)

// Verifies the treasure and claims such treasure.
func (t *TreasuresResource) VerifyAndClaim(c buffalo.Context) error {
	req := treasureReq{}
	oyster_utils.ParseReqBody(c.Request(), &req)

	addr := models.ComputeSectorDataMapAddress(req.GenesisHash, req.SectorIdx, req.NumChunks)
	iotaAddr := make([]giota.Address, len(addr))

	for i, address := range addr {
		iotaAddr[i] = giota.Address(address)
	}

	transactionsMap, err := services.FindTransactions(iotaAddr)

	if err != nil || !verifyAllPoW(iotaAddr, transactionsMap) {
		// indicate that PoW failure.
		return c.Render(200, r.JSON(treasureRes{Success: POW_FAILURE}))
	}

	return c.Render(200, r.JSON(treasureRes{
		Success: "true",
	}))

}

// Verify the Proof of work. Returns true if PoW is done. Otherwise, return false
func verifyAllPoW(iotaAddress []giota.Address, transactionsMap map[giota.Address][]giota.Transaction) bool {
	if len(transactionsMap) != len(iotaAddr) {
		return false
	}

	passedTimestamp := time.Now().AddDate(-1, 0, 0)

	for _, iotaAddress := range iotaAddr {
		if _, hasKey := transactionsMap[iotaAddress]; !hasKey {
			return false
		}

		transactions := transactionsMap[iotaAddress]
		// Check one the transactions has submit within the passed 1 year.
		isTransactionWithinTimePeriod := false
		for _, transaction := range transactions {
			if transaction.Timestamp.After(passedTimestamp) {
				isTransactionWithinTimePeriod = true
				break
			}
		}
		if !isTransactionWithinTimePeriod {
			return false
		}
	}
	return true
}

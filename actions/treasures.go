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
	EthKey          string `json:"EthKey"`
}

type treasureRes struct {
	Success string `json:"success"`
}

// Verifies the treasure and claims such treasure.
func (t *TreasuresResource) VerifyAndClaim(c buffalo.Context) error {
	req := treasureReq{}
	oyster_utils.ParseReqBody(c.Request(), &req)

	addr := models.ComputeSectorDataMapAddress(req.GenesisHash, req.SectorIdx, req.NumChunks)
	iotaAddr := make([]giota.Address, len(addr))

	for i, address := range addr {
		iotaAddr[i] = giota.Address(address)
	}

	transactionsMap := services.FindTransactions(iotaAddr)

	if len(transactionsMap) != len(iotaAddr) {
		// indicate that PoW failure.
	}

	passedTimestamp := time.Now().AddDate(-1, 0, 0)
	for _, iotaAddress := range iotaAddr {
		if _, hasKey := transactionsMap[iotaAddress]; !hasKey {
			// indicate that PoW failure
		}

		transactions := transactionsMap[iotaAddress]
		// Check all the transactions has submit within the passed 1 year.
		for _, transaction := range transactions {
			if !transaction.Timestamp.After(passedTimestamp) {
				// indicate that PoW failure
			}
		}
	}

	//ftr := &giota.FindTransactionsRequest{Bundles: []giota.Trytes{"DEXRPLKGBROUQMKCLMRPG9HFKCACDZ9AB9HOJQWERTYWERJNOYLW9PKLOGDUPC9DLGSUH9UHSKJOASJRU"}}
	//resp, err := api.FindTransactions(ftr)
	//
	//datamap1, vErr := models.GetDataMap(req.GenesisHash, req.NumChunks)
	//for _, d := range datamap1 {
	//
	//}
	// or

	//vErr, error := models.BuildDataMaps(req.GenesisHash, req.NumChunks)
	//datamap, err := models.GetDataMapByGenesisHashAndChunkIdx(req.GenesisHash, req.NumChunks)
	//for _, d := range datamap {
	//
	//}

	res := treasureRes{
		Success: "true",
	}

	return c.Render(200, r.JSON(res))
}

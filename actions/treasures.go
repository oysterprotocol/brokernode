package actions

import (
	"fmt"
	"os"
	"time"
	"strings"
	"github.com/gobuffalo/buffalo"
	"github.com/iotaledger/giota"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/utils"
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

func unixMilli(t time.Time) int64 {
	return t.Round(time.Millisecond).UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}

// Verifies the treasure and claims such treasure.
func (t *TreasuresResource) VerifyAndClaim(c buffalo.Context) error {
	req := treasureReq{}
	oyster_utils.ParseReqBody(c.Request(), &req)

	host_ip := os.Getenv("HOST_IP")
	provider := "http://" + host_ip + ":14265"
	api := giota.NewAPI(provider, nil)
	ftr := &giota.FindTransactionsRequest{Bundles: []giota.Trytes{"DEXRPLKGBROUQMKCLMRPG9HFKCACDZ9AB9HOJQWERTYWERJNOYLW9PKLOGDUPC9DLGSUH9UHSKJOASJRU"}}
	resp, err := api.FindTransactions(ftr)

	datamap1, vErr := models.GetDataMap(req.GenesisHash, req.NumChunks)
	for _, d := range datamap1 {
		
	}
	// or 
	
	vErr, error := models.BuildDataMaps(req.GenesisHash, req.NumChunks)
	datamap, err := models.GetDataMapByGenesisHashAndChunkIdx(req.GenesisHash, req.NumChunks)
	for _, d := range datamap {
		
	}
	
	var transactions[2] int64
	transactions[0] = 11
	transactions[0] = 1122
	epoch := unixMilli(time.Now().AddDate(-1, 0, 0))
	verifyAttached := true
	for _, transaction := range transactions {
		if transaction <= epoch {
			verifyAttached = false
		}
	}

	res := treasureRes{
		Success:       "true"
	}

	return c.Render(202, r.JSON(res))
}
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
	Success bool `json:"success"`
}

var IotaWrapper = services.IotaWrapper

// Verifies the treasure and claims such treasure.
func (t *TreasuresResource) VerifyAndClaim(c buffalo.Context) error {
	req := treasureReq{}
	oyster_utils.ParseReqBody(c.Request(), &req)

	addr := models.ComputeSectorDataMapAddress(req.GenesisHash, req.SectorIdx, req.NumChunks)
	verify, err = IotaWrapper.verifyTreasure(addr)

	res := treasureRes{
		Success: verify,
	}

	return c.Render(200, r.JSON(res))
}

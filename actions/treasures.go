package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
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
	EthKey          string `json:"ethKey"`
}

type treasureRes struct {
	Success bool `json:"success"`
}

// Visible for Unit Test
var IotaWrapper = services.IotaWrapper
var EthWrapper = services.EthWrapper

// Verifies the treasure and claims such treasure.
func (t *TreasuresResource) VerifyAndClaim(c buffalo.Context) error {
	req := treasureReq{}
	oyster_utils.ParseReqBody(c.Request(), &req)

	addr := models.ComputeSectorDataMapAddress(req.GenesisHash, req.SectorIdx, req.NumChunks)
	verify, err := IotaWrapper.VerifyTreasure(addr)

	ethAddr, addrErr := EthWrapper.GenerateEthAddrFromPrivateKey(req.EthKey)

	if err == nil && verify && addrErr == nil {
		verify = EthWrapper.ClaimPRL(services.StringToAddress(req.ReceiverEthAddr), ethAddr, req.EthKey)
	}

	res := treasureRes{
		Success: verify,
	}

	return c.Render(200, r.JSON(res))
}

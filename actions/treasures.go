package actions

import (
	"errors"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gobuffalo/buffalo"
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
	EthKey          string `json:"ethKey"`
}

type treasureRes struct {
	Success bool `json:"success"`
}

// Verifies the treasure and claims such treasure.
func (t *TreasuresResource) VerifyAndClaim(c buffalo.Context) error {
	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramTreasuresResourceVerifyAndClaim, start)

	req := treasureReq{}
	oyster_utils.ParseReqBody(c.Request(), &req)

	addr := models.ComputeSectorDataMapAddress(req.GenesisHash, req.SectorIdx, req.NumChunks)
	verify, err := IotaWrapper.VerifyTreasure(addr)

	_, keyErr := crypto.HexToECDSA(req.EthKey)

	if err != nil {
		c.Error(400, err)
	}
	if keyErr != nil {
		c.Error(400, keyErr)
	}
	if !verify {
		res := treasureRes{
			Success: verify,
		}
		return c.Render(200, r.JSON(res))
	}

	ethAddr := EthWrapper.GenerateEthAddrFromPrivateKey(req.EthKey)

	startingClaimClock, claimClockErr := EthWrapper.CheckClaimClock(ethAddr)

	if startingClaimClock.Int64() == int64(0) {
		err = errors.New("claim clock should be 1 or a timestamp but received 0")
		c.Error(400, err)
	} else if claimClockErr != nil {
		c.Error(400, err)
	} else {
		webnodeTreasureClaim := models.WebnodeTreasureClaim{
			GenesisHash:           req.GenesisHash,
			SectorIdx:             req.SectorIdx,
			NumChunks:             req.NumChunks,
			ReceiverETHAddr:       req.ReceiverEthAddr,
			TreasureETHAddr:       ethAddr.String(),
			TreasureETHPrivateKey: req.EthKey,
			StartingClaimClock:    startingClaimClock.Int64(),
		}

		vErr, err := models.DB.ValidateAndCreate(&webnodeTreasureClaim)
		if len(vErr.Errors) > 0 {
			oyster_utils.LogIfError(errors.New(vErr.Error()), nil)
		}
		if err != nil {
			oyster_utils.LogIfError(err, nil)
		}

		verify = err == nil && len(vErr.Errors) == 0
	}

	res := treasureRes{
		Success: verify &&
			startingClaimClock.Int64() != int64(0) &&
			claimClockErr == nil,
	}

	return c.Render(200, r.JSON(res))
}

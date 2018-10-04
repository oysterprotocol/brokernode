package actions_v2

import (
	"errors"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gobuffalo/buffalo"
	"github.com/oysterprotocol/brokernode/actions/utils"
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
	if err := oyster_utils.ParseReqBody(c.Request(), &req); err != nil {
		err = fmt.Errorf("Invalid request, unable to parse request body  %v", err)
		c.Error(400, err)
		return err
	}

	if req.EthKey == os.Getenv("TEST_MODE_WALLET_KEY") {
		return c.Render(200, actions_utils.Render.JSON(treasureRes{
			Success: true,
		}))
	}

	addr := oyster_utils.ComputeSectorDataMapAddress(req.GenesisHash, req.SectorIdx, req.NumChunks)
	verify, err := IotaWrapper.VerifyTreasure(addr)

	_, keyErr := crypto.HexToECDSA(req.EthKey)

	if err != nil {
		c.Error(400, err)
	}
	if keyErr != nil {
		c.Error(400, keyErr)
	}
	if !verify {
		return c.Render(200, actions_utils.Render.JSON(treasureRes{
			Success: verify,
		}))
	}

	ethAddr := EthWrapper.GenerateEthAddrFromPrivateKey(req.EthKey)

	startingClaimClock, claimClockErr := EthWrapper.CheckClaimClock(ethAddr)

	if startingClaimClock.Int64() == int64(0) {
		c.Error(400, errors.New("claim clock should be 1 or a timestamp but received 0"))
	} else if claimClockErr != nil {
		c.Error(400, claimClockErr)
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

		oyster_utils.LogIfError(errors.New(vErr.Error()), nil)
		oyster_utils.LogIfError(err, nil)

		verify = err == nil && len(vErr.Errors) == 0
	}

	res := treasureRes{
		Success: verify &&
			startingClaimClock.Int64() != int64(0) &&
			claimClockErr == nil,
	}

	return c.Render(200, actions_utils.Render.JSON(res))
}

package services

import (
	"github.com/oysterprotocol/brokernode/models"
	"github.com/ethereum/go-ethereum/common"
)

var EthMock Eth

func init() {
	SetUpMock()
}

func SetUpMock() {

	EthMock = Eth{
		ClaimUnusedPRLs: func(receiverAddress common.Address, treasureAddress common.Address, treasurePrivateKey string) (bool) {
			return true
		},
		SendGas: func([]models.CompletedUpload) error {
			return nil
		},
	}
}

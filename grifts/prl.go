package grifts

import (
	"fmt"
	"math/big"

	"github.com/markbates/grift/grift"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
)

var _ = grift.Namespace("db", func() {

	grift.Desc("send_prl_seed", "Adds a 'treasure' that needs PRL")
	grift.Add("send_prl_seed", func(c *grift.Context) error {

		privateKey := "bf0e4b5b8bbe67a7028fe8dae65cb27c0a7732a9c39a90c9621e54b16ffaf122"
		address := services.EthWrapper.GenerateEthAddrFromPrivateKey(privateKey)
		prlAmount := big.NewFloat(float64(.0099))
		prlAmountInWei := oyster_utils.ConvertToWeiUnit(prlAmount)

		treasure := models.Treasure{
			ETHAddr:   address.Hex(),
			ETHKey:    privateKey,
			PRLStatus: models.PRLWaiting,
			Address:   "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
			Message:   "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		}

		treasure.SetPRLAmount(prlAmountInWei)

		vErr, err := models.DB.ValidateAndCreate(&treasure)

		if err == nil && len(vErr.Errors) == 0 {
			fmt.Println("Treasure row added")
		}

		return nil
	})

	grift.Desc("send_prl_remove", "Removes the 'treasure' that needs PRL")
	grift.Add("send_prl_remove", func(c *grift.Context) error {

		privateKey := "bf0e4b5b8bbe67a7028fe8dae65cb27c0a7732a9c39a90c9621e54b16ffaf122"
		address := services.EthWrapper.GenerateEthAddrFromPrivateKey(privateKey)

		err := models.DB.RawQuery("DELETE from treasures WHERE eth_addr = ?", address.Hex()).All(&[]models.Treasure{})

		if err == nil {
			fmt.Println("Treasure row deleted")
		}

		return nil
	})
})

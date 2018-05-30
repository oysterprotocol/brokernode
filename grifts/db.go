package grifts

import (
	"errors"
	"fmt"
	"github.com/markbates/grift/grift"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"math/big"
	"os"
)

var _ = grift.Namespace("db", func() {

	grift.Desc("seed", "Seeds a database")
	grift.Add("seed", func(c *grift.Context) error {
		// Add DB seeding stuff here
		return nil
	})

	grift.Desc("send_prl_seed", "Adds a 'treasure' that needs PRL")
	grift.Add("send_prl_seed", func(c *grift.Context) error {

		griftPrivateKey := os.Getenv("GRIFT_ETH_PRIVATE_KEY")
		if griftPrivateKey == "" {
			errorString := "you haven't specified an eth private key to use for this grift"
			fmt.Println(errorString)
			return errors.New(errorString)
		}
		address := services.EthWrapper.GenerateEthAddrFromPrivateKey(griftPrivateKey)
		prlAmount := big.NewFloat(float64(.0001))
		prlAmountInWei := oyster_utils.ConvertToWeiUnit(prlAmount)

		treasure := models.Treasure{
			ETHAddr: address.Hex(),
			ETHKey:  griftPrivateKey,
			Address: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
			Message: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
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

		griftPrivateKey := os.Getenv("GRIFT_ETH_PRIVATE_KEY")
		if griftPrivateKey == "" {
			errorString := "you haven't specified an eth private key to use for this grift"
			fmt.Println(errorString)
			return errors.New(errorString)
		}
		address := services.EthWrapper.GenerateEthAddrFromPrivateKey(griftPrivateKey)

		err := models.DB.RawQuery("DELETE from treasures WHERE eth_addr = ?", address.Hex()).All(&[]models.Treasure{})

		if err == nil {
			fmt.Println("Treasure row deleted")
		}

		return nil
	})

	grift.Desc("delete_uploads", "Removes any sessions or data_maps in the db")
	grift.Add("delete_uploads", func(c *grift.Context) error {

		models.DB.RawQuery("DELETE from upload_sessions").All(&[]models.UploadSession{})
		models.DB.RawQuery("DELETE from data_maps").All(&[]models.UploadSession{})

		return nil
	})
})

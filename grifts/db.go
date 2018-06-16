package grifts

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/markbates/grift/grift"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"math/big"
	"os"
)

func getAddress() (common.Address, string, error) {
	griftPrivateKey := os.Getenv("GRIFT_ETH_PRIVATE_KEY")
	if griftPrivateKey == "" {
		errorString := "you haven't specified an eth private key to use for this grift"
		fmt.Println(errorString)
		return services.StringToAddress(""), griftPrivateKey, errors.New(errorString)
	}
	address := services.EthWrapper.GenerateEthAddrFromPrivateKey(griftPrivateKey)
	return address, griftPrivateKey, nil
}

var _ = grift.Namespace("db", func() {

	grift.Desc("seed", "Seeds a database")
	grift.Add("seed", func(c *grift.Context) error {
		// Add DB seeding stuff here
		return nil
	})

	grift.Desc("send_prl_seed", "Adds a 'treasure' that needs PRL")
	grift.Add("send_prl_seed", func(c *grift.Context) error {

		address, griftPrivateKey, err := getAddress()
		if err != nil {
			fmt.Println(err)
			return err
		}

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

		address, _, err := getAddress()
		if err != nil {
			fmt.Println(err)
			return err
		}

		err = models.DB.RawQuery("DELETE from treasures WHERE eth_addr = ?", address.Hex()).All(&[]models.Treasure{})

		if err == nil {
			fmt.Println("Treasure row deleted")
		}

		return nil
	})

	grift.Desc("set_to_prl_waiting", "Stages treasure for PRL")
	grift.Add("set_to_prl_waiting", func(c *grift.Context) error {

		address, _, err := getAddress()
		if err != nil {
			fmt.Println(err)
			return err
		}

		treasureToBury := models.Treasure{}

		err = models.DB.RawQuery("SELECT * from treasures where eth_addr = ?", address.Hex()).First(&treasureToBury)

		if err == nil {
			fmt.Println("Found transaction!")
		}

		treasureToBury.PRLStatus = models.PRLWaiting

		vErr, err := models.DB.ValidateAndUpdate(&treasureToBury)
		if err == nil && len(vErr.Errors) == 0 {
			fmt.Println("Updated!")
		}

		return nil
	})

	grift.Desc("set_to_prl_confirmed", "Stages treasure for gas")
	grift.Add("set_to_prl_confirmed", func(c *grift.Context) error {

		address, _, err := getAddress()
		if err != nil {
			fmt.Println(err)
			return err
		}

		treasureToBury := models.Treasure{}

		err = models.DB.RawQuery("SELECT * from treasures where eth_addr = ?", address.Hex()).First(&treasureToBury)

		if err == nil {
			fmt.Println("Found transaction!")
		}

		treasureToBury.PRLStatus = models.PRLConfirmed

		vErr, err := models.DB.ValidateAndUpdate(&treasureToBury)
		if err == nil && len(vErr.Errors) == 0 {
			fmt.Println("Updated!")
		}

		return nil
	})

	grift.Desc("set_to_gas_confirmed", "Stages treasure for bury()")
	grift.Add("set_to_gas_confirmed", func(c *grift.Context) error {

		address, _, err := getAddress()
		if err != nil {
			fmt.Println(err)
			return err
		}

		treasureToBury := models.Treasure{}

		err = models.DB.RawQuery("SELECT * from treasures where eth_addr = ?", address.Hex()).First(&treasureToBury)

		if err == nil {
			fmt.Println("Found transaction!")
		}

		treasureToBury.PRLStatus = models.PRLConfirmed

		vErr, err := models.DB.ValidateAndUpdate(&treasureToBury)
		if err == nil && len(vErr.Errors) == 0 {
			fmt.Println("Updated!")
		}

		return nil
	})

	grift.Desc("print_treasure", "Prints the treasure you are testing with")
	grift.Add("print_treasure", func(c *grift.Context) error {

		address, _, err := getAddress()
		if err != nil {
			fmt.Println(err)
			return err
		}

		treasureToBury := models.Treasure{}

		err = models.DB.RawQuery("SELECT * from treasures where eth_addr = ?", address.Hex()).First(&treasureToBury)

		if err == nil {
			fmt.Println("ETH Address:  " + treasureToBury.ETHAddr)
			fmt.Println("ETH Key:      " + treasureToBury.ETHKey)
			fmt.Println("Iota Address: " + treasureToBury.Address)
			fmt.Println("Iota Message: " + treasureToBury.Message)
			fmt.Println("PRL Status:   " + models.PRLStatusMap[treasureToBury.PRLStatus])
			fmt.Println("PRL Amount:   " + treasureToBury.PRLAmount)
		} else {
			fmt.Println(err)
		}

		return nil
	})

	grift.Desc("delete_uploads", "Removes any sessions or data_maps in the db")
	grift.Add("delete_uploads", func(c *grift.Context) error {

		models.DB.RawQuery("DELETE from upload_sessions").All(&[]models.UploadSession{})
		models.DB.RawQuery("DELETE from data_maps").All(&[]models.DataMap{})

		return nil
	})

	grift.Desc("delete_genesis_hashes", "Delete all stored genesis hashes")
	grift.Add("delete_genesis_hashes", func(c *grift.Context) error {

		models.DB.RawQuery("DELETE from stored_genesis_hashes").All(&[]models.StoredGenesisHash{})

		return nil
	})

	grift.Desc("reset_genesis_hashes", "Resets all stored genesis hashes to webnode count 0 and status unassigned")
	grift.Add("reset_genesis_hashes", func(c *grift.Context) error {

		storedGenHashCount := models.StoredGenesisHash{}

		count, err := models.DB.RawQuery("SELECT COUNT(*) from stored_genesis_hashes").Count(&storedGenHashCount)

		if count == 0 {
			fmt.Println("No stored genesis hashes available!")
			return nil
		}

		err = models.DB.RawQuery("UPDATE stored_genesis_hashes SET webnode_count = ? AND status = ?",
			0, models.StoredGenesisHashUnassigned).All(&[]models.StoredGenesisHash{})

		if err == nil {
			fmt.Println("Successfully reset all stored genesis hashes!")
		} else {
			fmt.Println(err)
			return err
		}

		return nil
	})

	grift.Desc("add_brokernodes", "add some brokernode addresses to the db")
	grift.Add("add_brokernodes", func(c *grift.Context) error {

		qaBrokerIPs := []string{
			"52.14.218.135", "18.217.133.146",
		}

		hostIP := os.Getenv("HOST_IP")

		for _, qaBrokerIP := range qaBrokerIPs {
			if qaBrokerIP != hostIP {
				vErr, err := models.DB.ValidateAndCreate(&models.Brokernode{
					Address: "http://" + qaBrokerIP + ":3000",
				})
				if err != nil || len(vErr.Errors) != 0 {
					fmt.Println(err)
					fmt.Println(vErr)
					return err
				}
			}
		}

		fmt.Println("Successfully added brokernodes to database!")

		return nil
	})

	grift.Desc("delete_brokernodes", "delete all brokernode addresses from the db")
	grift.Add("delete_brokernodes", func(c *grift.Context) error {

		err := models.DB.RawQuery("DELETE from brokernodes").All(&[]models.Brokernode{})

		if err != nil {
			fmt.Println(err)
			return err
		}

		fmt.Println("Successfully deleted brokernodes from database!")
		return nil
	})
})

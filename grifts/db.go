package grifts

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/markbates/grift/grift"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"github.com/shopspring/decimal"
	"math/big"
	"os"
	"strconv"
	"time"
)

const qaTrytes = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
const qaGenHashStartingChars = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"

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

		var numToCreate int
		if len(c.Args) == 0 {
			numToCreate = 1
		} else {
			numToCreate, _ = strconv.Atoi(c.Args[0])
		}

		for i := 0; i < numToCreate; i++ {
			address, griftPrivateKey, err := services.EthWrapper.GenerateEthAddr()
			fmt.Println("PRIVATE KEY IS:")
			fmt.Println(griftPrivateKey)

			if err != nil {
				fmt.Println(err)
				return err
			}

			//prlAmount := big.NewFloat(float64(.0001))
			//prlAmountInWei := oyster_utils.ConvertToWeiUnit(prlAmount)
			prlAmountInWei := big.NewInt(7800000000000001)

			treasure := models.Treasure{
				ETHAddr: address.Hex(),
				ETHKey:  griftPrivateKey,
				Address: qaTrytes,
				Message: qaTrytes,
			}

			treasure.SetPRLAmount(prlAmountInWei)

			vErr, err := models.DB.ValidateAndCreate(&treasure)

			if err == nil && len(vErr.Errors) == 0 {
				fmt.Println("Treasure row added")
			}
		}

		return nil
	})

	grift.Desc("send_prl_remove", "Removes the 'treasure' that needs PRL")
	grift.Add("send_prl_remove", func(c *grift.Context) error {

		err := models.DB.RawQuery("DELETE from treasures WHERE address = ?", qaTrytes).All(&[]models.Treasure{})

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

		treasuresToBury := []models.Treasure{}

		err := models.DB.RawQuery("SELECT * from treasures where address = ?", qaTrytes).All(&treasuresToBury)

		if err == nil {
			for _, treasureToBury := range treasuresToBury {
				fmt.Println("ETH Address:  " + treasureToBury.ETHAddr)
				fmt.Println("ETH Key:      " + treasureToBury.ETHKey)
				fmt.Println("Iota Address: " + treasureToBury.Address)
				fmt.Println("Iota Message: " + treasureToBury.Message)
				fmt.Println("PRL Status:   " + models.PRLStatusMap[treasureToBury.PRLStatus])
				fmt.Println("PRL Amount:   " + treasureToBury.PRLAmount)
			}
		} else {
			fmt.Println(err)
		}

		return nil
	})

	grift.Desc("delete_uploads", "Removes any sessions or data_maps in the db")
	grift.Add("delete_uploads", func(c *grift.Context) error {

		models.DB.RawQuery("DELETE from upload_sessions").All(&[]models.UploadSession{})
		models.DB.RawQuery("DELETE from data_maps").All(&[]models.DataMap{})

		// Clean up KvStore
		services.RemoveAllKvStoreData()
		services.InitKvStore()
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

	grift.Desc("claim_unused_test", "Adds a completed upload "+
		"to claim unused PRLs from")
	grift.Add("claim_unused_test", func(c *grift.Context) error {

		var numToCreate int
		if len(c.Args) == 0 {
			numToCreate = 1
		} else {
			numToCreate, _ = strconv.Atoi(c.Args[0])
		}

		for i := 0; i < numToCreate; i++ {
			address, griftPrivateKey, err := services.EthWrapper.GenerateEthAddr()
			fmt.Println("PRIVATE KEY IS:")
			fmt.Println(griftPrivateKey)

			if err != nil {
				fmt.Println(err)
				return err
			}

			//prlAmountInWei := big.NewInt(7800000000000001)

			prlAmount := big.NewFloat(float64(.0001))
			prlAmountInWei := oyster_utils.ConvertToWeiUnit(prlAmount)

			callMsg, _ := services.EthWrapper.CreateSendPRLMessage(services.MainWalletAddress,
				services.MainWalletPrivateKey,
				address, *prlAmountInWei)

			sendSuccess, _, _ := services.EthWrapper.SendPRLFromOyster(callMsg)
			if sendSuccess {
				fmt.Println("Sent successfully!")
				for {
					fmt.Println("Polling for PRL arrival")
					balance := services.EthWrapper.CheckPRLBalance(address)
					if balance.Int64() > 0 {
						fmt.Println("PRL arrived!")
						break
					}
					time.Sleep(10 * time.Second)
				}
			}

			fmt.Println("Now making a completed_upload")

			validChars := []rune("abcde123456789")
			genesisHashEndingChars := oyster_utils.RandSeq(10, validChars)

			completedUpload := models.CompletedUpload{
				GenesisHash:   qaGenHashStartingChars + genesisHashEndingChars,
				ETHAddr:       address.String(),
				ETHPrivateKey: griftPrivateKey,
			}

			vErr, err := models.DB.ValidateAndSave(&completedUpload)
			completedUpload.EncryptSessionEthKey()

			if len(vErr.Errors) > 0 {
				err := errors.New("validation errors making completed upload!")
				fmt.Println(err)
				return err
			}
			if err != nil {
				fmt.Println(err)
				return err
			}
			fmt.Println("Successfully created a completed_upload!")
		}
		return nil
	})

	grift.Desc("print_completed_uploads", "Prints the completed uploads")
	grift.Add("print_completed_uploads", func(c *grift.Context) error {

		completedUploads := []models.CompletedUpload{}

		err := models.DB.RawQuery("SELECT * from completed_uploads").All(&completedUploads)

		if err == nil {
			for _, completedUpload := range completedUploads {
				fmt.Println("Genesis hash:      " + completedUpload.GenesisHash)
				fmt.Println("ETH Address:       " + completedUpload.ETHAddr)
				fmt.Println("ETH Key:           " + completedUpload.ETHPrivateKey)
				decrypted := completedUpload.DecryptSessionEthKey()
				fmt.Println("decrypted ETH Key: " + decrypted)
				fmt.Println("PRL Status:        " + models.PRLClaimStatusMap[completedUpload.PRLStatus])
				fmt.Println("Gas Status:        " + models.GasTransferStatusMap[completedUpload.GasStatus])
				fmt.Println("________________________________________________________")
			}
		} else {
			fmt.Println(err)
		}

		return nil
	})

	grift.Desc("delete_completed_uploads", "Deletes the completed uploads")
	grift.Add("delete_completed_uploads", func(c *grift.Context) error {

		err := models.DB.RawQuery("DELETE from completed_uploads WHERE genesis_hash " +
			"LIKE " + "'" + qaGenHashStartingChars + "%';").All(&[]models.CompletedUpload{})

		if err == nil {
			fmt.Println("Completed uploads deleted")
		}

		return nil
	})

	grift.Desc("claim_treasure_test", "Creates a treasure to claim and then tries to claim it")
	grift.Add("claim_treasure_test", func(c *grift.Context) error {

		var numToCreate int
		if len(c.Args) == 0 {
			numToCreate = 1
		} else {
			numToCreate, _ = strconv.Atoi(c.Args[0])
		}

		for i := 0; i < numToCreate; i++ {
			address, privateKey, err := services.EthWrapper.GenerateEthAddr()
			fmt.Println("PRIVATE KEY IS:")
			fmt.Println(privateKey)

			if err != nil {
				fmt.Println(err)
				return err
			}

			//prlAmount := big.NewFloat(float64(.0001))
			//prlAmountInWei := oyster_utils.ConvertToWeiUnit(prlAmount)
			prlAmountInWei := big.NewInt(7800000000000001)

			treasure := models.Treasure{
				ETHAddr: address.Hex(),
				ETHKey:  privateKey,
				Address: qaTrytes,
				Message: qaTrytes,
			}

			treasure.SetPRLAmount(prlAmountInWei)

			vErr, err := models.DB.ValidateAndCreate(&treasure)

			if err == nil && len(vErr.Errors) == 0 {
				fmt.Println("Treasure row added")
			}

			for {
				buried, err := services.EthWrapper.CheckBuriedState(address)
				if err != nil {
					fmt.Println("ERROR CHECKING BURIED STATE!")
					return err
				}
				if buried {
					fmt.Println("IT WAS BURIED!!")
					time.Sleep(30 * time.Second)
					fmt.Println("MOVING ON")
					break
				}
				time.Sleep(3 * time.Second)
			}
			validChars := []rune("abcde123456789")
			genesisHashEndingChars := oyster_utils.RandSeq(10, validChars)

			treasureToClaim := models.WebnodeTreasureClaim{
				GenesisHash:           qaGenHashStartingChars + genesisHashEndingChars,
				ReceiverETHAddr:       "0x5C77fd6bbCBa6b40e23d083E0c0844B1D43784F5",
				TreasureETHAddr:       address.Hex(),
				TreasureETHPrivateKey: privateKey,
				SectorIdx:             0,
				NumChunks:             100,
				StartingClaimClock:    0,
			}

			vErr, err = models.DB.ValidateAndCreate(&treasureToClaim)

			if len(vErr.Errors) == 0 && err == nil {
				fmt.Println("Created a webnode treasure claim!")
			}
		}

		return nil
	})

	grift.Desc("delete_webnode_treasure_claims", "Deletes webnode treasure claims")
	grift.Add("delete_webnode_treasure_claims", func(c *grift.Context) error {

		err := models.DB.RawQuery("DELETE from webnode_treasure_claims WHERE genesis_hash " +
			"LIKE " + "'" + qaGenHashStartingChars + "%';").All(&[]models.CompletedUpload{})

		if err == nil {
			fmt.Println("Treasure claims deleted")
		}

		return nil
	})

	grift.Desc("print_webnode_treasure_claims", "Prints the treasure claims")
	grift.Add("print_webnode_treasure_claims", func(c *grift.Context) error {

		treasureClaims := []models.WebnodeTreasureClaim{}

		err := models.DB.RawQuery("SELECT * from webnode_treasure_claims").All(&treasureClaims)

		if err == nil {
			fmt.Println("Printing treasure claims")
			for _, treasureClaim := range treasureClaims {
				fmt.Println("Genesis hash:          " + treasureClaim.GenesisHash)
				fmt.Println("Receiver ETH Address:  " + treasureClaim.ReceiverETHAddr)
				fmt.Println("Treasure ETH Address:  " + treasureClaim.TreasureETHAddr)
				fmt.Println("Treasure ETH Key:      " + treasureClaim.TreasureETHPrivateKey)
				decrypted := treasureClaim.DecryptTreasureEthKey()
				fmt.Println("decrypted ETH Key:     " + decrypted)
				fmt.Println("PRL Status:            " + models.PRLClaimStatusMap[treasureClaim.ClaimPRLStatus])
				fmt.Println("Gas Status:            " + models.GasTransferStatusMap[treasureClaim.GasStatus])
				fmt.Println("________________________________________________________")
			}
		} else {
			fmt.Println(err)
		}

		return nil
	})

	grift.Desc("set_webnode_treasure_claim_statuses", "Sets the PRL and/or Gas statuses")
	grift.Add("set_webnode_treasure_claim_statuses", func(c *grift.Context) error {

		claimPRLStatus, err := strconv.Atoi(c.Args[0])
		gasStatus, err := strconv.Atoi(c.Args[1])

		if claimPRLStatus != 0 {
			err := models.DB.RawQuery("UPDATE webnode_treasure_claims set claim_prl_status = ?"+
				" WHERE genesis_hash "+
				"LIKE "+"'"+qaGenHashStartingChars+"%';", claimPRLStatus).All(&[]models.CompletedUpload{})
			if err != nil {
				fmt.Println(err)
				return err
			}
			fmt.Println("Setting claim prl statuses to " + models.PRLClaimStatusMap[models.PRLClaimStatus(claimPRLStatus)])
		}

		if gasStatus != 0 {
			err := models.DB.RawQuery("UPDATE webnode_treasure_claims set gas_status = ?"+
				" WHERE genesis_hash "+
				"LIKE "+"'"+qaGenHashStartingChars+"%';", gasStatus).All(&[]models.CompletedUpload{})
			if err != nil {
				fmt.Println(err)
				return err
			}
			fmt.Println("Setting gas statuses to " + models.GasTransferStatusMap[models.GasTransferStatus(gasStatus)])
		}

		if err == nil {
			fmt.Println("Treasure claims statuses changed")
		}

		return nil
	})

	grift.Desc("test_broker_txs", "Tests check_alpha_payments and check_beta_payments")
	grift.Add("test_broker_txs", func(c *grift.Context) error {

		alphaAddr, key, _ := services.EthWrapper.GenerateEthAddr()
		betaAddr, betaKey, _ := services.EthWrapper.GenerateEthAddr()

		validChars := []rune("abcde123456789")
		genesisHashEndingCharsAlpha := oyster_utils.RandSeq(10, validChars)

		//totalCost := decimal.NewFromFloat(float64(0.015625))
		totalCost := decimal.NewFromFloat(float64(0.0002))
		float64Cost, _ := totalCost.Float64()
		bigFloatCost := big.NewFloat(float64Cost)
		totalCostInWei := oyster_utils.ConvertToWeiUnit(bigFloatCost)

		brokerTxAlpha := models.BrokerBrokerTransaction{
			GenesisHash:   qaGenHashStartingChars + genesisHashEndingCharsAlpha,
			Type:          models.SessionTypeAlpha,
			ETHAddrAlpha:  alphaAddr.Hex(),
			ETHAddrBeta:   betaAddr.Hex(),
			ETHPrivateKey: key,
			TotalCost:     totalCost,
			PaymentStatus: models.BrokerTxAlphaPaymentPending,
		}

		vErr, err := models.DB.ValidateAndCreate(&brokerTxAlpha)
		if len(vErr.Errors) > 0 {
			fmt.Println(vErr.Error())
			return errors.New(vErr.Error())
		}
		if err != nil {
			fmt.Println(err)
			return err
		}

		genesisHashEndingCharsBeta := oyster_utils.RandSeq(10, validChars)
		brokerTxBeta := models.BrokerBrokerTransaction{
			GenesisHash:   qaGenHashStartingChars + genesisHashEndingCharsBeta,
			Type:          models.SessionTypeBeta,
			ETHAddrAlpha:  alphaAddr.Hex(),
			ETHAddrBeta:   betaAddr.Hex(),
			ETHPrivateKey: betaKey,
			TotalCost:     totalCost,
			PaymentStatus: models.BrokerTxAlphaPaymentPending,
		}

		vErr, err = models.DB.ValidateAndCreate(&brokerTxBeta)
		if len(vErr.Errors) > 0 {
			fmt.Println(vErr.Error())
			return errors.New(vErr.Error())
		}
		if err != nil {
			fmt.Println(err)
			return err
		}

		callMsg, _ := services.EthWrapper.CreateSendPRLMessage(
			services.MainWalletAddress,
			services.MainWalletPrivateKey,
			services.StringToAddress(brokerTxAlpha.ETHAddrAlpha), *totalCostInWei)

		sendSuccess, _, _ := services.EthWrapper.SendPRLFromOyster(callMsg)

		if !sendSuccess {
			fmt.Println(err)
			return err
		}

		fmt.Println("Alpha is: " + alphaAddr.Hex())
		fmt.Println("Beta is: " + betaAddr.Hex())

		return nil
	})

	grift.Desc("print_broker_txs", "Prints broker_broker_transactions")
	grift.Add("print_broker_txs", func(c *grift.Context) error {

		brokerTxs := []models.BrokerBrokerTransaction{}

		err := models.DB.RawQuery("SELECT * from broker_broker_transactions").All(&brokerTxs)

		if err == nil {
			fmt.Println("Printing broker transactions")
			for _, brokerTx := range brokerTxs {
				fmt.Println("Genesis hash:           " + brokerTx.GenesisHash)
				fmt.Println("Alpha ETH Address:      " + brokerTx.ETHAddrAlpha)
				fmt.Println("ETH Key:           " + brokerTx.ETHPrivateKey)
				decrypted := brokerTx.DecryptEthKey()
				fmt.Println("decrypted ETH Key:      " + decrypted)
				fmt.Println("Beta ETH Address:       " + brokerTx.ETHAddrBeta)
				fmt.Println("Payment status:   " + models.PaymentStatusMap[brokerTx.PaymentStatus])
				fmt.Println("________________________________________________________")
			}
		} else {
			fmt.Println(err)
		}

		return nil
		return nil
	})

	grift.Desc("delete_broker_txs", "Deletes the broker_txs")
	grift.Add("delete_broker_txs", func(c *grift.Context) error {

		err := models.DB.RawQuery("DELETE from broker_broker_transactions WHERE genesis_hash " +
			"LIKE " + "'" + qaGenHashStartingChars + "%';").All(&[]models.BrokerBrokerTransaction{})

		if err == nil {
			fmt.Println("Broker_txs deleted")
		}

		return nil
	})
})

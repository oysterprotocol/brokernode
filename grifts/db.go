package grifts

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/markbates/grift/grift"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/utils"
	"github.com/oysterprotocol/brokernode/utils/eth_gateway"
	"github.com/shopspring/decimal"
	"math/big"
	"math/rand"
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
		return eth_gateway.StringToAddress(""), griftPrivateKey, errors.New(errorString)
	}
	address := eth_gateway.EthWrapper.GenerateEthAddrFromPrivateKey(griftPrivateKey)
	return address, griftPrivateKey, nil
}

func GenerateChunkRequestsForTests(indexToStopAt int, genesisHash string) []models.ChunkReq {
	chunkReqs := []models.ChunkReq{}

	for i := 1; i <= indexToStopAt; i++ {

		asciiValue := ""
		for i := 0; i < 5; i++ {
			asciiValue += string(rand.Intn(255))
		}

		req := models.ChunkReq{
			Idx:  i,
			Hash: genesisHash,
			Data: asciiValue,
		}

		chunkReqs = append(chunkReqs, req)
	}
	return chunkReqs
}

func SessionSetUpForTest(session *models.UploadSession, mergedIndexes []int,
	indexToStopAt int) []oyster_utils.ChunkData {
	session.StartUploadSession()
	privateKeys := []string{}

	for i := 0; i < len(mergedIndexes); i++ {
		key := ""
		for j := 0; j < 9; j++ {
			key += strconv.Itoa(rand.Intn(8) + 1)
		}
		privateKeys = append(privateKeys, key+strconv.Itoa(i))
	}

	session.PaymentStatus = models.PaymentStatusConfirmed
	models.DB.ValidateAndUpdate(session)
	session.MakeTreasureIdxMap(mergedIndexes, privateKeys)

	chunkReqs := GenerateChunkRequestsForTests(indexToStopAt, session.GenesisHash)

	models.ProcessAndStoreChunkData(chunkReqs, session.GenesisHash, mergedIndexes, oyster_utils.TestValueTimeToLive)

	session.WaitForAllHashes(100)

	session.CreateTreasures()

	for {
		finishedMessages, _ := session.WaitForAllMessages(3)
		if finishedMessages {
			break
		}
	}

	bulkKeys := oyster_utils.GenerateBulkKeys(session.GenesisHash, models.MetaDataChunkIdx, int64(session.NumChunks))
	bulkChunkData, _ := models.GetMultiChunkData(oyster_utils.InProgressDir, session.GenesisHash,
		bulkKeys)

	return bulkChunkData
}

var _ = grift.Namespace("db", func() {

	grift.Desc("upload_qa", "Adds a qa upload")
	grift.Add("upload_qa", func(c *grift.Context) error {

		numChunks := 1800

		uploadSession1 := models.UploadSession{
			GenesisHash:    oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
			NumChunks:      numChunks,
			FileSizeBytes:  3000,
			Type:           models.SessionTypeAlpha,
			PaymentStatus:  models.PaymentStatusConfirmed,
			TreasureStatus: models.TreasureInDataMapPending,
		}

		SessionSetUpForTest(&uploadSession1, []int{15}, uploadSession1.NumChunks)

		uploadSession1.WaitForAllHashes(500)
		uploadSession1.WaitForAllMessages(500)

		time.Sleep(3 * time.Second)

		session := models.UploadSession{}
		models.DB.Find(&session, uploadSession1.ID)

		fmt.Println(session.PaymentStatus)
		fmt.Println(session.TreasureStatus)
		fmt.Println(session.AllDataReady)

		return nil
	})

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
			address, griftPrivateKey, err := eth_gateway.EthWrapper.GenerateEthAddr()
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

		err := models.DB.RawQuery("DELETE FROM treasures WHERE address = ?", qaTrytes).All(&[]models.Treasure{})

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

		err = models.DB.RawQuery("SELECT * FROM treasures where eth_addr = ?", address.Hex()).First(&treasureToBury)

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

		err = models.DB.RawQuery("SELECT * FROM treasures where eth_addr = ?", address.Hex()).First(&treasureToBury)

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

		err = models.DB.RawQuery("SELECT * FROM treasures where eth_addr = ?", address.Hex()).First(&treasureToBury)

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

	grift.Desc("print_qa_treasure", "Prints the qa treasure you are testing with")
	grift.Add("print_qa_treasure", func(c *grift.Context) error {

		treasuresToBury := []models.Treasure{}

		err := models.DB.RawQuery("SELECT * FROM treasures where address = ?", qaTrytes).All(&treasuresToBury)

		if err == nil {
			for _, treasureToBury := range treasuresToBury {
				fmt.Println("________________________________________________________")
				fmt.Println("ETH Address:  " + treasureToBury.ETHAddr)
				fmt.Println("ETH Key:      " + treasureToBury.ETHKey)
				fmt.Println("Iota Address: " + treasureToBury.Address)
				fmt.Println("Iota Message: " + treasureToBury.Message)
				fmt.Println("PRL Status:   " + models.PRLStatusMap[treasureToBury.PRLStatus])
				fmt.Println("PRL Amount:   " + treasureToBury.PRLAmount)
				fmt.Println("________________________________________________________")
			}
		} else {
			fmt.Println(err)
		}

		return nil
	})

	grift.Desc("print_treasure", "Prints all the treasures")
	grift.Add("print_treasure", func(c *grift.Context) error {

		treasuresToBury := []models.Treasure{}

		err := models.DB.RawQuery("SELECT * from treasures ORDER BY created_at ASC").All(&treasuresToBury)

		if err == nil {
			for _, treasureToBury := range treasuresToBury {
				fmt.Println("________________________________________________________")
				fmt.Println("Genesis Hash:  " + treasureToBury.GenesisHash)
				fmt.Println("ETH Address:   " + treasureToBury.ETHAddr)
				fmt.Println("ETH Key:       " + treasureToBury.ETHKey)
				fmt.Println("Iota Address:  " + treasureToBury.Address)
				fmt.Println("Iota Message:  " + treasureToBury.Message)
				fmt.Println("PRL Status:    " + models.PRLStatusMap[treasureToBury.PRLStatus])
				fmt.Println("PRL Amount:    " + treasureToBury.PRLAmount)
				fmt.Println("________________________________________________________")
			}
		} else {
			fmt.Println(err)
		}

		return nil
	})

	grift.Desc("delete_uploads", "Removes any sessions or data_maps in the db")
	grift.Add("delete_uploads", func(c *grift.Context) error {

		allSessions := []models.UploadSession{}
		models.DB.RawQuery("select * from upload_sessions").All(&allSessions)

		for i := range allSessions {
			dbNameMessage := oyster_utils.GetBadgerDBName([]string{oyster_utils.InProgressDir,
				allSessions[i].GenesisHash, oyster_utils.MessageDir})
			dbNameHash := oyster_utils.GetBadgerDBName([]string{oyster_utils.InProgressDir,
				allSessions[i].GenesisHash, oyster_utils.HashDir})
			oyster_utils.RemoveAllUniqueKvStoreData(dbNameMessage)
			oyster_utils.RemoveAllUniqueKvStoreData(dbNameHash)
			oyster_utils.CloseUniqueKvStore(dbNameMessage)
			oyster_utils.CloseUniqueKvStore(dbNameHash)
		}

		models.DB.RawQuery("DELETE FROM upload_sessions").All(&[]models.UploadSession{})

		return nil
	})

	grift.Desc("delete_genesis_hashes", "Delete all stored genesis hashes")
	grift.Add("delete_genesis_hashes", func(c *grift.Context) error {

		models.DB.RawQuery("DELETE FROM stored_genesis_hashes").All(&[]models.StoredGenesisHash{})

		return nil
	})

	grift.Desc("delete_old_genesis_hashes", "Delete all old stored genesis hashes")
	grift.Add("delete_old_genesis_hashes", func(c *grift.Context) error {

		models.DB.RawQuery("DELETE FROM stored_genesis_hashes WHERE "+
			"TIMESTAMPDIFF(hour, create_at, NOW()) >= ?", 5).All(&[]models.StoredGenesisHash{})

		return nil
	})

	grift.Desc("delete_genesis_hash", "Delete a specific genesis hash")
	grift.Add("delete_genesis_hash", func(c *grift.Context) error {

		models.DB.RawQuery("DELETE from stored_genesis_hashes WHERE genesis_hash = ?", c.Args[0]).All(&[]models.StoredGenesisHash{})

		return nil
	})

	grift.Desc("reset_genesis_hashes", "Resets all stored genesis hashes to webnode count 0 and status unassigned")
	grift.Add("reset_genesis_hashes", func(c *grift.Context) error {

		err := models.DB.RawQuery("UPDATE stored_genesis_hashes SET webnode_count = ? AND status = ?",
			0, models.StoredGenesisHashUnassigned).All(&[]models.StoredGenesisHash{})

		if err == nil {
			fmt.Println("Successfully reset all stored genesis hashes!")
		} else {
			fmt.Println(err)
			return err
		}

		return nil
	})

	grift.Desc("print_genesis_hashes", "Prints the stored genesis hashes")
	grift.Add("print_genesis_hashes", func(c *grift.Context) error {

		storedGenHashes := []models.StoredGenesisHash{}

		models.DB.RawQuery("SELECT * from stored_genesis_hashes").All(&storedGenHashes)

		for _, genHash := range storedGenHashes {

			treasureStatus := "TreasureBuried"

			if genHash.TreasureStatus == 1 {
				treasureStatus = "TreasurePending"
			}

			assignedStatus := "StoredGenesisHashAssigned"
			if genHash.Status == 1 {
				assignedStatus = "StoredGenesisHashUnassigned"
			}

			fmt.Println("____________________________________________")
			fmt.Println("Genesis hash:       " + genHash.GenesisHash)
			fmt.Println("Treasure status:    " + treasureStatus)
			fmt.Println("Assigned status:    " + assignedStatus)
			fmt.Println("Webnode count:      " + strconv.Itoa(genHash.WebnodeCount))
			fmt.Println("____________________________________________")
		}

		return nil
	})

	grift.Desc("set_gen_hashes_to_buried", "set all genesis hashes to TreasureBuried")
	grift.Add("set_gen_hashes_to_buried", func(c *grift.Context) error {

		err := models.DB.RawQuery("UPDATE stored_genesis_hashes SET treasure_status = ?",
			models.TreasureBuried).All(&[]models.StoredGenesisHash{})

		if err != nil {
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
		rebelBrokerIPs := []string{
			"18.188.230.212", "18.188.64.13",
		}

		hostIP := os.Getenv("HOST_IP")

		if hostIP == "52.14.218.135" || hostIP == "18.217.133.146" {
			for _, brokerIP := range qaBrokerIPs {
				if brokerIP != hostIP {
					vErr, err := models.DB.ValidateAndCreate(&models.Brokernode{
						Address: "http://" + brokerIP + ":3000",
					})
					if err != nil || len(vErr.Errors) != 0 {
						fmt.Println(err)
						fmt.Println(vErr)
						return err
					}
				}
			}
		} else {
			for _, brokerIP := range rebelBrokerIPs {
				if brokerIP != hostIP {
					vErr, err := models.DB.ValidateAndCreate(&models.Brokernode{
						Address: "http://" + brokerIP + ":3000",
					})
					if err != nil || len(vErr.Errors) != 0 {
						fmt.Println(err)
						fmt.Println(vErr)
						return err
					}
				}
			}
		}

		fmt.Println("Successfully added brokernodes to database!")

		return nil
	})

	grift.Desc("print_brokernodes", "print brokernodes")
	grift.Add("print_brokernodes", func(c *grift.Context) error {

		brokers := []models.Brokernode{}

		models.DB.RawQuery("SELECT * FROM brokernodes").All(&brokers)

		for _, broker := range brokers {
			fmt.Println(broker.String())
		}
		return nil
	})

	grift.Desc("delete_brokernodes", "delete all brokernode addresses from the db")
	grift.Add("delete_brokernodes", func(c *grift.Context) error {

		err := models.DB.RawQuery("DELETE FROM brokernodes").All(&[]models.Brokernode{})

		if err != nil {
			fmt.Println(err)
			return err
		}

		fmt.Println("Successfully deleted brokernodes from database!")
		return nil
	})

	grift.Desc("print_data_maps", "Prints all data_maps")
	grift.Add("print_data_maps", func(c *grift.Context) error {

		allSessions := []models.UploadSession{}

		models.DB.RawQuery("SELECT * from upload_sessions").All(&allSessions)

		dataMapAll := []oyster_utils.ChunkData{}

		for i := range allSessions {
			keys := oyster_utils.GenerateBulkKeys(allSessions[i].GenesisHash, models.MetaDataChunkIdx,
				int64(allSessions[i].NumChunks))
			chunkData, err := models.GetMultiChunkData(oyster_utils.InProgressDir,
				allSessions[i].GenesisHash, keys)
			if err == nil && len(chunkData) != 0 {
				dataMapAll = append(dataMapAll, chunkData...)
			}
		}

		for _, dataMap := range dataMapAll {
			fmt.Println("____________________________________________")
			fmt.Println("Address:            " + dataMap.Address)
			fmt.Println("Genesis Hash:       " + dataMap.GenesisHash)
			fmt.Println("Message (tryted):   " + dataMap.Message)
			fmt.Println("Message (raw):      " + dataMap.RawMessage)
			fmt.Println("____________________________________________")
		}

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
			address, griftPrivateKey, err := eth_gateway.EthWrapper.GenerateEthAddr()
			fmt.Println("PRIVATE KEY IS:")
			fmt.Println(griftPrivateKey)

			if err != nil {
				fmt.Println(err)
				return err
			}

			//prlAmountInWei := big.NewInt(7800000000000001)

			prlAmount := big.NewFloat(float64(.0001))
			prlAmountInWei := oyster_utils.ConvertToWeiUnit(prlAmount)

			callMsg, _ := eth_gateway.EthWrapper.CreateSendPRLMessage(eth_gateway.MainWalletAddress,
				eth_gateway.MainWalletPrivateKey,
				address, *prlAmountInWei)

			sendSuccess, _, _ := eth_gateway.EthWrapper.SendPRLFromOyster(callMsg)
			if sendSuccess {
				fmt.Println("Sent successfully!")
				for {
					fmt.Println("Polling for PRL arrival")
					balance := eth_gateway.EthWrapper.CheckPRLBalance(address)
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

	grift.Desc("print_upload_sessions", "Prints the upload_sessions")
	grift.Add("print_upload_sessions", func(c *grift.Context) error {

		uploads := []models.UploadSession{}

		err := models.DB.RawQuery("SELECT * from upload_sessions").All(&uploads)

		if err == nil {

			for _, upload := range uploads {

				session := "Beta"
				if upload.Type == models.SessionTypeAlpha {
					session = "Alpha"
				}

				paymentStatus := "PaymentStatusInvoiced"

				switch upload.PaymentStatus {
				case models.PaymentStatusPending:
					paymentStatus = "PaymentStatusPending"
				case models.PaymentStatusConfirmed:
					paymentStatus = "PaymentStatusConfirmed"
				case models.PaymentStatusError:
					paymentStatus = "PaymentStatusError"
				}

				treasureStatus := "TreasureGeneratingKeys"

				switch upload.TreasureStatus {
				case models.TreasureInDataMapPending:
					treasureStatus = "TreasureInDataMapPending"
				case models.TreasureInDataMapComplete:
					treasureStatus = "TreasureInDataMapComplete"
				}

				allDataReady := "AllDataReady"
				switch upload.AllDataReady {

				case models.AllDataNotReady:
					allDataReady = "AllDataNotReady"
				}

				fmt.Println("________________________________________________________")
				fmt.Println("Type:                " + session)
				fmt.Println("Genesis hash:        " + upload.GenesisHash)
				fmt.Println("ETH Address Alpha:   " + upload.ETHAddrAlpha.String)
				fmt.Println("ETH Address Beta:    " + upload.ETHAddrBeta.String)
				fmt.Println("ETH Key:             " + upload.ETHPrivateKey)
				decrypted := upload.DecryptSessionEthKey()
				fmt.Println("decrypted ETH Key:   " + decrypted)
				fmt.Println("Payment Status:      " + paymentStatus)
				fmt.Println("Treasure Status:     " + treasureStatus)
				fmt.Println("AllDataReady:        " + allDataReady)
				fmt.Print("NumChunks:           ")
				fmt.Println(upload.NumChunks)
				fmt.Print("NextIdxToAttach:     ")
				fmt.Println(upload.NextIdxToAttach)
				fmt.Print("NextIdxToVerify:     ")
				fmt.Println(upload.NextIdxToVerify)
				fmt.Println("________________________________________________________")
			}
		} else {
			fmt.Println(err)
		}

		return nil
	})

	grift.Desc("print_completed_uploads", "Prints the completed uploads")
	grift.Add("print_completed_uploads", func(c *grift.Context) error {

		completedUploads := []models.CompletedUpload{}

		err := models.DB.RawQuery("SELECT * FROM completed_uploads").All(&completedUploads)

		if err == nil {
			for _, completedUpload := range completedUploads {
				fmt.Println("________________________________________________________")
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

		err := models.DB.RawQuery("DELETE FROM completed_uploads WHERE genesis_hash " +
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
			address, privateKey, err := eth_gateway.EthWrapper.GenerateEthAddr()
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
				buried, err := eth_gateway.EthWrapper.CheckBuriedState(address)
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

	grift.Desc("delete_qa_webnode_treasure_claims", "Deletes QA webnode treasure claims")
	grift.Add("delete_qa_webnode_treasure_claims", func(c *grift.Context) error {

		err := models.DB.RawQuery("DELETE FROM webnode_treasure_claims WHERE genesis_hash " +
			"LIKE " + "'" + qaGenHashStartingChars + "%';").All(&[]models.CompletedUpload{})

		if err == nil {
			fmt.Println("Treasure claims deleted")
		}

		return nil
	})

	grift.Desc("delete_all_webnode_treasure_claims", "Deletes all webnode treasure claims")
	grift.Add("delete_all_webnode_treasure_claims", func(c *grift.Context) error {

		err := models.DB.RawQuery("DELETE from webnode_treasure_claims").All(&[]models.CompletedUpload{})

		if err == nil {
			fmt.Println("Treasure claims deleted")
		}

		return nil
	})

	grift.Desc("print_webnode_treasure_claims", "Prints the treasure claims")
	grift.Add("print_webnode_treasure_claims", func(c *grift.Context) error {

		treasureClaims := []models.WebnodeTreasureClaim{}

		err := models.DB.RawQuery("SELECT * FROM webnode_treasure_claims").All(&treasureClaims)

		if err == nil {
			fmt.Println("Printing treasure claims")
			for _, treasureClaim := range treasureClaims {

				fmt.Println("________________________________________________________")
				fmt.Println("Genesis hash:           " + treasureClaim.GenesisHash)
				fmt.Println("Receiver ETH Address:   " + treasureClaim.ReceiverETHAddr)
				fmt.Println("Treasure ETH Address:   " + treasureClaim.TreasureETHAddr)
				decrypted := treasureClaim.DecryptTreasureEthKey()
				fmt.Println("decrypted ETH Key:      " + decrypted)
				fmt.Println("PRL Status:             " + models.PRLClaimStatusMap[treasureClaim.ClaimPRLStatus])
				fmt.Println("Gas Status:             " + models.GasTransferStatusMap[treasureClaim.GasStatus])
				fmt.Println("Starting Claim Clock:   " + big.NewInt(treasureClaim.StartingClaimClock).String())

				claimClock, _ := eth_gateway.EthWrapper.CheckClaimClock(
					eth_gateway.StringToAddress(treasureClaim.TreasureETHAddr))

				fmt.Println("Current Claim Clock:    " + claimClock.String())
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

		alphaAddr, key, _ := eth_gateway.EthWrapper.GenerateEthAddr()
		betaAddr, betaKey, _ := eth_gateway.EthWrapper.GenerateEthAddr()

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

		callMsg, _ := eth_gateway.EthWrapper.CreateSendPRLMessage(
			eth_gateway.MainWalletAddress,
			eth_gateway.MainWalletPrivateKey,
			eth_gateway.StringToAddress(brokerTxAlpha.ETHAddrAlpha), *totalCostInWei)

		sendSuccess, _, _ := eth_gateway.EthWrapper.SendPRLFromOyster(callMsg)

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

		err := models.DB.RawQuery("SELECT * FROM broker_broker_transactions").All(&brokerTxs)

		if err == nil {
			fmt.Println("Printing broker transactions")
			for _, brokerTx := range brokerTxs {

				sessionType := "Beta"
				if brokerTx.Type == models.SessionTypeAlpha {
					sessionType = "Alpha"
				}
				fmt.Println("________________________________________________________")
				fmt.Println("Session type:           " + sessionType)
				fmt.Println("Genesis hash:           " + brokerTx.GenesisHash)
				fmt.Println("Alpha ETH Address:      " + brokerTx.ETHAddrAlpha)
				fmt.Println("Beta ETH Address:       " + brokerTx.ETHAddrBeta)
				fmt.Println("ETH Key:                " + brokerTx.ETHPrivateKey)
				decrypted := brokerTx.DecryptEthKey()
				fmt.Println("decrypted ETH Key:      " + decrypted)
				fmt.Println("Payment status:         " + models.PaymentStatusMap[brokerTx.PaymentStatus])
				fmt.Println("________________________________________________________")
			}
		} else {
			fmt.Println(err)
		}

		return nil
		return nil
	})

	grift.Desc("delete_qa_broker_txs", "Deletes the qa broker_txs")
	grift.Add("delete_qa_broker_txs", func(c *grift.Context) error {

		err := models.DB.RawQuery("DELETE FROM broker_broker_transactions WHERE genesis_hash " +
			"LIKE " + "'" + qaGenHashStartingChars + "%';").All(&[]models.BrokerBrokerTransaction{})

		if err == nil {
			fmt.Println("Broker_txs deleted")
		}

		return nil
	})

	grift.Desc("delete_all_broker_txs", "Deletes all the broker_txs")
	grift.Add("delete_all_broker_txs", func(c *grift.Context) error {

		err := models.DB.RawQuery("DELETE from broker_broker_transactions").All(&[]models.BrokerBrokerTransaction{})

		if err == nil {
			fmt.Println("Broker_txs deleted")
		}

		return nil
	})

	/*
		grift.Desc("set_data_maps_to_unassigned", "Sets all data maps to unassigned")
		grift.Add("set_data_maps_to_unassigned", func(c *grift.Context) error {

		TODO:  One way to handle this would be to reset the NextIdxToAttachedAnd
		NextIdxToVerify back to their starting values.  But, if we are currently moving data maps
		to the completed database during the upload, this won't work because the broker will try again
		to upload chunks that no longer exist in its in progress database for that session.  This
		approach would only work if we only move all data maps to completed when the whole session is finished.

			allSessions := []models.UploadSession{}
			models.DB.RawQuery("select * from upload_sessions").All(&allSessions)

			for i := range allSessions {

			}

			return nil
		})
	*/

	grift.Desc("add_genesis_hashes_to_treasure", "Add gen hashes to treasures that need them")
	grift.Add("add_genesis_hashes_to_treasure", func(c *grift.Context) error {

		// the old treasures don't have genesis hashes.  This updates them.
		err := models.DB.RawQuery("UPDATE treasures, completed_data_maps "+
			"SET treasures.genesis_hash = completed_data_maps.genesis_hash "+
			"WHERE treasures.genesis_hash = ? "+
			"AND treasures.address = completed_data_maps.address", "").All(&[]models.Treasure{})

		if err == nil {
			fmt.Println("genesish_hash set for all treasures")
		} else {
			return err
		}

		return nil
	})

	grift.Desc("count_treasures_that_need_genesis_hash", "Counts the treasures that need a genesis hash")
	grift.Add("count_treasures_that_need_genesis_hash", func(c *grift.Context) error {

		treasuresWithGenesisHash := []models.Treasure{}
		treasuresWithoutGenesisHash := []models.Treasure{}

		// the old treasures don't have genesis hashes.  This updates them.
		errWith := models.DB.RawQuery("SELECT address from treasures WHERE genesis_hash = ?",
			"").All(&treasuresWithGenesisHash)

		errWithout := models.DB.RawQuery("SELECT address from treasures WHERE genesis_hash != ?",
			"").All(&treasuresWithoutGenesisHash)

		if errWith == nil && errWithout == nil {
			fmt.Println("With genesis hashes:     " + strconv.Itoa(len(treasuresWithGenesisHash)))
			fmt.Println("Without genesis hashes:  " + strconv.Itoa(len(treasuresWithoutGenesisHash)))
		} else {
			return errors.New(errWith.Error() + errWithout.Error())
		}

		return nil
	})

	grift.Desc("set_completed_uploads_for_deploy", "Set status of all completed uploads")
	grift.Add("set_completed_uploads_for_deploy", func(c *grift.Context) error {

		// the old completed_uploads will be missing these statuses.  This updates them.

		err := models.DB.RawQuery("UPDATE completed_uploads SET prl_status = ?, gas_status = ?",
			models.PRLClaimSuccess, models.GasTransferLeftoversReclaimSuccess).All(&[]models.CompletedUpload{})

		if err == nil {
			fmt.Println("statuses set for all completed_uploads")
		} else {
			return err
		}

		return nil
	})

	grift.Desc("create_completed_upload", "create a completed upload")
	grift.Add("create_completed_upload", func(c *grift.Context) error {

		address, privateKey, err := eth_gateway.EthWrapper.GenerateEthAddr()
		fmt.Println("PRIVATE KEY IS:")
		fmt.Println(privateKey)

		if err != nil {
			fmt.Println(err)
			return err
		}

		validChars := []rune("abcde123456789")
		genesisHashEndingChars := oyster_utils.RandSeq(10, validChars)

		completedUpload := models.CompletedUpload{
			GenesisHash:   qaGenHashStartingChars + genesisHashEndingChars,
			ETHAddr:       address.String(),
			ETHPrivateKey: privateKey,
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

		return nil
	})
})

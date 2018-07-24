package jobs

import (
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"fmt"
	"github.com/gobuffalo/pop"
	"github.com/iotaledger/giota"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"strconv"
	"strings"
	"time"
)

var (
	temp2Count = 0
	temp3Count = 0
	temp4Count = 0
	temp5Count = 0
	temp6Count = 0
)

func init() {

}

func CleanMySqlDB() {

	models.DB.Dialect.Details().Pool = 50

	fmt.Println("Starting cleaning process again")

	//createTransferAddresses()
	//createTempCompletedDataMaps()
	//deleteCompletedEntries()

	//time.Sleep(30 * time.Second)

	//CleanMySqlDB()

	//assignGenHashesToTreasures()
	getDistinctEntries()

	time.Sleep(5 * time.Second)

	CleanMySqlDB()
}

func getDistinctEntries() {

	completedDataMaps := []models.CompletedDataMap{}

	models.DB.RawQuery("SELECT DISTINCT genesis_hash FROM completed_data_maps").All(&completedDataMaps)

	time.Sleep(1 * time.Second)

	for i, completedDataMap := range completedDataMaps {

		//models.DB.Close()
		//time.Sleep(1000 * time.Millisecond)
		//fmt.Println(models.DB.Dialect.Details().Pool)

		time.Sleep(1500 * time.Millisecond)

		storedGenHashes := []models.StoredGenesisHash{}
		models.DB.RawQuery("SELECT * FROM stored_genesis_hashes WHERE "+
			"genesis_hash = ?", completedDataMap.GenesisHash).All(&storedGenHashes)

		temp2Count, err := models.DB.Where("genesis_hash = ?",
			completedDataMap.GenesisHash).Count(models.Temp2CompletedDataMap{})

		if len(storedGenHashes) > 0 && temp2Count == storedGenHashes[0].NumChunks {
			time.Sleep(500 * time.Millisecond)
			models.DB.RawQuery("DELETE from completed_data_maps WHERE genesis_hash = ?",
				completedDataMap.GenesisHash).All(&[]models.CompletedDataMap{})
			time.Sleep(500 * time.Millisecond)
			continue
		}

		fmt.Println("Doing " + strconv.Itoa(i) + " out of " + strconv.Itoa(len(completedDataMaps)))

		models.DB.RawQuery("SELECT COUNT(*) FROM stored_genesis_hashes WHERE "+
			"genesis_hash = ?", completedDataMap.GenesisHash).All(&storedGenHashes)

		distinctEntries := []models.CompletedDataMap{}

		time.Sleep(1 * time.Second)

		err = models.DB.RawQuery("SELECT DISTINCT hash, address, chunk_idx, message, genesis_hash FROM "+
			"completed_data_maps WHERE genesis_hash = ?", completedDataMap.GenesisHash).All(&distinctEntries)

		err = models.DB.Transaction(func(tx *pop.Connection) error {
			// Passed in the connection

			operation, _ := oyster_utils.CreateDbUpdateOperation(&models.Temp2CompletedDataMap{})
			columnNames := operation.GetColumns()
			var values []string

			insertionCount := 0

			for i := 0; i < len(distinctEntries); i++ {

				tempDataMap := models.Temp2CompletedDataMap{
					Hash:        distinctEntries[i].Hash,
					GenesisHash: completedDataMap.GenesisHash,
					ChunkIdx:    distinctEntries[i].ChunkIdx,
					Address:     distinctEntries[i].Address,
					Message:     distinctEntries[i].Message,
				}

				// We use INSERT SQL query rather than default Create method.
				tempDataMap.BeforeCreate(nil)

				// Validate the data
				vErr, err := tempDataMap.Validate(nil)
				if err != nil {
					fmt.Println("ERROR VALIDATING TEMP DATA MAP")
					fmt.Println(err)
					panic(err)
					return err
				}
				if len(vErr.Errors) > 0 {
					fmt.Println("ERROR VALIDATING TEMP DATA MAP")
					fmt.Println(vErr.Error())
					panic(vErr.Error())
					return errors.New(vErr.Error())
				}

				values = append(values, fmt.Sprintf("(%s)", operation.GetNewInsertedValue(tempDataMap)))

				insertionCount++
				if insertionCount >= 10 {
					time.Sleep(500 * time.Millisecond)
					err = insertsIntoTempCompletedDataMapsTable(columnNames, strings.Join(values, oyster_utils.COLUMNS_SEPARATOR), len(values))
					if err != nil {
						panic(err)
					}
					insertionCount = 0
					values = nil
				}
			}
			if len(values) > 0 {
				time.Sleep(500 * time.Millisecond)
				err = insertsIntoTempCompletedDataMapsTable(columnNames, strings.Join(values, oyster_utils.COLUMNS_SEPARATOR), len(values))
				if err != nil {
					panic(err)
				}
			}

			return nil
		})
	}
}

func assignGenHashesToTreasures() {

	treasures := []models.Treasure{}

	models.DB.RawQuery("select * from treasures where genesis_hash = ?", "").All(&treasures)

	for _, treasure := range treasures {
		temp2 := []models.Temp2CompletedDataMap{}
		models.DB.RawQuery("select * from temp_2_completed_data_maps where address = ?", treasure.Address).All(&temp2)

		if len(temp2) > 0 {
			fmt.Println("Found a genesis hash!")
			treasure.GenesisHash = temp2[0].GenesisHash
			models.DB.ValidateAndSave(&treasure)
		}
	}
}

func moveEntriesToTempTable() {

	completedDataMaps := []models.CompletedDataMap{}

	err := models.DB.Where("").Limit(1000).All(&completedDataMaps)
	if err != nil {
		fmt.Println("ERROR IN INITIAL QUERY")
		fmt.Println(err)
		panic(err)
		return
	}

	fails := 0
	successes := 0
	for _, completedDataMap := range completedDataMaps {
		vErr, err := models.DB.ValidateAndCreate(&models.Temp2CompletedDataMap{
			GenesisHash: completedDataMap.GenesisHash,
			ChunkIdx:    completedDataMap.ChunkIdx,
			Hash:        completedDataMap.Hash,
			Status:      completedDataMap.Status,
			Message:     completedDataMap.Message,
			Address:     completedDataMap.Address,
		})
		if err != nil {
			fmt.Println("ERROR IN TEMP COMPLETED DATA MAP CREATION")
			fmt.Println(err)
			panic(err)
		}
		if len(vErr.Errors) != 0 {
			fmt.Println("ERROR IN TEMP COMPLETED DATA MAP CREATION")
			fmt.Println(vErr.Error())
			panic(vErr.Error())
		}

		treasureMatch := []models.Treasure{}

		err = models.DB.Where("address = ", completedDataMap.Address).All(&treasureMatch)
		if err != nil {
			fmt.Println("ERROR FINDING A TREASURE WITH MATCHING ADDRESS")
			fmt.Println(err)
			panic(err)
		}

		treasureMatchWasFound := false

		if len(treasureMatch) > 0 {
			treasureMatchWasFound = true
			for _, treasure := range treasureMatch {
				treasure.GenesisHash = completedDataMap.GenesisHash
				vErr, err := models.DB.ValidateAndUpdate(&treasure)
				if err != nil {
					fmt.Println("ERROR UPDATING TREASURE WITH NEW GENESIS HASH")
					fmt.Println(err)
					panic(err)
				}
				if len(vErr.Errors) != 0 {
					fmt.Println("ERROR UPDATING TREASURE WITH NEW GENESIS HASH")
					fmt.Println(vErr.Error())
					panic(vErr.Error())
				}
			}
		}

		tempCompletedDataMaps := []models.Temp2CompletedDataMap{}

		err = models.DB.Where("address = ?", completedDataMap.Address).All(&tempCompletedDataMaps)
		if err != nil {
			fmt.Println("ERROR FINDING A TEMP COMPLETED DATA MAP WITH MATCHING ADDRESS")
			fmt.Println(err)
			panic(err)
		}

		treasureMatch2 := []models.Treasure{}
		if treasureMatchWasFound {

			err = models.DB.Where("address = ? AND genesis_hash != ?", completedDataMap.Address,
				completedDataMap.GenesisHash).All(&treasureMatch2)
			if err != nil {
				fmt.Println("ERROR FINDING A TREASURE WITH MATCHING ADDRESS BUT NOT MATCHING GENESIS HASH")
				fmt.Println(err)
				panic(err)
			}
		}

		if len(treasureMatch2) == 0 && len(tempCompletedDataMaps) > 0 {
			fmt.Println("SUCCESS(Treasures): NO TREASURES WITH THIS ADDRESS BUT NOT THIS GENESIS HASH")

			tempCompletedDataMapSuccess := false

			tempCompletedDataMapSuccess = tempCompletedDataMaps[0].GenesisHash == completedDataMap.GenesisHash
			tempCompletedDataMapSuccess = tempCompletedDataMapSuccess &&
				tempCompletedDataMaps[0].Hash == completedDataMap.Hash
			tempCompletedDataMapSuccess = tempCompletedDataMapSuccess &&
				tempCompletedDataMaps[0].Address == completedDataMap.Address
			tempCompletedDataMapSuccess = tempCompletedDataMapSuccess &&
				tempCompletedDataMaps[0].Message == completedDataMap.Message
			tempCompletedDataMapSuccess = tempCompletedDataMapSuccess &&
				tempCompletedDataMaps[0].ChunkIdx == completedDataMap.ChunkIdx
			tempCompletedDataMapSuccess = tempCompletedDataMapSuccess &&
				tempCompletedDataMaps[0].Status == completedDataMap.Status

			if tempCompletedDataMapSuccess {
				fmt.Println("SUCCESS:  temp_completed_data_map exists!")
				successes++
			} else {
				fmt.Println("FAIL:  no temp_completed_data_map exists")
				fails++
			}

		} else {
			fails++
		}
	}

	fmt.Println("TOTAL SUCCESSES: ")
	fmt.Println(strconv.Itoa(successes))
	fmt.Println("TOTAL FAILURES:")
	fmt.Println(strconv.Itoa(fails))
}

func createTransferAddresses() {

	transferGenHashes := []models.TransferGenHash{}

	genHashCount, err := models.DB.Where("complete = ?", 0).Count(models.TransferGenHash{})
	addressesToTransferCount, err := models.DB.Where("complete = ?", 0).Count(models.TransferAddress{})

	if err != nil {
		fmt.Println("ERROR COUNTING GEN HASHES")
		fmt.Println(err)
		panic(err)
		return
	}

	fmt.Println("genHashCount")
	fmt.Println(genHashCount)

	fmt.Println("addressesToTransferCount")
	fmt.Println(addressesToTransferCount)

	if addressesToTransferCount <= 5000 && genHashCount > 0 {
		err := models.DB.Where("complete = ?", 0).Limit(5).All(&transferGenHashes)
		if err != nil {
			fmt.Println("ERROR GETTING GEN HASHES WAITING COMPLETION")
			fmt.Println(err)
			panic(err)
			return
		}
		for _, genHash := range transferGenHashes {
			currHash := genHash.GenesisHash

			err = models.DB.Transaction(func(tx *pop.Connection) error {
				// Passed in the connection

				operation, _ := oyster_utils.CreateDbUpdateOperation(&models.TransferAddress{})
				columnNames := operation.GetColumns()
				var values []string

				insertionCount := 0

				for i := 0; i < genHash.NumChunks; i++ {

					obfuscatedHash := oyster_utils.HashHex(currHash, sha512.New384())
					currAddr := string(oyster_utils.MakeAddress(obfuscatedHash))

					transferAddress := models.TransferAddress{
						Hash:        currHash,
						GenesisHash: genHash.GenesisHash,
						ChunkIdx:    i,
						Address:     currAddr,
						Complete:    0,
					}

					// We use INSERT SQL query rather than default Create method.
					transferAddress.BeforeCreate(nil)

					// Validate the data
					vErr, err := transferAddress.Validate(nil)
					if err != nil {
						fmt.Println("ERROR VALIDATING TRANSFER ADDRESS")
						fmt.Println(err)
						panic(err)
						return err
					}
					if len(vErr.Errors) > 0 {
						fmt.Println("ERROR VALIDATING TRANSFER ADDRESS")
						fmt.Println(vErr.Error())
						panic(vErr.Error())
						return errors.New(vErr.Error())
					}

					values = append(values, fmt.Sprintf("(%s)", operation.GetNewInsertedValue(transferAddress)))

					currHash = oyster_utils.HashHex(currHash, sha256.New())

					insertionCount++
					if insertionCount >= 10 {
						err = insertsIntoTransferAddressTable(columnNames, strings.Join(values, oyster_utils.COLUMNS_SEPARATOR), len(values))
						if err != nil {
							panic(err)
						}
						insertionCount = 0
						values = nil
					}
				}
				if len(values) > 0 {
					err = insertsIntoTransferAddressTable(columnNames, strings.Join(values, oyster_utils.COLUMNS_SEPARATOR), len(values))
					if err != nil {
						panic(err)
					}
				}

				genHash.Complete = 1
				vErr, err := models.DB.ValidateAndSave(&genHash)
				if err != nil {
					fmt.Println("ERROR SAVING GEN HASH AS COMPLETE")
					fmt.Println(err)
					panic(err)
					return err
				}
				if len(vErr.Errors) > 0 {
					fmt.Println("ERROR SAVING GEN HASH AS COMPLETE")
					fmt.Println(vErr.Error())
					panic(vErr.Error())
					return errors.New("Validation errors changing gen hash to complete")
				}
				return nil
			})
		}
	}
	fmt.Println("Done with createTransferAddresses()")
}

func insertsIntoTransferAddressTable(columnsName string, values string, valueSize int) error {
	if len(values) == 0 {
		return nil
	}

	rawQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", "transfer_addresses", columnsName, values)
	var err error
	for i := 0; i < oyster_utils.MAX_NUMBER_OF_SQL_RETRY; i++ {
		err = models.DB.RawQuery(rawQuery).All(&[]models.TransferAddress{})
		if err == nil {
			break
		}
	}
	return err
}

func insertsIntoTempCompletedDataMapsTable(columnsName string, values string, valueSize int) error {
	if len(values) == 0 {
		return nil
	}

	rawQuery := fmt.Sprintf("REPLACE INTO %s (%s) VALUES %s", "temp_2_completed_data_maps", columnsName, values)
	var err error
	for i := 0; i < oyster_utils.MAX_NUMBER_OF_SQL_RETRY; i++ {
		err = models.DB.RawQuery(rawQuery).All(&[]models.Temp2CompletedDataMap{})
		if err == nil {
			break
		}
	}
	return err
}

func createTempCompletedDataMaps() {

	transferAddresses := []models.TransferAddress{}

	err := models.DB.Where("complete = ?", 0).All(&transferAddresses)
	if err != nil {
		fmt.Println("ERROR RETRIEVING TRANSFER ADDRESSES")
		fmt.Println(err)
		panic(err)
		return
	}

	fmt.Println("number of transfer addresses")
	fmt.Println(len(transferAddresses))

	fmt.Println(temp2Count)
	fmt.Println(temp3Count)
	fmt.Println(temp4Count)
	fmt.Println(temp5Count)
	fmt.Println(temp6Count)

	for i := 0; i < len(transferAddresses); i += services.MaxNumberOfAddressPerFindTransactionRequest {
		end := i + services.MaxNumberOfAddressPerFindTransactionRequest

		if end > len(transferAddresses) {
			end = len(transferAddresses)
		}

		if i >= end {
			break
		}

		addresses := make([]giota.Address, 0, len(transferAddresses[i:end]))

		for _, row := range transferAddresses[i:end] {

			address, err := giota.ToAddress(row.Address)
			if err != nil {
				fmt.Println("ERROR CONVERTING TO TRYTES")
				fmt.Println(err)
				panic(err)
				return
			}
			addresses = append(addresses, address)
		}

		transactionsMap, err := IotaWrapper.FindTransactions(addresses)
		if err != nil {
			fmt.Println("ERROR FINDING TRANSACTIONS")
			fmt.Println(err)
			panic(err)
			return
		}

		for _, row := range transferAddresses[i:end] {
			address, err := giota.ToAddress(row.Address)
			if err != nil {
				fmt.Println("ERROR CONVERTING TO TRYTES")
				fmt.Println(err)
				panic(err)
				return
			}
			transactions := transactionsMap[address]

			fmt.Println(address)

			if len(transactions) > 0 {
				if temp2Count < 2500000 {

					err = models.DB.Transaction(func(tx *pop.Connection) error {
						// Passed in the connection

						tempDataMap := models.Temp2CompletedDataMap{
							Address:     row.Address,
							Message:     string(transactions[0].SignatureMessageFragment),
							Hash:        row.Hash,
							GenesisHash: row.GenesisHash,
							ChunkIdx:    row.ChunkIdx,
							Status:      models.Complete,
						}

						vErr, err := tx.ValidateAndSave(&tempDataMap)
						if err != nil {
							fmt.Println("ERROR MAKING TEMP COMPLETED DATA MAP")
							fmt.Println(err)
							panic(err)
							return err
						}
						if len(vErr.Errors) > 0 {
							fmt.Println("ERROR MAKING TEMP COMPLETED DATA MAP")
							fmt.Println(vErr.Error())
							panic(vErr.Error())
							return errors.New("validation errors!")
						}
						temp2Count++
						row.Complete = 1
						vErr, err = tx.ValidateAndSave(&row)
						if err != nil {
							fmt.Println("ERROR SETTING ADDRESS TO TRANSFER TO TRUE")
							fmt.Println(err)
							panic(err)
							return err
						}
						if len(vErr.Errors) > 0 {
							fmt.Println("ERROR SETTING ADDRESS TO TRANSFER TO TRUE")
							fmt.Println(vErr.Error())
							panic(vErr.Error())
							return errors.New("validation errors!")
						}
						fmt.Println("CREATED TEMP DATA MAP!")
						return nil
					})
				} else if temp3Count < 2500000 {
					err = models.DB.Transaction(func(tx *pop.Connection) error {
						// Passed in the connection

						tempDataMap := models.Temp3CompletedDataMap{
							Address:     row.Address,
							Message:     string(transactions[0].SignatureMessageFragment),
							Hash:        row.Hash,
							GenesisHash: row.GenesisHash,
							ChunkIdx:    row.ChunkIdx,
							Status:      models.Complete,
						}

						vErr, err := tx.ValidateAndSave(&tempDataMap)
						if err != nil {
							fmt.Println("ERROR MAKING TEMP COMPLETED DATA MAP")
							fmt.Println(err)
							panic(err)
							return err
						}
						if len(vErr.Errors) > 0 {
							fmt.Println("ERROR MAKING TEMP COMPLETED DATA MAP")
							fmt.Println(vErr.Error())
							panic(vErr.Error())
							return errors.New("validation errors!")
						}
						temp3Count++
						row.Complete = 1
						vErr, err = tx.ValidateAndSave(&row)
						if err != nil {
							fmt.Println("ERROR SETTING ADDRESS TO TRANSFER TO TRUE")
							fmt.Println(err)
							panic(err)
							return err
						}
						if len(vErr.Errors) > 0 {
							fmt.Println("ERROR SETTING ADDRESS TO TRANSFER TO TRUE")
							fmt.Println(vErr.Error())
							panic(vErr.Error())
							return errors.New("validation errors!")
						}
						fmt.Println("CREATED TEMP DATA MAP!")
						return nil
					})

				} else if temp4Count < 2500000 {
					err = models.DB.Transaction(func(tx *pop.Connection) error {
						// Passed in the connection

						tempDataMap := models.Temp4CompletedDataMap{
							Address:     row.Address,
							Message:     string(transactions[0].SignatureMessageFragment),
							Hash:        row.Hash,
							GenesisHash: row.GenesisHash,
							ChunkIdx:    row.ChunkIdx,
							Status:      models.Complete,
						}

						vErr, err := tx.ValidateAndSave(&tempDataMap)
						if err != nil {
							fmt.Println("ERROR MAKING TEMP COMPLETED DATA MAP")
							fmt.Println(err)
							panic(err)
							return err
						}
						if len(vErr.Errors) > 0 {
							fmt.Println("ERROR MAKING TEMP COMPLETED DATA MAP")
							fmt.Println(vErr.Error())
							panic(vErr.Error())
							return errors.New("validation errors!")
						}
						temp4Count++
						row.Complete = 1
						vErr, err = tx.ValidateAndSave(&row)
						if err != nil {
							fmt.Println("ERROR SETTING ADDRESS TO TRANSFER TO TRUE")
							fmt.Println(err)
							panic(err)
							return err
						}
						if len(vErr.Errors) > 0 {
							fmt.Println("ERROR SETTING ADDRESS TO TRANSFER TO TRUE")
							fmt.Println(vErr.Error())
							panic(vErr.Error())
							return errors.New("validation errors!")
						}
						fmt.Println("CREATED TEMP DATA MAP!")
						return nil
					})

				} else if temp5Count < 2500000 {
					err = models.DB.Transaction(func(tx *pop.Connection) error {
						// Passed in the connection

						tempDataMap := models.Temp5CompletedDataMap{
							Address:     row.Address,
							Message:     string(transactions[0].SignatureMessageFragment),
							Hash:        row.Hash,
							GenesisHash: row.GenesisHash,
							ChunkIdx:    row.ChunkIdx,
							Status:      models.Complete,
						}

						vErr, err := tx.ValidateAndSave(&tempDataMap)
						if err != nil {
							fmt.Println("ERROR MAKING TEMP COMPLETED DATA MAP")
							fmt.Println(err)
							panic(err)
							return err
						}
						if len(vErr.Errors) > 0 {
							fmt.Println("ERROR MAKING TEMP COMPLETED DATA MAP")
							fmt.Println(vErr.Error())
							panic(vErr.Error())
							return errors.New("validation errors!")
						}
						temp5Count++
						row.Complete = 1
						vErr, err = tx.ValidateAndSave(&row)
						if err != nil {
							fmt.Println("ERROR SETTING ADDRESS TO TRANSFER TO TRUE")
							fmt.Println(err)
							panic(err)
							return err
						}
						if len(vErr.Errors) > 0 {
							fmt.Println("ERROR SETTING ADDRESS TO TRANSFER TO TRUE")
							fmt.Println(vErr.Error())
							panic(vErr.Error())
							return errors.New("validation errors!")
						}
						fmt.Println("CREATED TEMP DATA MAP!")
						return nil
					})

				} else if temp6Count < 2500000 {
					err = models.DB.Transaction(func(tx *pop.Connection) error {
						// Passed in the connection

						tempDataMap := models.Temp6CompletedDataMap{
							Address:     row.Address,
							Message:     string(transactions[0].SignatureMessageFragment),
							Hash:        row.Hash,
							GenesisHash: row.GenesisHash,
							ChunkIdx:    row.ChunkIdx,
							Status:      models.Complete,
						}

						vErr, err := tx.ValidateAndSave(&tempDataMap)
						if err != nil {
							fmt.Println("ERROR MAKING TEMP COMPLETED DATA MAP")
							fmt.Println(err)
							panic(err)
							return err
						}
						if len(vErr.Errors) > 0 {
							fmt.Println("ERROR MAKING TEMP COMPLETED DATA MAP")
							fmt.Println(vErr.Error())
							panic(vErr.Error())
							return errors.New("validation errors!")
						}
						temp6Count++
						row.Complete = 1
						vErr, err = tx.ValidateAndSave(&row)
						if err != nil {
							fmt.Println("ERROR SETTING ADDRESS TO TRANSFER TO TRUE")
							fmt.Println(err)
							panic(err)
							return err
						}
						if len(vErr.Errors) > 0 {
							fmt.Println("ERROR SETTING ADDRESS TO TRANSFER TO TRUE")
							fmt.Println(vErr.Error())
							panic(vErr.Error())
							return errors.New("validation errors!")
						}
						fmt.Println("CREATED TEMP DATA MAP!")
						return nil
					})
				}
			} else {
				fmt.Println("Did not find any transactions")
			}
		}
	}
	fmt.Println("Done with createTempCompletedDataMaps()")
}

func deleteCompletedEntries() {
	err :=
		models.DB.RawQuery("DELETE FROM transfer_gen_hashes WHERE complete = ?", 1).All(&[]models.TransferGenHash{})
	if err != nil {
		fmt.Println("ERROR DELETING COMPLETED transfer_gen_hashes")
		fmt.Println(err)
		panic(err)
		return
	}
	err =
		models.DB.RawQuery("DELETE FROM transfer_addresses WHERE complete = ?", 1).All(&[]models.TransferAddress{})
	if err != nil {
		fmt.Println("ERROR DELETING COMPLETED transfer_addresses")
		fmt.Println(err)
		panic(err)
		return
	}
	fmt.Println("Done with deleteCompletedEntries()")

}

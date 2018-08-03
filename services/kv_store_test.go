package services_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
)

// a very big number that will lead to ErrTxnTooBig for write.
const guessedMaxBatchSize = 200000

var testDBID = []string{"prefix", "genHash", "data"}

func Test_KVStore_Init(t *testing.T) {
	err := services.InitKvStore()

	oyster_utils.AssertNoError(err, t, "Could not create Badger DB")
	defer services.CloseKvStore()
}

func Test_KVStore_MassBatchSet(t *testing.T) {
	services.InitKvStore()
	defer services.CloseKvStore()

	err := services.BatchSet(getKvPairs(guessedMaxBatchSize), models.TestValueTimeToLive)
	oyster_utils.AssertNoError(err, t, "")

	kvs, _ := services.BatchGet(&services.KVKeys{strconv.Itoa(guessedMaxBatchSize - 1)})
	oyster_utils.AssertTrue(len(*kvs) == 1, t, "Expect only an item")
}

func Test_KVStoreBatchGet(t *testing.T) {
	services.InitKvStore()
	defer services.CloseKvStore()

	services.BatchSet(&services.KVPairs{"key": "oyster"}, models.TestValueTimeToLive)

	kvs, err := services.BatchGet(&services.KVKeys{"key"})
	oyster_utils.AssertNoError(err, t, "Could not get key")

	oyster_utils.AssertTrue(len(*kvs) == 1, t, "")
	oyster_utils.AssertStringEqual((*kvs)["key"], "oyster", t)
}

func Test_KVStoreBatchGet_WithMissingKey(t *testing.T) {
	services.InitKvStore()
	defer services.CloseKvStore()

	services.BatchSet(&services.KVPairs{"key": "oyster"}, models.TestValueTimeToLive)

	kvs, err := services.BatchGet(&services.KVKeys{"key", "unknownKey"})
	oyster_utils.AssertNoError(err, t, "Could not get key")

	oyster_utils.AssertTrue(len(*kvs) == 1, t, "")
	oyster_utils.AssertStringEqual((*kvs)["key"], "oyster", t)
}

func Test_KVStore_MassBatchGet(t *testing.T) {
	services.InitKvStore()
	defer services.CloseKvStore()

	err := services.BatchSet(getKvPairs(guessedMaxBatchSize), models.TestValueTimeToLive)
	oyster_utils.AssertNoError(err, t, "")

	kvs, _ := services.BatchGet(getKeys(guessedMaxBatchSize))
	oyster_utils.AssertTrue(len(*kvs) == guessedMaxBatchSize, t, "")
}

func Test_KVStoreBatchDelete(t *testing.T) {
	services.InitKvStore()
	defer services.CloseKvStore()

	services.BatchSet(&services.KVPairs{"key1": "oyster1", "key2": "oyster2"}, models.TestValueTimeToLive)

	err := services.BatchDelete(&services.KVKeys{"key1"})
	oyster_utils.AssertNoError(err, t, "Could not delete key")

	kvs, err := services.BatchGet(&services.KVKeys{"key1"})
	oyster_utils.AssertNoError(err, t, "Could complete get key")
	oyster_utils.AssertTrue(len(*kvs) == 0, t, "")
}

func Test_KVStore_MassBatchDelete(t *testing.T) {
	services.InitKvStore()
	defer services.CloseKvStore()

	err := services.BatchSet(getKvPairs(guessedMaxBatchSize), models.TestValueTimeToLive)
	oyster_utils.AssertNoError(err, t, "")

	err = services.BatchDelete(getKeys(guessedMaxBatchSize))
	oyster_utils.AssertNoError(err, t, "")
}

func Test_KVStore_RemoveAllKvStoreData(t *testing.T) {
	services.InitKvStore()
	defer services.CloseKvStore()

	services.BatchSet(getKvPairs(2), models.TestValueTimeToLive)
	err := services.RemoveAllKvStoreData()
	oyster_utils.AssertNoError(err, t, "")

	services.InitKvStore()
	kvs, _ := services.BatchGet(getKeys(2))

	oyster_utils.AssertTrue(len(*kvs) == 0, t, "")
}

func Test_KVStore_GetMessageFromDataMap_FromKVStore(t *testing.T) {
	services.InitKvStore()
	defer services.CloseKvStore()

	services.BatchSet(getKvPairs(1), models.TestValueTimeToLive)

	dataMap := models.DataMap{
		MsgID: "0",
	}

	oyster_utils.AssertStringEqual(services.GetMessageFromDataMap(dataMap), "0", t)
}

func Test_KVStore_GetMessageFromDataMap_FromMessage(t *testing.T) {
	services.InitKvStore()
	defer services.CloseKvStore()

	dataMap := models.DataMap{
		MsgID:   "1",
		Message: "hello",
	}

	oyster_utils.AssertStringEqual(services.GetMessageFromDataMap(dataMap), "hello", t)
}

func getKvPairs(count int) *services.KVPairs {
	pairs := services.KVPairs{}
	for i := 0; i < count; i++ {
		pairs[strconv.Itoa(i)] = strconv.Itoa(i)
	}
	return &pairs
}

func getKeys(count int) *services.KVKeys {
	keys := services.KVKeys{}
	for i := 0; i < guessedMaxBatchSize; i++ {
		keys = append(keys, strconv.Itoa(i))
	}
	return &keys
}

func Test_KVStore_InitUniqueKvStore(t *testing.T) {
	err := services.InitUniqueKvStore(testDBID)
	dbName := services.GetBadgerDBName(testDBID)

	oyster_utils.AssertNoError(err, t, "Could not create Badger DB")
	defer services.CloseUniqueKvStore(dbName)
}

func Test_KVStore_MassBatchSetToUniqueDB(t *testing.T) {
	services.InitUniqueKvStore(testDBID)
	dbName := services.GetBadgerDBName(testDBID)
	defer services.CloseUniqueKvStore(dbName)

	err := services.BatchSetToUniqueDB(testDBID,
		getKvPairs(guessedMaxBatchSize), models.TestValueTimeToLive)
	oyster_utils.AssertNoError(err, t, "")

	kvs, _ := services.BatchGetFromUniqueDB(testDBID,
		&services.KVKeys{strconv.Itoa(guessedMaxBatchSize - 1)})
	oyster_utils.AssertTrue(len(*kvs) == 1, t, "Expect only an item")
}

func Test_KVStore_BatchGetFromUniqueDB(t *testing.T) {
	services.InitUniqueKvStore(testDBID)
	dbName := services.GetBadgerDBName(testDBID)
	defer services.CloseUniqueKvStore(dbName)

	services.BatchSetToUniqueDB(testDBID, &services.KVPairs{"key": "oyster"},
		models.TestValueTimeToLive)

	kvs, err := services.BatchGetFromUniqueDB(testDBID, &services.KVKeys{"key"})
	oyster_utils.AssertNoError(err, t, "Could not get key")

	oyster_utils.AssertTrue(len(*kvs) == 1, t, "")
	oyster_utils.AssertStringEqual((*kvs)["key"], "oyster", t)
}

func Test_KVStore_BatchGetFromUniqueDB_WithMissingKey(t *testing.T) {
	services.InitUniqueKvStore(testDBID)
	dbName := services.GetBadgerDBName(testDBID)
	defer services.CloseUniqueKvStore(dbName)

	services.BatchSetToUniqueDB(testDBID, &services.KVPairs{"key": "oyster"},
		models.TestValueTimeToLive)

	kvs, err := services.BatchGetFromUniqueDB(testDBID,
		&services.KVKeys{"key", "unknownKey"})
	oyster_utils.AssertNoError(err, t, "Could not get key")

	oyster_utils.AssertTrue(len(*kvs) == 1, t, "")
	oyster_utils.AssertStringEqual((*kvs)["key"], "oyster", t)
}

func Test_KVStore_MassBatchGetFromUniqueDB(t *testing.T) {
	services.InitUniqueKvStore(testDBID)
	dbName := services.GetBadgerDBName(testDBID)
	defer services.CloseUniqueKvStore(dbName)

	err := services.BatchSetToUniqueDB(testDBID,
		getKvPairs(guessedMaxBatchSize), models.TestValueTimeToLive)
	oyster_utils.AssertNoError(err, t, "")

	kvs, _ := services.BatchGetFromUniqueDB(testDBID, getKeys(guessedMaxBatchSize))
	oyster_utils.AssertTrue(len(*kvs) == guessedMaxBatchSize, t, "")
}

func Test_KVStore_BatchDeleteFromUniqueDB(t *testing.T) {
	services.InitUniqueKvStore(testDBID)
	dbName := services.GetBadgerDBName(testDBID)
	defer services.CloseUniqueKvStore(dbName)

	services.BatchSetToUniqueDB(testDBID,
		&services.KVPairs{"key1": "oyster1", "key2": "oyster2"}, models.TestValueTimeToLive)

	err := services.BatchDeleteFromUniqueDB(testDBID,
		&services.KVKeys{"key1"})
	oyster_utils.AssertNoError(err, t, "Could not delete key")

	kvs, err := services.BatchGetFromUniqueDB(testDBID,
		&services.KVKeys{"key1"})
	oyster_utils.AssertNoError(err, t, "Could complete get key")
	oyster_utils.AssertTrue(len(*kvs) == 0, t, "")
}

func Test_KVStore_Mass_BatchDeleteFromUniqueDB(t *testing.T) {
	services.InitUniqueKvStore(testDBID)
	dbName := services.GetBadgerDBName(testDBID)
	defer services.CloseUniqueKvStore(dbName)

	err := services.BatchSetToUniqueDB(testDBID,
		getKvPairs(guessedMaxBatchSize), models.TestValueTimeToLive)
	oyster_utils.AssertNoError(err, t, "")

	err = services.BatchDeleteFromUniqueDB(testDBID,
		getKeys(guessedMaxBatchSize))
	oyster_utils.AssertNoError(err, t, "")
}

func Test_KVStore_RemoveAllUniqueKvStoreData(t *testing.T) {
	services.InitUniqueKvStore(testDBID)
	dbName := services.GetBadgerDBName(testDBID)
	defer services.CloseUniqueKvStore(dbName)

	services.BatchSetToUniqueDB(testDBID, getKvPairs(2), models.TestValueTimeToLive)
	err := services.RemoveAllUniqueKvStoreData(dbName)
	oyster_utils.AssertNoError(err, t, "")

	services.InitUniqueKvStore(testDBID)
	kvs, _ := services.BatchGetFromUniqueDB(testDBID, getKeys(2))

	oyster_utils.AssertTrue(len(*kvs) == 0, t, "")
}

func Test_KVStore_RemoveAllKvStoreDataFromAllKvStores(t *testing.T) {

	dbID1 := []string{"prefix", "genhash1", services.MessageDir}
	dbID2 := []string{"prefix", "genhash2", services.MessageDir}

	services.InitUniqueKvStore(dbID1)
	dbName1 := services.GetBadgerDBName(dbID1)
	defer services.CloseUniqueKvStore(dbName1)

	services.InitUniqueKvStore(dbID2)
	dbName2 := services.GetBadgerDBName(dbID2)
	defer services.CloseUniqueKvStore(dbName2)

	services.BatchSetToUniqueDB(dbID1, getKvPairs(2), models.TestValueTimeToLive)
	services.BatchSetToUniqueDB(dbID2, getKvPairs(2), models.TestValueTimeToLive)

	services.RemoveAllKvStoreDataFromAllKvStores()

	services.InitUniqueKvStore(dbID1)
	services.InitUniqueKvStore(dbID2)

	kvs1, _ := services.BatchGetFromUniqueDB(dbID1, getKeys(2))
	kvs2, _ := services.BatchGetFromUniqueDB(dbID2, getKeys(2))

	oyster_utils.AssertTrue(len(*kvs1) == 0, t, "")
	oyster_utils.AssertTrue(len(*kvs2) == 0, t, "")
}

func Test_KVStore_GetChunkData(t *testing.T) {

	prefix := "somePrefix"
	genesisHash := "someGenHash"

	messageDBID := []string{prefix, genesisHash, services.MessageDir}
	messageDBName := services.GetBadgerDBName(messageDBID)

	hashDBID := []string{prefix, genesisHash, services.HashDir}
	hashDBName := services.GetBadgerDBName(hashDBID)

	hash := "abcdeff"
	message := "testMessage"
	testAddress := "PHYDCDNGNEKFPAHEO9VHR9UAKFLDLGICICIIJCNDLGMGIASAUCCHBCIDIDVCWC9EIGEFZAQDH9AIAIUFN"

	services.InitUniqueKvStore(messageDBID)
	services.InitUniqueKvStore(hashDBID)

	defer services.CloseUniqueKvStore(messageDBName)
	defer services.CloseUniqueKvStore(hashDBName)

	services.BatchSetToUniqueDB(messageDBID,
		&services.KVPairs{
			genesisHash + "_1": message + "1",
			genesisHash + "_2": message + "2",
			genesisHash + "_3": message + "3",
			genesisHash + "_4": message + "4",
			genesisHash + "_5": message + "5",
			genesisHash + "_6": message + "6"},
		models.TestValueTimeToLive)

	services.BatchSetToUniqueDB(hashDBID,
		&services.KVPairs{
			genesisHash + "_1": hash + "1",
			genesisHash + "_2": hash + "2",
			genesisHash + "_3": hash + "3",
			genesisHash + "_4": hash + "4",
			genesisHash + "_5": hash + "5",
			genesisHash + "_6": hash + "6"},
		models.TestValueTimeToLive)

	chunkData := services.GetChunkData(prefix, genesisHash, 3)

	trytes, _ := oyster_utils.ChunkMessageToTrytesWithStopper(message + "3")

	oyster_utils.AssertContainString(chunkData.Hash, hash, t)
	oyster_utils.AssertContainString(chunkData.Message, string(trytes), t)
	oyster_utils.AssertStringEqual(chunkData.Address, testAddress, t)
}

func Test_KVStore_DeleteDataFromUniqueDB_Ascending(t *testing.T) {

	prefix := "somePrefix"
	genesisHash := "someGenHash"

	messageDBID := []string{prefix, genesisHash, services.MessageDir}
	messageDBName := services.GetBadgerDBName(messageDBID)

	message := "testMessage"

	services.InitUniqueKvStore(messageDBID)
	defer services.CloseUniqueKvStore(messageDBName)

	keys := &services.KVKeys{
		genesisHash + "_1",
		genesisHash + "_2",
		genesisHash + "_3",
		genesisHash + "_4",
		genesisHash + "_5",
		genesisHash + "_6",
	}

	services.BatchSetToUniqueDB(messageDBID,
		&services.KVPairs{
			genesisHash + "_1": message + "1",
			genesisHash + "_2": message + "2",
			genesisHash + "_3": message + "3",
			genesisHash + "_4": message + "4",
			genesisHash + "_5": message + "5",
			genesisHash + "_6": message + "6"},
		models.TestValueTimeToLive)

	err := services.DeleteDataFromUniqueDB(messageDBID, genesisHash, 2, 4)

	oyster_utils.AssertNoError(err, t, "Could not delete from db for certain index range")

	kvs, _ := services.BatchGetFromUniqueDB(messageDBID, keys)
	oyster_utils.AssertTrue(len(*kvs) == 3, t, "")

	foundWhatWeWanted := false
	designatedKeysAreGone := false

	for _, value := range *kvs {
		foundWhatWeWanted = strings.Contains(value, "1") || strings.Contains(value, "5") ||
			strings.Contains(value, "6")

		designatedKeysAreGone = !strings.Contains(value, "2") && !strings.Contains(value, "3") &&
			!strings.Contains(value, "4")

		oyster_utils.AssertTrue(foundWhatWeWanted, t, "")
		oyster_utils.AssertTrue(designatedKeysAreGone, t, "")
	}
}

func Test_KVStore_DeleteDataFromUniqueDB_Descending(t *testing.T) {

	prefix := "somePrefix"
	genesisHash := "someGenHash"

	messageDBID := []string{prefix, genesisHash, services.MessageDir}
	messageDBName := services.GetBadgerDBName(messageDBID)

	message := "testMessage"

	services.InitUniqueKvStore(messageDBID)
	defer services.CloseUniqueKvStore(messageDBName)

	keys := &services.KVKeys{
		genesisHash + "_1",
		genesisHash + "_2",
		genesisHash + "_3",
		genesisHash + "_4",
		genesisHash + "_5",
		genesisHash + "_6",
	}

	services.BatchSetToUniqueDB(messageDBID,
		&services.KVPairs{
			genesisHash + "_1": message + "1",
			genesisHash + "_2": message + "2",
			genesisHash + "_3": message + "3",
			genesisHash + "_4": message + "4",
			genesisHash + "_5": message + "5",
			genesisHash + "_6": message + "6"},
		models.TestValueTimeToLive)

	err := services.DeleteDataFromUniqueDB(messageDBID, genesisHash, 4, 2)

	oyster_utils.AssertNoError(err, t, "Could not delete from db for certain index range")

	kvs, _ := services.BatchGetFromUniqueDB(messageDBID, keys)
	oyster_utils.AssertTrue(len(*kvs) == 3, t, "")

	foundWhatWeWanted := false
	designatedKeysAreGone := false

	for _, value := range *kvs {
		foundWhatWeWanted = strings.Contains(value, "1") || strings.Contains(value, "5") ||
			strings.Contains(value, "6")

		designatedKeysAreGone = !strings.Contains(value, "2") && !strings.Contains(value, "3") &&
			!strings.Contains(value, "4")

		oyster_utils.AssertTrue(foundWhatWeWanted, t, "")
		oyster_utils.AssertTrue(designatedKeysAreGone, t, "")
	}
}

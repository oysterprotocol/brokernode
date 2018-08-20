package oyster_utils_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/utils"
)

// a very big number that will lead to ErrTxnTooBig for write.
const guessedMaxBatchSize = 200000

var testDBID = []string{"prefix", "genHash", "data"}

func Test_KVStore_Init(t *testing.T) {
	err := oyster_utils.InitKvStore()

	oyster_utils.AssertNoError(err, t, "Could not create Badger DB")
	defer oyster_utils.CloseKvStore()
}

func Test_KVStore_MassBatchSet(t *testing.T) {
	oyster_utils.InitKvStore()
	defer oyster_utils.CloseKvStore()

	err := oyster_utils.BatchSet(getKvPairs(guessedMaxBatchSize), models.TestValueTimeToLive)
	oyster_utils.AssertNoError(err, t, "")

	kvs, _ := oyster_utils.BatchGet(&oyster_utils.KVKeys{strconv.Itoa(guessedMaxBatchSize - 1)})
	oyster_utils.AssertTrue(len(*kvs) == 1, t, "Expect only an item")
}

func Test_KVStoreBatchGet(t *testing.T) {
	oyster_utils.InitKvStore()
	defer oyster_utils.CloseKvStore()

	oyster_utils.BatchSet(&oyster_utils.KVPairs{"key": "oyster"}, models.TestValueTimeToLive)

	kvs, err := oyster_utils.BatchGet(&oyster_utils.KVKeys{"key"})
	oyster_utils.AssertNoError(err, t, "Could not get key")

	oyster_utils.AssertTrue(len(*kvs) == 1, t, "")
	oyster_utils.AssertStringEqual((*kvs)["key"], "oyster", t)
}

func Test_KVStoreBatchGet_WithMissingKey(t *testing.T) {
	oyster_utils.InitKvStore()
	defer oyster_utils.CloseKvStore()

	oyster_utils.BatchSet(&oyster_utils.KVPairs{"key": "oyster"}, models.TestValueTimeToLive)

	kvs, err := oyster_utils.BatchGet(&oyster_utils.KVKeys{"key", "unknownKey"})
	oyster_utils.AssertNoError(err, t, "Could not get key")

	oyster_utils.AssertTrue(len(*kvs) == 1, t, "")
	oyster_utils.AssertStringEqual((*kvs)["key"], "oyster", t)
}

func Test_KVStore_MassBatchGet(t *testing.T) {
	oyster_utils.InitKvStore()
	defer oyster_utils.CloseKvStore()

	err := oyster_utils.BatchSet(getKvPairs(guessedMaxBatchSize), models.TestValueTimeToLive)
	oyster_utils.AssertNoError(err, t, "")

	kvs, _ := oyster_utils.BatchGet(getKeys(guessedMaxBatchSize))
	oyster_utils.AssertTrue(len(*kvs) == guessedMaxBatchSize, t, "")
}

func Test_KVStoreBatchDelete(t *testing.T) {
	oyster_utils.InitKvStore()
	defer oyster_utils.CloseKvStore()

	oyster_utils.BatchSet(&oyster_utils.KVPairs{"key1": "oyster1", "key2": "oyster2"}, models.TestValueTimeToLive)

	err := oyster_utils.BatchDelete(&oyster_utils.KVKeys{"key1"})
	oyster_utils.AssertNoError(err, t, "Could not delete key")

	kvs, err := oyster_utils.BatchGet(&oyster_utils.KVKeys{"key1"})
	oyster_utils.AssertNoError(err, t, "Could complete get key")
	oyster_utils.AssertTrue(len(*kvs) == 0, t, "")
}

func Test_KVStore_MassBatchDelete(t *testing.T) {
	oyster_utils.InitKvStore()
	defer oyster_utils.CloseKvStore()

	err := oyster_utils.BatchSet(getKvPairs(guessedMaxBatchSize), models.TestValueTimeToLive)
	oyster_utils.AssertNoError(err, t, "")

	err = oyster_utils.BatchDelete(getKeys(guessedMaxBatchSize))
	oyster_utils.AssertNoError(err, t, "")
}

func Test_KVStore_RemoveAllKvStoreData(t *testing.T) {
	oyster_utils.InitKvStore()
	defer oyster_utils.CloseKvStore()

	oyster_utils.BatchSet(getKvPairs(2), models.TestValueTimeToLive)
	err := oyster_utils.RemoveAllKvStoreData()
	oyster_utils.AssertNoError(err, t, "")

	oyster_utils.InitKvStore()
	kvs, _ := oyster_utils.BatchGet(getKeys(2))

	oyster_utils.AssertTrue(len(*kvs) == 0, t, "")
}

func getKvPairs(count int) *oyster_utils.KVPairs {
	pairs := oyster_utils.KVPairs{}
	for i := 0; i < count; i++ {
		pairs[strconv.Itoa(i)] = strconv.Itoa(i)
	}
	return &pairs
}

func getKeys(count int) *oyster_utils.KVKeys {
	keys := oyster_utils.KVKeys{}
	for i := 0; i < guessedMaxBatchSize; i++ {
		keys = append(keys, strconv.Itoa(i))
	}
	return &keys
}

func Test_KVStore_InitUniqueKvStore(t *testing.T) {
	err := oyster_utils.InitUniqueKvStore(testDBID)
	dbName := oyster_utils.GetBadgerDBName(testDBID)

	oyster_utils.AssertNoError(err, t, "Could not create Badger DB")
	defer oyster_utils.CloseUniqueKvStore(dbName)
}

func Test_KVStore_MassBatchSetToUniqueDB(t *testing.T) {
	oyster_utils.InitUniqueKvStore(testDBID)
	dbName := oyster_utils.GetBadgerDBName(testDBID)
	defer oyster_utils.CloseUniqueKvStore(dbName)

	err := oyster_utils.BatchSetToUniqueDB(testDBID,
		getKvPairs(guessedMaxBatchSize), models.TestValueTimeToLive)
	oyster_utils.AssertNoError(err, t, "")

	kvs, _ := oyster_utils.BatchGetFromUniqueDB(testDBID,
		&oyster_utils.KVKeys{strconv.Itoa(guessedMaxBatchSize - 1)})
	oyster_utils.AssertTrue(len(*kvs) == 1, t, "Expect only an item")
}

func Test_KVStore_BatchGetFromUniqueDB(t *testing.T) {
	oyster_utils.InitUniqueKvStore(testDBID)
	dbName := oyster_utils.GetBadgerDBName(testDBID)
	defer oyster_utils.CloseUniqueKvStore(dbName)

	oyster_utils.BatchSetToUniqueDB(testDBID, &oyster_utils.KVPairs{"key": "oyster"},
		models.TestValueTimeToLive)

	kvs, err := oyster_utils.BatchGetFromUniqueDB(testDBID, &oyster_utils.KVKeys{"key"})
	oyster_utils.AssertNoError(err, t, "Could not get key")

	oyster_utils.AssertTrue(len(*kvs) == 1, t, "")
	oyster_utils.AssertStringEqual((*kvs)["key"], "oyster", t)
}

func Test_KVStore_BatchGetFromUniqueDB_WithMissingKey(t *testing.T) {
	oyster_utils.InitUniqueKvStore(testDBID)
	dbName := oyster_utils.GetBadgerDBName(testDBID)
	defer oyster_utils.CloseUniqueKvStore(dbName)

	oyster_utils.BatchSetToUniqueDB(testDBID, &oyster_utils.KVPairs{"key": "oyster"},
		models.TestValueTimeToLive)

	kvs, err := oyster_utils.BatchGetFromUniqueDB(testDBID,
		&oyster_utils.KVKeys{"key", "unknownKey"})
	oyster_utils.AssertNoError(err, t, "Could not get key")

	oyster_utils.AssertTrue(len(*kvs) == 1, t, "")
	oyster_utils.AssertStringEqual((*kvs)["key"], "oyster", t)
}

func Test_KVStore_MassBatchGetFromUniqueDB(t *testing.T) {
	oyster_utils.InitUniqueKvStore(testDBID)
	dbName := oyster_utils.GetBadgerDBName(testDBID)
	defer oyster_utils.CloseUniqueKvStore(dbName)

	err := oyster_utils.BatchSetToUniqueDB(testDBID,
		getKvPairs(guessedMaxBatchSize), models.TestValueTimeToLive)
	oyster_utils.AssertNoError(err, t, "")

	kvs, _ := oyster_utils.BatchGetFromUniqueDB(testDBID, getKeys(guessedMaxBatchSize))
	oyster_utils.AssertTrue(len(*kvs) == guessedMaxBatchSize, t, "")
}

func Test_KVStore_BatchDeleteFromUniqueDB(t *testing.T) {
	oyster_utils.InitUniqueKvStore(testDBID)
	dbName := oyster_utils.GetBadgerDBName(testDBID)
	defer oyster_utils.CloseUniqueKvStore(dbName)

	oyster_utils.BatchSetToUniqueDB(testDBID,
		&oyster_utils.KVPairs{"key1": "oyster1", "key2": "oyster2"}, models.TestValueTimeToLive)

	err := oyster_utils.BatchDeleteFromUniqueDB(testDBID,
		&oyster_utils.KVKeys{"key1"})
	oyster_utils.AssertNoError(err, t, "Could not delete key")

	kvs, err := oyster_utils.BatchGetFromUniqueDB(testDBID,
		&oyster_utils.KVKeys{"key1"})
	oyster_utils.AssertNoError(err, t, "Could complete get key")
	oyster_utils.AssertTrue(len(*kvs) == 0, t, "")
}

func Test_KVStore_Mass_BatchDeleteFromUniqueDB(t *testing.T) {
	oyster_utils.InitUniqueKvStore(testDBID)
	dbName := oyster_utils.GetBadgerDBName(testDBID)
	defer oyster_utils.CloseUniqueKvStore(dbName)

	err := oyster_utils.BatchSetToUniqueDB(testDBID,
		getKvPairs(guessedMaxBatchSize), models.TestValueTimeToLive)
	oyster_utils.AssertNoError(err, t, "")

	err = oyster_utils.BatchDeleteFromUniqueDB(testDBID,
		getKeys(guessedMaxBatchSize))
	oyster_utils.AssertNoError(err, t, "")
}

func Test_KVStore_RemoveAllUniqueKvStoreData(t *testing.T) {
	oyster_utils.InitUniqueKvStore(testDBID)
	dbName := oyster_utils.GetBadgerDBName(testDBID)
	defer oyster_utils.CloseUniqueKvStore(dbName)

	oyster_utils.BatchSetToUniqueDB(testDBID, getKvPairs(2), models.TestValueTimeToLive)
	err := oyster_utils.RemoveAllUniqueKvStoreData(dbName)

	oyster_utils.AssertNoError(err, t, "")

	oyster_utils.InitUniqueKvStore(testDBID)
	kvs, _ := oyster_utils.BatchGetFromUniqueDB(testDBID, getKeys(2))

	oyster_utils.AssertTrue(len(*kvs) == 0, t, "")
}

func Test_KVStore_RemoveAllKvStoreDataFromAllKvStores(t *testing.T) {

	dbID1 := []string{"prefix", "genhash1", oyster_utils.MessageDir}
	dbID2 := []string{"prefix", "genhash2", oyster_utils.MessageDir}

	oyster_utils.InitUniqueKvStore(dbID1)
	dbName1 := oyster_utils.GetBadgerDBName(dbID1)
	defer oyster_utils.CloseUniqueKvStore(dbName1)

	oyster_utils.InitUniqueKvStore(dbID2)
	dbName2 := oyster_utils.GetBadgerDBName(dbID2)
	defer oyster_utils.CloseUniqueKvStore(dbName2)

	oyster_utils.BatchSetToUniqueDB(dbID1, getKvPairs(2), models.TestValueTimeToLive)
	oyster_utils.BatchSetToUniqueDB(dbID2, getKvPairs(2), models.TestValueTimeToLive)

	oyster_utils.RemoveAllKvStoreDataFromAllKvStores()

	oyster_utils.InitUniqueKvStore(dbID1)
	oyster_utils.InitUniqueKvStore(dbID2)

	kvs1, _ := oyster_utils.BatchGetFromUniqueDB(dbID1, getKeys(2))
	kvs2, _ := oyster_utils.BatchGetFromUniqueDB(dbID2, getKeys(2))

	oyster_utils.AssertTrue(len(*kvs1) == 0, t, "")
	oyster_utils.AssertTrue(len(*kvs2) == 0, t, "")
}

func Test_KVStore_GetChunkData(t *testing.T) {

	prefix := "somePrefix"
	genesisHash := "someGenHash"

	messageDBID := []string{prefix, genesisHash, oyster_utils.MessageDir}
	messageDBName := oyster_utils.GetBadgerDBName(messageDBID)

	hashDBID := []string{prefix, genesisHash, oyster_utils.HashDir}
	hashDBName := oyster_utils.GetBadgerDBName(hashDBID)

	hash := "abcdeff"
	message := "testMessage"
	testAddress := "PHYDCDNGNEKFPAHEO9VHR9UAKFLDLGICICIIJCNDLGMGIASAUCCHBCIDIDVCWC9EIGEFZAQDH9AIAIUFN"

	oyster_utils.InitUniqueKvStore(messageDBID)
	oyster_utils.InitUniqueKvStore(hashDBID)

	defer oyster_utils.CloseUniqueKvStore(messageDBName)
	defer oyster_utils.CloseUniqueKvStore(hashDBName)

	oyster_utils.BatchSetToUniqueDB(messageDBID,
		&oyster_utils.KVPairs{
			genesisHash + "_1": message + "1",
			genesisHash + "_2": message + "2",
			genesisHash + "_3": message + "3",
			genesisHash + "_4": message + "4",
			genesisHash + "_5": message + "5",
			genesisHash + "_6": message + "6"},
		models.TestValueTimeToLive)

	oyster_utils.BatchSetToUniqueDB(hashDBID,
		&oyster_utils.KVPairs{
			genesisHash + "_1": hash + "1",
			genesisHash + "_2": hash + "2",
			genesisHash + "_3": hash + "3",
			genesisHash + "_4": hash + "4",
			genesisHash + "_5": hash + "5",
			genesisHash + "_6": hash + "6"},
		models.TestValueTimeToLive)

	chunkData := oyster_utils.GetChunkData(prefix, genesisHash, 3)

	trytes, _ := oyster_utils.ChunkMessageToTrytesWithStopper(message + "3")

	oyster_utils.AssertContainString(chunkData.Hash, hash, t)
	oyster_utils.AssertContainString(chunkData.Message, string(trytes), t)
	oyster_utils.AssertStringEqual(chunkData.Address, testAddress, t)
}

func Test_KVStore_DeleteDataFromUniqueDB_Ascending(t *testing.T) {

	prefix := "somePrefix"
	genesisHash := "someGenHash"

	messageDBID := []string{prefix, genesisHash, oyster_utils.MessageDir}
	messageDBName := oyster_utils.GetBadgerDBName(messageDBID)

	message := "testMessage"

	oyster_utils.InitUniqueKvStore(messageDBID)
	defer oyster_utils.CloseUniqueKvStore(messageDBName)

	keys := &oyster_utils.KVKeys{
		genesisHash + "_1",
		genesisHash + "_2",
		genesisHash + "_3",
		genesisHash + "_4",
		genesisHash + "_5",
		genesisHash + "_6",
	}

	oyster_utils.BatchSetToUniqueDB(messageDBID,
		&oyster_utils.KVPairs{
			genesisHash + "_1": message + "1",
			genesisHash + "_2": message + "2",
			genesisHash + "_3": message + "3",
			genesisHash + "_4": message + "4",
			genesisHash + "_5": message + "5",
			genesisHash + "_6": message + "6"},
		models.TestValueTimeToLive)

	err := oyster_utils.DeleteDataFromUniqueDB(messageDBID, genesisHash, 2, 4)

	oyster_utils.AssertNoError(err, t, "Could not delete from db for certain index range")

	kvs, _ := oyster_utils.BatchGetFromUniqueDB(messageDBID, keys)
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

	messageDBID := []string{prefix, genesisHash, oyster_utils.MessageDir}
	messageDBName := oyster_utils.GetBadgerDBName(messageDBID)

	message := "testMessage"

	oyster_utils.InitUniqueKvStore(messageDBID)
	defer oyster_utils.CloseUniqueKvStore(messageDBName)

	keys := &oyster_utils.KVKeys{
		genesisHash + "_1",
		genesisHash + "_2",
		genesisHash + "_3",
		genesisHash + "_4",
		genesisHash + "_5",
		genesisHash + "_6",
	}

	oyster_utils.BatchSetToUniqueDB(messageDBID,
		&oyster_utils.KVPairs{
			genesisHash + "_1": message + "1",
			genesisHash + "_2": message + "2",
			genesisHash + "_3": message + "3",
			genesisHash + "_4": message + "4",
			genesisHash + "_5": message + "5",
			genesisHash + "_6": message + "6"},
		models.TestValueTimeToLive)

	err := oyster_utils.DeleteDataFromUniqueDB(messageDBID, genesisHash, 4, 2)

	oyster_utils.AssertNoError(err, t, "Could not delete from db for certain index range")

	kvs, _ := oyster_utils.BatchGetFromUniqueDB(messageDBID, keys)
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

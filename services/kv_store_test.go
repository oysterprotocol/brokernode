package services_test

import (
	"strconv"
	"testing"

	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
)

// a very big number that will lead to ErrTxnTooBig for write.
const guessedMaxBatchSize = 200000

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

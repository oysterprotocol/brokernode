package services_test

import (
	"strconv"
	"testing"

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

	err := services.BatchSet(getKvPairs(guessedMaxBatchSize))
	oyster_utils.AssertNoError(err, t, "")

	kvs, _ := services.BatchGet(&services.KVKeys{strconv.Itoa(guessedMaxBatchSize - 1)})
	oyster_utils.AssertTrue(len(*kvs) == 1, t, "Expect only an item")
}

func Test_KVStoreBatchGet(t *testing.T) {
	services.InitKvStore()
	defer services.CloseKvStore()

	services.BatchSet(&services.KVPairs{"key": "oyster"})

	kvs, err := services.BatchGet(&services.KVKeys{"key"})
	oyster_utils.AssertNoError(err, t, "Could not get key")

	oyster_utils.AssertTrue(len(*kvs) == 1, t, "")
	oyster_utils.AssertStringEqual((*kvs)["key"], "oyster", t)
}

func Test_KVStoreBatchGet_WithMissingKey(t *testing.T) {
	services.InitKvStore()
	defer services.CloseKvStore()

	services.BatchSet(&services.KVPairs{"key": "oyster"})

	kvs, err := services.BatchGet(&services.KVKeys{"key", "unknownKey"})
	oyster_utils.AssertNoError(err, t, "Could not get key")

	oyster_utils.AssertTrue(len(*kvs) == 1, t, "")
	oyster_utils.AssertStringEqual((*kvs)["key"], "oyster", t)
}

func Test_KVStore_MassBatchGet(t *testing.T) {
	services.InitKvStore()
	defer services.CloseKvStore()

	err := services.BatchSet(getKvPairs(guessedMaxBatchSize))
	oyster_utils.AssertNoError(err, t, "")

	kvs, _ = services.BatchGet(getKeys(guessedMaxBatchSize))
	oyster_utils.AssertTrue(len(*kvs) == guessedMaxBatchSize, t, "")
}

func Test_KVStoreBatchDelete(t *testing.T) {
	services.InitKvStore()
	defer services.CloseKvStore()

	services.BatchSet(&services.KVPairs{"key1": "oyster1", "key2": "oyster2"})

	err := services.BatchDelete(&services.KVKeys{"key1"})
	oyster_utils.AssertNoError(err, t, "Could not delete key")

	kvs, err := services.BatchGet(&services.KVKeys{"key1"})
	oyster_utils.AssertNoError(err, t, "Could complete get key")
	oyster_utils.AssertTrue(len(*kvs) == 0, t, "")
}

func Test_KVStore_MassBatchDelete(t *testing.T) {
	services.InitKvStore()
	defer services.CloseKvStore()

	err := services.BatchSet(getKvPairs(guessedMaxBatchSize))
	oyster_utils.AssertNoError(err, t, "")

	err = services.BatchDelete(getKeys(guessedMaxBatchSize))
	oyster_utils.AssertNoError(err, t, "")
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

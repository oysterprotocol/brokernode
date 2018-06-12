package services_test

import (
	"strconv"
	"testing"

	"fmt"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
)

func Test_KVStore_Init(t *testing.T) {
	err := services.InitKvStore()

	oyster_utils.AssertNoError(err, t, "Could not create Badger DB")
	defer services.CloseKvStore()
}

func Test_KVStore_MassBatchSet(t *testing.T) {
	guessedMaxBatchSize := 200000
	services.InitKvStore()
	defer services.CloseKvStore()

	pairs := services.KVPairs{}
	// See whether it would break in Update method. And some part of it will be inserted
	for i := 0; i < guessedMaxBatchSize; i++ {
		pairs[strconv.Itoa(i)] = strconv.Itoa(i)
	}

	err := services.BatchSet(&pairs)
	oyster_utils.AssertError(err, t, "")

	kvs, _ := services.BatchGet(&services.KVKeys{strconv.Itoa(guessedMaxBatchSize - 1)})
	oyster_utils.AssertTrue(len(*kvs) == 0, t, "Expect only 0 item")
}

func Test_KVStore_MassBatchDelete(t *testing.T) {
	guessedMaxBatchSize := 200000
	services.InitKvStore()
	defer services.CloseKvStore()

	pairs1 := services.KVPairs{}
	for i := 0; i < guessedMaxBatchSize/2; i++ {
		pairs1[strconv.Itoa(i)] = strconv.Itoa(i)
	}
	oyster_utils.AssertNoError(services.BatchSet(&pairs1), t, "")

	pairs2 := services.KVPairs{}
	for i := guessedMaxBatchSize / 2; i < guessedMaxBatchSize; i++ {
		pairs2[strconv.Itoa(i)] = strconv.Itoa(i)
	}
	oyster_utils.AssertNoError(services.BatchSet(&pairs2), t, "")

	keys := services.KVKeys{}
	for i := 0; i < guessedMaxBatchSize; i++ {
		keys = append(keys, strconv.Itoa(i))
	}
	oyster_utils.AssertError(services.BatchDelete(&keys), t, "")
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

package services_test

import (
	"testing"

	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
)

func Test_KVStore_Init(t *testing.T) {
	err := services.InitKvStore()

	oyster_utils.AssertNoError(err, t, "Could not create Badger DB")
	defer services.CloseKvStore()
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

func Test_KVStoreGenKey(t *testing.T) {
	v := services.GenKvStoreKey("abc", 1)

	oyster_utils.AssertStringEqual(v, "abc_1", t)
}

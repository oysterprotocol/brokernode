package services_test

import (
	"testing"

	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
)

func Test_KVStore(t *testing.T) {
	db, err := services.InitKVStore()
	oyster_utils.AssertNoError(err, t, "Could not create Badger DB")
	defer db.Close()

	err = services.BatchSet(&services.KVPairs{"key": "oyster"})
	oyster_utils.AssertNoError(err, t, "Could not set key")

	kvs, err := services.BatchGet(&services.KVKeys{"key"})
	oyster_utils.AssertNoError(err, t, "Could not get key")

	val := (*kvs)["key"]
	oyster_utils.AssertStringEqual(val, "oyster", t)
}

func Test_KVStoreBatchGet_WithMissKeyValue(t *testing.T) {
	db, _ := services.InitKVStore()
	defer db.Close()

	err = services.BatchSet(&services.KVPairs{"key": "oyster"})

	kvs, err := services.BatchGet(&services.KVKeys{"key", "unknownKey"})
	oyster_utils.AssertNoError(err, t, "Could not get key")

	oyster_utils.AssertTrue(len(kvs) == 1, t, "")
	oyster_utils.AssertStringEqual((*kvs["key"]), "oyster", t)
}

func Test_KVStoreBatchDelete(t *testing.T) {
	db, _ := services.InitKVStore()
	defer db.Close()

	services.BatchSet(&KVPairs{"key1": "oyster1", "key2", "oyster2"})

	err := services.BatchDelete(&services.KVKeys{"key1"})
	oyster_utils.AssertNoError(err, t, "Could not delete key")
}

func Test_KVStoreGenKey(t *testing.T) {
	v := services.GenKvStoreKey("abc", 1)

	oyster_utils.AssertStringEqual(v, "abc_1", t)
}

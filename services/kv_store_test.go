package services

import (
	"testing"

	"github.com/oysterprotocol/brokernode/utils"
)

func Test_KVStore(t *testing.T) {
	db, err := InitKVStore()
	oyster_utils.AssertNoError(err, t, "Could not create Badger DB")
	defer db.Close()

	// Get key
	kvs, err := BatchGet(&KVKeys{"key"})
	oyster_utils.AssertError(err, t, "Could not get key")

	// Set key
	err = BatchSet(&KVPairs{"key": "oyster"})
	oyster_utils.AssertNoError(err, t, "Could not set key")

	// Get key
	kvs, err = BatchGet(&KVKeys{"key"})
	oyster_utils.AssertNoError(err, t, "Could not get key")
	val := (*kvs)["key"]
	oyster_utils.AssertStringEqual(val, "oyster", t)

	// Delete Key
	err = BatchDelete(&KVKeys{"key"})
	oyster_utils.AssertNoError(err, t, "Could not delete key")

	// Get key
	kvs, err = BatchGet(&KVKeys{"key"})
	oyster_utils.AssertError(err, t, "Could not get key")
}

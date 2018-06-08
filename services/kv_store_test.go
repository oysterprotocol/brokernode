package services

import (
	"testing"
)

func Test_KVStore(t *testing.T) {
	_, err := InitKVStore()
	if err != nil {
		t.Errorf("Could not create Badger DB: %v", err)
	}

	err = BatchSet(&KVPairs{"key": "oyster"})
	if err != nil {
		t.Errorf("Could not set key: %v", err)
	}

	kvs, err := BatchGet(&KVKeys{"key"})
	if err != nil {
		t.Errorf("Could not get key: %v", err)
	}

	val := (*kvs)["key"]
	if val != "oyster" {
		t.Errorf("Key value incorrect: %v != %v", val, "oyster")
	}

}

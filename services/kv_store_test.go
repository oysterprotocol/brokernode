package services

import (
	"testing"

	"github.com/dgraph-io/badger"
)

func Test_KVStore(t *testing.T) {
	db, err := InitKVStore()
	if err != nil {
		t.Errorf("Could not create Badger DB: %v", err)
	}

	err = BatchSet([]KVPair{KVPair{Key: "key", Val: "oyster"}})
	if err != nil {
		t.Errorf("Could not set key: %v", err)
	}

	var val string
	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("key"))
		if err != nil {
			return err
		}

		valBytes, err := item.Value()
		if err != nil {
			return err
		}
		val = string(valBytes)

		return nil
	})
	if err != nil {
		t.Errorf("Could not get key: %v", err)
	}

	if val != "oyster" {
		t.Errorf("Key value incorrect: %v != %v", val, "oyster")
	}

}

package services_test

import (
	"testing"

	"github.com/dgraph-io/badger"
	"github.com/oysterprotocol/brokernode/services"
)

func Test_InitKVStore(t *testing.T) {
	db, err := services.InitKVStore()
	if err != nil {
		t.Errorf("Could not create Badger DB: %v", err)
	}

	err = db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte("key"), []byte("oyster"))
		return err
	})

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

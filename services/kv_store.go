package services

import (
	"errors"
	"os"

	"github.com/dgraph-io/badger"
)

// const badgerDir = "/tmp/badger" // TODO: CHANGE THIS.
const badgerDir = "/var/lib/badger/prod"
const badgerDirTest = "/var/lib/badger/test"

// Singleton DB
var badgerDB *badger.DB

type KVPairs map[string]string
type KVKeys []string

// returns db so that caller can close connection when done.
func InitKVStore() (db *badger.DB, err error) {
	if badgerDB != nil {
		return badgerDB, nil
	}

	// Setup opts
	opts := badger.DefaultOptions

	if os.Getenv("GO_ENV") == "test" {
		opts.Dir = badgerDirTest
		opts.ValueDir = badgerDirTest
	} else {
		opts.Dir = badgerDir
		opts.ValueDir = badgerDir
	}

	db, err = badger.Open(opts)
	badgerDB = db

	return db, err
}

func BatchGet(ks *KVKeys) (kvs *KVPairs, err error) {
	kvs = &KVPairs{}
	if badgerDB == nil {
		return kvs, errors.New("badgerDB not initialized")
	}

	err = badgerDB.View(func(txn *badger.Txn) error {
		for _, k := range *ks {
			item, err := txn.Get([]byte(k))
			if err != nil {
				return err
			}

			val := ""
			if item != nil {
				valBytes, err := item.Value()
				if err != nil {
					return err
				}

				val = string(valBytes)
			}

			// Mutate KV map
			(*kvs)[k] = val
		}

		return nil
	})

	return
}

func BatchSet(kvs *KVPairs) (err error) {
	if badgerDB == nil {
		return errors.New("badgerDB not initialized")
	}

	return badgerDB.Update(func(txn *badger.Txn) error {
		for k, v := range *kvs {
			err := txn.Set([]byte(k), []byte(v))
			if err != nil {
				return err
			}

		}

		return nil
	})
}

package services

import (
	"os"

	"github.com/dgraph-io/badger"
)

// const badgerDir = "/tmp/badger" // TODO: CHANGE THIS.
const badgerDir = "/var/lib/badger/prod"
const badgerDirTest = "/var/lib/badger/test"

// Singleton DB
var badgerDB *badger.DB

type KVPair struct {
	Key string
	Val string
}

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

func BatchSet(kvs []KVPair) (err error) {
	return badgerDB.Update(func(txn *badger.Txn) error {
		for _, kv := range kvs {
			err := txn.Set([]byte(kv.Key), []byte(kv.Val))
			if err != nil {
				return err
			}

		}

		return nil
	})
}

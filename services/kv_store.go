package services

import (
	"errors"
	"fmt"
	"os"

	"github.com/dgraph-io/badger"
	"github.com/oysterprotocol/brokernode/utils"
)

// const badgerDir = "/tmp/badger" // TODO: CHANGE THIS.
const badgerDir = "/var/lib/badger/prod"
const badgerDirTest = "/var/lib/badger/test"

// Singleton DB
var badgerDB *badger.DB

var isKvStoreEnable bool

type KVPairs map[string]string
type KVKeys []string

func init() {
	// Currently disable it.
	isKvStoreEnable = false

	if isKvStoreEnable {
		InitKVStore()
	}
}

/* InitKVStore returns db so that caller can close connection when done.*/
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
	oyster_utils.LogIfError(err, nil)
	badgerDB = db

	return db, err
}

/*IsKvStoreEnabled returns true if KVStore is enabled. Check this before calling BatchGet/BatchSet.*/
func IsKvStoreEnabled() bool {
	return isKvStoreEnable
}

/*BatchGet returns KVPairs for a set of keys. Return partial result if error concurs.*/
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
	oyster_utils.LogIfError(err, map[string]interface{}{"batchSize": len(*ks)})

	return
}

/*BatchSet updates a set of KVPairs. Return error even partial result is updated.*/
func BatchSet(kvs *KVPairs) error {
	if badgerDB == nil {
		return errors.New("badgerDB not initialized")
	}

	err := badgerDB.Update(func(txn *badger.Txn) error {
		for k, v := range *kvs {
			if err := txn.Set([]byte(k), []byte(v)); err != nil {
				return err
			}
		}
		return nil
	})
	oyster_utils.LogIfError(err, map[string]interface{}{"batchSize": len(*kvs)})
	return err
}

/*GenKvStoreKey returns the key for inserting to KV-Store for data_maps.*/
func GenKvStoreKey(gensisHash string, chunkIdx int) string {
	return fmt.Sprintf("%s_%d", gensisHash, chunkIdx)
}

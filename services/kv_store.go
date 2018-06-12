package services

import (
	"errors"
	"os"

	"github.com/dgraph-io/badger"
	"github.com/oysterprotocol/brokernode/utils"
)

// const badgerDir = "/tmp/badger" // TODO: CHANGE THIS.
const badgerDir = "/var/lib/badger/prod"
const badgerDirTest = "/var/lib/badger/test"

// Singleton DB
var badgerDB *badger.DB
var dbNoInitError error
var isKvStoreEnable bool

type KVPairs map[string]string
type KVKeys []string

func init() {
	dbNoInitError = errors.New("badgerDB not initialized, Call InitKvStore() first")

	// Currently disable it.
	isKvStoreEnable = false
}

/*InitKvStore returns db so that caller can call CloseKvStore to close it when it is done.*/
func InitKvStore() (err error) {
	if badgerDB != nil {
		return nil
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

	badgerDB, err = badger.Open(opts)
	oyster_utils.LogIfError(err, nil)
	return err
}

/*CloseKvStore closes the db.*/
func CloseKvStore() {
	err := badgerDB.Close()
	oyster_utils.LogIfError(err, nil)
	badgerDB = nil
}

/*IsKvStoreEnabled returns true if KVStore is enabled. Check this before calling BatchGet/BatchSet.*/
func IsKvStoreEnabled() bool {
	return isKvStoreEnable
}

/*BatchGet returns KVPairs for a set of keys. It won't treat Key missing as error.*/
func BatchGet(ks *KVKeys) (kvs *KVPairs, err error) {
	kvs = &KVPairs{}
	if badgerDB == nil {
		return kvs, dbNoInitError
	}

	err = badgerDB.View(func(txn *badger.Txn) error {
		for _, k := range *ks {
			item, err := txn.Get([]byte(k))
			if err == badger.ErrKeyNotFound {
				continue
			}
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

/*BatchSet updates a set of KVPairs. Return error if any fails.*/
func BatchSet(kvs *KVPairs) error {
	if badgerDB == nil {
		return dbNoInitError
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

/*BatchDelete deletes a set of KVKeys, Return error if any fails.*/
func BatchDelete(ks *KVKeys) error {
	if badgerDB == nil {
		return dbNoInitError
	}

	err := badgerDB.Update(func(txn *badger.Txn) error {
		for _, key := range *ks {
			if err := txn.Delete([]byte(key)); err != nil {
				return err
			}
		}
		return nil
	})
	oyster_utils.LogIfError(err, map[string]interface{}{"batchSize": len(*ks)})
	return err
}

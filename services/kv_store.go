package services

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/dgraph-io/badger"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/utils"
)

// const badgerDir = "/tmp/badger" // TODO: CHANGE THIS.
const badgerDir = "/var/lib/badger/prod"

// Singleton DB
var badgerDB *badger.DB
var dbNoInitError error
var isKvStoreEnable bool
var badgerDirTest string

type KVPairs map[string]string
type KVKeys []string

func init() {
	dbNoInitError = errors.New("badgerDB not initialized, Call InitKvStore() first")

	badgerDirTest, _ = ioutil.TempDir("", "badgerForUnitTest")

	// enable unless explicitly disabled in .env file
	isKvStoreEnable = os.Getenv("KEY_VALUE_STORE_ENABLED") != "false"

	if IsKvStoreEnabled() {
		InitKvStore()
	}
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
func CloseKvStore() error {
	if badgerDB == nil {
		return nil
	}
	err := badgerDB.Close()
	oyster_utils.LogIfError(err, nil)
	badgerDB = nil
	return err
}

/*RemoveAllKvStoreData removes all the data. Caller should call InitKvStore() again to create a new one.*/
func RemoveAllKvStoreData() error {
	if err := CloseKvStore(); err != nil {
		return err
	}

	var dir string
	if os.Getenv("GO_ENV") == "test" {
		dir = badgerDirTest
	} else {
		dir = badgerDir
	}
	err := os.RemoveAll(dir)
	oyster_utils.LogIfError(err, map[string]interface{}{"badgerDir": dir})
	return err
}

/*GetBadgerDb returns the underlying the database. If not call InitKvStore(), it will return nil*/
func GetBadgerDb() *badger.DB {
	return badgerDB
}

/*IsKvStoreEnabled returns true if KVStore is enabled. Check this before calling BatchGet/BatchSet.*/
func IsKvStoreEnabled() bool {
	return isKvStoreEnable
}

/*DataMapGet returns the message reference by dataMap.*/
func GetMessageFromDataMap(dataMap models.DataMap) string {
	value := dataMap.Message
	if IsKvStoreEnabled() {
		values, _ := BatchGet(&KVKeys{dataMap.MsgID})
		if v, hasKey := (*values)[dataMap.MsgID]; hasKey {
			value = v
		}
	}

	if dataMap.MsgStatus == models.MsgStatusUploadedHaveNotEncoded {
		message, _ := oyster_utils.ChunkMessageToTrytesWithStopper(value)
		value = string(message)
	}

	return value
}

/*BatchGet returns KVPairs for a set of keys. It won't treat Key missing as error.*/
func BatchGet(ks *KVKeys) (kvs *KVPairs, err error) {
	kvs = &KVPairs{}
	if badgerDB == nil {
		return kvs, dbNoInitError
	}

	err = badgerDB.View(func(txn *badger.Txn) error {
		for _, k := range *ks {
			// Skip any empty keys.
			if k == "" {
				continue
			}

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

	var err error
	txn := badgerDB.NewTransaction(true)
	for k, v := range *kvs {
		if k == "" {
			err = errors.New("BatchSet does not accept key as empty string")
			break
		}

		e := txn.Set([]byte(k), []byte(v))
		if e == nil {
			continue
		}

		if e == badger.ErrTxnTooBig {
			e = nil
			if commitErr := txn.Commit(nil); commitErr != nil {
				e = commitErr
			} else {
				txn = badgerDB.NewTransaction(true)
				e = txn.Set([]byte(k), []byte(v))
			}
		}

		if e != nil {
			err = e
			break
		}
	}

	defer txn.Discard()
	if err == nil {
		err = txn.Commit(nil)
	}
	oyster_utils.LogIfError(err, map[string]interface{}{"batchSize": len(*kvs)})
	return err
}

/*BatchDelete deletes a set of KVKeys, Return error if any fails.*/
func BatchDelete(ks *KVKeys) error {
	if badgerDB == nil {
		return dbNoInitError
	}

	var err error
	txn := badgerDB.NewTransaction(true)
	for _, key := range *ks {
		e := txn.Delete([]byte(key))
		if e == nil {
			continue
		}

		if e == badger.ErrTxnTooBig {
			e = nil
			if commitErr := txn.Commit(nil); commitErr != nil {
				e = commitErr
			} else {
				txn = badgerDB.NewTransaction(true)
				e = txn.Delete([]byte(key))
			}
		}

		if e != nil {
			err = e
			break
		}
	}

	defer txn.Discard()
	if err == nil {
		err = txn.Commit(nil)
	}

	oyster_utils.LogIfError(err, map[string]interface{}{"batchSize": len(*ks)})
	return err
}

/*DeleteMsgDatas deletes the data referred by dataMaps. */
func DeleteMsgDatas(dataMaps []models.DataMap) {
	if !IsKvStoreEnabled() {
		return
	}

	var keys KVKeys
	for _, dm := range dataMaps {
		keys = append(keys, dm.MsgID)
	}
	BatchDelete(&keys)
}

package services

import (
	"errors"
	"io/ioutil"
	"os"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/utils"
)

// const badgerDir = "/tmp/badger" // TODO: CHANGE THIS.
const badgerDir = "/var/lib/badger/prod"
const CompletedDir = "completed"
const InProgressDir = "current"
const HashDir = "hash"
const MessageDir = "message"

// Singleton DB
var badgerDB *badger.DB
var dbNoInitError error
var isKvStoreEnable bool
var badgerDirTest string
var DBMap KVDBMap

type KVDBMap map[string]*badger.DB
type KVPairs map[string]string
type KVKeys []string

type ChunkData struct {
	Address string
	Hash    string
	Message string
}

func init() {
	DBMap = make(KVDBMap)
	dbNoInitError = errors.New("badgerDB not initialized, Call InitKvStore() first")

	badgerDirTest, _ = ioutil.TempDir("", "badgerForUnitTest")

	// enable unless explicitly disabled in .env file
	isKvStoreEnable = os.Getenv("KEY_VALUE_STORE_ENABLED") != "false"

	if IsKvStoreEnabled() {
		err := InitKvStore()
		// If error in init the KV store. Just crash and fail the entirely process and wait for restart.
		if err != nil {
			panic(err.Error())
		}
	}
}

/*InitUniqueKvStore creates a new K:V store associated with a particular upload*/
func InitUniqueKvStore(dirName string) error {
	if DBMap[dirName] != nil {
		return nil
	}

	opts := badger.DefaultOptions

	if os.Getenv("GO_ENV") == "test" {
		opts.Dir = badgerDirTest + "/" + dirName
		opts.ValueDir = badgerDirTest + "/" + dirName
	} else {
		opts.Dir = badgerDir + "/" + dirName
		opts.ValueDir = badgerDir + "/" + dirName
	}

	db, err := badger.Open(opts)
	oyster_utils.LogIfError(err, nil)
	if err == nil {
		DBMap[dirName] = db
	}
	return err
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

/*CloseUniqueKvStore closes the K:V store associated with a particular upload.*/
func CloseUniqueKvStore(dirName string) error {
	if DBMap[dirName] == nil {
		return nil
	}
	err := DBMap[dirName].Close()
	oyster_utils.LogIfError(err, nil)
	DBMap[dirName] = nil
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

/*RemoveAllUniqueKvStoreData removes all the data associated with a particular K:V store.*/
func RemoveAllUniqueKvStoreData(dirName string) error {
	if err := CloseUniqueKvStore(dirName); err != nil {
		return err
	}

	var dir string
	if os.Getenv("GO_ENV") == "test" {
		dir = badgerDirTest + "/" + dirName
	} else {
		dir = badgerDir + "/" + dirName
	}

	err := os.RemoveAll(dir)
	oyster_utils.LogIfError(err, map[string]interface{}{"badgerDir": dir})
	return err
}

/*RemoveAllKvStoreDataFromAllKvStores removes all the data associated with all K:V stores.*/
func RemoveAllKvStoreDataFromAllKvStores() []error {
	var errArray []error
	for dirName := range DBMap {
		if err := CloseUniqueKvStore(dirName); err != nil {
			errArray = append(errArray, err)
			continue
		}

		var dir string
		if os.Getenv("GO_ENV") == "test" {
			dir = badgerDirTest + "/" + dirName
		} else {
			dir = badgerDir + "/" + dirName
		}
		err := os.RemoveAll(dir)
		oyster_utils.LogIfError(err, map[string]interface{}{"badgerDir": dir})
	}
	return errArray
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

/*GetUniqueBadgerDb returns a database associated with an upload.  If not initialized this will return nil. */
func GetUniqueBadgerDb(dirName string) *badger.DB {
	return DBMap[dirName]
}

/*GetOrInitUniqueBadgerDB returns a database associated with an upload. */
func GetOrInitUniqueBadgerDB(dirName string) *badger.DB {
	db := GetUniqueBadgerDb(dirName)
	if db != nil {
		return db
	}

	err := InitUniqueKvStore(dirName)
	oyster_utils.LogIfError(err, nil)
	return GetUniqueBadgerDb(dirName)
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

/*GetChunkData returns the message, hash, and address for a chunk.*/
func GetChunkData(prefix string, genesisHash string, chunkIdx int) ChunkData {

	key := oyster_utils.GenerateBadgerKey(prefix, genesisHash, chunkIdx)

	hashDBDir := prefix + "/" + HashDir + "/" + genesisHash
	messageDBDir := prefix + "/" + MessageDir + "/" + genesisHash

	GetOrInitUniqueBadgerDB(hashDBDir)
	GetOrInitUniqueBadgerDB(messageDBDir)

	hash := ""
	message := ""

	hashValues, _ := BatchGetFromUniqueDB(hashDBDir,
		&KVKeys{key})
	if v, hasKey := (*hashValues)[key]; hasKey {
		hash = v
	}

	messageValues, _ := BatchGetFromUniqueDB(messageDBDir,
		&KVKeys{key})
	if v, hasKey := (*messageValues)[key]; hasKey {
		message = v
	}

	trytes, _ := oyster_utils.ChunkMessageToTrytesWithStopper(message)
	message = string(trytes)

	address := oyster_utils.Sha256ToAddress(hash)

	return ChunkData{
		Address: address,
		Message: message,
		Hash:    hash,
	}
}

/*BatchGetFromUniqueDB returns KVPairs for a set of keys from a specific DB.
It won't treat Key missing as error.*/
func BatchGetFromUniqueDB(dirName string, ks *KVKeys) (kvs *KVPairs, err error) {
	kvs = &KVPairs{}
	if DBMap[dirName] == nil {
		DBMap[dirName] = GetOrInitUniqueBadgerDB(dirName)
	}

	err = DBMap[dirName].View(func(txn *badger.Txn) error {
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

/*BatchSetToUniqueDB updates a set of KVPairs in a unique database.
Return error if any fails.*/
func BatchSetToUniqueDB(dirName string, kvs *KVPairs, ttl time.Duration) error {
	if DBMap[dirName] == nil {
		DBMap[dirName] = GetOrInitUniqueBadgerDB(dirName)
	}

	var err error
	txn := DBMap[dirName].NewTransaction(true)
	for k, v := range *kvs {
		if k == "" {
			err = errors.New("BatchSetToUniqueDB does not accept key as empty string")
			break
		}

		e := txn.SetWithTTL([]byte(k), []byte(v), ttl)
		if e == nil {
			continue
		}

		if e == badger.ErrTxnTooBig {
			e = nil
			if commitErr := txn.Commit(nil); commitErr != nil {
				e = commitErr
			} else {
				txn = DBMap[dirName].NewTransaction(true)
				e = txn.SetWithTTL([]byte(k), []byte(v), ttl)
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

/*BatchSet updates a set of KVPairs. Return error if any fails.*/
func BatchSet(kvs *KVPairs, ttl time.Duration) error {
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

		e := txn.SetWithTTL([]byte(k), []byte(v), ttl)
		if e == nil {
			continue
		}

		if e == badger.ErrTxnTooBig {
			e = nil
			if commitErr := txn.Commit(nil); commitErr != nil {
				e = commitErr
			} else {
				txn = badgerDB.NewTransaction(true)
				e = txn.SetWithTTL([]byte(k), []byte(v), ttl)
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

/*BatchDeleteFromUniqueDB deletes a set of KVKeys from a specific DB.
Return error if any fails.*/
func BatchDeleteFromUniqueDB(dirName string, ks *KVKeys) error {
	if DBMap[dirName] == nil {
		DBMap[dirName] = GetOrInitUniqueBadgerDB(dirName)
	}

	var err error
	txn := DBMap[dirName].NewTransaction(true)
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
				txn = DBMap[dirName].NewTransaction(true)
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

/*DeleteDataFromUniqueDB deletes the data from a specific DB for all
chunks within a certain range.  Note that endingIdx is inclusive */
func DeleteDataFromUniqueDB(dirName string, prefix string, genesisHash string, startingIdx int,
	endingIdx int) {

	var keys KVKeys
	step := 0
	stop := 0

	if startingIdx < endingIdx {
		step = 1
		stop = endingIdx + step
	} else if startingIdx > endingIdx {
		step = -1
		stop = endingIdx + step
	} else {
		keys = append(keys, oyster_utils.GenerateBadgerKey(prefix, genesisHash, startingIdx))
	}

	for i := startingIdx; i != stop; i = i + step {
		keys = append(keys, oyster_utils.GenerateBadgerKey(prefix, genesisHash, i))
	}

	BatchDeleteFromUniqueDB(dirName, &keys)
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

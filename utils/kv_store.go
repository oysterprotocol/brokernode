package oyster_utils

import (
	"errors"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dgraph-io/badger"
)

// const badgerDir = "/tmp/badger" // TODO: CHANGE THIS.
const badgerDir = "/var/lib/badger/prod"

/*CompletedDir is a directory for completed data maps*/
const CompletedDir = "complete"

/*InProgressDir is a directory for chunks we are still working on*/
const InProgressDir = "current"

/*HashDir is a directory for hashes*/
const HashDir = "hash"

/*MessageDir is a directory for messages*/
const MessageDir = "message"

/*KeyDelimiter is a delimiter character used in badger keys*/
const KeyDelimiter = '_'

/*TestValueTimeToLive is some default value we can use in unit
tests for K:V pairs in badger*/
const TestValueTimeToLive = 3 * time.Minute

// Singleton DB
var badgerDB *badger.DB
var dbNoInitError error
var isKvStoreEnable bool
var badgerDirTest string
var dbMap KVDBMap
var prodDBMap KVDBMap
var unitTestDBMap KVDBMap

/*KVDBMap is a type which is a map with strings as keys and DBData as the value.  We use this type to make a
data structure tracking all the unique DBs that are still alive*/
type KVDBMap map[string]DBData

/*KVPairs is a type.  Map key strings to value strings*/
type KVPairs map[string]string

/*KVKeys is a type.  An array of key strings*/
type KVKeys []string

/*DBData defines what data to expect for each database stored in DVDBMap*/
type DBData struct {
	DatabaseName  string
	DirectoryPath string
	Database      *badger.DB
}

/*ChunkData is the type of response we will give when a caller wants data about a specific chunk*/
type ChunkData struct {
	Address     string
	Hash        string
	Message     string
	RawMessage  string
	Idx         int64
	GenesisHash string
}

func init() {

	prodDBMap = make(KVDBMap)
	unitTestDBMap = make(KVDBMap)
	dbMap = make(KVDBMap)

	dbNoInitError = errors.New("badgerDB not initialized, Call InitKvStore() first")

	badgerDirTest, _ = ioutil.TempDir("", "badgerForUnitTest")

	// enable unless explicitly disabled in .env file
	isKvStoreEnable = os.Getenv("KEY_VALUE_STORE_ENABLED") != "false"

	if DataMapStorageMode == DataMapsInSQL {
		err := InitKvStore()
		// If error in init the KV store. Just crash and fail the entirely process and wait for restart.
		if err != nil {
			panic(err.Error())
		}
	}
}

func getDBMap() *KVDBMap {
	if os.Getenv("GO_ENV") == "test" {
		return &unitTestDBMap
	} else {
		return &prodDBMap
	}
}

//
func GenerateBulkKeys(genHash string, startingIdx int64, endingIdx int64) *KVKeys {

	var keys KVKeys
	step := int64(0)
	stop := int64(0)

	if startingIdx < endingIdx {
		step = 1
		stop = endingIdx + step
	} else if startingIdx > endingIdx {
		step = -1
		stop = endingIdx + step
	} else {
		keys = append(keys, GetBadgerKey([]string{genHash, strconv.FormatInt(int64(startingIdx), 10)}))
	}

	for i := startingIdx; i != stop; i = i + step {
		keys = append(keys, GetBadgerKey([]string{genHash, strconv.FormatInt(int64(i), 10)}))
	}

	return &keys
}

func GetChunkIdxFromKey(key string) int64 {
	s := key
	if i := strings.LastIndexByte(key, KeyDelimiter); i >= 0 {
		s = s[i+1:]
	}
	i, err := strconv.ParseInt(s, 10, 64)
	LogIfError(err, nil)
	return i
}

/*GetBadgerDirName will make a directory path from an array of strings.
will/look/like/this
*/
func GetBadgerDirName(dirs []string) string {
	return buildBadgerName(dirs, string(os.PathSeparator))
}

/*GetBadgerDBName will make a DB name from an array of strings
will_look_like_this
*/
func GetBadgerDBName(names []string) string {
	return buildBadgerName(names, string(KeyDelimiter))
}

/*GetBadgerKey will make a key for a key value pair from an array of strings
someGenHash_1
*/
func GetBadgerKey(keyStrings []string) string {
	return buildBadgerName(keyStrings, string(KeyDelimiter)+string(KeyDelimiter))
}

func buildBadgerName(names []string, separator string) string {
	returnName := names[0]

	for i := 1; i < len(names); i++ {
		returnName += separator + names[i]
	}

	return returnName
}

/*InitUniqueKvStore creates a new K:V store associated with a particular upload*/
func InitUniqueKvStore(dbID []string) error {
	dbName := GetBadgerDBName(dbID)
	dirPath := GetBadgerDirName(dbID)
	if dbMap[dbName].Database != nil {
		return nil
	}

	opts := badger.DefaultOptions

	if os.Getenv("GO_ENV") == "test" {

		dir := badgerDirTest + string(os.PathSeparator) + dirPath

		if _, err := os.Stat(dir); os.IsNotExist(err) {
			os.MkdirAll(dir, os.ModeDir)
		}

		opts.Dir = dir
		opts.ValueDir = dir
	} else {
		dir := badgerDir + string(os.PathSeparator) + dirPath

		if _, err := os.Stat(dir); os.IsNotExist(err) {
			os.MkdirAll(dir, os.ModeDir)
		}

		opts.Dir = dir
		opts.ValueDir = dir
	}

	db, err := badger.Open(opts)
	LogIfError(err, nil)

	if err == nil {
		dbMap[dbName] = DBData{
			Database:      db,
			DatabaseName:  dbName,
			DirectoryPath: dirPath,
		}
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
	LogIfError(err, nil)
	return err
}

/*CloseUniqueKvStore closes the K:V store associated with a particular upload.*/
func CloseUniqueKvStore(dbName string) error {
	if dbMap[dbName].Database == nil {
		return nil
	}
	err := dbMap[dbName].Database.Close()
	LogIfError(err, nil)

	if _, ok := dbMap[dbName]; ok {
		delete(dbMap, dbName)
	}

	return err
}

/*CloseKvStore closes the db.*/
func CloseKvStore() error {
	if badgerDB == nil {
		return nil
	}
	err := badgerDB.Close()
	LogIfError(err, nil)
	badgerDB = nil
	return err
}

/*RemoveAllUniqueKvStoreData removes all the data associated with a particular K:V store.*/
func RemoveAllUniqueKvStoreData(dbName string) error {
	directoryPath := dbMap[dbName].DirectoryPath

	if err := CloseUniqueKvStore(dbName); err != nil {
		return err
	}

	var dir string
	if os.Getenv("GO_ENV") == "test" {
		dir = badgerDirTest + string(os.PathSeparator) + directoryPath
	} else {
		dir = badgerDir + string(os.PathSeparator) + directoryPath
	}

	err := os.RemoveAll(dir)

	LogIfError(err, map[string]interface{}{"badgerDir": dir})
	return err
}

/*RemoveAllKvStoreDataFromAllKvStores removes all the data associated with all K:V stores.*/
func RemoveAllKvStoreDataFromAllKvStores() []error {
	var errArray []error
	for dbName := range dbMap {
		directoryPath := dbMap[dbName].DirectoryPath
		if err := CloseUniqueKvStore(dbName); err != nil {
			errArray = append(errArray, err)
			continue
		}

		var dir string
		if os.Getenv("GO_ENV") == "test" {
			dir = badgerDirTest + string(os.PathSeparator) + directoryPath
		} else {
			dir = badgerDir + string(os.PathSeparator) + directoryPath
		}
		err := os.RemoveAll(dir)
		LogIfError(err, map[string]interface{}{"badgerDir": dir})
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
	LogIfError(err, map[string]interface{}{"badgerDir": dir})
	return err
}

/*GetUniqueBadgerDb returns a database associated with an upload.  If not initialized this will return nil. */
func GetUniqueBadgerDb(dbName string) *badger.DB {
	return dbMap[dbName].Database
}

/*GetOrInitUniqueBadgerDB returns a database associated with an upload. */
func GetOrInitUniqueBadgerDB(dbID []string) *badger.DB {
	dbName := GetBadgerDBName(dbID)

	db := GetUniqueBadgerDb(dbName)
	if db != nil {
		return db
	}

	err := InitUniqueKvStore(dbID)

	if err == badger.ErrConflict {

	}

	LogIfError(err, nil)
	return GetUniqueBadgerDb(dbName)
}

/*GetBadgerDb returns the underlying the database. If not call InitKvStore(), it will return nil*/
func GetBadgerDb() *badger.DB {
	return badgerDB
}

/*IsKvStoreEnabled returns true if KVStore is enabled. Check this before calling BatchGet/BatchSet.*/
func IsKvStoreEnabled() bool {
	return isKvStoreEnable
}

/*GetChunkData returns the message, hash, and address for a chunk.*/
func GetChunkData(prefix string, genesisHash string, chunkIdx int64) ChunkData {

	rawMessage := GetMessageData(prefix, genesisHash, chunkIdx)
	hash := GetHashData(prefix, genesisHash, chunkIdx)

	address := ""
	message := ""

	if rawMessage != "" {
		trytes, _ := ChunkMessageToTrytesWithStopper(rawMessage)
		message = string(trytes)
	}

	if hash != "" {
		address = Sha256ToAddress(hash)
	}

	return ChunkData{
		Address:     address,
		RawMessage:  rawMessage,
		Message:     message,
		Hash:        hash,
		Idx:         chunkIdx,
		GenesisHash: genesisHash,
	}
}

func GetMessageData(prefix string, genesisHash string, chunkIdx int64) string {

	key := GetBadgerKey([]string{genesisHash, strconv.FormatInt(int64(chunkIdx), 10)})

	rawMessage := ""

	msgValues, _ := BatchGetFromUniqueDB([]string{prefix, genesisHash, MessageDir},
		&KVKeys{key})
	if v, hasKey := (*msgValues)[key]; hasKey {
		rawMessage = v
	}

	return rawMessage
}

func GetHashData(prefix string, genesisHash string, chunkIdx int64) string {

	key := GetBadgerKey([]string{genesisHash, strconv.FormatInt(int64(chunkIdx), 10)})

	hash := ""

	hashValues, _ := BatchGetFromUniqueDB([]string{prefix, genesisHash, HashDir},
		&KVKeys{key})
	if v, hasKey := (*hashValues)[key]; hasKey {
		hash = v
	}

	return hash
}

/*GetBulkChunkData returns the message, hash, and address for a large number of chunks.*/
func GetBulkChunkData(prefix string, genesisHash string, ks *KVKeys) ([]ChunkData, error) {

	var chunkData []ChunkData

	hashValues, errHash := BatchGetFromUniqueDB([]string{prefix, genesisHash, HashDir},
		ks)
	LogIfError(errHash, nil)

	messageValues, errMessage := BatchGetFromUniqueDB([]string{prefix, genesisHash, MessageDir},
		ks)
	LogIfError(errMessage, nil)

	if errHash != nil {
		return chunkData, errHash
	}
	if errMessage != nil {
		return chunkData, errMessage
	}

	for _, key := range *(ks) {
		_, hasMessageKey := (*messageValues)[key]
		_, hasHashKey := (*hashValues)[key]

		if hasMessageKey && hasHashKey {
			message := ""
			trytes, _ := ChunkMessageToTrytesWithStopper((*messageValues)[key])
			message = string(trytes)

			address := Sha256ToAddress((*hashValues)[key])
			chunkData = append(chunkData, ChunkData{
				Address:     address,
				Hash:        (*hashValues)[key],
				Message:     message,
				RawMessage:  (*messageValues)[key],
				Idx:         GetChunkIdxFromKey(key),
				GenesisHash: genesisHash,
			})
		}
	}
	return chunkData, nil
}

/*BatchGetFromUniqueDB returns KVPairs for a set of keys from a specific DB.
It won't treat Key missing as error.*/
func BatchGetFromUniqueDB(dbID []string, ks *KVKeys) (kvs *KVPairs, err error) {
	kvs = &KVPairs{}
	dbName := GetBadgerDBName(dbID)
	var db *badger.DB
	if dbMap[dbName].Database == nil {
		db = GetOrInitUniqueBadgerDB(dbID)
	} else {
		db = dbMap[dbName].Database
	}

	if db == nil {
		return kvs, errors.New("nil database")
	}

	err = db.View(func(txn *badger.Txn) error {
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
	LogIfError(err, map[string]interface{}{"batchSize": len(*ks)})

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
	LogIfError(err, map[string]interface{}{"batchSize": len(*ks)})

	return
}

/*BatchSetToUniqueDB updates a set of KVPairs in a unique database.
Return error if any fails.*/
func BatchSetToUniqueDB(dbID []string, kvs *KVPairs, ttl time.Duration) error {
	ttl = getTTL(ttl)
	dbName := GetBadgerDBName(dbID)
	var db *badger.DB
	if dbMap[dbName].Database == nil {
		db = GetOrInitUniqueBadgerDB(dbID)
	} else {
		db = dbMap[dbName].Database
	}

	var err error
	txn := db.NewTransaction(true)
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
				txn = db.NewTransaction(true)
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
	LogIfError(err, map[string]interface{}{"batchSize": len(*kvs)})
	return err
}

/*BatchSet updates a set of KVPairs. Return error if any fails.*/
func BatchSet(kvs *KVPairs, ttl time.Duration) error {
	ttl = getTTL(ttl)
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
	LogIfError(err, map[string]interface{}{"batchSize": len(*kvs)})
	return err
}

/*BatchDeleteFromUniqueDB deletes a set of KVKeys from a specific DB.
Return error if any fails.*/
func BatchDeleteFromUniqueDB(dbID []string, ks *KVKeys) error {
	dbName := GetBadgerDBName(dbID)
	var db *badger.DB
	if dbMap[dbName].Database == nil {
		db = GetOrInitUniqueBadgerDB(dbID)
	} else {
		db = dbMap[dbName].Database
	}

	var err error
	txn := db.NewTransaction(true)
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
				txn = db.NewTransaction(true)
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

	LogIfError(err, map[string]interface{}{"batchSize": len(*ks)})
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

	LogIfError(err, map[string]interface{}{"batchSize": len(*ks)})
	return err
}

/*DeleteDataFromUniqueDB deletes the data from a specific DB for all
chunks within a certain range.  Note that endingIdx is inclusive */
func DeleteDataFromUniqueDB(dbID []string, genesisHash string, startingIdx int64,
	endingIdx int64) error {

	keys := GenerateBulkKeys(genesisHash, startingIdx, endingIdx)

	err := BatchDeleteFromUniqueDB(dbID, keys)
	return err
}

func AllChunkDataHasArrived(chunkData ChunkData) bool {
	return chunkData.Address != "" && chunkData.Message != "" && chunkData.Hash != ""
}

func getTTL(ttl time.Duration) time.Duration {
	if os.Getenv("GO_ENV") != "test" {
		return ttl
	} else {
		return TestValueTimeToLive
	}
}

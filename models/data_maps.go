package models

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/oysterprotocol/brokernode/utils"
	"golang.org/x/crypto/sha3"
)

const (
	FileBytesChunkSize = float64(2187)

	MaxSideChainLength = 1000 // need to determine what this number should be

	DataMapTableName = "data_maps"

	// The max number of values to insert to db via Sql: INSERT INTO table_name VALUES.
	MaxNumberOfValueForInsertOperation = 10
)

const (
	Pending int = iota + 1
	Unassigned
	Unverified
	Complete
	Confirmed
	Error = -1
)

const (
	MsgStatusUnmigrated int = iota
	MsgStatusNotUploaded
	MsgStatusUploaded
)

type DataMap struct {
	ID             uuid.UUID `json:"id" db:"id"`
	CreatedAt      time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time `json:"updatedAt" db:"updated_at"`
	Status         int       `json:"status" db:"status"`
	NodeID         string    `json:"nodeID" db:"node_id"`
	NodeType       string    `json:"nodeType" db:"node_type"`
	Message        string    `json:"message" db:"message"`
	MsgID          string    `json:"msgId" db:"msg_id"`
	MsgStatus      int       `json:"msgStatus" db:"msg_status"`
	TrunkTx        string    `json:"trunkTx" db:"trunk_tx"`
	BranchTx       string    `json:"branchTx" db:"branch_tx"`
	GenesisHash    string    `json:"genesisHash" db:"genesis_hash"`
	ChunkIdx       int       `json:"chunkIdx" db:"chunk_idx"`
	Hash           string    `json:"hash" db:"hash"`
	ObfuscatedHash string    `json:"obfuscatedHash" db:"obfuscated_hash"`
	Address        string    `json:"address" db:"address"`
}

type TypeAndChunkMap struct {
	Type   int       `json:"type"`
	Chunks []DataMap `json:"chunks"`
}

var SortOrder map[int]string

var (
	TreasurePrefix        = hex.EncodeToString([]byte("Treasure: "))
	TreasurePayloadLength = len(TreasurePrefix) + 96
	TreasureChunkPadding  = int(FileBytesChunkSize) - TreasurePayloadLength
)

func init() {
	SortOrder = make(map[int]string, 2)
	SortOrder[SessionTypeAlpha] = "asc"
	SortOrder[SessionTypeBeta] = "desc"
}

// String is not required by pop and may be deleted
func (d DataMap) String() string {
	jd, _ := json.Marshal(d)
	return string(jd)
}

// DataMaps is not required by pop and may be deleted
type DataMaps []DataMap

// String is not required by pop and may be deleted
func (d DataMaps) String() string {
	jd, _ := json.Marshal(d)
	return string(jd)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (d *DataMap) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: d.GenesisHash, Name: "GenesisHash"},
		&validators.IntIsGreaterThan{Field: d.ChunkIdx, Name: "ChunkIdx", Compared: -1},
		&validators.StringIsPresent{Field: d.Hash, Name: "Hash"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (d *DataMap) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (d *DataMap) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

/**
 * Callbacks
 */

func (d *DataMap) BeforeCreate(tx *pop.Connection) error {
	d.MsgID = d.generateMsgId()

	// After adding MsgStatus column, all previous value will be 0/MsgStatusUnmigrated.
	// And all new value should be default to MsgStatusNotUploaded.
	if d.MsgStatus == MsgStatusUnmigrated {
		d.MsgStatus = MsgStatusNotUploaded
	}
	return nil
}

func (d *DataMap) EncryptEthKey(unencryptedKey string) (string, error) {

	session := UploadSession{}

	err := DB.RawQuery("SELECT * from upload_sessions WHERE genesis_hash = ?", d.GenesisHash).First(&session)
	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return "", err
	}

	hashedSessionID := oyster_utils.HashHex(hex.EncodeToString([]byte(fmt.Sprint(session.ID))), sha3.New256())
	hashedChunkCreationTime := oyster_utils.HashHex(hex.EncodeToString([]byte(fmt.Sprint(d.CreatedAt))), sha3.New256())

	decryptedKey := oyster_utils.Encrypt(hashedSessionID, unencryptedKey, hashedChunkCreationTime)
	return hex.EncodeToString(decryptedKey), nil
}

func (d *DataMap) DecryptEthKey(encryptedKey string) (string, error) {

	session := UploadSession{}

	err := DB.RawQuery("SELECT * from upload_sessions WHERE genesis_hash = ?", d.GenesisHash).First(&session)
	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return "", err
	}

	hashedSessionID := oyster_utils.HashHex(hex.EncodeToString([]byte(fmt.Sprint(session.ID))), sha3.New256())
	hashedChunkCreationTime := oyster_utils.HashHex(hex.EncodeToString([]byte(fmt.Sprint(d.CreatedAt))), sha3.New256())

	decryptedKey := oyster_utils.Decrypt(hashedSessionID, encryptedKey, hashedChunkCreationTime)
	return hex.EncodeToString(decryptedKey), nil
}

func (d *DataMap) generateMsgId() string {
	return fmt.Sprintf("%v__%d", d.GenesisHash, d.ChunkIdx)
}

// Computes a particular sectorIdx addresses in term of DataMaps. Limit by maxNumbOfHashes.
func ComputeSectorDataMapAddress(genHash string, sectorIdx int, maxNumOfHashes int) []string {
	var addr []string

	currHash := genHash
	for i := 0; i < sectorIdx*oyster_utils.FileSectorInChunkSize; i++ {
		currHash = oyster_utils.HashHex(currHash, sha256.New())
	}

	for i := 0; i < maxNumOfHashes; i++ {
		obfuscatedHash := oyster_utils.HashHex(currHash, sha512.New384())
		currAddr := string(oyster_utils.MakeAddress(obfuscatedHash))

		addr = append(addr, currAddr)
		currHash = oyster_utils.HashHex(currHash, sha256.New())
	}
	return addr
}

// BuildDataMaps builds the datamap and inserts them into the DB.
func BuildDataMaps(genHash string, numChunks int) (vErr *validate.Errors, err error) {

	fileChunksCount := numChunks

	operation, _ := oyster_utils.CreateDbUpdateOperation(&DataMap{})
	columnNames := operation.GetColumns()
	var values []string

	currHash := genHash
	insertionCount := 0
	for i := 0; i < fileChunksCount; i++ {

		obfuscatedHash := oyster_utils.HashHex(currHash, sha512.New384())
		currAddr := string(oyster_utils.MakeAddress(obfuscatedHash))

		dataMap := DataMap{
			GenesisHash:    genHash,
			ChunkIdx:       i,
			Hash:           currHash,
			ObfuscatedHash: obfuscatedHash,
			Address:        currAddr,
			Status:         Pending,
		}
		// Validate the data
		vErr, _ = dataMap.Validate(nil)
		values = append(values, fmt.Sprintf("(%s)", operation.GetNewInsertedValue(dataMap)))

		currHash = oyster_utils.HashHex(currHash, sha256.New())

		insertionCount++
		if insertionCount >= MaxNumberOfValueForInsertOperation {
			err = insertsIntoDataMapsTable(columnNames, strings.Join(values, oyster_utils.COLUMNS_SEPARATOR), len(values))
			insertionCount = 0
			values = nil
		}
	}
	if len(values) > 0 {
		err = insertsIntoDataMapsTable(columnNames, strings.Join(values, oyster_utils.COLUMNS_SEPARATOR), len(values))
	}
	return
}

/*@TODO is this file the best place for this method?*/
func CreateTreasurePayload(ethereumSeed string, sha256Hash string, maxSideChainLength int) (string, error) {
	keyLocation := rand.Intn(maxSideChainLength)

	currentHash := sha256Hash
	for i := 0; i <= keyLocation; i++ {
		currentHash = oyster_utils.HashHex(currentHash, sha3.New256())
	}

	encryptedResult := oyster_utils.Encrypt(currentHash, TreasurePrefix+ethereumSeed, sha256Hash)

	treasurePayload := string(oyster_utils.BytesToTrytes(encryptedResult)) + oyster_utils.RandSeq(TreasureChunkPadding, oyster_utils.TrytesAlphabet)

	return treasurePayload, nil
}

func GetUnassignedGenesisHashes() ([]interface{}, error) {

	var genesisHashesUnassigned = []DataMap{}

	err := DB.RawQuery("SELECT distinct genesis_hash FROM data_maps WHERE status = ? || status = ?",
		Unassigned,
		Error).All(&genesisHashesUnassigned)

	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return nil, err
	}

	genHashInterface := make([]interface{}, len(genesisHashesUnassigned))

	for i, genHash := range genesisHashesUnassigned {
		genHashInterface[i] = genHash.GenesisHash
	}

	// return value is an interface like this:
	// genHashes := []interface{}{"abcdeff1", "abcdeff2", "abcdeff3", "abcdeff4"}
	// need it in this form for "Where in {?)" queries
	return genHashInterface, nil
}

func GetUnassignedChunks() (dataMaps []DataMap, err error) {
	dataMaps = []DataMap{}
	err = DB.Where("status = ? OR status = ?", Unassigned, Error).All(&dataMaps)
	oyster_utils.LogIfError(err, nil)
	return dataMaps, err
}

func GetAllUnassignedChunksBySession(session UploadSession) (dataMaps []DataMap, err error) {
	dataMaps = []DataMap{}

	if session.Type == SessionTypeAlpha {
		err = DB.Where("genesis_hash = ? AND status = ? OR status = ? ORDER BY chunk_idx asc",
			session.GenesisHash, Unassigned, Error).All(&dataMaps)
	} else {
		err = DB.Where("genesis_hash = ? AND status = ? OR status = ? ORDER BY chunk_idx desc",
			session.GenesisHash, Unassigned, Error).All(&dataMaps)
	}
	oyster_utils.LogIfError(err, nil)
	return dataMaps, err
}

func GetUnassignedChunksBySession(session UploadSession, limit int) (dataMaps []DataMap, err error) {
	dataMaps = []DataMap{}

	if session.Type == SessionTypeAlpha {
		err = DB.Where("genesis_hash = ? AND status = ? OR status = ? ORDER BY chunk_idx asc LIMIT ?",
			session.GenesisHash, Unassigned, Error, limit).All(&dataMaps)
	} else {
		err = DB.Where("genesis_hash = ? AND status = ? OR status = ? ORDER BY chunk_idx desc LIMIT ?",
			session.GenesisHash, Unassigned, Error, limit).All(&dataMaps)
	}

	oyster_utils.LogIfError(err, nil)
	return dataMaps, err
}

func GetPendingChunksBySession(session UploadSession, limit int) (dataMaps []DataMap, err error) {
	dataMaps = []DataMap{}

	err = DB.Where("genesis_hash = ? AND status = ? LIMIT ?",
		session.GenesisHash, Pending, limit).All(&dataMaps)

	oyster_utils.LogIfError(err, nil)
	return dataMaps, err
}

func AttachUnassignedChunksToGenHashMap(genesisHashes []interface{}) (map[string]TypeAndChunkMap, error) {

	/* TODO: this method was an attempt at more sophisticated chunk prioritizing but the tests are being
	flaky.  After mainnet revisit this.
	*/

	if len(genesisHashes) > 0 {

		incompleteSessions := []UploadSession{}
		dataMaps := []DataMap{}

		err := DB.Where("genesis_hash in (?)", genesisHashes...).All(&incompleteSessions)

		if err != nil {
			oyster_utils.LogIfError(err, nil)
			return nil, err
		}
		if len(incompleteSessions) <= 0 {
			return nil, nil
		}

		hashAndTypeMap := map[string]TypeAndChunkMap{}
		for _, session := range incompleteSessions {
			if session.Type == SessionTypeAlpha {
				err = DB.RawQuery("SELECT * from data_maps where genesis_hash = ? AND status = ? OR status = ? ORDER BY chunk_idx asc",
					session.GenesisHash,
					Unassigned,
					Error).All(&dataMaps)
			} else {
				err = DB.RawQuery("SELECT * from data_maps where genesis_hash = ? AND status = ? OR status = ? ORDER BY chunk_idx desc",
					session.GenesisHash,
					Unassigned,
					Error).All(&dataMaps)
			}

			typeAndChunkMap := TypeAndChunkMap{
				Type:   session.Type,
				Chunks: dataMaps,
			}
			hashAndTypeMap[string(session.GenesisHash)] = typeAndChunkMap
		}

		return hashAndTypeMap, nil

	}
	return nil, nil
}

func insertsIntoDataMapsTable(columnsName string, values string, valueSize int) error {
	if len(values) == 0 {
		return nil
	}

	rawQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", DataMapTableName, columnsName, values)
	var err error
	for i := 0; i < oyster_utils.MAX_NUMBER_OF_SQL_RETRY; i++ {
		err = DB.RawQuery(rawQuery).All(&[]DataMap{})
		if err == nil {
			break
		}
	}
	oyster_utils.LogIfError(err, map[string]interface{}{"MaxRetry": oyster_utils.MAX_NUMBER_OF_SQL_RETRY, "NumOfRecord": valueSize})
	return err
}

// GetDataMapByGenesisHashAndChunkIdx lets you pass in genesis hash and chunk idx as
// parameters to get a specific data map
func GetDataMapByGenesisHashAndChunkIdx(genesisHash string, chunkIdx int) ([]DataMap, error) {
	dataMaps := []DataMap{}
	err := DB.Where("genesis_hash = ?",
		genesisHash).Where("chunk_idx = ?", chunkIdx).All(&dataMaps)

	oyster_utils.LogIfError(err, nil)
	return dataMaps, err
}

func MapChunkIndexesAndAddresses(chunks []DataMap) ([]string, []int) {

	addrs := make([]string, 0, len(chunks))
	indexes := make([]int, 0, len(chunks))

	for _, chunk := range chunks {
		addrs = append(addrs, chunk.Address)
		indexes = append(indexes, chunk.ChunkIdx)
	}

	return addrs, indexes
}

func GetDataMap(genHash string, numChunks int) (dataMaps []DataMap, vErr *validate.Errors) {

	fileChunksCount := numChunks

	if oyster_utils.BrokerMode != oyster_utils.TestModeNoTreasure {
		fileChunksCount = oyster_utils.GetTotalFileChunkIncludingBuriedPearlsUsingNumChunks(numChunks)
	}

	currHash := genHash
	for i := 0; i < fileChunksCount; i++ {
		obfuscatedHash := oyster_utils.HashHex(currHash, sha512.New384())
		currAddr := string(oyster_utils.MakeAddress(obfuscatedHash))

		dataMap := DataMap{
			GenesisHash:    genHash,
			ChunkIdx:       i,
			Hash:           currHash,
			ObfuscatedHash: obfuscatedHash,
			Address:        currAddr,
			Status:         Pending,
		}
		// Validate the data
		var err error
		vErr, err = dataMap.Validate(nil)
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			continue
		}
		dataMaps = append(dataMaps, dataMap)

	}
	return
}

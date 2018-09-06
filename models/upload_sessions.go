package models

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/crypto/sha3"
	"math"
	"math/big"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/pop/nulls"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/oysterprotocol/brokernode/utils"
	"github.com/shopspring/decimal"
)

/*ChunkReq is the form in which webinterface will send the data for each chunk*/
type ChunkReq struct {
	Idx  int    `json:"idx"`
	Data string `json:"data"`
	Hash string `json:"hash"` // This is GenesisHash.
}

type Invoice struct {
	Cost       decimal.Decimal `json:"cost"`
	EthAddress nulls.String    `json:"ethAddress"`
}

type TreasureMap struct {
	Sector int    `json:"sector"`
	Idx    int    `json:"idx"`
	Key    string `json:"key"`
}

type UploadSession struct {
	ID                   uuid.UUID `json:"id" db:"id"`
	CreatedAt            time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt            time.Time `json:"updatedAt" db:"updated_at"`
	GenesisHash          string    `json:"genesisHash" db:"genesis_hash"`
	NumChunks            int       `json:"numChunks" db:"num_chunks"`
	FileSizeBytes        uint64    `json:"fileSizeBytes" db:"file_size_bytes"` // In Trytes rather than Bytes
	StorageLengthInYears int       `json:"storageLengthInYears" db:"storage_length_in_years"`
	Type                 int       `json:"type" db:"type"`

	ETHAddrAlpha   nulls.String    `json:"ethAddrAlpha" db:"eth_addr_alpha"`
	ETHAddrBeta    nulls.String    `json:"ethAddrBeta" db:"eth_addr_beta"`
	ETHPrivateKey  string          `db:"eth_private_key"`
	TotalCost      decimal.Decimal `json:"totalCost" db:"total_cost"`
	PaymentStatus  int             `json:"paymentStatus" db:"payment_status"`
	TreasureStatus int             `json:"treasureStatus" db:"treasure_status"`

	TreasureIdxMap nulls.String `json:"treasureIdxMap" db:"treasure_idx_map"`
	Version        uint32       `json:"version" db:"version"`

	NextIdxToAttach int64 `json:"nextIdxToAttach" db:"next_idx_to_attach"`
	NextIdxToVerify int64 `json:"nextIdxToVerify" db:"next_idx_to_verify"`

	AllDataReady int `json:"allDataReady" db:"all_data_ready"`
}

const (
	AllDataNotReady int = iota + 1
	AllDataReady
)

// Enum for upload session type.
const (
	SessionTypeAlpha int = iota + 1
	SessionTypeBeta
)

const (
	PaymentStatusInvoiced int = iota + 1
	PaymentStatusPending
	PaymentStatusConfirmed
	PaymentStatusError = -1
)

const (
	TreasureGeneratingKeys int = iota + 1
	TreasureInDataMapPending
	TreasureInDataMapComplete
)

const (
	/*MaxTimesToCheckForAllChunks is the maximum number of times we will check for the chunks in WaitForAllChunks*/
	MaxTimesToCheckForAllChunks = 5000
	/*MaxBadgerInsertions is the maximum number of entries to add to badger at a time in
	BuildDataMaps*/
	MaxBadgerInsertions = 100
	/*FileBytesChunkSize is the total number of bytes in ascii that can fit in a chunk*/
	FileBytesChunkSize = float64(2187)
	/*MaxSideChainLength is the maximum length of the sidechain that we will create for the encrypting of treasure*/
	MaxSideChainLength = 1000
	/*DataMapsTimeToLive will cause data_maps
	message data to be garbage collected after two days.*/
	DataMapsTimeToLive = 2 * 24 * time.Hour
	/*CompletedDataMapsTimeToLive will cause completed_data_maps
	message data to be garbage collected after 3 weeks.*/
	CompletedDataMapsTimeToLive = 21 * 24 * time.Hour
)

var (
	/*TreasurePrefix appears at the beginning of every treasure payload*/
	TreasurePrefix = hex.EncodeToString([]byte("Treasure: "))
	/*TreasurePayloadLength - the length of the actual payload*/
	TreasurePayloadLength = len(TreasurePrefix) + 96
	/*TreasureChunkPadding - the length of padding to add after the payload*/
	TreasureChunkPadding = int(FileBytesChunkSize) - TreasurePayloadLength
	/*StoragePeg is how much storage in GB that one PRL will pay for, for a year.
	Long-term we will query the smart contract for this value*/
	StoragePeg = decimal.NewFromFloat(float64(64))
)

// String is not required by pop and may be deleted
func (u UploadSession) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// UploadSessions is not required by pop and may be deleted
type UploadSessions []UploadSession

// String is not required by pop and may be deleted
func (u UploadSessions) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

/**
 * Validations
 */

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (u *UploadSession) Validate(tx *pop.Connection) (*validate.Errors, error) {
	err := validate.Validate(
		&validators.StringIsPresent{Field: u.GenesisHash, Name: "GenesisHash"},
		&validators.IntIsPresent{Field: u.NumChunks, Name: "NumChunks"},
	)
	if u.FileSizeBytes == 0 {
		err.Add("FileSizeByte", "FileSizeByte is 0 and it should be a positive value.")
	}
	return err, nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (u *UploadSession) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (u *UploadSession) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

/**
 * Callbacks
 */

func (u *UploadSession) BeforeCreate(tx *pop.Connection) error {
	// Defaults to alpha session.
	if u.Type != SessionTypeBeta {
		u.Type = SessionTypeAlpha
	}

	if u.AllDataReady == 0 {
		u.AllDataReady = AllDataNotReady
	}

	switch oyster_utils.BrokerMode {
	case oyster_utils.ProdMode:
		// Defaults to paymentStatusPending
		if u.PaymentStatus == 0 {
			if oyster_utils.PaymentMode == oyster_utils.UserIsPaying {
				u.PaymentStatus = PaymentStatusInvoiced
			} else {
				u.PaymentStatus = PaymentStatusConfirmed
			}
		}

		// Defaults to treasureGeneratingKeys
		if u.TreasureStatus == 0 {
			u.TreasureStatus = TreasureGeneratingKeys
		}
	case oyster_utils.TestModeDummyTreasure:
		// Defaults to paymentStatusPaid
		if u.PaymentStatus == 0 {
			u.PaymentStatus = PaymentStatusConfirmed
		}

		// Defaults to treasureBurying
		if u.TreasureStatus == 0 {
			u.TreasureStatus = TreasureGeneratingKeys
		}
	case oyster_utils.TestModeNoTreasure:
		// Defaults to paymentStatusPaid
		if u.PaymentStatus == 0 {
			u.PaymentStatus = PaymentStatusConfirmed
		}

		// Defaults to treasureBuried
		if u.TreasureStatus == 0 {
			u.TreasureStatus = TreasureInDataMapComplete
		}
	}

	return nil
}

/**
 * Methods
 */

// BuildDataMapsForSession builds the datamap and inserts them into the badger DB.
func BuildDataMapsForSession(genHash string, numChunks int) (err error) {

	if oyster_utils.DataMapStorageMode == oyster_utils.DataMapsInSQL {
		vErr, err := BuildDataMapsForSessionInSQL(genHash, numChunks)
		if err != nil {
			return err
		}
		if vErr.HasAny() {
			return errors.New(vErr.Error())
		}
	} else {
		fileChunksCount := numChunks

		currHash := genHash
		insertionCount := 0

		dbID := []string{oyster_utils.InProgressDir, genHash, oyster_utils.HashDir}

		db := oyster_utils.GetOrInitUniqueBadgerDB(dbID)
		if db == nil {
			err := errors.New("error creating unique badger DB")
			oyster_utils.LogIfError(err, nil)
			return err
		}

		kvPairs := oyster_utils.KVPairs{}

		for i := 0; i < fileChunksCount; i++ {

			kvPairs[oyster_utils.GetBadgerKey([]string{genHash, strconv.Itoa(i)})] = currHash
			currHash = oyster_utils.HashHex(currHash, sha256.New())

			insertionCount++
			if insertionCount >= MaxBadgerInsertions {
				err = oyster_utils.BatchSetToUniqueDB(dbID, &kvPairs, DataMapsTimeToLive)
				if err == nil {
					insertionCount = 0
					kvPairs = oyster_utils.KVPairs{}
				}
				oyster_utils.LogIfError(err, nil)
			}
		}

		if len(kvPairs) > 0 {
			err = oyster_utils.BatchSetToUniqueDB(dbID, &kvPairs, DataMapsTimeToLive)
			oyster_utils.LogIfError(err, nil)
			if err != nil {
				panic(err)
			}
		}
	}

	return err
}

// BuildDataMapsForSessionInSQL builds the datamap and inserts them into the sql DB.
func BuildDataMapsForSessionInSQL(genHash string, numChunks int) (vErr *validate.Errors, err error) {

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
			MsgID:          oyster_utils.GetBadgerKey([]string{genHash, strconv.Itoa(i)}),
		}
		// We use INSERT SQL query rather than default Create method.
		dataMap.BeforeCreate(nil)

		// Validate the data
		vErr, _ = dataMap.Validate(nil)
		oyster_utils.LogIfValidationError(
			"validation errors for creating dataMap for batch insertion.", vErr, nil)
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

// StartUploadSession will generate dataMaps and save the session and dataMaps
// to the DB.
func (u *UploadSession) StartUploadSession() (vErr *validate.Errors, err error) {
	u.calculatePayment()

	vErr, err = DB.ValidateAndCreate(u)
	if err != nil || len(vErr.Errors) > 0 {
		oyster_utils.LogIfError(err, nil)
		oyster_utils.LogIfValidationError("validation error for creating UploadSession.", vErr, nil)
		return
	}

	u.EncryptSessionEthKey()

	if oyster_utils.BrokerMode != oyster_utils.TestModeNoTreasure {
		u.NumChunks = oyster_utils.GetTotalFileChunkIncludingBuriedPearlsUsingNumChunks(u.NumChunks)
		DB.ValidateAndUpdate(u)
	}

	switch u.Type {
	case SessionTypeAlpha:
		u.NextIdxToAttach = 0
		u.NextIdxToVerify = 0
	case SessionTypeBeta:
		u.NextIdxToAttach = int64(u.NumChunks - 1)
		u.NextIdxToVerify = int64(u.NumChunks - 1)
	}
	DB.ValidateAndUpdate(u)

	go func() {
		err = BuildDataMapsForSession(u.GenesisHash, u.NumChunks)
		oyster_utils.LogIfError(err, nil)
	}()

	return
}

/*StartSessionAndWaitForChunks is a substitute for StartUploadSession intended to be used in unit tests*/
func (u *UploadSession) StartSessionAndWaitForChunks(maxTimesToCheckForAllChunks int) (bool, *validate.Errors, error) {
	vErr, err := u.StartUploadSession()
	if vErr.HasAny() || err != nil {
		oyster_utils.LogIfError(err, nil)
		oyster_utils.LogIfValidationError("vErr during StartUploadSession", vErr, nil)
		return false, vErr, err
	}
	chunksAreFinished, err := u.WaitForAllChunks(maxTimesToCheckForAllChunks)
	oyster_utils.LogIfError(err, nil)
	return chunksAreFinished, vErr, err
}

func (u *UploadSession) GetInvoice() Invoice {

	var ethAddress nulls.String

	if u.Type != SessionTypeAlpha {
		ethAddress = u.ETHAddrBeta
	} else {
		ethAddress = u.ETHAddrAlpha
	}

	return Invoice{
		EthAddress: ethAddress,
		Cost:       u.TotalCost,
	}
}

func (u *UploadSession) calculatePayment() {

	// convert all variables to decimal format
	storagePeg := GetStoragePeg()
	fileSizeInBytes := decimal.NewFromFloat(float64(u.FileSizeBytes))
	storageLength := decimal.NewFromFloat(float64(u.StorageLengthInYears))

	// calculate total cost
	fileSizeInKB := fileSizeInBytes.Div(decimal.NewFromFloat(float64(oyster_utils.FileChunkSizeInByte)))
	numChunks := fileSizeInKB.Add(decimal.NewFromFloat(float64(1))).Ceil()
	numSectors := numChunks.Div(decimal.NewFromFloat(float64(oyster_utils.FileSectorInChunkSize))).Ceil()
	costPerYear := numSectors.Div(storagePeg)
	u.TotalCost = costPerYear.Mul(storageLength)
}

func (u *UploadSession) GetTreasureMap() ([]TreasureMap, error) {
	var err error
	treasureIndex := []TreasureMap{}
	if u.TreasureIdxMap.Valid {
		// only do this if the string value is valid
		err = json.Unmarshal([]byte(u.TreasureIdxMap.String), &treasureIndex)
		oyster_utils.LogIfError(err, nil)
	}

	return treasureIndex, err
}

func (u *UploadSession) SetTreasureMap(treasureIndexMap []TreasureMap) error {
	treasureString, err := json.Marshal(treasureIndexMap)
	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return err
	}
	u.TreasureIdxMap = nulls.String{string(treasureString), true}
	_, err = DB.ValidateAndSave(u)
	oyster_utils.LogIfError(err, nil)
	return err
}

// Sets the TreasureIdxMap with Sector, Idx, and Key
func (u *UploadSession) MakeTreasureIdxMap(mergedIndexes []int, privateKeys []string) {

	treasureIndexArray := make([]TreasureMap, 0)

	for i, mergedIndex := range mergedIndexes {

		key, err := u.EncryptTreasureChunkEthKey(privateKeys[i])
		if err != nil {
			oyster_utils.LogIfError(errors.New(err.Error()+" in MakeTreasureIdxMap"),
				nil)
			return
		}

		treasureIndexArray = append(treasureIndexArray, TreasureMap{
			Sector: i,
			Idx:    mergedIndex,
			Key:    key,
		})
	}

	treasureString, err := json.Marshal(treasureIndexArray)
	if err != nil {
		oyster_utils.LogIfError(err, nil)
	}

	u.TreasureIdxMap = nulls.String{string(treasureString), true}
	u.TreasureStatus = TreasureInDataMapPending

	DB.ValidateAndSave(u)
}

func (u *UploadSession) GetTreasureIndexes() ([]int, error) {
	treasureMap, err := u.GetTreasureMap()
	if err != nil {
		oyster_utils.LogIfError(err, nil)
	}
	treasureIndexArray := make([]int, 0)
	for _, treasure := range treasureMap {
		treasureIndexArray = append(treasureIndexArray, treasure.Idx)
	}
	return treasureIndexArray, err
}

func (u *UploadSession) GetPRLsPerTreasure() (*big.Float, error) {

	indexes, err := u.GetTreasureIndexes()
	if err != nil {
		fmt.Println("Cannot get prls per treasures in models/upload_sessions.go" + err.Error())
		// captured error in upstream method
		return big.NewFloat(0), err
	}

	prlTotal := u.TotalCost.Rat()
	numerator := prlTotal.Num()
	denominator := prlTotal.Denom()

	prlTotalFloat := new(big.Float).Quo(new(big.Float).SetInt(numerator), new(big.Float).SetInt(denominator))
	prlTotalToBuryFloat := new(big.Float).Quo(prlTotalFloat, new(big.Float).SetInt(big.NewInt(int64(2))))

	totalChunks := oyster_utils.GetTotalFileChunkIncludingBuriedPearlsUsingNumChunks(u.NumChunks)
	totalSectors := float64(math.Ceil(float64(totalChunks) / float64(oyster_utils.FileSectorInChunkSize)))

	if int(totalSectors) != len(indexes) {
		err = errors.New("length of treasure indexes does not match totalSectors in models/upload_sessions.go")
		oyster_utils.LogIfError(err, nil)
		return big.NewFloat(0), err
	}

	prlPerSector := new(big.Float).Quo(prlTotalToBuryFloat, big.NewFloat(totalSectors))

	return prlPerSector, nil
}

func GetStoragePeg() decimal.Decimal {
	return StoragePeg // TODO: write code to query smart contract to get real storage peg
}

func (u *UploadSession) GetPaymentStatus() string {
	switch u.PaymentStatus {
	case PaymentStatusInvoiced:
		return "invoiced"
	case PaymentStatusPending:
		return "pending"
	case PaymentStatusConfirmed:
		return "confirmed"
	default:
		return "error"
	}
}

func (u *UploadSession) EncryptSessionEthKey() {
	u.ETHPrivateKey = oyster_utils.ReturnEncryptedEthKey(u.ID, u.CreatedAt, u.ETHPrivateKey)
	DB.ValidateAndSave(u)
}

func (u *UploadSession) DecryptSessionEthKey() string {
	return oyster_utils.ReturnDecryptedEthKey(u.ID, u.CreatedAt, u.ETHPrivateKey)
}

/*EncryptTreasureChunkEthKey encrypts the eth key of the treasure chunk*/
func (u *UploadSession) EncryptTreasureChunkEthKey(unencryptedKey string) (string, error) {
	encryptedKey := oyster_utils.ReturnEncryptedEthKey(u.ID, u.CreatedAt, unencryptedKey)
	return encryptedKey, nil
}

/*DecryptTreasureChunkEthKey decrypts the eth key of the treasure chunk*/
func (u *UploadSession) DecryptTreasureChunkEthKey(encryptedKey string) (string, error) {
	return oyster_utils.ReturnDecryptedEthKey(u.ID, u.CreatedAt, encryptedKey), nil
}

/*WaitForAllChunks is a blocking call that will wait for all chunks or false and an error if we get an
error, or if we have checked too many times.  This is mainly intended for use in unit tests.*/
func (u *UploadSession) WaitForAllChunks(maxTimesToCheckForAllChunks int) (bool, error) {

	hashesFinished, errHashes := u.WaitForAllHashes(MaxTimesToCheckForAllChunks)
	messagesFinished, errMessages := u.WaitForAllMessages(MaxTimesToCheckForAllChunks)

	if errHashes != nil {
		return false, errHashes
	}
	if errMessages != nil {
		return false, errMessages
	}

	return hashesFinished && messagesFinished, nil
}

/*WaitForAllHashes is a blocking call that will wait for all hashes of a file or return false and an
error if we get an error, or if we have checked too many times.  This is mainly intended for use in unit tests.*/
func (u *UploadSession) WaitForAllHashes(maxTimesToCheckForAllChunks int) (bool, error) {

	timesChecked := 0
	for {
		ready := u.CheckIfAllHashesAreReady()
		timesChecked++

		if ready {
			break
		} else {
			time.Sleep(2000 * time.Millisecond)
		}

		if timesChecked >= maxTimesToCheckForAllChunks {
			return false, errors.New("checked for the chunks too many times")
		}
	}
	return true, nil
}

/*WaitForAllMessages is a blocking call that will wait for all messages of a file or return false and an
error if we get an error, or if we have checked too many times.  This is mainly intended for use in unit tests.*/
func (u *UploadSession) WaitForAllMessages(maxTimesToCheckForAllChunks int) (bool, error) {

	timesChecked := 0
	for {
		ready := u.CheckIfAllMessagesAreReady()
		timesChecked++

		if ready {
			break
		} else {
			time.Sleep(2000 * time.Millisecond)
		}

		if timesChecked >= maxTimesToCheckForAllChunks {
			return false, errors.New("checked for the chunks too many times")
		}
	}
	return true, nil
}

/*CheckIfAllDataIsReady wraps calls which will check if the message data and hash data have both been created*/
func (u *UploadSession) CheckIfAllDataIsReady() bool {

	return u.AllDataReady == AllDataReady || (u.CheckIfAllHashesAreReady() && u.CheckIfAllMessagesAreReady())
}

/*CheckIfAllHashesAreReady verifies that all the hashes for the file chunks have been created.  It returns true if
the first and last chunk hashes are present*/
func (u *UploadSession) CheckIfAllHashesAreReady() bool {

	if oyster_utils.DataMapStorageMode == oyster_utils.DataMapsInBadger {
		chunkDataStartHash := oyster_utils.GetHashData(oyster_utils.InProgressDir, u.GenesisHash, 0)
		chunkDataEndHash := oyster_utils.GetHashData(oyster_utils.InProgressDir, u.GenesisHash, int64(u.NumChunks-1))

		return chunkDataStartHash != "" && chunkDataEndHash != ""
	}

	startIdx := 0
	lastIdx := u.NumChunks - 1

	dataMapStart := []DataMap{}
	dataMapEnd := []DataMap{}

	DB.Where("genesis_hash = ? AND chunk_idx = ?", u.GenesisHash, int(startIdx)).All(&dataMapStart)
	DB.Where("genesis_hash = ? AND chunk_idx = ?", u.GenesisHash, int(lastIdx)).All(&dataMapEnd)

	return len(dataMapStart) == 1 && len(dataMapEnd) == 1
}

/*CheckIfAllMessagesAreReady verifies that all the messages for the file chunks have been created.  It returns true if
the first, last, and treasure chunk messages are present*/
func (u *UploadSession) CheckIfAllMessagesAreReady() bool {

	treasureIndexes, err := u.GetTreasureIndexes()
	if err != nil || (len(treasureIndexes) == 0 && oyster_utils.BrokerMode != oyster_utils.TestModeNoTreasure) {
		oyster_utils.LogIfError(err, nil)
		return false
	}

	if oyster_utils.DataMapStorageMode == oyster_utils.DataMapsInBadger {
		return checkIfAllMessagesAreReadyInBadger(treasureIndexes, u)
	}
	return checkIfAllMessagesAreReadyInSQL(treasureIndexes, u)
}

func checkIfAllMessagesAreReadyInBadger(treasureIndexes []int, u *UploadSession) bool {

	chunkDataStartMessage := oyster_utils.GetMessageData(oyster_utils.InProgressDir, u.GenesisHash, 0)
	chunkDataEndMessage := oyster_utils.GetMessageData(oyster_utils.InProgressDir, u.GenesisHash, int64(u.NumChunks-1))

	allMessagesFound := false
	if chunkDataStartMessage != "" && chunkDataEndMessage != "" {
		allMessagesFound = true
		if len(treasureIndexes) > 0 {
			for _, index := range treasureIndexes {
				chunkDataTreasureMessage :=
					oyster_utils.GetMessageData(oyster_utils.InProgressDir, u.GenesisHash, int64(index))
				if chunkDataTreasureMessage == "" {
					allMessagesFound = false
				}
			}
		}
	}
	return allMessagesFound
}

func checkIfAllMessagesAreReadyInSQL(treasureIndexes []int, u *UploadSession) bool {
	chunkDataStart := GetSingleChunkData(oyster_utils.InProgressDir, u.GenesisHash, 0)
	chunkDataEnd := GetSingleChunkData(oyster_utils.InProgressDir, u.GenesisHash, int64(u.NumChunks-1))

	allMessagesFound := false
	if chunkDataStart.Message != "" && chunkDataEnd.Message != "" {
		allMessagesFound = true
		if len(treasureIndexes) > 0 {
			for _, index := range treasureIndexes {
				chunkDataTreasure :=
					GetSingleChunkData(oyster_utils.InProgressDir, u.GenesisHash, int64(index))
				if chunkDataTreasure.Message == "" {
					allMessagesFound = false
				}
			}
		}
	}
	return allMessagesFound
}

/*GetUnassignedChunksBySession returns the chunk data for chunks that need attaching for a particular session*/
func (u *UploadSession) GetUnassignedChunksBySession(offset int) (chunkData []oyster_utils.ChunkData, err error) {
	var stopChunkIdx int64

	if u.Type == SessionTypeAlpha {
		stopChunkIdx = u.NextIdxToAttach + int64(offset)
		if stopChunkIdx > int64(u.NumChunks-1) {
			stopChunkIdx = int64(u.NumChunks - 1)
		}
	} else {
		stopChunkIdx = u.NextIdxToAttach - int64(offset)
		if stopChunkIdx < 0 {
			stopChunkIdx = int64(0)
		}
	}

	keys := oyster_utils.GenerateBulkKeys(u.GenesisHash, u.NextIdxToAttach, stopChunkIdx)

	chunkData, err = GetMultiChunkData(oyster_utils.InProgressDir, u.GenesisHash, keys)
	oyster_utils.LogIfError(err, nil)

	return chunkData, err
}

/*MoveChunksToCompleted receives some chunks for a session and moves them to a separate DB for completed chunks*/
func (u *UploadSession) MoveChunksToCompleted(chunks []oyster_utils.ChunkData) {

	if oyster_utils.DataMapStorageMode == oyster_utils.DataMapsInBadger {

		for i := range chunks {
			inProgressMessageDBID := []string{oyster_utils.InProgressDir, u.GenesisHash, oyster_utils.MessageDir}
			inProgressHashDBID := []string{oyster_utils.InProgressDir, u.GenesisHash, oyster_utils.HashDir}

			completeMessageDBID := []string{oyster_utils.CompletedDir, u.GenesisHash, oyster_utils.MessageDir}
			completeHashDBID := []string{oyster_utils.CompletedDir, u.GenesisHash, oyster_utils.HashDir}

			key := oyster_utils.GetBadgerKey([]string{u.GenesisHash, strconv.FormatInt(chunks[i].Idx, 10)})

			kvMessage := oyster_utils.KVPairs{key: chunks[i].RawMessage}
			kvHash := oyster_utils.KVPairs{key: chunks[i].Hash}

			errMessage := oyster_utils.BatchSetToUniqueDB(completeMessageDBID, &kvMessage,
				CompletedDataMapsTimeToLive)
			oyster_utils.LogIfError(errMessage, nil)
			errHash := oyster_utils.BatchSetToUniqueDB(completeHashDBID, &kvHash,
				CompletedDataMapsTimeToLive)
			oyster_utils.LogIfError(errHash, nil)

			if errMessage == nil && errHash == nil {
				oyster_utils.BatchDeleteFromUniqueDB(inProgressMessageDBID, &oyster_utils.KVKeys{key})
				oyster_utils.BatchDeleteFromUniqueDB(inProgressHashDBID, &oyster_utils.KVKeys{key})
			} else if errMessage != nil {
				oyster_utils.LogIfError(errors.New(errMessage.Error()+" while saving message to completed db "+
					"in MoveChunksToCompleted in models/upload_sessions"), nil)
			} else if errHash != nil {
				oyster_utils.LogIfError(errors.New(errHash.Error()+" while saving hash to completed db "+
					"in MoveChunksToCompleted in models/upload_sessions"), nil)
			}
		}
	} else {
		moveToComplete(chunks)
	}
}

/*MoveAllChunksToCompleted moves all the chunks for an in-progress session to completed.*/
func (u *UploadSession) MoveAllChunksToCompleted() error {
	if oyster_utils.DataMapStorageMode == oyster_utils.DataMapsInBadger {
		return moveAllChunksToCompletedBadger(u)
	}
	return moveAllChunksToCompletedSQL(u)
}

func moveAllChunksToCompletedBadger(u *UploadSession) error {
	inProgressMessageDBID := []string{oyster_utils.InProgressDir, u.GenesisHash, oyster_utils.MessageDir}
	inProgressHashDBID := []string{oyster_utils.InProgressDir, u.GenesisHash, oyster_utils.HashDir}

	completeMessageDBID := []string{oyster_utils.CompletedDir, u.GenesisHash, oyster_utils.MessageDir}
	completeHashDBID := []string{oyster_utils.CompletedDir, u.GenesisHash, oyster_utils.HashDir}

	for i := 0; i <= u.NumChunks-1; {

		stop := int64(i + MaxBadgerInsertions)
		if stop > int64(u.NumChunks-1) {
			stop = int64(u.NumChunks - 1)
		}

		keys := oyster_utils.GenerateBulkKeys(u.GenesisHash, int64(i), stop)

		kvMessages, err := oyster_utils.BatchGetFromUniqueDB(inProgressMessageDBID, keys)
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			return err
		}
		kvHashes, err := oyster_utils.BatchGetFromUniqueDB(inProgressHashDBID, keys)
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			return err
		}

		errMessage := oyster_utils.BatchSetToUniqueDB(completeMessageDBID, kvMessages, CompletedDataMapsTimeToLive)
		if errMessage != nil {
			oyster_utils.LogIfError(errMessage, nil)
			return errMessage
		}
		errHash := oyster_utils.BatchSetToUniqueDB(completeHashDBID, kvHashes, CompletedDataMapsTimeToLive)
		if errHash != nil {
			oyster_utils.LogIfError(errHash, nil)
			return errHash
		}

		oyster_utils.BatchDeleteFromUniqueDB(inProgressMessageDBID, keys)
		oyster_utils.BatchDeleteFromUniqueDB(inProgressHashDBID, keys)

		if i == u.NumChunks-1 {
			break
		}

		i = i + MaxBadgerInsertions

		if i > u.NumChunks-1 {
			i = u.NumChunks - 1
		}
	}

	return nil
}

func moveAllChunksToCompletedSQL(u *UploadSession) error {
	for ok, i := true, 0; ok; ok = i < u.NumChunks {
		end := i + MaxBadgerInsertions

		if end > u.NumChunks {
			end = u.NumChunks
		}

		if i >= end {
			break
		}

		bulkKeys := oyster_utils.GenerateBulkKeys(u.GenesisHash, int64(i), int64(end))

		chunks, err := GetMultiChunkData(oyster_utils.InProgressDir, u.GenesisHash, bulkKeys)
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			return err
		}

		if len(chunks[i:end]) > 0 {
			err := moveToComplete(chunks[i:end])
			if err != nil {
				oyster_utils.LogIfError(err, nil)
				return err
			}
		}
		i += MaxBadgerInsertions
	}
	err := DB.RawQuery("DELETE FROM data_maps WHERE genesis_hash = ?", u.GenesisHash).All(&[]DataMap{})
	return err
}

func moveToComplete(dataMaps []oyster_utils.ChunkData) error {
	if len(dataMaps) == 0 {
		return nil
	}

	existedDataMaps := []CompletedDataMap{}
	DB.RawQuery("SELECT address FROM completed_data_maps WHERE genesis_hash = ?", dataMaps[0].GenesisHash).All(&existedDataMaps)
	existedMap := make(map[string]bool)
	for _, dm := range existedDataMaps {
		existedMap[dm.Address] = true
	}

	messagsKvPairs := oyster_utils.KVPairs{}
	var upsertedValues []string
	dbOperation, _ := oyster_utils.CreateDbUpdateOperation(&CompletedDataMap{})
	hasValidationError := false

	for _, dataMap := range dataMaps {
		if _, hasKey := existedMap[dataMap.Address]; hasKey {
			continue
		}

		completedDataMap := CompletedDataMap{
			GenesisHash: dataMap.GenesisHash,
			ChunkIdx:    int(dataMap.Idx),
			Hash:        dataMap.Hash,
			Address:     dataMap.Address,
			MsgStatus:   MsgStatusUploadedHaveNotEncoded,
			MsgID:       oyster_utils.GenerateBadgerKey(CompletedDataMapsMsgIDPrefix, dataMap.GenesisHash, int(dataMap.Idx)),
		}

		if vErr, err := completedDataMap.Validate(nil); err != nil || vErr.HasAny() {
			oyster_utils.LogIfValidationError("CompletedDataMap validation failed", vErr, nil)
			oyster_utils.LogIfError(err, nil)
			hasValidationError = true
			return errors.New("completedDataMap validation error in moveToComplete")
		}

		messagsKvPairs[completedDataMap.MsgID] = dataMap.RawMessage
		upsertedValues = append(upsertedValues, dbOperation.GetNewInsertedValue(completedDataMap))
	}

	errBatchUpsert := BatchUpsert("completed_data_maps", upsertedValues, dbOperation.GetColumns(), nil)
	if errBatchUpsert != nil {
		return errors.New("BatchUpsert failed")
	}

	errBatchSet := oyster_utils.BatchSet(&messagsKvPairs, CompletedDataMapsTimeToLive)
	if errBatchSet != nil {
		return errors.New("BatchSet failed")
	}

	if hasValidationError {
		return errors.New("Partial update failed")
	}

	for _, dataMap := range dataMaps {
		err := DB.RawQuery("DELETE FROM data_maps WHERE genesis_hash = ? and chunk_idx = ?",
			dataMap.GenesisHash, int(dataMap.Idx)).All(&[]UploadSessions{})
		oyster_utils.LogIfError(err, nil)
	}
	return nil
}

/*UpdateIndexWithVerifiedChunks receives some chunks and will update the session's NextIdxToVerify property.
Starting at its current NextIdxToVerify, it checks the next index and verifies that there is a chunk matching that
index in the array passed in.  When it finds a missing index it sets the NextIdxToVerify to that index.*/
func (u *UploadSession) UpdateIndexWithVerifiedChunks(chunks []oyster_utils.ChunkData) {

	treasureIndexes, _ := u.GetTreasureIndexes()

	nextIdxToVerifyStart := u.NextIdxToVerify
	nextIdxToVerifyNew := u.NextIdxToVerify

	idxMap := make(map[int64]bool)
	treasureIdxMap := make(map[int64]bool)

	for i := range chunks {
		idxMap[chunks[i].Idx] = true
	}
	for _, index := range treasureIndexes {
		treasureIdxMap[int64(index)] = true
	}

	step := int64(1)

	if u.Type == SessionTypeBeta {
		step = int64(-1)
	}

	for i := nextIdxToVerifyStart; i != nextIdxToVerifyStart+(int64(len(chunks))*step); i = i + step {
		if _, ok := treasureIdxMap[i]; !ok {
			if _, ok := idxMap[i]; !ok {
				break
			}
		} else {
			chunk := GetSingleChunkData(oyster_utils.CompletedDir, u.GenesisHash, int64(i))
			if chunk.Hash == "" {
				break
			}
		}
		nextIdxToVerifyNew = i + step
	}

	if u.Type == SessionTypeAlpha {
		u.NextIdxToVerify = int64(math.Min(float64(nextIdxToVerifyNew), float64(u.NextIdxToAttach)))
	} else {
		u.NextIdxToVerify = int64(math.Max(float64(nextIdxToVerifyNew), float64(u.NextIdxToAttach)))
	}

	vErr, err := DB.ValidateAndUpdate(u)
	if err != nil {
		oyster_utils.LogIfError(errors.New(err.Error()+" updating NextIdxToVerify in "+
			"UpdateIndexWithVerifiedChunks in models/upload_sessions.go"), nil)
	}
	if vErr.HasAny() {
		oyster_utils.LogIfValidationError(" updating NextIdxToVerify in "+
			"UpdateIndexWithVerifiedChunks in models/upload_sessions.go", vErr, nil)
	}
}

/*UpdateIndexWithAttachedChunks receives some chunks and will update the session's NextIdxToAttach property.
Starting at its current NextIdxToAttach, it checks the next index and verifies that there is a chunk matching that
index in the array passed in.  When it finds a missing index it sets the NextIdxToAttach to that index.*/
func (u *UploadSession) UpdateIndexWithAttachedChunks(chunks []oyster_utils.ChunkData) {

	treasureIndexes, _ := u.GetTreasureIndexes()

	nextIdxToAttachStart := u.NextIdxToAttach
	nextIdxToAttachNew := u.NextIdxToAttach

	idxMap := make(map[int64]bool)
	treasureIdxMap := make(map[int64]bool)

	for i := range chunks {
		idxMap[chunks[i].Idx] = true
	}
	for _, index := range treasureIndexes {
		treasureIdxMap[int64(index)] = true
	}

	step := int64(1)

	if u.Type == SessionTypeBeta {
		step = int64(-1)
	}

	for i := nextIdxToAttachStart; i != nextIdxToAttachStart+(int64(len(chunks))*step); i = i + step {
		if _, ok := treasureIdxMap[i]; !ok {
			if _, ok := idxMap[i]; !ok {
				break
			}
		} else {
			chunk := GetSingleChunkData(oyster_utils.CompletedDir, u.GenesisHash, int64(i))
			if chunk.Hash == "" {
				break
			}
		}
		nextIdxToAttachNew = i + step
	}

	u.NextIdxToAttach = nextIdxToAttachNew

	vErr, err := DB.ValidateAndUpdate(u)
	if err != nil {
		oyster_utils.LogIfError(errors.New(err.Error()+" updating NextIdxToAttach in "+
			"UpdateIndexWithAttachedChunks in models/upload_sessions.go"), nil)
	}
	if vErr.HasAny() {
		oyster_utils.LogIfValidationError(" updating NextIdxToAttach in "+
			"UpdateIndexWithAttachedChunks in models/upload_sessions.go", vErr, nil)
	}
}

/*DownGradeIndexesOnUnattachedChunks receives chunks that are not attached or that were attached incorrectly.
If it finds a chunk that is unattached that is below (if alpha) or above (if beta) its current indexes it will
change its indexes to match these unattached chunks.*/
func (u *UploadSession) DownGradeIndexesOnUnattachedChunks(chunks []oyster_utils.ChunkData) {
	if len(chunks) == 0 {
		return
	}

	maxIdx := int64(chunks[0].Idx)
	minIdx := int64(chunks[0].Idx)

	var nextIdxToAttach int64
	var nextIdxToVerify int64

	for _, chunk := range chunks {
		maxIdx = int64(math.Max(float64(chunk.Idx), float64(maxIdx)))
		minIdx = int64(math.Min(float64(chunk.Idx), float64(minIdx)))
	}

	if u.Type == SessionTypeAlpha {
		nextIdxToAttach = int64(math.Min(float64(minIdx), float64(u.NextIdxToAttach)))
		nextIdxToVerify = int64(math.Min(float64(minIdx), float64(u.NextIdxToVerify)))
	} else {
		nextIdxToAttach = int64(math.Max(float64(maxIdx), float64(u.NextIdxToAttach)))
		nextIdxToVerify = int64(math.Max(float64(maxIdx), float64(u.NextIdxToVerify)))
	}

	u.NextIdxToVerify = nextIdxToVerify
	u.NextIdxToAttach = nextIdxToAttach

	vErr, err := DB.ValidateAndUpdate(u)
	oyster_utils.LogIfValidationError("updating session in verify_data_maps", vErr, nil)
	oyster_utils.LogIfError(err, nil)
}

/*SetTreasureMessage sets the treasure message for a treasure chunk of a session.*/
func (u *UploadSession) SetTreasureMessage(treasureIndex int, treasurePayload string,
	ttl time.Duration) (err error) {

	if oyster_utils.DataMapStorageMode == oyster_utils.DataMapsInBadger {
		key := oyster_utils.GetBadgerKey([]string{u.GenesisHash, strconv.Itoa(treasureIndex)})
		err = oyster_utils.BatchSetToUniqueDB([]string{oyster_utils.InProgressDir, u.GenesisHash,
			oyster_utils.MessageDir}, &oyster_utils.KVPairs{key: treasurePayload}, ttl)
	} else {
		key := oyster_utils.GenerateBadgerKey("", u.GenesisHash, treasureIndex)
		err = oyster_utils.BatchSet(&oyster_utils.KVPairs{key: treasurePayload}, ttl)
	}
	return err
}

/*GetSessionsWithIncompleteData gets all the sessions for which we don't have all the data.*/
func GetSessionsWithIncompleteData() ([]UploadSession, error) {
	sessions := []UploadSession{}

	err := DB.RawQuery("SELECT * FROM upload_sessions WHERE all_data_ready = ?",
		AllDataNotReady).All(&sessions)

	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return nil, err
	}

	return sessions, nil
}

/*GetSessionsByAge gets all the sessions eligible for chunk attachment.*/
func GetSessionsByAge() ([]UploadSession, error) {
	sessionsByAge := []UploadSession{}

	err := DB.RawQuery("SELECT * FROM upload_sessions WHERE payment_status = ? AND "+
		"treasure_status = ? AND all_data_ready = ? ORDER BY created_at ASC",
		PaymentStatusConfirmed, TreasureInDataMapComplete, AllDataReady).All(&sessionsByAge)

	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return []UploadSession{}, err
	}

	return sessionsByAge, nil
}

/*GetCompletedSessions gets all the sessions whose index values suggest that they are completed.*/
func GetCompletedSessions() ([]UploadSession, error) {
	completedSessions := []UploadSession{}
	sessions, err := GetSessionsByAge()

	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return nil, err
	}

	for _, session := range sessions {
		stop := session.NumChunks
		if session.Type == SessionTypeBeta {
			stop = -1
		}
		if session.NextIdxToVerify == int64(stop) {
			completedSessions = append(completedSessions, session)
		}
	}

	return completedSessions, nil
}

/*GetSessionsThatNeedTreasure checks for sessions which the user has paid their PRL but in which
we have not yet buried the treasure.*/
func GetSessionsThatNeedTreasure() ([]UploadSession, error) {
	unburiedSessions := []UploadSession{}

	err := DB.Where("payment_status = ? AND treasure_status = ?",
		PaymentStatusConfirmed, TreasureInDataMapPending).All(&unburiedSessions)
	oyster_utils.LogIfError(err, nil)

	return unburiedSessions, err
}

/*GetReadySessions gets all the sessions eligible for chunk attachment which still need chunks to be attached.*/
func GetReadySessions() ([]UploadSession, error) {
	sessions := []UploadSession{}
	readySessions := []UploadSession{}

	sessions, err := GetSessionsByAge()

	for _, session := range sessions {
		if session.Type == SessionTypeBeta && session.NextIdxToAttach != -1 {
			readySessions = append(readySessions, session)
		} else if session.Type == SessionTypeAlpha && session.NextIdxToAttach < int64(session.NumChunks) {
			readySessions = append(readySessions, session)
		}
	}

	return readySessions, err
}

/* SetBrokerTransactionToPaid will find the the upload session's corresponding broker_broker_transaction and set it
to paid.  This will happen if the polling in actions/upload_sessions.go detects payment before the job in
check_alpha_payments.go
*/
func SetBrokerTransactionToPaid(session UploadSession) error {
	err := DB.RawQuery("UPDATE broker_broker_transactions set payment_status = ? WHERE "+
		"payment_status != ? AND genesis_hash = ?",
		BrokerTxAlphaPaymentConfirmed,
		BrokerTxAlphaPaymentConfirmed,
		session.GenesisHash).All(&[]BrokerBrokerTransaction{})

	oyster_utils.LogIfError(err, nil)
	return err
}

/*CreateTreasurePayload makes a payload for a treasure chunk by encrypting an ethereum private key using a sidechain
hash as the key.*/
func CreateTreasurePayload(ethereumSeed string, sha256Hash string, maxSideChainLength int) (string, error) {
	keyLocation := rand.Intn(maxSideChainLength)

	currentHash := sha256Hash
	for i := 0; i <= keyLocation; i++ {
		currentHash = oyster_utils.HashHex(currentHash, sha3.New256())
	}

	encryptedResult := oyster_utils.Encrypt(currentHash, TreasurePrefix+ethereumSeed, sha256Hash)

	treasurePayload := string(oyster_utils.BytesToTrytes(encryptedResult)) + oyster_utils.RandSeq(TreasureChunkPadding,
		oyster_utils.TrytesAlphabet)

	return treasurePayload, nil
}

/*GetChunkForWebnodePoW gets a chunk for webnode PoW and updates the index of the session.*/
func GetChunkForWebnodePoW() (oyster_utils.ChunkData, error) {
	var chunkData oyster_utils.ChunkData
	var err error
	var vErr *validate.Errors
	var i = 0
	sessions, err := GetReadySessions()

	for i := range sessions {
		if sessions[i].Type == SessionTypeAlpha {
			if sessions[i].NextIdxToAttach != int64(sessions[i].NumChunks-1) {
				chunkData = GetSingleChunkData(oyster_utils.InProgressDir,
					sessions[i].GenesisHash,
					sessions[i].NextIdxToAttach)
				sessions[i].NextIdxToAttach++
				break
			}
		} else {
			if sessions[i].NextIdxToAttach != int64(0) {
				chunkData = GetSingleChunkData(oyster_utils.InProgressDir,
					sessions[i].GenesisHash,
					sessions[i].NextIdxToAttach)
				sessions[i].NextIdxToAttach--
				break
			}
		}
	}

	vErr, err = DB.ValidateAndUpdate(&sessions[i])
	oyster_utils.LogIfValidationError("updating session in verify_data_maps", vErr, nil)
	oyster_utils.LogIfError(err, nil)
	return chunkData, err
}

/*ProcessAndStoreChunkData receives the genesis hash, chunk idx, and message from the client
and adds it to the badger database*/
func ProcessAndStoreChunkData(chunks []ChunkReq, genesisHash string, treasureIdxMap []int, ttl time.Duration) {

	if oyster_utils.DataMapStorageMode == oyster_utils.DataMapsInSQL {
		ProcessAndStoreChunkDataInSQL(chunks, genesisHash, treasureIdxMap)
		return
	}

	// the keys in this chunks map have already transformed indexes
	chunksMap := convertToBadgerKeyedMapForChunks(chunks, genesisHash, treasureIdxMap)

	dbID := []string{oyster_utils.InProgressDir, chunks[0].Hash, oyster_utils.MessageDir}

	db := oyster_utils.GetOrInitUniqueBadgerDB(dbID)
	if db == nil {
		return
	}

	batchSetKvMap := oyster_utils.KVPairs{} // Store chunk.Data into KVStore
	for key, chunk := range chunksMap {
		batchSetKvMap[key] = chunk.Data
	}

	err := oyster_utils.BatchSetToUniqueDB(dbID, &batchSetKvMap, ttl)
	oyster_utils.LogIfError(err, nil)
	if err != nil {
		panic(err)
	}
}

// convertToBadgerKeyedMapForChunks converts chunkReq into maps where the key is the badger msg_id.
// Return minChunkId and maxChunkId.
func convertToBadgerKeyedMapForChunks(chunks []ChunkReq, genesisHash string, treasureIdxMap []int) map[string]ChunkReq {
	chunksMap := make(map[string]ChunkReq)

	var chunkIdx int
	for _, chunk := range chunks {
		if oyster_utils.BrokerMode == oyster_utils.TestModeNoTreasure {
			chunkIdx = chunk.Idx
		} else {
			chunkIdx = oyster_utils.TransformIndexWithBuriedIndexes(chunk.Idx, treasureIdxMap)
		}

		key := oyster_utils.GetBadgerKey([]string{genesisHash, strconv.Itoa(chunkIdx)})
		chunksMap[key] = chunk
	}
	return chunksMap
}

/*ProcessAndStoreChunkDataInSQL receives chunk data from the client and stores the data in data_maps rows.*/
func ProcessAndStoreChunkDataInSQL(chunks []ChunkReq, genesisHash string, treasureIdxMap []int) {
	// the keys in this chunks map have already transformed indexes
	chunksMap := convertToBadgerKeyedMapForChunks(chunks, genesisHash, treasureIdxMap)

	batchSetKvMap := oyster_utils.KVPairs{} // Store chunk.Data into KVStore
	for key, chunk := range chunksMap {

		batchSetKvMap[key] = chunk.Data
	}

	oyster_utils.BatchSet(&batchSetKvMap, DataMapsTimeToLive)
}

/*GetSingleChunkData gets data about a single chunk.*/
func GetSingleChunkData(prefix string, genesisHash string, chunkIdx int64) oyster_utils.ChunkData {
	if oyster_utils.DataMapStorageMode == oyster_utils.DataMapsInBadger {
		return oyster_utils.GetChunkData(prefix, genesisHash, chunkIdx)
	}

	inProgressDataMaps := []DataMap{}
	completedDataMaps := []CompletedDataMap{}
	key := ""
	address := ""
	hash := ""

	if prefix == oyster_utils.InProgressDir {
		key = oyster_utils.GenerateBadgerKey("", genesisHash, int(chunkIdx))
		DB.Where("genesis_hash = ? AND chunk_idx = ?", genesisHash, int(chunkIdx)).All(&inProgressDataMaps)

		if len(inProgressDataMaps) > 0 {
			address = inProgressDataMaps[0].Address
			hash = inProgressDataMaps[0].Hash
		}
	} else {
		key = oyster_utils.GenerateBadgerKey(CompletedDataMapsMsgIDPrefix, genesisHash, int(chunkIdx))
		DB.Where("genesis_hash = ? AND chunk_idx = ?", genesisHash, int(chunkIdx)).All(&completedDataMaps)

		if len(completedDataMaps) > 0 {
			address = completedDataMaps[0].Address
			hash = completedDataMaps[0].Hash
		}
	}
	chunkData := oyster_utils.ChunkData{
		Address:     address,
		Hash:        hash,
		GenesisHash: genesisHash,
		Idx:         chunkIdx,
	}
	return returnChunkDataSQL(key, chunkData)
}

func returnChunkDataSQL(key string, chunkData oyster_utils.ChunkData) oyster_utils.ChunkData {
	rawMessage := ""
	message := ""
	values, _ := oyster_utils.BatchGet(&oyster_utils.KVKeys{key})
	if v, hasKey := (*values)[key]; hasKey {
		rawMessage = v
	}

	if rawMessage != "" {
		trytesMessage, err := oyster_utils.ChunkMessageToTrytesWithStopper(rawMessage)
		oyster_utils.LogIfError(err, nil)
		if err == nil {
			message = string(trytesMessage)
		}
	}

	return oyster_utils.ChunkData{
		Address:     chunkData.Address,
		RawMessage:  rawMessage,
		Message:     message,
		Hash:        chunkData.Hash,
		Idx:         chunkData.Idx,
		GenesisHash: chunkData.GenesisHash,
	}
}

/*GetMultiChunkData gets data about multiple chunks.  It will only return data for a chunk if both the hash and message
is ready.*/
func GetMultiChunkData(prefix string, genesisHash string, ks *oyster_utils.KVKeys) ([]oyster_utils.ChunkData, error) {
	if oyster_utils.DataMapStorageMode == oyster_utils.DataMapsInBadger {
		return oyster_utils.GetBulkChunkData(prefix, genesisHash, ks)
	}

	chunkData := []oyster_utils.ChunkData{}

	for _, key := range *(ks) {
		chunkIdx := oyster_utils.GetChunkIdxFromKey(key)
		singleChunkData := GetSingleChunkData(prefix, genesisHash, chunkIdx)

		if singleChunkData.Hash != "" && singleChunkData.RawMessage != "" {
			chunkData = append(chunkData, singleChunkData)
		}
	}
	return chunkData, nil
}

/*GetMultiChunkDataFromAnyDB gets data about multiple chunks.  It will get the data regardless of whether the chunks
are in in-progress or complete database*/
func GetMultiChunkDataFromAnyDB(genesisHash string, ks *oyster_utils.KVKeys) ([]oyster_utils.ChunkData, error) {
	if oyster_utils.DataMapStorageMode == oyster_utils.DataMapsInBadger {
		chunkDataInProgress, err := oyster_utils.GetBulkChunkData(oyster_utils.InProgressDir, genesisHash, ks)
		chunkDataComplete := []oyster_utils.ChunkData{}
		oyster_utils.LogIfError(err, nil)
		if len(chunkDataInProgress) != len(*ks) {
			chunkDataComplete, err = oyster_utils.GetBulkChunkData(oyster_utils.CompletedDir, genesisHash, ks)
			oyster_utils.LogIfError(err, nil)
		}
		return reassembleChunks(chunkDataInProgress, chunkDataComplete, ks), nil
	}

	chunkData := []oyster_utils.ChunkData{}

	for _, key := range *(ks) {
		chunkIdx := oyster_utils.GetChunkIdxFromKey(key)
		singleChunkDataInProgress := GetSingleChunkData(oyster_utils.InProgressDir, genesisHash, chunkIdx)

		if singleChunkDataInProgress.Hash != "" && singleChunkDataInProgress.RawMessage != "" {
			chunkData = append(chunkData, singleChunkDataInProgress)
		} else {
			singleChunkDataComplete := GetSingleChunkData(oyster_utils.CompletedDir, genesisHash, chunkIdx)

			if singleChunkDataComplete.Hash != "" && singleChunkDataComplete.RawMessage != "" {
				chunkData = append(chunkData, singleChunkDataComplete)
			}
		}
	}
	return chunkData, nil
}

func reassembleChunks(chunkDataInProgress []oyster_utils.ChunkData, chunkDataComplete []oyster_utils.ChunkData,
	ks *oyster_utils.KVKeys) []oyster_utils.ChunkData {

	if len(chunkDataInProgress) == 0 {
		return chunkDataComplete
	}
	if len(chunkDataComplete) == 0 {
		return chunkDataInProgress
	}

	chunkData := []oyster_utils.ChunkData{}
	keyChunkMap := make(map[string]oyster_utils.ChunkData)

	for _, chunk := range chunkDataInProgress {
		key := oyster_utils.GetBadgerKey([]string{chunk.GenesisHash,
			strconv.FormatInt(int64(chunk.Idx), 10)})
		keyChunkMap[key] = chunk
	}

	for _, chunk := range chunkDataComplete {
		key := oyster_utils.GetBadgerKey([]string{chunk.GenesisHash,
			strconv.FormatInt(int64(chunk.Idx), 10)})
		keyChunkMap[key] = chunk
	}

	for _, key := range *ks {
		if _, ok := keyChunkMap[key]; ok {
			chunkData = append(chunkData, keyChunkMap[key])
		}
	}

	return chunkData
}

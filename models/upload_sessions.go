package models

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/crypto/sha3"
	"math"
	"math/big"
	"math/rand"
	"strconv"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/pop/nulls"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/oysterprotocol/brokernode/utils"
	"github.com/shopspring/decimal"
)

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

// BuildDataMaps builds the datamap and inserts them into the DB.
func BuildDataMapsForSession(genHash string, numChunks int) (err error) {

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
		if err == nil {
			finished, err := u.WaitForAllChunks(MaxTimesToCheckForAllChunks)
			oyster_utils.LogIfError(err, nil)
			if finished {
				u.AllDataReady = AllDataReady
			}
			DB.ValidateAndUpdate(u)
		}
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

func (u *UploadSession) EncryptTreasureChunkEthKey(unencryptedKey string) (string, error) {
	encryptedKey := oyster_utils.ReturnEncryptedEthKey(u.ID, u.CreatedAt, unencryptedKey)
	return encryptedKey, nil
}

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

func (u *UploadSession) CheckIfAllDataIsReady() bool {

	return u.CheckIfAllHashesAreReady() && u.CheckIfAllMessagesAreReady()
}

func (u *UploadSession) CheckIfAllHashesAreReady() bool {

	chunkDataStartHash := oyster_utils.GetHashData(oyster_utils.InProgressDir, u.GenesisHash, 0)
	chunkDataEndHash := oyster_utils.GetHashData(oyster_utils.InProgressDir, u.GenesisHash, int64(u.NumChunks-1))

	return chunkDataStartHash != "" && chunkDataEndHash != ""
}

func (u *UploadSession) CheckIfAllMessagesAreReady() bool {
	treasureIndexes, err := u.GetTreasureIndexes()
	if err != nil || (len(treasureIndexes) == 0 && oyster_utils.BrokerMode != oyster_utils.TestModeNoTreasure) {
		oyster_utils.LogIfError(err, nil)
		return false
	}

	chunkDataStartMessage := oyster_utils.GetMessageData(oyster_utils.InProgressDir, u.GenesisHash, 0)
	chunkDataEndMessage := oyster_utils.GetMessageData(oyster_utils.InProgressDir, u.GenesisHash, int64(u.NumChunks-1))

	allMessagesFound := false
	if chunkDataStartMessage != "" && chunkDataEndMessage != "" {
		allMessagesFound = true
		if len(treasureIndexes) > 0 {
			for _, index := range treasureIndexes {
				chunkDataTreasureHash :=
					oyster_utils.GetMessageData(oyster_utils.InProgressDir, u.GenesisHash, int64(index))
				if chunkDataTreasureHash == "" {
					allMessagesFound = false
				}
			}
		}
	}
	return allMessagesFound
}

func (u *UploadSession) GetUnassignedChunksBySession(limit int) (chunkData []oyster_utils.ChunkData, err error) {
	var stopChunkIdx int64

	if u.Type == SessionTypeAlpha {
		stopChunkIdx = u.NextIdxToAttach + int64(limit) - 1
		if stopChunkIdx > int64(u.NumChunks-1) {
			stopChunkIdx = int64(u.NumChunks - 1)
		}
	} else {
		stopChunkIdx = u.NextIdxToAttach - int64(limit) + 1
		if stopChunkIdx < 0 {
			stopChunkIdx = int64(0)
		}
	}

	keys := oyster_utils.GenerateBulkKeys(u.GenesisHash, u.NextIdxToAttach, stopChunkIdx)

	chunkData, err = oyster_utils.GetBulkChunkData(oyster_utils.InProgressDir, u.GenesisHash, keys)
	oyster_utils.LogIfError(err, nil)

	return chunkData, err
}

func (u *UploadSession) MoveChunksToCompleted(chunks []oyster_utils.ChunkData) {
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
}

func (u *UploadSession) MoveAllChunksToCompleted() error {
	inProgressMessageDBID := []string{oyster_utils.InProgressDir, u.GenesisHash, oyster_utils.MessageDir}
	inProgressHashDBID := []string{oyster_utils.InProgressDir, u.GenesisHash, oyster_utils.HashDir}

	completeMessageDBID := []string{oyster_utils.CompletedDir, u.GenesisHash, oyster_utils.MessageDir}
	completeHashDBID := []string{oyster_utils.CompletedDir, u.GenesisHash, oyster_utils.HashDir}

	keys := oyster_utils.GenerateBulkKeys(u.GenesisHash, 0, int64(u.NumChunks)-1)

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

	return nil
}

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
			key := oyster_utils.GetBadgerKey([]string{u.GenesisHash, strconv.FormatInt(int64(i), 10)})
			kvs, err := oyster_utils.BatchGetFromUniqueDB([]string{oyster_utils.CompletedDir, u.GenesisHash,
				oyster_utils.HashDir}, &oyster_utils.KVKeys{key})
			oyster_utils.LogIfError(err, nil)
			if _, hasKey := (*kvs)[key]; !hasKey {
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
			key := oyster_utils.GetBadgerKey([]string{u.GenesisHash, strconv.FormatInt(int64(i), 10)})
			kvs, err := oyster_utils.BatchGetFromUniqueDB([]string{oyster_utils.CompletedDir, u.GenesisHash,
				oyster_utils.HashDir}, &oyster_utils.KVKeys{key})
			oyster_utils.LogIfError(err, nil)
			if _, hasKey := (*kvs)[key]; !hasKey {
				break
			} else {
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

func (u *UploadSession) DownGradeIndexesOnUnattachedChunks(chunks []oyster_utils.ChunkData) {
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

func GetSessionsByAge() ([]UploadSession, error) {
	sessionsByAge := []UploadSession{}

	err := DB.RawQuery("SELECT * FROM upload_sessions WHERE payment_status = ? AND "+
		"treasure_status = ? AND all_data_ready = ? ORDER BY created_at ASC",
		PaymentStatusConfirmed, TreasureInDataMapComplete, AllDataReady).All(&sessionsByAge)

	if err != nil {
		fmt.Println(err)
		oyster_utils.LogIfError(err, nil)
		return nil, err
	}

	return sessionsByAge, nil
}

func GetCompletedSessions() ([]UploadSession, error) {
	completedSessions := []UploadSession{}
	sessions, err := GetSessionsByAge()

	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return nil, err
	}

	for _, session := range sessions {
		stop := session.NumChunks - 1
		step := 1
		if session.Type == SessionTypeBeta {
			stop = 0
			step = -1
		}
		if session.NextIdxToVerify == int64(stop+step) {
			completedSessions = append(completedSessions, session)
		}
	}

	return completedSessions, nil
}

// GetSessionsThatNeedKeysEncrypted checks for sessions which the user has paid their PRL but in which
// we have not yet encrypted the keys
func GetSessionsThatNeedKeysEncrypted() ([]UploadSession, error) {
	needKeysEncrypted := []UploadSession{}

	err := DB.Where("payment_status = ? AND treasure_status = ?",
		PaymentStatusConfirmed, TreasureGeneratingKeys).All(&needKeysEncrypted)
	oyster_utils.LogIfError(err, nil)

	return needKeysEncrypted, err
}

// GetSessionsThatNeedTreasure checks for sessions which the user has paid their PRL but in which
// we have not yet buried the treasure.
func GetSessionsThatNeedTreasure() ([]UploadSession, error) {
	unburiedSessions := []UploadSession{}

	err := DB.Where("payment_status = ? AND treasure_status = ?",
		PaymentStatusConfirmed, TreasureInDataMapPending).All(&unburiedSessions)
	oyster_utils.LogIfError(err, nil)

	return unburiedSessions, err
}

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

func GetChunkForWebnodePoW() (oyster_utils.ChunkData, error) {
	var chunkData oyster_utils.ChunkData
	var err error
	var vErr *validate.Errors
	var i = 0
	sessions, err := GetReadySessions()

	for i := range sessions {
		if sessions[i].Type == SessionTypeAlpha {
			if sessions[i].NextIdxToAttach != int64(sessions[i].NumChunks-1) {
				chunkData = oyster_utils.GetChunkData(oyster_utils.InProgressDir,
					sessions[i].GenesisHash,
					sessions[i].NextIdxToAttach)
				sessions[i].NextIdxToAttach++
				break
			}
		} else {
			if sessions[i].NextIdxToAttach != int64(0) {
				chunkData = oyster_utils.GetChunkData(oyster_utils.InProgressDir,
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

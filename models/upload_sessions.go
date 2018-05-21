package models

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/oysterprotocol/brokernode/utils"
	"golang.org/x/crypto/sha3"
	"math"
	"time"

	"github.com/getsentry/raven-go"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/pop/nulls"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

// Enum for upload session type.
const (
	SessionTypeAlpha int = iota + 1
	SessionTypeBeta
)

type Invoice struct {
	Cost       float64      `json:"cost"`
	EthAddress nulls.String `json:"ethAddress"`
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
	FileSizeBytes        int       `json:"fileSizeBytes" db:"file_size_bytes"` // In Trytes rather than Bytes
	StorageLengthInYears int       `json:"storageLengthInYears" db:"storage_length_in_years"`
	Type                 int       `json:"type" db:"type"`

	ETHAddrAlpha  nulls.String `json:"ethAddrAlpha" db:"eth_addr_alpha"`
	ETHAddrBeta   nulls.String `json:"ethAddrBeta" db:"eth_addr_beta"`
	ETHPrivateKey string       `db:"eth_private_key"`
	// TODO: Floats shouldn't be used for prices, use https://github.com/shopspring/decimal.
	TotalCost      float64 `json:"totalCost" db:"total_cost"`
	PaymentStatus  int     `json:"paymentStatus" db:"payment_status"`
	TreasureStatus int     `json:"treasureStatus" db:"treasure_status"`

	TreasureIdxMap nulls.String `json:"treasureIdxMap" db:"treasure_idx_map"`
	// TreasureIdxMap slices.Int `json:"treasureIdxMap" db:"treasure_idx_map"`
}

const (
	PaymentStatusInvoiced int = iota + 1
	PaymentStatusPending
	PaymentStatusConfirmed
	PaymentStatusError = -1
)

const (
	TreasureGeneratingKeys int = iota + 1
	TreasureBurying
	TreasureBuried
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
	return validate.Validate(
		&validators.StringIsPresent{Field: u.GenesisHash, Name: "GenesisHash"},
		&validators.IntIsPresent{Field: u.NumChunks, Name: "NumChunks"},
		&validators.IntIsPresent{Field: u.FileSizeBytes, Name: "FileSize"},
	), nil
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

	switch oyster_utils.BrokerMode {
	case oyster_utils.ProdMode:
		// Defaults to paymentStatusPending
		if u.PaymentStatus == 0 {
			u.PaymentStatus = PaymentStatusInvoiced
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
			u.TreasureStatus = TreasureBurying
		}
	case oyster_utils.TestModeNoTreasure:
		// Defaults to paymentStatusPaid
		if u.PaymentStatus == 0 {
			u.PaymentStatus = PaymentStatusConfirmed
		}

		// Defaults to treasureBuried
		if u.TreasureStatus == 0 {
			u.TreasureStatus = TreasureBuried
		}
	}

	return nil
}

/**
 * Methods
 */

// StartUploadSession will generate dataMaps and save the session and dataMaps
// to the DB.
func (u *UploadSession) StartUploadSession() (vErr *validate.Errors, err error) {
	if u.Type == SessionTypeAlpha {
		u.calculatePayment()
	}

	vErr, err = DB.ValidateAndCreate(u)
	if err != nil || len(vErr.Errors) > 0 {
		fmt.Println(err)
		raven.CaptureError(err, nil)
		return
	}

	u.EncryptSessionEthKey()

	vErr, err = BuildDataMaps(u.GenesisHash, u.NumChunks)
	return
}

// TODO: Chunk this to smaller batches?
// DataMapsForSession fetches the datamaps associated with the session.
func (u *UploadSession) DataMapsForSession() (dMaps *[]DataMap, err error) {
	dMaps = &[]DataMap{}
	err = DB.RawQuery("SELECT * from data_maps WHERE genesis_hash = ? ORDER BY chunk_idx asc", u.GenesisHash).All(dMaps)
	if err != nil {
		raven.CaptureError(err, nil)
	}

	return
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
	storagePeg := getStoragePeg()
	fileSizeGigaBytes := int(math.Ceil(float64(u.FileSizeBytes / 1000000000)))
	if fileSizeGigaBytes < 1 {
		fileSizeGigaBytes = 1
	}

	u.TotalCost = float64(storagePeg * u.StorageLengthInYears * fileSizeGigaBytes)
}

func (u *UploadSession) GetTreasureMap() ([]TreasureMap, error) {
	var err error
	treasureIndex := []TreasureMap{}
	if u.TreasureIdxMap.Valid {
		// only do this if the string value is valid
		err = json.Unmarshal([]byte(u.TreasureIdxMap.String), &treasureIndex)
		if err != nil {
			fmt.Println(err)
			raven.CaptureError(err, nil)
		}
	}

	return treasureIndex, err
}

func (u *UploadSession) SetTreasureMap(treasureIndexMap []TreasureMap) error {
	var err error
	u.TreasureIdxMap = nulls.String{}
	DB.ValidateAndSave(u)
	treasureString, err := json.Marshal(treasureIndexMap)
	if err != nil {
		fmt.Println(err)
		raven.CaptureError(err, nil)
		return err
	}
	u.TreasureIdxMap = nulls.String{string(treasureString), true}
	DB.ValidateAndSave(u)
	return nil
}

// Sets the TreasureIdxMap with Sector, Idx, and Key
func (u *UploadSession) MakeTreasureIdxMap(mergedIndexes []int, privateKeys []string) {

	treasureIndexArray := make([]TreasureMap, 0)

	for i, mergedIndex := range mergedIndexes {
		treasureChunks, err := GetDataMapByGenesisHashAndChunkIdx(u.GenesisHash, mergedIndex)
		if err != nil {
			fmt.Println(err)
			return
		}
		if len(treasureChunks) == 0 || len(treasureChunks) > 1 {
			err = errors.New("did not find a chunk that matched genesis_hash and chunk_idx in MakeTreasureIdxMap, or " +
				"found duplicate chunks")
			raven.CaptureError(err, nil)
			return
		}

		hashedSessionID := oyster_utils.HashString(fmt.Sprint(u.ID), sha3.New256())
		hashedChunkCreationTime := oyster_utils.HashString(fmt.Sprint(treasureChunks[0].CreatedAt), sha3.New256())

		encryptedKey := oyster_utils.Encrypt(hashedSessionID, privateKeys[i], hashedChunkCreationTime)

		treasureIndexArray = append(treasureIndexArray, TreasureMap{
			Sector: i,
			Idx:    mergedIndex,
			Key:    hex.EncodeToString(encryptedKey),
		})
	}

	treasureString, err := json.Marshal(treasureIndexArray)
	if err != nil {
		fmt.Println(err)
		raven.CaptureError(err, nil)
	}

	u.TreasureIdxMap = nulls.String{string(treasureString), true}
	DB.ValidateAndSave(u)
}

func (u *UploadSession) GetTreasureIndexes() ([]int, error) {
	treasureMap, err := u.GetTreasureMap()
	if err != nil {
		fmt.Println(err)
		raven.CaptureError(err, nil)
	}
	treasureIndexArray := make([]int, 0)
	for _, treasure := range treasureMap {
		treasureIndexArray = append(treasureIndexArray, treasure.Idx)
	}
	return treasureIndexArray, err
}

func (u *UploadSession) BulkMarkDataMapsAsUnassigned() error {
	err := DB.RawQuery("UPDATE data_maps SET status = ? "+
		"WHERE genesis_hash = ? AND status = ? AND message != ?",
		Unassigned,
		u.GenesisHash,
		Pending,
		DataMap{}.Message).All(&[]DataMap{})
	if err != nil {
		raven.CaptureError(err, nil)
	}
	return err
}

func getStoragePeg() int {
	return 1 // TODO: write code to query smart contract to get real storage peg
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
	hashedSessionID := oyster_utils.HashString(fmt.Sprint(u.ID), sha3.New256())
	hashedCreationTime := oyster_utils.HashString(fmt.Sprint(u.CreatedAt), sha3.New256())

	encryptedKey := oyster_utils.Encrypt(hashedSessionID, u.ETHPrivateKey, hashedCreationTime)

	u.ETHPrivateKey = hex.EncodeToString(encryptedKey)
	DB.ValidateAndSave(u)
}

func (u *UploadSession) DecryptSessionEthKey() string {
	hashedSessionID := oyster_utils.HashString(fmt.Sprint(u.ID), sha3.New256())
	hashedCreationTime := oyster_utils.HashString(fmt.Sprint(u.CreatedAt), sha3.New256())

	decryptedKey := oyster_utils.Decrypt(hashedSessionID, u.ETHPrivateKey, hashedCreationTime)

	return hex.EncodeToString(decryptedKey)
}

func GetSessionsByAge() ([]UploadSession, error) {
	sessionsByAge := []UploadSession{}

	err := DB.RawQuery("SELECT * from upload_sessions WHERE payment_status = ? AND "+
		"treasure_status = ? ORDER BY created_at asc", PaymentStatusConfirmed, TreasureBuried).All(&sessionsByAge)

	if err != nil {
		fmt.Println(err)
		raven.CaptureError(err, nil)
		return nil, err
	}

	return sessionsByAge, nil
}

// GetSessionsThatNeedTreasure checks for sessions which the user has paid their PRL but in which
// we have not yet buried the treasure.
func GetSessionsThatNeedTreasure() ([]UploadSession, error) {
	unburiedSessions := []UploadSession{}

	err := DB.Where("payment_status = ? AND treasure_status = ?",
		PaymentStatusConfirmed, TreasureBurying).All(&unburiedSessions)
	if err != nil {
		raven.CaptureError(err, nil)
	}

	return unburiedSessions, err
}

func GetReadySessions() ([]UploadSession, error) {
	readySessions := []UploadSession{}

	err := DB.Where("payment_status = ? AND treasure_status = ?",
		PaymentStatusConfirmed, TreasureBuried).All(&readySessions)
	if err != nil {
		raven.CaptureError(err, nil)
	}

	return readySessions, err
}

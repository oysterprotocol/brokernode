package models

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"os"
	"time"

	"github.com/oysterprotocol/brokernode/utils"
	"golang.org/x/crypto/sha3"

	"github.com/getsentry/raven-go"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/pop/nulls"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/shopspring/decimal"
)

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
	FileSizeBytes        int       `json:"fileSizeBytes" db:"file_size_bytes"` // In Trytes rather than Bytes
	StorageLengthInYears int       `json:"storageLengthInYears" db:"storage_length_in_years"`
	Type                 int       `json:"type" db:"type"`

	ETHAddrAlpha   nulls.String    `json:"ethAddrAlpha" db:"eth_addr_alpha"`
	ETHAddrBeta    nulls.String    `json:"ethAddrBeta" db:"eth_addr_beta"`
	ETHPrivateKey  string          `db:"eth_private_key"`
	TotalCost      decimal.Decimal `json:"totalCost" db:"total_cost"`
	PaymentStatus  int             `json:"paymentStatus" db:"payment_status"`
	TreasureStatus int             `json:"treasureStatus" db:"treasure_status"`

	TreasureIdxMap nulls.String `json:"treasureIdxMap" db:"treasure_idx_map"`
}

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

var StoragePeg = decimal.NewFromFloat(float64(64)) // GB per year per PRL; TODO: query smart contract for real storage peg

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
			if os.Getenv("OYSTER_PAYS") == "" {
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
			u.TreasureStatus = TreasureInDataMapPending
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

// StartUploadSession will generate dataMaps and save the session and dataMaps
// to the DB.
func (u *UploadSession) StartUploadSession() (vErr *validate.Errors, err error) {
	u.calculatePayment()

	vErr, err = DB.ValidateAndCreate(u)
	if err != nil || len(vErr.Errors) > 0 {
		fmt.Println(err)
		raven.CaptureError(err, nil)
		return
	}

	u.EncryptSessionEthKey()

	if oyster_utils.BrokerMode != oyster_utils.TestModeNoTreasure {
		u.NumChunks = oyster_utils.GetTotalFileChunkIncludingBuriedPearlsUsingNumChunks(u.NumChunks)
		DB.ValidateAndUpdate(u)
	}

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
	_, err = DB.ValidateAndSave(u)
	oyster_utils.LogIfError(err)
	treasureString, err := json.Marshal(treasureIndexMap)
	if err != nil {
		fmt.Println(err)
		raven.CaptureError(err, nil)
		return err
	}
	u.TreasureIdxMap = nulls.String{string(treasureString), true}
	_, err = DB.ValidateAndSave(u)
	oyster_utils.LogIfError(err)
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

		encryptedKey, err := treasureChunks[0].EncryptEthKey(privateKeys[i])
		if err != nil {
			fmt.Println(err)
			return
		}

		treasureIndexArray = append(treasureIndexArray, TreasureMap{
			Sector: i,
			Idx:    mergedIndex,
			Key:    encryptedKey,
		})
	}

	treasureString, err := json.Marshal(treasureIndexArray)
	if err != nil {
		fmt.Println(err)
		raven.CaptureError(err, nil)
	}

	u.TreasureIdxMap = nulls.String{string(treasureString), true}
	u.TreasureStatus = TreasureInDataMapPending

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
		fmt.Println(err)
		raven.CaptureError(err, nil)
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
	hashedSessionID := oyster_utils.HashHex(hex.EncodeToString([]byte(fmt.Sprint(u.ID))), sha3.New256())
	hashedCreationTime := oyster_utils.HashHex(hex.EncodeToString([]byte(fmt.Sprint(u.CreatedAt))), sha3.New256())

	encryptedKey := oyster_utils.Encrypt(hashedSessionID, u.ETHPrivateKey, hashedCreationTime)

	u.ETHPrivateKey = hex.EncodeToString(encryptedKey)
	DB.ValidateAndSave(u)
}

func (u *UploadSession) DecryptSessionEthKey() string {
	hashedSessionID := oyster_utils.HashHex(hex.EncodeToString([]byte(fmt.Sprint(u.ID))), sha3.New256())
	hashedCreationTime := oyster_utils.HashHex(hex.EncodeToString([]byte(fmt.Sprint(u.CreatedAt))), sha3.New256())

	decryptedKey := oyster_utils.Decrypt(hashedSessionID, u.ETHPrivateKey, hashedCreationTime)

	return hex.EncodeToString(decryptedKey)
}

func GetSessionsByAge() ([]UploadSession, error) {
	sessionsByAge := []UploadSession{}

	err := DB.RawQuery("SELECT * from upload_sessions WHERE payment_status = ? AND "+
		"treasure_status = ? ORDER BY created_at asc", PaymentStatusConfirmed, TreasureInDataMapComplete).All(&sessionsByAge)

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
		PaymentStatusConfirmed, TreasureInDataMapPending).All(&unburiedSessions)
	if err != nil {
		raven.CaptureError(err, nil)
	}

	return unburiedSessions, err
}

func GetReadySessions() ([]UploadSession, error) {
	readySessions := []UploadSession{}

	err := DB.Where("payment_status = ? AND treasure_status = ?",
		PaymentStatusConfirmed, TreasureInDataMapComplete).All(&readySessions)
	if err != nil {
		raven.CaptureError(err, nil)
	}

	return readySessions, err
}

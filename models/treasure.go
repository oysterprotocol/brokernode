package models

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/getsentry/raven-go"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"github.com/oysterprotocol/brokernode/utils"
	"golang.org/x/crypto/sha3"
	"math/big"
	"time"
)

type PRLStatus int

// IMPORTANT:  Do not remove Message and Address from
// this struct; they are used for encryption
type Treasure struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	ETHAddr   string    `json:"ethAddr" db:"eth_addr"`
	ETHKey    string    `json:"ethKey" db:"eth_key"`
	PRLAmount string    `json:"prlAmount" db:"prl_amount"`
	PRLStatus PRLStatus `json:"prlStatus" db:"prl_status"`
	Message   string    `json:"message" db:"message"`
	Address   string    `json:"address" db:"address"`
}

const (
	// organizing these as a sequence to simplify queries and testing,
	// and because the process is a sequence anyway
	PRLWaiting PRLStatus = iota + 1
	PRLPending
	PRLConfirmed
	GasPending
	GasConfirmed
	BuryPending
	BuryConfirmed

	// error states
	PRLError  = -1
	GasError  = -2
	BuryError = -3
)

// String is not required by pop and may be deleted
func (t Treasure) String() string {
	jt, _ := json.Marshal(t)
	return string(jt)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (t *Treasure) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (t *Treasure) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (t *Treasure) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

/**
 * Callbacks
 */

func (t *Treasure) BeforeCreate(tx *pop.Connection) error {
	// Defaults to PRLWaiting.
	if t.PRLStatus == 0 {
		t.PRLStatus = PRLWaiting
	}

	t.EncryptTreasureEthKey()

	return nil
}

func (t *Treasure) SetPRLAmount(bigInt *big.Int) (string, error) {
	prlAmountAsBytes, err := bigInt.MarshalJSON()
	if err != nil {
		fmt.Println(err)
		raven.CaptureError(err, nil)
		return "", err
	}
	t.PRLAmount = string(prlAmountAsBytes)
	DB.ValidateAndUpdate(t)

	return t.PRLAmount, nil
}

func (t *Treasure) GetPRLAmount() *big.Int {

	prlAmountAsBytes := []byte(t.PRLAmount)
	var bigInt big.Int
	bigInt.UnmarshalJSON(prlAmountAsBytes)

	return &bigInt
}

func GetTreasuresToBuryByPRLStatus(prlStatus PRLStatus) ([]Treasure, error) {
	treasureToBury := []Treasure{}
	err := DB.RawQuery("SELECT * from treasures where prl_status = ?", prlStatus).All(&treasureToBury)
	if err != nil {
		fmt.Println(err)
		raven.CaptureError(err, nil)
		return treasureToBury, err
	}
	return treasureToBury, nil
}

func GetAllTreasuresToBury() ([]Treasure, error) {
	allTreasures := []Treasure{}
	err := DB.RawQuery("SELECT * from treasures").All(&allTreasures)
	return allTreasures, err
}

func (t *Treasure) EncryptTreasureEthKey() {
	hashedMessage := oyster_utils.HashHex(hex.EncodeToString([]byte(t.Message)), sha3.New256())
	hashedAddress := oyster_utils.HashHex(hex.EncodeToString([]byte(t.Address)), sha3.New256())
	encryptedKey := oyster_utils.Encrypt(hashedMessage, t.ETHKey, hashedAddress)
	t.ETHKey = hex.EncodeToString(encryptedKey)
}

func (t *Treasure) DecryptTreasureEthKey() string {
	hashedMessage := oyster_utils.HashHex(hex.EncodeToString([]byte(t.Message)), sha3.New256())
	hashedAddress := oyster_utils.HashHex(hex.EncodeToString([]byte(t.Address)), sha3.New256())
	decryptedKey := oyster_utils.Decrypt(hashedMessage, t.ETHKey, hashedAddress)
	return hex.EncodeToString(decryptedKey)
}

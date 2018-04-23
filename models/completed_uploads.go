package models

import (
	"encoding/json"
	"errors"
	"github.com/getsentry/raven-go"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

type CompletedUpload struct {
	ID            uuid.UUID `json:"id" db:"id"`
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt     time.Time `json:"updatedAt" db:"updated_at"`
	GenesisHash   string    `json:"genesisHash" db:"genesis_hash"`
	ETHAddr       string    `json:"ethAddr" db:"eth_addr"`
	ETHPrivateKey string    `json:"ethPrivateKey" db:"eth_private_key"`
	PRLStatus     int       `json:"prlStatus" db:"prl_status"`
	GasStatus     int       `json:"gasStatus" db:"gas_status"`
}

const (
	PRLClaimNotStarted int = iota + 1
	PRLClaimProcessing
	PRLClaimSuccess
	PRLClaimError = -1
)

const (
	GasTransferNotStarted int = iota + 1
	GasTransferProcessing
	GasTransferSuccess
	GasTransferError = -1
)

// String is not required by pop and may be deleted
func (c CompletedUpload) String() string {
	jc, _ := json.Marshal(c)
	return string(jc)
}

// CompletedUploads is not required by pop and may be deleted
type CompletedUploads []CompletedUpload

// String is not required by pop and may be deleted
func (c CompletedUploads) String() string {
	jc, _ := json.Marshal(c)
	return string(jc)
}

/**
 * Validations
 */

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (c *CompletedUpload) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: c.GenesisHash, Name: "GenesisHash"},
		&validators.StringIsPresent{Field: c.GenesisHash, Name: "EthAddr"},
		&validators.StringIsPresent{Field: c.GenesisHash, Name: "EthPrivateKey"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (c *CompletedUpload) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (c *CompletedUpload) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

/**
 * Callbacks
 */

func (c *CompletedUpload) BeforeCreate(tx *pop.Connection) error {

	// Defaults to ClaimNotBegun
	if c.PRLStatus == 0 {
		c.PRLStatus = PRLClaimNotStarted
	}

	// Defaults to GasTransferNotStarted
	if c.GasStatus == 0 {
		c.GasStatus = GasTransferNotStarted
	}
	return nil
}

/**
 * Methods
 */
func NewCompletedUpload(session UploadSession) error {

	var err error

	switch session.Type {
	case SessionTypeAlpha:
		_, err = DB.ValidateAndSave(&CompletedUpload{
			GenesisHash:   session.GenesisHash,
			ETHAddr:       session.ETHAddrAlpha.String,
			ETHPrivateKey: session.ETHPrivateKey})
	case SessionTypeBeta:
		_, err = DB.ValidateAndSave(&CompletedUpload{
			GenesisHash:   session.GenesisHash,
			ETHAddr:       session.ETHAddrBeta.String,
			ETHPrivateKey: session.ETHPrivateKey})
	default:
		err = errors.New("no session type provided for session in method models.NewCompletedUpload")
	}

	return err
}

func GetRowsByGasAndPRLStatus(gasStatus int, prlStatus int) (uploads []CompletedUpload, err error) {
	err = DB.Where("gas_status = ? AND prl_status = ?", gasStatus, prlStatus).All(&uploads)
	return uploads, err
}

func GetRowsByGasStatus(gasStatus int) (uploads []CompletedUpload, err error) {
	err = DB.Where("gas_status = ?", gasStatus).All(&uploads)
	return uploads, err
}

func SetGasStatus(uploads []CompletedUpload, newGasStatus int) {
	for _, upload := range uploads {
		upload.GasStatus = newGasStatus
		DB.ValidateAndSave(&upload)
	}
}

func GetRowsByPRLStatus(prlStatus int) (uploads []CompletedUpload, err error) {
	err = DB.Where("prl_status = ?", prlStatus).All(&uploads)
	return uploads, err
}

func SetPRLStatus(uploads []CompletedUpload, newPRLStatus int) {
	for _, upload := range uploads {
		upload.PRLStatus = newPRLStatus
		DB.ValidateAndSave(&upload)
	}
}

func GetTimedOutGasTransfers(thresholdTime time.Time) (uploads []CompletedUpload, err error) {
	err = DB.Where("gas_status = ? AND updated_at <= ?",
		GasTransferProcessing,
		thresholdTime).All(&uploads)
	return uploads, err
}

func GetTimedOutPRLTransfers(thresholdTime time.Time) (uploads []CompletedUpload, err error) {
	err = DB.Where("prl_status = ? AND updated_at <= ?",
		PRLClaimProcessing,
		thresholdTime).All(&uploads)
	return uploads, err
}

func SetGasStatusByAddress(transactionAddress string, newGasStatus int) {
	uploadRow := CompletedUpload{}
	err := DB.Where("eth_addr = ?", transactionAddress).First(&uploadRow)
	if err != nil {
		raven.CaptureError(err, nil)
		return
	}
	if uploadRow.ID == uuid.Nil {
		return
	}
	uploadRow.GasStatus = newGasStatus
	DB.ValidateAndSave(&uploadRow)
}

func SetPRLStatusByAddress(transactionAddress string, newPRLStatus int) {
	uploadRow := CompletedUpload{}
	err := DB.Where("eth_addr = ?", transactionAddress).First(&uploadRow)
	if err != nil {
		raven.CaptureError(err, nil)
		return
	}
	if uploadRow.ID == uuid.Nil {
		return
	}
	uploadRow.PRLStatus = newPRLStatus
	DB.ValidateAndSave(&uploadRow)
}

func DeleteCompletedClaims() error {
	err := DB.RawQuery("DELETE from completed_uploads WHERE prl_status = ?", PRLClaimSuccess).All(&[]CompletedUpload{})
	if err != nil {
		return err
	}

	return nil
}

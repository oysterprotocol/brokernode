package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/getsentry/raven-go"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/oysterprotocol/brokernode/utils"
)

type CompletedUpload struct {
	ID            uuid.UUID         `json:"id" db:"id"`
	CreatedAt     time.Time         `json:"createdAt" db:"created_at"`
	UpdatedAt     time.Time         `json:"updatedAt" db:"updated_at"`
	GenesisHash   string            `json:"genesisHash" db:"genesis_hash"`
	ETHAddr       string            `json:"ethAddr" db:"eth_addr"`
	ETHPrivateKey string            `json:"ethPrivateKey" db:"eth_private_key"`
	PRLStatus     PRLClaimStatus    `json:"prlStatus" db:"prl_status"`
	GasStatus     GasTransferStatus `json:"gasStatus" db:"gas_status"`
	Version       uint32            `json:"version" db:"version"`
}

type PRLClaimStatus int
type GasTransferStatus int

const (
	PRLClaimNotStarted PRLClaimStatus = iota + 1
	PRLClaimProcessing
	PRLClaimSuccess
	PRLClaimError = -1
)

const (
	GasTransferNotStarted GasTransferStatus = iota + 1
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
	oyster_utils.LogIfError(err, map[string]interface{}{"sessionType": session.Type})

	return err
}

func GetRowsByGasAndPRLStatus(gasStatus GasTransferStatus, prlStatus PRLClaimStatus) (uploads []CompletedUpload, err error) {
	err = DB.Where("gas_status = ? AND prl_status = ?", gasStatus, prlStatus).All(&uploads)
	if err != nil {
		raven.CaptureError(err, nil)
	}

	return uploads, err
}

func GetRowsByGasStatus(gasStatus GasTransferStatus) (uploads []CompletedUpload, err error) {
	err = DB.Where("gas_status = ?", gasStatus).All(&uploads)
	if err != nil {
		raven.CaptureError(err, nil)
	}

	return uploads, err
}

func SetGasStatus(uploads []CompletedUpload, newGasStatus GasTransferStatus) {
	for _, upload := range uploads {
		upload.GasStatus = newGasStatus
		DB.ValidateAndSave(&upload)
	}
}

func GetRowsByPRLStatus(prlStatus PRLClaimStatus) (uploads []CompletedUpload, err error) {
	err = DB.Where("prl_status = ?", prlStatus).All(&uploads)
	if err != nil {
		raven.CaptureError(err, nil)
	}

	return uploads, err
}

func SetPRLStatus(uploads []CompletedUpload, newPRLStatus PRLClaimStatus) {
	for _, upload := range uploads {
		upload.PRLStatus = newPRLStatus
		DB.ValidateAndSave(&upload)
	}
}

func GetTimedOutGasTransfers(thresholdTime time.Time) (uploads []CompletedUpload, err error) {
	err = DB.Where("gas_status = ? AND updated_at <= ?",
		GasTransferProcessing,
		thresholdTime).All(&uploads)
	if err != nil {
		raven.CaptureError(err, nil)
	}

	return uploads, err
}

func GetTimedOutPRLTransfers(thresholdTime time.Time) (uploads []CompletedUpload, err error) {
	err = DB.Where("prl_status = ? AND updated_at <= ?",
		PRLClaimProcessing,
		thresholdTime).All(&uploads)
	if err != nil {
		raven.CaptureError(err, nil)
	}

	return uploads, err
}

func SetGasStatusByAddress(transactionAddress string, newGasStatus GasTransferStatus) {
	uploadRow := CompletedUpload{}
	err := DB.Where("eth_addr = ?", transactionAddress).First(&uploadRow)
	if err != nil {
		fmt.Println(err)
		raven.CaptureError(err, nil)
		return
	}
	if uploadRow.ID == uuid.Nil {
		return
	}
	uploadRow.GasStatus = newGasStatus
	DB.ValidateAndSave(&uploadRow)
}

func SetPRLStatusByAddress(transactionAddress string, newPRLStatus PRLClaimStatus) {
	uploadRow := CompletedUpload{}
	err := DB.Where("eth_addr = ?", transactionAddress).First(&uploadRow)
	if err != nil {
		fmt.Println(err)
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
		fmt.Println(err)
		raven.CaptureError(err, nil)
		return err
	}

	return nil
}

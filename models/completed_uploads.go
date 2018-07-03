package models

import (
	"encoding/json"
	"errors"
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
	PRLTxHash     string            `json:"prlTxHash" db:"prl_tx_hash"`
	PRLTxNonce    int64             `json:"prlTxNonce" db:"prl_tx_nonce"`
	GasStatus     GasTransferStatus `json:"gasStatus" db:"gas_status"`
	GasTxHash     string            `json:"gasTxHash" db:"gas_tx_hash"`
	GasTxNonce    int64             `json:"gasTxNonce" db:"gas_tx_nonce"`
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
	GasTransferLeftoversReclaimProcessing
	GasTransferLeftoversReclaimSuccess

	GasTransferError                 = -1
	GasTransferLeftoversReclaimError = -2
)

var PRLClaimStatusMap = make(map[PRLClaimStatus]string)
var GasTransferStatusMap = make(map[GasTransferStatus]string)

func init() {
	PRLClaimStatusMap[PRLClaimNotStarted] = "PRLClaimNotStarted"
	PRLClaimStatusMap[PRLClaimProcessing] = "PRLClaimProcessing"
	PRLClaimStatusMap[PRLClaimSuccess] = "PRLClaimSuccess"
	PRLClaimStatusMap[PRLClaimError] = "PRLClaimError"

	GasTransferStatusMap[GasTransferNotStarted] = "GasTransferNotStarted"
	GasTransferStatusMap[GasTransferProcessing] = "GasTransferProcessing"
	GasTransferStatusMap[GasTransferSuccess] = "GasTransferSuccess"
	GasTransferStatusMap[GasTransferLeftoversReclaimProcessing] = "GasTransferLeftoversReclaimProcessing"
	GasTransferStatusMap[GasTransferLeftoversReclaimSuccess] = "GasTransferLeftoversReclaimSuccess"
	GasTransferStatusMap[GasTransferError] = "GasTransferError"
}

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
		&validators.StringIsPresent{Field: c.ETHAddr, Name: "EthAddr"},
		&validators.StringIsPresent{Field: c.ETHPrivateKey, Name: "EthPrivateKey"},
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

func (c *CompletedUpload) EncryptSessionEthKey() {
	c.ETHPrivateKey = oyster_utils.ReturnEncryptedEthKey(c.ID, c.CreatedAt, c.ETHPrivateKey)
	DB.ValidateAndSave(c)
}

func (c *CompletedUpload) DecryptSessionEthKey() string {
	return oyster_utils.ReturnDecryptedEthKey(c.ID, c.CreatedAt, c.ETHPrivateKey)
}

/**
 * Methods
 */
func NewCompletedUpload(session UploadSession) error {

	var err error
	var vErr *validate.Errors
	privateKey := session.DecryptSessionEthKey()
	completedUpload := CompletedUpload{}

	switch session.Type {
	case SessionTypeAlpha:
		completedUpload = CompletedUpload{
			GenesisHash:   session.GenesisHash,
			ETHAddr:       session.ETHAddrAlpha.String,
			ETHPrivateKey: privateKey,
		}

		vErr, err = DB.ValidateAndSave(&completedUpload)
		if err != nil {
			oyster_utils.LogIfError(err, nil)
		}
		if len(vErr.Errors) != 0 {
			oyster_utils.LogIfValidationError(
				"validation errors for creating completedUpload with SessionTypeAlpha.", vErr, nil)
		}
	case SessionTypeBeta:
		completedUpload = CompletedUpload{
			GenesisHash:   session.GenesisHash,
			ETHAddr:       session.ETHAddrBeta.String,
			ETHPrivateKey: privateKey,
		}

		vErr, err = DB.ValidateAndSave(&completedUpload)
		if err != nil {
			oyster_utils.LogIfError(err, nil)
		}
		if len(vErr.Errors) != 0 {
			oyster_utils.LogIfValidationError(
				"validation errors for creating completedUpload with SessionTypeBeta.", vErr, nil)
		}
	default:
		err = errors.New("no session type provided for session in method models.NewCompletedUpload")
		oyster_utils.LogIfError(err, map[string]interface{}{"sessionType": session.Type})
		return err
	}

	completedUpload.EncryptSessionEthKey()

	return nil
}

func GetRowsByGasAndPRLStatus(gasStatus GasTransferStatus, prlStatus PRLClaimStatus) (uploads []CompletedUpload, err error) {
	err = DB.Where("gas_status = ? AND prl_status = ?", gasStatus, prlStatus).All(&uploads)
	oyster_utils.LogIfError(err, nil)
	return uploads, err
}

func GetRowsByGasStatus(gasStatus GasTransferStatus) (uploads []CompletedUpload, err error) {
	err = DB.Where("gas_status = ?", gasStatus).All(&uploads)
	oyster_utils.LogIfError(err, nil)

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
	oyster_utils.LogIfError(err, nil)

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
	oyster_utils.LogIfError(err, nil)

	return uploads, err
}

func GetTimedOutPRLTransfers(thresholdTime time.Time) (uploads []CompletedUpload, err error) {
	err = DB.Where("prl_status = ? AND updated_at <= ?",
		PRLClaimProcessing,
		thresholdTime).All(&uploads)
	oyster_utils.LogIfError(err, nil)

	return uploads, err
}

func GetTimedOutGasReclaims(thresholdTime time.Time) (uploads []CompletedUpload, err error) {
	err = DB.Where("gas_status = ? AND updated_at <= ?",
		GasTransferLeftoversReclaimProcessing,
		thresholdTime).All(&uploads)
	oyster_utils.LogIfError(err, nil)

	return uploads, err
}

func SetGasStatusByAddress(transactionAddress string, newGasStatus GasTransferStatus) {
	uploadRow := CompletedUpload{}
	err := DB.Where("eth_addr = ?", transactionAddress).First(&uploadRow)
	if err != nil {
		oyster_utils.LogIfError(err, nil)
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
		oyster_utils.LogIfError(err, nil)
		return
	}
	if uploadRow.ID == uuid.Nil {
		return
	}
	uploadRow.PRLStatus = newPRLStatus
	DB.ValidateAndSave(&uploadRow)
}

func DeleteCompletedClaims() error {
	err := DB.RawQuery("DELETE from completed_uploads WHERE gas_status = ?",
		GasTransferLeftoversReclaimSuccess).All(&[]CompletedUpload{})
	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return err
	}

	return nil
}

package models

import (
	"encoding/json"
	"errors"
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
	ETHPrivateKey string    `db:"eth_private_key"`
	ClaimStatus   int       `json:"claimStatus" db:"claim_status"`
}

const (
	ClaimNotBegun int = iota + 1
	ClaimInProcess
	ClaimSuccess
	ClaimError = -1
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
	if c.ClaimStatus == 0 {
		c.ClaimStatus = ClaimNotBegun
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

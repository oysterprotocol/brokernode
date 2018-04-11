package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

// Enum for upload session type.
const (
	SessionTypeAlpha int = iota + 1
	SessionTypeBeta
)

const (
	Unpaid int = iota + 1
	Paid
)

const (
	Unburied int = iota + 1
	Buried
)

type UploadSession struct {
	ID             uuid.UUID `json:"id" db:"id"`
	CreatedAt      time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time `json:"updatedAt" db:"updated_at"`
	GenesisHash    string    `json:"genesisHash" db:"genesis_hash"`
	FileSizeBytes  int       `json:"fileSizeBytes" db:"file_size_bytes"`
	Type           int       `json:"type" db:"type"`
	PaymentStatus  int       `json:"paymentStatus" db:"payment_status"`
	TreasureStatus int       `json:"treasureStatus" db:"treasure_status"`
}

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
		&validators.IntIsPresent{Field: u.FileSizeBytes, Name: "FileSizeBytes"},
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

	return nil
}

/**
 * Methods
 */

// StartUploadSession will generate dataMaps and save the session and dataMaps
// to the DB.
func (u *UploadSession) StartUploadSession() (vErr *validate.Errors, err error) {
	vErr, err = DB.ValidateAndCreate(u)
	if err != nil || len(vErr.Errors) > 0 {
		return
	}

	vErr, err = BuildDataMaps(u.GenesisHash, u.FileSizeBytes)
	return
}

// TODO: Chunk this to smaller batches?
// DataMapsForSession fetches the datamaps associated with the session.
func (u *UploadSession) DataMapsForSession() (dMaps *[]DataMap, err error) {
	dMaps = &[]DataMap{}
	err = DB.RawQuery("SELECT * from data_maps WHERE genesis_hash = ? ORDER BY chunk_idx asc", u.GenesisHash).All(dMaps)

	return
}

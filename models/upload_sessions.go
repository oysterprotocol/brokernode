package models

import (
	"encoding/json"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"time"
)

// Enum for upload session type.
const (
	SessionTypeAlpha int = iota + 1
	SessionTypeBeta
)

type UploadSession struct {
	ID            uuid.UUID `json:"id" db:"id"`
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt     time.Time `json:"updatedAt" db:"updated_at"`
	GenesisHash   string    `json:"genesisHash" db:"genesis_hash"`
	FileSizeBytes int       `json:"fileSizeBytes" db:"file_size_bytes"`
	Type          int       `json:"type" db:"type"`
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

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (u *UploadSession) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: u.GenesisHash, Name: "GenesisHash"},
		&validators.IntIsPresent{Field: u.FileSizeBytes, Name: "FileSizeBytes"},
		&validators.IntIsPresent{Field: u.Type, Name: "Type"},
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

// StartUploadSession will generate dataMaps and save the session and dataMaps
// to the DB.
func (u *UploadSession) StartUploadSession() (vErr *validate.Errors, err error) {
	// Defaults to alpha session.
	if u.Type != SessionTypeBeta {
		u.Type = SessionTypeAlpha
	}
	vErr, err = DB.ValidateAndCreate(u)
	if err != nil || len(vErr.Errors) > 0 {
		return
	}

	vErr, err = BuildDataMaps(u.GenesisHash, u.FileSizeBytes)
	return
}

package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

type DataMap struct {
	ID          string    `json:"id" db:"id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	Status      int       `json:"status" db:"status"`
	HooknodeIp  string    `json:"hooknode_ip" db:"hooknode_ip"`
	Message     string    `json:"message" db:"message"`
	TrunkTx     string    `json:"trunk_tx" db:"trunk_tx"`
	BranchTx    string    `json:"branch_tx" db:"branch_tx"`
	GenesisHash string    `json:"genesis_hash" db:"genesis_hash"`
	ChunkIdx    int       `json:"chunk_idx" db:"chunk_idx"`
	Hash        string    `json:"hash" db:"hash"`
	Address     string    `json:"address" db:"address"`
}

// String is not required by pop and may be deleted
func (d DataMap) String() string {
	jd, _ := json.Marshal(d)
	return string(jd)
}

// DataMaps is not required by pop and may be deleted
type DataMaps []DataMap

// String is not required by pop and may be deleted
func (d DataMaps) String() string {
	jd, _ := json.Marshal(d)
	return string(jd)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (d *DataMap) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: d.HooknodeIp, Name: "HooknodeIp"},
		&validators.StringIsPresent{Field: d.Message, Name: "Message"},
		&validators.StringIsPresent{Field: d.TrunkTx, Name: "TrunkTx"},
		&validators.StringIsPresent{Field: d.ID, Name: "ID"},
		&validators.StringIsPresent{Field: d.GenesisHash, Name: "GenesisHash"},
		&validators.StringIsPresent{Field: d.Hash, Name: "Hash"},
		&validators.StringIsPresent{Field: d.Address, Name: "Address"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (d *DataMap) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (d *DataMap) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

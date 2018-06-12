package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

type CompletedDataMap struct {
	ID             int       `json:"id" db:"id"`
	CreatedAt      time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time `json:"updatedAt" db:"updated_at"`
	Status         int       `json:"status" db:"status"`
	NodeID         string    `json:"nodeID" db:"node_id"`
	NodeType       string    `json:"nodeType" db:"node_type"`
	Message        string    `json:"message" db:"message"`
	MsgID          string    `json:"msgId" db:"msg_id"`
	TrunkTx        string    `json:"trunkTx" db:"trunk_tx"`
	BranchTx       string    `json:"branchTx" db:"branch_tx"`
	GenesisHash    string    `json:"genesisHash" db:"genesis_hash"`
	ChunkIdx       int       `json:"chunkIdx" db:"chunk_idx"`
	Hash           string    `json:"hash" db:"hash"`
	ObfuscatedHash string    `json:"obfuscatedHash" db:"obfuscated_hash"`
	Address        string    `json:"address" db:"address"`
}

func init() {
}

// String is not required by pop and may be deleted
func (d CompletedDataMap) String() string {
	jd, _ := json.Marshal(d)
	return string(jd)
}

// CompletedDataMaps is not required by pop and may be deleted
type CompletedDataMaps []CompletedDataMap

// String is not required by pop and may be deleted
func (d CompletedDataMaps) String() string {
	jd, _ := json.Marshal(d)
	return string(jd)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (d *CompletedDataMap) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: d.GenesisHash, Name: "GenesisHash"},
		&validators.IntIsGreaterThan{Field: d.ChunkIdx, Name: "ChunkIdx", Compared: -1},
		&validators.StringIsPresent{Field: d.Hash, Name: "Hash"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (d *CompletedDataMap) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (d *CompletedDataMap) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

package models

import (
	"encoding/json"
	"fmt"
	"github.com/gobuffalo/uuid"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

type TransferAddress struct {
	ID          uuid.UUID `json:"id" db:"id"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
	ChunkIdx    int       `json:"chunkIdx" db:"chunk_idx"`
	GenesisHash string    `json:"genesisHash" db:"genesis_hash"`
	Hash        string    `json:"hash" db:"hash"`
	Message     string    `json:"message" db:"message"`
	Address     string    `json:"address" db:"address"`
	Complete    int       `json:"complete" db:"complete"`
}

// String is not required by pop and may be deleted
func (t TransferAddress) String() string {
	jd, _ := json.Marshal(t)
	return string(jd)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) methot.
// This method is not required and may be deletet.
func (t *TransferAddress) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: t.GenesisHash, Name: "GenesisHash"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" methot.
// This method is not required and may be deletet.
func (t *TransferAddress) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" methot.
// This method is not required and may be deletet.
func (t *TransferAddress) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// BeforeCreate runs every time when TransferAddress is createt.
func (t *TransferAddress) BeforeCreate(tx *pop.Connection) error {
	return nil
}

type TransferGenHash struct {
	ID          uuid.UUID `json:"id" db:"id"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
	NumChunks   int       `json:"numChunks" db:"num_chunks"`
	GenesisHash string    `json:"genesisHash" db:"genesis_hash"`
	Complete    int       `json:"complete" db:"complete"`
}

// String is not required by pop and may be deleted
func (t TransferGenHash) String() string {
	jd, _ := json.Marshal(t)
	return string(jd)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) methot.
// This method is not required and may be deletet.
func (t *TransferGenHash) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: t.GenesisHash, Name: "GenesisHash"},
		&validators.IntIsGreaterThan{Field: t.NumChunks, Name: "NumChunks", Compared: -1},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" methot.
// This method is not required and may be deletet.
func (t *TransferGenHash) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" methot.
// This method is not required and may be deletet.
func (t *TransferGenHash) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// BeforeCreate runs every time when TransferGenHash is createt.
func (t *TransferGenHash) BeforeCreate(tx *pop.Connection) error {
	return nil
}

type Temp2CompletedDataMap struct {
	ID             uuid.UUID `json:"id" db:"id"`
	CreatedAt      time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time `json:"updatedAt" db:"updated_at"`
	Status         int       `json:"status" db:"status"`
	NodeID         string    `json:"nodeID" db:"node_id"`
	NodeType       string    `json:"nodeType" db:"node_type"`
	Message        string    `json:"message" db:"message"`
	TrunkTx        string    `json:"trunkTx" db:"trunk_tx"`
	BranchTx       string    `json:"branchTx" db:"branch_tx"`
	GenesisHash    string    `json:"genesisHash" db:"genesis_hash"`
	ChunkIdx       int       `json:"chunkIdx" db:"chunk_idx"`
	Hash           string    `json:"hash" db:"hash"`
	ObfuscatedHash string    `json:"obfuscatedHash" db:"obfuscated_hash"`
	Address        string    `json:"address" db:"address"`
}

func (t2 Temp2CompletedDataMap) TableName() string {
	return "temp_2_completed_data_maps"
}

// String is not required by pop and may be deleted
func (t2 Temp2CompletedDataMap) String() string {
	jd, _ := json.Marshal(t2)
	return string(jd)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) methot.
// This method is not required and may be deletet.
func (t2 *Temp2CompletedDataMap) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: t2.GenesisHash, Name: "GenesisHash"},
		&validators.IntIsGreaterThan{Field: t2.ChunkIdx, Name: "ChunkIdx", Compared: -1},
		&validators.StringIsPresent{Field: t2.Hash, Name: "Hash"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" methot.
// This method is not required and may be deletet.
func (t2 *Temp2CompletedDataMap) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" methot.
// This method is not required and may be deletet.
func (t2 *Temp2CompletedDataMap) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// BeforeCreate runs every time when Temp2CompletedDataMap is createt.
func (t2 *Temp2CompletedDataMap) BeforeCreate(tx *pop.Connection) error {
	return nil
}

func (t2 *Temp2CompletedDataMap) generateMsgId() string {
	return fmt.Sprintf("temp2CompleteDataMap_%v__%d", t2.GenesisHash, t2.ChunkIdx)
}

type Temp3CompletedDataMap struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
	Status    int       `json:"status" db:"status"`
	NodeID    string    `json:"nodeID" db:"node_id"`
	NodeType  string    `json:"nodeType" db:"node_type"`
	Message   string    `json:"message" db:"message"`

	TrunkTx        string `json:"trunkTx" db:"trunk_tx"`
	BranchTx       string `json:"branchTx" db:"branch_tx"`
	GenesisHash    string `json:"genesisHash" db:"genesis_hash"`
	ChunkIdx       int    `json:"chunkIdx" db:"chunk_idx"`
	Hash           string `json:"hash" db:"hash"`
	ObfuscatedHash string `json:"obfuscatedHash" db:"obfuscated_hash"`
	Address        string `json:"address" db:"address"`
}

func (t3 Temp3CompletedDataMap) TableName() string {
	return "temp_3_completed_data_maps"
}

// String is not required by pop and may be deleted
func (t3 Temp3CompletedDataMap) String() string {
	jd, _ := json.Marshal(t3)
	return string(jd)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) methot.
// This method is not required and may be deletet.
func (t3 *Temp3CompletedDataMap) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: t3.GenesisHash, Name: "GenesisHash"},
		&validators.IntIsGreaterThan{Field: t3.ChunkIdx, Name: "ChunkIdx", Compared: -1},
		&validators.StringIsPresent{Field: t3.Hash, Name: "Hash"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" methot.
// This method is not required and may be deletet.
func (t3 *Temp3CompletedDataMap) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" methot.
// This method is not required and may be deletet.
func (t3 *Temp3CompletedDataMap) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// BeforeCreate runs every time when Temp3CompletedDataMap is createt.
func (t3 *Temp3CompletedDataMap) BeforeCreate(tx *pop.Connection) error {
	return nil
}

func (t3 *Temp3CompletedDataMap) generateMsgId() string {
	return fmt.Sprintf("temp3CompleteDataMap_%v__%d", t3.GenesisHash, t3.ChunkIdx)
}

type Temp4CompletedDataMap struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
	Status    int       `json:"status" db:"status"`
	NodeID    string    `json:"nodeID" db:"node_id"`
	NodeType  string    `json:"nodeType" db:"node_type"`
	Message   string    `json:"message" db:"message"`

	TrunkTx        string `json:"trunkTx" db:"trunk_tx"`
	BranchTx       string `json:"branchTx" db:"branch_tx"`
	GenesisHash    string `json:"genesisHash" db:"genesis_hash"`
	ChunkIdx       int    `json:"chunkIdx" db:"chunk_idx"`
	Hash           string `json:"hash" db:"hash"`
	ObfuscatedHash string `json:"obfuscatedHash" db:"obfuscated_hash"`
	Address        string `json:"address" db:"address"`
}

func (t4 Temp4CompletedDataMap) TableName() string {
	return "temp_4_completed_data_maps"
}

// String is not required by pop and may be deleted
func (t4 Temp4CompletedDataMap) String() string {
	jd, _ := json.Marshal(t4)
	return string(jd)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) methot.
// This method is not required and may be deletet.
func (t4 *Temp4CompletedDataMap) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: t4.GenesisHash, Name: "GenesisHash"},
		&validators.IntIsGreaterThan{Field: t4.ChunkIdx, Name: "ChunkIdx", Compared: -1},
		&validators.StringIsPresent{Field: t4.Hash, Name: "Hash"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" methot.
// This method is not required and may be deletet.
func (t4 *Temp4CompletedDataMap) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" methot.
// This method is not required and may be deletet.
func (t4 *Temp4CompletedDataMap) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// BeforeCreate runs every time when Temp4CompletedDataMap is createt.
func (t4 *Temp4CompletedDataMap) BeforeCreate(tx *pop.Connection) error {
	return nil
}

func (t4 *Temp4CompletedDataMap) generateMsgId() string {
	return fmt.Sprintf("temp4CompleteDataMap_%v__%d", t4.GenesisHash, t4.ChunkIdx)
}

type Temp5CompletedDataMap struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
	Status    int       `json:"status" db:"status"`
	NodeID    string    `json:"nodeID" db:"node_id"`
	NodeType  string    `json:"nodeType" db:"node_type"`
	Message   string    `json:"message" db:"message"`

	TrunkTx        string `json:"trunkTx" db:"trunk_tx"`
	BranchTx       string `json:"branchTx" db:"branch_tx"`
	GenesisHash    string `json:"genesisHash" db:"genesis_hash"`
	ChunkIdx       int    `json:"chunkIdx" db:"chunk_idx"`
	Hash           string `json:"hash" db:"hash"`
	ObfuscatedHash string `json:"obfuscatedHash" db:"obfuscated_hash"`
	Address        string `json:"address" db:"address"`
}

func (t5 Temp5CompletedDataMap) TableName() string {
	return "temp_5_completed_data_maps"
}

// String is not required by pop and may be deleted
func (t5 Temp5CompletedDataMap) String() string {
	jd, _ := json.Marshal(t5)
	return string(jd)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) methot.
// This method is not required and may be deletet.
func (t5 *Temp5CompletedDataMap) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: t5.GenesisHash, Name: "GenesisHash"},
		&validators.IntIsGreaterThan{Field: t5.ChunkIdx, Name: "ChunkIdx", Compared: -1},
		&validators.StringIsPresent{Field: t5.Hash, Name: "Hash"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" methot.
// This method is not required and may be deletet.
func (t5 *Temp5CompletedDataMap) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" methot.
// This method is not required and may be deletet.
func (t5 *Temp5CompletedDataMap) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// BeforeCreate runs every time when Temp5CompletedDataMap is createt.
func (t5 *Temp5CompletedDataMap) BeforeCreate(tx *pop.Connection) error {
	return nil
}

func (t5 *Temp5CompletedDataMap) generateMsgId() string {
	return fmt.Sprintf("temp5CompleteDataMap_%v__%d", t5.GenesisHash, t5.ChunkIdx)
}

type Temp6CompletedDataMap struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
	Status    int       `json:"status" db:"status"`
	NodeID    string    `json:"nodeID" db:"node_id"`
	NodeType  string    `json:"nodeType" db:"node_type"`
	Message   string    `json:"message" db:"message"`

	TrunkTx        string `json:"trunkTx" db:"trunk_tx"`
	BranchTx       string `json:"branchTx" db:"branch_tx"`
	GenesisHash    string `json:"genesisHash" db:"genesis_hash"`
	ChunkIdx       int    `json:"chunkIdx" db:"chunk_idx"`
	Hash           string `json:"hash" db:"hash"`
	ObfuscatedHash string `json:"obfuscatedHash" db:"obfuscated_hash"`
	Address        string `json:"address" db:"address"`
}

func (t6 Temp6CompletedDataMap) TableName() string {
	return "temp_6_completed_data_maps"
}

// String is not required by pop and may be deleted
func (t6 Temp6CompletedDataMap) String() string {
	jd, _ := json.Marshal(t6)
	return string(jd)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) methot.
// This method is not required and may be deletet.
func (t6 *Temp6CompletedDataMap) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: t6.GenesisHash, Name: "GenesisHash"},
		&validators.IntIsGreaterThan{Field: t6.ChunkIdx, Name: "ChunkIdx", Compared: -1},
		&validators.StringIsPresent{Field: t6.Hash, Name: "Hash"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" methot.
// This method is not required and may be deletet.
func (t6 *Temp6CompletedDataMap) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" methot.
// This method is not required and may be deletet.
func (t6 *Temp6CompletedDataMap) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// BeforeCreate runs every time when Temp6CompletedDataMap is createt.
func (t6 *Temp6CompletedDataMap) BeforeCreate(tx *pop.Connection) error {
	return nil
}

func (t6 *Temp6CompletedDataMap) generateMsgId() string {
	return fmt.Sprintf("temp6CompleteDataMap_%v__%d", t6.GenesisHash, t6.ChunkIdx)
}

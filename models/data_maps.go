package models

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

var ChunkStatus map[string]int

const fileBytesChunkSize = float64(2817)

type DataMap struct {
	ID          int       `json:"id" db:"id"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
	Status      int       `json:"status" db:"status"`
	HooknodeIP  string    `json:"hooknodeIP" db:"hooknode_ip"`
	Message     string    `json:"message" db:"message"`
	TrunkTx     string    `json:"trunkTx" db:"trunk_tx"`
	BranchTx    string    `json:"branchTx" db:"branch_tx"`
	GenesisHash string    `json:"genesisHash" db:"genesis_hash"`
	ChunkIdx    int       `json:"chunkIdx" db:"chunk_idx"`
	Hash        string    `json:"hash" db:"hash"`
	Address     string    `json:"address" db:"address"`
}

func init() {
	SetChunkStatuses()
}

func SetChunkStatuses() {
	ChunkStatus = make(map[string]int)
	ChunkStatus["pending"] = 0
	ChunkStatus["unassigned"] = 1
	ChunkStatus["unverified"] = 2
	ChunkStatus["complete"] = 3
	ChunkStatus["confirmed"] = 4
	ChunkStatus["error"] = 5
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
		&validators.StringIsPresent{Field: d.GenesisHash, Name: "GenesisHash"},
		&validators.IntIsGreaterThan{Field: d.ChunkIdx, Name: "ChunkIdx", Compared: -1},
		&validators.StringIsPresent{Field: d.Hash, Name: "Hash"},
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

// BuildDataMaps builds the datamap and inserts them into the DB.
func BuildDataMaps(genHash string, fileBytesCount int) (vErr *validate.Errors, err error) {
	fileChunksCount := int(math.Ceil(float64(fileBytesCount) / fileBytesChunkSize))

	currHash := genHash
	for i := 0; i <= fileChunksCount; i++ {
		// TODO: Batch these inserts.
		vErr, err = DB.ValidateAndCreate(&DataMap{
			GenesisHash: genHash,
			ChunkIdx:    i,
			Hash:        currHash,
		})

		currHash = hashString(currHash)
	}

	return
}

func hashString(str string) (h string) {
	shaHash := sha256.New()
	shaHash.Write([]byte(str))
	h = hex.EncodeToString(shaHash.Sum(nil))
	return
}

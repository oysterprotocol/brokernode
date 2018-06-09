package models

import (
	"encoding/json"
	"errors"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"github.com/oysterprotocol/brokernode/utils"
	"time"
)

const (
	StoredGenesisHashUnassigned int = iota + 1
	StoredGenesisHashAssigned
)

const WebnodeCountLimit = 2

type StoredGenesisHash struct {
	ID            uuid.UUID `json:"id" db:"id"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
	GenesisHash   string    `json:"genesisHash" db:"genesis_hash"`
	FileSizeBytes int       `json:"fileSizeBytes" db:"file_size_bytes"`
	NumChunks     int       `json:"numChunks" db:"num_chunks"`
	WebnodeCount  int       `json:"webnodeCount" db:"webnode_count"`
	Status        int       `json:"status" db:"status"`
}

// String is not required by pop and may be deleted
func (s StoredGenesisHash) String() string {
	js, _ := json.Marshal(s)
	return string(js)
}

// StoredGenesisHashes is not required by pop and may be deleted
type StoredGenesisHashes []StoredGenesisHash

// String is not required by pop and may be deleted
func (s StoredGenesisHashes) String() string {
	js, _ := json.Marshal(s)
	return string(js)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (s *StoredGenesisHash) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (s *StoredGenesisHash) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (s *StoredGenesisHash) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

func (s *StoredGenesisHash) BeforeCreate(tx *pop.Connection) error {

	// Defaults to StoredGenesisHashUnassigned
	if s.Status == 0 {
		s.Status = StoredGenesisHashUnassigned
	}

	return nil
}

//TODO: unit test this method
func GetGenesisHashForWebnode(existingGenesisHashes []string) (StoredGenesisHash, error) {
	//existingGenesisHashes are genesis hashes that the webnode already has
	storedGenesisHashes := []StoredGenesisHash{}

	existingGenHashMap := make(map[string]bool)
	for _, genHash := range existingGenesisHashes {
		existingGenHashMap[genHash] = true
	}

	err := DB.Where("webnode_count < ? AND status = ?",
		WebnodeCountLimit, StoredGenesisHashUnassigned).All(&storedGenesisHashes)

	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return StoredGenesisHash{}, err
	}

	for _, storedGenHash := range storedGenesisHashes {
		if _, ok := existingGenHashMap[storedGenHash.GenesisHash]; !ok {
			return storedGenHash, nil
		}
	}

	return StoredGenesisHash{}, errors.New("no genesis hashes to sell")
}

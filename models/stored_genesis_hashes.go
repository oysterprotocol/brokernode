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

const (
	TreasurePending int = iota + 1
	TreasureBuried
)

const (
	WebnodeCountLimit  = 2
	NoGenHashesMessage = "no genesis hashes to sell, or none that this webnode needs"
)

const (
	/*
		HoursToAssumeTreasureHasBeenBuried is the number of hours after which we can safely assume the treasure
		has been buried.  This is needed in case the webnode is transacting with the beta broker but the alpha
		did the treasure burying, or vice versa.
	*/
	HoursToAssumeTreasureHasBeenBuried = 8
)

type StoredGenesisHash struct {
	ID             uuid.UUID `json:"id" db:"id"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
	GenesisHash    string    `json:"genesisHash" db:"genesis_hash"`
	FileSizeBytes  uint64    `json:"fileSizeBytes" db:"file_size_bytes"`
	NumChunks      int       `json:"numChunks" db:"num_chunks"`
	WebnodeCount   int       `json:"webnodeCount" db:"webnode_count"`
	Status         int       `json:"status" db:"status"`
	TreasureStatus int       `json:"treasureStatus" db:"treasure_status"`
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

	if s.TreasureStatus == 0 {
		s.TreasureStatus = TreasurePending
	}

	return nil
}

func GetGenesisHashForWebnode(existingGenesisHashes []string) (StoredGenesisHash, error) {
	//existingGenesisHashes are genesis hashes that the webnode already has
	treasureBuriedGenesisHashes := []StoredGenesisHash{}
	treasureLikelyBuriedGenesisHashes := []StoredGenesisHash{}

	existingGenHashMap := make(map[string]bool)
	for _, genHash := range existingGenesisHashes {
		existingGenHashMap[genHash] = true
	}

	err := DB.Where("webnode_count < ? AND status = ? AND treasure_status = ? ORDER BY created_at asc",
		WebnodeCountLimit, StoredGenesisHashUnassigned, TreasureBuried).All(&treasureBuriedGenesisHashes)
	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return StoredGenesisHash{}, err
	}

	err = DB.Where("webnode_count < ? AND status = ?  AND TIMESTAMPDIFF(hour, created_at, NOW()) >= ? "+
		"ORDER BY created_at asc",
		WebnodeCountLimit, StoredGenesisHashUnassigned,
		HoursToAssumeTreasureHasBeenBuried).All(&treasureLikelyBuriedGenesisHashes)
	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return StoredGenesisHash{}, err
	}

	treasureBuriedGenesisHashes = append(treasureBuriedGenesisHashes, treasureLikelyBuriedGenesisHashes...)

	for _, storedGenHash := range treasureBuriedGenesisHashes {
		if _, ok := existingGenHashMap[storedGenHash.GenesisHash]; !ok {
			return storedGenHash, nil
		}
	}

	return StoredGenesisHash{}, errors.New(NoGenHashesMessage)
}

func SetToTreasureBuriedByGenesisHash(genesisHash string) error {
	err := DB.RawQuery("UPDATE stored_genesis_hashes SET treasure_status = ? WHERE genesis_hash = ?",
		TreasureBuried, genesisHash).All(&[]StoredGenesisHash{})
	return err
}

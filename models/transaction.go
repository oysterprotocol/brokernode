package models

import (
	"encoding/json"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"time"
)

const (
	TransactionTypeBrokernode int = iota + 1
	TransactionTypeWebnode
	TransactionTypeGenesisHash
)

const (
	TransactionStatusPending int = iota + 1
	TransactionStatusComplete
)

type Transaction struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	Status    int       `json:"status" db:"status"`
	Type      int       `json:"type" db:"type"`
	DataMapID uuid.UUID `json:"data_map_id" db:"data_map_id"`
	Purchase  string    `json:"purchase" db:"purchase"`
}

// String is not required by pop and may be deleted
func (t Transaction) String() string {
	jt, _ := json.Marshal(t)
	return string(jt)
}

// Transactions is not required by pop and may be deleted
type Transactions []Transaction

// String is not required by pop and may be deleted
func (t Transactions) String() string {
	jt, _ := json.Marshal(t)
	return string(jt)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (t *Transaction) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.IntIsPresent{Field: t.Status, Name: "Status"},
		&validators.IntIsPresent{Field: t.Type, Name: "Type"},
		&validators.StringIsPresent{Field: t.Purchase, Name: "Purchase"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (t *Transaction) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (t *Transaction) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

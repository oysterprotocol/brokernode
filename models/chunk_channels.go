package models

import (
	"encoding/json"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"time"
)

/*
Intended to replace hooknodes table until we add hooknodes back in
 */

type ChunkChannel struct {
	ID              uuid.UUID `json:"id" db:"id"`
	ChannelID       string    `json:"channel_id" db:"channel_id"`
	ChunksProcessed int       `json:"chunks_processed" db:"chunks_processed"`
	EstReadyTime    time.Time `json:"est_ready_time" db:"est_ready_time"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// String is not required by pop and may be deleted
func (c ChunkChannel) String() string {
	jc, _ := json.Marshal(c)
	return string(jc)
}

// ChunkChannels is not required by pop and may be deleted
type ChunkChannels []ChunkChannel

// String is not required by pop and may be deleted
func (c ChunkChannels) String() string {
	jc, _ := json.Marshal(c)
	return string(jc)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (c *ChunkChannel) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (c *ChunkChannel) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (c *ChunkChannel) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

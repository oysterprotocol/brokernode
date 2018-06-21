package models

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"github.com/oysterprotocol/brokernode/utils"
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

var (
	letters = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	wg      sync.WaitGroup
)

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

// GetReadyChannels grabs all of the channels that are ready
func GetReadyChannels() ([]ChunkChannel, error) {

	channels := []ChunkChannel{}

	err := DB.RawQuery("SELECT * from chunk_channels WHERE "+
		"est_ready_time <= ? ORDER BY est_ready_time;", time.Now()).All(&channels)

	oyster_utils.LogIfError(err, nil)

	return channels, err
}

// GetReadyChannels grabs one ready channel
func GetOneReadyChannel() (ChunkChannel, error) {

	channel := ChunkChannel{}

	err := DB.RawQuery("SELECT * from chunk_channels WHERE "+
		"est_ready_time <= ? ORDER BY est_ready_time;", time.Now()).First(&channel)
	oyster_utils.LogIfError(err, nil)
	return channel, err
}

func MakeChannels(powProcs int) ([]ChunkChannel, error) {

	wg.Add(1)
	var err error

	go func(err *error) {
		defer wg.Done()
		*err = DB.Transaction(func(DB *pop.Connection) error {
			err := DB.RawQuery("DELETE from chunk_channels;").All(&[]ChunkChannel{})
			if err != nil {
				oyster_utils.LogIfError(err, nil)
				return err
			}

			for i := 0; i < powProcs; i++ {

				var err error
				channel := ChunkChannel{}
				channel.ChannelID = oyster_utils.RandSeq(10, letters)
				channel.EstReadyTime = time.Now().Add(-5 * time.Second)
				channel.ChunksProcessed = 0

				_, err = DB.ValidateAndSave(&channel)
				if err != nil {
					oyster_utils.LogIfError(err, nil)
					return err
				}
			}

			return nil
		})
	}(&err)

	wg.Wait()

	oyster_utils.LogIfError(err, nil)

	channels := []ChunkChannel{}

	err = DB.RawQuery("SELECT * from chunk_channels;").All(&channels)
	oyster_utils.LogIfError(err, nil)

	return channels, err
}

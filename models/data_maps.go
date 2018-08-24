package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/oysterprotocol/brokernode/utils"
)

const (
	/*DataMapTableName is the name of the data maps table in SQL*/
	DataMapTableName = "data_maps"

	// The max number of values to insert to db via Sql: INSERT INTO table_name VALUES.
	MaxNumberOfValueForInsertOperation = 10
)

const (
	Pending int = iota + 1
	Unassigned
	Unverified
	Complete
	Confirmed
	Error = -1
)

const (
	// Default value before adding msg_status column.
	MsgStatusUnmigrated int = iota
	// When client does not upload any data to brokernode.
	MsgStatusNotUploaded
	// When client has upload the data chunk to the brokernode without encoding it.
	MsgStatusUploadedHaveNotEncoded
	// When client does not need to encode upload data chunk.
	MsgStatusUploadedNoNeedEncode
)

type DataMap struct {
	ID             uuid.UUID `json:"id" db:"id"`
	CreatedAt      time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time `json:"updatedAt" db:"updated_at"`
	Status         int       `json:"status" db:"status"`
	NodeID         string    `json:"nodeID" db:"node_id"`
	NodeType       string    `json:"nodeType" db:"node_type"`
	Message        string    `json:"message" db:"message"`
	MsgID          string    `json:"msgId" db:"msg_id"`
	MsgStatus      int       `json:"msgStatus" db:"msg_status"`
	TrunkTx        string    `json:"trunkTx" db:"trunk_tx"`
	BranchTx       string    `json:"branchTx" db:"branch_tx"`
	GenesisHash    string    `json:"genesisHash" db:"genesis_hash"`
	ChunkIdx       int       `json:"chunkIdx" db:"chunk_idx"`
	Hash           string    `json:"hash" db:"hash"`
	ObfuscatedHash string    `json:"obfuscatedHash" db:"obfuscated_hash"`
	Address        string    `json:"address" db:"address"`
}

type TypeAndChunkMap struct {
	Type   int       `json:"type"`
	Chunks []DataMap `json:"chunks"`
}

var (
	/*MsgStatusMap is for pretty printing in the grifts and elsewhere*/
	MsgStatusMap = make(map[int]string)
	/*StatusMap is for pretty printing in the grifts and elsewhere*/
	StatusMap = make(map[int]string)
)

func init() {
	StatusMap[Pending] = "Pending"
	StatusMap[Unassigned] = "Unassigned"
	StatusMap[Unverified] = "Unverified"
	StatusMap[Complete] = "Complete"
	StatusMap[Confirmed] = "Confirmed"
	StatusMap[Error] = "Error"

	MsgStatusMap[MsgStatusUnmigrated] = "MsgStatusUnmigrated"
	MsgStatusMap[MsgStatusNotUploaded] = "MsgStatusNotUploaded"
	MsgStatusMap[MsgStatusUploadedHaveNotEncoded] = "MsgStatusUploadedHaveNotEncoded"
	MsgStatusMap[MsgStatusUploadedNoNeedEncode] = "MsgStatusUploadedNoNeedEncode"
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
		&validators.StringIsPresent{Field: d.MsgID, Name: "MsgID"},
		&validators.IntIsGreaterThan{Field: d.MsgStatus, Name: "MsgStatus", Compared: MsgStatusUnmigrated},
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

/**
 * Callbacks
 */

func (d *DataMap) BeforeCreate(tx *pop.Connection) error {
	d.MsgID = d.generateMsgId()

	// After adding MsgStatus column, all previous value will be 0/MsgStatusUnmigrated.
	// And all new value should be default to MsgStatusNotUploaded.
	if d.MsgStatus == MsgStatusUnmigrated {
		d.MsgStatus = MsgStatusNotUploaded
	}
	return nil
}

func (d *DataMap) generateMsgId() string {
	return oyster_utils.GenerateBadgerKey("", d.GenesisHash, d.ChunkIdx)
}

func insertsIntoDataMapsTable(columnsName string, values string, valueSize int) error {
	if len(values) == 0 {
		return nil
	}

	rawQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", DataMapTableName, columnsName, values)
	var err error
	for i := 0; i < oyster_utils.MAX_NUMBER_OF_SQL_RETRY; i++ {
		err = DB.RawQuery(rawQuery).All(&[]DataMap{})
		if err == nil {
			break
		}
	}
	oyster_utils.LogIfError(err, map[string]interface{}{"MaxRetry": oyster_utils.MAX_NUMBER_OF_SQL_RETRY, "NumOfRecord": valueSize})
	return err
}

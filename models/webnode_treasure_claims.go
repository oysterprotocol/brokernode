package models

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gobuffalo/pop"
	"github.com/oysterprotocol/brokernode/utils"
	"golang.org/x/crypto/sha3"
	"math/big"
	"time"

	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

type WebnodeTreasureClaim struct {
	ID                    uuid.UUID         `json:"id" db:"id"`
	CreatedAt             time.Time         `json:"createdAt" db:"created_at"`
	UpdatedAt             time.Time         `json:"updatedAt" db:"updated_at"`
	GenesisHash           string            `json:"genesisHash" db:"genesis_hash"`
	SectorIdx             int               `json:"sectorIdx" db:"sector_idx"`
	NumChunks             int               `json:"numChunks" db:"num_chunks"`
	ReceiverETHAddr       string            `json:"receiverEthAddr" db:"receiver_eth_addr"`
	TreasureETHAddr       string            `json:"treasureEthAddr" db:"treasure_eth_addr"`
	TreasureETHPrivateKey string            `json:"treasureEthPrivateKey" db:"treasure_eth_private_key"`
	StartingClaimClock    int64             `json:"startingClaimClock" db:"starting_claim_clock"`
	ClaimPRLStatus        PRLClaimStatus    `json:"claimPrlStatus" db:"claim_prl_status"`
	ClaimPRLTxHash        string            `json:"claimPrlTxHash" db:"claim_prl_tx_hash"`
	ClaimPRLTxNonce       int64             `json:"claimPrlTxNonce" db:"claim_prl_tx_nonce"`
	GasStatus             GasTransferStatus `json:"gasStatus" db:"gas_status"`
	GasTxHash             string            `json:"gasTxHash" db:"gas_tx_hash"`
	GasTxNonce            int64             `json:"gasTxNonce" db:"gas_tx_nonce"`
}

// String is not required by pop and may be deleted
func (w WebnodeTreasureClaim) String() string {
	jc, _ := json.Marshal(w)
	return string(jc)
}

/**
 * Validations
 */

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (w *WebnodeTreasureClaim) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: w.GenesisHash, Name: "GenesisHash"},
		&validators.StringIsPresent{Field: w.ReceiverETHAddr, Name: "ReceiverETHAddr"},
		&validators.StringIsPresent{Field: w.TreasureETHAddr, Name: "TreasureETHAddr"},
		&validators.StringIsPresent{Field: w.TreasureETHPrivateKey, Name: "TreasureETHPrivateKey"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (w *WebnodeTreasureClaim) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (w *WebnodeTreasureClaim) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

/**
 * Callbacks
 */

func (w *WebnodeTreasureClaim) BeforeCreate(tx *pop.Connection) error {
	// Defaults to PRLClaimNotStarted.
	if w.ClaimPRLStatus == 0 {
		w.ClaimPRLStatus = PRLClaimNotStarted
	}

	// Defaults to GasTransferNotStarted.
	if w.GasStatus == 0 {
		w.GasStatus = GasTransferNotStarted
	}

	if w.StartingClaimClock == 0 {
		startingClock := big.NewInt(-1)
		w.StartingClaimClock = startingClock.Int64()
	}

	return nil
}

func (w *WebnodeTreasureClaim) AfterCreate(tx *pop.Connection) error {

	w.EncryptTreasureEthKey()

	return nil
}

func (w *WebnodeTreasureClaim) EncryptTreasureEthKey() {
	hashedSessionID := oyster_utils.HashHex(hex.EncodeToString([]byte(fmt.Sprint(w.ID))), sha3.New256())
	hashedCreationTime := oyster_utils.HashHex(hex.EncodeToString([]byte(fmt.Sprint(w.CreatedAt.Clock()))), sha3.New256())

	encryptedKey := oyster_utils.Encrypt(hashedSessionID, w.TreasureETHPrivateKey, hashedCreationTime)

	w.TreasureETHPrivateKey = hex.EncodeToString(encryptedKey)
	DB.ValidateAndSave(w)
}

func (w *WebnodeTreasureClaim) DecryptTreasureEthKey() string {
	hashedSessionID := oyster_utils.HashHex(hex.EncodeToString([]byte(fmt.Sprint(w.ID))), sha3.New256())
	hashedCreationTime := oyster_utils.HashHex(hex.EncodeToString([]byte(fmt.Sprint(w.CreatedAt.Clock()))), sha3.New256())

	decryptedKey := oyster_utils.Decrypt(hashedSessionID, w.TreasureETHPrivateKey, hashedCreationTime)

	return hex.EncodeToString(decryptedKey)
}

func GetTreasureClaimsByGasAndPRLStatus(gasStatus GasTransferStatus, prlStatus PRLClaimStatus) (treasureClaims []WebnodeTreasureClaim, err error) {
	err = DB.Where("gas_status = ? AND claim_prl_status = ?", gasStatus, prlStatus).All(&treasureClaims)
	oyster_utils.LogIfError(err, nil)
	return treasureClaims, err
}

func GetTreasureClaimsByGasStatus(gasStatus GasTransferStatus) (treasureClaims []WebnodeTreasureClaim, err error) {
	err = DB.Where("gas_status = ?", gasStatus).All(&treasureClaims)
	oyster_utils.LogIfError(err, nil)

	return treasureClaims, err
}

func GetTreasureClaimsByPRLStatus(prlStatus PRLClaimStatus) (treasureClaims []WebnodeTreasureClaim, err error) {
	err = DB.Where("claim_prl_status = ?", prlStatus).All(&treasureClaims)
	oyster_utils.LogIfError(err, nil)

	return treasureClaims, err
}

func GetTreasureClaimsWithTimedOutGasTransfers(thresholdTime time.Time) (treasureClaims []WebnodeTreasureClaim, err error) {
	err = DB.Where("gas_status = ? AND updated_at <= ?",
		GasTransferProcessing,
		thresholdTime).All(&treasureClaims)
	oyster_utils.LogIfError(err, nil)

	return treasureClaims, err
}

func GetTreasureClaimsWithTimedOutPRLClaims(thresholdTime time.Time) (treasureClaims []WebnodeTreasureClaim, err error) {
	err = DB.Where("claim_prl_status = ? AND updated_at <= ?",
		PRLClaimProcessing,
		thresholdTime).All(&treasureClaims)
	oyster_utils.LogIfError(err, nil)

	return treasureClaims, err
}

func GetTreasureClaimsWithTimedOutGasReclaims(thresholdTime time.Time) (treasureClaims []WebnodeTreasureClaim, err error) {
	err = DB.Where("gas_status = ? AND updated_at <= ?",
		GasTransferLeftoversReclaimProcessing,
		thresholdTime).All(&treasureClaims)
	oyster_utils.LogIfError(err, nil)

	return treasureClaims, err
}

func DeleteCompletedTreasureClaims() error {
	err := DB.RawQuery("DELETE from webnode_treasure_claims WHERE gas_status = ?",
		GasTransferLeftoversReclaimSuccess).All(&[]WebnodeTreasureClaim{})
	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return err
	}

	return nil
}

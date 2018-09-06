package models

import (
	"encoding/json"
	"math/big"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"github.com/oysterprotocol/brokernode/utils"
	"github.com/shopspring/decimal"
)

/* BrokerBrokerTransaction defines the model for the table which will deal with broker to broker transactions */
type BrokerBrokerTransaction struct {
	ID            uuid.UUID       `json:"id" db:"id"`
	GenesisHash   string          `json:"genesisHash" db:"genesis_hash"`
	CreatedAt     time.Time       `json:"createdAt" db:"created_at"`
	UpdatedAt     time.Time       `json:"updatedAt" db:"updated_at"`
	Type          int             `json:"type" db:"type"`
	ETHAddrAlpha  string          `json:"ethAddrAlpha" db:"eth_addr_alpha"`
	ETHAddrBeta   string          `json:"ethAddrBeta" db:"eth_addr_beta"`
	ETHPrivateKey string          `json:"ethPrivateKey" db:"eth_private_key"`
	TotalCost     decimal.Decimal `json:"totalCost" db:"total_cost"`
	PaymentStatus PaymentStatus   `json:"paymentStatus" db:"payment_status"`
}

/* Payment status will hold the status of the payment of the broker_broker_transaction row */
type PaymentStatus int

const (
	BrokerTxAlphaPaymentPending PaymentStatus = iota + 1
	BrokerTxAlphaPaymentConfirmed
	BrokerTxGasPaymentPending
	BrokerTxGasPaymentConfirmed
	BrokerTxBetaPaymentPending
	BrokerTxBetaPaymentConfirmed

	/* These error statuses assigned these ints so we can multiply by -1 to
	set back to the previous state in the sequence, for retrying
	*/
	BrokerTxAlphaPaymentError PaymentStatus = -1
	BrokerTxGasPaymentError   PaymentStatus = -2
	BrokerTxBetaPaymentError  PaymentStatus = -4
)

/* PaymentStatusMap is used for pretty printing the payment statuses */
var PaymentStatusMap = make(map[PaymentStatus]string)

func init() {
	PaymentStatusMap[BrokerTxAlphaPaymentPending] = "BrokerTxAlphaPaymentPending"
	PaymentStatusMap[BrokerTxAlphaPaymentConfirmed] = "BrokerTxAlphaPaymentConfirmed"
	PaymentStatusMap[BrokerTxGasPaymentPending] = "BrokerTxGasPaymentPending"
	PaymentStatusMap[BrokerTxGasPaymentConfirmed] = "BrokerTxGasPaymentConfirmed"
	PaymentStatusMap[BrokerTxGasPaymentError] = "BrokerTxGasPaymentError"
	PaymentStatusMap[BrokerTxBetaPaymentPending] = "BrokerTxBetaPaymentPending"
	PaymentStatusMap[BrokerTxBetaPaymentConfirmed] = "BrokerTxBetaPaymentConfirmed"

	PaymentStatusMap[BrokerTxAlphaPaymentError] = "BrokerTxAlphaPaymentError"
	PaymentStatusMap[BrokerTxGasPaymentError] = "BrokerTxGasPaymentError"
	PaymentStatusMap[BrokerTxBetaPaymentError] = "BrokerTxBetaPaymentError"
}

// String is not required by pop and may be deleted
func (b BrokerBrokerTransaction) String() string {
	ju, _ := json.Marshal(b)
	return string(ju)
}

/**
 * Validations
 */

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (b *BrokerBrokerTransaction) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (b *BrokerBrokerTransaction) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (b *BrokerBrokerTransaction) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

/**
 * Callbacks
 */
func (b *BrokerBrokerTransaction) BeforeCreate(tx *pop.Connection) error {
	// Defaults to alpha session.
	if b.Type != SessionTypeBeta {
		b.Type = SessionTypeAlpha
	}

	switch oyster_utils.BrokerMode {
	case oyster_utils.ProdMode:
		// Defaults to BrokerTxAlphaPaymentPending unless oyster is paying
		if b.PaymentStatus == 0 {
			if oyster_utils.PaymentMode == oyster_utils.UserIsPaying {
				b.PaymentStatus = BrokerTxAlphaPaymentPending
			} else {
				b.PaymentStatus = BrokerTxBetaPaymentConfirmed
			}
		}
	default:
		// Defaults to BrokerTxBetaPaymentConfirmed
		if b.PaymentStatus == 0 {
			b.PaymentStatus = BrokerTxBetaPaymentConfirmed
		}
	}

	return nil
}

func (b *BrokerBrokerTransaction) AfterCreate(tx *pop.Connection) error {
	if b.Type == SessionTypeAlpha {
		b.EncryptEthKey()
	}

	return nil
}

/**
 * Methods
 */

/*EncryptEthKey encrypts the private eth key*/
func (b *BrokerBrokerTransaction) EncryptEthKey() {
	b.ETHPrivateKey = oyster_utils.ReturnEncryptedEthKey(b.ID, b.CreatedAt, b.ETHPrivateKey)
	DB.ValidateAndSave(b)
}

/*DecryptEthKey decrypts the private eth key*/
func (b *BrokerBrokerTransaction) DecryptEthKey() string {
	return oyster_utils.ReturnDecryptedEthKey(b.ID, b.CreatedAt, b.ETHPrivateKey)
}

/*NewBrokerBrokerTransaction creates a new broker_broker_transaction that corresponds to a session */
func NewBrokerBrokerTransaction(session *UploadSession) bool {

	if oyster_utils.BrokerMode != oyster_utils.ProdMode {
		return false
	}

	var paymentStatus PaymentStatus

	switch session.PaymentStatus {
	case PaymentStatusInvoiced:
		paymentStatus = BrokerTxAlphaPaymentPending
	case PaymentStatusPending:
		paymentStatus = BrokerTxAlphaPaymentPending
	case PaymentStatusConfirmed:
		if oyster_utils.PaymentMode == oyster_utils.UserIsPaying {
			paymentStatus = BrokerTxAlphaPaymentConfirmed
		} else {
			paymentStatus = BrokerTxBetaPaymentConfirmed
		}
	case PaymentStatusError:
		paymentStatus = BrokerTxAlphaPaymentError
	default:
		paymentStatus = BrokerTxAlphaPaymentPending
	}

	privateKey := session.DecryptSessionEthKey()

	brokerTx := BrokerBrokerTransaction{
		GenesisHash:   session.GenesisHash,
		Type:          session.Type,
		ETHAddrAlpha:  session.ETHAddrAlpha.String,
		ETHAddrBeta:   session.ETHAddrBeta.String,
		ETHPrivateKey: privateKey,
		TotalCost:     session.TotalCost,
		PaymentStatus: paymentStatus,
	}

	vErr, err := DB.ValidateAndCreate(&brokerTx)
	oyster_utils.LogIfError(err, nil)
	oyster_utils.LogIfValidationError("BrokerBrokerTransaction validation failed", vErr, nil)

	return err == nil && len(vErr.Errors) == 0
}

/*GetTotalCostInWei takes the TotalCost and converts it to wei units*/
func (b *BrokerBrokerTransaction) GetTotalCostInWei() *big.Int {
	float64Cost, _ := b.TotalCost.Float64()
	return oyster_utils.ConvertToWeiUnit(big.NewFloat(float64Cost))
}

/*GetTransactionsBySessionTypesAndPaymentStatuses accepts an array of session types and payment statuses and returns
broker_broker_transactions that match*/
func GetTransactionsBySessionTypesAndPaymentStatuses(sessionTypes []int, paymentStatuses []PaymentStatus) ([]BrokerBrokerTransaction, error) {

	txsToReturn := make([]BrokerBrokerTransaction, 0)

	if len(sessionTypes) == 2 || len(sessionTypes) == 0 {
		for _, paymentStatus := range paymentStatuses {
			brokerTxs := []BrokerBrokerTransaction{}

			err := DB.RawQuery("SELECT * FROM broker_broker_transactions WHERE payment_status = ?",
				paymentStatus).All(&brokerTxs)

			if err != nil {
				oyster_utils.LogIfError(err, nil)
				return brokerTxs, err
			}
			txsToReturn = append(txsToReturn, brokerTxs...)
		}
	} else if len(sessionTypes) == 1 {
		for _, paymentStatus := range paymentStatuses {
			brokerTxs := []BrokerBrokerTransaction{}

			err := DB.RawQuery("SELECT * FROM broker_broker_transactions WHERE type = ? AND "+
				"payment_status = ?",
				sessionTypes[0],
				paymentStatus).All(&brokerTxs)

			if err != nil {
				oyster_utils.LogIfError(err, nil)
				return brokerTxs, err
			}
			txsToReturn = append(txsToReturn, brokerTxs...)
		}
	}
	return txsToReturn, nil
}

/*GetTransactionsBySessionTypesPaymentStatusesAndTime accepts an array of session types, an array of payment statuses,
and a time threshold and returns broker_broker_transactions that match*/
func GetTransactionsBySessionTypesPaymentStatusesAndTime(sessionTypes []int, paymentStatuses []PaymentStatus,
	thresholdTime time.Time) ([]BrokerBrokerTransaction, error) {

	txsToReturn := make([]BrokerBrokerTransaction, 0)

	if len(sessionTypes) == 2 || len(sessionTypes) == 0 {
		for _, paymentStatus := range paymentStatuses {
			brokerTxs := []BrokerBrokerTransaction{}

			err := DB.RawQuery("SELECT * FROM broker_broker_transactions WHERE payment_status = ? "+
				"AND updated_at <= ?",
				paymentStatus,
				thresholdTime).All(&brokerTxs)

			if err != nil {
				oyster_utils.LogIfError(err, nil)
				return brokerTxs, err
			}
			txsToReturn = append(txsToReturn, brokerTxs...)
		}
	} else if len(sessionTypes) == 1 {
		for _, paymentStatus := range paymentStatuses {
			brokerTxs := []BrokerBrokerTransaction{}

			err := DB.RawQuery("SELECT * FROM broker_broker_transactions WHERE type = ? AND "+
				"payment_status = ? AND updated_at <= ?",
				sessionTypes[0],
				paymentStatus,
				thresholdTime).All(&brokerTxs)

			if err != nil {
				oyster_utils.LogIfError(err, nil)
				return brokerTxs, err
			}
			txsToReturn = append(txsToReturn, brokerTxs...)
		}
	}
	return txsToReturn, nil
}

/* SetUploadSessionToPaid will find the upload_session that corresponds to the broker_broker_transaction that has just
been paid and set it to paid */
func SetUploadSessionToPaid(brokerTx BrokerBrokerTransaction) error {
	err := DB.RawQuery("UPDATE upload_sessions set payment_status = ? WHERE "+
		"payment_status != ? AND genesis_hash = ?",
		PaymentStatusConfirmed,
		PaymentStatusConfirmed,
		brokerTx.GenesisHash).All(&[]UploadSession{})

	oyster_utils.LogIfError(err, nil)
	return err
}

func (brokerTx *BrokerBrokerTransaction) GetMetaChunk() (DataMap, error) {
	metaIdx := 0 // NOTE: This will change with rev2.

	metaChunk := DataMap{}
	err := DB.
		Where("genesis_hash = ? AND chunk_idx = ?", brokerTx.GenesisHash, metaIdx).
		First(&metaChunk)
	oyster_utils.LogIfError(err, nil)

	// TODO: Check if msg is empty, and throw error.
	return metaChunk, err
}

/* DeleteCompletedBrokerTransactions deletes any brokerTxs for which both alpha and beta are paid */
func DeleteCompletedBrokerTransactions() {
	err := DB.RawQuery("DELETE FROM broker_broker_transactions WHERE "+
		"payment_status = ?",
		BrokerTxBetaPaymentConfirmed,
	).All(&[]BrokerBrokerTransaction{})

	oyster_utils.LogIfError(err, nil)
}

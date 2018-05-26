package jobs

import (
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/utils"
	"github.com/gobuffalo/pop"
	"github.com/oysterprotocol/brokernode/services"
)

// Allows client to override this value. For testing.
var unpaidExpirationInHour = 24

func init() {
	SetUnpaidExpiration(0, true)
}

/*RemoveUnpaidUploadSession cleans up unpload_sessions and data_maps talbe for expired/unpaid session. */
func RemoveUnpaidUploadSession() {
	sessions := []models.UploadSession{}
	err := models.DB.RawQuery("SELECT * from upload_sessions WHERE payment_status != ?", models.PaymentStatusConfirmed).All(&sessions)
	if err != nil {
		oyster_utils.LogIfError(err)
		return
	}

	for _, session := range sessions {
		balance := EthWrapper.CheckBalance(services.StringToAddress(session.ETHAddrAlpha.String))
		if balance.Int64() > 0 {
			continue
		}

		oyster_utils.LogIfError(models.DB.Transaction(func(tx *pop.Connection) error {
			if err := tx.RawQuery("DELETE from data_maps WHERE genesis_hash = ?", sessions.GenesisHash).All(&[]models.DataMaps{});
				 err != nil {
				return nil
			}
			if err := tx.RawQuery("DELETE from upload_sessions WHERE id = ?", session.ID).All(&[]models.UploadSession{});
				 err != nil {
				return err
			}
			return nil			
		})
	}
}

/*SetUnpaidExpiration allows caller to control how long the expiration duration. */
func SetUnpaidExpiration(hour int, setToDefault bool) {
	if setToDefault {
		unpaidExpirationInHour = 24
	} else {
		unpaidExpirationInHour = hour
	}
}

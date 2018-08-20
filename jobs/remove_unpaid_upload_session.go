package jobs

import (
	"errors"
	"github.com/gobuffalo/pop"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
)

/*UnpaidExpirationInHour means number of hours before it should remove unpaid upload session. */
const UnpaidExpirationInHour = 24

/*RemoveUnpaidUploadSession cleans up upload_sessions and data_maps table for expired/unpaid session. */
func RemoveUnpaidUploadSession(PrometheusWrapper services.PrometheusService) {

	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramRemoveUnpaidUploadSession, start)

	sessions := []models.UploadSession{}
	err := models.DB.RawQuery(
		"SELECT * FROM upload_sessions WHERE payment_status != ? AND TIMESTAMPDIFF(hour, updated_at, NOW()) >= ?",
		models.PaymentStatusConfirmed, UnpaidExpirationInHour).All(&sessions)
	if err != nil {
		oyster_utils.LogIfError(errors.New(err.Error()+" while finding old unpaid sessions in "+
			"remove_unpaid_upload_session"), nil)
		return
	}

	for _, session := range sessions {
		balance := EthWrapper.CheckPRLBalance(services.StringToAddress(session.ETHAddrAlpha.String))
		if balance.Int64() > 0 {
			continue
		}

		dbNameMessage := oyster_utils.GetBadgerDBName([]string{oyster_utils.InProgressDir,
			session.GenesisHash, oyster_utils.MessageDir})
		dbNameHash := oyster_utils.GetBadgerDBName([]string{oyster_utils.InProgressDir,
			session.GenesisHash, oyster_utils.HashDir})

		errMessage := oyster_utils.CloseUniqueKvStore(dbNameMessage)
		if errMessage != nil {
			oyster_utils.LogIfError(errors.New(errMessage.Error()+" deleting message data from "+
				dbNameMessage+" in RemoveUnpaidUploadSession"), nil)
			continue
		}
		errHash := oyster_utils.CloseUniqueKvStore(dbNameHash)
		if errHash != nil {
			oyster_utils.LogIfError(errors.New(errHash.Error()+" deleting hash data from "+
				dbNameHash+" in RemoveUnpaidUploadSession"), nil)
			continue
		}

		err := models.DB.Transaction(func(tx *pop.Connection) error {
			return tx.RawQuery("DELETE FROM upload_sessions WHERE id = ?", session.ID).All(&[]models.UploadSession{})
		})
		oyster_utils.LogIfError(err, nil)
	}
}

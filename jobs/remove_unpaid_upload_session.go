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

		if oyster_utils.DataMapStorageMode == oyster_utils.DataMapsInBadger {
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
		} else {
			dataMaps := []models.DataMap{}
			err := models.DB.Transaction(func(tx *pop.Connection) error {
				if err := tx.RawQuery("DELETE FROM data_maps WHERE genesis_hash = ?", session.GenesisHash).All(&dataMaps); err != nil {
					return err
				}
				return tx.RawQuery("DELETE FROM upload_sessions WHERE id = ?", session.ID).All(&[]models.UploadSession{})
			})
			if err == nil {
				var keys oyster_utils.KVKeys
				for _, dm := range dataMaps {
					keys = append(keys, dm.MsgID)
				}
				oyster_utils.BatchDelete(&keys)
			} else {
				oyster_utils.LogIfError(errors.New(err.Error()+" in transaction in "+
					"remove_unpaid_upload_session"), nil)
			}
		}

		err := models.DB.Transaction(func(tx *pop.Connection) error {
			return tx.RawQuery("DELETE FROM upload_sessions WHERE id = ?", session.ID).All(&[]models.UploadSession{})
		})

		oyster_utils.LogIfError(err, nil)
	}
}

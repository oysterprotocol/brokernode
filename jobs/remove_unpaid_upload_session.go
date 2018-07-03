package jobs

import (
	"github.com/gobuffalo/pop"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
)

/*UnpaidExpirationInHour means number of hours before it should remove unpaid upload session. */
const UnpaidExpirationInHour = 24

/*RemoveUnpaidUploadSession cleans up unpload_sessions and data_maps talbe for expired/unpaid session. */
func RemoveUnpaidUploadSession(PrometheusWrapper services.PrometheusService) {

	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramRemoveUnpaidUploadSession, start)

	sessions := []models.UploadSession{}
	err := models.DB.RawQuery(
		"SELECT * from upload_sessions WHERE payment_status != ? AND TIMESTAMPDIFF(hour, updated_at, NOW()) >= ?",
		models.PaymentStatusConfirmed, UnpaidExpirationInHour).All(&sessions)
	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return
	}

	for _, session := range sessions {
		balance := EthWrapper.CheckPRLBalance(services.StringToAddress(session.ETHAddrAlpha.String))
		if balance.Int64() > 0 {
			continue
		}

		dataMaps := []models.DataMap{}
		err := models.DB.Transaction(func(tx *pop.Connection) error {
			if err := tx.RawQuery("DELETE from data_maps WHERE genesis_hash = ?", session.GenesisHash).All(&dataMaps); err != nil {
				return err
			}
			if err := tx.RawQuery("DELETE from upload_sessions WHERE id = ?", session.ID).All(&[]models.UploadSession{}); err != nil {
				return err
			}
			return nil
		})
		oyster_utils.LogIfError(err, nil)
		if err == nil {
			services.DeleteMsgDatas(dataMaps)
		}
	}
}

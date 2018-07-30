package jobs

import (
	"errors"
	"time"

	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"gopkg.in/segmentio/analytics-go.v3"
)

func UpdateTimeOutDataMaps(thresholdTime time.Time, PrometheusWrapper services.PrometheusService) {

	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramUpdateTimeOutDataMaps, start)

	timedOutDataMaps := []models.DataMap{}

	err := models.DB.Where("status = ? AND updated_at <= ?", models.Unverified, thresholdTime).All(&timedOutDataMaps)
	oyster_utils.LogIfError(errors.New(err.Error()+" getting timed-out data_maps in UpdateTimeOutDataMaps() "+
		"in update_timed_out_data_maps"), nil)

	if len(timedOutDataMaps) > 0 {

		//when we bring back hooknodes, do decrement score somewhere in here

		var updatedDms []string
		dbOperation, _ := oyster_utils.CreateDbUpdateOperation(&models.DataMap{})

		for _, timedOutDataMap := range timedOutDataMaps {

			timedOutDataMap.Status = models.Unassigned
			updatedDms = append(updatedDms, dbOperation.GetUpdatedValue(timedOutDataMap))
		}

		err := models.BatchUpsert(
			"data_maps",
			updatedDms,
			dbOperation.GetColumns(),
			[]string{"status"})

		oyster_utils.LogIfError(errors.New(err.Error()+" updating timed-out data_maps in UpdateTimeOutDataMaps() "+
			"in update_timed_out_data_maps"), nil)

		oyster_utils.LogToSegment("update_timed_out_data_maps: chunks_timed_out", analytics.NewProperties().
			//Set("address", timedOutDataMap.Address).
			//Set("genesis_hash", timedOutDataMap.GenesisHash).
			//Set("chunk_idx", timedOutDataMap.ChunkIdx))
			Set("num_chunks", len(timedOutDataMaps)))
	}
}

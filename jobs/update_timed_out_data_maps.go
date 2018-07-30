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

		for _, timedOutDataMap := range timedOutDataMaps {

			// TODO batch this
			timedOutDataMap.Status = models.Unassigned
			models.DB.ValidateAndSave(&timedOutDataMap)
		}
		oyster_utils.LogToSegment("update_timed_out_data_maps: chunks_timed_out", analytics.NewProperties().
			//Set("address", timedOutDataMap.Address).
			//Set("genesis_hash", timedOutDataMap.GenesisHash).
			//Set("chunk_idx", timedOutDataMap.ChunkIdx))
			Set("num_chunks", len(timedOutDataMaps)))
	}
}

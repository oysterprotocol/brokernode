package jobs

import (
	"github.com/getsentry/raven-go"
	"github.com/oysterprotocol/brokernode/models"
	"time"
)

func init() {
}

func UpdateTimeOutDataMaps(thresholdTime time.Time) {

	timedOutDataMaps := []models.DataMap{}

	err := models.DB.Where("status = ? AND updated_at <= ?", models.Unverified, thresholdTime).All(&timedOutDataMaps)
	if err != nil {
		raven.CaptureError(err, nil)
	}

	if len(timedOutDataMaps) > 0 {

		//when we bring back hooknodes, do decrement somewhere in here

		for _, timedOutDataMap := range timedOutDataMaps {
			//go services.SegmentClient.Enqueue(analytics.Track{
			//	Event:  "chunk_timed_out",
			//	UserId: services.GetLocalIP(),
			//	Properties: analytics.NewProperties().
			//		Set("address", timedOutDataMap.Address).
			//		Set("genesis_hash", timedOutDataMap.GenesisHash).
			//		Set("chunk_idx", timedOutDataMap.ChunkIdx),
			//})

			timedOutDataMap.Status = models.Unassigned
			models.DB.ValidateAndSave(&timedOutDataMap)
		}
	}
}

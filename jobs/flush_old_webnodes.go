package jobs

import (
	"time"

	raven "github.com/getsentry/raven-go"
	"github.com/oysterprotocol/brokernode/models"
)

func init() {
}

func FlushOldWebNodes(thresholdTime time.Time) {

	webnodes := []models.Webnode{}
	err := models.DB.Where("updated_at <= ?", thresholdTime).All(&webnodes)

	if err != nil {
		raven.CaptureError(err, nil)
	} else {
		for i := 0; i < len(webnodes); i++ {
			webnode := webnodes[i]
			models.DB.Destroy(&webnode)
		}
	}
}

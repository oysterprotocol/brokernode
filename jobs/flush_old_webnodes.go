package jobs

import (
	"fmt"
	"github.com/oysterprotocol/brokernode/utils"
	"gopkg.in/segmentio/analytics-go.v3"
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
			oyster_utils.LogToSegment("flush_old_webnode", analytics.NewProperties().
				Set("webnode_id", fmt.Sprint(webnodes[i].ID)).
				Set("webnode_address", webnodes[i].Address))

			webnode := webnodes[i]
			models.DB.Destroy(&webnode)
		}
	}
}

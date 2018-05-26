package jobs

import (
	"fmt"
	"time"

	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/utils"
	"gopkg.in/segmentio/analytics-go.v3"
)

func FlushOldWebNodes(thresholdTime time.Time) {

	webnodes := []models.Webnode{}
	err := models.DB.Where("updated_at <= ?", thresholdTime).All(&webnodes)

	if err != nil {
		oyster_utils.LogIfError(err)
		return
	}

	for i := 0; i < len(webnodes); i++ {
		oyster_utils.LogToSegment("flush_old_wednodes: flushing_old_webnode", analytics.NewProperties().
			Set("webnode_id", fmt.Sprint(webnodes[i].ID)).
			Set("webnode_address", webnodes[i].Address))

		webnode := webnodes[i]
		err := models.DB.Destroy(&webnode)
		oyster_utils.LogIfError(err)
	}
}

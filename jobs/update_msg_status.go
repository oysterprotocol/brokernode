package jobs

import (
	"github.com/oysterprotocol/brokernode/services"
)

/*UpdateMsgStatus checks badger to verify that message data for particular chunks has arrived, and if so,
updates the msg_status field of those chunks*/
func UpdateMsgStatus(PrometheusWrapper services.PrometheusService) {

}

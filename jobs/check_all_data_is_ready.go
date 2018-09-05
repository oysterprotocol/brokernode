package jobs

import (
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
)

/*CheckAllDataIsReady checks if all data for sessions has been stored*/
func CheckAllDataIsReady(PrometheusWrapper services.PrometheusService) {

	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramCheckAllDataIsReady, start)

	CheckSessionsWithIncompleteData()

	sessions, err := models.GetSessionsWithIncompleteData()

	if err != nil {
		oyster_utils.LogIfError(err, nil)
	}

	for _, session := range sessions {

		ready := session.CheckIfAllDataIsReady()

		if ready {
			u := models.UploadSession{}
			models.DB.Find(&u, session.ID)
			u.AllDataReady = models.AllDataReady
			models.DB.ValidateAndUpdate(&u)
		}
	}
}

/*CheckSessionsWithIncompleteData grabs the sessions with incomplete data
and checks if they are complete.*/
func CheckSessionsWithIncompleteData() {
	sessions, err := models.GetSessionsWithIncompleteData()

	if err != nil {
		oyster_utils.LogIfError(err, nil)
	}

	for _, session := range sessions {

		ready := session.CheckIfAllDataIsReady()

		if ready {
			u := models.UploadSession{}
			models.DB.Find(&u, session.ID)
			u.AllDataReady = models.AllDataReady
			models.DB.ValidateAndUpdate(&u)
		}
	}
}

package jobs

import (
	"github.com/getsentry/raven-go"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
)

func init() {
}

func ProcessUnassignedChunks(iotaWrapper services.IotaService) {

	chunks, err := GetUnassignedChunks()

	if err != nil {
		raven.CaptureError(err, nil)
	} else {
		if len(chunks) > 0 {

			for i := 0; i < len(chunks); i += BundleSize {
				end := i + BundleSize

				if end > len(chunks) {
					end = len(chunks)
				}

				iotaWrapper.ProcessChunks(chunks[i:end], false)
			}
		}
	}
}

func GetUnassignedChunks() (dataMaps []models.DataMap, err error) {

	query := models.DB.Where("status = ?", models.Unassigned)
	dataMaps = []models.DataMap{}
	err = query.All(&dataMaps)
	if err != nil {
		raven.CaptureError(err, nil)
	}

	return dataMaps, err
}


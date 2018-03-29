package jobs

import (
	"github.com/getsentry/raven-go"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"

)

func init() {
}

func ProcessUnassignedChunks(iotaWrapper services.IotaService) {

	channels, _ := models.GetReadyChannels()

	if len(channels) > 0 {
		AssignChunksToChannels(&channels, iotaWrapper)
	}
}

func AssignChunksToChannels(channels *[]models.ChunkChannel, iotaWrapper services.IotaService) {

	/*
	TODO:  More sophisticated chunk grabbing.  I.e. only grab as many as
	we have ready channels for, and try to grab an equal number per unique
	genesis hash
	 */

	chunks, err := GetUnassignedChunks()

	if err != nil {
		raven.CaptureError(err, nil)
	} else {
		if len(chunks) > 0 {

			j := 0

			for _, channel := range *channels {
				end := j + BundleSize

				if end > len(chunks) {
					end = len(chunks)
				}

				if j >= end {
					break
				}

				iotaWrapper.SendChunksToChannel(chunks[j:end], &channel)

				j += BundleSize
				if j > len(chunks) {
					break
				}
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

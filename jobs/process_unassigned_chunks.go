package jobs

import (
	"fmt"
	"github.com/getsentry/raven-go"
	"github.com/oysterprotocol/brokernode/models"
)

func init() {
}

func ProcessUnassignedChunks(processChunks ProcessChunksFunc) {

	chunks, err := GetUnassignedChunks()

	if err != nil {
		raven.CaptureError(err, nil)
	} else {
		processChunks(chunks)
	}
}

func GetUnassignedChunks() (dataMaps []models.DataMap, err error) {

	//query := models.DB.Where("status = ? AND updated_at >= ?", strconv.Itoa(models.ChunkStatus["unassigned"]), thresholdTime)
	query := models.DB.Where("status = ?", models.Unassigned)
	dataMaps = []models.DataMap{}
	err = query.All(&dataMaps)
	if err != nil {
		raven.CaptureError(err, nil)
		fmt.Printf("%v\n", err)
	} else {
		fmt.Print("Success!\n")
	}

	return dataMaps, err
}

func processChunks(dataMaps []models.DataMap) {
	if len(dataMaps) > 0 {

		for i := 0; i < len(dataMaps); i += BundleSize {
			end := i + BundleSize

			if end > len(dataMaps) {
				end = len(dataMaps)
			}

			// send to broker code that processes these
			//dataMaps[i:end]
		}
	}
}
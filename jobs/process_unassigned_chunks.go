package jobs

import (
	"time"
	"fmt"
	//"errors"
	"github.com/gobuffalo/pop"
	"github.com/getsentry/raven-go"
	"github.com/oysterprotocol/brokernode/models"
)

func init() {
}

func handle(thresholdTime time.Time) {
	tx, err := pop.Connect("development")
	if err != nil {
		raven.CaptureError(err, nil)
	}

	query := tx.Where("status = 'unassigned' AND updated_at >= ?", thresholdTime)
	dataMaps := []models.DataMap{}
	err = query.All(&dataMaps)
	if err != nil {
		fmt.Print("ERROR!\n")
		fmt.Printf("%v\n", err)
	} else {
		fmt.Print("Success!\n")
		fmt.Printf("%v\n", dataMaps)

		if len(dataMaps) > 0 {

			chunkSize := 10 // put 10 somewhere else

			for i := 0; i < len(dataMaps); i += chunkSize {
				end := i + chunkSize

				if end > len(dataMaps) {
					end = len(dataMaps)
				}

				processChunks(dataMaps[i:end]);
			}
		}
	}
}

func processChunks(dataMaps []models.DataMap) {
	fmt.Println(dataMaps)
}

package jobs

import (
	"fmt"
	"github.com/gobuffalo/pop"
	"github.com/getsentry/raven-go"
	"github.com/oysterprotocol/brokernode/models"
	"strconv"
)

var BundleSize = 10

type ProcessChunksFunc func(dataMaps []models.DataMap)

type ChunkProcessor struct {
	processChunks ProcessChunksFunc
}

func init() {
	NewChunkProcessor(processChunks)
}

func NewChunkProcessor(processChunks ProcessChunksFunc) *ChunkProcessor {
	return &ChunkProcessor{processChunks: processChunks}
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

	tx, err := pop.Connect("test")
	if err != nil {
		raven.CaptureError(err, nil)
	}

	models.SetChunkStatuses()

	//query := tx.Where("status = ? AND updated_at >= ?", strconv.Itoa(models.ChunkStatus["unassigned"]), thresholdTime)
	query := tx.Where("status = ?", strconv.Itoa(models.ChunkStatus["unassigned"]))
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
		}
	}
}
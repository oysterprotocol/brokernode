package jobs

import (
	"fmt"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
)

var BundleSize = 10

type ProcessChunksFunc func(dataMaps []models.DataMap)

type ChunkProcessor struct {
	processChunks ProcessChunksFunc
}

var IotaWrapper services.IotaService

func init() {
	NewChunkProcessor(processChunks)
	IotaWrapper := services.IotaWrapper
	fmt.Println(IotaWrapper)
}

func NewChunkProcessor(processChunks ProcessChunksFunc) *ChunkProcessor {
	return &ChunkProcessor{processChunks: processChunks}
}

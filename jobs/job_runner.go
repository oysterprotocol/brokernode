package jobs

import "github.com/oysterprotocol/brokernode/models"

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
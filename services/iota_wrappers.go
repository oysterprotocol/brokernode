package services

import (
	"github.com/oysterprotocol/brokernode/models"
)

type IotaWrapper struct {
}

type FilteredChunk struct {
	MatchesTangle      []models.DataMap
	DoesNotMatchTangle []models.DataMap
	NotAttached        []models.DataMap
}

func VerifyChunkMessageMatchesTangle(chunks []models.DataMap) []models.DataMap {
	return chunks
}

func ProcessChunks(chunks []models.DataMap, reattachIfAlreadyAttached bool) []models.DataMap {
	return chunks
}

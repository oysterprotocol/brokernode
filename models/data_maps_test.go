package models_test

import (
	"github.com/oysterprotocol/brokernode/models"
)

func (ms *ModelSuite) Test_BuilDataMaps() {
	genHash := "genHashTest"
	fileBytesCount := 9000

	models.BuildDataMaps(genHash, fileBytesCount)

	expectedHashes := []string{
		"genHashTest",
		"a973bb9fbbbdb35ff0c918e2eb017ee599b9135382cf1ffabf4daf8247a42a64",
		"8ea51f148b6ca31e48825453290e9087dbbea5c58bcd13442f5b6610990b3290",
		"85b29c4787c5af60002584d9d985d69b1d4fe8022927803692ca323922bd3228",
		"53ad8a87078369b60d6e1a39b0cb5512801be47e84c5b249fd6fa844cc2bc776",
	}

	dMaps := []models.DataMap{}
	ms.DB.Where("genesis_hash = ?", genHash).Order("chunk_idx asc").All(&dMaps)

	for i, dMap := range dMaps {
		ms.Equal(dMap.Hash, expectedHashes[i])
	}
}

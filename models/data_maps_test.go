package models_test

import (
	"github.com/oysterprotocol/brokernode/models"
)

func (ms *ModelSuite) Test_BuildDataMaps() {
	genHash := "genHashTest"
	fileBytesCount := 9000

	vErr, err := models.BuildDataMaps(genHash, fileBytesCount)
	ms.Nil(err)
	ms.Equal(0, len(vErr.Errors))

	expectedHashes := []string{
		"genHashTest",
		"a973bb9fbbbdb35ff0c918e2eb017ee599b9135382cf1ffabf4daf8247a42a64",
		"8ea51f148b6ca31e48825453290e9087dbbea5c58bcd13442f5b6610990b3290",
		"85b29c4787c5af60002584d9d985d69b1d4fe8022927803692ca323922bd3228",
		"53ad8a87078369b60d6e1a39b0cb5512801be47e84c5b249fd6fa844cc2bc776",
		"44bedaa422b51f2e8d6f393341ae8a64b7ae93ce42eb4d8f7d0fb79fe3bb76b9",
		"c01e47a63a7fa5033ff3b7145fdb5cbe36987de667e9c5a84d995d169f850ddc",
	}

	dMaps := []models.DataMap{}
	ms.DB.Where("genesis_hash = ?", genHash).Order("chunk_idx asc").All(&dMaps)

	for i, dMap := range dMaps {
		ms.Equal(expectedHashes[i], dMap.Hash)
	}
}

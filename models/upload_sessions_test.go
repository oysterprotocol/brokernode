package models_test

import (
	"github.com/oysterprotocol/brokernode/models"
)

func (ms *ModelSuite) Test_StartUploadSession() {
	genHash := "genHashTest"
	fileSizeBytes := 123

	u := models.UploadSession{
		GenesisHash:   genHash,
		FileSizeBytes: fileSizeBytes,
	}

	vErr, err := u.StartUploadSession()
	ms.Nil(err)
	ms.Equal(0, len(vErr.Errors))

	uSession := models.UploadSession{}
	ms.DB.Where("genesis_hash = ?", genHash).First(&uSession)

	ms.Equal(genHash, uSession.GenesisHash)
	ms.Equal(fileSizeBytes, uSession.FileSizeBytes)
	ms.Equal(models.SessionTypeAlpha, uSession.Type)
}

func (ms *ModelSuite) Test_DataMapsForSession() {
	genHash := "genHashTest"
	fileSizeBytes := 123

	u := models.UploadSession{
		GenesisHash:   genHash,
		FileSizeBytes: fileSizeBytes,
	}

	vErr, err := u.StartUploadSession()
	ms.Nil(err)
	ms.Equal(0, len(vErr.Errors))

	expectedHashes := []string{
		"genHashTest",
		"a973bb9fbbbdb35ff0c918e2eb017ee599b9135382cf1ffabf4daf8247a42a64",
		"8ea51f148b6ca31e48825453290e9087dbbea5c58bcd13442f5b6610990b3290",
		"85b29c4787c5af60002584d9d985d69b1d4fe8022927803692ca323922bd3228",
		"53ad8a87078369b60d6e1a39b0cb5512801be47e84c5b249fd6fa844cc2bc776",
	}

	dMaps, err := u.DataMapsForSession()
	ms.Nil(err)

	for i, dMap := range *dMaps {
		ms.Equal(expectedHashes[i], dMap.Hash)
	}
}

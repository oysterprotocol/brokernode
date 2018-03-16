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

	err := u.StartUploadSession()
	ms.Nil(err)

	uSession := models.UploadSession{}
	ms.DB.Where("genesis_hash = ?", genHash).First(&uSession)

	ms.Equal(uSession.GenesisHash, genHash)
	ms.Equal(uSession.FileSizeBytes, fileSizeBytes)
	ms.Equal(uSession.Type, models.SessionTypeAlpha)
}

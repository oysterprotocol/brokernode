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

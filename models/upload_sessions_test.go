package models_test

import (
	"github.com/oysterprotocol/brokernode/models"
)

func (ms *ModelSuite) Test_StartUploadSession() {
	genHash := "genHashTest"
	fileSizeBytes := 123
	storageLengthInYears := 2

	u := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          genHash,
		FileSizeBytes:        fileSizeBytes,
		StorageLengthInYears: storageLengthInYears,
	}

	vErr, err := u.StartUploadSession()
	ms.Nil(err)
	ms.Equal(0, len(vErr.Errors))

	uSession := models.UploadSession{}
	ms.DB.Where("genesis_hash = ?", genHash).First(&uSession)

	ms.Equal(genHash, uSession.GenesisHash)
	ms.Equal(fileSizeBytes, uSession.FileSizeBytes)
	ms.Equal(models.SessionTypeAlpha, uSession.Type)
	ms.Equal(2.0, uSession.TotalCost)
	ms.Equal(2, uSession.StorageLengthInYears)
}

func (ms *ModelSuite) Test_DataMapsForSession() {
	genHash := "genHashTest"
	fileSizeBytes := 123
	storageLengthInYears := 3

	u := models.UploadSession{
		GenesisHash:          genHash,
		FileSizeBytes:        fileSizeBytes,
		StorageLengthInYears: storageLengthInYears,
	}

	vErr, err := u.StartUploadSession()
	ms.Nil(err)
	ms.Equal(0, len(vErr.Errors))

	expectedHashes := []string{
		"ef7edf0decd95c9e094184dca8641b68bb3ca0f69fec086341893816c68f7d9d408131fa01a66cf95f05b2a038185db9",
		"86ad8449bd1b32bcd86d86cfe7b3b6453f391c0c0df57956a2dff53f55709af3cd43a983ef46263cf8e361ae15734b33",
		"fbb914b1ba9cc663be0eb7b2570209af5caccfe5b7bba65e832c683072a969715e1b23866ce97ddb765fefe9b991e652",
		"e697116fd36a697f327f4682fd6f72250933bc61184fc36ff89badf749779aadf643b2e4f3fcd22fa9c07a6ce89c99a5",
		"167b2e33d17a4a96c6ad7216cd49c664b056efd30c08d65a354d1a5eb9cc9dbcb2f639495269f7ef5e56b8e62777edfc",
	}

	dMaps, err := u.DataMapsForSession()
	ms.Nil(err)

	for i, dMap := range *dMaps {
		ms.Equal(expectedHashes[i], dMap.ObfuscatedHash)
	}
}

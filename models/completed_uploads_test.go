package models_test

import (
	"github.com/gobuffalo/pop/nulls"
	"github.com/oysterprotocol/brokernode/models"
)

func (ms *ModelSuite) Test_NewCompletedUpload() {
	fileBytesCount := 2500

	err := models.NewCompletedUpload(models.UploadSession{
		GenesisHash:   "genHash1",
		FileSizeBytes: fileBytesCount,
		Type:          models.SessionTypeBeta,
		ETHAddrAlpha:  nulls.String{string("SOME_ALPHA_ETH_ADDRESS"), true},
		ETHAddrBeta:   nulls.String{string("SOME_BETA_ETH_ADDRESS"), true},
		ETHPrivateKey: "SOME_PRIVATE_KEY",
	})
	ms.Equal(nil, err)

	err = models.NewCompletedUpload(models.UploadSession{
		GenesisHash:   "genHash2",
		FileSizeBytes: fileBytesCount,
		Type:          models.SessionTypeAlpha,
		ETHAddrAlpha:  nulls.String{string("SOME_ALPHA_ETH_ADDRESS"), true},
		ETHAddrBeta:   nulls.String{string("SOME_BETA_ETH_ADDRESS"), true},
		ETHPrivateKey: "SOME_PRIVATE_KEY",
	})
	ms.Equal(nil, err)

	err = models.NewCompletedUpload(models.UploadSession{ // no session type
		GenesisHash:   "genHash3",
		FileSizeBytes: fileBytesCount,
		ETHAddrAlpha:  nulls.String{string("SOME_ALPHA_ETH_ADDRESS"), true},
		ETHAddrBeta:   nulls.String{string("SOME_BETA_ETH_ADDRESS"), true},
		ETHPrivateKey: "SOME_PRIVATE_KEY",
	})
	ms.NotEqual(err, nil)
	ms.Equal("no session type provided for session in method models.NewCompletedUpload", err.Error())

	completedUploads := []models.CompletedUpload{}
	err = ms.DB.All(&completedUploads)
	ms.Equal(nil, err)

	ms.Equal(2, len(completedUploads))

	for _, completedUpload := range completedUploads {
		ms.Equal(true, completedUpload.GenesisHash == "genHash1" ||
			completedUpload.GenesisHash == "genHash2")
		if completedUpload.GenesisHash == "genHash1" {
			ms.Equal("SOME_BETA_ETH_ADDRESS", completedUpload.ETHAddr)
		}
		if completedUpload.GenesisHash == "genHash2" {
			ms.Equal("SOME_ALPHA_ETH_ADDRESS", completedUpload.ETHAddr)
		}
	}
}

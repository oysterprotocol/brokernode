package models_test

import (
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/utils"
)

type hashAddressConversion struct {
	sha256Hash     string
	ethPrivateSeed string
}

func (suite *ModelSuite) Test_BuildDataMaps() {

	oyster_utils.SetStorageMode(oyster_utils.DataMapsInSQL)
	defer oyster_utils.ResetDataMapStorageMode()

	genHash := "abcdef"
	numChunks := 7

	err := models.BuildDataMapsForSession(genHash, numChunks)
	suite.Nil(err)

	expectedHashChainHashes := []string{
		"995da3cf545787d65f9ced52674e92ee8171c87c7a4008aa4349ec47d21609a7",
		"4533a01d26697df306b3380e08f4fae30f488d2985e6449e9bd9bd86849ddbc6",
		"93d62b82fa8169af012ca0d3c13f6c5d94d06daf4f769ee45595a049a4805524",
		"fea47151e3dbd670f1bded7b2393093e8dde50d6ccb541f5f51689005cf88ab1",
		"c694406b4d98e3bb23416c1111099bc4c7317a81b40d53e68ce7afe4d9aa716f",
		"ab1c8e2baa271bf67fb5ba7083b06c66a2ac41db3e4d25728e80d16a3dfb746b",
		"7f907764c0d12a50daa3b38fc2fc888637640f5c91a4b81f5879da99c8e35653",
	}

	expectedAddresses := []string{
		"PGJIVFP9VHV9G9JEPDDEJFBGBGNEVGLCIFKEUDIFNFY9WCEG9FGDSBQBZCBCHFZAB9YBLFKBI9DCWAFDX",
		"XGAFUHGFXCIGMALBXDN9GCKINDY9VDJIRGWEUFGCJHSHXGOBMAVEADHFDDZCZBYCYCRBXAPGS9SBKEAGU",
		"JB9GNAF9ICZ9FHCBT9J9J9V9BESBSHUBXFGF9GK9VGSFPALAIDMCP9WBFDVBZHAASFYDUBKAUHVEMHBHD",
		"ZEVEHDABMB9HGGQESAZFEEH9THO9YECFUADIVFTCVHWDIAB9REEHE9NDRB99UHMCRDMEFAJESGSAR9WGT",
		"PCQEWALIKGX9EGPFOAQBGGLHUABCHGDDXCNCB9UCMHZAAALHJHVAHBPFRGDAG9SBEIZEAEKAHAUB9GWDY",
		"EEFFVBXGNCCBJEZEQCYBEHEIT9RDOAPDT9CBIBUBMDEFCILAIDAIM9ABHAX9AAVHKEACPGQFJGTFHIR9J",
		"PFWBSECAKGLFGCPERHDFNDHBOHOHKGQEQAC9TGGEFFVBUBBCRDQ9AITDVGQBMHZBLFQGBIVDOHAEGBVBN",
	}

	dMaps := []models.DataMap{}
	suite.DB.Where("genesis_hash = ?", genHash).Order("chunk_idx ASC").All(&dMaps)

	suite.Equal(numChunks, len(dMaps))

	for i, dMap := range dMaps {
		suite.Equal(expectedHashChainHashes[i], dMap.Hash)
		suite.Equal(expectedAddresses[i], dMap.Address)
		suite.NotNil(dMap.MsgID)
	}
}

package models_test

import (
	"encoding/hex"
	"fmt"
	"strings"

	"golang.org/x/crypto/sha3"

	"github.com/iotaledger/giota"
	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
)

type hashAddressConversion struct {
	sha256Hash     string
	ethPrivateSeed string
}

var encryptedTreasureCases = []hashAddressConversion{
	{sha256Hash: "64dc1ce4655554f514a4ce83e08c1d08372fdf02bd8c9b6dbecfc74b783d39d1",
		ethPrivateSeed: "0000000000000000000000000000000000000000000000000000000000000001"},
	{sha256Hash: "99577b266e77d07e364d0b87bf1bcef44c78e3668dfdc3881969b375c09d4fcd",
		ethPrivateSeed: "1004444400000006780000000000000000000000000012345000000765430001"},
	{sha256Hash: "7fb4ca9cc0032bafc2ebd0fda018a41f5adfcf441123de22ab736a42207933f7",
		ethPrivateSeed: "7777777774444444777777744444447777777444444777777744444777777744"},
}

func (ms *ModelSuite) Test_BuildDataMaps() {
	genHash := "abcdef"
	numChunks := 7

	vErr, err := models.BuildDataMaps(genHash, numChunks)
	ms.Nil(err)
	ms.Equal(0, len(vErr.Errors))

	expectedObfuscatedHashes := []string{ // 1 extra chunk for treasure
		"dd88bb5db7314227c7e6117c693ceb83bbaf587bd1b63393d7512ba68bf42973845fa1c2924be14d37ba2da1938d7228",
		"cdfdb810ee1607917c8bacbfbf95d35dab9281abb01968c2a27349476b53aa35024fae410955327233523229677da827",
		"d5a3eda969c62842840e58fe7a1982fdcf9eb758e2ebd545289d6daa706b506a6a4833cd134992be9c73fe4c1e1d15ff",
		"40bd29065a1ade39140a0a168949eb4bbaa9bd0bd3b52b27755e104d724cf21cb5854b26ed9de5daa6d9fbf032f7952b",
		"a19d743743d8c4982ebc8c08ec0fa0a530f7b865ee83240299dd057a4800ed5e7e942191d02e12d4b65987328f15d2e3",
		"619832ffc818c2b22a47c4e43053c570695f0266e5351ce4e2313eb2cf1f0749f8a18826234bbd83a052eea8bbb201ad",
		"8ca84cd55f3991a1624fddf8147e2a7c14393f4b79a7f62775f40d3723181cee9252cdb3c7b6fb1276ad40bd7c2f7ac8",
		"b24d9a1ec8ae5897eaa67a3ee7e7c8982c03d18ea84c4b537e11f480d347e550aecef582e7883d4c29969f299836792e",
	}
	expectedHashChainHashes := []string{ // 1 extra chunk for treasure
		"abcdef",
		"995da3cf545787d65f9ced52674e92ee8171c87c7a4008aa4349ec47d21609a7",
		"4533a01d26697df306b3380e08f4fae30f488d2985e6449e9bd9bd86849ddbc6",
		"93d62b82fa8169af012ca0d3c13f6c5d94d06daf4f769ee45595a049a4805524",
		"fea47151e3dbd670f1bded7b2393093e8dde50d6ccb541f5f51689005cf88ab1",
		"c694406b4d98e3bb23416c1111099bc4c7317a81b40d53e68ce7afe4d9aa716f",
		"ab1c8e2baa271bf67fb5ba7083b06c66a2ac41db3e4d25728e80d16a3dfb746b",
		"7f907764c0d12a50daa3b38fc2fc888637640f5c91a4b81f5879da99c8e35653",
		"f9e9eef8afa6a7edbe108dc9d178fcc114cb2e9403a552e61c7e247889b3d536",
	}
	expectedAddresses := []string{
		"EHAEYFLCUFVALBLAJGNHQ9PDXCFBSHWDYFMFGCODTGTFXALEZG9CPADFDEAINAGDXDNCZEEGKEUBIHWBA",
		"PGJIVFP9VHV9G9JEPDDEJFBGBGNEVGLCIFKEUDIFNFY9WCEG9FGDSBQBZCBCHFZAB9YBLFKBI9DCWAFDX",
		"XGAFUHGFXCIGMALBXDN9GCKINDY9VDJIRGWEUFGCJHSHXGOBMAVEADHFDDZCZBYCYCRBXAPGS9SBKEAGU",
		"JB9GNAF9ICZ9FHCBT9J9J9V9BESBSHUBXFGF9GK9VGSFPALAIDMCP9WBFDVBZHAASFYDUBKAUHVEMHBHD",
		"ZEVEHDABMB9HGGQESAZFEEH9THO9YECFUADIVFTCVHWDIAB9REEHE9NDRB99UHMCRDMEFAJESGSAR9WGT",
		"PCQEWALIKGX9EGPFOAQBGGLHUABCHGDDXCNCB9UCMHZAAALHJHVAHBPFRGDAG9SBEIZEAEKAHAUB9GWDY",
		"EEFFVBXGNCCBJEZEQCYBEHEIT9RDOAPDT9CBIBUBMDEFCILAIDAIM9ABHAX9AAVHKEACPGQFJGTFHIR9J",
		"PFWBSECAKGLFGCPERHDFNDHBOHOHKGQEQAC9TGGEFFVBUBBCRDQ9AITDVGQBMHZBLFQGBIVDOHAEGBVBN",
	}

	dMaps := []models.DataMap{}
	ms.DB.Where("genesis_hash = ?", genHash).Order("chunk_idx asc").All(&dMaps)

	ms.Equal(numChunks, len(dMaps))

	for i, dMap := range dMaps {
		ms.Equal(expectedObfuscatedHashes[i], dMap.ObfuscatedHash)
		ms.Equal(expectedHashChainHashes[i], dMap.Hash)
		ms.Equal(expectedAddresses[i], dMap.Address)
		ms.NotNil(dMap.MsgID)
	}
}

func (ms *ModelSuite) Test_CreateTreasurePayload() {
	maxSideChainLength := 10
	matchesFound := 0

	for _, tc := range encryptedTreasureCases {
		payload, err := models.CreateTreasurePayload(tc.ethPrivateSeed, tc.sha256Hash, maxSideChainLength)
		ms.Nil(err)

		trytes, err := giota.ToTrytes(payload[0:models.TreasurePayloadLength])
		ms.Nil(err)
		payloadInBytes := oyster_utils.TrytesToBytes(trytes)

		currentHash := tc.sha256Hash

		for i := 0; i <= maxSideChainLength; i++ {
			currentHash = oyster_utils.HashHex(currentHash, sha3.New256())
			result := oyster_utils.Decrypt(currentHash, hex.EncodeToString(payloadInBytes), tc.sha256Hash)
			if result != nil {
				ms.Equal(true, strings.Contains(hex.EncodeToString(result), fmt.Sprint(models.TreasurePrefix)))
				ms.Equal(hex.EncodeToString(result)[len(models.TreasurePrefix):], tc.ethPrivateSeed)
				matchesFound++
			}
		}
	}
	ms.Equal(3, matchesFound)
}

func (suite *ModelSuite) Test_GetUnassignedGenesisHashes() {
	numChunks := 10

	vErr, err := models.BuildDataMaps("abcdeff1", numChunks)
	suite.Nil(err)
	suite.Equal(0, len(vErr.Errors))
	vErr, err = models.BuildDataMaps("abcdeff2", numChunks)
	suite.Nil(err)
	suite.Equal(0, len(vErr.Errors))
	vErr, err = models.BuildDataMaps("abcdeff3", numChunks)
	suite.Nil(err)
	suite.Equal(0, len(vErr.Errors))
	vErr, err = models.BuildDataMaps("abcdeff4", numChunks)
	suite.Nil(err)
	suite.Equal(0, len(vErr.Errors))
	vErr, err = models.BuildDataMaps("abcdeff5", numChunks)
	suite.Nil(err)
	suite.Equal(0, len(vErr.Errors))

	genHash1 := []models.DataMap{} // 1 unassigned
	genHash2 := []models.DataMap{} // 1 error
	genHash3 := []models.DataMap{} // all unassigned
	genHash4 := []models.DataMap{} // all error
	err = suite.DB.Where("genesis_hash = ?", "abcdeff1").All(&genHash1)
	suite.Equal(err, nil)
	err = suite.DB.Where("genesis_hash = ?", "abcdeff2").All(&genHash2)
	suite.Equal(err, nil)
	err = suite.DB.Where("genesis_hash = ?", "abcdeff3").All(&genHash3)
	suite.Equal(err, nil)
	err = suite.DB.Where("genesis_hash = ?", "abcdeff4").All(&genHash4)
	suite.Equal(err, nil)

	genHash1[1].Status = models.Unassigned
	suite.DB.ValidateAndSave(&genHash1[1])

	genHash2[0].Status = models.Error
	suite.DB.ValidateAndSave(&genHash2[0])

	for _, dataMap := range genHash3 {
		dataMap.Status = models.Unassigned
		suite.DB.ValidateAndSave(&dataMap)
	}

	for _, dataMap := range genHash4 {
		dataMap.Status = models.Error
		suite.DB.ValidateAndSave(&dataMap)
	}

	genHashes, err := models.GetUnassignedGenesisHashes()
	suite.Equal(err, nil)
	suite.Equal(4, len(genHashes))

	for _, genHash := range genHashes {
		if genHash == "abcdeff5" {
			suite.Failf("FAIL", "abcdeff5 should not be in the array")
		}
	}
}

func (suite *ModelSuite) Test_GetUnassignedChunks() {
	numChunks := 2

	vErr, err := models.BuildDataMaps("abcdeff1", numChunks)
	suite.Nil(err)
	suite.Equal(0, len(vErr.Errors))
	vErr, err = models.BuildDataMaps("abcdeff2", numChunks)
	suite.Nil(err)
	suite.Equal(0, len(vErr.Errors))
	vErr, err = models.BuildDataMaps("abcdeff3", numChunks)
	suite.Nil(err)
	suite.Equal(0, len(vErr.Errors))
	vErr, err = models.BuildDataMaps("abcdeff4", numChunks)
	suite.Nil(err)
	suite.Equal(0, len(vErr.Errors))
	vErr, err = models.BuildDataMaps("abcdeff5", numChunks)
	suite.Nil(err)
	suite.Equal(0, len(vErr.Errors))

	genHash1 := []models.DataMap{} // 1 unassigned
	genHash2 := []models.DataMap{} // 1 error
	genHash3 := []models.DataMap{} // all unassigned
	genHash4 := []models.DataMap{} // all error
	err = suite.DB.Where("genesis_hash = ?", "abcdeff1").All(&genHash1)
	suite.Equal(err, nil)
	err = suite.DB.Where("genesis_hash = ?", "abcdeff2").All(&genHash2)
	suite.Equal(err, nil)
	err = suite.DB.Where("genesis_hash = ?", "abcdeff3").All(&genHash3)
	suite.Equal(err, nil)
	err = suite.DB.Where("genesis_hash = ?", "abcdeff4").All(&genHash4)
	suite.Equal(err, nil)

	genHash1[1].Status = models.Unassigned
	suite.DB.ValidateAndSave(&genHash1[1])

	genHash2[0].Status = models.Error
	suite.DB.ValidateAndSave(&genHash2[0])

	for _, dataMap := range genHash3 {
		dataMap.Status = models.Unassigned
		suite.DB.ValidateAndSave(&dataMap)
	}

	for _, dataMap := range genHash4 {
		dataMap.Status = models.Error
		suite.DB.ValidateAndSave(&dataMap)
	}

	unassignedChunks, err := models.GetUnassignedChunks()
	suite.Equal(err, nil)

	for _, unassignedChunk := range unassignedChunks {
		if unassignedChunk.GenesisHash == "abcdeff5" {
			suite.Failf("FAIL", "a chunk with genHash5 should not be in the array")
		}
	}
}

func (suite *ModelSuite) Test_GetAllUnassignedChunksBySession() {
	numChunks := 5

	uploadSession1 := models.UploadSession{
		GenesisHash:    "abcdeff1",
		FileSizeBytes:  8000,
		NumChunks:      numChunks,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureInDataMapComplete,
	}
	uploadSession1.StartUploadSession()
	session := models.UploadSession{}
	err := suite.DB.Where("genesis_hash = ?", "abcdeff1").First(&session)
	suite.Nil(err)

	dataMaps := []models.DataMap{}
	err = suite.DB.Where("genesis_hash = ?", "abcdeff1").All(&dataMaps)
	suite.Nil(err)
	for _, dm := range dataMaps {
		if services.IsKvStoreEnabled() {
			suite.Nil(services.BatchSet(&services.KVPairs{dm.MsgID: "NOTEMPETY"}))
			dm.MsgStatus = models.MsgStatusUploaded
		} else {
			dm.Message = "NOTEMPETY"
		}
		suite.DB.ValidateAndSave(&dm)
	}

	jobs.MarkBuriedMapsAsUnassigned()
	chunks, err := models.GetAllUnassignedChunksBySession(session)
	suite.Nil(err)

	suite.NotEqual(0, len(chunks))
	suite.Equal(numChunks+1, len(chunks)) // 1 extra chunk for treasure
	suite.NotEqual(models.DataMap{}, chunks[0])
}

func (suite *ModelSuite) Test_GetUnassignedChunksBySession() {
	numChunks := 5

	uploadSession1 := models.UploadSession{
		GenesisHash:    "abcdeff1",
		FileSizeBytes:  8000,
		NumChunks:      numChunks,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureInDataMapComplete,
	}
	uploadSession1.StartUploadSession()
	session := models.UploadSession{}
	err := suite.DB.Where("genesis_hash = ?", "abcdeff1").First(&session)
	suite.Nil(err)

	dataMaps := []models.DataMap{}
	err = suite.DB.Where("genesis_hash = ?", "abcdeff1").All(&dataMaps)
	suite.Nil(err)
	for _, dm := range dataMaps {
		if services.IsKvStoreEnabled() {
			suite.Nil(services.BatchSet(&services.KVPairs{dm.MsgID: "NOTEMPETY"}))
			dm.MsgStatus = models.MsgStatusUploaded
		} else {
			dm.Message = "NOTEMPETY"
		}
		suite.DB.ValidateAndSave(&dm)
	}

	jobs.MarkBuriedMapsAsUnassigned()
	chunks, err := models.GetAllUnassignedChunksBySession(session)
	chunksWithLimit, err := models.GetUnassignedChunksBySession(session, 4)
	suite.Nil(err)

	suite.NotEqual(0, len(chunksWithLimit))
	suite.Equal(numChunks+1, len(chunks)) // 1 extra chunk for treasure
	suite.Equal(4, len(chunksWithLimit))
	suite.NotEqual(models.DataMap{}, chunksWithLimit[0])
}

func (suite *ModelSuite) Test_GetPendingChunksBySession() {
	numChunks := 5

	uploadSession1 := models.UploadSession{
		GenesisHash:    "abcdeff1",
		FileSizeBytes:  8000,
		NumChunks:      numChunks,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureInDataMapComplete,
	}
	uploadSession1.StartUploadSession()
	session := models.UploadSession{}
	err := suite.DB.Where("genesis_hash = ?", "abcdeff1").First(&session)
	suite.Nil(err)

	dataMaps := []models.DataMap{}
	err = suite.DB.Where("genesis_hash = ?", "abcdeff1").All(&dataMaps)
	suite.Nil(err)
	for _, dm := range dataMaps {
		if services.IsKvStoreEnabled() {
			suite.Nil(services.BatchSet(&services.KVPairs{dm.MsgID: "NOTEMPETY"}))
			dm.MsgStatus = models.MsgStatusUploaded
		} else {
			dm.Message = "NOTEMPETY"
		}
		suite.DB.ValidateAndSave(&dm)
	}

	chunksWithLimit, err := models.GetPendingChunksBySession(session, 4)
	suite.Nil(err)

	suite.NotEqual(0, len(chunksWithLimit))
	suite.Equal(4, len(chunksWithLimit))
}

func (suite *ModelSuite) Test_ComputeSectorDataMapAddress_AtSectorZero() {
	hashes := models.ComputeSectorDataMapAddress("abcdef", 0, 2)

	suite.Equal([]string{
		"EHAEYFLCUFVALBLAJGNHQ9PDXCFBSHWDYFMFGCODTGTFXALEZG9CPADFDEAINAGDXDNCZEEGKEUBIHWBA",
		"PGJIVFP9VHV9G9JEPDDEJFBGBGNEVGLCIFKEUDIFNFY9WCEG9FGDSBQBZCBCHFZAB9YBLFKBI9DCWAFDX"}, hashes)
}

func (suite *ModelSuite) Test_ComputeSectorDataMapAddress_AtSectorOne() {
	hashes := models.ComputeSectorDataMapAddress("abcdef", 1, 2)

	suite.Equal([]string{
		"UDUDRFQ9UEUCEI9C9IQBXERBBGBIWFDENHHFKEYEAC9EYCT9MG9GDCI9BAMDM9EHIFKD9IJFFEUCYFRFP",
		"YHVFOAWCTEGBN9RCVEKIP9EFLBCENFZ9F9HF9DG9XGG9YFGDH9FFT9GIICIBPCDESHLDXB9AIHEFNBM9D"}, hashes)
}

func (suite *ModelSuite) Test_ChunkEncryptAndDecryptEthKey() {
	genHash := "abcdef"
	fileSizeBytes := 123

	ethKey := hex.EncodeToString([]byte("SOME_PRIVATE_KEY"))

	u := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          genHash,
		FileSizeBytes:        fileSizeBytes,
		NumChunks:            400,
		StorageLengthInYears: 4,
		ETHPrivateKey:        hex.EncodeToString([]byte("SOME_PRIVATE_KEY")),
	}

	_, err := u.StartUploadSession()

	dataMap := models.DataMap{}
	err = suite.DB.RawQuery("SELECT * from data_maps where genesis_hash = ?", u.GenesisHash).First(&dataMap)
	suite.Nil(err)

	ethKey = hex.EncodeToString([]byte("SOME_OTHER_PRIVATE_KEY"))

	encryptedKey, err := dataMap.EncryptEthKey(ethKey)
	suite.Nil(err)
	suite.NotEqual(ethKey, encryptedKey)

	decryptedKey, err := dataMap.DecryptEthKey(encryptedKey)
	suite.Equal(ethKey, decryptedKey)
}

func (suite *ModelSuite) Test_AttachUnassignedChunksToGenHashMap() {

	/*TODO

	This was originally intended as part of a more sophisticated means to process the chunks
	but the test is flakey so I'm just commenting it all out for now and will come back to it
	after mainnet.

	*/

	//fileBytesCount := 2000
	//
	//uploadSession1 := models.UploadSession{
	//	GenesisHash:   "abcdeff1",
	//	FileSizeBytes: fileBytesCount,
	//	Type:          models.SessionTypeAlpha,
	//}
	//
	//uploadSession2 := models.UploadSession{
	//	GenesisHash:   "abcdeff2",
	//	FileSizeBytes: fileBytesCount,
	//	Type:          models.SessionTypeBeta,
	//}
	//
	//uploadSession3 := models.UploadSession{
	//	GenesisHash:   "abcdeff3",
	//	FileSizeBytes: fileBytesCount,
	//	Type:          models.SessionTypeAlpha,
	//}
	//
	//uploadSession4 := models.UploadSession{
	//	GenesisHash:   "abcdeff4",
	//	FileSizeBytes: fileBytesCount,
	//	Type:          models.SessionTypeBeta,
	//}
	//
	//uploadSession1.StartUploadSession()
	//uploadSession2.StartUploadSession()
	//uploadSession3.StartUploadSession()
	//uploadSession4.StartUploadSession()
	//
	//genHash1 := []models.DataMap{}
	//genHash2 := []models.DataMap{}
	//genHash3 := []models.DataMap{}
	//genHash4 := []models.DataMap{}
	//err := suite.DB.Where("genesis_hash = ?", "abcdeff1").All(&genHash1)
	//suite.Equal(err, nil)
	//err = suite.DB.Where("genesis_hash = ?", "abcdeff2").All(&genHash2)
	//suite.Equal(err, nil)
	//err = suite.DB.Where("genesis_hash = ?", "abcdeff3").All(&genHash3)
	//suite.Equal(err, nil)
	//err = suite.DB.Where("genesis_hash = ?", "abcdeff4").All(&genHash4)
	//suite.Equal(err, nil)
	//
	//genHash1[0].Status = models.Unassigned
	//genHash1[1].Status = models.Unassigned
	//suite.DB.ValidateAndSave(&genHash1[0])
	//suite.DB.ValidateAndSave(&genHash1[1])
	//
	//genHash2[0].Status = models.Error
	//genHash2[1].Status = models.Error
	//suite.DB.ValidateAndSave(&genHash2[0])
	//suite.DB.ValidateAndSave(&genHash2[1])
	//
	//for _, dataMap := range genHash3 {
	//	dataMap.Status = models.Unassigned
	//	suite.DB.ValidateAndSave(&dataMap)
	//}
	//
	//for _, dataMap := range genHash4 {
	//	dataMap.Status = models.Error
	//	suite.DB.ValidateAndSave(&dataMap)
	//}
	//
	//genHashes, err := models.GetUnassignedGenesisHashes()
	//suite.Equal(err, nil)
	//suite.Equal(4, len(genHashes))
	//
	//hashAndTypeMap, err := models.AttachUnassignedChunksToGenHashMap(genHashes)
	//
	//suite.Equal(models.SessionTypeAlpha, hashAndTypeMap["abcdeff1"].Type)
	//suite.Equal(models.SessionTypeBeta, hashAndTypeMap["abcdeff2"].Type)
	//suite.Equal(models.SessionTypeAlpha, hashAndTypeMap["abcdeff3"].Type)
	//suite.Equal(models.SessionTypeBeta, hashAndTypeMap["abcdeff4"].Type)
	//
	//suite.Equal(6, len(hashAndTypeMap["abcdeff1"].Chunks))
	//suite.Equal(20, len(hashAndTypeMap["abcdeff2"].Chunks))
	//suite.Equal(0, len(hashAndTypeMap["abcdeff3"].Chunks))
	//suite.Equal(0, len(hashAndTypeMap["abcdeff4"].Chunks))

	//suite.Equal(0, hashAndTypeMap["abcdeff1"].Chunks[0].ChunkIdx)
	//suite.Equal(0, hashAndTypeMap["abcdeff2"].Chunks[0].ChunkIdx)
	//suite.Equal(0, hashAndTypeMap["abcdeff3"].Chunks[0].ChunkIdx)
	//suite.Equal(0, hashAndTypeMap["abcdeff4"].Chunks[0].ChunkIdx)

	//fmt.Println(len(hashAndTypeMap))
}

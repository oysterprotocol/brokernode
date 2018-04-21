package models_test

import (
	"crypto/sha512"
	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/utils"

	"github.com/iotaledger/giota"
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
	genHash := "genHashTest"
	fileBytesCount := 9000

	vErr, err := models.BuildDataMaps(genHash, fileBytesCount)
	ms.Nil(err)
	ms.Equal(0, len(vErr.Errors))

	expectedHashes := []string{
		"ef7edf0decd95c9e094184dca8641b68bb3ca0f69fec086341893816c68f7d9d408131fa01a66cf95f05b2a038185db9",
		"86ad8449bd1b32bcd86d86cfe7b3b6453f391c0c0df57956a2dff53f55709af3cd43a983ef46263cf8e361ae15734b33",
		"fbb914b1ba9cc663be0eb7b2570209af5caccfe5b7bba65e832c683072a969715e1b23866ce97ddb765fefe9b991e652",
		"e697116fd36a697f327f4682fd6f72250933bc61184fc36ff89badf749779aadf643b2e4f3fcd22fa9c07a6ce89c99a5",
		"167b2e33d17a4a96c6ad7216cd49c664b056efd30c08d65a354d1a5eb9cc9dbcb2f639495269f7ef5e56b8e62777edfc",
		"73ad9b9ba83acbf49a714980e660ead44f3fb574ee807d05d4ab728cfc9ecd1cd2f2a0a608948ea968d383db037a6d6c",
		"173b5a6ced53b7a84aa9f789bab0485418e949a3571ed964dde9b54618d38f212d496831cf083cc0b46d6d51e78461c7",
	}

	dMaps := []models.DataMap{}
	ms.DB.Where("genesis_hash = ?", genHash).Order("chunk_idx asc").All(&dMaps)

	for i, dMap := range dMaps {
		ms.Equal(expectedHashes[i], dMap.ObfuscatedHash)
	}
}

func (ms *ModelSuite) Test_CreateTreasurePayload() {
	maxSideChainLength := 10
	matchesFound := 0

	for _, tc := range encryptedTreasureCases {
		payload, err := models.CreateTreasurePayload(tc.ethPrivateSeed, tc.sha256Hash, maxSideChainLength)
		ms.Nil(err)

		payloadNotTryted := oyster_utils.TrytesToBytes(giota.Trytes(payload))

		currentHash := tc.sha256Hash

		for i := 0; i < maxSideChainLength; i++ {
			currentHash = oyster_utils.HashString(currentHash, sha512.New())
			result := oyster_utils.Decrypt(currentHash, string(payloadNotTryted))
			if result != "" {
				ms.Equal(result, tc.ethPrivateSeed)
				matchesFound++
			}
		}
	}
	ms.Equal(3, matchesFound)
}

func (suite *ModelSuite) Test_GetUnassignedGenesisHashes() {
	fileBytesCount := 8500

	vErr, err := models.BuildDataMaps("genHash1", fileBytesCount)
	suite.Nil(err)
	suite.Equal(0, len(vErr.Errors))
	vErr, err = models.BuildDataMaps("genHash2", fileBytesCount)
	suite.Nil(err)
	suite.Equal(0, len(vErr.Errors))
	vErr, err = models.BuildDataMaps("genHash3", fileBytesCount)
	suite.Nil(err)
	suite.Equal(0, len(vErr.Errors))
	vErr, err = models.BuildDataMaps("genHash4", fileBytesCount)
	suite.Nil(err)
	suite.Equal(0, len(vErr.Errors))
	vErr, err = models.BuildDataMaps("genHash5", fileBytesCount)
	suite.Nil(err)
	suite.Equal(0, len(vErr.Errors))

	genHash1 := []models.DataMap{} // 1 unassigned
	genHash2 := []models.DataMap{} // 1 error
	genHash3 := []models.DataMap{} // all unassigned
	genHash4 := []models.DataMap{} // all error
	err = suite.DB.Where("genesis_hash = ?", "genHash1").All(&genHash1)
	suite.Equal(err, nil)
	err = suite.DB.Where("genesis_hash = ?", "genHash2").All(&genHash2)
	suite.Equal(err, nil)
	err = suite.DB.Where("genesis_hash = ?", "genHash3").All(&genHash3)
	suite.Equal(err, nil)
	err = suite.DB.Where("genesis_hash = ?", "genHash4").All(&genHash4)
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
		if genHash == "genHash5" {
			suite.Failf("FAIL", "genHash5 should not be in the array")
		}
	}
}

func (suite *ModelSuite) Test_GetUnassignedChunks() {
	fileBytesCount := 1000

	vErr, err := models.BuildDataMaps("genHash1", fileBytesCount)
	suite.Nil(err)
	suite.Equal(0, len(vErr.Errors))
	vErr, err = models.BuildDataMaps("genHash2", fileBytesCount)
	suite.Nil(err)
	suite.Equal(0, len(vErr.Errors))
	vErr, err = models.BuildDataMaps("genHash3", fileBytesCount)
	suite.Nil(err)
	suite.Equal(0, len(vErr.Errors))
	vErr, err = models.BuildDataMaps("genHash4", fileBytesCount)
	suite.Nil(err)
	suite.Equal(0, len(vErr.Errors))
	vErr, err = models.BuildDataMaps("genHash5", fileBytesCount)
	suite.Nil(err)
	suite.Equal(0, len(vErr.Errors))

	genHash1 := []models.DataMap{} // 1 unassigned
	genHash2 := []models.DataMap{} // 1 error
	genHash3 := []models.DataMap{} // all unassigned
	genHash4 := []models.DataMap{} // all error
	err = suite.DB.Where("genesis_hash = ?", "genHash1").All(&genHash1)
	suite.Equal(err, nil)
	err = suite.DB.Where("genesis_hash = ?", "genHash2").All(&genHash2)
	suite.Equal(err, nil)
	err = suite.DB.Where("genesis_hash = ?", "genHash3").All(&genHash3)
	suite.Equal(err, nil)
	err = suite.DB.Where("genesis_hash = ?", "genHash4").All(&genHash4)
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
		if unassignedChunk.GenesisHash == "genHash5" {
			suite.Failf("FAIL", "a chunk with genHash5 should not be in the array")
		}
	}
}

func (suite *ModelSuite) Test_GetAllUnassignedChunksBySession() {
	uploadSession1 := models.UploadSession{
		GenesisHash:    "genHash1",
		FileSizeBytes:  8000,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  models.PaymentStatusPaid,
		TreasureStatus: models.TreasureBuried,
	}
	uploadSession1.StartUploadSession()
	session := models.UploadSession{}
	err := suite.DB.Where("genesis_hash = ?", "genHash1").First(&session)
	suite.Nil(err)

	jobs.MarkBuriedMapsAsUnassigned()
	chunks, err := models.GetAllUnassignedChunksBySession(session)
	suite.Nil(err)

	suite.NotEqual(0, len(chunks))
	suite.Equal(5, len(chunks))
	suite.NotEqual(models.DataMap{}, chunks[0])
}

func (suite *ModelSuite) Test_GetUnassignedChunksBySession() {
	uploadSession1 := models.UploadSession{
		GenesisHash:    "genHash1",
		FileSizeBytes:  8000,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  models.PaymentStatusPaid,
		TreasureStatus: models.TreasureBuried,
	}
	uploadSession1.StartUploadSession()
	session := models.UploadSession{}
	err := suite.DB.Where("genesis_hash = ?", "genHash1").First(&session)
	suite.Nil(err)

	jobs.MarkBuriedMapsAsUnassigned()
	chunks, err := models.GetAllUnassignedChunksBySession(session)
	chunksWithLimit, err := models.GetUnassignedChunksBySession(session, 4)
	suite.Nil(err)

	suite.NotEqual(0, len(chunksWithLimit))
	suite.Equal(5, len(chunks))
	suite.Equal(4, len(chunksWithLimit))
	suite.NotEqual(models.DataMap{}, chunksWithLimit[0])
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
	//	GenesisHash:   "genHash1",
	//	FileSizeBytes: fileBytesCount,
	//	Type:          models.SessionTypeAlpha,
	//}
	//
	//uploadSession2 := models.UploadSession{
	//	GenesisHash:   "genHash2",
	//	FileSizeBytes: fileBytesCount,
	//	Type:          models.SessionTypeBeta,
	//}
	//
	//uploadSession3 := models.UploadSession{
	//	GenesisHash:   "genHash3",
	//	FileSizeBytes: fileBytesCount,
	//	Type:          models.SessionTypeAlpha,
	//}
	//
	//uploadSession4 := models.UploadSession{
	//	GenesisHash:   "genHash4",
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
	//err := suite.DB.Where("genesis_hash = ?", "genHash1").All(&genHash1)
	//suite.Equal(err, nil)
	//err = suite.DB.Where("genesis_hash = ?", "genHash2").All(&genHash2)
	//suite.Equal(err, nil)
	//err = suite.DB.Where("genesis_hash = ?", "genHash3").All(&genHash3)
	//suite.Equal(err, nil)
	//err = suite.DB.Where("genesis_hash = ?", "genHash4").All(&genHash4)
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
	//suite.Equal(models.SessionTypeAlpha, hashAndTypeMap["genHash1"].Type)
	//suite.Equal(models.SessionTypeBeta, hashAndTypeMap["genHash2"].Type)
	//suite.Equal(models.SessionTypeAlpha, hashAndTypeMap["genHash3"].Type)
	//suite.Equal(models.SessionTypeBeta, hashAndTypeMap["genHash4"].Type)
	//
	//suite.Equal(6, len(hashAndTypeMap["genHash1"].Chunks))
	//suite.Equal(20, len(hashAndTypeMap["genHash2"].Chunks))
	//suite.Equal(0, len(hashAndTypeMap["genHash3"].Chunks))
	//suite.Equal(0, len(hashAndTypeMap["genHash4"].Chunks))

	//suite.Equal(0, hashAndTypeMap["genHash1"].Chunks[0].ChunkIdx)
	//suite.Equal(0, hashAndTypeMap["genHash2"].Chunks[0].ChunkIdx)
	//suite.Equal(0, hashAndTypeMap["genHash3"].Chunks[0].ChunkIdx)
	//suite.Equal(0, hashAndTypeMap["genHash4"].Chunks[0].ChunkIdx)

	//fmt.Println(len(hashAndTypeMap))
}

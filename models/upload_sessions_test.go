package models_test

import (
	"encoding/hex"
	"fmt"
	"github.com/gobuffalo/pop/nulls"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"github.com/shopspring/decimal"
	"math/big"
	"time"
)

func (suite *ModelSuite) Test_BigFileSize() {
	fileSizeBytes := uint64(9223372036854775808) // 2^63+1, more than signed int64 range.
	u := models.UploadSession{
		GenesisHash:   "hello",
		NumChunks:     2,
		FileSizeBytes: fileSizeBytes,
	}

	validateAndCreate(suite, &u)

	uploadSession := models.UploadSession{}
	suite.Nil(suite.DB.Find(&uploadSession, u.ID))

	suite.Equal(uploadSession.FileSizeBytes, fileSizeBytes)
}

func (suite *ModelSuite) Test_StartUploadSession() {

	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	defer oyster_utils.ResetBrokerMode()

	genHash := "abcdef"
	fileSizeBytes := uint64(123)
	numChunks := 2
	storageLengthInYears := 2

	u := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          genHash,
		FileSizeBytes:        fileSizeBytes,
		NumChunks:            numChunks,
		StorageLengthInYears: storageLengthInYears,
	}

	vErr, err := u.StartUploadSession()
	suite.Nil(err)
	suite.False(vErr.HasAny())

	uSession := models.UploadSession{}
	suite.DB.Where("genesis_hash = ?", genHash).First(&uSession)

	suite.Equal(genHash, uSession.GenesisHash)
	suite.Equal(fileSizeBytes, uSession.FileSizeBytes)
	suite.Equal(numChunks+1, uSession.NumChunks)
	suite.Equal(models.SessionTypeAlpha, uSession.Type)
	suite.Equal(decimal.NewFromFloatWithExponent(0.03125, -5), uSession.TotalCost)
	suite.Equal(2, uSession.StorageLengthInYears)
}

func (suite *ModelSuite) Test_DataMapsForSession() {
	genHash := "abcdef"
	numChunks := 2
	storageLengthInYears := 3

	u := models.UploadSession{
		GenesisHash:          genHash,
		FileSizeBytes:        123,
		NumChunks:            numChunks,
		StorageLengthInYears: storageLengthInYears,
	}

	vErr, err := u.StartUploadSession()
	suite.Nil(err)
	suite.False(vErr.HasAny())

	expectedHashes := []string{
		"dd88bb5db7314227c7e6117c693ceb83bbaf587bd1b63393d7512ba68bf42973845fa1c2924be14d37ba2da1938d7228",
		"cdfdb810ee1607917c8bacbfbf95d35dab9281abb01968c2a27349476b53aa35024fae410955327233523229677da827",
		"d5a3eda969c62842840e58fe7a1982fdcf9eb758e2ebd545289d6daa706b506a6a4833cd134992be9c73fe4c1e1d15ff",
	}

	dMaps, err := u.DataMapsForSession()
	suite.Nil(err)

	for i, dMap := range *dMaps {
		suite.Equal(expectedHashes[i], dMap.ObfuscatedHash)
	}
}

func (suite *ModelSuite) Test_TreasureMapGetterAndSetter() {
	genHash := "abcdef"
	numChunks := 2
	storageLengthInYears := 3

	// This map seems pointless but it makes the testing
	// in the for loop later on a bit simpler
	t := map[int]models.TreasureMap{}
	t[5] = models.TreasureMap{
		Sector: 1,
		Idx:    5,
		Key:    "firstKey",
	}
	t[78] = models.TreasureMap{
		Sector: 2,
		Idx:    78,
		Key:    "secondKey",
	}

	treasureIndexArray := make([]models.TreasureMap, 0)
	treasureIndexArray = append(treasureIndexArray, t[5])
	treasureIndexArray = append(treasureIndexArray, t[78])

	// do not format this.  It needs to not have new lines in it
	testMap := `[{"sector":` + fmt.Sprint(t[5].Sector) + `,"idx":` + fmt.Sprint(t[5].Idx) + `,"key":"` + fmt.Sprint(t[5].Key) + `"},{"sector":` + fmt.Sprint(t[78].Sector) + `,"idx":` + fmt.Sprint(t[78].Idx) + `,"key":"` + fmt.Sprint(t[78].Key) + `"}]`

	u := models.UploadSession{
		GenesisHash:          genHash,
		FileSizeBytes:        123,
		NumChunks:            numChunks,
		StorageLengthInYears: storageLengthInYears,
	}

	vErr, err := u.StartUploadSession()
	suite.Nil(err)
	suite.False(vErr.HasAny())

	suite.Nil(u.SetTreasureMap(treasureIndexArray))

	treasureIdxMap, err := u.GetTreasureMap()
	suite.Nil(err)

	session := models.UploadSession{}
	err = suite.DB.Where("genesis_hash = ?", u.GenesisHash).First(&session)

	suite.Equal(testMap, session.TreasureIdxMap.String)

	suite.Equal(treasureIndexArray, treasureIdxMap)
	suite.Equal(2, len(treasureIdxMap))

	for _, entry := range treasureIdxMap {
		_, ok := t[entry.Idx]
		suite.True(ok)
		suite.Equal(entry.Sector, t[entry.Idx].Sector)
		suite.Equal(entry.Key, t[entry.Idx].Key)
		suite.Equal(entry.Idx, t[entry.Idx].Idx)
	}
}

func (suite *ModelSuite) Test_GetSessionsByAge() {

	err := suite.DB.RawQuery("DELETE FROM upload_sessions").All(&[]models.UploadSession{})
	suite.Nil(err)

	uploadSession1 := models.UploadSession{
		GenesisHash:    "abcdeff1",
		FileSizeBytes:  5000,
		NumChunks:      7,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureInDataMapComplete,
	}
	uploadSession2 := models.UploadSession{ // this one will be newest and last in the array
		GenesisHash:    "abcdeff2",
		FileSizeBytes:  5000,
		NumChunks:      7,
		Type:           models.SessionTypeBeta,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureInDataMapComplete,
	}
	uploadSession3 := models.UploadSession{ // this one will be oldest and first in the array
		GenesisHash:    "abcdeff3",
		FileSizeBytes:  5000,
		NumChunks:      7,
		Type:           models.SessionTypeBeta,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureInDataMapComplete,
	}
	uploadSession4 := models.UploadSession{ // will not be in the array
		GenesisHash:    "abcdeff4",
		FileSizeBytes:  5000,
		NumChunks:      7,
		Type:           models.SessionTypeBeta,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureInDataMapPending,
	}
	uploadSession5 := models.UploadSession{ // will not be in the array
		GenesisHash:    "abcdeff5",
		FileSizeBytes:  5000,
		NumChunks:      7,
		Type:           models.SessionTypeBeta,
		PaymentStatus:  models.PaymentStatusPending,
		TreasureStatus: models.TreasureInDataMapPending,
	}

	vErr, err := uploadSession1.StartUploadSession()
	suite.Nil(err)
	suite.False(vErr.HasAny())
	vErr, err = uploadSession2.StartUploadSession()
	suite.Nil(err)
	suite.False(vErr.HasAny())
	vErr, err = uploadSession3.StartUploadSession()
	suite.Nil(err)
	suite.False(vErr.HasAny())
	vErr, err = uploadSession4.StartUploadSession()
	suite.Nil(err)
	suite.False(vErr.HasAny())
	vErr, err = uploadSession5.StartUploadSession()
	suite.Nil(err)
	suite.False(vErr.HasAny())

	// set uploadSession3 to be the oldest
	err = suite.DB.RawQuery("UPDATE upload_sessions SET created_at = ? WHERE genesis_hash = ?",
		time.Now().Add(-10*time.Second), "abcdeff3").All(&[]models.UploadSession{})

	// set uploadSession2 to be the newest
	err = suite.DB.RawQuery("UPDATE upload_sessions SET created_at = ? WHERE genesis_hash = ?",
		time.Now().Add(10*time.Second), "abcdeff2").All(&[]models.UploadSession{})

	sessions, err := models.GetSessionsByAge()
	suite.Nil(err)

	//verify that the oldest session (uploadSession3) is first in the array
	suite.Equal("abcdeff3", sessions[0].GenesisHash)
	suite.Equal("abcdeff1", sessions[1].GenesisHash)
	suite.Equal("abcdeff2", sessions[2].GenesisHash)
	suite.Equal(3, len(sessions))
}

func (suite *ModelSuite) Test_MakeTreasureIdxMap() {

	defer oyster_utils.ResetBrokerMode()
	oyster_utils.SetBrokerMode(oyster_utils.TestModeDummyTreasure)

	sectorSize := 100

	genHash := "abcdef"
	numChunks := 250
	storageLengthInYears := 3
	alphaIndexes := []int{2, 121, 245}
	betaIndexes := []int{9, 89, 230}

	u := models.UploadSession{
		GenesisHash:          genHash,
		FileSizeBytes:        123,
		NumChunks:            numChunks,
		StorageLengthInYears: storageLengthInYears,
	}

	vErr, err := u.StartUploadSession()
	mergedIndexes, err := oyster_utils.MergeIndexes(alphaIndexes, betaIndexes, sectorSize, numChunks)

	suite.Nil(err)
	privateKeys := []string{
		"9999999999999999999999999999999999999999999999999999999999999999",
		"9999999999999999999999999999999999999999999999999999999999999999",
		"9999999999999999999999999999999999999999999999999999999999999999",
	}

	u.MakeTreasureIdxMap(mergedIndexes, privateKeys)

	treasureIdxMap, err := u.GetTreasureMap()
	suite.Nil(err)
	suite.False(vErr.HasAny())

	suite.Equal(0, treasureIdxMap[0].Sector)
	suite.Equal(1, treasureIdxMap[1].Sector)
	suite.Equal(2, treasureIdxMap[2].Sector)

	// This will break anytime we change the hashing method.
	suite.Equal(68, treasureIdxMap[0].Idx)
	suite.Equal(148, treasureIdxMap[1].Idx)
	suite.Equal(210, treasureIdxMap[2].Idx)
}

func (suite *ModelSuite) Test_GetTreasureIndexes() {

	defer oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	oyster_utils.SetBrokerMode(oyster_utils.TestModeDummyTreasure)

	sectorSize := 100

	genHash := "abcdef"
	numChunks := 250
	storageLengthInYears := 3
	alphaIndexes := []int{2, 121, 245}
	betaIndexes := []int{9, 89, 230}

	u := models.UploadSession{
		GenesisHash:          genHash,
		FileSizeBytes:        123,
		NumChunks:            numChunks,
		StorageLengthInYears: storageLengthInYears,
	}

	expectedIndexes := make([]int, 0)
	expectedIndexes = append(expectedIndexes, 68)
	expectedIndexes = append(expectedIndexes, 148)
	expectedIndexes = append(expectedIndexes, 210)

	vErr, err := u.StartUploadSession()
	suite.Nil(err)
	suite.False(vErr.HasAny())

	mergedIndexes, err := oyster_utils.MergeIndexes(alphaIndexes, betaIndexes, sectorSize, numChunks)
	suite.Nil(err)
	privateKeys, err := services.EthWrapper.GenerateKeys(len(mergedIndexes))
	suite.Nil(err)
	u.MakeTreasureIdxMap(mergedIndexes, privateKeys)
	actualIndexes, err := u.GetTreasureIndexes()

	// This will break anytime we change the hashing method.
	suite.Equal(expectedIndexes, actualIndexes)
}

func (suite *ModelSuite) Test_EncryptAndDecryptEthKey() {
	genHash := "abcdef"
	ethKey := hex.EncodeToString([]byte("SOME_PRIVATE_KEY"))

	u := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          genHash,
		FileSizeBytes:        123,
		NumChunks:            400,
		StorageLengthInYears: 4,
		ETHPrivateKey:        ethKey,
	}

	vErr, err := u.StartUploadSession()
	suite.Nil(err)
	suite.False(vErr.HasAny())

	suite.NotEqual(ethKey, u.ETHPrivateKey) // it should be encrypted by now

	decryptedKey := u.DecryptSessionEthKey()

	suite.Equal(ethKey, decryptedKey)
}

func (suite *ModelSuite) Test_CalculatePayment_Less_Than_1_GB() {

	currentStoragePeg := models.StoragePeg

	defer func() { models.StoragePeg = currentStoragePeg }()

	models.StoragePeg = decimal.NewFromFloat(float64(64))
	storageLengthInYears := 3

	u := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          "abcdef",
		NumChunks:            2,
		FileSizeBytes:        9000000,
		StorageLengthInYears: storageLengthInYears,
	}

	vErr, err := u.StartUploadSession()
	invoice := u.GetInvoice()
	suite.Nil(err)
	suite.False(vErr.HasAny())

	// expecting to be charged for 1 full sector even though we aren't using the whole sector
	suite.Equal(decimal.New(468750000000000, -16), invoice.Cost)
}

func (suite *ModelSuite) Test_CalculatePayment_Greater_Than_1_GB() {

	defer oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	oyster_utils.SetBrokerMode(oyster_utils.TestModeDummyTreasure)

	currentStoragePeg := models.StoragePeg

	defer func() { models.StoragePeg = currentStoragePeg }()

	models.StoragePeg = decimal.NewFromFloat(float64(64))
	storageLengthInYears := 3

	u := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          "abcdef",
		NumChunks:            2,
		FileSizeBytes:        1500000000,
		StorageLengthInYears: storageLengthInYears,
	}

	vErr, err := u.StartUploadSession()
	invoice := u.GetInvoice()
	suite.Nil(err)
	suite.False(vErr.HasAny())

	// expecting to be charged for 2 full sectors even though we're only using 1.5
	suite.Equal(decimal.New(937500000000000, -16), invoice.Cost)
}

func (suite *ModelSuite) Test_CalculatePayment_1_Chunk_Less_Than_2_GB() {

	defer oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	oyster_utils.SetBrokerMode(oyster_utils.TestModeDummyTreasure)

	currentStoragePeg := models.StoragePeg

	defer func() { models.StoragePeg = currentStoragePeg }()

	fileSizeBytes := uint64(1999999 * oyster_utils.FileChunkSizeInByte)

	models.StoragePeg = decimal.NewFromFloat(float64(64))
	storageLengthInYears := 3

	u := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          "abcdef",
		NumChunks:            2,
		FileSizeBytes:        fileSizeBytes,
		StorageLengthInYears: storageLengthInYears,
	}

	vErr, err := u.StartUploadSession()
	invoice := u.GetInvoice()
	suite.Nil(err)
	suite.False(vErr.HasAny())

	// expecting to be charged for 2 sectors
	suite.Equal(decimal.New(937500000000000, -16), invoice.Cost)
}

func (suite *ModelSuite) Test_CalculatePayment_2_GB() {

	defer oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	oyster_utils.SetBrokerMode(oyster_utils.TestModeDummyTreasure)

	currentStoragePeg := models.StoragePeg

	defer func() { models.StoragePeg = currentStoragePeg }()

	fileSizeBytes := uint64(2000000 * oyster_utils.FileChunkSizeInByte)

	models.StoragePeg = decimal.NewFromFloat(float64(64))
	storageLengthInYears := 3

	u := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          "abcdef",
		NumChunks:            2,
		FileSizeBytes:        fileSizeBytes,
		StorageLengthInYears: storageLengthInYears,
	}

	vErr, err := u.StartUploadSession()
	invoice := u.GetInvoice()
	suite.Nil(err)
	suite.False(vErr.HasAny())

	// expecting to be charged for 3 sectors
	// we are 1 chunk over
	suite.Equal(decimal.New(1406250000000000, -16), invoice.Cost)
}

func (suite *ModelSuite) Test_GetPRLsPerTreasure() {
	defer oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	oyster_utils.SetBrokerMode(oyster_utils.TestModeDummyTreasure)

	totalCost := 5
	numSectors := 3

	genHash := "abcdef"
	numChunks := 250
	storageLengthInYears := 3
	mergedIndexes := []int{2, 121, 245}
	privateKeys := []string{
		"9999999999999999999999999999999999999999999999999999999999999999",
		"9999999999999999999999999999999999999999999999999999999999999999",
		"9999999999999999999999999999999999999999999999999999999999999999",
	}

	u := models.UploadSession{
		GenesisHash:          genHash,
		FileSizeBytes:        123,
		NumChunks:            numChunks,
		StorageLengthInYears: storageLengthInYears,
	}

	u.StartUploadSession()
	u.MakeTreasureIdxMap(mergedIndexes, privateKeys)
	u.NumChunks = 2500000
	u.TotalCost = decimal.NewFromFloat(float64(totalCost))
	suite.DB.ValidateAndUpdate(&u)

	prlsPerTreasure, err := u.GetPRLsPerTreasure()
	suite.Nil(err)

	expectedPRLsPerTreasure := new(big.Float).Quo(
		new(big.Float).SetInt(big.NewInt(int64(totalCost))),
		new(big.Float).SetInt(big.NewInt(int64(numSectors*2))))
	// multiplying numSectors x2, since brokers get to keep half the PRL

	suite.Equal(expectedPRLsPerTreasure, prlsPerTreasure)
}

func (suite *ModelSuite) Test_PaymentStatus() {
	u := models.UploadSession{}

	u.PaymentStatus = models.PaymentStatusConfirmed
	suite.Equal("confirmed", u.GetPaymentStatus())

	u.PaymentStatus = models.PaymentStatusInvoiced
	suite.Equal("invoiced", u.GetPaymentStatus())

	u.PaymentStatus = models.PaymentStatusPending
	suite.Equal("pending", u.GetPaymentStatus())

	u.PaymentStatus = 100
	suite.Equal("error", u.GetPaymentStatus())

	u.PaymentStatus = models.PaymentStatusError
	suite.Equal("error", u.GetPaymentStatus())
}

func (suite *ModelSuite) Test_SetBrokerTransactionToPaid() {

	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	defer oyster_utils.ResetBrokerMode()

	genHash := "abcdef"
	fileSizeBytes := 123
	numChunks := 2
	storageLengthInYears := 2
	privateKey := "abcdef1234567890"
	startingEthAddr := "0000000000"

	u := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          genHash,
		FileSizeBytes:        uint64(fileSizeBytes),
		NumChunks:            numChunks,
		StorageLengthInYears: storageLengthInYears,
		ETHPrivateKey:        privateKey,
		ETHAddrAlpha:         nulls.String{string(startingEthAddr), true},
		ETHAddrBeta:          nulls.String{string(startingEthAddr), true},
		TotalCost:            totalCost,
	}

	vErr, err := u.StartUploadSession()
	suite.Nil(err)
	suite.Equal(0, len(vErr.Errors))

	uSession := models.UploadSession{}
	suite.DB.Where("genesis_hash = ?", genHash).All(&uSession)

	models.NewBrokerBrokerTransaction(&uSession)

	brokerTxs := returnAllBrokerBrokerTxs(suite)
	suite.Equal(1, len(brokerTxs))

	suite.Equal(models.BrokerTxAlphaPaymentPending, brokerTxs[0].PaymentStatus)

	models.SetBrokerTransactionToPaid(uSession)

	brokerTxs = returnAllBrokerBrokerTxs(suite)
	suite.Equal(1, len(brokerTxs))

	suite.Equal(models.BrokerTxAlphaPaymentConfirmed, brokerTxs[0].PaymentStatus)
}

func (suite *ModelSuite) Test_BulkMarkDataMapsAsUnassigned_BeforeMigration() {
	// Setup
	genesisHash := "beforeMigration"

	dbOperation, _ := oyster_utils.CreateDbUpdateOperation(&models.DataMap{})
	columnsName := dbOperation.GetColumns()

	dataMap := models.DataMap{
		GenesisHash: genesisHash,
		Status:      models.Pending,
		ChunkIdx:    0,
		Hash:        "0",
		MsgID:       "msg_0",
		Message:     "abc",
		MsgStatus:   models.MsgStatusUnmigrated,
	})
	suite.Nil(
		suite.DB.RawQuery(fmt.Sprintf("INSERT INTO data_maps (%s) VALUES %s", 
			columnsName, dbOperation.GetNewInsertedValue(dataMap))).All(&[]DataMap{}))
	
	dataMap = models.DataMap{
		GenesisHash: genesisHash,
		Status:      models.Pending,
		ChunkIdx:    1,
		Hash:        "1",
		MsgID:       "msg_1",
		Message:     "123",
		MsgStatus:   models.MsgStatusUnmigrated,
	})
	suite.Nil(
		suite.DB.RawQuery(fmt.Sprintf("INSERT INTO data_maps (%s) VALUES %s", 
			columnsName, dbOperation.GetNewInsertedValue(dataMap))).All(&[]DataMap{}))

	dm := []models.DataMap{}
	suite.Nil(
		suite.DB.RawQuery("SELECT * FROM data_maps WHERE status = ? AND genesis_hash = ?",
			models.Pending,
			genesisHash).All(&dm))
	suite.Equal(2, len(dm))

	u := models.UploadSession{
		GenesisHash: genesisHash,
	}

	// Testing
	suite.Nil(u.BulkMarkDataMapsAsUnassigned())

	// Assert
	dm = []models.DataMap{}
	suite.Nil(
		suite.DB.RawQuery("SELECT * FROM data_maps WHERE status = ? AND genesis_hash = ?",
			models.Unassigned,
			genesisHash).All(&dm))
	suite.Equal(2, len(dm))
}

func (suite *ModelSuite) Test_BulkMarkDataMapsAsUnassigned_AfterMigration() {
	// Setup
	genesisHash := "afterMigration"

	validateAndCreate(suite, &models.DataMap{
		GenesisHash: genesisHash,
		ChunkIdx:    0,
		Hash:        "0",
		MsgID:       "msg_0",
		Status:      models.Pending,
		MsgStatus:   models.MsgStatusUploadedNoNeedEncode,
	})

	validateAndCreate(suite, &models.DataMap{
		GenesisHash: genesisHash,
		ChunkIdx:    1,
		Hash:        "1",
		MsgID:       "msg_1",
		Status:      models.Pending,
		MsgStatus:   models.MsgStatusUploadedHaveNotEncoded,
	})

	dm := []models.DataMap{}
	suite.Nil(
		suite.DB.RawQuery("SELECT * FROM data_maps WHERE status = ? AND genesis_hash = ?",
			models.Pending,
			genesisHash).All(&dm))
	suite.Equal(2, len(dm))

	u := models.UploadSession{
		GenesisHash: genesisHash,
	}

	// Testing
	suite.Nil(u.BulkMarkDataMapsAsUnassigned())

	// Assert
	dm = []models.DataMap{}
	suite.Nil(
		suite.DB.RawQuery("SELECT * FROM data_maps WHERE status = ? AND genesis_hash = ?",
			models.Unassigned,
			genesisHash).All(&dm))
	suite.Equal(2, len(dm))
}

func validateAndCreate(suite *ModelSuite, model interface{}) {
	vErr, err := suite.DB.ValidateAndCreate(model)
	suite.Nil(err)
	suite.False(vErr.HasAny())
}

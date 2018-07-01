package models_test

import (
	"encoding/hex"
	"fmt"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"github.com/shopspring/decimal"
	"math/big"
	"time"
)

func (ms *ModelSuite) Test_BigFileSize() {
	u := models.UploadSession{
		GenesisHash:   "hello",
		NumChunks:     2,
		FileSizeBytes: 9223372036854775808, // 2^63+1, if signed,
	}

	vErr, err := ms.DB.ValidateAndCreate(&u)
	ms.Nil(err)
	ms.False(vErr.HasAny())
}

func (ms *ModelSuite) Test_StartUploadSession() {
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
	ms.Nil(err)
	ms.False(vErr.HasAny())

	uSession := models.UploadSession{}
	ms.DB.Where("genesis_hash = ?", genHash).First(&uSession)

	ms.Equal(genHash, uSession.GenesisHash)
	ms.Equal(fileSizeBytes, uSession.FileSizeBytes)
	ms.Equal(numChunks+1, uSession.NumChunks)
	ms.Equal(models.SessionTypeAlpha, uSession.Type)
	ms.Equal(decimal.NewFromFloatWithExponent(0.03125, -5), uSession.TotalCost)
	ms.Equal(2, uSession.StorageLengthInYears)
}

func (ms *ModelSuite) Test_DataMapsForSession() {
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
	ms.Nil(err)
	ms.False(vErr.HasAny())

	expectedHashes := []string{
		"dd88bb5db7314227c7e6117c693ceb83bbaf587bd1b63393d7512ba68bf42973845fa1c2924be14d37ba2da1938d7228",
		"cdfdb810ee1607917c8bacbfbf95d35dab9281abb01968c2a27349476b53aa35024fae410955327233523229677da827",
		"d5a3eda969c62842840e58fe7a1982fdcf9eb758e2ebd545289d6daa706b506a6a4833cd134992be9c73fe4c1e1d15ff",
	}

	dMaps, err := u.DataMapsForSession()
	ms.Nil(err)

	for i, dMap := range *dMaps {
		ms.Equal(expectedHashes[i], dMap.ObfuscatedHash)
	}
}

func (ms *ModelSuite) Test_TreasureMapGetterAndSetter() {
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
	ms.Nil(err)
	ms.False(vErr.HasAny())

	ms.Nil(u.SetTreasureMap(treasureIndexArray))

	treasureIdxMap, err := u.GetTreasureMap()
	ms.Nil(err)

	session := models.UploadSession{}
	err = ms.DB.Where("genesis_hash = ?", u.GenesisHash).First(&session)

	ms.Equal(testMap, session.TreasureIdxMap.String)

	ms.Equal(treasureIndexArray, treasureIdxMap)
	ms.Equal(2, len(treasureIdxMap))

	for _, entry := range treasureIdxMap {
		_, ok := t[entry.Idx]
		ms.Equal(true, ok)
		ms.Equal(t[entry.Idx].Sector, entry.Sector)
		ms.Equal(t[entry.Idx].Key, entry.Key)
		ms.Equal(t[entry.Idx].Idx, entry.Idx)
	}
}

func (ms *ModelSuite) Test_GetSessionsByAge() {

	err := ms.DB.RawQuery("DELETE from upload_sessions").All(&[]models.UploadSession{})
	ms.Nil(err)

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
	ms.Nil(err)
	ms.False(vErr.HasAny())
	vErr, err = uploadSession2.StartUploadSession()
	ms.Nil(err)
	ms.False(vErr.HasAny())
	vErr, err = uploadSession3.StartUploadSession()
	ms.Nil(err)
	ms.False(vErr.HasAny())
	vErr, err = uploadSession4.StartUploadSession()
	ms.Nil(err)
	ms.False(vErr.HasAny())
	vErr, err = uploadSession5.StartUploadSession()
	ms.Nil(err)
	ms.False(vErr.HasAny())

	// set uploadSession3 to be the oldest
	err = ms.DB.RawQuery("UPDATE upload_sessions SET created_at = ? WHERE genesis_hash = ?",
		time.Now().Add(-10*time.Second), "abcdeff3").All(&[]models.UploadSession{})

	// set uploadSession2 to be the newest
	err = ms.DB.RawQuery("UPDATE upload_sessions SET created_at = ? WHERE genesis_hash = ?",
		time.Now().Add(10*time.Second), "abcdeff2").All(&[]models.UploadSession{})

	sessions, err := models.GetSessionsByAge()
	ms.Nil(err)

	//verify that the oldest session (uploadSession3) is first in the array
	ms.Equal("abcdeff3", sessions[0].GenesisHash)
	ms.Equal("abcdeff1", sessions[1].GenesisHash)
	ms.Equal("abcdeff2", sessions[2].GenesisHash)
	ms.Equal(3, len(sessions))
}

func (ms *ModelSuite) Test_MakeTreasureIdxMap() {

	defer oyster_utils.ResetBrokerMode()
	oyster_utils.SetBrokerMode(oyster_utils.TestModeDummyTreasure)

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
	mergedIndexes, err := oyster_utils.MergeIndexes(alphaIndexes, betaIndexes)

	ms.Nil(err)
	privateKeys := []string{
		"9999999999999999999999999999999999999999999999999999999999999999",
		"9999999999999999999999999999999999999999999999999999999999999999",
		"9999999999999999999999999999999999999999999999999999999999999999",
	}

	u.MakeTreasureIdxMap(mergedIndexes, privateKeys)

	treasureIdxMap, err := u.GetTreasureMap()
	ms.Nil(err)
	ms.False(vErr.HasAny())

	ms.Equal(0, treasureIdxMap[0].Sector)
	ms.Equal(1, treasureIdxMap[1].Sector)
	ms.Equal(2, treasureIdxMap[2].Sector)

	// When we update MergeIndexes to hash the indexes, this test will start failing
	ms.Equal(5, treasureIdxMap[0].Idx)
	ms.Equal(105, treasureIdxMap[1].Idx)
	ms.Equal(237, treasureIdxMap[2].Idx)
}

func (ms *ModelSuite) Test_GetTreasureIndexes() {

	defer oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	oyster_utils.SetBrokerMode(oyster_utils.TestModeDummyTreasure)

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
	expectedIndexes = append(expectedIndexes, 5)
	expectedIndexes = append(expectedIndexes, 105)
	expectedIndexes = append(expectedIndexes, 237)

	vErr, err := u.StartUploadSession()
	ms.Nil(err)
	ms.False(vErr.HasAny())

	mergedIndexes, err := oyster_utils.MergeIndexes(alphaIndexes, betaIndexes)
	ms.Nil(err)
	privateKeys, err := services.EthWrapper.GenerateKeys(len(mergedIndexes))
	ms.Nil(err)
	u.MakeTreasureIdxMap(mergedIndexes, privateKeys)
	actualIndexes, err := u.GetTreasureIndexes()

	// When we update MergeIndexes to hash the indexes, this test will start failing
	ms.Equal(expectedIndexes, actualIndexes)
}

func (ms *ModelSuite) Test_EncryptAndDecryptEthKey() {
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
	ms.Nil(err)
	ms.False(vErr.HasAny())

	ms.NotEqual(ethKey, u.ETHPrivateKey) // it should be encrypted by now

	decryptedKey := u.DecryptSessionEthKey()

	ms.Equal(ethKey, decryptedKey)
}

func (ms *ModelSuite) Test_CalculatePayment_Less_Than_1_GB() {

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
	ms.Nil(err)
	ms.False(vErr.HasAny())

	// expecting to be charged for 1 full sector even though we aren't using the whole sector
	ms.Equal(decimal.New(468750000000000, -16), invoice.Cost)
}

func (ms *ModelSuite) Test_CalculatePayment_Greater_Than_1_GB() {

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
	ms.Nil(err)
	ms.False(vErr.HasAny())

	// expecting to be charged for 2 full sectors even though we're only using 1.5
	ms.Equal(decimal.New(937500000000000, -16), invoice.Cost)
}

func (ms *ModelSuite) Test_CalculatePayment_1_Chunk_Less_Than_2_GB() {

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
	ms.Nil(err)
	ms.False(vErr.HasAny())

	// expecting to be charged for 2 sectors
	ms.Equal(decimal.New(937500000000000, -16), invoice.Cost)
}

func (ms *ModelSuite) Test_CalculatePayment_2_GB() {

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
	ms.Nil(err)
	ms.False(vErr.HasAny())

	// expecting to be charged for 3 sectors
	// we are 1 chunk over
	ms.Equal(decimal.New(1406250000000000, -16), invoice.Cost)
}

func (ms *ModelSuite) Test_GetPRLsPerTreasure() {
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
	ms.DB.ValidateAndUpdate(&u)

	prlsPerTreasure, err := u.GetPRLsPerTreasure()
	ms.Nil(err)

	expectedPRLsPerTreasure := new(big.Float).Quo(
		new(big.Float).SetInt(big.NewInt(int64(totalCost))),
		new(big.Float).SetInt(big.NewInt(int64(numSectors*2))))
	// multiplying numSectors x2, since brokers get to keep half the PRL

	ms.Equal(expectedPRLsPerTreasure, prlsPerTreasure)
}

func (ms *ModelSuite) Test_PaymentStatus() {
	u := models.UploadSession{}

	u.PaymentStatus = models.PaymentStatusConfirmed
	ms.Equal(u.GetPaymentStatus(), "confirmed")

	u.PaymentStatus = models.PaymentStatusInvoiced
	ms.Equal(u.GetPaymentStatus(), "invoiced")

	u.PaymentStatus = models.PaymentStatusPending
	ms.Equal(u.GetPaymentStatus(), "pending")

	u.PaymentStatus = 100
	ms.Equal(u.GetPaymentStatus(), "error")

	u.PaymentStatus = models.PaymentStatusError
	ms.Equal(u.GetPaymentStatus(), "error")
}

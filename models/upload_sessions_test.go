package models_test

import (
	"encoding/hex"
	"fmt"
	"github.com/gobuffalo/pop/nulls"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/utils"
	"github.com/shopspring/decimal"
	"math/big"
	"strconv"
	"time"
)

func (suite *ModelSuite) Test_BigFileSize() {
	fileSizeBytes := uint64(9223372036854775808) // 2^63+1, more than signed int64 range.
	u := models.UploadSession{
		GenesisHash:   "hello",
		NumChunks:     2,
		FileSizeBytes: fileSizeBytes,
	}

	vErr, err := suite.DB.ValidateAndCreate(&u)
	suite.Nil(err)
	suite.False(vErr.HasAny())

	uploadSession := models.UploadSession{}
	suite.Nil(suite.DB.Find(&uploadSession, u.ID))

	suite.Equal(uploadSession.FileSizeBytes, fileSizeBytes)
}

func (suite *ModelSuite) Test_StartUploadSession() {

	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	defer oyster_utils.ResetBrokerMode()

	fileSizeBytes := uint64(123)
	numChunks := 2
	storageLengthInYears := 2

	u := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        fileSizeBytes,
		NumChunks:            numChunks,
		StorageLengthInYears: storageLengthInYears,
	}

	vErr, err := u.StartUploadSession()
	suite.False(vErr.HasAny())
	suite.Nil(err)

	suite.Nil(err)
	suite.False(vErr.HasAny())

	uSession := models.UploadSession{}
	suite.DB.Where("genesis_hash = ?", u.GenesisHash).First(&uSession)

	suite.Equal(u.GenesisHash, uSession.GenesisHash)
	suite.Equal(fileSizeBytes, uSession.FileSizeBytes)
	suite.Equal(numChunks+1, uSession.NumChunks)
	suite.Equal(models.SessionTypeAlpha, uSession.Type)
	suite.Equal(decimal.NewFromFloatWithExponent(0.03125, -5), uSession.TotalCost)
	suite.Equal(2, uSession.StorageLengthInYears)

	// verify indexes for alpha session
	suite.Equal(int64(0), uSession.NextIdxToVerify)
	suite.Equal(int64(0), uSession.NextIdxToAttach)

	u2 := models.UploadSession{
		Type:                 models.SessionTypeBeta,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        fileSizeBytes,
		NumChunks:            numChunks,
		StorageLengthInYears: storageLengthInYears,
	}

	vErr, err = u2.StartUploadSession()
	suite.False(vErr.HasAny())
	suite.Nil(err)

	suite.Nil(err)
	suite.False(vErr.HasAny())

	uSession2 := models.UploadSession{}
	suite.DB.Where("genesis_hash = ?", u2.GenesisHash).First(&uSession2)

	// verify indexes for beta session
	suite.Equal(int64(uSession2.NumChunks-1), uSession2.NextIdxToVerify)
	suite.Equal(int64(uSession2.NumChunks-1), uSession2.NextIdxToAttach)
}

func (suite *ModelSuite) Test_TreasureMapGetterAndSetter() {
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
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        123,
		NumChunks:            numChunks,
		StorageLengthInYears: storageLengthInYears,
	}

	vErr, err := u.StartUploadSession()
	suite.False(vErr.HasAny())
	suite.Nil(err)

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
		GenesisHash:    oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:  5000,
		NumChunks:      7,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureInDataMapComplete,
		AllDataReady:   models.AllDataReady,
	}
	uploadSession2 := models.UploadSession{ // this one will be newest and last in the array
		GenesisHash:    oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:  5000,
		NumChunks:      7,
		Type:           models.SessionTypeBeta,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureInDataMapComplete,
		AllDataReady:   models.AllDataReady,
	}
	uploadSession3 := models.UploadSession{ // this one will be oldest and first in the array
		GenesisHash:    oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:  5000,
		NumChunks:      7,
		Type:           models.SessionTypeBeta,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureInDataMapComplete,
		AllDataReady:   models.AllDataReady,
	}
	uploadSession4 := models.UploadSession{ // will not be in the array
		GenesisHash:    oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:  5000,
		NumChunks:      7,
		Type:           models.SessionTypeBeta,
		PaymentStatus:  models.PaymentStatusConfirmed,
		TreasureStatus: models.TreasureInDataMapPending,
		AllDataReady:   models.AllDataReady,
	}
	uploadSession5 := models.UploadSession{ // will not be in the array
		GenesisHash:    oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:  5000,
		NumChunks:      7,
		Type:           models.SessionTypeBeta,
		PaymentStatus:  models.PaymentStatusPending,
		TreasureStatus: models.TreasureInDataMapPending,
		AllDataReady:   models.AllDataReady,
	}

	vErr, err := uploadSession1.StartUploadSession()
	suite.False(vErr.HasAny())
	suite.Nil(err)
	vErr, err = uploadSession2.StartUploadSession()
	suite.False(vErr.HasAny())
	suite.Nil(err)
	vErr, err = uploadSession3.StartUploadSession()
	suite.False(vErr.HasAny())
	suite.Nil(err)
	vErr, err = uploadSession4.StartUploadSession()
	suite.False(vErr.HasAny())
	suite.Nil(err)
	vErr, err = uploadSession5.StartUploadSession()
	suite.False(vErr.HasAny())
	suite.Nil(err)

	// set uploadSession3 to be the oldest
	err = suite.DB.RawQuery("UPDATE upload_sessions SET created_at = ? WHERE genesis_hash = ?",
		time.Now().Add(-10*time.Second), uploadSession3.GenesisHash).All(&[]models.UploadSession{})

	// set uploadSession2 to be the newest
	err = suite.DB.RawQuery("UPDATE upload_sessions SET created_at = ? WHERE genesis_hash = ?",
		time.Now().Add(10*time.Second), uploadSession2.GenesisHash).All(&[]models.UploadSession{})

	sessions, err := models.GetSessionsByAge()
	suite.Nil(err)

	suite.Equal(3, len(sessions))

	//verify that the oldest session (uploadSession3) is first in the array
	suite.Equal(uploadSession3.GenesisHash, sessions[0].GenesisHash)
	suite.Equal(uploadSession1.GenesisHash, sessions[1].GenesisHash)
	suite.Equal(uploadSession2.GenesisHash, sessions[2].GenesisHash)
}

func (suite *ModelSuite) Test_GetSessionsThatNeedKeysEncrypted() {
	defer oyster_utils.ResetBrokerMode()
	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)

	numChunks := 250
	storageLengthInYears := 3

	u1 := models.UploadSession{
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        123,
		NumChunks:            numChunks,
		StorageLengthInYears: storageLengthInYears,
		PaymentStatus:        models.PaymentStatusConfirmed,
		TreasureStatus:       models.TreasureGeneratingKeys,
	}
	vErr, err := u1.StartUploadSession()
	suite.False(vErr.HasAny())
	suite.Nil(err)

	u2 := models.UploadSession{
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        123,
		NumChunks:            numChunks,
		StorageLengthInYears: storageLengthInYears,
		PaymentStatus:        models.PaymentStatusInvoiced,
		TreasureStatus:       models.TreasureGeneratingKeys,
	}
	vErr, err = u2.StartUploadSession()
	suite.False(vErr.HasAny())
	suite.Nil(err)

	u3 := models.UploadSession{
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        123,
		NumChunks:            numChunks,
		StorageLengthInYears: storageLengthInYears,
		PaymentStatus:        models.PaymentStatusConfirmed,
		TreasureStatus:       models.TreasureInDataMapPending,
	}
	vErr, err = u3.StartUploadSession()
	suite.False(vErr.HasAny())
	suite.Nil(err)

	sessions, err := models.GetSessionsThatNeedKeysEncrypted()
	suite.Nil(err)
	suite.Equal(1, len(sessions))
	suite.Equal(u1.GenesisHash, sessions[0].GenesisHash)
}

func (suite *ModelSuite) Test_GetSessionsThatNeedTreasure() {
	defer oyster_utils.ResetBrokerMode()
	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)

	numChunks := 250
	storageLengthInYears := 3

	u1 := models.UploadSession{
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        123,
		NumChunks:            numChunks,
		StorageLengthInYears: storageLengthInYears,
		PaymentStatus:        models.PaymentStatusConfirmed,
		TreasureStatus:       models.TreasureGeneratingKeys,
	}
	vErr, err := u1.StartUploadSession()
	suite.False(vErr.HasAny())
	suite.Nil(err)

	u2 := models.UploadSession{
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        123,
		NumChunks:            numChunks,
		StorageLengthInYears: storageLengthInYears,
		PaymentStatus:        models.PaymentStatusInvoiced,
		TreasureStatus:       models.TreasureInDataMapPending,
	}
	vErr, err = u2.StartUploadSession()
	suite.False(vErr.HasAny())
	suite.Nil(err)

	u3 := models.UploadSession{
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        123,
		NumChunks:            numChunks,
		StorageLengthInYears: storageLengthInYears,
		PaymentStatus:        models.PaymentStatusConfirmed,
		TreasureStatus:       models.TreasureInDataMapPending,
	}
	vErr, err = u3.StartUploadSession()
	suite.False(vErr.HasAny())
	suite.Nil(err)

	sessions, err := models.GetSessionsThatNeedTreasure()
	suite.Nil(err)
	suite.Equal(1, len(sessions))
	suite.Equal(u3.GenesisHash, sessions[0].GenesisHash)
}

func (suite *ModelSuite) Test_MakeTreasureIdxMap() {

	defer oyster_utils.ResetBrokerMode()
	oyster_utils.SetBrokerMode(oyster_utils.TestModeDummyTreasure)

	sectorSize := 100

	numChunks := 250
	storageLengthInYears := 3
	alphaIndexes := []int{2, 121, 245}
	betaIndexes := []int{9, 89, 230}

	u := models.UploadSession{
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        123,
		NumChunks:            numChunks,
		StorageLengthInYears: storageLengthInYears,
	}

	vErr, err := u.StartUploadSession()
	suite.False(vErr.HasAny())
	suite.Nil(err)
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

	decryptedKey0, _ := u.DecryptTreasureChunkEthKey(treasureIdxMap[0].Key)
	decryptedKey1, _ := u.DecryptTreasureChunkEthKey(treasureIdxMap[1].Key)
	decryptedKey2, _ := u.DecryptTreasureChunkEthKey(treasureIdxMap[2].Key)

	suite.Equal("9999999999999999999999999999999999999999999999999999999999999999",
		decryptedKey0)
	suite.Equal("9999999999999999999999999999999999999999999999999999999999999999",
		decryptedKey1)
	suite.Equal("9999999999999999999999999999999999999999999999999999999999999999",
		decryptedKey2)
}

func (suite *ModelSuite) Test_GetTreasureIndexes() {

	defer oyster_utils.ResetBrokerMode()
	oyster_utils.SetBrokerMode(oyster_utils.TestModeDummyTreasure)

	numChunks := 250
	storageLengthInYears := 3
	expectedIndexes := []int{5, 121, 225}

	testMap := `[{
		"sector": 0,
		"idx": 5,
		"key": "firstKey"
		},
		{
		"sector": 1,
		"idx": 121,
		"key": "secondKey"
		},
		{
		"sector": 2,
		"idx": 225,
		"key": "thirdKey"
		}]`

	u := models.UploadSession{
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        123,
		NumChunks:            numChunks,
		StorageLengthInYears: storageLengthInYears,
		TreasureIdxMap:       nulls.String{string(testMap), true},
	}

	chunkReqs := GenerateChunkRequests(numChunks, u.GenesisHash)
	models.ProcessAndStoreChunkData(chunkReqs, u.GenesisHash, expectedIndexes, oyster_utils.TestValueTimeToLive)

	vErr, err := u.StartUploadSession()
	suite.Nil(err)
	suite.False(vErr.HasAny())

	actualIndexes, err := u.GetTreasureIndexes()

	suite.Equal(expectedIndexes, actualIndexes)
}

func (suite *ModelSuite) Test_EncryptAndDecryptEthKey() {
	ethKey := hex.EncodeToString([]byte("SOME_PRIVATE_KEY"))

	u := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
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

func (suite *ModelSuite) Test_WaitForAllChunks() {

	storageLengthInYears := 3

	u := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		NumChunks:            200,
		FileSizeBytes:        9000000,
		StorageLengthInYears: storageLengthInYears,
		PaymentStatus:        models.PaymentStatusConfirmed,
		TreasureStatus:       models.TreasureInDataMapPending,
	}

	mergedIndexes := []int{45}
	chunkReqs := GenerateChunkRequests(200, u.GenesisHash)
	models.ProcessAndStoreChunkData(chunkReqs, u.GenesisHash, mergedIndexes, oyster_utils.TestValueTimeToLive)

	vErr, err := u.StartUploadSession()
	suite.Nil(err)
	suite.False(vErr.HasAny())

	privateKeys := []string{"0000000001"}

	u.MakeTreasureIdxMap(mergedIndexes, privateKeys)
	for {
		err = u.SetTreasureMessage(mergedIndexes[0], "SOMEVALUE", oyster_utils.TestValueTimeToLive)
		if err == nil {
			break
		}
	}

	allChunksExist, err := u.WaitForAllChunks(500)
	suite.True(allChunksExist)
	suite.Nil(err)
}

func (suite *ModelSuite) Test_CalculatePayment_Less_Than_1_GB() {

	currentStoragePeg := models.StoragePeg

	defer func() { models.StoragePeg = currentStoragePeg }()

	models.StoragePeg = decimal.NewFromFloat(float64(64))
	storageLengthInYears := 3

	u := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
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

	defer oyster_utils.ResetBrokerMode()
	oyster_utils.SetBrokerMode(oyster_utils.TestModeDummyTreasure)

	currentStoragePeg := models.StoragePeg

	defer func() { models.StoragePeg = currentStoragePeg }()

	models.StoragePeg = decimal.NewFromFloat(float64(64))
	storageLengthInYears := 3

	u := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
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

	defer oyster_utils.ResetBrokerMode()
	oyster_utils.SetBrokerMode(oyster_utils.TestModeDummyTreasure)

	currentStoragePeg := models.StoragePeg

	defer func() { models.StoragePeg = currentStoragePeg }()

	fileSizeBytes := uint64(1999999 * oyster_utils.FileChunkSizeInByte)

	models.StoragePeg = decimal.NewFromFloat(float64(64))
	storageLengthInYears := 3

	u := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
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

	defer oyster_utils.ResetBrokerMode()
	oyster_utils.SetBrokerMode(oyster_utils.TestModeDummyTreasure)

	currentStoragePeg := models.StoragePeg

	defer func() { models.StoragePeg = currentStoragePeg }()

	fileSizeBytes := uint64(2000000 * oyster_utils.FileChunkSizeInByte)

	models.StoragePeg = decimal.NewFromFloat(float64(64))
	storageLengthInYears := 3

	u := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
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
	defer oyster_utils.ResetBrokerMode()
	oyster_utils.SetBrokerMode(oyster_utils.TestModeDummyTreasure)

	totalCost := 5
	numSectors := 3

	numChunks := 250
	storageLengthInYears := 3
	mergedIndexes := []int{2, 121, 245}
	privateKeys := []string{
		"9999999999999999999999999999999999999999999999999999999999999999",
		"9999999999999999999999999999999999999999999999999999999999999999",
		"9999999999999999999999999999999999999999999999999999999999999999",
	}

	u := models.UploadSession{
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
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

func (ms *ModelSuite) Test_SetBrokerTransactionToPaid() {

	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	defer oyster_utils.ResetBrokerMode()

	fileSizeBytes := 123
	numChunks := 2
	storageLengthInYears := 2
	privateKey := "abcdef1234567890"
	startingEthAddr := "0000000000"

	u := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(fileSizeBytes),
		NumChunks:            numChunks,
		StorageLengthInYears: storageLengthInYears,
		ETHPrivateKey:        privateKey,
		ETHAddrAlpha:         nulls.String{string(startingEthAddr), true},
		ETHAddrBeta:          nulls.String{string(startingEthAddr), true},
		TotalCost:            totalCost,
	}

	vErr, err := u.StartUploadSession()
	ms.Nil(err)
	ms.Equal(0, len(vErr.Errors))

	uSession := models.UploadSession{}
	ms.DB.Where("genesis_hash = ?", u.GenesisHash).All(&uSession)

	models.NewBrokerBrokerTransaction(&uSession)

	brokerTxs := returnAllBrokerBrokerTxs(ms)
	ms.Equal(1, len(brokerTxs))

	ms.Equal(models.BrokerTxAlphaPaymentPending, brokerTxs[0].PaymentStatus)

	models.SetBrokerTransactionToPaid(uSession)

	brokerTxs = returnAllBrokerBrokerTxs(ms)
	ms.Equal(1, len(brokerTxs))

	ms.Equal(models.BrokerTxAlphaPaymentConfirmed, brokerTxs[0].PaymentStatus)
}

func (suite *ModelSuite) Test_ProcessAndStoreChunkData_badger() {
	oyster_utils.SetStorageMode(oyster_utils.DataMapsInBadger)
	defer oyster_utils.ResetDataMapStorageMode()

	oyster_utils.RemoveAllKvStoreDataFromAllKvStores()

	genHash := oyster_utils.RandSeq(6, []rune("abcdef0123456789"))

	mergedIndexes := []int{5}
	chunkReqs := GenerateChunkRequests(10, genHash)
	models.ProcessAndStoreChunkData(chunkReqs, genHash, mergedIndexes, oyster_utils.TestValueTimeToLive)

	allMessagesPresent := true
	treasureIdx := 5
	treasurePresent := false

	for i := 0; i < 11; i++ {
		chunkData := models.GetSingleChunkData(oyster_utils.InProgressDir, genHash, int64(i))
		if chunkData.RawMessage == "" && i != treasureIdx {
			allMessagesPresent = false
		}
		if i == treasureIdx && chunkData.RawMessage != "" {
			treasurePresent = true
		}
	}

	suite.True(allMessagesPresent)
	suite.False(treasurePresent)
}

func (suite *ModelSuite) Test_ProcessAndStoreChunkData_sql() {
	oyster_utils.SetStorageMode(oyster_utils.DataMapsInSQL)
	defer oyster_utils.ResetDataMapStorageMode()

	genHash := oyster_utils.RandSeq(6, []rune("abcdef0123456789"))

	mergedIndexes := []int{5}
	chunkReqs := GenerateChunkRequests(10, genHash)
	models.ProcessAndStoreChunkData(chunkReqs, genHash, mergedIndexes, oyster_utils.TestValueTimeToLive)

	allMessagesPresent := true
	treasureIdx := 5
	treasurePresent := false

	time.Sleep(5 * time.Second)

	for i := 0; i < 11; i++ {
		chunkData := models.GetSingleChunkData(oyster_utils.InProgressDir, genHash, int64(i))
		if chunkData.RawMessage == "" && i != treasureIdx {
			allMessagesPresent = false
		}
		if i == treasureIdx && chunkData.RawMessage != "" {
			treasurePresent = true
		}
	}

	suite.True(allMessagesPresent)
	suite.False(treasurePresent)
}

func (suite *ModelSuite) Test_BuildDataMapsForSession_badger() {
	oyster_utils.SetStorageMode(oyster_utils.DataMapsInBadger)
	defer oyster_utils.ResetDataMapStorageMode()

	genHash := "abcdef44"
	numChunks := 7

	u := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          genHash,
		FileSizeBytes:        uint64(7000),
		NumChunks:            numChunks,
		StorageLengthInYears: 2,
	}

	suite.DB.ValidateAndCreate(&u)
	err := models.BuildDataMapsForSession(genHash, numChunks)
	suite.Nil(err)

	expectedHashesWithKeys := make(map[string]string)

	expectedHashesWithKeys[oyster_utils.GetBadgerKey([]string{u.GenesisHash, "0"})] = genHash
	expectedHashesWithKeys[oyster_utils.GetBadgerKey([]string{u.GenesisHash, "1"})] =
		"c55fa76fed24435e2722bfbfa905e173d894acf8ad7e5542093e2bc525705f70"
	expectedHashesWithKeys[oyster_utils.GetBadgerKey([]string{u.GenesisHash, "2"})] =
		"2b73088a9316322905f5b1c734baa844d8b12fd01b8c7431362d85ac51c612b1"
	expectedHashesWithKeys[oyster_utils.GetBadgerKey([]string{u.GenesisHash, "3"})] =
		"9e670d841a421eddd9eae598298da1efe2cef072981381978e5baae295bb0819"
	expectedHashesWithKeys[oyster_utils.GetBadgerKey([]string{u.GenesisHash, "4"})] =
		"c7f1dfabcb8467b0199e942b2f0d93d8073e759912a826c6a670f31a850e24e7"
	expectedHashesWithKeys[oyster_utils.GetBadgerKey([]string{u.GenesisHash, "5"})] =
		"37bdfe84e67e0d7cba26085cd9346613814f5a91928a111468fe127c3635a912"
	expectedHashesWithKeys[oyster_utils.GetBadgerKey([]string{u.GenesisHash, "6"})] =
		"89683d310638d51174e71da6786ef8185f9e7fea3314f20414b93a20f7feb57e"

	finishedHashes, _ := u.WaitForAllHashes(500)
	suite.True(finishedHashes)

	for i := 0; i < numChunks; i++ {
		singleChunkData := models.GetSingleChunkData(oyster_utils.InProgressDir, u.GenesisHash, int64(i))

		key := oyster_utils.GetBadgerKey([]string{u.GenesisHash, strconv.Itoa(i)})

		suite.Equal(expectedHashesWithKeys[key], singleChunkData.Hash)
	}
}

func (suite *ModelSuite) Test_BuildDataMapsForSession_sql() {
	oyster_utils.SetStorageMode(oyster_utils.DataMapsInSQL)
	defer oyster_utils.ResetDataMapStorageMode()

	genHash := "abcdef44"
	numChunks := 7

	u := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          genHash,
		FileSizeBytes:        uint64(7000),
		NumChunks:            numChunks,
		StorageLengthInYears: 2,
	}

	suite.DB.ValidateAndCreate(&u)
	err := models.BuildDataMapsForSession(genHash, numChunks)
	suite.Nil(err)

	expectedHashesWithKeys := make(map[string]string)

	expectedHashesWithKeys[oyster_utils.GetBadgerKey([]string{u.GenesisHash, "0"})] = genHash
	expectedHashesWithKeys[oyster_utils.GetBadgerKey([]string{u.GenesisHash, "1"})] =
		"c55fa76fed24435e2722bfbfa905e173d894acf8ad7e5542093e2bc525705f70"
	expectedHashesWithKeys[oyster_utils.GetBadgerKey([]string{u.GenesisHash, "2"})] =
		"2b73088a9316322905f5b1c734baa844d8b12fd01b8c7431362d85ac51c612b1"
	expectedHashesWithKeys[oyster_utils.GetBadgerKey([]string{u.GenesisHash, "3"})] =
		"9e670d841a421eddd9eae598298da1efe2cef072981381978e5baae295bb0819"
	expectedHashesWithKeys[oyster_utils.GetBadgerKey([]string{u.GenesisHash, "4"})] =
		"c7f1dfabcb8467b0199e942b2f0d93d8073e759912a826c6a670f31a850e24e7"
	expectedHashesWithKeys[oyster_utils.GetBadgerKey([]string{u.GenesisHash, "5"})] =
		"37bdfe84e67e0d7cba26085cd9346613814f5a91928a111468fe127c3635a912"
	expectedHashesWithKeys[oyster_utils.GetBadgerKey([]string{u.GenesisHash, "6"})] =
		"89683d310638d51174e71da6786ef8185f9e7fea3314f20414b93a20f7feb57e"

	finishedHashes, _ := u.WaitForAllHashes(500)
	suite.True(finishedHashes)

	for i := 0; i < numChunks; i++ {
		singleChunkData := models.GetSingleChunkData(oyster_utils.InProgressDir, u.GenesisHash, int64(i))

		key := oyster_utils.GetBadgerKey([]string{u.GenesisHash, strconv.Itoa(i)})

		suite.Equal(expectedHashesWithKeys[key], singleChunkData.Hash)
	}
}

func (suite *ModelSuite) Test_EncryptTreasureChunkEthKey_DecryptTreasureChunkEthKey() {

	ethKey := hex.EncodeToString([]byte("SOME_PRIVATE_KEY"))

	u := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        123,
		NumChunks:            400,
		StorageLengthInYears: 4,
		ETHPrivateKey:        hex.EncodeToString([]byte("SOME_OTHER_KEY")),
	}

	_, err := u.StartUploadSession()
	suite.Nil(err)

	encryptedKey, err := u.EncryptTreasureChunkEthKey(ethKey)
	decryptedKey, err := u.DecryptTreasureChunkEthKey(encryptedKey)

	suite.Equal(ethKey, decryptedKey)

}

func (suite *ModelSuite) Test_WaitForAllHashes() {

	numChunks := 7

	u := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(7000),
		NumChunks:            numChunks,
		StorageLengthInYears: 2,
	}

	suite.DB.ValidateAndCreate(&u)
	err := models.BuildDataMapsForSession(u.GenesisHash, numChunks)
	suite.Nil(err)

	finishedHashes, _ := u.WaitForAllHashes(500)
	suite.True(finishedHashes)

	u2 := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(7000),
		NumChunks:            numChunks,
		StorageLengthInYears: 2,
	}

	suite.DB.ValidateAndCreate(&u2)

	finishedHashes, _ = u2.WaitForAllHashes(1)
	// This will be false since we never built the data maps
	suite.False(finishedHashes)
}

func (suite *ModelSuite) Test_WaitForAllMessages() {
	oyster_utils.SetBrokerMode(oyster_utils.TestModeNoTreasure)
	defer oyster_utils.ResetBrokerMode()

	numChunks := 7

	u := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(7000),
		NumChunks:            numChunks,
		StorageLengthInYears: 2,
	}

	suite.DB.ValidateAndCreate(&u)
	u.MakeTreasureIdxMap([]int{}, []string{})

	finishedMessages, _ := u.WaitForAllMessages(1)
	// This will be false since we did not send any chunk requests
	suite.False(finishedMessages)

	chunkReqs := GenerateChunkRequests(numChunks, u.GenesisHash)

	firstBatch := chunkReqs[0:2]
	secondBatch := chunkReqs[2:numChunks]

	models.ProcessAndStoreChunkData(firstBatch, u.GenesisHash, []int{}, oyster_utils.TestValueTimeToLive)

	finishedMessages, _ = u.WaitForAllMessages(2)
	// This will be false since we have not yet sent all the chunks
	suite.False(finishedMessages)

	models.ProcessAndStoreChunkData(secondBatch, u.GenesisHash, []int{}, oyster_utils.TestValueTimeToLive)

	finishedMessages, _ = u.WaitForAllMessages(10)
	// This will be true since we have sent all the chunks
	suite.True(finishedMessages)
}

func (suite *ModelSuite) Test_CheckIfAllDataIsReady() {
	/*
		Currently no need to test this method since it just returns the result
		of CheckIfAllMessagesAreReady() && CheckIfAllHashesAreReady()
	*/
}
func (suite *ModelSuite) Test_CheckIfAllHashesAreReady_badger() {
	oyster_utils.SetStorageMode(oyster_utils.DataMapsInBadger)
	defer oyster_utils.ResetDataMapStorageMode()

	numChunks := 7

	u := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(7000),
		NumChunks:            numChunks,
		StorageLengthInYears: 2,
	}

	suite.DB.ValidateAndCreate(&u)
	err := models.BuildDataMapsForSession(u.GenesisHash, numChunks)
	suite.Nil(err)

	finishedHashes := u.CheckIfAllHashesAreReady()
	suite.True(finishedHashes)

	u2 := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(7000),
		NumChunks:            numChunks,
		StorageLengthInYears: 2,
	}

	suite.DB.ValidateAndCreate(&u2)

	// calling this to make sure badger has had time to finish
	u.WaitForAllHashes(10)
	finishedHashes = u2.CheckIfAllHashesAreReady()
	// This will be false since we never built the data maps
	suite.False(finishedHashes)
}

func (suite *ModelSuite) Test_CheckIfAllHashesAreReady_sql() {
	oyster_utils.SetStorageMode(oyster_utils.DataMapsInSQL)
	defer oyster_utils.ResetDataMapStorageMode()

	numChunks := 7

	u := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(7000),
		NumChunks:            numChunks,
		StorageLengthInYears: 2,
	}

	suite.DB.ValidateAndCreate(&u)
	err := models.BuildDataMapsForSession(u.GenesisHash, numChunks)
	suite.Nil(err)

	finishedHashes := u.CheckIfAllHashesAreReady()
	suite.True(finishedHashes)

	u2 := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(7000),
		NumChunks:            numChunks,
		StorageLengthInYears: 2,
	}

	suite.DB.ValidateAndCreate(&u2)

	// calling this to make sure badger has had time to finish
	u.WaitForAllHashes(10)
	finishedHashes = u2.CheckIfAllHashesAreReady()
	// This will be false since we never built the data maps
	suite.False(finishedHashes)
}

func (suite *ModelSuite) Test_CheckIfAllMessagesAreReady_badger() {
	oyster_utils.SetBrokerMode(oyster_utils.TestModeNoTreasure)
	defer oyster_utils.ResetBrokerMode()

	oyster_utils.SetStorageMode(oyster_utils.DataMapsInBadger)
	defer oyster_utils.ResetDataMapStorageMode()

	numChunks := 7

	u := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(7000),
		NumChunks:            numChunks,
		StorageLengthInYears: 2,
	}

	suite.DB.ValidateAndCreate(&u)
	u.MakeTreasureIdxMap([]int{}, []string{})

	finishedMessages := u.CheckIfAllMessagesAreReady()
	// This will be false since we did not send any chunk requests
	suite.False(finishedMessages)

	chunkReqs := GenerateChunkRequests(numChunks, u.GenesisHash)

	firstBatch := chunkReqs[0:2]
	secondBatch := chunkReqs[2:numChunks]

	models.ProcessAndStoreChunkData(firstBatch, u.GenesisHash, []int{}, oyster_utils.TestValueTimeToLive)

	finishedMessages = u.CheckIfAllMessagesAreReady()
	// This will be false since we have not yet sent all the chunks
	suite.False(finishedMessages)

	models.ProcessAndStoreChunkData(secondBatch, u.GenesisHash, []int{}, oyster_utils.TestValueTimeToLive)

	// calling this to make sure badger has had time to finish
	u.WaitForAllMessages(10)
	finishedMessages = u.CheckIfAllMessagesAreReady()
	// This will be true since we have sent all the chunks
	suite.True(finishedMessages)
}

func (suite *ModelSuite) Test_CheckIfAllMessagesAreReady_sql() {
	oyster_utils.SetBrokerMode(oyster_utils.TestModeNoTreasure)
	defer oyster_utils.ResetBrokerMode()

	oyster_utils.SetStorageMode(oyster_utils.DataMapsInSQL)
	defer oyster_utils.ResetDataMapStorageMode()

	numChunks := 7

	u := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(7000),
		NumChunks:            numChunks,
		StorageLengthInYears: 2,
	}

	suite.DB.ValidateAndCreate(&u)
	u.MakeTreasureIdxMap([]int{}, []string{})

	finishedMessages := u.CheckIfAllMessagesAreReady()
	// This will be false since we did not send any chunk requests
	suite.False(finishedMessages)

	chunkReqs := GenerateChunkRequests(numChunks, u.GenesisHash)

	firstBatch := chunkReqs[0:2]
	secondBatch := chunkReqs[2:numChunks]

	models.ProcessAndStoreChunkData(firstBatch, u.GenesisHash, []int{}, oyster_utils.TestValueTimeToLive)

	finishedMessages = u.CheckIfAllMessagesAreReady()
	// This will be false since we have not yet sent all the chunks
	suite.False(finishedMessages)

	models.ProcessAndStoreChunkData(secondBatch, u.GenesisHash, []int{}, oyster_utils.TestValueTimeToLive)

	// calling this to make sure badger has had time to finish
	u.WaitForAllMessages(10)
	finishedMessages = u.CheckIfAllMessagesAreReady()
	// This will be true since we have sent all the chunks
	suite.True(finishedMessages)
}

func (suite *ModelSuite) Test_GetUnassignedChunksBySession_alpha() {

	oyster_utils.SetBrokerMode(oyster_utils.TestModeNoTreasure)
	defer oyster_utils.ResetBrokerMode()

	numChunks := 9

	u := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(9000),
		NumChunks:            numChunks,
		StorageLengthInYears: 2,
	}

	suite.DB.ValidateAndCreate(&u)
	err := models.BuildDataMapsForSession(u.GenesisHash, numChunks)
	suite.Nil(err)
	u.MakeTreasureIdxMap([]int{}, []string{})

	chunkReqs := GenerateChunkRequests(numChunks, u.GenesisHash)
	models.ProcessAndStoreChunkData(chunkReqs, u.GenesisHash, []int{}, oyster_utils.TestValueTimeToLive)

	u.WaitForAllHashes(100)
	u.WaitForAllMessages(100)

	u.NextIdxToAttach = 2
	suite.DB.ValidateAndUpdate(&u)

	chunkData, err := u.GetUnassignedChunksBySession(3)

	suite.Equal(3, len(chunkData))

	for _, chunk := range chunkData {
		suite.True(chunk.Idx >= 2 && chunk.Idx <= 5)
	}

	chunkData, err = u.GetUnassignedChunksBySession(100)

	suite.Equal(7, len(chunkData))

	for _, chunk := range chunkData {
		suite.True(chunk.Idx >= 2 && chunk.Idx <= int64(u.NumChunks-1))
	}
}

func (suite *ModelSuite) Test_GetUnassignedChunksBySession_beta() {
	oyster_utils.SetBrokerMode(oyster_utils.TestModeNoTreasure)
	defer oyster_utils.ResetBrokerMode()

	numChunks := 9

	u := models.UploadSession{
		Type:                 models.SessionTypeBeta,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(9000),
		NumChunks:            numChunks,
		StorageLengthInYears: 2,
	}

	suite.DB.ValidateAndCreate(&u)
	err := models.BuildDataMapsForSession(u.GenesisHash, numChunks)
	suite.Nil(err)
	u.MakeTreasureIdxMap([]int{}, []string{})

	chunkReqs := GenerateChunkRequests(numChunks, u.GenesisHash)
	models.ProcessAndStoreChunkData(chunkReqs, u.GenesisHash, []int{}, oyster_utils.TestValueTimeToLive)

	u.WaitForAllHashes(100)
	u.WaitForAllMessages(100)

	u.NextIdxToAttach = 6
	suite.DB.ValidateAndUpdate(&u)

	chunkData, err := u.GetUnassignedChunksBySession(3)

	suite.Equal(3, len(chunkData))

	for _, chunk := range chunkData {
		suite.True(chunk.Idx >= 3 && chunk.Idx <= 6)
	}

	chunkData, err = u.GetUnassignedChunksBySession(100)

	suite.Equal(7, len(chunkData))

	for _, chunk := range chunkData {
		suite.True(chunk.Idx >= 0 && chunk.Idx <= 6)
	}
}

func (suite *ModelSuite) Test_MoveChunksToCompleted_badger() {
	oyster_utils.SetBrokerMode(oyster_utils.TestModeNoTreasure)
	defer oyster_utils.ResetBrokerMode()

	oyster_utils.SetStorageMode(oyster_utils.DataMapsInBadger)
	defer oyster_utils.ResetDataMapStorageMode()

	numChunks := 9

	u := models.UploadSession{
		Type:                 models.SessionTypeBeta,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(9000),
		NumChunks:            numChunks,
		StorageLengthInYears: 2,
	}

	suite.DB.ValidateAndCreate(&u)
	err := models.BuildDataMapsForSession(u.GenesisHash, numChunks)
	suite.Nil(err)
	u.MakeTreasureIdxMap([]int{}, []string{})

	chunkReqs := GenerateChunkRequests(numChunks, u.GenesisHash)
	models.ProcessAndStoreChunkData(chunkReqs, u.GenesisHash, []int{}, oyster_utils.TestValueTimeToLive)

	u.WaitForAllHashes(100)
	u.WaitForAllMessages(100)

	bulkKeys := oyster_utils.GenerateBulkKeys(u.GenesisHash, 0, 2)
	chunkData, err :=
		models.GetMultiChunkData(oyster_utils.InProgressDir, u.GenesisHash, bulkKeys)
	suite.Nil(err)
	suite.Equal(3, len(chunkData))

	u.MoveChunksToCompleted(chunkData)

	chunkData, err =
		models.GetMultiChunkData(oyster_utils.InProgressDir, u.GenesisHash, bulkKeys)
	suite.Nil(err)
	suite.Equal(0, len(chunkData))

	chunkData, err =
		models.GetMultiChunkData(oyster_utils.CompletedDir, u.GenesisHash, bulkKeys)
	suite.Nil(err)
	suite.Equal(3, len(chunkData))
}

func (suite *ModelSuite) Test_MoveChunksToCompleted_sql() {
	oyster_utils.SetBrokerMode(oyster_utils.TestModeNoTreasure)
	defer oyster_utils.ResetBrokerMode()

	oyster_utils.SetStorageMode(oyster_utils.DataMapsInSQL)
	defer oyster_utils.ResetDataMapStorageMode()

	numChunks := 9

	u := models.UploadSession{
		Type:                 models.SessionTypeBeta,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(9000),
		NumChunks:            numChunks,
		StorageLengthInYears: 2,
	}

	suite.DB.ValidateAndCreate(&u)
	err := models.BuildDataMapsForSession(u.GenesisHash, numChunks)
	suite.Nil(err)
	u.MakeTreasureIdxMap([]int{}, []string{})

	chunkReqs := GenerateChunkRequests(numChunks, u.GenesisHash)
	models.ProcessAndStoreChunkData(chunkReqs, u.GenesisHash, []int{}, oyster_utils.TestValueTimeToLive)

	u.WaitForAllHashes(100)
	u.WaitForAllMessages(100)

	bulkKeys := oyster_utils.GenerateBulkKeys(u.GenesisHash, 0, 2)
	chunkData, err :=
		models.GetMultiChunkData(oyster_utils.InProgressDir, u.GenesisHash, bulkKeys)
	suite.Nil(err)
	suite.Equal(3, len(chunkData))

	u.MoveChunksToCompleted(chunkData)

	chunkData, err =
		models.GetMultiChunkData(oyster_utils.InProgressDir, u.GenesisHash, bulkKeys)
	suite.Nil(err)
	suite.Equal(0, len(chunkData))

	chunkData, err =
		models.GetMultiChunkData(oyster_utils.CompletedDir, u.GenesisHash, bulkKeys)
	suite.Nil(err)
	suite.Equal(3, len(chunkData))
}

func (suite *ModelSuite) Test_MoveAllChunksToCompleted_badger() {
	oyster_utils.SetBrokerMode(oyster_utils.TestModeNoTreasure)
	defer oyster_utils.ResetBrokerMode()

	oyster_utils.SetStorageMode(oyster_utils.DataMapsInBadger)
	defer oyster_utils.ResetDataMapStorageMode()

	numChunks := 9

	u := models.UploadSession{
		Type:                 models.SessionTypeBeta,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(9000),
		NumChunks:            numChunks,
		StorageLengthInYears: 2,
	}

	suite.DB.ValidateAndCreate(&u)
	err := models.BuildDataMapsForSession(u.GenesisHash, numChunks)
	suite.Nil(err)
	u.MakeTreasureIdxMap([]int{}, []string{})

	chunkReqs := GenerateChunkRequests(numChunks, u.GenesisHash)
	models.ProcessAndStoreChunkData(chunkReqs, u.GenesisHash, []int{}, oyster_utils.TestValueTimeToLive)

	u.WaitForAllHashes(100)
	u.WaitForAllMessages(100)

	bulkKeys := oyster_utils.GenerateBulkKeys(u.GenesisHash, 0, int64(u.NumChunks-1))
	chunkData, err :=
		models.GetMultiChunkData(oyster_utils.InProgressDir, u.GenesisHash, bulkKeys)
	suite.Nil(err)
	suite.Equal(numChunks, len(chunkData))

	u.MoveChunksToCompleted(chunkData)

	chunkData, err =
		models.GetMultiChunkData(oyster_utils.InProgressDir, u.GenesisHash, bulkKeys)
	suite.Nil(err)
	suite.Equal(0, len(chunkData))

	chunkData, err =
		models.GetMultiChunkData(oyster_utils.CompletedDir, u.GenesisHash, bulkKeys)
	suite.Nil(err)
	suite.Equal(numChunks, len(chunkData))
}

func (suite *ModelSuite) Test_MoveAllChunksToCompleted_sql() {
	oyster_utils.SetBrokerMode(oyster_utils.TestModeNoTreasure)
	defer oyster_utils.ResetBrokerMode()

	oyster_utils.SetStorageMode(oyster_utils.DataMapsInSQL)
	defer oyster_utils.ResetDataMapStorageMode()

	numChunks := 9

	u := models.UploadSession{
		Type:                 models.SessionTypeBeta,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(9000),
		NumChunks:            numChunks,
		StorageLengthInYears: 2,
	}

	suite.DB.ValidateAndCreate(&u)
	err := models.BuildDataMapsForSession(u.GenesisHash, numChunks)
	suite.Nil(err)
	u.MakeTreasureIdxMap([]int{}, []string{})

	chunkReqs := GenerateChunkRequests(numChunks, u.GenesisHash)
	models.ProcessAndStoreChunkData(chunkReqs, u.GenesisHash, []int{}, oyster_utils.TestValueTimeToLive)

	u.WaitForAllHashes(100)
	u.WaitForAllMessages(100)

	bulkKeys := oyster_utils.GenerateBulkKeys(u.GenesisHash, 0, int64(u.NumChunks-1))
	chunkData, err :=
		models.GetMultiChunkData(oyster_utils.InProgressDir, u.GenesisHash, bulkKeys)
	suite.Nil(err)
	suite.Equal(numChunks, len(chunkData))

	u.MoveChunksToCompleted(chunkData)

	chunkData, err =
		models.GetMultiChunkData(oyster_utils.InProgressDir, u.GenesisHash, bulkKeys)
	suite.Nil(err)
	suite.Equal(0, len(chunkData))

	chunkData, err =
		models.GetMultiChunkData(oyster_utils.CompletedDir, u.GenesisHash, bulkKeys)
	suite.Nil(err)
	suite.Equal(numChunks, len(chunkData))
}

func (suite *ModelSuite) Test_UpdateIndexWithVerifiedChunks_alpha_treasure_not_complete() {
	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	defer oyster_utils.ResetBrokerMode()

	numChunks := 9

	u := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(9000),
		NumChunks:            numChunks,
		StorageLengthInYears: 2,
	}

	mergedIndexes := []int{5}

	bulkChunkData := SessionSetUpForTest(&u, mergedIndexes, u.NumChunks)

	u.NextIdxToAttach = int64(u.NumChunks) - 1
	u.NextIdxToVerify = 2
	suite.DB.ValidateAndUpdate(&u)

	bulkKeys := oyster_utils.GenerateBulkKeys(u.GenesisHash, 2, int64(u.NumChunks)-1)
	bulkChunkData, err := models.GetMultiChunkData(oyster_utils.InProgressDir, u.GenesisHash,
		bulkKeys)
	suite.Nil(err)

	u.UpdateIndexWithVerifiedChunks(bulkChunkData)

	suite.Equal(int64(mergedIndexes[0]), u.NextIdxToVerify)
}

func (suite *ModelSuite) Test_UpdateIndexWithVerifiedChunks_alpha_treasure_complete() {
	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	defer oyster_utils.ResetBrokerMode()

	numChunks := 9

	u := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(9000),
		NumChunks:            numChunks,
		StorageLengthInYears: 2,
	}

	mergedIndexes := []int{5}

	bulkChunkData := SessionSetUpForTest(&u, mergedIndexes, u.NumChunks)

	u.NextIdxToAttach = int64(u.NumChunks) - 1
	u.NextIdxToVerify = 2
	suite.DB.ValidateAndUpdate(&u)

	bulkKeys := oyster_utils.GenerateBulkKeys(u.GenesisHash, 2, int64(u.NumChunks)-1)
	bulkChunkData, err := models.GetMultiChunkData(oyster_utils.InProgressDir, u.GenesisHash,
		bulkKeys)
	suite.Nil(err)

	u.MoveChunksToCompleted([]oyster_utils.ChunkData{bulkChunkData[3]}) // the index of the treasure chunks

	u.UpdateIndexWithVerifiedChunks(bulkChunkData)

	suite.Equal(int64(u.NumChunks-1), u.NextIdxToVerify)
}

func (suite *ModelSuite) Test_UpdateIndexWithVerifiedChunks_beta_treasure_not_complete() {
	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	defer oyster_utils.ResetBrokerMode()

	numChunks := 9

	u := models.UploadSession{
		Type:                 models.SessionTypeBeta,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(9000),
		NumChunks:            numChunks,
		StorageLengthInYears: 2,
	}

	mergedIndexes := []int{5}

	bulkChunkData := SessionSetUpForTest(&u, mergedIndexes, u.NumChunks)

	u.NextIdxToAttach = -1
	u.NextIdxToVerify = int64(u.NumChunks - 3)
	suite.DB.ValidateAndUpdate(&u)

	bulkKeys := oyster_utils.GenerateBulkKeys(u.GenesisHash, int64(u.NumChunks-3), 0)
	bulkChunkData, err := models.GetMultiChunkData(oyster_utils.InProgressDir, u.GenesisHash,
		bulkKeys)
	suite.Nil(err)

	u.UpdateIndexWithVerifiedChunks(bulkChunkData)

	suite.Equal(int64(mergedIndexes[0]), u.NextIdxToVerify)
}

func (suite *ModelSuite) Test_UpdateIndexWithVerifiedChunks_beta_treasure_complete() {
	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	defer oyster_utils.ResetBrokerMode()

	numChunks := 9

	u := models.UploadSession{
		Type:                 models.SessionTypeBeta,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(9000),
		NumChunks:            numChunks,
		StorageLengthInYears: 2,
	}

	mergedIndexes := []int{5}

	bulkChunkData := SessionSetUpForTest(&u, mergedIndexes, u.NumChunks)

	u.NextIdxToAttach = -1
	u.NextIdxToVerify = int64(u.NumChunks - 3)
	suite.DB.ValidateAndUpdate(&u)

	bulkKeys := oyster_utils.GenerateBulkKeys(u.GenesisHash, int64(u.NumChunks-3), 0)
	bulkChunkData, err := models.GetMultiChunkData(oyster_utils.InProgressDir, u.GenesisHash,
		bulkKeys)
	suite.Nil(err)

	u.MoveChunksToCompleted([]oyster_utils.ChunkData{bulkChunkData[2]}) // the index of the treasure chunks

	u.UpdateIndexWithVerifiedChunks(bulkChunkData)

	suite.Equal(int64(-1), u.NextIdxToVerify)
}

func (suite *ModelSuite) Test_UpdateIndexWithAttachedChunks_alpha_treasure_not_complete() {
	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	defer oyster_utils.ResetBrokerMode()

	numChunks := 9

	u := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(9000),
		NumChunks:            numChunks,
		StorageLengthInYears: 2,
	}

	mergedIndexes := []int{5}

	bulkChunkData := SessionSetUpForTest(&u, mergedIndexes, u.NumChunks)

	u.NextIdxToAttach = 2
	suite.DB.ValidateAndUpdate(&u)

	bulkKeys := oyster_utils.GenerateBulkKeys(u.GenesisHash, 2, int64(u.NumChunks)-1)
	bulkChunkData, err := models.GetMultiChunkData(oyster_utils.InProgressDir, u.GenesisHash,
		bulkKeys)
	suite.Nil(err)

	u.UpdateIndexWithAttachedChunks(bulkChunkData)

	suite.Equal(int64(mergedIndexes[0]), u.NextIdxToAttach)
}

func (suite *ModelSuite) Test_UpdateIndexWithAttachedChunks_alpha_treasure_complete() {
	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	defer oyster_utils.ResetBrokerMode()

	numChunks := 9

	u := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(9000),
		NumChunks:            numChunks,
		StorageLengthInYears: 2,
	}

	mergedIndexes := []int{5}

	bulkChunkData := SessionSetUpForTest(&u, mergedIndexes, u.NumChunks)

	u.NextIdxToAttach = 2
	suite.DB.ValidateAndUpdate(&u)

	bulkKeys := oyster_utils.GenerateBulkKeys(u.GenesisHash, 2, int64(u.NumChunks)-1)
	bulkChunkData, err := models.GetMultiChunkData(oyster_utils.InProgressDir, u.GenesisHash,
		bulkKeys)
	suite.Nil(err)

	u.MoveChunksToCompleted([]oyster_utils.ChunkData{bulkChunkData[3]}) // the index of the treasure chunks

	u.UpdateIndexWithAttachedChunks(bulkChunkData)

	suite.Equal(int64(u.NumChunks), u.NextIdxToAttach)
}

func (suite *ModelSuite) Test_UpdateIndexWithAttachedChunks_beta_treasure_not_complete() {
	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	defer oyster_utils.ResetBrokerMode()

	numChunks := 9

	u := models.UploadSession{
		Type:                 models.SessionTypeBeta,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(9000),
		NumChunks:            numChunks,
		StorageLengthInYears: 2,
	}

	mergedIndexes := []int{5}

	bulkChunkData := SessionSetUpForTest(&u, mergedIndexes, u.NumChunks)

	u.NextIdxToAttach = int64(u.NumChunks - 3)
	suite.DB.ValidateAndUpdate(&u)

	bulkKeys := oyster_utils.GenerateBulkKeys(u.GenesisHash, int64(u.NumChunks-3), 0)
	bulkChunkData, err := models.GetMultiChunkData(oyster_utils.InProgressDir, u.GenesisHash,
		bulkKeys)
	suite.Nil(err)

	u.UpdateIndexWithAttachedChunks(bulkChunkData)

	suite.Equal(int64(mergedIndexes[0]), u.NextIdxToAttach)
}

func (suite *ModelSuite) Test_UpdateIndexWithAttachedChunks_beta_treasure_complete() {
	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	defer oyster_utils.ResetBrokerMode()

	numChunks := 9

	u := models.UploadSession{
		Type:                 models.SessionTypeBeta,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(9000),
		NumChunks:            numChunks,
		StorageLengthInYears: 2,
	}

	mergedIndexes := []int{5}

	bulkChunkData := SessionSetUpForTest(&u, mergedIndexes, u.NumChunks)

	u.NextIdxToAttach = int64(u.NumChunks - 3)
	suite.DB.ValidateAndUpdate(&u)

	bulkKeys := oyster_utils.GenerateBulkKeys(u.GenesisHash, int64(u.NumChunks-3), 0)
	bulkChunkData, err := models.GetMultiChunkData(oyster_utils.InProgressDir, u.GenesisHash,
		bulkKeys)
	suite.Nil(err)

	u.MoveChunksToCompleted([]oyster_utils.ChunkData{bulkChunkData[2]}) // the index of the treasure chunks

	u.UpdateIndexWithAttachedChunks(bulkChunkData)

	suite.Equal(int64(-1), u.NextIdxToAttach)
}

func (suite *ModelSuite) Test_DownGradeIndexesOnUnattachedChunks_alpha() {
	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	defer oyster_utils.ResetBrokerMode()

	numChunks := 15

	u := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(15000),
		NumChunks:            numChunks,
		StorageLengthInYears: 2,
		NextIdxToAttach:      12,
		NextIdxToVerify:      11,
	}

	suite.DB.ValidateAndUpdate(&u)

	chunkData := []oyster_utils.ChunkData{}
	for i := 6; i < numChunks; i++ {
		chunkData = append(chunkData, oyster_utils.ChunkData{
			Idx: int64(i),
		})
	}

	u.DownGradeIndexesOnUnattachedChunks(chunkData)

	suite.Equal(int64(6), u.NextIdxToVerify)
	suite.Equal(int64(6), u.NextIdxToAttach)
}

func (suite *ModelSuite) Test_DownGradeIndexesOnUnattachedChunks_beta() {
	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	defer oyster_utils.ResetBrokerMode()

	numChunks := 15

	u := models.UploadSession{
		Type:                 models.SessionTypeBeta,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(15000),
		NumChunks:            numChunks,
		StorageLengthInYears: 2,
		NextIdxToAttach:      1,
		NextIdxToVerify:      2,
	}

	suite.DB.ValidateAndUpdate(&u)

	chunkData := []oyster_utils.ChunkData{}
	for i := 6; i > -1; i-- {
		chunkData = append(chunkData, oyster_utils.ChunkData{
			Idx: int64(i),
		})
	}

	u.DownGradeIndexesOnUnattachedChunks(chunkData)

	suite.Equal(int64(6), u.NextIdxToVerify)
	suite.Equal(int64(6), u.NextIdxToAttach)
}

func (suite *ModelSuite) Test_GetCompletedSessions() {
	numChunks := 15

	u1 := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(15000),
		NumChunks:            numChunks,
		StorageLengthInYears: 1,
		NextIdxToAttach:      int64(numChunks),
		NextIdxToVerify:      int64(numChunks),
		AllDataReady:         models.AllDataReady,
		TreasureStatus:       models.TreasureInDataMapComplete,
		PaymentStatus:        models.PaymentStatusConfirmed,
	}

	vErr, err := suite.DB.ValidateAndCreate(&u1)
	suite.False(vErr.HasAny())
	suite.Nil(err)

	u2 := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(15000),
		NumChunks:            numChunks,
		StorageLengthInYears: 1,
		NextIdxToAttach:      int64(numChunks / 2),
		NextIdxToVerify:      int64(numChunks / 2),
		AllDataReady:         models.AllDataReady,
		TreasureStatus:       models.TreasureInDataMapComplete,
		PaymentStatus:        models.PaymentStatusConfirmed,
	}

	vErr, err = suite.DB.ValidateAndCreate(&u2)
	suite.False(vErr.HasAny())
	suite.Nil(err)

	u3 := models.UploadSession{
		Type:                 models.SessionTypeBeta,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(15000),
		NumChunks:            numChunks,
		StorageLengthInYears: 1,
		NextIdxToAttach:      int64(-1),
		NextIdxToVerify:      int64(-1),
		AllDataReady:         models.AllDataReady,
		TreasureStatus:       models.TreasureInDataMapComplete,
		PaymentStatus:        models.PaymentStatusConfirmed,
	}

	vErr, err = suite.DB.ValidateAndCreate(&u3)
	suite.False(vErr.HasAny())
	suite.Nil(err)

	u4 := models.UploadSession{
		Type:                 models.SessionTypeBeta,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(15000),
		NumChunks:            numChunks,
		StorageLengthInYears: 1,
		NextIdxToAttach:      int64(numChunks / 2),
		NextIdxToVerify:      int64(numChunks / 2),
		AllDataReady:         models.AllDataReady,
		TreasureStatus:       models.TreasureInDataMapComplete,
		PaymentStatus:        models.PaymentStatusConfirmed,
	}

	vErr, err = suite.DB.ValidateAndCreate(&u4)
	suite.False(vErr.HasAny())
	suite.Nil(err)

	completedSessions, err := models.GetCompletedSessions()
	suite.Nil(err)
	suite.Equal(2, len(completedSessions))

	for _, session := range completedSessions {
		suite.True(session.GenesisHash == u1.GenesisHash || session.GenesisHash == u3.GenesisHash)
	}
}

func (suite *ModelSuite) Test_GetReadySessions() {
	numChunks := 15

	u1 := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(15000),
		NumChunks:            numChunks,
		StorageLengthInYears: 1,
		NextIdxToAttach:      int64(numChunks + 1),
		NextIdxToVerify:      int64(numChunks + 1),
		AllDataReady:         models.AllDataReady,
		TreasureStatus:       models.TreasureInDataMapComplete,
		PaymentStatus:        models.PaymentStatusConfirmed,
	}

	vErr, err := suite.DB.ValidateAndCreate(&u1)
	suite.False(vErr.HasAny())
	suite.Nil(err)

	u2 := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(15000),
		NumChunks:            numChunks,
		StorageLengthInYears: 1,
		NextIdxToAttach:      int64(numChunks / 2),
		NextIdxToVerify:      int64(numChunks / 2),
		AllDataReady:         models.AllDataReady,
		TreasureStatus:       models.TreasureInDataMapComplete,
		PaymentStatus:        models.PaymentStatusConfirmed,
	}

	vErr, err = suite.DB.ValidateAndCreate(&u2)
	suite.False(vErr.HasAny())
	suite.Nil(err)

	u3 := models.UploadSession{
		Type:                 models.SessionTypeBeta,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(15000),
		NumChunks:            numChunks,
		StorageLengthInYears: 1,
		NextIdxToAttach:      int64(-1),
		NextIdxToVerify:      int64(-1),
		AllDataReady:         models.AllDataReady,
		TreasureStatus:       models.TreasureInDataMapComplete,
		PaymentStatus:        models.PaymentStatusConfirmed,
	}

	vErr, err = suite.DB.ValidateAndCreate(&u3)
	suite.False(vErr.HasAny())
	suite.Nil(err)

	u4 := models.UploadSession{
		Type:                 models.SessionTypeBeta,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(15000),
		NumChunks:            numChunks,
		StorageLengthInYears: 1,
		NextIdxToAttach:      int64(numChunks / 2),
		NextIdxToVerify:      int64(numChunks / 2),
		AllDataReady:         models.AllDataReady,
		TreasureStatus:       models.TreasureInDataMapComplete,
		PaymentStatus:        models.PaymentStatusConfirmed,
	}

	vErr, err = suite.DB.ValidateAndCreate(&u4)
	suite.False(vErr.HasAny())
	suite.Nil(err)

	readySessions, err := models.GetReadySessions()
	suite.Nil(err)
	suite.Equal(2, len(readySessions))

	for _, session := range readySessions {
		suite.True(session.GenesisHash == u2.GenesisHash || session.GenesisHash == u4.GenesisHash)
	}
}

func (suite *ModelSuite) Test_GetChunkForWebnodePoW() {

	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	defer oyster_utils.ResetBrokerMode()

	numChunks := 9

	u := models.UploadSession{
		Type:                 models.SessionTypeBeta,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(9000),
		NumChunks:            numChunks,
		StorageLengthInYears: 2,
		NextIdxToAttach:      int64(numChunks / 2),
		NextIdxToVerify:      int64(numChunks / 2),
		AllDataReady:         models.AllDataReady,
		TreasureStatus:       models.TreasureInDataMapComplete,
		PaymentStatus:        models.PaymentStatusConfirmed,
	}

	mergedIndexes := []int{5}

	SessionSetUpForTest(&u, mergedIndexes, u.NumChunks)

	oldNextIdxToAttach := u.NextIdxToAttach

	chunkForWebnode, err := models.GetChunkForWebnodePoW()
	suite.Nil(err)
	suite.Equal(oldNextIdxToAttach, chunkForWebnode.Idx)

	session := models.UploadSession{}
	suite.DB.First(&session)

	suite.Equal(oldNextIdxToAttach-1, session.NextIdxToAttach)
}

func (suite *ModelSuite) Test_SetTreasureMessage_badger() {
	oyster_utils.SetStorageMode(oyster_utils.DataMapsInBadger)
	defer oyster_utils.ResetDataMapStorageMode()

	treasureIdx := 3
	treasurePayload := "SOMEPAYLOAD"

	u := models.UploadSession{
		Type:                 models.SessionTypeBeta,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(5000),
		NumChunks:            5,
		StorageLengthInYears: 2,
	}

	suite.DB.ValidateAndCreate(&u)

	u.SetTreasureMessage(treasureIdx, treasurePayload, oyster_utils.TestValueTimeToLive)

	chunkData := models.GetSingleChunkData(oyster_utils.InProgressDir, u.GenesisHash, int64(treasureIdx))

	suite.Equal(treasurePayload, chunkData.Message)
}

func (suite *ModelSuite) Test_SetTreasureMessage_sql() {
	oyster_utils.SetStorageMode(oyster_utils.DataMapsInSQL)
	defer oyster_utils.ResetDataMapStorageMode()

	treasureIdx := 4
	treasurePayload := "SOMEPAYLOAD"

	u := models.UploadSession{
		Type:                 models.SessionTypeBeta,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(5000),
		NumChunks:            5,
		StorageLengthInYears: 2,
	}

	suite.DB.ValidateAndCreate(&u)

	u.SetTreasureMessage(treasureIdx, treasurePayload, oyster_utils.TestValueTimeToLive)

	chunkData := models.GetSingleChunkData(oyster_utils.InProgressDir, u.GenesisHash, int64(treasureIdx))

	suite.Equal(treasurePayload, chunkData.Message)
}

func (suite *ModelSuite) Test_GetSingleChunkData_badger() {
	oyster_utils.SetStorageMode(oyster_utils.DataMapsInBadger)
	defer oyster_utils.ResetDataMapStorageMode()

	numChunks := 9

	u := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(9000),
		NumChunks:            numChunks,
		StorageLengthInYears: 2,
	}

	mergedIndexes := []int{5}

	SessionSetUpForTest(&u, mergedIndexes, u.NumChunks)

	chunkData := models.GetSingleChunkData(oyster_utils.InProgressDir, u.GenesisHash, int64(0))

	suite.Equal(u.GenesisHash, chunkData.Hash)
	suite.NotEqual("", chunkData.Message)
}

func (suite *ModelSuite) Test_GetSingleChunkData_sql() {
	oyster_utils.SetStorageMode(oyster_utils.DataMapsInSQL)
	defer oyster_utils.ResetDataMapStorageMode()

	numChunks := 9

	u := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(9000),
		NumChunks:            numChunks,
		StorageLengthInYears: 2,
	}

	mergedIndexes := []int{5}

	SessionSetUpForTest(&u, mergedIndexes, u.NumChunks)

	chunkData := models.GetSingleChunkData(oyster_utils.InProgressDir, u.GenesisHash, int64(0))

	suite.Equal(u.GenesisHash, chunkData.Hash)
	suite.NotEqual("", chunkData.Message)
}

func (suite *ModelSuite) Test_GetMultiChunkData_badger() {
	oyster_utils.SetStorageMode(oyster_utils.DataMapsInBadger)
	defer oyster_utils.ResetDataMapStorageMode()

	numChunks := 9

	u := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(9000),
		NumChunks:            numChunks,
		StorageLengthInYears: 2,
	}

	mergedIndexes := []int{5}

	SessionSetUpForTest(&u, mergedIndexes, u.NumChunks)

	key1 := oyster_utils.GetBadgerKey([]string{u.GenesisHash, strconv.Itoa(0)})
	key2 := oyster_utils.GetBadgerKey([]string{u.GenesisHash, strconv.Itoa(5)})
	key3 := oyster_utils.GetBadgerKey([]string{u.GenesisHash, strconv.Itoa(u.NumChunks - 1)})

	chunkData, err := models.GetMultiChunkData(oyster_utils.InProgressDir, u.GenesisHash,
		&oyster_utils.KVKeys{key1, key2, key3})
	suite.Nil(err)

	suite.NotEqual("", chunkData[0].Hash)
	suite.NotEqual("", chunkData[0].Message)
	suite.NotEqual("", chunkData[1].Hash)
	suite.NotEqual("", chunkData[1].Message)
	suite.NotEqual("", chunkData[2].Hash)
	suite.NotEqual("", chunkData[2].Message)
}

func (suite *ModelSuite) Test_GetMultiChunkData_sql() {
	oyster_utils.SetStorageMode(oyster_utils.DataMapsInSQL)
	defer oyster_utils.ResetDataMapStorageMode()

	oyster_utils.InitKvStore()

	numChunks := 9

	u := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        uint64(9000),
		NumChunks:            numChunks,
		StorageLengthInYears: 2,
	}

	mergedIndexes := []int{5}

	SessionSetUpForTest(&u, mergedIndexes, u.NumChunks)

	key1 := oyster_utils.GetBadgerKey([]string{u.GenesisHash, strconv.Itoa(0)})
	key2 := oyster_utils.GetBadgerKey([]string{u.GenesisHash, strconv.Itoa(5)})
	key3 := oyster_utils.GetBadgerKey([]string{u.GenesisHash, strconv.Itoa(u.NumChunks - 1)})

	chunkData, err := models.GetMultiChunkData(oyster_utils.InProgressDir, u.GenesisHash,
		&oyster_utils.KVKeys{key1, key2, key3})
	suite.Nil(err)

	suite.NotEqual("", chunkData[0].Hash)
	suite.NotEqual("", chunkData[0].Message)
	suite.NotEqual("", chunkData[1].Hash)
	suite.NotEqual("", chunkData[1].Message)
	suite.NotEqual("", chunkData[2].Hash)
	suite.NotEqual("", chunkData[2].Message)
}

package jobs_test

import (
	"math/big"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
)

func (suite *JobsSuite) Test_RemoveUnpaid_uploadSessionsAndDataMap() {

	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	defer oyster_utils.ResetBrokerMode()

	jobs.EthWrapper = services.Eth{
		CheckPRLBalance: func(addr common.Address) *big.Int {
			return big.NewInt(0)
		},
	}
	UploadSessionsAndDataMap_Expired := "aaaaaa"
	UploadSessionsAndDataMap_NoExpired := "bbbbbb"

	addStartUploadSession(suite, UploadSessionsAndDataMap_Expired, models.PaymentStatusInvoiced, true)
	addStartUploadSession(suite, UploadSessionsAndDataMap_NoExpired, models.PaymentStatusInvoiced, false)

	jobs.RemoveUnpaidUploadSession(jobs.PrometheusWrapper)

	verifyData(suite, UploadSessionsAndDataMap_NoExpired, true)
}

func (suite *JobsSuite) Test_RemoveUnpaid_allPaid() {
	AllPaid := "cccccc"

	addStartUploadSession(suite, AllPaid, models.PaymentStatusConfirmed, true)

	jobs.RemoveUnpaidUploadSession(jobs.PrometheusWrapper)

	verifyData(suite, AllPaid, true)
}

func (suite *JobsSuite) Test_RemoveUnpaid_hasBalance() {
	jobs.EthWrapper = services.Eth{
		CheckPRLBalance: func(addr common.Address) *big.Int {
			return big.NewInt(10)
		},
	}
	HasBalance := "dddddd"
	addStartUploadSession(suite, HasBalance, models.PaymentStatusInvoiced, true)

	jobs.RemoveUnpaidUploadSession(jobs.PrometheusWrapper)

	verifyData(suite, HasBalance, true)
}

func (suite *JobsSuite) Test_RemoveUnpaid_OnlyRemoveUploadSession() {
	jobs.EthWrapper = services.Eth{
		CheckPRLBalance: func(addr common.Address) *big.Int {
			return big.NewInt(0)
		},
	}
	OnlyRemoveUploadSession_Expired := "eeeeee"
	OnlyRemoveUploadSession_NoExpired := "ffffff"

	addOnlySession(suite, OnlyRemoveUploadSession_Expired, models.PaymentStatusInvoiced, true)
	addOnlySession(suite, OnlyRemoveUploadSession_NoExpired, models.PaymentStatusInvoiced, false)

	jobs.RemoveUnpaidUploadSession(jobs.PrometheusWrapper)

	verifyData(suite, OnlyRemoveUploadSession_NoExpired, false)
}

func addStartUploadSession(suite *JobsSuite, genesisHash string, paymentStatus int, isExpired bool) {
	session := models.UploadSession{
		GenesisHash:    genesisHash,
		FileSizeBytes:  8000,
		NumChunks:      5,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  paymentStatus,
		TreasureStatus: models.TreasureInDataMapPending,
	}
	_, err := session.StartUploadSession()
	suite.Nil(err)

	mergedIndexes := []int{3}
	privateKeys := []string{"0000000001"}
	session.MakeTreasureIdxMap(mergedIndexes, privateKeys)

	chunkReqs1 := GenerateChunkRequests(session.NumChunks, session.GenesisHash)
	models.ProcessAndStoreChunkData(chunkReqs1, session.GenesisHash, mergedIndexes, models.TestValueTimeToLive)

	session.WaitForAllHashes(500)

	treasureIndexMap, err := session.GetTreasureMap()

	for _, entry := range treasureIndexMap {

		key := oyster_utils.GetBadgerKey([]string{session.GenesisHash, strconv.Itoa(entry.Idx)})
		oyster_utils.BatchSetToUniqueDB([]string{oyster_utils.InProgressDir, session.GenesisHash,
			oyster_utils.MessageDir}, &oyster_utils.KVPairs{key: "someDummyMessage"}, models.TestValueTimeToLive)
	}
	session.TreasureStatus = models.TreasureInDataMapComplete
	models.DB.ValidateAndSave(&session)

	session.WaitForAllMessages(500)

	if isExpired {
		exceedLimitUpdateTime := time.Now().Add(-(jobs.UnpaidExpirationInHour + 1) * time.Hour)
		// Force to update updated_at field
		err = suite.DB.RawQuery("UPDATE upload_sessions SET updated_at = ? WHERE id = ?",
			exceedLimitUpdateTime.Format(oyster_utils.SqlTimeFormat), session.ID).All(&[]models.UploadSession{})
		suite.Nil(err)
	}
}

func addOnlySession(suite *JobsSuite, genesisHash string, paymentStatus int, isExpired bool) {
	session := models.UploadSession{
		GenesisHash:    genesisHash,
		FileSizeBytes:  8000,
		NumChunks:      2,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  paymentStatus,
		TreasureStatus: models.TreasureInDataMapComplete,
	}
	_, err := suite.DB.ValidateAndCreate(&session)
	suite.Nil(err)

	if isExpired {
		exceedLimitUpdateTime := time.Now().Add(-(jobs.UnpaidExpirationInHour + 1) * time.Hour)
		// Force to update updated_at field
		err = suite.DB.RawQuery("UPDATE upload_sessions SET updated_at = ? WHERE id = ?",
			exceedLimitUpdateTime.Format(oyster_utils.SqlTimeFormat), session.ID).All(&[]models.UploadSession{})
		suite.Nil(err)
	}
}

func verifyData(suite *JobsSuite, expectedGenesisHash string, expectToHaveDataMap bool) {
	var sessions []models.UploadSession
	suite.Nil(suite.DB.RawQuery("SELECT * FROM upload_sessions").All(&sessions))
	suite.Equal(1, len(sessions))
	suite.Equal(expectedGenesisHash, sessions[0].GenesisHash)

	if expectToHaveDataMap {

		keys := oyster_utils.GenerateBulkKeys(sessions[0].GenesisHash, 0, int64(sessions[0].NumChunks-1))

		chunkData, _ := oyster_utils.GetBulkChunkData(oyster_utils.InProgressDir, sessions[0].GenesisHash, keys)
		suite.True(len(chunkData) == sessions[0].NumChunks)
		for _, dataMap := range chunkData {
			suite.Equal(expectedGenesisHash, dataMap.GenesisHash)
		}
	}
}

package jobs_test

import (
	"github.com/oysterprotocol/brokernode/utils/eth_gateway"
	"math/big"
	"math/rand"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/utils"
)

func (suite *JobsSuite) Test_RemoveUnpaid_uploadSessionsAndDataMap_badger() {
	oyster_utils.SetStorageMode(oyster_utils.DataMapsInBadger)
	defer oyster_utils.ResetDataMapStorageMode()

	removeAllUploadSessions(suite)

	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	defer oyster_utils.ResetBrokerMode()

	jobs.EthWrapper = eth_gateway.Eth{
		CheckPRLBalance: func(addr common.Address) *big.Int {
			return big.NewInt(0)
		},
	}

	UploadSessionsAndDataMap_Expired := "aaaaaaa1"
	UploadSessionsAndDataMap_NoExpired := "bbbbbbb1"

	addStartUploadSession(suite, UploadSessionsAndDataMap_Expired, models.PaymentStatusInvoiced, true)
	addStartUploadSession(suite, UploadSessionsAndDataMap_NoExpired, models.PaymentStatusInvoiced, false)

	jobs.RemoveUnpaidUploadSession(jobs.PrometheusWrapper)

	verifyData(suite, UploadSessionsAndDataMap_NoExpired, true)
}

func (suite *JobsSuite) Test_RemoveUnpaid_allPaid_badger() {
	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	defer oyster_utils.ResetBrokerMode()

	oyster_utils.SetStorageMode(oyster_utils.DataMapsInBadger)
	defer oyster_utils.ResetDataMapStorageMode()

	removeAllUploadSessions(suite)

	AllPaid := "ccccccc1"

	addStartUploadSession(suite, AllPaid, models.PaymentStatusConfirmed, true)

	jobs.RemoveUnpaidUploadSession(jobs.PrometheusWrapper)

	verifyData(suite, AllPaid, true)
}

func (suite *JobsSuite) Test_RemoveUnpaid_hasBalance_badger() {
	oyster_utils.SetStorageMode(oyster_utils.DataMapsInBadger)
	defer oyster_utils.ResetDataMapStorageMode()

	removeAllUploadSessions(suite)

	jobs.EthWrapper = eth_gateway.Eth{
		CheckPRLBalance: func(addr common.Address) *big.Int {
			return big.NewInt(10)
		},
	}
	HasBalance := "ddddddd1"
	addStartUploadSession(suite, HasBalance, models.PaymentStatusInvoiced, true)

	jobs.RemoveUnpaidUploadSession(jobs.PrometheusWrapper)

	verifyData(suite, HasBalance, true)
}

func (suite *JobsSuite) Test_RemoveUnpaid_OnlyRemoveUploadSession_badger() {
	oyster_utils.SetStorageMode(oyster_utils.DataMapsInBadger)
	defer oyster_utils.ResetDataMapStorageMode()

	removeAllUploadSessions(suite)

	jobs.EthWrapper = eth_gateway.Eth{
		CheckPRLBalance: func(addr common.Address) *big.Int {
			return big.NewInt(0)
		},
	}
	OnlyRemoveUploadSession_Expired := "eeeeeee1"
	OnlyRemoveUploadSession_NoExpired := "fffffff1"

	addOnlySession(suite, OnlyRemoveUploadSession_Expired, models.PaymentStatusInvoiced, true)
	addOnlySession(suite, OnlyRemoveUploadSession_NoExpired, models.PaymentStatusInvoiced, false)

	jobs.RemoveUnpaidUploadSession(jobs.PrometheusWrapper)

	verifyData(suite, OnlyRemoveUploadSession_NoExpired, false)
}

func (suite *JobsSuite) Test_RemoveUnpaid_uploadSessionsAndDataMap_sql() {
	oyster_utils.SetStorageMode(oyster_utils.DataMapsInSQL)
	defer oyster_utils.ResetDataMapStorageMode()

	removeAllUploadSessions(suite)

	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	defer oyster_utils.ResetBrokerMode()

	jobs.EthWrapper = eth_gateway.Eth{
		CheckPRLBalance: func(addr common.Address) *big.Int {
			return big.NewInt(0)
		},
	}
	UploadSessionsAndDataMap_Expired := "aaaaaaa2"
	UploadSessionsAndDataMap_NoExpired := "bbbbbbb2"

	addStartUploadSession(suite, UploadSessionsAndDataMap_Expired, models.PaymentStatusInvoiced, true)
	addStartUploadSession(suite, UploadSessionsAndDataMap_NoExpired, models.PaymentStatusInvoiced, false)

	jobs.RemoveUnpaidUploadSession(jobs.PrometheusWrapper)

	verifyData(suite, UploadSessionsAndDataMap_NoExpired, true)
}

func (suite *JobsSuite) Test_RemoveUnpaid_allPaid_sql() {
	oyster_utils.SetStorageMode(oyster_utils.DataMapsInSQL)
	defer oyster_utils.ResetDataMapStorageMode()

	removeAllUploadSessions(suite)

	AllPaid := "ccccccc2"

	addStartUploadSession(suite, AllPaid, models.PaymentStatusConfirmed, true)

	jobs.RemoveUnpaidUploadSession(jobs.PrometheusWrapper)

	verifyData(suite, AllPaid, true)
}

func (suite *JobsSuite) Test_RemoveUnpaid_hasBalance_sql() {
	oyster_utils.SetStorageMode(oyster_utils.DataMapsInSQL)
	defer oyster_utils.ResetDataMapStorageMode()

	removeAllUploadSessions(suite)

	jobs.EthWrapper = eth_gateway.Eth{
		CheckPRLBalance: func(addr common.Address) *big.Int {
			return big.NewInt(10)
		},
	}
	HasBalance := "ddddddd2"
	addStartUploadSession(suite, HasBalance, models.PaymentStatusInvoiced, true)

	jobs.RemoveUnpaidUploadSession(jobs.PrometheusWrapper)

	verifyData(suite, HasBalance, true)
}

func (suite *JobsSuite) Test_RemoveUnpaid_OnlyRemoveUploadSession_sql() {
	oyster_utils.SetStorageMode(oyster_utils.DataMapsInSQL)
	defer oyster_utils.ResetDataMapStorageMode()

	removeAllUploadSessions(suite)

	jobs.EthWrapper = eth_gateway.Eth{
		CheckPRLBalance: func(addr common.Address) *big.Int {
			return big.NewInt(0)
		},
	}
	OnlyRemoveUploadSession_Expired := "eeeeeee2"
	OnlyRemoveUploadSession_NoExpired := "fffffff2"

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

	_, err := suite.DB.ValidateAndCreate(&session)
	suite.Nil(err)
	models.BuildDataMapsForSession(session.GenesisHash, session.NumChunks)

	mergedIndexes := []int{3}
	key := ""
	for j := 0; j < 10; j++ {
		key += strconv.Itoa(rand.Intn(8) + 1)
	}
	privateKeys := []string{key}
	session.MakeTreasureIdxMap(mergedIndexes, privateKeys)

	chunkReqs1 := GenerateChunkRequests(session.NumChunks, session.GenesisHash)
	models.ProcessAndStoreChunkData(chunkReqs1, session.GenesisHash, mergedIndexes, oyster_utils.TestValueTimeToLive)

	session.WaitForAllHashes(500)

	session.CreateTreasures()

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

		keys := oyster_utils.GenerateBulkKeys(sessions[0].GenesisHash, 0, int64(sessions[0].NumChunks))

		chunkData, _ := models.GetMultiChunkData(oyster_utils.InProgressDir, sessions[0].GenesisHash, keys)

		suite.True(len(chunkData) == sessions[0].NumChunks)
		for _, dataMap := range chunkData {
			suite.Equal(expectedGenesisHash, dataMap.GenesisHash)
		}
	}
}

func removeAllUploadSessions(suite *JobsSuite) {
	suite.DB.RawQuery("DELETE FROM upload_sessions").All(&models.UploadSession{})
}

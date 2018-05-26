package jobs_test

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
)

func (suite *JobsSuite) Test_RemoveUnpaid_uploadSessionsAndDataMap() {
	jobs.EthWrapper = services.Eth{
		CheckBalance: func(addr common.Address) *big.Int {
			return big.NewInt(0)
		},
	}

	addStartUploadSession(suite, "UploadSessionsAndDataMap_Expired", models.PaymentStatusInvoiced, true)
	addStartUploadSession(suite, "UploadSessionsAndDataMap_NoExpired", models.PaymentStatusInvoiced, false)

	jobs.RemoveUnpaidUploadSession()

	verifyData(suite, "UploadSessionsAndDataMap_NoExpired")
}

func (suite *JobsSuite) Test_RemoveUnpaid_allPaid() {
	addStartUploadSession(suite, "AllPaid", models.PaymentStatusConfirmed, true)

	jobs.RemoveUnpaidUploadSession()

	verifyData(suite, "AllPaid")
}

func (suite *JobsSuite) Test_RemoveUnpaid_hasBalance() {
	jobs.EthWrapper = services.Eth{
		CheckBalance: func(addr common.Address) *big.Int {
			return big.NewInt(10)
		},
	}

	addStartUploadSession(suite, "HasBalance", models.PaymentStatusInvoiced, true)

	jobs.RemoveUnpaidUploadSession()

	verifyData(suite, "HasBalance")
}

func (suite *JobsSuite) Test_RemoveUnpaid_OnlyRemoveUploadSession() {
	jobs.EthWrapper = services.Eth{
		CheckBalance: func(addr common.Address) *big.Int {
			return big.NewInt(0)
		},
	}

	addOnlySession(suite, "OnlyRemoveUploadSession_Expired", models.PaymentStatusInvoiced, true)
	addOnlySession(suite, "OnlyRemoveUploadSession_NoExpired", models.PaymentStatusInvoiced, false)

	jobs.RemoveUnpaidUploadSession()

	verifyData(suite, "OnlyRemoveUploadSession_NoExpired")
}

func addStartUploadSession(suite *JobsSuite, genesisHash string, paymentStatus int, isExpired bool) {
	session := models.UploadSession{
		GenesisHash:    genesisHash,
		FileSizeBytes:  8000,
		NumChunks:      2,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  paymentStatus,
		TreasureStatus: models.TreasureInDataMapComplete,
	}
	_, err := session.StartUploadSession()
	suite.Nil(err)

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

func verifyData(suite *JobsSuite, expectedDenesisHash string) {
	var sessions []models.UploadSession
	suite.Nil(suite.DB.RawQuery("SELECT * from upload_sessions").All(&sessions))
	suite.Equal(1, len(sessions))
	suite.Equal(expectedDenesisHash, sessions[0].GenesisHash)

	var dataMaps []models.DataMap
	suite.Nil(suite.DB.RawQuery("SELECT * from data_maps").All(&dataMaps))
	for _, dataMap := range dataMaps {
		suite.Equal(expectedDenesisHash, dataMap.GenesisHash)
	}
}

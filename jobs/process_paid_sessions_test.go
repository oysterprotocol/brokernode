package jobs_test

import (
	"github.com/gobuffalo/pop/nulls"
	//"github.com/oysterprotocol/brokernode/jobs"
	//"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/models"
)

func (suite *JobsSuite) Test_ProcessPaidSessions() {
	fileBytesCount := 750000

	uploadSession1 := models.UploadSession{
		GenesisHash:    "genHash1",
		FileSizeBytes:  fileBytesCount,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  models.PaymentStatusPaid,
		TreasureStatus: models.TreasureUnburied,
		TreasureIdxMap: nulls.String{"[1, 220, 355]", true},
	}

	uploadSession1.StartUploadSession()

	uploadSession2 := models.UploadSession{
		GenesisHash:    "genHash2",
		FileSizeBytes:  fileBytesCount,
		Type:           models.SessionTypeAlpha,
		PaymentStatus:  models.PaymentStatusPaid,
		TreasureStatus: models.TreasureBuried,
		TreasureIdxMap: nulls.String{"[45, 278, 305]", true},
	}

	uploadSession2.StartUploadSession()

	paidButUnburied := []models.DataMap{}
	err := suite.DB.Where("genesis_hash = ?", "genHash1").All(&paidButUnburied)
	suite.Equal(err, nil)

	paidAndUnburied := []models.DataMap{}
	err = suite.DB.Where("genesis_hash = ?", "genHash1").All(&paidAndUnburied)
	suite.Equal(err, nil)

	suite.NotEqual(0, len(paidButUnburied))
	suite.NotEqual(0, len(paidAndUnburied))

	for _, dMap := range paidButUnburied {
		suite.Equal("", dMap.Message)
	}

	jobs.ProcessPaidSessions()

	paidButUnburied = []models.DataMap{}
	err = suite.DB.Where("genesis_hash = ?", "genHash1").All(&paidButUnburied)
	suite.Equal(err, nil)

	/*@TODO after getting the slices and "where in" queries working, verify that the chunks with the correct
	indexes no longer have "" for their "message" fields and verify that all data map statuses have changed to
	"Unassigned"
	*/
}

package jobs_test

import (
//"github.com/oysterprotocol/brokernode/jobs"
//"github.com/oysterprotocol/brokernode/models"
)

func (suite *JobsSuite) Test_ProcessPaidSessions() {
	//fileBytesCount := 3500
	//
	//uploadSession1 := models.UploadSession{
	//	GenesisHash:    "genHash1",
	//	FileSizeBytes:  fileBytesCount,
	//	Type:           models.SessionTypeAlpha,
	//	PaymentStatus:  models.Paid,
	//	TreasureStatus: models.Unburied,
	//}
	//
	//uploadSession1.StartUploadSession()
	//
	//uploadSession2 := models.UploadSession{
	//	GenesisHash:    "genHash2",
	//	FileSizeBytes:  fileBytesCount,
	//	Type:           models.SessionTypeAlpha,
	//	PaymentStatus:  models.Paid,
	//	TreasureStatus: models.Buried,
	//}
	//
	//uploadSession2.StartUploadSession()
	//
	//paidButUnburied := []models.DataMap{}
	//err := suite.DB.Where("genesis_hash = ?", "genHash1").All(&paidButUnburied)
	//suite.Equal(err, nil)
	//
	//paidAndUnburied := []models.DataMap{}
	//err = suite.DB.Where("genesis_hash = ?", "genHash1").All(&paidAndUnburied)
	//suite.Equal(err, nil)
	//
	//suite.NotEqual(0, len(paidButUnburied))
	//suite.NotEqual(0, len(paidAndUnburied))
	//
	//for _, dMap := range paidButUnburied {
	//	suite.Equal("", dMap.Message)
	//}
	//
	//jobs.ProcessPaidSessions()
	//
	//paidButUnburied = []models.DataMap{}
	//err = suite.DB.Where("genesis_hash = ?", "genHash1").All(&paidButUnburied)
	//suite.Equal(err, nil)
	////for _, dMap := range paidButUnburied {
	//suite.NotEqual("", paidButUnburied[0].Message)
}

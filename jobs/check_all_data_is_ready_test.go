package jobs_test

import (
	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/utils"
)

func (suite *JobsSuite) Test_CheckSessionsWithIncompleteData() {
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

	SessionSetUpForTest(&u, mergedIndexes, u.NumChunks)

	u.AllDataReady = models.AllDataNotReady
	suite.DB.ValidateAndUpdate(&u)

	session := &models.UploadSession{}
	models.DB.Find(session, u.ID)

	suite.Equal(models.AllDataNotReady, session.AllDataReady)

	jobs.CheckSessionsWithIncompleteData()

	session = &models.UploadSession{}
	models.DB.Find(session, u.ID)

	suite.Equal(models.AllDataReady, session.AllDataReady)
}

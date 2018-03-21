package jobs_test

import (
	"github.com/oysterprotocol/brokernode/models"
)

func (suite *JobsSuite) Test_VerifyDataMaps() {
	genHash := "someGenHash"
	fileBytesCount := 30000

	vErr, err := models.BuildDataMaps(genHash, fileBytesCount)
	suite.Nil(err)
	suite.Equal(0, len(vErr.Errors))

	allDataMaps := []models.DataMap{}

	err = suite.DB.All(&allDataMaps)
	suite.Equal(12, len(allDataMaps))

	// make half the data maps "Unverified"
	for i := 0; i < (len(allDataMaps)/2); i++ {
		allDataMaps[i].Status = models.Unverified
		suite.DB.ValidateAndSave(&allDataMaps[i])
	}

	// will have to wait until we can mock these iota methods
}
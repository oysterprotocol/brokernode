package jobs_test

import (
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/jobs"
	"time"
)

func (suite *JobsSuite) Test_UpdateTimedOutDataMaps() {

	// populate data_maps
	genHash := "someGenHash"
	fileBytesCount := 18000

	vErr, err := models.BuildDataMaps(genHash, fileBytesCount)
	suite.Nil(err)
	suite.Equal(0, len(vErr.Errors))

	// check that it is the length we expect
	allDataMaps := []models.DataMap{}
	err = suite.DB.All(&allDataMaps)
	suite.Equal(8, len(allDataMaps))

	// make data maps unverified
	for i := 0; i < 8; i++ {
		allDataMaps[i].Status = models.Unverified
		suite.DB.ValidateAndSave(&allDataMaps[i])
	}

	// call method under test, passing in our mock of our iota methods
	jobs.UpdateTimeOutDataMaps(time.Now().Add(60 * time.Second))

	allDataMaps = []models.DataMap{}
	err = suite.DB.All(&allDataMaps)

	suite.Equal(8, len(allDataMaps))

	for _, dataMap := range allDataMaps {
		suite.Equal(models.Unassigned, dataMap.Status)
	}
}

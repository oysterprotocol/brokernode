package jobs_test

import (
	"time"

	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
)

func (suite *JobsSuite) Test_UpdateTimedOutDataMaps() {

	// populate data_maps
	genHash := "abcdef"
	numChunks := 10

	vErr, err := models.BuildDataMaps(genHash, numChunks)
	suite.Nil(err)
	suite.False(vErr.HasAny())

	// check that it is the length we expect
	allDataMaps := []models.DataMap{}
	err = suite.DB.All(&allDataMaps)
	suite.Equal(numChunks, len(allDataMaps))

	// make data maps unverified
	for i := 0; i < 10; i++ {
		allDataMaps[i].Status = models.Unverified
		suite.DB.ValidateAndSave(&allDataMaps[i])
	}

	// call method under test, passing in our mock of our iota methods
	jobs.UpdateTimeOutDataMaps(time.Now().Add(60*time.Second), jobs.PrometheusWrapper)

	allDataMaps = []models.DataMap{}
	err = suite.DB.All(&allDataMaps)

	suite.Equal(numChunks, len(allDataMaps))

	for _, dataMap := range allDataMaps {
		if services.GetMessageFromDataMap(dataMap) != "" {
			// if no message, will not mark as Unassigned
			// the treasure chunk currently has no message
			suite.Equal(models.Unassigned, dataMap.Status)
		}
	}
}

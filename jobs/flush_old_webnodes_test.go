package jobs_test

import (
	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/models"
	"time"
)

func (suite *JobsSuite) Test_FlushOldWebnodes() {

	// Testing that old webnodes get removed

	_, err := models.DB.ValidateAndCreate(&models.Webnode{Address: "oldWebnodeC"})
	_, err = models.DB.ValidateAndCreate(&models.Webnode{Address: "oldWebnodeD"})

	// setting the threshold time to be more recent than when the webnodes were updated
	thresholdTime := time.Now().Add(time.Millisecond * 10000)

	suite.Nil(err)

	webnodes := []models.Webnode{}

	// check that there are 2 webnodes
	err = suite.DB.All(&webnodes)
	suite.Nil(err)
	suite.Equal(2, len(webnodes))

	// call method under test
	jobs.FlushOldWebNodes(thresholdTime, jobs.PrometheusWrapper)

	webnodes = []models.Webnode{}

	// checking that all webnodes have been removed
	err = suite.DB.All(&webnodes)
	suite.Nil(err)
	suite.Equal(0, len(webnodes))

	// Testing that it doesn't remove new webnodes

	_, err = models.DB.ValidateAndCreate(&models.Webnode{Address: "newWebnodeA"})
	_, err = models.DB.ValidateAndCreate(&models.Webnode{Address: "newWebnodeB"})

	// setting the threshold time to be less recent than when the webnodes were updated
	thresholdTime = time.Now().Add(time.Millisecond * -10000)

	suite.Nil(err)

	webnodes = []models.Webnode{}

	// check that there are 2 webnodes
	err = suite.DB.All(&webnodes)
	suite.Nil(err)
	suite.Equal(2, len(webnodes))

	// call method under test
	jobs.FlushOldWebNodes(thresholdTime, jobs.PrometheusWrapper)

	webnodes = []models.Webnode{}

	// checking that all webnodes are still there
	err = suite.DB.All(&webnodes)
	suite.Nil(err)
	suite.Equal(2, len(webnodes))
}

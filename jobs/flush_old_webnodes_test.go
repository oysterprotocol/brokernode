package jobs_test

import (
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/jobs"
	"time"
)

func (suite *JobsSuite) Test_FlushOldWebnodes() {

	_, err := models.DB.ValidateAndCreate(&models.Webnode{WebnodeID: "oldWebnodeC"})
	_, err = models.DB.ValidateAndCreate(&models.Webnode{WebnodeID: "oldWebnodeD"})

	thresholdTime := time.Now().Add(time.Millisecond * 400)
	time.Sleep(time.Millisecond * 400)

	_, err = models.DB.ValidateAndCreate(&models.Webnode{WebnodeID: "newWebnodeA"})
	_, err = models.DB.ValidateAndCreate(&models.Webnode{WebnodeID: "newWebnodeB"})

	suite.Equal(err, nil)

	webnodes := []models.Webnode{}

	// check that there are 4 webnodes
	err = suite.DB.All(&webnodes)
	suite.Equal(err, nil)
	suite.Equal(4, len(webnodes))

	// call method under test
	jobs.FlushOldWebNodes(thresholdTime)

	webnodes = []models.Webnode{}

	// checking that there are only 2 webnodes
	err = suite.DB.All(&webnodes)
	suite.Equal(err, nil)
	suite.Equal(2, len(webnodes))
}

package jobs_test

import (
	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/models"
)

func (suite *JobsSuite) Test_ProcessUnassignedChunks() {
	genHash := "genHashTest"
	fileBytesCount := 8500

	vErr, err := models.BuildDataMaps(genHash, fileBytesCount)
	suite.Nil(err)
	suite.Equal(0, len(vErr.Errors))

	dataMaps := []models.DataMap{}
	suite.DB.All(&dataMaps)

	unassignedTimedOut := models.DataMap{}
	unassignedTimedOut = dataMaps[0]
	unassignedTimedOut.Status = models.Unassigned
	unassignedTimedOut.Message = "unassignedTimedOut"
	suite.DB.ValidateAndSave(&unassignedTimedOut)

	unassignedNotTimedOut := models.DataMap{}
	unassignedNotTimedOut = dataMaps[1]
	unassignedNotTimedOut.Status = models.Unassigned
	unassignedNotTimedOut.Message = "unassignedNotTimedOut"
	suite.DB.ValidateAndSave(&unassignedNotTimedOut)

	assignedTimedOut := models.DataMap{}
	assignedTimedOut = dataMaps[2]
	assignedTimedOut.Status = models.Pending
	assignedTimedOut.Message = "assignedTimedOut"
	suite.DB.ValidateAndSave(&assignedTimedOut)

	result, err := jobs.GetUnassignedChunks()

	suite.Equal(2, len(result))

	for _, dMap := range result {
		suite.Equal(models.Unassigned, dMap.Status)
	}
}

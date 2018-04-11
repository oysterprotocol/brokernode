package jobs_test

import (
	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/models"
)

func (suite *JobsSuite) Test_PurgeCompletedSessions() {
	fileBytesCount := 2500

	uploadSession1 := models.UploadSession{
		GenesisHash:   "genHash1",
		FileSizeBytes: fileBytesCount,
		Type:          models.SessionTypeAlpha,
	}

	vErr, err := uploadSession1.StartUploadSession()
	suite.Equal(0, len(vErr.Errors))
	suite.Equal(err, nil)

	uploadSession2 := models.UploadSession{
		GenesisHash:   "genHash2",
		FileSizeBytes: fileBytesCount,
		Type:          models.SessionTypeAlpha,
	}

	vErr, err = uploadSession2.StartUploadSession()
	suite.Equal(0, len(vErr.Errors))
	suite.Equal(err, nil)

	uploadSession3 := models.UploadSession{
		GenesisHash:   "genHash3",
		FileSizeBytes: fileBytesCount,
		Type:          models.SessionTypeAlpha,
	}

	vErr, err = uploadSession3.StartUploadSession()
	suite.Equal(0, len(vErr.Errors))
	suite.Equal(err, nil)

	allDataMaps := []models.DataMap{}
	err = suite.DB.All(&allDataMaps)
	suite.Equal(err, nil)

	completedDataMaps := []models.CompletedDataMap{}
	err = suite.DB.All(&completedDataMaps)
	suite.Equal(err, nil)

	uploadSessions := []models.UploadSession{}
	err = suite.DB.All(&uploadSessions)
	suite.Equal(err, nil)

	storedGenHashes := []models.StoredGenesisHash{}
	err = suite.DB.All(&storedGenHashes)
	suite.Equal(err, nil)

	// verify initial lengths are what we expected
	suite.Equal(9, len(allDataMaps))
	suite.Equal(0, len(completedDataMaps))
	suite.Equal(3, len(uploadSessions))
	suite.Equal(0, len(storedGenHashes))

	// set all chunks of first data map to complete or confirmed
	allDone := []models.DataMap{}
	err = suite.DB.Where("genesis_hash = ?", "genHash1").All(&allDone)
	suite.Equal(err, nil)

	for _, dataMap := range allDone {
		dataMap.Status = models.Complete
		suite.DB.ValidateAndSave(&dataMap)
	}

	// set one of them to "confirmed"
	allDone[1].Status = models.Confirmed
	suite.DB.ValidateAndSave(&allDone[1])

	// set one chunk of second data map to complete
	someDone := []models.DataMap{}
	err = suite.DB.Where("genesis_hash = ?", "genHash2").All(&someDone)
	suite.Equal(err, nil)

	someDone[0].Status = models.Complete
	suite.DB.ValidateAndSave(&someDone[0])

	//call method under test
	jobs.PurgeCompletedSessions()

	allDataMaps = []models.DataMap{}
	err = suite.DB.All(&allDataMaps)
	suite.Equal(err, nil)

	completedDataMaps = []models.CompletedDataMap{}
	err = suite.DB.All(&completedDataMaps)
	suite.Equal(err, nil)

	uploadSessions = []models.UploadSession{}
	err = suite.DB.All(&uploadSessions)
	suite.Equal(err, nil)

	storedGenHashes = []models.StoredGenesisHash{}
	err = suite.DB.All(&storedGenHashes)
	suite.Equal(err, nil)

	// verify final lengths are what we expected
	suite.Equal(6, len(allDataMaps))
	suite.Equal(3, len(completedDataMaps))
	suite.Equal(2, len(uploadSessions))
	suite.Equal(1, len(storedGenHashes))

	// for good measure, verify that it's only "genHash1" in completed_data_maps
	// and that "genHash1" is not in data_maps at all
	genHash1InDataMaps := []models.DataMap{}
	err = suite.DB.Where("genesis_hash = ?", "genHash1").All(&genHash1InDataMaps)
	suite.Equal(0, len(genHash1InDataMaps))
	suite.Equal(err, nil)

	genHash1Completed := []models.CompletedDataMap{}
	err = suite.DB.Where("genesis_hash = ?", "genHash1").All(&genHash1Completed)
	suite.Equal(3, len(genHash1Completed))
	suite.Equal(err, nil)
}

package jobs_test

import (
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/jobs"
)

func (suite *JobsSuite) Test_PurgeCompletedSessions() {
	fileBytesCount := 2500

	uploadSession1 := models.UploadSession{
		GenesisHash:   "genHash1",
		FileSizeBytes: fileBytesCount,
		Type:          models.SessionTypeAlpha,
	}

	uploadSession1.StartUploadSession()

	uploadSession2 := models.UploadSession{
		GenesisHash:   "genHash2",
		FileSizeBytes: fileBytesCount,
		Type:          models.SessionTypeAlpha,
	}

	uploadSession2.StartUploadSession()

	uploadSession3 := models.UploadSession{
		GenesisHash:   "genHash3",
		FileSizeBytes: fileBytesCount,
		Type:          models.SessionTypeAlpha,
	}

	uploadSession3.StartUploadSession()

	allDataMaps := []models.DataMap{}
	err := suite.DB.All(&allDataMaps)
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
	suite.Equal(6, len(allDataMaps))
	suite.Equal(0, len(completedDataMaps))
	suite.Equal(3, len(uploadSessions))
	suite.Equal(0, len(storedGenHashes))

	// set all chunks of first data map to complete or confirmed
	allDone := []models.DataMap{}
	err = suite.DB.Where("genesis_hash = ?", "genHash1").All(&allDone)
	suite.Equal(err, nil)

	allDone[0].Status = models.Complete
	allDone[1].Status = models.Confirmed

	suite.DB.ValidateAndSave(&allDone[0])
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
	suite.Equal(4, len(allDataMaps))
	suite.Equal(2, len(completedDataMaps))
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
	suite.Equal(2, len(genHash1Completed))
	suite.Equal(err, nil)
}

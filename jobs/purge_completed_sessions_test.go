package jobs_test

import (
	"github.com/gobuffalo/pop/nulls"
	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/utils"
)

func (suite *JobsSuite) Test_PurgeCompletedSessions() {

	// running these tests in TestModeNoTreasure so we will not expect
	// numChunks + 1 in subsequent tests, just numChunks
	oyster_utils.SetBrokerMode(oyster_utils.TestModeNoTreasure)
	defer oyster_utils.ResetBrokerMode()

	fileBytesCount := uint64(30000)
	numChunks := 30
	privateKey := "1111111111111111111111111111111111111111111111111111111111111111"
	genHash := oyster_utils.RandSeq(8, []rune("abcdef0123456789"))

	uploadSession1 := models.UploadSession{
		GenesisHash:   genHash,
		FileSizeBytes: fileBytesCount,
		NumChunks:     numChunks,
		Type:          models.SessionTypeBeta,
		ETHAddrAlpha:  nulls.String{string("SOME_ALPHA_ETH_ADDRESS1"), true},
		ETHAddrBeta:   nulls.String{string("SOME_BETA_ETH_ADDRESS1"), true},
		ETHPrivateKey: privateKey,
	}

	SessionSetUpForTest(&uploadSession1, []int{15, 20}, uploadSession1.NumChunks)

	genHash = oyster_utils.RandSeq(8, []rune("abcdef0123456789"))

	uploadSession2 := models.UploadSession{
		GenesisHash:   genHash,
		FileSizeBytes: fileBytesCount,
		NumChunks:     numChunks,
		Type:          models.SessionTypeAlpha,
		ETHAddrAlpha:  nulls.String{string("SOME_ALPHA_ETH_ADDRESS2"), true},
		ETHAddrBeta:   nulls.String{string("SOME_BETA_ETH_ADDRESS2"), true},
		ETHPrivateKey: privateKey,
	}

	SessionSetUpForTest(&uploadSession2, []int{11, 22}, uploadSession2.NumChunks)

	genHash = oyster_utils.RandSeq(8, []rune("abcdef0123456789"))

	uploadSession3 := models.UploadSession{
		GenesisHash:   genHash,
		FileSizeBytes: fileBytesCount,
		NumChunks:     numChunks,
		Type:          models.SessionTypeAlpha,
		ETHAddrAlpha:  nulls.String{string("SOME_ALPHA_ETH_ADDRESS3"), true},
		ETHAddrBeta:   nulls.String{string("SOME_BETA_ETH_ADDRESS3"), true},
		ETHPrivateKey: privateKey,
	}

	SessionSetUpForTest(&uploadSession3, []int{12, 18}, uploadSession3.NumChunks)

	finished1, _ := uploadSession1.WaitForAllHashes(500)
	finished2, _ := uploadSession2.WaitForAllHashes(500)
	finished3, _ := uploadSession3.WaitForAllHashes(500)

	suite.True(finished1 && finished2 && finished3)

	finished1, _ = uploadSession1.WaitForAllMessages(500)
	finished2, _ = uploadSession2.WaitForAllMessages(500)
	finished3, _ = uploadSession3.WaitForAllMessages(500)

	suite.True(finished1 && finished2 && finished3)

	// Set all sessions to states that will cause them to be picked up by
	// GetSessionsByAge
	sessions := []models.UploadSession{}
	suite.DB.All(&sessions)
	for _, session := range sessions {
		session.PaymentStatus = models.PaymentStatusConfirmed
		session.TreasureStatus = models.TreasureInDataMapComplete
		session.AllDataReady = models.AllDataReady
		suite.DB.ValidateAndUpdate(&session)
	}

	session1Keys := oyster_utils.GenerateBulkKeys(uploadSession1.GenesisHash, 0,
		int64(uploadSession1.NumChunks-1))
	chunksSession1InProgress, _ := oyster_utils.GetBulkChunkData(oyster_utils.InProgressDir, uploadSession1.GenesisHash,
		session1Keys)
	chunksSession1Completed, _ := oyster_utils.GetBulkChunkData(oyster_utils.CompletedDir, uploadSession1.GenesisHash,
		session1Keys)

	session2Keys := oyster_utils.GenerateBulkKeys(uploadSession2.GenesisHash, 0,
		int64(uploadSession2.NumChunks-1))
	chunksSession2InProgress, _ := oyster_utils.GetBulkChunkData(oyster_utils.InProgressDir, uploadSession2.GenesisHash,
		session2Keys)
	chunksSession2Completed, _ := oyster_utils.GetBulkChunkData(oyster_utils.CompletedDir, uploadSession2.GenesisHash,
		session2Keys)

	session3Keys := oyster_utils.GenerateBulkKeys(uploadSession3.GenesisHash, 0,
		int64(uploadSession3.NumChunks-1))
	chunksSession3InProgress, _ := oyster_utils.GetBulkChunkData(oyster_utils.InProgressDir, uploadSession3.GenesisHash,
		session3Keys)
	chunksSession3Completed, _ := oyster_utils.GetBulkChunkData(oyster_utils.CompletedDir, uploadSession3.GenesisHash,
		session3Keys)

	uploadSessions := []models.UploadSession{}
	err := suite.DB.All(&uploadSessions)
	suite.Nil(err)

	completedUploads := []models.CompletedUpload{}
	err = suite.DB.All(&completedUploads)
	suite.Nil(err)

	// verify initial lengths are what we expected
	suite.Equal(numChunks, len(chunksSession1InProgress))
	suite.Equal(numChunks, len(chunksSession2InProgress))
	suite.Equal(numChunks, len(chunksSession3InProgress))

	suite.Equal(0, len(chunksSession1Completed))
	suite.Equal(0, len(chunksSession2Completed))
	suite.Equal(0, len(chunksSession3Completed))

	suite.Equal(3, len(uploadSessions))
	suite.Equal(0, len(completedUploads))

	// set first session's NextIdxToVerify so that it will be regarded as complete
	firstSession := models.UploadSession{}
	suite.DB.Where("genesis_hash = ?", uploadSession1.GenesisHash).First(&firstSession)
	firstSession.NextIdxToAttach = -1
	firstSession.NextIdxToVerify = -1
	vErr, err := suite.DB.ValidateAndSave(&firstSession)
	suite.False(vErr.HasAny())
	suite.Nil(err)

	// set second session's indexes to midway through the map
	secondSession := models.UploadSession{}
	suite.DB.Where("genesis_hash = ?", uploadSession2.GenesisHash).First(&secondSession)
	secondSession.NextIdxToAttach = 16
	secondSession.NextIdxToVerify = 15
	suite.DB.ValidateAndSave(&secondSession)

	//call method under test
	jobs.PurgeCompletedSessions(jobs.PrometheusWrapper)

	chunksSession1InProgress, _ = oyster_utils.GetBulkChunkData(oyster_utils.InProgressDir, uploadSession1.GenesisHash,
		session1Keys)
	chunksSession1Completed, _ = oyster_utils.GetBulkChunkData(oyster_utils.CompletedDir, uploadSession1.GenesisHash,
		session1Keys)

	chunksSession2InProgress, _ = oyster_utils.GetBulkChunkData(oyster_utils.InProgressDir, uploadSession2.GenesisHash,
		session2Keys)
	chunksSession2Completed, _ = oyster_utils.GetBulkChunkData(oyster_utils.CompletedDir, uploadSession2.GenesisHash,
		session2Keys)

	chunksSession3InProgress, _ = oyster_utils.GetBulkChunkData(oyster_utils.InProgressDir, uploadSession3.GenesisHash,
		session3Keys)
	chunksSession3Completed, _ = oyster_utils.GetBulkChunkData(oyster_utils.CompletedDir, uploadSession3.GenesisHash,
		session3Keys)

	uploadSessions = []models.UploadSession{}
	err = suite.DB.All(&uploadSessions)
	suite.Nil(err)

	completedUploads = []models.CompletedUpload{}
	err = suite.DB.All(&completedUploads)
	suite.Nil(err)

	// verify final lengths are what we expected
	suite.Equal(0, len(chunksSession1InProgress))
	suite.Equal(numChunks, len(chunksSession1Completed))
	suite.Equal(0, len(chunksSession2Completed))
	suite.Equal(numChunks, len(chunksSession2InProgress))
	suite.Equal(0, len(chunksSession3Completed))
	suite.Equal(numChunks, len(chunksSession3InProgress))
	suite.Equal(2, len(uploadSessions))
	suite.Equal(1, len(completedUploads))

	suite.Equal("SOME_BETA_ETH_ADDRESS1", completedUploads[0].ETHAddr)
	suite.Equal(uploadSession1.GenesisHash, completedUploads[0].GenesisHash)
}

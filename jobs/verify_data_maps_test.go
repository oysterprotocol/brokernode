package jobs_test

import (
	"github.com/gobuffalo/pop/nulls"
	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"time"
)

var (
	verifyChunkMessagesMatchesRecordMockCalled_verify = false
)

func (suite *JobsSuite) Test_VerifyDataMaps_verify_all_beta() {
	// make suite available inside mock methods
	Suite = *suite

	verifyChunkMessagesMatchesRecordMockCalled_verify = false

	// assign the mock methods for this test
	// return all chunks as being attached to the tangle, i.e. all verified
	IotaMock.VerifyChunkMessagesMatchRecord = func(chunks []oyster_utils.ChunkData) (filteredChunks services.FilteredChunk, err error) {

		// our mock was called
		verifyChunkMessagesMatchesRecordMockCalled_verify = true

		matchesTangle := []oyster_utils.ChunkData{}
		doesNotMatchTangle := []oyster_utils.ChunkData{}
		notAttached := []oyster_utils.ChunkData{}

		matchesTangle = append(matchesTangle, chunks...)

		return services.FilteredChunk{
			MatchesTangle:      matchesTangle,
			NotAttached:        notAttached,
			DoesNotMatchTangle: doesNotMatchTangle,
		}, err
	}

	models.MakeChannels(3)

	// populate data_maps
	numChunks := 10

	uploadSession1 := models.UploadSession{
		GenesisHash:   oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes: uint64(numChunks * 1000),
		NumChunks:     numChunks,
		Type:          models.SessionTypeBeta,
		ETHAddrAlpha:  nulls.String{string("SOME_ALPHA_ETH_ADDRESS1"), true},
		ETHAddrBeta:   nulls.String{string("SOME_BETA_ETH_ADDRESS1"), true},
		ETHPrivateKey: "000000000000005432",
	}

	SessionSetUpForTest(&uploadSession1, []int{5}, uploadSession1.NumChunks)

	session1Keys := oyster_utils.GenerateBulkKeys(uploadSession1.GenesisHash, 0,
		int64(uploadSession1.NumChunks))
	chunksSession1InProgress, _ := models.GetMultiChunkData(oyster_utils.InProgressDir, uploadSession1.GenesisHash,
		session1Keys)
	chunksSession1Completed, _ := models.GetMultiChunkData(oyster_utils.CompletedDir, uploadSession1.GenesisHash,
		session1Keys)

	// check that it is the length we expect
	suite.Equal(numChunks+1, len(chunksSession1InProgress))
	suite.Equal(0, len(chunksSession1Completed))

	// set all data maps to need verifying
	session := models.UploadSession{}
	suite.DB.First(&session)
	session.PaymentStatus = models.PaymentStatusConfirmed
	session.TreasureStatus = models.TreasureInDataMapComplete
	session.AllDataReady = models.AllDataReady
	suite.DB.ValidateAndUpdate(&session)

	session.NextIdxToAttach = -1
	session.NextIdxToVerify = int64(session.NumChunks - 1)
	suite.DB.ValidateAndUpdate(&session)

	suite.Equal(int64(session.NumChunks-1), session.NextIdxToVerify)

	// call method under test, passing in our mock of our iota methods
	jobs.VerifyDataMaps(IotaMock, jobs.PrometheusWrapper, time.Now().Add(1*time.Minute))

	// verify the mock methods were called
	suite.True(verifyChunkMessagesMatchesRecordMockCalled_verify)

	session = models.UploadSession{}
	suite.DB.First(&session)

	suite.Equal(int64(-1), session.NextIdxToVerify)

	suite.True(verifyChunkMessagesMatchesRecordMockCalled_verify)
}

func (suite *JobsSuite) Test_VerifyDataMaps_verify_some_beta() {
	// make suite available inside mock methods
	Suite = *suite

	verifyChunkMessagesMatchesRecordMockCalled_verify = false

	// assign the mock methods for this test
	// return all chunks as being attached to the tangle, i.e. all verified
	IotaMock.VerifyChunkMessagesMatchRecord = func(chunks []oyster_utils.ChunkData) (filteredChunks services.FilteredChunk, err error) {

		// our mock was called
		verifyChunkMessagesMatchesRecordMockCalled_verify = true

		matchesTangle := []oyster_utils.ChunkData{}
		doesNotMatchTangle := []oyster_utils.ChunkData{}
		notAttached := []oyster_utils.ChunkData{}

		matchesTangle = append(matchesTangle, chunks[0], chunks[1], chunks[2])
		doesNotMatchTangle = append(doesNotMatchTangle, chunks[3])
		for i := 4; i < len(chunks); i++ {
			notAttached = append(notAttached, chunks[i])
		}

		return services.FilteredChunk{
			MatchesTangle:      matchesTangle,
			NotAttached:        notAttached,
			DoesNotMatchTangle: doesNotMatchTangle,
		}, err
	}

	models.MakeChannels(3)

	// populate data_maps
	numChunks := 10

	uploadSession1 := models.UploadSession{
		GenesisHash:   oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes: uint64(numChunks * 1000),
		NumChunks:     numChunks,
		Type:          models.SessionTypeBeta,
		ETHAddrAlpha:  nulls.String{string("SOME_ALPHA_ETH_ADDRESS1"), true},
		ETHAddrBeta:   nulls.String{string("SOME_BETA_ETH_ADDRESS1"), true},
		ETHPrivateKey: "000000000000005432",
	}

	SessionSetUpForTest(&uploadSession1, []int{5}, uploadSession1.NumChunks)

	session1Keys := oyster_utils.GenerateBulkKeys(uploadSession1.GenesisHash, 0,
		int64(uploadSession1.NumChunks))
	chunksSession1InProgress, _ := models.GetMultiChunkData(oyster_utils.InProgressDir, uploadSession1.GenesisHash,
		session1Keys)
	chunksSession1Completed, _ := models.GetMultiChunkData(oyster_utils.CompletedDir, uploadSession1.GenesisHash,
		session1Keys)

	// check that it is the length we expect
	suite.Equal(numChunks+1, len(chunksSession1InProgress))
	suite.Equal(0, len(chunksSession1Completed))

	// set all data maps to need verifying
	session := models.UploadSession{}
	suite.DB.First(&session)
	session.PaymentStatus = models.PaymentStatusConfirmed
	session.TreasureStatus = models.TreasureInDataMapComplete
	session.AllDataReady = models.AllDataReady
	suite.DB.ValidateAndUpdate(&session)

	session.NextIdxToAttach = -1
	session.NextIdxToVerify = int64(session.NumChunks - 1)
	suite.DB.ValidateAndUpdate(&session)

	suite.Equal(int64(session.NumChunks-1), session.NextIdxToVerify)

	// call method under test, passing in our mock of our iota methods
	jobs.VerifyDataMaps(IotaMock, jobs.PrometheusWrapper, time.Now().Add(1*time.Minute))

	// verify the mock methods were called
	suite.True(verifyChunkMessagesMatchesRecordMockCalled_verify)

	session = models.UploadSession{}
	suite.DB.First(&session)

	suite.Equal(int64(7), session.NextIdxToVerify)

	suite.True(verifyChunkMessagesMatchesRecordMockCalled_verify)
}

func (suite *JobsSuite) Test_VerifyDataMaps_verify_all_alpha() {
	// make suite available inside mock methods
	Suite = *suite

	verifyChunkMessagesMatchesRecordMockCalled_verify = false

	// assign the mock methods for this test
	// return all chunks as being attached to the tangle, i.e. all verified
	IotaMock.VerifyChunkMessagesMatchRecord = func(chunks []oyster_utils.ChunkData) (filteredChunks services.FilteredChunk, err error) {

		// our mock was called
		verifyChunkMessagesMatchesRecordMockCalled_verify = true

		matchesTangle := []oyster_utils.ChunkData{}
		doesNotMatchTangle := []oyster_utils.ChunkData{}
		notAttached := []oyster_utils.ChunkData{}

		matchesTangle = append(matchesTangle, chunks...)

		return services.FilteredChunk{
			MatchesTangle:      matchesTangle,
			NotAttached:        notAttached,
			DoesNotMatchTangle: doesNotMatchTangle,
		}, err
	}

	models.MakeChannels(3)

	// populate data_maps
	numChunks := 10

	uploadSession1 := &models.UploadSession{
		GenesisHash:   oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes: uint64(numChunks * 1000),
		NumChunks:     numChunks,
		Type:          models.SessionTypeAlpha,
		ETHAddrAlpha:  nulls.String{string("SOME_ALPHA_ETH_ADDRESS1"), true},
		ETHAddrBeta:   nulls.String{string("SOME_BETA_ETH_ADDRESS1"), true},
		ETHPrivateKey: "000000000000005432",
	}

	SessionSetUpForTest(uploadSession1, []int{5}, uploadSession1.NumChunks)

	session1Keys := oyster_utils.GenerateBulkKeys(uploadSession1.GenesisHash, 0,
		int64(uploadSession1.NumChunks))
	chunksSession1InProgress, _ := models.GetMultiChunkData(oyster_utils.InProgressDir, uploadSession1.GenesisHash,
		session1Keys)
	chunksSession1Completed, _ := models.GetMultiChunkData(oyster_utils.CompletedDir, uploadSession1.GenesisHash,
		session1Keys)

	// check that it is the length we expect
	suite.Equal(numChunks+1, len(chunksSession1InProgress))
	suite.Equal(0, len(chunksSession1Completed))

	uploadSession1.PaymentStatus = models.PaymentStatusConfirmed
	uploadSession1.TreasureStatus = models.TreasureInDataMapComplete
	uploadSession1.AllDataReady = models.AllDataReady
	suite.DB.ValidateAndUpdate(uploadSession1)

	uploadSession1.NextIdxToAttach = int64(uploadSession1.NumChunks)
	uploadSession1.NextIdxToVerify = 0
	suite.DB.ValidateAndUpdate(uploadSession1)

	suite.Equal(int64(0), uploadSession1.NextIdxToVerify)

	// call method under test, passing in our mock of our iota methods
	jobs.VerifyDataMaps(IotaMock, jobs.PrometheusWrapper, time.Now().Add(1*time.Minute))

	// verify the mock methods were called
	suite.True(verifyChunkMessagesMatchesRecordMockCalled_verify)

	session := models.UploadSession{}
	suite.DB.First(&session)

	suite.Equal(int64(uploadSession1.NumChunks), session.NextIdxToVerify)
	suite.True(verifyChunkMessagesMatchesRecordMockCalled_verify)
}

func (suite *JobsSuite) Test_VerifyDataMaps_verify_some_alpha() {
	// make suite available inside mock methods
	Suite = *suite

	verifyChunkMessagesMatchesRecordMockCalled_verify = false

	// assign the mock methods for this test
	// return all chunks as being attached to the tangle, i.e. all verified
	IotaMock.VerifyChunkMessagesMatchRecord = func(chunks []oyster_utils.ChunkData) (filteredChunks services.FilteredChunk, err error) {

		// our mock was called
		verifyChunkMessagesMatchesRecordMockCalled_verify = true

		matchesTangle := []oyster_utils.ChunkData{}
		doesNotMatchTangle := []oyster_utils.ChunkData{}
		notAttached := []oyster_utils.ChunkData{}

		matchesTangle = append(matchesTangle, chunks[0], chunks[1], chunks[2])
		doesNotMatchTangle = append(doesNotMatchTangle, chunks[3])
		for i := 4; i < len(chunks); i++ {
			notAttached = append(notAttached, chunks[i])
		}

		return services.FilteredChunk{
			MatchesTangle:      matchesTangle,
			NotAttached:        notAttached,
			DoesNotMatchTangle: doesNotMatchTangle,
		}, err
	}

	models.MakeChannels(3)

	// populate data_maps
	numChunks := 10

	uploadSession1 := models.UploadSession{
		GenesisHash:   oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes: uint64(numChunks * 1000),
		NumChunks:     numChunks,
		Type:          models.SessionTypeAlpha,
		ETHAddrAlpha:  nulls.String{string("SOME_ALPHA_ETH_ADDRESS1"), true},
		ETHAddrBeta:   nulls.String{string("SOME_BETA_ETH_ADDRESS1"), true},
		ETHPrivateKey: "000000000000005432",
	}

	SessionSetUpForTest(&uploadSession1, []int{5}, uploadSession1.NumChunks)

	session1Keys := oyster_utils.GenerateBulkKeys(uploadSession1.GenesisHash, 0,
		int64(uploadSession1.NumChunks))
	chunksSession1InProgress, _ := models.GetMultiChunkData(oyster_utils.InProgressDir, uploadSession1.GenesisHash,
		session1Keys)
	chunksSession1Completed, _ := models.GetMultiChunkData(oyster_utils.CompletedDir, uploadSession1.GenesisHash,
		session1Keys)

	// check that it is the length we expect
	suite.Equal(numChunks+1, len(chunksSession1InProgress))
	suite.Equal(0, len(chunksSession1Completed))

	// set all data maps to need verifying
	session := models.UploadSession{}
	suite.DB.First(&session)
	session.PaymentStatus = models.PaymentStatusConfirmed
	session.TreasureStatus = models.TreasureInDataMapComplete
	session.AllDataReady = models.AllDataReady
	suite.DB.ValidateAndUpdate(&session)

	session.NextIdxToAttach = int64(session.NumChunks)
	session.NextIdxToVerify = 0
	suite.DB.ValidateAndUpdate(&session)

	suite.Equal(int64(0), session.NextIdxToVerify)

	// call method under test, passing in our mock of our iota methods
	jobs.VerifyDataMaps(IotaMock, jobs.PrometheusWrapper, time.Now().Add(1*time.Minute))

	// verify the mock methods were called
	suite.True(verifyChunkMessagesMatchesRecordMockCalled_verify)

	session = models.UploadSession{}
	suite.DB.First(&session)

	// we only set 3 to be matching the tangle
	suite.Equal(int64(3), session.NextIdxToVerify)
}

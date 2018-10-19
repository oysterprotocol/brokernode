package jobs_test

import (
	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"github.com/pkg/errors"
	"time"
)

func (suite *JobsSuite) Test_AttachTreasureTransactions_success() {
	IotaMock.DoPoW = func(chunks []oyster_utils.ChunkData) error {
		return nil
	}

	// create one treasure for each SignedStatus
	makeOneTreasureOfEachSignedStatus(suite)

	// call method under test
	jobs.AttachTreasureTransactions(IotaMock)

	/*
		Expectations:
			-there will be no treasures with SignedStatus TreasureAttachError
			-there will be no treasures with SignedStatus TreasureSigned
			-there will be 3 treasures with SignedStatus TreasureSignedAndAttached. 2 of them
			are newly assigned and one was created in makeOneTreasureOfEachSignedStatus()
	*/

	numTreasureAttachError := 0
	numTreasureSigned := 0
	numTreasureSignedAndAttached := 0

	treasures := []models.Treasure{}
	suite.DB.All(&treasures)

	for _, treasure := range treasures {
		if treasure.SignedStatus == models.TreasureAttachError {
			numTreasureAttachError++
		}
		if treasure.SignedStatus == models.TreasureSigned {
			numTreasureSigned++
		}
		if treasure.SignedStatus == models.TreasureSignedAndAttached {
			numTreasureSignedAndAttached++
		}
	}

	suite.Equal(0, numTreasureAttachError)
	suite.Equal(0, numTreasureSigned)
	suite.Equal(3, numTreasureSignedAndAttached)
}

func (suite *JobsSuite) Test_AttachTreasureTransactions_error() {
	IotaMock.DoPoW = func(chunks []oyster_utils.ChunkData) error {
		return errors.New("something went wrong")
	}

	// create one treasure for each SignedStatus
	makeOneTreasureOfEachSignedStatus(suite)

	// call method under test
	jobs.AttachTreasureTransactions(IotaMock)

	/*
		Expectations:
			-there will be 2 treasures with SignedStatus TreasureAttachError. 1 was created in
			makeOneTreasureOfEachSignedStatus() and 1 other will have this status after
			we call the method under test
			-there will be no treasures with SignedStatus TreasureSigned
			-there will be 1 treasures with SignedStatus TreasureSignedAndAttached.  Just the one that
			we initially created in makeOneTreasureOfEachSignedStatus()
	*/

	numTreasureAttachError := 0
	numTreasureSigned := 0
	numTreasureSignedAndAttached := 0

	treasures := []models.Treasure{}
	suite.DB.All(&treasures)

	for _, treasure := range treasures {
		if treasure.SignedStatus == models.TreasureAttachError {
			numTreasureAttachError++
		}
		if treasure.SignedStatus == models.TreasureSigned {
			numTreasureSigned++
		}
		if treasure.SignedStatus == models.TreasureSignedAndAttached {
			numTreasureSignedAndAttached++
		}
	}

	suite.Equal(2, numTreasureAttachError)
	suite.Equal(0, numTreasureSigned)
	suite.Equal(1, numTreasureSignedAndAttached)
}

func (suite *JobsSuite) Test_VerifyTreasureTransactions_error_while_verifying() {
	IotaMock.VerifyChunkMessagesMatchRecord = func(chunks []oyster_utils.ChunkData) (filteredChunks services.FilteredChunk, err error) {
		return services.FilteredChunk{
			MatchesTangle:      []oyster_utils.ChunkData{},
			NotAttached:        []oyster_utils.ChunkData{},
			DoesNotMatchTangle: []oyster_utils.ChunkData{},
		}, errors.New("some error happened")
	}

	// create one treasure for each SignedStatus
	makeOneTreasureOfEachSignedStatus(suite)

	// call method under test
	jobs.VerifyTreasureTransactions(IotaMock, time.Now())

	/*
		Expectations:
			-there was an error while verifying so nothing should have changed
	*/

	oneOfEachStatus := true

	for status := range models.SignedStatusMap {
		treasures := []models.Treasure{}
		suite.DB.Where("signed_status = ?", status).All(&treasures)
		if len(treasures) != 1 {
			oneOfEachStatus = false
		}
	}

	suite.True(oneOfEachStatus)
}

func (suite *JobsSuite) Test_VerifyTreasureTransactions_success() {
	IotaMock.VerifyChunkMessagesMatchRecord = func(chunks []oyster_utils.ChunkData) (filteredChunks services.FilteredChunk, err error) {
		return services.FilteredChunk{
			MatchesTangle:      chunks,
			NotAttached:        []oyster_utils.ChunkData{},
			DoesNotMatchTangle: []oyster_utils.ChunkData{},
		}, nil
	}

	// create one treasure for each SignedStatus
	makeOneTreasureOfEachSignedStatus(suite)

	// call method under test
	jobs.VerifyTreasureTransactions(IotaMock, time.Now())

	/*
		Expectations:
			-there will be 2 treasures with SignedStatus TreasureSignedAndAttachmentVerified. 1 was created in
			makeOneTreasureOfEachSignedStatus() and 1 other will have this status after
			we call the method under test
			-there will be no treasures with SignedStatus TreasureSignedAndAttached
	*/

	numTreasureSignedAndAttached := 0
	numTreasureSignedAndAttachmentVerified := 0

	treasures := []models.Treasure{}
	suite.DB.All(&treasures)

	for _, treasure := range treasures {
		if treasure.SignedStatus == models.TreasureSignedAndAttachmentVerified {
			numTreasureSignedAndAttachmentVerified++
		}
		if treasure.SignedStatus == models.TreasureSignedAndAttached {
			numTreasureSignedAndAttached++
		}
	}

	suite.Equal(2, numTreasureSignedAndAttachmentVerified)
	suite.Equal(0, numTreasureSignedAndAttached)
}

func (suite *JobsSuite) Test_VerifyTreasureTransactions_does_not_match_record() {
	IotaMock.VerifyChunkMessagesMatchRecord = func(chunks []oyster_utils.ChunkData) (filteredChunks services.FilteredChunk, err error) {
		return services.FilteredChunk{
			MatchesTangle:      []oyster_utils.ChunkData{},
			NotAttached:        []oyster_utils.ChunkData{},
			DoesNotMatchTangle: chunks,
		}, nil
	}

	// create one treasure for each SignedStatus
	makeOneTreasureOfEachSignedStatus(suite)

	// call method under test
	jobs.VerifyTreasureTransactions(IotaMock, time.Now())

	/*
		Expectations:
			-there will be 2 treasures with SignedStatus TreasureAttachError. 1 was created in
			makeOneTreasureOfEachSignedStatus() and 1 other will have this status after
			we call the method under test
			-there will be no treasures with SignedStatus TreasureSignedAndAttached
	*/

	numTreasureSignedAndAttached := 0
	numTreasureAttachError := 0

	treasures := []models.Treasure{}
	suite.DB.All(&treasures)

	for _, treasure := range treasures {
		if treasure.SignedStatus == models.TreasureAttachError {
			numTreasureAttachError++
		}
		if treasure.SignedStatus == models.TreasureSignedAndAttached {
			numTreasureSignedAndAttached++
		}
	}

	suite.Equal(2, numTreasureAttachError)
	suite.Equal(0, numTreasureSignedAndAttached)
}

func (suite *JobsSuite) Test_VerifyTreasureTransactions_not_attached_not_timed_out() {
	IotaMock.VerifyChunkMessagesMatchRecord = func(chunks []oyster_utils.ChunkData) (filteredChunks services.FilteredChunk, err error) {
		return services.FilteredChunk{
			MatchesTangle:      []oyster_utils.ChunkData{},
			NotAttached:        chunks,
			DoesNotMatchTangle: []oyster_utils.ChunkData{},
		}, nil
	}

	// create one treasure for each SignedStatus
	makeOneTreasureOfEachSignedStatus(suite)

	// call method under test
	jobs.VerifyTreasureTransactions(IotaMock, time.Now().Add(-1*time.Hour))

	/*
		Expectations:
			-nothing will have changed because although the chunk is not attached, it is also
			not yet timed out
	*/

	oneOfEachStatus := true

	for status := range models.SignedStatusMap {
		treasures := []models.Treasure{}
		suite.DB.Where("signed_status = ?", status).All(&treasures)
		if len(treasures) != 1 {
			oneOfEachStatus = false
		}
	}

	suite.True(oneOfEachStatus)
}

func (suite *JobsSuite) Test_VerifyTreasureTransactions_not_attached_timed_out() {
	IotaMock.VerifyChunkMessagesMatchRecord = func(chunks []oyster_utils.ChunkData) (filteredChunks services.FilteredChunk, err error) {
		return services.FilteredChunk{
			MatchesTangle:      []oyster_utils.ChunkData{},
			NotAttached:        chunks,
			DoesNotMatchTangle: []oyster_utils.ChunkData{},
		}, nil
	}

	// create one treasure for each SignedStatus
	makeOneTreasureOfEachSignedStatus(suite)

	// call method under test
	jobs.VerifyTreasureTransactions(IotaMock, time.Now().Add(1*time.Hour))

	/*
		Expectations:
			-the chunk is not attached and has timed out, so there will be 2 treasures with
			SignedStatus TreasureAttachError. 1 was created in makeOneTreasureOfEachSignedStatus()
			and 1 other will have this status after we call the method under test
			-there will be no treasures with SignedStatus TreasureSignedAndAttached
	*/

	numTreasureSignedAndAttached := 0
	numTreasureAttachError := 0

	treasures := []models.Treasure{}
	suite.DB.All(&treasures)

	for _, treasure := range treasures {
		if treasure.SignedStatus == models.TreasureAttachError {
			numTreasureAttachError++
		}
		if treasure.SignedStatus == models.TreasureSignedAndAttached {
			numTreasureSignedAndAttached++
		}
	}

	suite.Equal(2, numTreasureAttachError)
	suite.Equal(0, numTreasureSignedAndAttached)
}

func makeOneTreasureOfEachSignedStatus(suite *JobsSuite) {

	generateTreasuresToBury(suite, len(models.SignedStatusMap), models.PRLWaiting)

	treasures := []models.Treasure{}
	suite.DB.All(&treasures)
	suite.Equal(len(models.SignedStatusMap), len(treasures))

	var i = 0
	for status := range models.SignedStatusMap {
		treasures[i].SignedStatus = status
		vErr, err := suite.DB.ValidateAndSave(&treasures[i])
		suite.Nil(err)
		suite.False(vErr.HasAny())
		i++
	}
}

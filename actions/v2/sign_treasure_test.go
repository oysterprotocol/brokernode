package actions_v2

import (
	"encoding/json"
	"github.com/gobuffalo/uuid"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/utils"
	"io/ioutil"
	"strings"
)

func (suite *ActionSuite) Test_GetUnsignedTreasure_no_session_found() {
	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	defer oyster_utils.ResetBrokerMode()

	res := suite.JSON("/api/v2/unsigned-treasure/abcdef").Get()

	suite.Equal(500, res.Code)
	suite.True(strings.Contains(res.Body.String(), "no rows in result set"))
}

func (suite *ActionSuite) Test_GetUnsignedTreasure_in_test_mode_no_treasure() {
	oyster_utils.SetBrokerMode(oyster_utils.TestModeNoTreasure)
	defer oyster_utils.ResetBrokerMode()

	fileSizeBytes := uint64(123)
	numChunks := 2
	storageLengthInYears := 2

	u := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:        fileSizeBytes,
		NumChunks:            numChunks,
		StorageLengthInYears: storageLengthInYears,
	}

	u.StartUploadSession()

	res := suite.JSON("/api/v2/unsigned-treasure/" + u.ID.String()).Get()

	suite.Equal(200, res.Code)

	// Parse response
	resParsed := unsignedTreasureRes{}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	suite.Nil(err)
	err = json.Unmarshal(bodyBytes, &resParsed)
	suite.Nil(err)

	suite.Equal(false, resParsed.Available)
	suite.Equal([]TreasurePayload{}, resParsed.UnsignedTreasure)

	// Get session
	uploadSession := &models.UploadSession{}
	models.DB.Find(uploadSession, u.ID)
	suite.Equal(models.AllDataReady, uploadSession.AllDataReady)
}

func (suite *ActionSuite) Test_GetUnsignedTreasure_not_responsible() {
	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	defer oyster_utils.ResetBrokerMode()

	fileSizeBytes := uint64(123)
	numChunks := 2
	storageLengthInYears := 2

	u := models.UploadSession{
		Type:                         models.SessionTypeAlpha,
		GenesisHash:                  oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:                fileSizeBytes,
		NumChunks:                    numChunks,
		StorageLengthInYears:         storageLengthInYears,
		TreasureResponsibilityStatus: models.TreasureNotResponsible,
	}

	u.StartUploadSession()

	res := suite.JSON("/api/v2/unsigned-treasure/" + u.ID.String()).Get()

	suite.Equal(200, res.Code)

	// Parse response
	resParsed := unsignedTreasureRes{}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	suite.Nil(err)
	err = json.Unmarshal(bodyBytes, &resParsed)
	suite.Nil(err)

	suite.Equal(false, resParsed.Available)
	suite.Equal([]TreasurePayload{}, resParsed.UnsignedTreasure)

	// Get session
	uploadSession := &models.UploadSession{}
	models.DB.Find(uploadSession, u.ID)
	suite.Equal(models.AllDataReady, uploadSession.AllDataReady)
}

func (suite *ActionSuite) Test_GetUnsignedTreasure_responsible_no_treasures() {
	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	defer oyster_utils.ResetBrokerMode()

	fileSizeBytes := uint64(123)
	numChunks := 2
	storageLengthInYears := 2

	u := models.UploadSession{
		Type:                         models.SessionTypeAlpha,
		GenesisHash:                  oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:                fileSizeBytes,
		NumChunks:                    numChunks,
		StorageLengthInYears:         storageLengthInYears,
		TreasureResponsibilityStatus: models.TreasureResponsibleNotAttached,
	}

	u.StartUploadSession()

	res := suite.JSON("/api/v2/unsigned-treasure/" + u.ID.String()).Get()

	suite.Equal(500, res.Code)

	// Parse response
	resParsed := unsignedTreasureRes{}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	suite.Nil(err)
	err = json.Unmarshal(bodyBytes, &resParsed)
	suite.Nil(err)

	suite.Equal(false, resParsed.Available)
	suite.Equal([]TreasurePayload(nil), resParsed.UnsignedTreasure)
}

func (suite *ActionSuite) Test_GetUnsignedTreasure_responsible_with_treasures() {

	/*

		This test will fail until more rev2 changes are added.  Re-enable when ready.

		oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
		defer oyster_utils.ResetBrokerMode()

		fileSizeBytes := uint64(123)
		numChunks := 9
		storageLengthInYears := 2

		u := models.UploadSession{
			Type:                         models.SessionTypeAlpha,
			GenesisHash:                  oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
			FileSizeBytes:                fileSizeBytes,
			NumChunks:                    numChunks,
			StorageLengthInYears:         storageLengthInYears,
			TreasureResponsibilityStatus: models.TreasureResponsibleNotAttached,
		}

		mergedIndexes := []int{5}

		SessionSetUpForTest(&u, mergedIndexes, u.NumChunks)

		res := suite.JSON("/api/v2/unsigned-treasure/" + u.ID.String()).Get()

		suite.Equal(200, res.Code)

		// Parse response
		resParsed := unsignedTreasureRes{}
		bodyBytes, err := ioutil.ReadAll(res.Body)
		suite.Nil(err)
		err = json.Unmarshal(bodyBytes, &resParsed)
		suite.Nil(err)

		suite.Equal(true, resParsed.Available)

		treasures := []models.Treasure{}
		suite.DB.Where("genesis_hash = ?", u.GenesisHash).All(&treasures)

		suite.Equal(1, len(treasures))

		suite.Equal(treasures[0].Message, resParsed.UnsignedTreasure[0].TreasurePayload)
		suite.Equal(treasures[0].ID, resParsed.UnsignedTreasure[0].ID)
		suite.Equal(treasures[0].Idx, resParsed.UnsignedTreasure[0].Idx)

		// Get session
		uploadSession := &models.UploadSession{}
		models.DB.Find(uploadSession, u.ID)
		suite.Equal(models.AllDataNotReady, uploadSession.AllDataReady)

	*/
}

func (suite *ActionSuite) Test_SignTreasure_no_sessions() {
	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	defer oyster_utils.ResetBrokerMode()

	treasurePayloads := []TreasurePayload{{
		ID:              uuid.NewV3(uuid.UUID{}, "blahblah"),
		Idx:             int64(5),
		TreasurePayload: "treasurePayload",
	}}

	res := suite.JSON("/api/v2/signed-treasure/" + "abcdef").Put(map[string]interface{}{
		"signedTreasure": treasurePayloads,
	})

	suite.Equal(500, res.Code)
	suite.True(strings.Contains(res.Body.String(), "no rows in result set"))
}

func (suite *ActionSuite) Test_SignTreasure_test_mode_no_treasure() {
	oyster_utils.SetBrokerMode(oyster_utils.TestModeNoTreasure)
	defer oyster_utils.ResetBrokerMode()

	fileSizeBytes := uint64(123)
	numChunks := 9
	storageLengthInYears := 2

	u := models.UploadSession{
		Type:                         models.SessionTypeAlpha,
		GenesisHash:                  oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:                fileSizeBytes,
		NumChunks:                    numChunks,
		StorageLengthInYears:         storageLengthInYears,
		TreasureResponsibilityStatus: models.TreasureResponsibleNotAttached,
	}

	suite.DB.ValidateAndCreate(&u)

	treasurePayloads := []TreasurePayload{{
		ID:              uuid.NewV3(uuid.UUID{}, "blahblah"),
		Idx:             int64(5),
		TreasurePayload: "treasurePayload",
	}}

	res := suite.JSON("/api/v2/signed-treasure/" + u.ID.String()).Put(map[string]interface{}{
		"signedTreasure": treasurePayloads,
	})

	suite.Equal(200, res.Code)

	// Parse response
	resParsed := unsignedTreasureRes{}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	suite.Nil(err)
	err = json.Unmarshal(bodyBytes, &resParsed)
	suite.Nil(err)

	// Get session
	uploadSession := &models.UploadSession{}
	models.DB.Find(uploadSession, u.ID)
	suite.Equal(models.AllDataReady, uploadSession.AllDataReady)
}

func (suite *ActionSuite) Test_SignTreasure_treasure_not_responsible() {
	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	defer oyster_utils.ResetBrokerMode()

	fileSizeBytes := uint64(123)
	numChunks := 9
	storageLengthInYears := 2

	u := models.UploadSession{
		Type:                         models.SessionTypeAlpha,
		GenesisHash:                  oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:                fileSizeBytes,
		NumChunks:                    numChunks,
		StorageLengthInYears:         storageLengthInYears,
		TreasureResponsibilityStatus: models.TreasureNotResponsible,
		AllDataReady:                 models.AllDataNotReady,
	}

	suite.DB.ValidateAndCreate(&u)

	treasurePayloads := []TreasurePayload{{
		ID:              uuid.NewV3(uuid.UUID{}, "blahblah"),
		Idx:             int64(5),
		TreasurePayload: "treasurePayload",
	}}

	res := suite.JSON("/api/v2/signed-treasure/" + u.ID.String()).Put(map[string]interface{}{
		"signedTreasure": treasurePayloads,
	})

	suite.Equal(200, res.Code)

	// Parse response
	resParsed := unsignedTreasureRes{}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	suite.Nil(err)
	err = json.Unmarshal(bodyBytes, &resParsed)
	suite.Nil(err)

	// Get session
	uploadSession := &models.UploadSession{}
	models.DB.Find(uploadSession, u.ID)
	suite.Equal(models.AllDataReady, uploadSession.AllDataReady)
}

func (suite *ActionSuite) Test_SignTreasure_treasure_responsible_no_treasures() {
	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	defer oyster_utils.ResetBrokerMode()

	fileSizeBytes := uint64(123)
	numChunks := 9
	storageLengthInYears := 2

	u := models.UploadSession{
		Type:                         models.SessionTypeAlpha,
		GenesisHash:                  oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:                fileSizeBytes,
		NumChunks:                    numChunks,
		StorageLengthInYears:         storageLengthInYears,
		TreasureResponsibilityStatus: models.TreasureResponsibleNotAttached,
		AllDataReady:                 models.AllDataNotReady,
	}

	suite.DB.ValidateAndCreate(&u)

	treasurePayloads := []TreasurePayload{{
		ID:              uuid.NewV3(uuid.UUID{}, "blahblah"),
		Idx:             int64(5),
		TreasurePayload: "treasurePayload",
	}}

	res := suite.JSON("/api/v2/signed-treasure/" + u.ID.String()).Put(map[string]interface{}{
		"signedTreasure": treasurePayloads,
	})

	suite.Equal(500, res.Code)

	suite.True(strings.Contains(res.Body.String(), "no rows in result set"))
	suite.True(strings.Contains(res.Body.String(), "treasure does not exist or error finding treasure"))
}

func (suite *ActionSuite) Test_SignTreasure_treasure_responsible_with_treasures() {
	/*

		This test will fail until more rev2 changes are added.  Re-enable when ready.

	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	defer oyster_utils.ResetBrokerMode()

	fileSizeBytes := uint64(123)
	numChunks := 9
	storageLengthInYears := 2

	u := models.UploadSession{
		Type:                         models.SessionTypeAlpha,
		GenesisHash:                  oyster_utils.RandSeq(6, []rune("abcdef0123456789")),
		FileSizeBytes:                fileSizeBytes,
		NumChunks:                    numChunks,
		StorageLengthInYears:         storageLengthInYears,
		TreasureResponsibilityStatus: models.TreasureResponsibleNotAttached,
	}

	// set session up for test so that TreasureIdxMap gets created
	SessionSetUpForTest(&u, []int{5}, u.NumChunks)

	// call unsigned treasures endpoint so that an entry in the treasures
	// table is created
	res := suite.JSON("/api/v2/unsigned-treasure/" + u.ID.String()).Get()

	// Parse response
	resParsed := unsignedTreasureRes{}
	bodyBytes, _ := ioutil.ReadAll(res.Body)
	json.Unmarshal(bodyBytes, &resParsed)

	initialTreasurePayload := resParsed.UnsignedTreasure[0]

	treasures := []models.Treasure{}
	suite.DB.Where("genesis_hash = ?", u.GenesisHash).All(&treasures)

	// verify the ID, Idx, and Message all match the entry in our treasures table
	suite.Equal(initialTreasurePayload.TreasurePayload, treasures[0].Message)
	suite.Equal(initialTreasurePayload.ID, treasures[0].ID)
	suite.Equal(initialTreasurePayload.Idx, treasures[0].Idx)

	// verify that the SignedStatus is TreasureUnsigned
	suite.Equal(models.TreasureUnsigned, treasures[0].SignedStatus)

	// Get session
	uploadSession := &models.UploadSession{}
	models.DB.Find(uploadSession, u.ID)

	// verify that all the data is not yet ready
	suite.Equal(models.AllDataNotReady, uploadSession.AllDataReady)

	// create a treasure payload with a un updated ("signed") message
	treasurePayloads := []TreasurePayload{{
		ID:              initialTreasurePayload.ID,
		Idx:             initialTreasurePayload.Idx,
		TreasurePayload: "someDifferentPayload",
	}}

	// call signed treasures endpoint
	res = suite.JSON("/api/v2/signed-treasure/" + u.ID.String()).Put(map[string]interface{}{
		"signedTreasure": treasurePayloads,
	})

	suite.Equal(200, res.Code)

	treasures = []models.Treasure{}
	suite.DB.Where("genesis_hash = ?", u.GenesisHash).All(&treasures)

	// verify message is not the same as the initial message
	suite.NotEqual(initialTreasurePayload.TreasurePayload, treasures[0].Message)

	// verify ID and Idx are still the same
	suite.Equal(initialTreasurePayload.ID, treasures[0].ID)
	suite.Equal(initialTreasurePayload.Idx, treasures[0].Idx)

	// verify message matches the message we passed in
	suite.Equal(treasurePayloads[0].TreasurePayload, treasures[0].Message)

	// verify that the SignedStatus is TreasureSigned
	suite.Equal(models.TreasureSigned, treasures[0].SignedStatus)

	// Get session
	uploadSession = &models.UploadSession{}
	models.DB.Find(uploadSession, u.ID)

	// verify that all the data is ready
	suite.Equal(models.AllDataReady, uploadSession.AllDataReady)

	*/
}

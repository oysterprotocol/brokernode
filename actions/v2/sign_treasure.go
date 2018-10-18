package actions_v2

import (
	"errors"
	"fmt"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/uuid"
	"github.com/oysterprotocol/brokernode/actions/utils"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/utils"
	"gopkg.in/segmentio/analytics-go.v3"
	"os"
	"time"
)

type SignTreasureResource struct {
	buffalo.Resource
}

type TreasurePayload struct {
	ID              uuid.UUID `json:"id"`
	Idx             int64     `json:"idx"`
	TreasurePayload string    `json:"treasurePayload"`
}

// Request Response structs
type unsignedTreasureRes struct {
	UnsignedTreasure []TreasurePayload `json:"unsignedTreasure"`
	Available        bool              `json:"available"`
}

type SignedTreasureReq struct {
	SignedTreasure []TreasurePayload `json:"signedTreasure"`
}

/*GetUnsignedTreasure signifies that all chunks have been uploaded and
gets the unsigned treasure so the client can sign it*/
func (usr *SignTreasureResource) GetUnsignedTreasure(c buffalo.Context) error {

	if os.Getenv("DEPLOY_IN_PROGRESS") == "true" {
		err := errors.New("Deployment in progress.  Try again later")
		c.Error(400, err)
		return err
	}

	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramSignTreasureGetUnsigned, start)

	// Get session
	uploadSession := &models.UploadSession{}
	err := models.DB.Find(uploadSession, c.Param("id"))

	if err != nil {
		oyster_utils.LogIfError(err, nil)
		c.Error(400, err)
		return err
	}

	if oyster_utils.BrokerMode == oyster_utils.TestModeNoTreasure ||
		uploadSession.TreasureResponsibilityStatus == models.TreasureNotResponsible {

		if err == nil {
			uploadSession.AllDataReady = models.AllDataReady
			models.DB.ValidateAndUpdate(uploadSession)
		}

		res := unsignedTreasureRes{
			Available:        false,
			UnsignedTreasure: []TreasurePayload{},
		}
		return c.Render(200, actions_utils.Render.JSON(res))
	}

	defer oyster_utils.TimeTrack(time.Now(), "actions/sign_treasure: getting_unsigned_treasure",
		analytics.NewProperties().
			Set("id", uploadSession.ID).
			Set("genesis_hash", uploadSession.GenesisHash))

	uploadSession.CreateTreasures()

	treasures := []models.Treasure{}
	err = models.DB.Where("genesis_hash = ?", uploadSession.GenesisHash).All(&treasures)
	if err != nil {
		oyster_utils.LogIfError(err, nil)
		c.Error(400, err)
		return err
	}

	if len(treasures) == 0 && uploadSession.TreasureResponsibilityStatus != models.TreasureNotResponsible {

		err = errors.New("broker was supposed to attach the treasure but no treasures were found")
		oyster_utils.LogIfError(err, nil)
		c.Error(400, err)
		return err
	}

	treasurePayloads := []TreasurePayload{}
	for _, treasure := range treasures {
		treasurePayloads = append(treasurePayloads, TreasurePayload{
			ID:              treasure.ID,
			Idx:             treasure.Idx,
			TreasurePayload: treasure.Message,
		})
	}

	res := unsignedTreasureRes{
		Available:        true,
		UnsignedTreasure: treasurePayloads,
	}

	return c.Render(200, actions_utils.Render.JSON(res))
}

/*SignTreasure stores the signed treasure*/
func (usr *SignTreasureResource) SignTreasure(c buffalo.Context) error {

	if os.Getenv("DEPLOY_IN_PROGRESS") == "true" {
		err := errors.New("Deployment in progress.  Try again later")
		c.Error(400, err)
		return err
	}

	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramSignTreasureSetSigned, start)

	// Get session
	uploadSession := &models.UploadSession{}
	err := models.DB.Find(uploadSession, c.Param("id"))

	if err != nil {
		oyster_utils.LogIfError(err, nil)
		c.Error(400, err)
		return err
	}

	if oyster_utils.BrokerMode == oyster_utils.TestModeNoTreasure ||
		uploadSession.TreasureResponsibilityStatus == models.TreasureNotResponsible {

		if err == nil {
			uploadSession.AllDataReady = models.AllDataReady
			models.DB.ValidateAndUpdate(uploadSession)
		}

		return c.Render(200, actions_utils.Render.JSON(map[string]bool{"success": true}))
	}

	defer oyster_utils.TimeTrack(time.Now(), "actions/sign_treasure: signing_treasure", analytics.NewProperties().
		Set("id", uploadSession.ID).
		Set("genesis_hash", uploadSession.GenesisHash))

	req := SignedTreasureReq{}
	if err := oyster_utils.ParseReqBody(c.Request(), &req); err != nil {
		err = fmt.Errorf("Invalid request, unable to parse request body  %v", err)
		c.Error(400, err)
		return err
	}

	for _, signedTreasure := range req.SignedTreasure {
		// Get treasure
		treasure := &models.Treasure{}
		err := models.DB.Find(treasure, signedTreasure.ID)
		oyster_utils.LogIfError(err, nil)
		if err == nil {
			treasure.Message = signedTreasure.TreasurePayload
			treasure.SignedStatus = models.TreasureSigned
			vErr, err := models.DB.ValidateAndUpdate(treasure)
			oyster_utils.LogIfError(err, nil)
			oyster_utils.LogIfValidationError("error updating with signed treasure", vErr, nil)
		} else {
			newErr := errors.New(err.Error() + " - treasure does not exist or error finding treasure")
			oyster_utils.LogIfError(newErr, nil)
			c.Error(400, newErr)
			return newErr
		}
	}

	uploadSession.AllDataReady = models.AllDataReady
	vErr, err := models.DB.ValidateAndUpdate(uploadSession)

	oyster_utils.LogIfError(err, nil)
	oyster_utils.LogIfValidationError("error updating with signed treasure", vErr, nil)

	return c.Render(200, actions_utils.Render.JSON(map[string]bool{"success": true}))
}

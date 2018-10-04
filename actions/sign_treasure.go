package actions

import (
	"errors"
	"fmt"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/uuid"
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
		fmt.Println(err)
		c.Error(400, err)
		return err
	}

	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramSignTreasureGetUnsigned, start)

	// Get session
	uploadSession := &models.UploadSession{}
	err := models.DB.Find(uploadSession, c.Param("id"))

	if oyster_utils.BrokerMode == oyster_utils.TestModeNoTreasure {

		uploadSession.AllDataReady = models.AllDataReady
		models.DB.ValidateAndUpdate(uploadSession)

		res := unsignedTreasureRes{
			Available:        false,
			UnsignedTreasure: []TreasurePayload{},
		}
		return c.Render(200, r.JSON(res))
	}

	defer oyster_utils.TimeTrack(time.Now(), "actions/sign_treasure: getting_unsigned_treasure", analytics.NewProperties().
		Set("id", uploadSession.ID).
		Set("genesis_hash", uploadSession.GenesisHash))

	if err != nil {
		oyster_utils.LogIfError(err, nil)
		c.Error(400, err)
		return err
	}
	if uploadSession == nil {
		err := errors.New("error finding session in GetUnsignedTreasure")
		oyster_utils.LogIfError(err, nil)
		c.Error(400, err)
		return err
	}

	if uploadSession.TreasureResponsibilityStatus == models.TreasureNotResponsible {
		uploadSession.AllDataReady = models.AllDataReady
		models.DB.ValidateAndUpdate(uploadSession)

		res := unsignedTreasureRes{
			Available:        false,
			UnsignedTreasure: []TreasurePayload{},
		}
		return c.Render(200, r.JSON(res))
	}

	CreateTreasures(uploadSession)

	//treasureMap, err := uploadSession.GetTreasureMap()
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
			ID:  treasure.ID,
			Idx: treasure.Idx,
			// TODO:  send the RawMessage or the Message?  Message already in trytes.
			TreasurePayload: treasure.RawMessage,
		})
	}

	res := unsignedTreasureRes{
		Available:        true,
		UnsignedTreasure: treasurePayloads,
	}

	return c.Render(200, r.JSON(res))
}

/*SignTreasure stores the signed treasure*/
func (usr *SignTreasureResource) SignTreasure(c buffalo.Context) error {

	if os.Getenv("DEPLOY_IN_PROGRESS") == "true" {
		err := errors.New("Deployment in progress.  Try again later")
		fmt.Println(err)
		c.Error(400, err)
		return err
	}

	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramSignTreasureSetSigned, start)

	// Get session
	uploadSession := &models.UploadSession{}
	err := models.DB.Find(uploadSession, c.Param("id"))

	if oyster_utils.BrokerMode == oyster_utils.TestModeNoTreasure {

		uploadSession.AllDataReady = models.AllDataReady
		models.DB.ValidateAndUpdate(uploadSession)

		return c.Render(200, r.JSON(map[string]bool{"success": true}))
	}

	defer oyster_utils.TimeTrack(time.Now(), "actions/sign_treasure: signing_treasure", analytics.NewProperties().
		Set("id", uploadSession.ID).
		Set("genesis_hash", uploadSession.GenesisHash))

	if err != nil {
		oyster_utils.LogIfError(err, nil)
		c.Error(400, err)
		return err
	}
	if uploadSession == nil {
		err := errors.New("error finding session in SignTreasure")
		oyster_utils.LogIfError(err, nil)
		c.Error(400, err)
		return err
	}
	if uploadSession.TreasureResponsibilityStatus == models.TreasureNotResponsible {
		uploadSession.AllDataReady = models.AllDataReady
		models.DB.ValidateAndUpdate(uploadSession)

		return c.Render(200, r.JSON(map[string]bool{"success": true}))
	}

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
		if err == nil && treasure.ID != uuid.Nil {
			// TODO find out if Jet is going to send the raw payload or if it will already be
			// trytes
			treasure.RawMessage = signedTreasure.TreasurePayload
			treasure.Message = signedTreasure.TreasurePayload
			treasure.SignedStatus = models.TreasureSigned
			vErr, err := models.DB.ValidateAndUpdate(treasure)
			oyster_utils.LogIfError(err, nil)
			oyster_utils.LogIfValidationError("error updating with signed treasure", vErr, nil)
		} else {
			err := errors.New("treasure does not exist or error finding treasure")
			oyster_utils.LogIfError(err, nil)
			c.Error(400, err)
			return err
		}
	}

	uploadSession.AllDataReady = models.AllDataReady
	vErr, err := models.DB.ValidateAndUpdate(uploadSession)

	oyster_utils.LogIfError(err, nil)
	oyster_utils.LogIfValidationError("error updating with signed treasure", vErr, nil)

	return c.Render(200, r.JSON(map[string]bool{"success": true}))
}

/*CreateTreasures creates the entries in the treasures table*/
func CreateTreasures(session *models.UploadSession) {

	treasureIdxMapArray, err := session.GetTreasureMap()
	if err != nil {
		fmt.Println("Cannot create treasures to bury in sign_treasure: " + err.Error())
		// already captured error in upstream function
		return
	}
	if len(treasureIdxMapArray) == 0 {
		fmt.Println("Cannot create treasures to bury in sign_treasure: " + "treasureIdxMapArray is empty")
		return
	}

	treasureIdxMap := make(map[int64]models.TreasureMap)
	for _, treasureIdxEntry := range treasureIdxMapArray {
		treasureIdxMap[int64(treasureIdxEntry.Idx)] = treasureIdxEntry
	}

	prlPerTreasure, err := session.GetPRLsPerTreasure()
	if err != nil {
		fmt.Println("Cannot create treasures to bury in sign_treasure: " + err.Error())
		// captured error in upstream method
		return
	}

	prlInWei := oyster_utils.ConvertToWeiUnit(prlPerTreasure)

	for idx, treasureChunk := range treasureIdxMap {

		chunkDataEncryptionChunk := models.GetSingleChunkData(oyster_utils.InProgressDir, session.GenesisHash,
			int64(treasureChunk.EncryptionIdx))

		treasureAddress := models.GetTreasureAddress(oyster_utils.InProgressDir, session.GenesisHash,
			idx)

		decryptedKey, err := session.DecryptTreasureChunkEthKey(treasureChunk.Key)

		treasurePayloadRaw, err := models.CreateTreasurePayloadRaw(decryptedKey, chunkDataEncryptionChunk.RawMessage,
			models.MaxSideChainLength)

		// TODO: should use BytesToTrytes or AsciiToTrytes?
		treasurePayloadTryted := string(oyster_utils.BytesToTrytes([]byte(treasurePayloadRaw)))

		if err != nil {
			fmt.Println("Cannot create treasures to bury in sign_treasure: " + err.Error())
			// already captured error in upstream function
			continue
		}

		if decryptedKey == os.Getenv("TEST_MODE_WALLET_KEY") {
			continue
		}

		if oyster_utils.BrokerMode == oyster_utils.ProdMode {
			ethAddress := EthWrapper.GenerateEthAddrFromPrivateKey(decryptedKey)

			treasureToBury := models.Treasure{
				GenesisHash:     session.GenesisHash,
				ETHAddr:         ethAddress.Hex(),
				ETHKey:          decryptedKey,
				Address:         treasureAddress,
				Message:         treasurePayloadTryted,
				RawMessage:      treasurePayloadRaw,
				SignedStatus:    models.TreasureUnsigned,
				EncryptionIndex: int64(treasureChunk.EncryptionIdx),
				Idx:             int64(treasureChunk.Idx),
			}

			treasureToBury.SetPRLAmount(prlInWei)

			models.DB.ValidateAndCreate(&treasureToBury)
		}
	}
}

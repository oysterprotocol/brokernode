package actions

import (
	"bytes"
	"encoding/json"
	"net/http"

	raven "github.com/getsentry/raven-go"
	"github.com/gobuffalo/buffalo"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/pkg/errors"
)

type UploadSessionResource struct {
	buffalo.Resource
}

// Request Response structs

type uploadSessionCreateReq struct {
	GenesisHash          string `json:"genesisHash"`
	FileSizeBytes        int    `json:"fileSizeBytes"`
	BetaIP               string `json:"betaIp"`
	StorageLengthInYears int    `json:"storageLengthInYears"`
}

type uploadSessionCreateRes struct {
	ID            string               `json:"id"`
	UploadSession models.UploadSession `json:"uploadSession"`
	BetaSessionID string               `json:"betaSessionId"`
	Invoice       invoice              `json:"invoice"`
}

type invoice struct {
	Cost       int    `json:"cost"`
	EthAddress string `json:"ethAddress"`
}

type chunkReq struct {
	Idx  int    `json:"idx"`
	Data string `json:"data"`
	Hash string `json:"hash"`
}

type UploadSessionUpdateReq struct {
	Chunks []chunkReq `json:"chunks"`
}

// Create creates an upload session.
func (usr *UploadSessionResource) Create(c buffalo.Context) error {
	req := uploadSessionCreateReq{}
	parseReqBody(c.Request(), &req)

	// TODO: Handle PRL Payments

	// Start Beta Session.

	var betaSessionID = ""
	if req.BetaIP != "" {
		betaReq, err := json.Marshal(req)
		if err != nil {
			c.Render(400, r.JSON(map[string]string{"Error starting Beta": err.Error()}))
			return err
		}

		reqBetaBody := bytes.NewBuffer(betaReq)

		// Should we be hardcoding the port?
		betaURL := req.BetaIP + ":3000/api/v2/upload-sessions/beta"
		betaRes, err := http.Post(betaURL, "application/json", reqBetaBody)

		if err != nil {
			c.Render(400, r.JSON(map[string]string{"Error starting Beta": err.Error()}))
			return err
		}
		betaSessionRes := &uploadSessionCreateRes{}
		parseResBody(betaRes, betaSessionRes)
		betaSessionID = betaSessionRes.ID
	}

	// Start Alpha Session.

	u := models.UploadSession{
		Type:          models.SessionTypeAlpha,
		GenesisHash:   req.GenesisHash,
		FileSizeBytes: req.FileSizeBytes,
		TotalCost:     1.23, // TODO: Real price
	}
	vErr, err := u.StartUploadSession()
	if err != nil {
		return err
	}

	if len(vErr.Errors) > 0 {
		c.Render(422, r.JSON(vErr.Errors))
		return err
	}

	res := uploadSessionCreateRes{
		UploadSession: u,
		ID:            u.ID.String(),
		BetaSessionID: betaSessionID,
		Invoice:       CreateInvoice(req.StorageLengthInYears, req.FileSizeBytes),
	}
	return c.Render(200, r.JSON(res))
}

// Update uploads a chunk associated with an upload session.
func (usr *UploadSessionResource) Update(c buffalo.Context) error {

	req := UploadSessionUpdateReq{}
	parseReqBody(c.Request(), &req)

	// Get session
	uploadSession := &models.UploadSession{}
	err := models.DB.Find(uploadSession, c.Param("id"))
	if err != nil || uploadSession == nil {
		c.Render(400, r.JSON(map[string]string{"Error finding session": errors.WithStack(err).Error()}))
		return err
	}

	// Update dMaps to have chunks async
	go func() {
		// Map over chunks from request
		// TODO: Batch processing DB upserts.
		dMaps := make([]models.DataMap, len(req.Chunks))
		for i, chunk := range req.Chunks {
			// Fetch DataMap
			dm := models.DataMap{}
			err := models.DB.RawQuery(
				"SELECT * from data_maps WHERE genesis_hash = ? AND chunk_idx = ?", uploadSession.GenesisHash, chunk.Idx).First(&dm)

			if err != nil {
				raven.CaptureError(err, nil)
			}

			// Simple check if hashes match.
			if chunk.Hash == dm.GenesisHash {
				// Updates dmap in DB.
				dm.Message = chunk.Data
				dm.Status = models.Unassigned
				models.DB.ValidateAndSave(&dm)
			}

			dMaps[i] = dm
		}
	}()

	return c.Render(202, r.JSON(map[string]bool{"success": true}))
}

// CreateBeta creates an upload session on the beta broker.
func (usr *UploadSessionResource) CreateBeta(c buffalo.Context) error {
	req := uploadSessionCreateReq{}
	parseReqBody(c.Request(), &req)

	u := models.UploadSession{
		Type:          models.SessionTypeBeta,
		GenesisHash:   req.GenesisHash,
		FileSizeBytes: req.FileSizeBytes,
	}
	vErr, err := u.StartUploadSession()
	if err != nil {
		return err
	}

	if len(vErr.Errors) > 0 {
		c.Render(422, r.JSON(vErr.Errors))
		return err
	}

	res := uploadSessionCreateRes{
		UploadSession: u,
		ID:            u.ID.String(),
		Invoice:       CreateInvoice(req.StorageLengthInYears, req.FileSizeBytes),
	}
	return c.Render(200, r.JSON(res))
}

func CreateInvoice(storageLengthInYears int, fileSizeBytes int) invoice {
	invoice := invoice{
		EthAddress: models.GetEthAddress(),
		Cost:       models.CalculatePayment(storageLengthInYears, fileSizeBytes),
	}

	return invoice
}

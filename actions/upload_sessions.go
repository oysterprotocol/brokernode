package actions

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/pkg/errors"
)

type UploadSessionResource struct {
	buffalo.Resource
}

// Request Response structs

type uploadSessionCreateReq struct {
	GenesisHash   string `json:"genesisHash"`
	FileSizeBytes int    `json:"fileSizeBytes"`
	BetaIP        string `json:"betaIp"`
}

type uploadSessionCreateRes struct {
	UploadSession models.UploadSession `json:"id"`
	BetaSessionID string               `json:"betaSessionId"`
}

type chunkReq struct {
	idx  int    `json:idx`
	data string `json:data`
	hash string `json:hash`
}

type uploadSessionUpdateReq struct {
	chunks []chunkReq `json:chunks`
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
		betaURL := "https://" + req.BetaIP + ":3000/upload-sessions/beta"
		betaRes, err := http.Post(betaURL, "application/json", reqBetaBody)

		if err != nil {
			c.Render(400, r.JSON(map[string]string{"Error starting Beta": err.Error()}))
			return err
		}

		betaSessionRes := uploadSessionCreateRes{}
		parseResBody(betaRes, betaSessionRes)
		betaSessionID = betaSessionRes.UploadSession.ID.String()
	}

	// Start Alpha Session.

	u := models.UploadSession{
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
		BetaSessionID: betaSessionID,
	}
	return c.Render(200, r.JSON(res))
}

// Update uploads a chunk associated with an upload session.
func (usr *UploadSessionResource) Update(c buffalo.Context) error {
	req := uploadSessionUpdateReq{}
	parseReqBody(c.Request(), &req)

	// Get session
	tx := c.Value("tx").(*pop.Connection)
	uploadSession := &models.UploadSession{}
	err := tx.Find(uploadSession, c.Param("id"))
	if err != nil || uploadSession == nil {
		c.Render(400, r.JSON(map[string]string{"Error finding session": errors.WithStack(err).Error()}))
		return err
	}

	// Update dMaps to have chunks async
	go func() {
		// Map over chunks from request
		// TODO: Batch processing DB upserts.
		dMaps := make([]models.DataMap, len(req.chunks))
		for i, chunk := range req.chunks {
			// Fetch DataMap
			dm := models.DataMap{}
			tx.RawQuery(
				"SELECT * from data_maps WHERE genesis_hash ? AND chunk_idx = ?", uploadSession.GenesisHash, chunk.idx).First(&dm)

			// Simple check if hashes match.
			if chunk.hash == dm.Hash {
				// Updates dmap in DB.
				dm.Message = chunk.data
				dm.Status = models.Unassigned
				tx.ValidateAndSave(&dm)
			}

			dMaps[i] = dm
		}

		// Should we still do this here?
		iotaService := services.IotaService{}
		go iotaService.ProcessChunks(dMaps, false)
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

	res := uploadSessionCreateRes{UploadSession: u}
	return c.Render(200, r.JSON(res))
}

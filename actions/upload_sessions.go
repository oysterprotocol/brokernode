package actions

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/oysterprotocol/brokernode/models"
)

type UploadSessionResource struct {
	buffalo.Resource
}

// Request Response structs

type uploadSessionCreateReq struct {
	GenesisHash   string `json:"genesisHash"`
	FileSizeBytes int    `json:"fileSizeBytes"`
	BetaIP        string `json:"betaIP"`
}

type uploadSessionCreateRes struct {
	UploadSession models.UploadSession `json:"uploadSession"`
	BetaSessionID string               `json:"betaSessionID"`
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

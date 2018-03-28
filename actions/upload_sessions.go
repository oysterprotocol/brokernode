package actions

import (
	"github.com/oysterprotocol/brokernode/models"

	"github.com/gobuffalo/buffalo"
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
	// TODO: Start session with beta.

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

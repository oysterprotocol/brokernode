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
	models.UploadSession
	// TODO: Add beta session id.
}

// Create creates an upload session.
func (usr *UploadSessionResource) Create(c buffalo.Context) error {
	req := uploadSessionCreateReq{}
	ParseReqBody(c.Request(), &req)

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

	return c.Render(200, r.JSON(u))
}

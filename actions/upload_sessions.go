package actions

import (
	"fmt"

	"github.com/gobuffalo/buffalo"
)

type UploadSessionResource struct {
	buffalo.Resource
}

// Request parsing

type uploadSessionCreateReq struct {
	GenesisHash   string `json:"genesisHash"`
	FileSizeBytes int    `json:"fileSizeBytes"`
	BetaIP        string `json:"betaIP"`
}

// Create creates an upload session.
func (usr *UploadSessionResource) Create(c buffalo.Context) error {
	req := uploadSessionCreateReq{}
	ParseReqBody(c.Request(), &req)

	fmt.Println(req.GenesisHash)
	return c.Render(200, r.JSON(map[string]string{"message": req.GenesisHash}))
}

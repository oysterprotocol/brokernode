package actions_v3

import (
	"github.com/gobuffalo/buffalo"
	"github.com/oysterprotocol/brokernode/actions/utils"
)

type UploadSessionResource struct {
	buffalo.Resource
}

type uploadSessionCreateRes struct {
}

func (usr *UploadSessionResource) Create(c buffalo.Context) error {
	return c.Render(200, actions_utils.Render.JSON(uploadSessionCreateRes{}))
}

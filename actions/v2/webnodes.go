package actions_v2

import (
	"fmt"
	"os"

	"github.com/gobuffalo/buffalo"
	"github.com/oysterprotocol/brokernode/actions/utils"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/utils"
)

type WebnodeResource struct {
	buffalo.Resource
}

// Request Response structs

type webnodeCreateReq struct {
	Address string `json:"address"`
}

type webnodeCreateRes struct {
	Webnode models.Webnode `json:"id"`
}

// Creates a webnode.
func (usr *WebnodeResource) Create(c buffalo.Context) error {

	if os.Getenv("TANGLE_MAINTENANCE") == "true" {
		return c.Render(403, actions_utils.Render.JSON(map[string]string{"error": "This broker is undergoing tangle maintenance"}))
	}

	if os.Getenv("DEPLOY_IN_PROGRESS") == "true" {
		return c.Render(403, actions_utils.Render.JSON(map[string]string{"error": "Deployment in progress.  Try again later"}))
	}

	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramWebnodeResourceCreate, start)

	req := webnodeCreateReq{}
	if err := oyster_utils.ParseReqBody(c.Request(), &req); err != nil {
		err = fmt.Errorf("Invalid request, unable to parse request body  %v", err)
		c.Error(400, err)
		return err
	}

	w := models.Webnode{
		Address: req.Address,
	}

	res := webnodeCreateRes{
		Webnode: w,
	}

	return c.Render(200, actions_utils.Render.JSON(res))
}

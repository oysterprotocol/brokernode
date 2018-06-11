package actions

import (
	"github.com/gobuffalo/buffalo"
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
	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramWebnodeResourceCreate, start)

	req := webnodeCreateReq{}
	oyster_utils.ParseReqBody(c.Request(), &req)

	w := models.Webnode{
		Address: req.Address,
	}

	res := webnodeCreateRes{
		Webnode: w,
	}

	return c.Render(200, r.JSON(res))
}

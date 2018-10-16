package actions_v3

import (
	"github.com/gobuffalo/buffalo"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
)

// Visible for Unit Test
var IotaWrapper = services.IotaWrapper
var EthWrapper = oyster_utils.EthWrapper
var PrometheusWrapper = services.PrometheusWrapper

func RegisterApi(app *buffalo.App) *buffalo.App {
	apiV3 := app.Group("/api/v3")

	uploadSessionResourceV3 := UploadSessionResourceV3{}
	apiV3.PUT("upload-sessions/{id}", uploadSessionResourceV3.Update)
	apiV3.POST("upload-sessions", uploadSessionResourceV3.Create)
	apiV3.POST("upload-sessions/beta", uploadSessionResourceV3.CreateBeta)

	return apiV3
}

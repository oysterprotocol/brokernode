package actions_v3

import (
	"github.com/gobuffalo/buffalo"
	"github.com/oysterprotocol/brokernode/services"
)

// Visible for Unit Test
var IotaWrapper = services.IotaWrapper
var EthWrapper = services.EthWrapper
var PrometheusWrapper = services.PrometheusWrapper

func RegisterApi(app *buffalo.App) *buffalo.App {
	apiV3 := app.Group("/api/v3")

	uploadSessionResourceV3 := UploadSessionResourceV3{}
	apiV3.PUT("upload-sessions/{id}", uploadSessionResourceV3.Update)

	return apiV3
}

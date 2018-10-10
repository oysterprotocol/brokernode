package actions_v2

import (
	"github.com/gobuffalo/buffalo"
	"github.com/oysterprotocol/brokernode/services"
)

// Visible for Unit Test
var IotaWrapper = services.IotaWrapper
var EthWrapper = services.EthWrapper
var PrometheusWrapper = services.PrometheusWrapper

func RegisterApi(app *buffalo.App) *buffalo.App {
	apiV2 := app.Group("/api/v2")

	// UploadSessions
	uploadSessionResourceV2 := UploadSessionResourceV2{}
	apiV2.POST("upload-sessions", uploadSessionResourceV2.Create)
	apiV2.PUT("upload-sessions/{id}", uploadSessionResourceV2.Update)
	apiV2.POST("upload-sessions/beta", uploadSessionResourceV2.CreateBeta)
	apiV2.GET("upload-sessions/{id}", uploadSessionResourceV2.GetPaymentStatus)

	// Webnodes
	webnodeResource := WebnodeResource{}
	apiV2.POST("supply/webnodes", webnodeResource.Create)

	// Transactions
	transactionBrokernodeResource := TransactionBrokernodeResource{}
	apiV2.POST("demand/transactions/brokernodes", transactionBrokernodeResource.Create)
	apiV2.PUT("demand/transactions/brokernodes/{id}", transactionBrokernodeResource.Update)

	transactionGenesisHashResource := TransactionGenesisHashResource{}
	apiV2.POST("demand/transactions/genesis_hashes", transactionGenesisHashResource.Create)
	apiV2.PUT("demand/transactions/genesis_hashes/{id}", transactionGenesisHashResource.Update)

	// Treasures
	treasures := TreasuresResource{}
	apiV2.POST("treasures", treasures.VerifyAndClaim)

	return apiV2
}

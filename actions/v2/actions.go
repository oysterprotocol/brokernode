package actions_v2

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

	// Treasure signing
	signTreasureResource := SignTreasureResource{}
	apiV2.GET("unsigned-treasure/{id}", signTreasureResource.GetUnsignedTreasure)
	apiV2.PUT("signed-treasure/{id}", signTreasureResource.SignTreasure)

	// Status
	statusResource := StatusResource{}
	apiV2.GET("status", statusResource.CheckStatus)

	return apiV2
}

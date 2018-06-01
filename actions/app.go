package actions

import (
	"os"

	raven "github.com/getsentry/raven-go"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/middleware"
	"github.com/gobuffalo/buffalo/middleware/ssl"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/x/sessions"
	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/rs/cors"
	"github.com/unrolled/secure"
	"github.com/prometheus/client_golang/prometheus"
)

// ENV is used to help switch settings based on where the
// application is being run. Default is "development".
var ENV = envy.Get("GO_ENV", "development")
var app *buffalo.App

// Visible for Unit Test
var IotaWrapper = services.IotaWrapper
var EthWrapper = services.EthWrapper
var PrometheusWrapper = services.PrometheusWrapper

var HistogramTreasuresVerifyAndClaim = PrometheusWrapper.PrepareHistogram("treasures_verify_and_claim_seconds", "HistogramTreasuresVerifyAndClaimSeconds", "code")
var HistogramUploadSessionResourceCreate = PrometheusWrapper.PrepareHistogram("upload_session_resource_create_seconds", "HistogramUploadSessionResourceCreateSeconds", "code")
var HistogramUploadSessionResourceUpdate = PrometheusWrapper.PrepareHistogram("upload_session_resource_update_seconds", "HistogramUploadSessionResourceUpdateSeconds", "code")
var HistogramUploadSessionResourceCreateBeta = PrometheusWrapper.PrepareHistogram("upload_session_resource_create_beta_seconds", "HistogramUploadSessionResourceCreateBetaSeconds", "code")
var HistogramUploadSessionResourceGetPaymentStatus = PrometheusWrapper.PrepareHistogram("upload_session_resource_get_payment_status_seconds", "HistogramUploadSessionResourceGetPaymentStatusSeconds", "code")
var HistogramWebnodeResourceCreate = PrometheusWrapper.PrepareHistogram("webnode_resource_create_seconds", "HistogramWebnodeResourceCreateSeconds", "code")
var HistogramTransactionBrokernodeResourceCreate = PrometheusWrapper.PrepareHistogram("transaction_brokernode_resource_create_seconds", "HistogramTransactionBrokernodeResourceCreateSeconds", "code")
var HistogramTransactionBrokernodeResourceUpdate = PrometheusWrapper.PrepareHistogram("transaction_brokernode_resource_update_seconds", "HistogramTransactionBrokernodeResourceUpdateSeconds", "code")
var HistogramTransactionGenesisHashResourceCreate = PrometheusWrapper.PrepareHistogram("transaction_genesis_hash_resource_create_seconds", "HistogramTransactionGenesisHashResourceCreateSeconds", "code")
var HistogramTransactionGenesisHashResourceUpdate = PrometheusWrapper.PrepareHistogram("transaction_genesis_hash_resource_seconds", "HistogramTransactionGenesisHashResourceUpdateSeconds", "code")

// App is where all routes and middleware for buffalo
// should be defined. This is the nerve center of your
// application.
func App() *buffalo.App {
	if app == nil {
		app = buffalo.New(buffalo.Options{
			Env:          ENV,
			LooseSlash:   true,
			SessionStore: sessions.Null{},
			PreWares: []buffalo.PreWare{
				cors.AllowAll().Handler,
			},
			SessionName: "_brokernode_session",
			WorkerOff:   false,
			Worker:      jobs.OysterWorker,
		})

		// Setup sentry
		ravenDSN := os.Getenv("SENTRY_DSN")
		if ravenDSN != "" {
			raven.SetDSN(ravenDSN)
		}

		// Automatically redirect to SSL
		app.Use(ssl.ForceSSL(secure.Options{
			SSLRedirect:     ENV == "production",
			SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
		}))

		// Set the request content type to JSON
		app.Use(middleware.SetContentType("application/json"))

		if ENV == "development" {
			app.Use(middleware.ParameterLogger)
		}

		// Wraps each request in a transaction.
		//  c.Value("tx").(*pop.PopTransaction)
		// Remove to disable this.
		app.Use(middleware.PopTransaction(models.DB))

		app.GET("/", HomeHandler)

		app.GET("/metrics", buffalo.WrapHandler(prometheus.Handler()))

		apiV2 := app.Group("/api/v2")

		// UploadSessions
		uploadSessionResource := UploadSessionResource{}
		// apiV2.Resource("/upload-sessions", &UploadSessionResource{&buffalo.BaseResource{}})
		apiV2.POST("upload-sessions", uploadSessionResource.Create)
		apiV2.PUT("upload-sessions/{id}", uploadSessionResource.Update)
		apiV2.POST("upload-sessions/beta", uploadSessionResource.CreateBeta)
		apiV2.GET("upload-sessions/{id}", uploadSessionResource.GetPaymentStatus)

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
	}

	return app
}

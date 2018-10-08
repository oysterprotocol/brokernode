package actions

import (
	"os"

	raven "github.com/getsentry/raven-go"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/middleware"
	"github.com/gobuffalo/buffalo/middleware/ssl"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/x/sessions"
	"github.com/oysterprotocol/brokernode/actions/v2"
	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/cors"
	"github.com/unrolled/secure"
)

// ENV is used to help switch settings based on where the
// application is being run. Default is "development".
var ENV = envy.Get("GO_ENV", "development")
var app *buffalo.App

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
		uploadSessionResource := actions_v2.UploadSessionResourceV2{}
		// apiV2.Resource("/upload-sessions", &UploadSessionResource{&buffalo.BaseResource{}})
		apiV2.POST("upload-sessions", uploadSessionResource.Create)
		apiV2.PUT("upload-sessions/{id}", uploadSessionResource.Update)
		apiV2.POST("upload-sessions/beta", uploadSessionResource.CreateBeta)
		apiV2.GET("upload-sessions/{id}", uploadSessionResource.GetPaymentStatus)

		// Webnodes
		webnodeResource := actions_v2.WebnodeResource{}
		apiV2.POST("supply/webnodes", webnodeResource.Create)

		// Transactions
		transactionBrokernodeResource := actions_v2.TransactionBrokernodeResource{}
		apiV2.POST("demand/transactions/brokernodes", transactionBrokernodeResource.Create)
		apiV2.PUT("demand/transactions/brokernodes/{id}", transactionBrokernodeResource.Update)

		transactionGenesisHashResource := actions_v2.TransactionGenesisHashResource{}
		apiV2.POST("demand/transactions/genesis_hashes", transactionGenesisHashResource.Create)
		apiV2.PUT("demand/transactions/genesis_hashes/{id}", transactionGenesisHashResource.Update)

		// Treasures
		treasures := actions_v2.TreasuresResource{}
		apiV2.POST("treasures", treasures.VerifyAndClaim)

		// Status
		statusResource := StatusResource{}
		apiV2.GET("status", statusResource.CheckStatus)
	}

	oyster_utils.StartProfile()
	defer oyster_utils.StopProfile()

	return app
}

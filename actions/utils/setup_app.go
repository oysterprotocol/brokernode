package actions_utils

import (
	"os"

	raven "github.com/getsentry/raven-go"

	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/models"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/middleware"
	"github.com/gobuffalo/buffalo/middleware/ssl"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/x/sessions"
	"github.com/rs/cors"
	"github.com/unrolled/secure"
)

// ENV is used to help switch settings based on where the
// application is being run. Default is "development".
var ENV = envy.Get("GO_ENV", "development")

func CreateBuffaloApp() *buffalo.App {

	app := buffalo.New(buffalo.Options{
		Env:          ENV,
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

	return app
}

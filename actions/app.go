package actions

import (
	"github.com/oysterprotocol/brokernode/actions/utils"
	"github.com/oysterprotocol/brokernode/actions/v2"
	"github.com/oysterprotocol/brokernode/actions/v3"
	"github.com/oysterprotocol/brokernode/utils"

	"github.com/gobuffalo/buffalo"
	"github.com/prometheus/client_golang/prometheus"
)

var app *buffalo.App

// App is where all routes and middleware for buffalo
// should be defined. This is the nerve center of your
// application.
func App() *buffalo.App {
	if app == nil {
		app = actions_utils.CreateBuffaloApp()

		app.GET("/", HomeHandler)

		app.GET("/metrics", buffalo.WrapHandler(prometheus.Handler()))

		actions_v2.RegisterApi(app)

		actions_v3.RegisterApi(app)
	}

	oyster_utils.StartProfile()
	defer oyster_utils.StopProfile()

	return app
}

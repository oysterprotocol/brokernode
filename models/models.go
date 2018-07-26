package models

import (
	"log"
	"time"

	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/pop"
	"github.com/oysterprotocol/brokernode/utils"
)

// DB is a connection to your database to be used
// throughout your application.
var DB *pop.Connection

const TestValueTimeToLive = 3 * time.Minute

func init() {
	var err error
	env := envy.Get("GO_ENV", "development")
	DB, err = pop.Connect(env)
	if err != nil {
		log.Fatal(err)
	}

	oyster_utils.SetLogInfoForDatabaseUrl(DB.URL())
	pop.Debug = env == "development"
}

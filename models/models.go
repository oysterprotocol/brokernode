package models

import (
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/pop"
	"github.com/oysterprotocol/brokernode/utils"
	"log"
	"time"
)

// DB is a connection to your database to be used
// throughout your application.
var DB *pop.Connection

/*TestValueTimeToLive is some default value we can use in unit
tests for K:V pairs in badger*/
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

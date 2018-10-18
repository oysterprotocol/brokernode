package models

import (
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/pop"
	"github.com/oysterprotocol/brokernode/utils"
	"github.com/oysterprotocol/brokernode/utils/eth_gateway"
	"log"
)

// DB is a connection to your database to be used
// throughout your application.
var DB *pop.Connection

var (
	/*EthWrapper will be used to interact with the ethereum blockchain*/
	EthWrapper = eth_gateway.EthWrapper
)

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

package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/oysterprotocol/brokernode/actions"
	"github.com/oysterprotocol/brokernode/services"
)

func main() {
	pop.Debug = false
	// Setup rand. See https://til.hashrocket.com/posts/355f31f19c-seeding-golangs-rand
	rand.Seed(time.Now().Unix())
	app := actions.App()

	// Setup KV Store
	db, err := services.InitKVStore()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := app.Serve(); err != nil {
		log.Fatal(err)
	}
}

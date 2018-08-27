package main

import (
	"github.com/oysterprotocol/brokernode/utils"
	"log"
	"math/rand"
	"os"
	"runtime"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/oysterprotocol/brokernode/actions"
)

func main() {
	// Allow running on multiple cores. Kinda weird that this is manual?
	// https://golang.org/pkg/runtime/#GOMAXPROCS
	runtime.GOMAXPROCS(runtime.NumCPU())

	pop.Debug = false
	// Setup rand. See https://til.hashrocket.com/posts/355f31f19c-seeding-golangs-rand
	rand.Seed(time.Now().Unix())
	app := actions.App()
	defer oyster_utils.CloseKvStore()

	if os.Getenv("DATA_MAPS_IN_BADGER") != "true" {
		// Setup KV Store
		err := oyster_utils.InitKvStore()
		if err != nil {
			log.Fatal(err)
		}
	}

	if err := app.Serve(); err != nil {
		log.Fatal(err)
	}
}

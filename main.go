package main

import (
	"github.com/oysterprotocol/brokernode/actions"
	"log"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	app := actions.App()
	if err := app.Serve(); err != nil {
		log.Fatal(err)
	}
}

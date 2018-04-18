package main

import (
	"github.com/gobuffalo/pop"
	"github.com/oysterprotocol/brokernode/actions"
	"log"
)

func main() {
	pop.Debug = false
	app := actions.App()
	if err := app.Serve(); err != nil {
		log.Fatal(err)
	}
}

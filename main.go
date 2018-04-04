package main

import (
	"github.com/oysterprotocol/brokernode/actions"
	"log"
)

func main() {
	app := actions.App()
	if err := app.Serve(); err != nil {
		log.Fatal(err)
	}
}

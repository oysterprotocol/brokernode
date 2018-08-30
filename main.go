package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/oysterprotocol/brokernode/utils"
	"log"
	"math/rand"
	"os"
	"runtime"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/pop/logging"
	"github.com/oysterprotocol/brokernode/actions"
)

var Log = false

var oysterLogger = func(lvl logging.Level, s string, args ...interface{}) {
	if lvl == logging.SQL {
		if len(args) > 0 {
			xargs := make([]string, len(args))
			for i, a := range args {
				switch a.(type) {
				case string:
					xargs[i] = fmt.Sprintf("%q", a)
				default:
					xargs[i] = fmt.Sprintf("%v", a)
				}
			}
			s = fmt.Sprintf("%s - %s | %s", lvl, s, xargs)
		} else {
			s = fmt.Sprintf("%s - %s", lvl, s)
		}
	} else {
		s = fmt.Sprintf(s, args...)
		s = fmt.Sprintf("%s - %s", lvl, s)
	}
	if pop.Color {
		s = color.YellowString(s)
	}
	if Log {
		fmt.Println(s)
	}
}

func main() {
	// Allow running on multiple cores. Kinda weird that this is manual?
	// https://golang.org/pkg/runtime/#GOMAXPROCS
	runtime.GOMAXPROCS(runtime.NumCPU())

	pop.Debug = false
	pop.SetLogger(oysterLogger)

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

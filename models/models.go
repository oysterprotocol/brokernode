package models

import (
	"log"

	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
)

// DB is a connection to your database to be used
// throughout your application.
var DB *pop.Connection
var EmptyUUID uuid.UUID

func init() {
	var err error
	env := envy.Get("GO_ENV", "development")
	DB, err = pop.Connect(env)
	EmptyUUID, _ = uuid.FromString("00000000-0000-0000-0000-000000000000")
	if err != nil {
		log.Fatal(err)
	}
	pop.Debug = env == "development"
}

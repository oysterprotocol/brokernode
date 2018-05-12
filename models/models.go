package models

import (
	"log"

	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
)

// List a set of API we used for pop.Connection.
// Add more method if needed for pop.Connection. See https://godoc.org/github.com/markbates/pop
type DBInterface interface {
	RawQuery(stmt string, args ...interface{}) *pop.Query
	Where(stmt string, args ...interface{}) *pop.Query
	ValidateAndSave(model interface{}, excludeColumns ...string) (*validate.Errors, error)
	ValidateAndCreate(model interface{}, excludeColumns ...string) (*validate.Errors, error)
	Transaction(fn func(tx *pop.Connection) error) error
	Destroy(model interface{}) error
	Limit(limit int) *pop.Query
	Save(model interface{}, excludeColumns ...string) error
	Find(model interface{}, id interface{}) error

	// See PR: https://github.com/gobuffalo/pop/pull/14. the return type changes from "*pop.Query" to "*pop.Connection"
	// in this PR.
	Eager(fields ...string) *pop.Connection
}

// DB is a connection to your database to be used
// throughout your application.
var DB DBInterface

func init() {
	var err error
	env := envy.Get("GO_ENV", "development")
	DB, err = pop.Connect(env)
	if err != nil {
		log.Fatal(err)
	}
	pop.Debug = env == "development"
}

package jobs_test

import (
	"testing"
	"github.com/gobuffalo/suite"
)

type ModelSuite struct {
	*suite.Model
}

func Test_Jobs(t *testing.T) {
	as := &ModelSuite{suite.NewModel()}
	suite.Run(t, as)
}

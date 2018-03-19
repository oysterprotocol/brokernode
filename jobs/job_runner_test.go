package jobs_test

import (
	"github.com/gobuffalo/suite"
	"testing"
)

type JobsSuite struct {
	*suite.Model
}

//func (suite *JobsSuite) SetupSuite() {
//}
//
//func (suite *JobsSuite) TearDownSuite() {
//}
//
//func (suite *JobsSuite) SetupTest() {
//}
//
//func (suite *JobsSuite) TearDownTest() {
//}

func Test_JobsSuite(t *testing.T) {
	as := &JobsSuite{suite.NewModel()}
	suite.Run(t, as)
}

package actions

import (
	"testing"

	"github.com/gobuffalo/suite"
	"github.com/oysterprotocol/brokernode/utils"
)

type ActionSuite struct {
	*suite.Action
}


func (suite *ActionSuite) SetupTest() {
	suite.Action.SetupTest()

	suite.Nil(oyster_utils.InitKvStore())
}

func Test_ActionSuite(t *testing.T) {
	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	defer oyster_utils.ResetBrokerMode()
	as := &ActionSuite{suite.NewAction(App())}
	suite.Run(t, as)
}

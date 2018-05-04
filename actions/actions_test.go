package actions

import (
	"github.com/oysterprotocol/brokernode/utils"
	"testing"

	"github.com/gobuffalo/suite"
)

type ActionSuite struct {
	*suite.Action
}

func Test_ActionSuite(t *testing.T) {
	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	defer oyster_utils.ResetBrokerMode()
	as := &ActionSuite{suite.NewAction(App())}
	suite.Run(t, as)
}

package actions_v2

import (
	"github.com/gobuffalo/suite"
	"github.com/oysterprotocol/brokernode/actions"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"testing"
)

type ActionSuite struct {
	*suite.Action
}

func (suite *ActionSuite) SetupTest() {
	suite.Action.SetupTest()

	suite.Nil(oyster_utils.InitKvStore())

	EthWrapper = services.EthWrapper
	IotaWrapper = services.IotaWrapper
}

func Test_ActionSuite(t *testing.T) {
	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	defer oyster_utils.ResetBrokerMode()
	as := &ActionSuite{suite.NewAction(actions.App())}
	suite.Run(t, as)
}

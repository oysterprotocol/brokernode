package actions

import (
	"testing"

	"github.com/gobuffalo/suite"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
)

type ActionSuite struct {
	*suite.Action
}

func (suite *ActionSuite) SetupTest() {
	suite.Action.SetupTest()

	if oyster_utils.IsKvStoreEnabled() {
		suite.Nil(oyster_utils.RemoveAllKvStoreData())
		suite.Nil(oyster_utils.InitKvStore())
	}

	EthWrapper = services.EthWrapper
	IotaWrapper = services.IotaWrapper
}

func Test_ActionSuite(t *testing.T) {
	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	defer oyster_utils.ResetBrokerMode()
	as := &ActionSuite{suite.NewAction(App())}
	suite.Run(t, as)
}

package actions_v2

import (
	"github.com/gobuffalo/suite"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
)

type ActionSuite struct {
	*suite.Action
}

func (suite *ActionSuite) SetupTest() {
	suite.Action.SetupTest()

	suite.Nil(oyster_utils.InitKvStore())

	EthWrapper = oyster_utils.EthWrapper
	IotaWrapper = services.IotaWrapper
}

package actions_v2

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

	suite.Nil(oyster_utils.InitKvStore())

	EthWrapper = services.EthWrapper
	IotaWrapper = services.IotaWrapper
}

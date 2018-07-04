package models_test

import (
	"testing"

	"github.com/gobuffalo/suite"
	"github.com/oysterprotocol/brokernode/utils"
)

type ModelSuite struct {
	*suite.Model
}

func Test_ModelSuite(t *testing.T) {
	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	defer oyster_utils.ResetBrokerMode()
	as := &ModelSuite{suite.NewModel()}
	suite.Run(t, as)
}

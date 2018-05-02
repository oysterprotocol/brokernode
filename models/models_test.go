package models_test

import (
	"github.com/gobuffalo/suite"
	"github.com/oysterprotocol/brokernode/utils"
	"testing"
)

type ModelSuite struct {
	*suite.Model
}

func Test_ModelSuite(t *testing.T) {
	defer oyster_utils.ResetBrokerMode()
	oyster_utils.SetBrokerMode(oyster_utils.ProdMode)
	as := &ModelSuite{suite.NewModel()}
	suite.Run(t, as)
}

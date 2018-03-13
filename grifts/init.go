package grifts

import (
	"github.com/gobuffalo/buffalo"
	"github.com/oysterprotocol/brokernode/actions"
)

func init() {
	buffalo.Grifts(actions.App())
}

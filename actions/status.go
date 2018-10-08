package actions

import (
	"os"

	"github.com/gobuffalo/buffalo"
	"github.com/oysterprotocol/brokernode/actions/utils"
)

/*StatusResource is a resource for the status endpoint*/
type StatusResource struct {
	buffalo.Resource
}

// Response structs
type checkStatusRes struct {
	Available bool `json:"available"`
}

/*CheckStatus checks conditions to determine if the brokernode is available.
We can add more conditions to this method as needed*/
func (status *StatusResource) CheckStatus(c buffalo.Context) error {

	available := os.Getenv("DEPLOY_IN_PROGRESS") != "true"

	res := checkStatusRes{
		Available: available,
	}

	return c.Render(200, actions_utils.Render.JSON(res))
}

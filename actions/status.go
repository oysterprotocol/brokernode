package actions

import (
	"github.com/gobuffalo/buffalo"
	"os"
)

type StatusResource struct {
	buffalo.Resource
}

// Response structs
type checkAvailabilityRes struct {
	Available bool `json:"available"`
}

/*CheckAvailability checks conditions to determine if the brokernode is available.
We can add more conditions to this method as needed*/
func (status *StatusResource) CheckAvailability(c buffalo.Context) error {

	available := os.Getenv("DEPLOY_IN_PROGRESS") != "true"

	res := checkAvailabilityRes{
		Available: available,
	}

	return c.Render(200, r.JSON(res))
}

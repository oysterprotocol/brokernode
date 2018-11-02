package actions_v2

import (
	"os"
	"strconv"

	"github.com/gobuffalo/buffalo"
	"github.com/oysterprotocol/brokernode/actions/utils"
)

/*StatusResource is a resource for the status endpoint*/
type StatusResource struct {
	buffalo.Resource
}

// Response structs
type checkStatusRes struct {
	Available      bool `json:"available"`
	NumChunksLimit int  `json:"numChunksLimit"`
}

/*CheckStatus checks conditions to determine if the brokernode is available.
We can add more conditions to this method as needed*/
func (status *StatusResource) CheckStatus(c buffalo.Context) error {

	numChunksLimit := -1 //unlimited

	available := os.Getenv("DEPLOY_IN_PROGRESS") != "true"
	if v, err := strconv.Atoi(os.Getenv("NUM_CHUNKS_LIMIT")); err == nil {
		numChunksLimit = v
	}

	res := checkStatusRes{
		Available:      available,
		NumChunksLimit: numChunksLimit,
	}

	return c.Render(200, actions_utils.Render.JSON(res))
}

package actions_utils

import (
	"github.com/gobuffalo/buffalo/render"
)

var Render *render.Engine

func init() {
	Render = render.New(render.Options{})
}

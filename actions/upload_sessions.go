package actions

import "github.com/gobuffalo/buffalo"

type UploadSessionResource struct {
	buffalo.Resource
}

// Create creates an upload session.
func (usr *UploadSessionResource) Create(c buffalo.Context) error {
	c.Param("genesisHash")
	return c.Render(200, r.JSON(map[string]string{"message": "Welcome to Buffalo!"}))
}

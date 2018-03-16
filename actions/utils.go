package actions

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// parseReqBody take a request and parses the body to the target interface.
func parseReqBody(req *http.Request, dest interface{}) (err error) {
	body := req.Body
	defer body.Close()

	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return
	}
	err = json.Unmarshal(bodyBytes, dest)

	return
}

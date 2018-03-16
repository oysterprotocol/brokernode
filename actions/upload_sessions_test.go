package actions

import (
	"encoding/json"
	"io/ioutil"

	"github.com/oysterprotocol/brokernode/models"
)

func (as *ActionSuite) Test_UploadSessionsCreate() {
	res := as.JSON("/api/v2/upload-sessions").Post(map[string]interface{}{
		"genesisHash":   "genesisHashTest",
		"fileSizeBytes": 123,
	})

	// Parse response
	resParsed := uploadSessionCreateRes{}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	as.Nil(err)
	err = json.Unmarshal(bodyBytes, &resParsed)
	as.Nil(err)

	as.Equal(200, res.Code)
	as.Equal("genesisHashTest", resParsed.GenesisHash)
	as.Equal(123, resParsed.FileSizeBytes)
	as.Equal(models.SessionTypeAlpha, resParsed.Type)
}

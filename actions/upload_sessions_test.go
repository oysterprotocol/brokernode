package actions

import (
	"encoding/json"
	"io/ioutil"

	"github.com/oysterprotocol/brokernode/models"
)

func (as *ActionSuite) Test_UploadSessionsCreate() {
	res := as.JSON("/api/v2/upload-sessions").Post(map[string]interface{}{
		"genesisHash":          "genesisHashTest",
		"fileSizeBytes":        123,
		"storageLengthInYears": 1,
	})

	// Parse response
	resParsed := uploadSessionCreateRes{}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	as.Nil(err)
	err = json.Unmarshal(bodyBytes, &resParsed)
	as.Nil(err)

	as.Equal(200, res.Code)
	as.Equal("genesisHashTest", resParsed.UploadSession.GenesisHash)
	as.Equal(123, resParsed.UploadSession.FileSizeBytes)
	as.Equal(models.SessionTypeAlpha, resParsed.UploadSession.Type)
	as.NotEqual(0, resParsed.Invoice.Cost)
	as.NotEqual("", resParsed.Invoice.EthAddress)
}

func (as *ActionSuite) Test_UploadSessionsCreateBeta() {
	res := as.JSON("/api/v2/upload-sessions/beta").Post(map[string]interface{}{
		"genesisHash":          "genesisHashTest",
		"fileSizeBytes":        123,
		"storageLengthInYears": 1,
		"alphaTreasureIndexes": []int{1},
	})

	// Parse response
	resParsed := uploadSessionCreateRes{}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	as.Nil(err)
	err = json.Unmarshal(bodyBytes, &resParsed)
	as.Nil(err)

	as.Equal(200, res.Code)
	as.Equal("genesisHashTest", resParsed.UploadSession.GenesisHash)
	as.Equal(123, resParsed.UploadSession.FileSizeBytes)
	as.Equal(models.SessionTypeBeta, resParsed.UploadSession.Type)
	as.True(1 == len(resParsed.BetaTreasureIndexes))
	as.NotEqual(0, resParsed.Invoice.Cost)
	as.NotEqual("", resParsed.Invoice.EthAddress)
}

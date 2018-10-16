package actions_v2

import (
	"encoding/json"
	"io/ioutil"
)

func (suite *ActionSuite) Test_CheckStatus() {
	res := suite.JSON("/api/v2/status").Get()

	suite.Equal(200, res.Code)

	// Parse response
	resParsed := checkStatusRes{}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	suite.Nil(err)
	err = json.Unmarshal(bodyBytes, &resParsed)
	suite.Nil(err)

	suite.Equal(true, resParsed.Available == true || resParsed.Available == false)
}

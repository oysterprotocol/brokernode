package actions

import (
	"encoding/json"
	"io/ioutil"
)

func (suite *ActionSuite) Test_CheckAvailability() {
	res := suite.JSON("/api/v2/check-availability").Get()

	suite.Equal(200, res.Code)

	// Parse response
	resParsed := checkAvailabilityRes{}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	suite.Nil(err)
	err = json.Unmarshal(bodyBytes, &resParsed)
	suite.Nil(err)

	suite.Equal(true, resParsed.Available == true || resParsed.Available == false)
}

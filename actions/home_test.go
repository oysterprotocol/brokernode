package actions

func (suite *ActionSuite) Test_HomeHandler() {
	res := suite.JSON("/").Get()
	suite.Equal(200, res.Code)
	suite.Contains(res.Body.String(), "Welcome to Buffalo")
}

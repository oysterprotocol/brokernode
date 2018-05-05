package actions

func (as *ActionSuite) Test_VerifyAndClaim_NoError() {
	res := as.JSON("/api/v2/treasures").Post(map[string]interface{}{
		"receiverEthAddr": "receiverEthAddr",
		"genesisHash":     "123",
	})

	as.Equal(200, res.Code)
}

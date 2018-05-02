package actions

func (as *ActionSuite) Test_VerifyAndClaim_NoError() {
	t := TreasuresResource{}

	as.Nil(t.VerifyAndClaim(nil))
}

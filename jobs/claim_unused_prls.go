package jobs

func init() {
}

func ClaimUnusedPRLs() {
	CheckExistingClaimAttempts()
	StartNewClaims()
}

func CheckExistingClaimAttempts() {
	RetryClaims()
}

func RetryClaims() {

}

func StartNewClaims() {

}

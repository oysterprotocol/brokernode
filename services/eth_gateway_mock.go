package services

var EthMock Eth

func init() {
	SetUpMock()
}

func SetUpMock() {

	EthMock = Eth{}
}

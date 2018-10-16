package oyster_utils

var EthMock Eth

func init() {
	SetUpMock()
}

func SetUpMock() {

	EthMock = Eth{}
}

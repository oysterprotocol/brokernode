package oyster_utils

import (
	"log"
	"os"
)

/*BrokerModeStatus reflects what mode the broker is in:
ProdMode
TestModeDummyTreasure
TestModeNoTreasure*/
type BrokerModeStatus int

/*PoWModeStatus reflects whether the PoW on the broker is enabled or not*/
type PoWModeStatus int

const (
	/*ProdMode - the broker is requiring real PRL for uploads*/
	ProdMode BrokerModeStatus = iota + 1
	/*TestModeDummyTreasure - the broker does not require PRL and will bury a dummy treasure*/
	TestModeDummyTreasure
	/*TestModeNoTreasure - the broker does not require PRL and will not bury a treasure*/
	TestModeNoTreasure
)

const (
	/*PoWEnabled - the broker is doing PoW*/
	PoWEnabled PoWModeStatus = iota + 1
	/*PoWDisabled - the broker is not doing PoW (for webnode QAing)*/
	PoWDisabled
)

/*BrokerMode - the mode the broker is in (prod, dummy treasure, etc.)*/
var BrokerMode BrokerModeStatus

/*PoWMode - whether PoW is enabled or disabled*/
var PoWMode PoWModeStatus

func init() {

	ResetPoWMode()
	ResetBrokerMode()
}

/*ResetPoWMode - resets the PoWMode to whatever is in the .env file*/
func ResetPoWMode() {
	powMode := os.Getenv("DISABLE_POW")

	var mode PoWModeStatus

	switch powMode {
	case "true":
		mode = PoWDisabled
	case "false":
		mode = PoWEnabled
	case "":
		mode = PoWEnabled
	default:
		mode = PoWEnabled
	}
	SetPoWMode(mode)
}

/*ResetBrokerMode - resets the broker mode to whatever is in the .env file*/
func ResetBrokerMode() {
	brokerMode := os.Getenv("MODE")

	var mode BrokerModeStatus

	switch brokerMode {
	case "PROD_MODE":
		mode = ProdMode
	case "TEST_MODE_NO_TREASURE":
		mode = TestModeNoTreasure
	case "TEST_MODE_DUMMY_TREASURE":
		mode = TestModeDummyTreasure
	default:
		mode = ProdMode
	}
	SetBrokerMode(mode)
}

/*SetPoWMode - allow to change the mode within the code (such as for unit tests)*/
func SetPoWMode(powMode PoWModeStatus) {
	PoWMode = powMode
}

/*SetBrokerMode - allow to change the mode within the code (such as for unit tests)*/
func SetBrokerMode(brokerMode BrokerModeStatus) {
	BrokerMode = brokerMode

	switch brokerMode {
	case ProdMode:
		log.Println("Make sure you set the correct mode in .env!  Broker running in PROD_MODE")
	case TestModeNoTreasure:
		log.Println("Make sure you set the correct mode in .env!  Broker running in TEST_MODE_NO_TREASURE")
	case TestModeDummyTreasure:
		log.Println("Make sure you set the correct mode in .env!  Broker running in TEST_MODE_DUMMY_TREASURE")
	default:
		log.Println("No MODE given, defaulting to PROD_MODE")
	}
}

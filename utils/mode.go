package oyster_utils

import (
	"log"
	"os"
)

type BrokerModeStatus int
type PoWModeStatus int

const (
	ProdMode BrokerModeStatus = iota + 1
	TestModeDummyTreasure
	TestModeNoTreasure
)

const (
	PoWEnabled PoWModeStatus = iota + 1
	PoWDisabled
)

var BrokerMode BrokerModeStatus
var PoWMode PoWModeStatus

func init() {

	ResetPoWMode()
	ResetBrokerMode()
}

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

func SetPoWMode(powMode PoWModeStatus) {
	PoWMode = powMode
}

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

package oyster_utils

import (
	"log"
	"os"
)

type ModeStatus int

const (
	ProdMode ModeStatus = iota + 1
	TestModeDummyTreasure
	TestModeNoTreasure
)

var BrokerMode ModeStatus

func init() {

	ResetBrokerMode()
}

func ResetBrokerMode() {
	brokerMode := os.Getenv("MODE")

	var mode ModeStatus

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

func SetBrokerMode(brokerMode ModeStatus) {
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

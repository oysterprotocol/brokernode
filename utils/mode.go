package oyster_utils

import (
	"github.com/getsentry/raven-go"
	"github.com/joho/godotenv"
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

	// Load ENV variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
		raven.CaptureError(err, nil)
	}

	brokerMode := os.Getenv("MODE")

	switch brokerMode {
	case "PROD_MODE":
		log.Println("Broker mode set to PROD_MODE")
		BrokerMode = ProdMode
	case "TEST_MODE_NO_TREASURE":
		log.Println("Make sure you set the correct mode in .env file!  Broker running in TEST_MODE_NO_TREASURE")
		BrokerMode = TestModeNoTreasure
	case "TEST_MODE_DUMMY_TREASURE":
		log.Println("Make sure you set the correct mode in .env!  Broker running in TEST_MODE_DUMMY_TREASURE")
		BrokerMode = TestModeDummyTreasure
	default:
		log.Println("No MODE given, defaulting to PROD_MODE")
		BrokerMode = ProdMode
	}
}

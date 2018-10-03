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

/*OysterPaysStatus reflects whether uploads are free--this is mostly needed for unit tests*/
type OysterPaysStatus int

/*DataMapsStorageStatus reflects whether we are storing data maps in sql or badger*/
type DataMapsStorageStatus int

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

const (
	/*UserIsPaying is when the user is paying for their own uploads*/
	UserIsPaying OysterPaysStatus = iota + 1
	/*OysterIsPaying is when Oyster is paying for the uploads*/
	OysterIsPaying
)

const (
	/*DataMapsInSQL is when the data maps are stored in SQL*/
	DataMapsInSQL DataMapsStorageStatus = iota + 1
	/*DataMapsInBadger is when the data maps are stored in badger*/
	DataMapsInBadger
	/*DataMapsInS3 is when the data maps are stored in S3*/
	DataMapsInS3
)

/*BrokerMode - the mode the broker is in (prod, dummy treasure, etc.)*/
var BrokerMode BrokerModeStatus

/*PoWMode - whether PoW is enabled or disabled*/
var PoWMode PoWModeStatus

/*PaymentMode - whether Oyster is paying or not*/
var PaymentMode OysterPaysStatus

/*DataMapStorageMode - where we are storing the data maps*/
var DataMapStorageMode DataMapsStorageStatus

func init() {

	ResetPoWMode()
	ResetBrokerMode()
	ResetPaymentMode()
	ResetDataMapStorageMode()
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

/*ResetPaymentMode - resets the payment mode to whatever is in the .env file*/
func ResetPaymentMode() {
	paymentMode := os.Getenv("OYSTER_PAYS")

	var mode OysterPaysStatus

	switch paymentMode {
	case "true":
		mode = OysterIsPaying
	default:
		mode = UserIsPaying
	}
	SetPaymentMode(mode)
}

/*ResetDataMapStorageMode - resets the storage mode to whatever is in the .env file*/
func ResetDataMapStorageMode() {
	dataMapsInBadger := os.Getenv("DATA_MAPS_IN_BADGER")

	var mode DataMapsStorageStatus

	switch dataMapsInBadger {
	case "true":
		mode = DataMapsInBadger
	default:
		mode = DataMapsInSQL
	}
	SetStorageMode(mode)
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

/*SetPaymentMode - allow to change the mode within the code (such as for unit tests)*/
func SetPaymentMode(paymentMode OysterPaysStatus) {
	PaymentMode = paymentMode
}

/*SetStorageMode - change where we are storing the data maps*/
func SetStorageMode(dataStorageMode DataMapsStorageStatus) {

	switch dataStorageMode {
	case DataMapsInSQL:
		InitKvStore()
		log.Println("Make sure you set the correct mode in .env!  Broker storage mode is DataMapsInSQL")
	case DataMapsInBadger:
		log.Println("Make sure you set the correct mode in .env!  Broker storage mode is DataMapsInBadger")
	}

	DataMapStorageMode = dataStorageMode
}

package actions

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
)

type mockWaitForTransfer struct {
	hasCalled        bool
	input_brokerAddr common.Address
	output_int       *big.Int
	output_error     error
}

type mockSendPrl struct {
	hasCalled   bool
	input_msg   services.OysterCallMsg
	output_bool bool
}

func (as *ActionSuite) Test_UploadSessionsCreate() {
	mockWaitForTransfer := mockWaitForTransfer{
		output_error: nil,
		output_int:   big.NewInt(100),
	}
	mockSendPrl := mockSendPrl{
		output_bool: true,
	}
	EthWrapper = services.Eth{
		WaitForTransfer: mockWaitForTransfer.waitForTransfer,
		SendPRL:         mockSendPrl.sendPrl,
		GenerateEthAddr: services.EthWrapper.GenerateEthAddr,
		GenerateKeys:    services.EthWrapper.GenerateKeys,
	}

	res := as.JSON("/api/v2/upload-sessions").Post(map[string]interface{}{
		"genesisHash":          "genesisHashTest",
		"fileSizeBytes":        123,
		"numChunks":            2,
		"storageLengthInYears": 1,
	})

	// Parse response
	resParsed := uploadSessionCreateRes{}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	as.Nil(err)
	err = json.Unmarshal(bodyBytes, &resParsed)
	as.Nil(err)

	as.Equal(200, res.Code)
	as.Equal("genesisHashTest", resParsed.UploadSession.GenesisHash)
	as.Equal(123, resParsed.UploadSession.FileSizeBytes)
	as.Equal(models.SessionTypeAlpha, resParsed.UploadSession.Type)
	as.NotEqual(0, resParsed.Invoice.Cost)
	as.NotEqual("", resParsed.Invoice.EthAddress)

	time.Sleep(50 * time.Millisecond) // Force it to wait for goroutine to excute.

	as.True(mockWaitForTransfer.hasCalled)
	as.Equal(services.StringToAddress(resParsed.UploadSession.ETHAddrAlpha.String), mockWaitForTransfer.input_brokerAddr)
	as.True(mockSendPrl.hasCalled)
	as.Equal(big.NewInt(50), &mockSendPrl.input_msg.Amount)

	verifyPaymentConfirmation(as, resParsed.ID)
}

func (as *ActionSuite) Test_UploadSessionsCreateBeta() {
	mockWaitForTransfer := mockWaitForTransfer{
		output_error: nil,
		output_int:   big.NewInt(100),
	}
	mockSendPrl := mockSendPrl{}

	EthWrapper = services.Eth{
		WaitForTransfer: mockWaitForTransfer.waitForTransfer,
		SendPRL:         mockSendPrl.sendPrl,
		GenerateEthAddr: services.EthWrapper.GenerateEthAddr,
		GenerateKeys:    services.EthWrapper.GenerateKeys,
	}

	res := as.JSON("/api/v2/upload-sessions/beta").Post(map[string]interface{}{
		"genesisHash":          "genesisHashTest",
		"fileSizeBytes":        123,
		"numChunks":            2,
		"storageLengthInYears": 1,
		"alphaTreasureIndexes": []int{1},
	})

	// Parse response
	resParsed := uploadSessionCreateBetaRes{}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	as.Nil(err)
	err = json.Unmarshal(bodyBytes, &resParsed)
	as.Nil(err)

	as.Equal(200, res.Code)
	as.Equal("genesisHashTest", resParsed.UploadSession.GenesisHash)
	as.Equal(123, resParsed.UploadSession.FileSizeBytes)
	as.Equal(models.SessionTypeBeta, resParsed.UploadSession.Type)
	as.True(1 == len(resParsed.BetaTreasureIndexes))
	as.NotEqual(0, resParsed.Invoice.Cost)
	as.NotEqual("", resParsed.Invoice.EthAddress)

	time.Sleep(50 * time.Millisecond) // Force it to wait for goroutine to excute.

	as.True(mockWaitForTransfer.hasCalled)
	as.Equal(services.StringToAddress(resParsed.UploadSession.ETHAddrAlpha.String), mockWaitForTransfer.input_brokerAddr)
	as.False(mockSendPrl.hasCalled)

	verifyPaymentConfirmation(as, resParsed.ID)
}

func (as *ActionSuite) Test_UploadSessionsGetPaymentStatus_Paid() {
	//setup
	uploadSession1 := models.UploadSession{
		GenesisHash:   "genHash1",
		FileSizeBytes: 123,
		NumChunks:     2,
		PaymentStatus: models.PaymentStatusConfirmed,
	}

	uploadSession1.StartUploadSession()

	session := models.UploadSession{}
	err := as.DB.Where("genesis_hash = ?", "genHash1").First(&session)
	as.Equal(err, nil)

	//execute method
	res := as.JSON("/api/v2/upload-sessions/" + fmt.Sprint(session.ID)).Get()

	// Parse response
	resParsed := paymentStatusCreateRes{}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	as.Nil(err)
	err = json.Unmarshal(bodyBytes, &resParsed)
	as.Nil(err)

	as.Equal("confirmed", resParsed.PaymentStatus)
}

func (as *ActionSuite) Test_UploadSessionsGetPaymentStatus_Pending() {
	//setup
	uploadSession1 := models.UploadSession{
		GenesisHash:   "genHash1",
		FileSizeBytes: 123,
		NumChunks:     2,
		PaymentStatus: models.PaymentStatusPending,
	}

	uploadSession1.StartUploadSession()

	session := models.UploadSession{}
	err := as.DB.Where("genesis_hash = ?", "genHash1").First(&session)
	as.Equal(err, nil)

	//execute method
	res := as.JSON("/api/v2/upload-sessions/" + fmt.Sprint(session.ID)).Get()

	// Parse response
	resParsed := paymentStatusCreateRes{}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	as.Nil(err)
	err = json.Unmarshal(bodyBytes, &resParsed)
	as.Nil(err)

	as.Equal("pending", resParsed.PaymentStatus)
}

func (as *ActionSuite) Test_UploadSessionsGetPaymentStatus_Invoiced() {
	//setup
	uploadSession1 := models.UploadSession{
		GenesisHash:   "genHash1",
		FileSizeBytes: 123,
		NumChunks:     2,
		PaymentStatus: models.PaymentStatusInvoiced,
	}

	uploadSession1.StartUploadSession()

	session := models.UploadSession{}
	err := as.DB.Where("genesis_hash = ?", "genHash1").First(&session)
	as.Equal(err, nil)

	//execute method
	res := as.JSON("/api/v2/upload-sessions/" + fmt.Sprint(session.ID)).Get()

	// Parse response
	resParsed := paymentStatusCreateRes{}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	as.Nil(err)
	err = json.Unmarshal(bodyBytes, &resParsed)
	as.Nil(err)

	as.Equal("invoiced", resParsed.PaymentStatus)
}

func (as *ActionSuite) Test_UploadSessionsGetPaymentStatus_Error() {
	//setup
	uploadSession1 := models.UploadSession{
		GenesisHash:   "genHash1",
		FileSizeBytes: 123,
		NumChunks:     2,
		PaymentStatus: models.PaymentStatusError,
	}

	uploadSession1.StartUploadSession()

	session := models.UploadSession{}
	err := as.DB.Where("genesis_hash = ?", "genHash1").First(&session)
	as.Equal(err, nil)

	//execute method
	res := as.JSON("/api/v2/upload-sessions/" + fmt.Sprint(session.ID)).Get()

	// Parse response
	resParsed := paymentStatusCreateRes{}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	as.Nil(err)
	err = json.Unmarshal(bodyBytes, &resParsed)
	as.Nil(err)

	as.Equal("error", resParsed.PaymentStatus)
}

func (as *ActionSuite) Test_UploadSessionsGetPaymentStatus_DoesntExist() {
	//res := as.JSON("/api/v2/upload-sessions/" + "noIDFound").Get()

	//TODO: Return better error response when ID does not exist
}

func verifyPaymentConfirmation(as *ActionSuite, sessionId string) {
	session := models.UploadSession{}
	err := as.DB.Find(&session, sessionId)
	as.Nil(err)
	as.True(session.PaymentStatus == models.PaymentStatusConfirmed)
}

func (v *mockWaitForTransfer) waitForTransfer(brokerAddr common.Address) (*big.Int, error) {
	v.hasCalled = true
	v.input_brokerAddr = brokerAddr
	return v.output_int, v.output_error
}

func (v *mockSendPrl) sendPrl(msg services.OysterCallMsg) bool {
	v.hasCalled = true
	v.input_msg = msg
	return v.output_bool
}

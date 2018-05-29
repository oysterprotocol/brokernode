package actions

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gobuffalo/pop/nulls"
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

type mockCheckPRLBalance struct {
	hasCalled  bool
	input_addr common.Address
	output_int *big.Int
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
		"genesisHash":          "abcdef",
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
	as.Equal("abcdef", resParsed.UploadSession.GenesisHash)
	as.Equal(123, resParsed.UploadSession.FileSizeBytes)
	as.Equal(models.SessionTypeAlpha, resParsed.UploadSession.Type)
	as.NotEqual(0, resParsed.Invoice.Cost)
	as.NotEqual("", resParsed.Invoice.EthAddress)

	time.Sleep(50 * time.Millisecond) // Force it to wait for goroutine to excute.

	as.True(mockWaitForTransfer.hasCalled)
	as.Equal(services.StringToAddress(resParsed.UploadSession.ETHAddrAlpha.String), mockWaitForTransfer.input_brokerAddr)

	// mockCheckPRLBalance will result a positive value, and Alpha knows that beta has such balance, it won't send
	// it again.
	as.False(mockSendPrl.hasCalled)

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
		"genesisHash":          "abcdef",
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
	as.Equal("abcdef", resParsed.UploadSession.GenesisHash)
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
	mockCheckPRLBalance := mockCheckPRLBalance{}
	EthWrapper = services.Eth{
		CheckPRLBalance: mockCheckPRLBalance.checkPRLBalance,
	}
	uploadSession1 := models.UploadSession{
		GenesisHash:   "abcdeff1",
		FileSizeBytes: 123,
		NumChunks:     2,
		PaymentStatus: models.PaymentStatusConfirmed,
	}

	resParsed := getPaymentStatus(uploadSession1, as)

	as.Equal("confirmed", resParsed.PaymentStatus)
	as.False(mockCheckPRLBalance.hasCalled)
}

func (as *ActionSuite) Test_UploadSessionsGetPaymentStatus_NoConfirmButCheckComplete() {
	//setup
	mockCheckPRLBalance := mockCheckPRLBalance{
		output_int: big.NewInt(10),
	}
	mockSendPrl := mockSendPrl{}
	EthWrapper = services.Eth{
		CheckPRLBalance: mockCheckPRLBalance.checkPRLBalance,
		SendPRL:         mockSendPrl.sendPrl,
	}
	uploadSession1 := models.UploadSession{
		GenesisHash:   "abcdeff1",
		FileSizeBytes: 123,
		NumChunks:     2,
		PaymentStatus: models.PaymentStatusPending,
		ETHAddrAlpha:  nulls.NewString("alpha"),
		ETHAddrBeta:   nulls.NewString("beta"),
	}

	resParsed := getPaymentStatus(uploadSession1, as)

	as.Equal("confirmed", resParsed.PaymentStatus)

	// checkPRLBalance has been called twice. 1st for Alpha, and 2nd for Beta, we only record Beta addr
	// Since both time, it returns a positive balance, thus, alpha won't call sendPrl method.
	as.True(mockCheckPRLBalance.hasCalled)
	as.False(mockSendPrl.hasCalled)
	as.Equal(services.StringToAddress(uploadSession1.ETHAddrBeta.String), mockCheckPRLBalance.input_addr)

	session := models.UploadSession{}
	as.Nil(as.DB.Find(&session, resParsed.ID))
	as.Equal(models.PaymentStatusConfirmed, session.PaymentStatus)
}

func (as *ActionSuite) Test_UploadSessionsGetPaymentStatus_NoConfirmAndCheckIncomplete() {
	//setup
	mockCheckPRLBalance := mockCheckPRLBalance{
		output_int: big.NewInt(0),
	}
	EthWrapper = services.Eth{
		CheckPRLBalance: mockCheckPRLBalance.checkPRLBalance,
	}
	uploadSession1 := models.UploadSession{
		GenesisHash:   "abcdeff1",
		FileSizeBytes: 123,
		NumChunks:     2,
		PaymentStatus: models.PaymentStatusInvoiced,
		ETHAddrAlpha:  nulls.NewString("alpha"),
	}

	resParsed := getPaymentStatus(uploadSession1, as)

	as.Equal("invoiced", resParsed.PaymentStatus)
	as.True(mockCheckPRLBalance.hasCalled)
	as.Equal(services.StringToAddress(uploadSession1.ETHAddrAlpha.String), mockCheckPRLBalance.input_addr)

	session := models.UploadSession{}
	as.Nil(as.DB.Find(&session, resParsed.ID))
	as.Equal(models.PaymentStatusInvoiced, session.PaymentStatus)
}

func (as *ActionSuite) Test_UploadSessionsGetPaymentStatus_BetaConfirmed() {
	mockCheckPRLBalance := mockCheckPRLBalance{
		output_int: big.NewInt(10),
	}
	mockSendPrl := mockSendPrl{}
	EthWrapper = services.Eth{
		CheckPRLBalance: mockCheckPRLBalance.checkPRLBalance,
		SendPRL:         mockSendPrl.sendPrl,
	}

	uploadSession1 := models.UploadSession{
		Type:          models.SessionTypeBeta,
		GenesisHash:   "abcdeff1",
		FileSizeBytes: 123,
		NumChunks:     2,
		PaymentStatus: models.PaymentStatusInvoiced,
		ETHAddrAlpha:  nulls.NewString("alpha"),
	}

	resParsed := getPaymentStatus(uploadSession1, as)

	as.Equal("confirmed", resParsed.PaymentStatus)
	as.True(mockCheckPRLBalance.hasCalled)
	as.False(mockSendPrl.hasCalled)

	session := models.UploadSession{}
	as.Nil(as.DB.Find(&session, resParsed.ID))
	as.Equal(models.PaymentStatusConfirmed, session.PaymentStatus)
}

func (as *ActionSuite) Test_UploadSessionsGetPaymentStatus_DoesntExist() {
	//res := as.JSON("/api/v2/upload-sessions/" + "noIDFound").Get()

	//TODO: Return better error response when ID does not exist
}

func getPaymentStatus(seededUploadSession models.UploadSession, as *ActionSuite) paymentStatusCreateRes {
	seededUploadSession.StartUploadSession()

	session := models.UploadSession{}
	as.Nil(as.DB.Where("genesis_hash = ?", seededUploadSession.GenesisHash).First(&session))

	//execute method
	res := as.JSON("/api/v2/upload-sessions/" + fmt.Sprint(session.ID)).Get()

	// Parse response
	resParsed := paymentStatusCreateRes{}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	as.Nil(err)

	as.Nil(json.Unmarshal(bodyBytes, &resParsed))

	return resParsed
}

func verifyPaymentConfirmation(as *ActionSuite, sessionId string) {
	session := models.UploadSession{}
	err := as.DB.Find(&session, sessionId)
	as.Nil(err)
	as.True(session.PaymentStatus == models.PaymentStatusConfirmed)
}

func (v *mockWaitForTransfer) waitForTransfer(brokerAddr common.Address, transferType string) (*big.Int, error) {
	v.hasCalled = true
	v.input_brokerAddr = brokerAddr
	return v.output_int, v.output_error
}

func (v *mockSendPrl) sendPrl(msg services.OysterCallMsg) bool {
	v.hasCalled = true
	v.input_msg = msg
	return v.output_bool
}

func (v *mockCheckPRLBalance) checkPRLBalance(addr common.Address) *big.Int {
	v.hasCalled = true
	v.input_addr = addr
	return v.output_int
}

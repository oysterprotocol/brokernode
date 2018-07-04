package actions

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"strconv"
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

func (suite *ActionSuite) Test_UploadSessionsCreate() {
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

	res := suite.JSON("/api/v2/upload-sessions").Post(map[string]interface{}{
		"genesisHash":          "abcdef",
		"fileSizeBytes":        123,
		"numChunks":            2,
		"storageLengthInYears": 1,
	})

	// Parse response
	resParsed := uploadSessionCreateRes{}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	suite.Nil(err)
	err = json.Unmarshal(bodyBytes, &resParsed)
	suite.Nil(err)

	suite.Equal(200, res.Code)
	suite.Equal("abcdef", resParsed.UploadSession.GenesisHash)
	suite.Equal(uint64(123), resParsed.UploadSession.FileSizeBytes)
	suite.Equal(models.SessionTypeAlpha, resParsed.UploadSession.Type)
	suite.NotEqual(0, resParsed.Invoice.Cost)
	suite.NotEqual("", resParsed.Invoice.EthAddress)

	time.Sleep(50 * time.Millisecond) // Force it to wait for goroutine to excute.

	// TODO: fix waitForTransfer and uncomment it out in
	// actions/upload_sessions.go then uncomment out these tests.
	//suite.True(mockWaitForTransfer.hasCalled)
	//suite.Equal(services.StringToAddress(resParsed.UploadSession.ETHAddrAlpha.String), mockWaitForTransfer.input_brokerAddr)

	// mockCheckPRLBalance will result a positive value, and Alpha knows that beta has such balance, it won't send
	// it again.
	suite.False(mockSendPrl.hasCalled)

	// TODO: fix waitForTransfer and uncomment it out in
	// actions/upload_sessions.go then uncomment out these tests.
	// verifyPaymentConfirmation(as, resParsed.ID)
}

func (suite *ActionSuite) Test_UploadSessionsCreateBeta() {
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

	res := suite.JSON("/api/v2/upload-sessions/beta").Post(map[string]interface{}{
		"genesisHash":          "abcdef",
		"fileSizeBytes":        123,
		"numChunks":            2,
		"storageLengthInYears": 1,
		"alphaTreasureIndexes": []int{1},
	})

	// Parse response
	resParsed := uploadSessionCreateBetaRes{}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	suite.Nil(err)
	err = json.Unmarshal(bodyBytes, &resParsed)
	suite.Nil(err)

	suite.Equal(200, res.Code)
	suite.Equal("abcdef", resParsed.UploadSession.GenesisHash)
	suite.Equal(uint64(123), resParsed.UploadSession.FileSizeBytes)
	suite.Equal(models.SessionTypeBeta, resParsed.UploadSession.Type)
	suite.Equal(1, len(resParsed.BetaTreasureIndexes))
	suite.NotEqual(0, resParsed.Invoice.Cost)
	suite.NotEqual("", resParsed.Invoice.EthAddress)

	time.Sleep(50 * time.Millisecond) // Force it to wait for goroutine to excute.

	// TODO: fix waitForTransfer and uncomment it out in
	// actions/upload_sessions.go then uncomment out this test.
	//suite.True(mockWaitForTransfer.hasCalled)
	suite.Equal(services.StringToAddress(resParsed.UploadSession.ETHAddrAlpha.String), mockWaitForTransfer.input_brokerAddr)
	suite.False(mockSendPrl.hasCalled)

	// TODO: fix waitForTransfer and uncomment it out in
	// actions/upload_sessions.go then uncomment out these tests.
	// verifyPaymentConfirmation(as, resParsed.ID)
}

func (suite *ActionSuite) Test_UploadSessionsGetPaymentStatus_Paid() {
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

	resParsed := getPaymentStatus(uploadSession1, suite)

	suite.Equal("confirmed", resParsed.PaymentStatus)
	suite.False(mockCheckPRLBalance.hasCalled)
}

func (suite *ActionSuite) Test_UploadSessionsGetPaymentStatus_NoConfirmButCheckComplete() {
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

	resParsed := getPaymentStatus(uploadSession1, suite)

	suite.Equal("confirmed", resParsed.PaymentStatus)

	/* checkPRLBalance has been called once just for alpha.  Sending
	to beta now occurs in a job. */
	suite.True(mockCheckPRLBalance.hasCalled)
	suite.False(mockSendPrl.hasCalled)
	suite.Equal(services.StringToAddress(uploadSession1.ETHAddrAlpha.String), mockCheckPRLBalance.input_addr)

	session := models.UploadSession{}
	suite.Nil(suite.DB.Find(&session, resParsed.ID))
	suite.Equal(models.PaymentStatusConfirmed, session.PaymentStatus)
}

func (suite *ActionSuite) Test_UploadSessionsGetPaymentStatus_NoConfirmAndCheckIncomplete() {
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

	resParsed := getPaymentStatus(uploadSession1, suite)

	suite.Equal("invoiced", resParsed.PaymentStatus)
	suite.True(mockCheckPRLBalance.hasCalled)
	suite.Equal(services.StringToAddress(uploadSession1.ETHAddrAlpha.String), mockCheckPRLBalance.input_addr)

	session := models.UploadSession{}
	suite.Nil(suite.DB.Find(&session, resParsed.ID))
	suite.Equal(models.PaymentStatusInvoiced, session.PaymentStatus)
}

func (suite *ActionSuite) Test_UploadSessionsGetPaymentStatus_BetaConfirmed() {
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

	resParsed := getPaymentStatus(uploadSession1, suite)

	suite.Equal("confirmed", resParsed.PaymentStatus)
	suite.True(mockCheckPRLBalance.hasCalled)
	suite.False(mockSendPrl.hasCalled)

	session := models.UploadSession{}
	suite.Nil(suite.DB.Find(&session, resParsed.ID))
	suite.Equal(models.PaymentStatusConfirmed, session.PaymentStatus)
}

func (suite *ActionSuite) Test_UploadSessionsGetPaymentStatus_DoesntExist() {
	//res := suite.JSON("/api/v2/upload-sessions/" + "noIDFound").Get()

	//TODO: Return better error response when ID does not exist
}

func (suite *ActionSuite) Test_ProcessAndStoreDataMap_ProcessAll() {
	vErr, err := suite.DB.ValidateAndCreate(&models.DataMap{
		GenesisHash: "123",
		ChunkIdx:    1,
		Hash:        "1",
	})
	suite.Nil(err)
	suite.False(vErr.HasAny())
	vErr, err = suite.DB.ValidateAndCreate(&models.DataMap{
		GenesisHash: "123",
		ChunkIdx:    2,
		Hash:        "2",
	})
	suite.Nil(err)
	suite.False(vErr.HasAny())

	ProcessAndStoreChunkData([]chunkReq{
		chunkReq{Idx: 1, Data: strconv.Itoa(1), Hash: "123"},
		chunkReq{Idx: 2, Data: strconv.Itoa(2), Hash: "123"}},
		"123",
		[]int{})

	var dms []models.DataMap
	suite.Nil(suite.DB.Where("genesis_hash = 123").All(&dms))

	suite.Equal(len(dms), 2)

	for _, dm := range dms {
		suite.Equal(dm.MsgStatus, models.MsgStatusUploaded)
		suite.Equal(dm.Message, "")

		values, err := services.BatchGet(&services.KVKeys{dm.MsgID})
		suite.Nil(err)
		suite.Equal(len(*values), 1)
		suite.Equal((*values)[dm.MsgID], strconv.Itoa(dm.ChunkIdx))
	}
}

func (suite *ActionSuite) Test_ProcessAndStoreDataMap_ProcessSome() {
	vErr, err := suite.DB.ValidateAndCreate(&models.DataMap{
		GenesisHash: "123",
		ChunkIdx:    3,
		Hash:        "3",
	})
	suite.Nil(err)
	suite.False(vErr.HasAny())
	vErr, err = suite.DB.ValidateAndCreate(&models.DataMap{
		GenesisHash: "123",
		ChunkIdx:    4,
		Hash:        "4",
	})

	ProcessAndStoreChunkData([]chunkReq{
		chunkReq{Idx: 3, Data: strconv.Itoa(3), Hash: "123"}},
		"123",
		[]int{})

	var dms []models.DataMap
	suite.Nil(suite.DB.Where("genesis_hash = 123").All(&dms))

	suite.Equal(len(dms), 2)

	for _, dm := range dms {
		isProccessed := dm.ChunkIdx == 3

		if isProccessed {
			suite.Equal(dm.MsgStatus, models.MsgStatusUploaded)
		} else {
			suite.Equal(dm.MsgStatus, models.MsgStatusNotUploaded)
		}

		values, err := services.BatchGet(&services.KVKeys{dm.MsgID})
		suite.Nil(err)

		if isProccessed {
			suite.Equal(len(*values), 1)
			suite.Equal((*values)[dm.MsgID], strconv.Itoa(dm.ChunkIdx))
		} else {
			suite.Equal(len(*values), 0)
		}
	}
}

func getPaymentStatus(seededUploadSession models.UploadSession, suite *ActionSuite) paymentStatusCreateRes {
	seededUploadSession.StartUploadSession()

	session := models.UploadSession{}
	suite.Nil(suite.DB.Where("genesis_hash = ?", seededUploadSession.GenesisHash).First(&session))

	//execute method
	res := suite.JSON("/api/v2/upload-sessions/" + fmt.Sprint(session.ID)).Get()

	// Parse response
	resParsed := paymentStatusCreateRes{}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	suite.Nil(err)

	suite.Nil(json.Unmarshal(bodyBytes, &resParsed))

	return resParsed
}

func verifyPaymentConfirmation(sessionId string, suite *ActionSuite) {
	session := models.UploadSession{}
	err := suite.DB.Find(&session, sessionId)
	suite.Nil(err)
	suite.Equal(models.PaymentStatusConfirmed, session.PaymentStatus)
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

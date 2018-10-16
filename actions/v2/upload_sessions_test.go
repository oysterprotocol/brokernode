package actions_v2

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/oysterprotocol/brokernode/utils"
	"io/ioutil"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gobuffalo/pop/nulls"
	"github.com/oysterprotocol/brokernode/models"
)

type mockWaitForTransfer struct {
	hasCalled        bool
	input_brokerAddr common.Address
	output_int       *big.Int
	output_error     error
}

type mockSendPrl struct {
	hasCalled   bool
	input_msg   oyster_utils.OysterCallMsg
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
	EthWrapper = oyster_utils.Eth{
		WaitForTransfer: mockWaitForTransfer.waitForTransfer,
		SendPRL:         mockSendPrl.sendPrl,
		GenerateEthAddr: oyster_utils.EthWrapper.GenerateEthAddr,
		GenerateKeys:    oyster_utils.EthWrapper.GenerateKeys,
	}

	genHash := oyster_utils.RandSeq(8, []rune("abcdef0123456789"))

	res := suite.JSON("/api/v2/upload-sessions").Post(map[string]interface{}{
		"genesisHash":          genHash,
		"fileSizeBytes":        123,
		"numChunks":            2,
		"storageLengthInYears": 1,
	})

	// Parse response
	resParsed := uploadSessionCreateResV2{}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	suite.Nil(err)
	err = json.Unmarshal(bodyBytes, &resParsed)
	suite.Nil(err)

	suite.Equal(200, res.Code)
	suite.Equal(genHash, resParsed.UploadSession.GenesisHash)
	suite.Equal(uint64(123), resParsed.UploadSession.FileSizeBytes)
	suite.Equal(models.SessionTypeAlpha, resParsed.UploadSession.Type)
	suite.NotEqual(0, resParsed.Invoice.Cost)
	suite.NotEqual("", resParsed.Invoice.EthAddress)

	time.Sleep(50 * time.Millisecond) // Force it to wait for goroutine to excute.

	// TODO: fix waitForTransfer and uncomment it out in
	// actions/upload_sessions.go then uncomment out these tests.
	//suite.True(mockWaitForTransfer.hasCalled)
	//suite.Equal(oyster_utils.StringToAddress(resParsed.UploadSession.ETHAddrAlpha.String), mockWaitForTransfer.input_brokerAddr)

	// mockCheckPRLBalance will result a positive value, and Alpha knows that beta has such balance, it won't send
	// it again.
	suite.False(mockSendPrl.hasCalled)

	// TODO: fix waitForTransfer and uncomment it out in
	// actions/upload_sessions.go then uncomment out these tests.
	// verifyPaymentConfirmation(as, resParsed.ID)

	chunkData := models.GetSingleChunkData(oyster_utils.InProgressDir, genHash, models.MetaDataChunkIdx)

	expectedHash := oyster_utils.HashHex(genHash, sha256.New())

	suite.Equal(expectedHash, chunkData.Hash)

	brokerTx := []models.BrokerBrokerTransaction{}

	suite.DB.Where("genesis_hash = ?", genHash).All(&brokerTx)

	suite.Equal(1, len(brokerTx))
}

func (suite *ActionSuite) Test_UploadSessionsCreateBeta() {
	mockWaitForTransfer := mockWaitForTransfer{
		output_error: nil,
		output_int:   big.NewInt(100),
	}
	mockSendPrl := mockSendPrl{}

	EthWrapper = oyster_utils.Eth{
		WaitForTransfer: mockWaitForTransfer.waitForTransfer,
		SendPRL:         mockSendPrl.sendPrl,
		GenerateEthAddr: oyster_utils.EthWrapper.GenerateEthAddr,
		GenerateKeys:    oyster_utils.EthWrapper.GenerateKeys,
	}

	genHash := oyster_utils.RandSeq(8, []rune("abcdef0123456789"))

	res := suite.JSON("/api/v2/upload-sessions/beta").Post(map[string]interface{}{
		"genesisHash":          genHash,
		"fileSizeBytes":        123,
		"numChunks":            2,
		"storageLengthInYears": 1,
		"alphaTreasureIndexes": []int{1},
	})

	// Parse response
	resParsed := uploadSessionCreateBetaResV2{}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	suite.Nil(err)
	err = json.Unmarshal(bodyBytes, &resParsed)
	suite.Nil(err)

	suite.Equal(200, res.Code)
	suite.Equal(genHash, resParsed.UploadSession.GenesisHash)
	suite.Equal(uint64(123), resParsed.UploadSession.FileSizeBytes)
	suite.Equal(models.SessionTypeBeta, resParsed.UploadSession.Type)
	suite.Equal(1, len(resParsed.BetaTreasureIndexes))
	suite.NotEqual(0, resParsed.Invoice.Cost)
	suite.NotEqual("", resParsed.Invoice.EthAddress)

	time.Sleep(50 * time.Millisecond) // Force it to wait for goroutine to excute.

	// TODO: fix waitForTransfer and uncomment it out in
	// actions/upload_sessions.go then uncomment out this test.
	//suite.True(mockWaitForTransfer.hasCalled)
	suite.Equal(oyster_utils.StringToAddress(resParsed.UploadSession.ETHAddrAlpha.String), mockWaitForTransfer.input_brokerAddr)
	suite.False(mockSendPrl.hasCalled)

	// TODO: fix waitForTransfer and uncomment it out in
	// actions/upload_sessions.go then uncomment out these tests.
	// verifyPaymentConfirmation(as, resParsed.ID)

	chunkData := models.GetSingleChunkData(oyster_utils.InProgressDir, genHash, models.MetaDataChunkIdx)

	expectedHash := oyster_utils.HashHex(genHash, sha256.New())

	suite.Equal(expectedHash, chunkData.Hash)

	brokerTx := []models.BrokerBrokerTransaction{}

	suite.DB.Where("genesis_hash = ?", genHash).All(&brokerTx)

	suite.Equal(1, len(brokerTx))
}

func (suite *ActionSuite) Test_UploadSessionsGetPaymentStatus_Paid() {
	//setup
	mockCheckPRLBalance := mockCheckPRLBalance{}
	EthWrapper = oyster_utils.Eth{
		CheckPRLBalance: mockCheckPRLBalance.checkPRLBalance,
	}

	genHash := oyster_utils.RandSeq(8, []rune("abcdef0123456789"))

	uploadSession1 := models.UploadSession{
		GenesisHash:   genHash,
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
	EthWrapper = oyster_utils.Eth{
		CheckPRLBalance: mockCheckPRLBalance.checkPRLBalance,
		SendPRL:         mockSendPrl.sendPrl,
	}

	genHash := oyster_utils.RandSeq(8, []rune("abcdef0123456789"))

	uploadSession1 := models.UploadSession{
		GenesisHash:   genHash,
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
	suite.Equal(oyster_utils.StringToAddress(uploadSession1.ETHAddrAlpha.String), mockCheckPRLBalance.input_addr)

	session := models.UploadSession{}
	suite.Nil(suite.DB.Find(&session, resParsed.ID))
	suite.Equal(models.PaymentStatusConfirmed, session.PaymentStatus)
}

func (suite *ActionSuite) Test_UploadSessionsGetPaymentStatus_NoConfirmAndCheckIncomplete() {
	//setup
	mockCheckPRLBalance := mockCheckPRLBalance{
		output_int: big.NewInt(0),
	}
	EthWrapper = oyster_utils.Eth{
		CheckPRLBalance: mockCheckPRLBalance.checkPRLBalance,
	}

	genHash := oyster_utils.RandSeq(8, []rune("abcdef0123456789"))

	uploadSession1 := models.UploadSession{
		GenesisHash:   genHash,
		FileSizeBytes: 123,
		NumChunks:     2,
		PaymentStatus: models.PaymentStatusInvoiced,
		ETHAddrAlpha:  nulls.NewString("alpha"),
	}

	resParsed := getPaymentStatus(uploadSession1, suite)

	suite.Equal("invoiced", resParsed.PaymentStatus)
	suite.True(mockCheckPRLBalance.hasCalled)
	suite.Equal(oyster_utils.StringToAddress(uploadSession1.ETHAddrAlpha.String), mockCheckPRLBalance.input_addr)

	session := models.UploadSession{}
	suite.Nil(suite.DB.Find(&session, resParsed.ID))
	suite.Equal(models.PaymentStatusInvoiced, session.PaymentStatus)
}

func (suite *ActionSuite) Test_UploadSessionsGetPaymentStatus_BetaConfirmed() {
	mockCheckPRLBalance := mockCheckPRLBalance{
		output_int: big.NewInt(10),
	}
	mockSendPrl := mockSendPrl{}
	EthWrapper = oyster_utils.Eth{
		CheckPRLBalance: mockCheckPRLBalance.checkPRLBalance,
		SendPRL:         mockSendPrl.sendPrl,
	}

	genHash := oyster_utils.RandSeq(8, []rune("abcdef0123456789"))

	uploadSession1 := models.UploadSession{
		Type:          models.SessionTypeBeta,
		GenesisHash:   genHash,
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

func getPaymentStatus(seededUploadSession models.UploadSession, suite *ActionSuite) paymentStatusCreateResV2 {
	seededUploadSession.StartUploadSession()

	session := models.UploadSession{}
	suite.Nil(suite.DB.Where("genesis_hash = ?", seededUploadSession.GenesisHash).First(&session))

	//execute method
	res := suite.JSON("/api/v2/upload-sessions/" + fmt.Sprint(session.ID)).Get()

	// Parse response
	resParsed := paymentStatusCreateResV2{}
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

func (v *mockSendPrl) sendPrl(msg oyster_utils.OysterCallMsg) bool {
	v.hasCalled = true
	v.input_msg = msg
	return v.output_bool
}

func (v *mockCheckPRLBalance) checkPRLBalance(addr common.Address) *big.Int {
	v.hasCalled = true
	v.input_addr = addr
	return v.output_int
}

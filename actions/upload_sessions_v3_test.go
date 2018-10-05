package actions

import (
	"encoding/json"
	"io/ioutil"
	"math/big"
	"time"

	oyster_utils "github.com/oysterprotocol/brokernode/utils"

	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
)

func (suite *ActionSuite) Test_UploadSessionsV3Create() {
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

	genHash := oyster_utils.RandSeq(8, []rune("abcdef0123456789"))

	res := suite.JSON("/api/v3/upload-sessions").Post(map[string]interface{}{
		"genesisHash":          genHash,
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
	suite.Equal(genHash, resParsed.UploadSession.GenesisHash)
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

	chunkData := models.GetSingleChunkData(oyster_utils.InProgressDir, genHash, int64(0))

	suite.Equal(genHash, chunkData.Hash)

	brokerTx := []models.BrokerBrokerTransaction{}

	suite.DB.Where("genesis_hash = ?", genHash).All(&brokerTx)

	suite.Equal(1, len(brokerTx))
}

func (suite *ActionSuite) Test_UploadSessionsV3CreateBeta() {
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

	genHash := oyster_utils.RandSeq(8, []rune("abcdef0123456789"))

	res := suite.JSON("/api/v3/upload-sessions/beta").Post(map[string]interface{}{
		"genesisHash":          genHash,
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
	suite.Equal(services.StringToAddress(resParsed.UploadSession.ETHAddrAlpha.String), mockWaitForTransfer.input_brokerAddr)
	suite.False(mockSendPrl.hasCalled)

	// TODO: fix waitForTransfer and uncomment it out in
	// actions/upload_sessions.go then uncomment out these tests.
	// verifyPaymentConfirmation(as, resParsed.ID)

	chunkData := models.GetSingleChunkData(oyster_utils.InProgressDir, genHash, int64(0))

	suite.Equal(genHash, chunkData.Hash)

	brokerTx := []models.BrokerBrokerTransaction{}

	suite.DB.Where("genesis_hash = ?", genHash).All(&brokerTx)

	suite.Equal(1, len(brokerTx))
}

func (suite *ActionSuite) Test_UploadSessionsV3GetPaymentStatus_Paid() {
	//setup
	mockCheckPRLBalance := mockCheckPRLBalance{}
	EthWrapper = services.Eth{
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

package actions

import (
	"encoding/json"
	"io/ioutil"

	"fmt"
	"github.com/oysterprotocol/brokernode/models"
)

func (as *ActionSuite) Test_UploadSessionsCreate() {
	res := as.JSON("/api/v2/upload-sessions").Post(map[string]interface{}{
		"genesisHash":          "genesisHashTest",
		"fileSizeBytes":        123,
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
}

func (as *ActionSuite) Test_UploadSessionsCreateBeta() {
	res := as.JSON("/api/v2/upload-sessions/beta").Post(map[string]interface{}{
		"genesisHash":          "genesisHashTest",
		"fileSizeBytes":        123,
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
}

func (as *ActionSuite) Test_UploadSessionsGetPaymentStatus_Paid() {
	//setup
	uploadSession1 := models.UploadSession{
		GenesisHash:   "genHash1",
		FileSizeBytes: 123,
		PaymentStatus: models.PaymentStatusPaid,
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

	as.Equal("paid", resParsed.PaymentStatus)
}

func (as *ActionSuite) Test_UploadSessionsGetPaymentStatus_Pending() {
	//setup
	uploadSession1 := models.UploadSession{
		GenesisHash:   "genHash1",
		FileSizeBytes: 123,
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

func (as *ActionSuite) Test_UploadSessionsGetPaymentStatus_Error() {
	//setup
	uploadSession1 := models.UploadSession{
		GenesisHash:   "genHash1",
		FileSizeBytes: 123,
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

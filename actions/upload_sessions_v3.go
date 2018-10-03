package actions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/nulls"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	oyster_utils "github.com/oysterprotocol/brokernode/utils"
	"github.com/pkg/errors"
	analytics "gopkg.in/segmentio/analytics-go.v3"
)

type UploadSessionResourceV3 struct {
	buffalo.Resource
}

// Request Response structs

type uploadSessionCreateReqV3 struct {
	GenesisHash          string         `json:"genesisHash"`
	NumChunks            int            `json:"numChunks"`
	FileSizeBytes        uint64         `json:"fileSizeBytes"` // This is Trytes instead of Byte
	BetaIP               string         `json:"betaIp"`
	StorageLengthInYears int            `json:"storageLengthInYears"`
	AlphaTreasureIndexes []int          `json:"alphaTreasureIndexes"`
	Invoice              models.Invoice `json:"invoice"`
	Version              uint32         `json:"version"`
}

type uploadSessionCreateResV3 struct {
	ID            string               `json:"id"`
	UploadSession models.UploadSession `json:"uploadSession"`
	BetaSessionID string               `json:"betaSessionId"`
	Invoice       models.Invoice       `json:"invoice"`
}

type uploadSessionCreateBetaResV3 struct {
	ID                  string               `json:"id"`
	UploadSession       models.UploadSession `json:"uploadSession"`
	BetaSessionID       string               `json:"betaSessionId"`
	Invoice             models.Invoice       `json:"invoice"`
	BetaTreasureIndexes []int                `json:"betaTreasureIndexes"`
}

type UploadSessionUpdateReqV3 struct {
	Chunks []models.ChunkReq `json:"chunks"`
}

type paymentStatusCreateResV3 struct {
	ID            string `json:"id"`
	PaymentStatus string `json:"paymentStatus"`
}

var NumChunksLimitV3 = -1 //unlimited

func init() {

}

// Create creates an upload session.
func (usr *UploadSessionResourceV3) Create(c buffalo.Context) error {

	if os.Getenv("DEPLOY_IN_PROGRESS") == "true" {
		err := errors.New("Deployment in progress.  Try again later")
		fmt.Println(err)
		c.Error(400, err)
		return err
	}

	if v, err := strconv.Atoi(os.Getenv("NUM_CHUNKS_LIMIT")); err == nil {
		NumChunksLimitV3 = v
	}

	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramUploadSessionResourceCreate, start)

	req := uploadSessionCreateReqV3{}
	if err := oyster_utils.ParseReqBody(c.Request(), &req); err != nil {
		err = fmt.Errorf("Invalid request, unable to parse request body  %v", err)
		c.Error(400, err)
		return err
	}

	if NumChunksLimitV3 != -1 && req.NumChunks > NumChunksLimitV3 {
		err := errors.New("This broker has a limit of " + fmt.Sprint(NumChunksLimitV3) + " file chunks.")
		fmt.Println(err)
		c.Error(400, err)
		return err
	}

	alphaEthAddr, privKey, _ := EthWrapper.GenerateEthAddr()

	// Start Alpha Session.
	alphaSession := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          req.GenesisHash,
		FileSizeBytes:        req.FileSizeBytes,
		NumChunks:            req.NumChunks,
		StorageLengthInYears: req.StorageLengthInYears,
		ETHAddrAlpha:         nulls.NewString(alphaEthAddr.Hex()),
		ETHPrivateKey:        privKey,
		Version:              req.Version,
	}

	defer oyster_utils.TimeTrack(time.Now(), "actions/upload_sessions: create_alpha_session", analytics.NewProperties().
		Set("id", alphaSession.ID).
		Set("genesis_hash", alphaSession.GenesisHash).
		Set("file_size_byes", alphaSession.FileSizeBytes).
		Set("num_chunks", alphaSession.NumChunks).
		Set("storage_years", alphaSession.StorageLengthInYears))

	if oyster_utils.DataMapStorageMode == oyster_utils.DataMapsInBadger {

		dbID := []string{oyster_utils.InProgressDir, alphaSession.GenesisHash, oyster_utils.HashDir}

		db := oyster_utils.GetOrInitUniqueBadgerDB(dbID)
		if db == nil {
			err := errors.New("error creating unique badger DB for hashes")
			oyster_utils.LogIfError(err, nil)
			c.Error(400, err)
			return err
		}
	}

	vErr, err := alphaSession.StartUploadSession()
	if err != nil || vErr.HasAny() {
		err = fmt.Errorf("StartUploadSession error: %v and validation error: %v", err, vErr)
		c.Error(400, err)
		return err
	}

	invoice := alphaSession.GetInvoice()

	// Mutates this because copying in golang sucks...
	req.Invoice = invoice

	req.AlphaTreasureIndexes = oyster_utils.GenerateInsertedIndexesForPearl(oyster_utils.ConvertToByte(req.FileSizeBytes))

	// Start Beta Session.
	var betaSessionID = ""
	var betaTreasureIndexes []int
	hasBeta := req.BetaIP != ""
	if hasBeta {
		betaReq, err := json.Marshal(req)
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			c.Error(400, err)
			return err
		}

		reqBetaBody := bytes.NewBuffer(betaReq)

		// Should we be hardcoding the port?
		betaURL := req.BetaIP + ":3000/api/v2/upload-sessions/beta"
		betaRes, err := http.Post(betaURL, "application/json", reqBetaBody)
		defer betaRes.Body.Close() // we need to close the connection
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			c.Error(400, err)
			return err
		}
		betaSessionRes := &uploadSessionCreateBetaResV3{}
		if err := oyster_utils.ParseResBody(betaRes, betaSessionRes); err != nil {
			err = fmt.Errorf("Unable to communicate with Beta node: %v", err)
			// This should consider as BadRequest since the client pick the beta node.
			c.Error(400, err)
			return err
		}

		betaSessionID = betaSessionRes.ID

		betaTreasureIndexes = betaSessionRes.BetaTreasureIndexes
		alphaSession.ETHAddrBeta = betaSessionRes.UploadSession.ETHAddrBeta
	}

	if err := models.DB.Save(&alphaSession); err != nil {
		oyster_utils.LogIfError(err, nil)
		c.Error(400, err)
		return err
	}

	models.NewBrokerBrokerTransaction(&alphaSession)

	if hasBeta {
		mergedIndexes, _ := oyster_utils.MergeIndexes(req.AlphaTreasureIndexes, betaTreasureIndexes,
			oyster_utils.FileSectorInChunkSize, req.NumChunks)

		if len(mergedIndexes) == 0 && oyster_utils.BrokerMode != oyster_utils.TestModeNoTreasure {
			err := errors.New("no indexes selected for treasure")
			fmt.Println(err)
			c.Error(400, err)
			return err
		}

		for {
			privateKeys, err := EthWrapper.GenerateKeys(len(mergedIndexes))
			if err != nil {
				err := errors.New("Could not generate eth keys: " + err.Error())
				fmt.Println(err)
				c.Error(400, err)
				return err
			}
			if len(mergedIndexes) != len(privateKeys) {
				err := errors.New("privateKeys and mergedIndexes should have the same length")
				oyster_utils.LogIfError(err, nil)
				c.Error(400, err)
				return err
			}
			// Update alpha treasure idx map.
			alphaSession.MakeTreasureIdxMap(mergedIndexes, privateKeys)

			treasureIndexes, _ := alphaSession.GetTreasureIndexes()

			if alphaSession.TreasureStatus == models.TreasureInDataMapPending &&
				alphaSession.TreasureIdxMap.Valid && alphaSession.TreasureIdxMap.String != "" &&
				len(treasureIndexes) == len(mergedIndexes) {
				models.DB.ValidateAndUpdate(&alphaSession)
				break
			}
		}
	}

	res := uploadSessionCreateResV3{
		UploadSession: alphaSession,
		ID:            alphaSession.ID.String(),
		BetaSessionID: betaSessionID,
		Invoice:       invoice,
	}
	//go waitForTransferAndNotifyBeta(
	//	res.UploadSession.ETHAddrAlpha.String, res.UploadSession.ETHAddrBeta.String, res.ID)

	return c.Render(200, r.JSON(res))
}

// Update uploads a chunk associated with an upload session.
func (usr *UploadSessionResourceV3) Update(c buffalo.Context) error {
	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramUploadSessionResourceUpdate, start)

	req := UploadSessionUpdateReqV3{}
	if err := oyster_utils.ParseReqBody(c.Request(), &req); err != nil {
		err = fmt.Errorf("Invalid request, unable to parse request body  %v", err)
		c.Error(400, err)
		return err
	}

	// Get session
	uploadSession := &models.UploadSession{}
	err := models.DB.Find(uploadSession, c.Param("id"))

	defer oyster_utils.TimeTrack(time.Now(), "actions/upload_sessions: updating_session", analytics.NewProperties().
		Set("id", uploadSession.ID).
		Set("genesis_hash", uploadSession.GenesisHash).
		Set("file_size_byes", uploadSession.FileSizeBytes).
		Set("num_chunks", uploadSession.NumChunks).
		Set("storage_years", uploadSession.StorageLengthInYears))

	if err != nil {
		oyster_utils.LogIfError(err, nil)
		c.Error(400, err)
		return err
	}
	if uploadSession == nil {
		err := errors.New("Error finding sessions")
		oyster_utils.LogIfError(err, nil)
		c.Error(400, err)
		return err
	}

	treasureIdxMap, err := uploadSession.GetTreasureIndexes()

	if oyster_utils.DataMapStorageMode == oyster_utils.DataMapsInBadger {
		dbID := []string{oyster_utils.InProgressDir, uploadSession.GenesisHash, oyster_utils.MessageDir}

		db := oyster_utils.GetOrInitUniqueBadgerDB(dbID)
		if db == nil {
			err := errors.New("error creating unique badger DB for messages")
			oyster_utils.LogIfError(err, nil)
			c.Error(400, err)
			return err
		}
	}

	// Update dMaps to have chunks async
	go func() {
		defer oyster_utils.TimeTrack(time.Now(), "actions/upload_sessions: async_datamap_updates", analytics.NewProperties().
			Set("id", uploadSession.ID).
			Set("genesis_hash", uploadSession.GenesisHash).
			Set("file_size_byes", uploadSession.FileSizeBytes).
			Set("num_chunks", uploadSession.NumChunks).
			Set("storage_years", uploadSession.StorageLengthInYears))

		models.ProcessAndStoreChunkData(req.Chunks, uploadSession.GenesisHash, treasureIdxMap,
			models.DataMapsTimeToLive)
	}()

	return c.Render(202, r.JSON(map[string]bool{"success": true}))
}

// CreateBeta creates an upload session on the beta broker.
func (usr *UploadSessionResourceV3) CreateBeta(c buffalo.Context) error {
	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramUploadSessionResourceCreateBeta, start)

	req := uploadSessionCreateReqV3{}
	if err := oyster_utils.ParseReqBody(c.Request(), &req); err != nil {
		err = fmt.Errorf("Invalid request, unable to parse request body  %v", err)
		c.Error(400, err)
		return err
	}

	betaTreasureIndexes := oyster_utils.GenerateInsertedIndexesForPearl(oyster_utils.ConvertToByte(req.FileSizeBytes))

	// Generates ETH address.
	betaEthAddr, privKey, _ := EthWrapper.GenerateEthAddr()

	u := models.UploadSession{
		Type:                 models.SessionTypeBeta,
		GenesisHash:          req.GenesisHash,
		NumChunks:            req.NumChunks,
		FileSizeBytes:        req.FileSizeBytes,
		StorageLengthInYears: req.StorageLengthInYears,
		TotalCost:            req.Invoice.Cost,
		ETHAddrAlpha:         req.Invoice.EthAddress,
		ETHAddrBeta:          nulls.NewString(betaEthAddr.Hex()),
		ETHPrivateKey:        privKey,
		Version:              req.Version,
	}

	defer oyster_utils.TimeTrack(time.Now(), "actions/upload_sessions: create_beta_session", analytics.NewProperties().
		Set("id", u.ID).
		Set("genesis_hash", u.GenesisHash).
		Set("file_size_byes", u.FileSizeBytes).
		Set("num_chunks", u.NumChunks).
		Set("storage_years", u.StorageLengthInYears))

	if oyster_utils.DataMapStorageMode == oyster_utils.DataMapsInBadger {
		dbID := []string{oyster_utils.InProgressDir, u.GenesisHash, oyster_utils.HashDir}

		db := oyster_utils.GetOrInitUniqueBadgerDB(dbID)
		if db == nil {
			err := errors.New("error creating unique badger DB for hashes")
			oyster_utils.LogIfError(err, nil)
			c.Error(400, err)
			return err
		}
	}

	vErr, err := u.StartUploadSession()

	if err != nil || vErr.HasAny() {
		err = fmt.Errorf("Can't startUploadSession with validation error: %v and err: %v", vErr, err)
		c.Error(400, err)
		return err
	}

	if len(vErr.Errors) > 0 {
		c.Render(422, r.JSON(vErr.Errors))
		return err
	}

	mergedIndexes, err := oyster_utils.MergeIndexes(req.AlphaTreasureIndexes, betaTreasureIndexes,
		oyster_utils.FileSectorInChunkSize, req.NumChunks)

	if len(mergedIndexes) == 0 && oyster_utils.BrokerMode != oyster_utils.TestModeNoTreasure {
		err := errors.New("no indexes selected for treasure")
		fmt.Println(err)
		c.Error(400, err)
		return err
	}

	if err != nil {
		fmt.Println(err)
		c.Error(400, err)
		return err
	}
	for {
		privateKeys, err := EthWrapper.GenerateKeys(len(mergedIndexes))
		if err != nil {
			err := errors.New("Could not generate eth keys: " + err.Error())
			fmt.Println(err)
			c.Error(400, err)
			return err
		}
		if len(mergedIndexes) != len(privateKeys) {
			err := errors.New("privateKeys and mergedIndexes should have the same length")
			fmt.Println(err)
			c.Error(400, err)
			return err
		}
		u.MakeTreasureIdxMap(mergedIndexes, privateKeys)

		treasureIndexes, err := u.GetTreasureIndexes()

		if u.TreasureStatus == models.TreasureInDataMapPending &&
			u.TreasureIdxMap.Valid && u.TreasureIdxMap.String != "" &&
			len(treasureIndexes) == len(mergedIndexes) {
			models.DB.ValidateAndUpdate(&u)
			break
		}
	}

	models.NewBrokerBrokerTransaction(&u)

	res := uploadSessionCreateBetaResV3{
		UploadSession:       u,
		ID:                  u.ID.String(),
		Invoice:             u.GetInvoice(),
		BetaTreasureIndexes: betaTreasureIndexes,
	}
	//go waitForTransferAndNotifyBeta(
	//	res.UploadSession.ETHAddrAlpha.String, res.UploadSession.ETHAddrBeta.String, res.ID)

	return c.Render(200, r.JSON(res))
}

func (usr *UploadSessionResourceV3) GetPaymentStatus(c buffalo.Context) error {
	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramUploadSessionResourceGetPaymentStatus, start)

	session := models.UploadSession{}
	err := models.DB.Find(&session, c.Param("id"))

	if err != nil {
		c.Error(400, err)
		oyster_utils.LogIfError(err, nil)
		return err
	}
	if (session == models.UploadSession{}) {
		err := errors.New("Did not find session that matched id" + c.Param("id"))
		oyster_utils.LogIfError(err, nil)
		c.Error(400, err)
		return err
	}

	// Force to check the status
	if session.PaymentStatus != models.PaymentStatusConfirmed {
		balance := EthWrapper.CheckPRLBalance(services.StringToAddress(session.ETHAddrAlpha.String))
		if balance.Int64() > 0 {
			previousPaymentStatus := session.PaymentStatus
			session.PaymentStatus = models.PaymentStatusConfirmed
			err = models.DB.Save(&session)
			if err != nil {
				session.PaymentStatus = previousPaymentStatus
			} else {
				models.SetBrokerTransactionToPaid(session)
			}
		}
	}

	res := paymentStatusCreateResV3{
		ID:            session.ID.String(),
		PaymentStatus: session.GetPaymentStatus(),
	}

	return c.Render(200, r.JSON(res))
}

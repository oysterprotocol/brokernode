package actions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"os"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/nulls"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"gopkg.in/segmentio/analytics-go.v3"

	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type UploadSessionResource struct {
	buffalo.Resource
}

// Request Response structs

type uploadSessionCreateReq struct {
	GenesisHash          string         `json:"genesisHash"`
	NumChunks            int            `json:"numChunks"`
	FileSizeBytes        uint64         `json:"fileSizeBytes"` // This is Trytes instead of Byte
	BetaIP               string         `json:"betaIp"`
	StorageLengthInYears int            `json:"storageLengthInYears"`
	AlphaTreasureIndexes []int          `json:"alphaTreasureIndexes"`
	Invoice              models.Invoice `json:"invoice"`
	Version              uint32         `json:"version"`
}

type uploadSessionCreateRes struct {
	ID            string               `json:"id"`
	UploadSession models.UploadSession `json:"uploadSession"`
	BetaSessionID string               `json:"betaSessionId"`
	Invoice       models.Invoice       `json:"invoice"`
}

type uploadSessionCreateBetaRes struct {
	ID                  string               `json:"id"`
	UploadSession       models.UploadSession `json:"uploadSession"`
	BetaSessionID       string               `json:"betaSessionId"`
	Invoice             models.Invoice       `json:"invoice"`
	BetaTreasureIndexes []int                `json:"betaTreasureIndexes"`
}

type chunkReq struct {
	Idx  int    `json:"idx"`
	Data string `json:"data"`
	Hash string `json:"hash"` // This is GenesisHash.
}

type UploadSessionUpdateReq struct {
	Chunks []chunkReq `json:"chunks"`
}

type paymentStatusCreateRes struct {
	ID            string `json:"id"`
	PaymentStatus string `json:"paymentStatus"`
}

const SQL_BATCH_SIZE = 10

var NumChunksLimit = -1 //unlimited

func init() {
	if v, err := strconv.Atoi(os.Getenv("NUM_CHUNKS_LIMIT")); err == nil {
		NumChunksLimit = v
	}
}

// Create creates an upload session.
func (usr *UploadSessionResource) Create(c buffalo.Context) error {
	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramUploadSessionResourceCreate, start)

	req := uploadSessionCreateReq{}
	oyster_utils.ParseReqBody(c.Request(), &req)

	if NumChunksLimit != -1 && req.NumChunks > NumChunksLimit {
		err := errors.New("This broker has a limit of " + fmt.Sprint(NumChunksLimit) + " file chunks.")
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

	vErr, err := alphaSession.StartUploadSession()
	if err != nil {
		fmt.Println(err)
		return err
	}

	invoice := alphaSession.GetInvoice()

	// Mutates this because copying in golang sucks...
	req.Invoice = invoice

	req.AlphaTreasureIndexes = oyster_utils.GenerateInsertedIndexesForPearl(oyster_utils.ConvertToByte(req.FileSizeBytes))

	// Start Beta Session.
	var betaSessionID = ""
	var betaTreasureIndexes []int
	if req.BetaIP != "" {
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
		betaSessionRes := &uploadSessionCreateBetaRes{}
		oyster_utils.ParseResBody(betaRes, betaSessionRes)
		betaSessionID = betaSessionRes.ID

		betaTreasureIndexes = betaSessionRes.BetaTreasureIndexes
		alphaSession.ETHAddrBeta = betaSessionRes.UploadSession.ETHAddrBeta
	}

	err = models.DB.Save(&alphaSession)
	if err != nil {
		oyster_utils.LogIfError(err, nil)
		c.Error(400, err)
		return err
	}

	models.NewBrokerBrokerTransaction(&alphaSession)

	mergedIndexes, _ := oyster_utils.MergeIndexes(req.AlphaTreasureIndexes, betaTreasureIndexes, oyster_utils.FileSectorInChunkSize, req.NumChunks)
	if err != nil {
		// not doing error handling here, relying on beta to throw the error since returning
		// an error here breaks the unit tests
		fmt.Println(err)
	}

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

	if len(vErr.Errors) > 0 {
		c.Render(422, r.JSON(vErr.Errors))
		return err
	}

	res := uploadSessionCreateRes{
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
func (usr *UploadSessionResource) Update(c buffalo.Context) error {
	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramUploadSessionResourceUpdate, start)

	req := UploadSessionUpdateReq{}
	oyster_utils.ParseReqBody(c.Request(), &req)

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

	// Update dMaps to have chunks async
	go func() {
		defer oyster_utils.TimeTrack(time.Now(), "actions/upload_sessions: async_datamap_updates", analytics.NewProperties().
			Set("id", uploadSession.ID).
			Set("genesis_hash", uploadSession.GenesisHash).
			Set("file_size_byes", uploadSession.FileSizeBytes).
			Set("num_chunks", uploadSession.NumChunks).
			Set("storage_years", uploadSession.StorageLengthInYears))

		ProcessAndStoreChunkData(req.Chunks, uploadSession.GenesisHash, treasureIdxMap)
	}()

	return c.Render(202, r.JSON(map[string]bool{"success": true}))
}

// CreateBeta creates an upload session on the beta broker.
func (usr *UploadSessionResource) CreateBeta(c buffalo.Context) error {
	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramUploadSessionResourceCreateBeta, start)

	req := uploadSessionCreateReq{}
	oyster_utils.ParseReqBody(c.Request(), &req)

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

	vErr, err := u.StartUploadSession()

	if err != nil {
		fmt.Println(err)
		c.Error(400, err)
		return err
	}

	if len(vErr.Errors) > 0 {
		c.Render(422, r.JSON(vErr.Errors))
		return err
	}

	models.NewBrokerBrokerTransaction(&u)

	mergedIndexes, err := oyster_utils.MergeIndexes(req.AlphaTreasureIndexes, betaTreasureIndexes, oyster_utils.FileSectorInChunkSize, req.NumChunks)
	if err != nil {
		fmt.Println(err)
		c.Error(400, err)
		return err
	}
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
	u.MakeTreasureIdxMap(mergedIndexes, privateKeys)

	res := uploadSessionCreateBetaRes{
		UploadSession:       u,
		ID:                  u.ID.String(),
		Invoice:             u.GetInvoice(),
		BetaTreasureIndexes: betaTreasureIndexes,
	}
	//go waitForTransferAndNotifyBeta(
	//	res.UploadSession.ETHAddrAlpha.String, res.UploadSession.ETHAddrBeta.String, res.ID)

	return c.Render(200, r.JSON(res))
}

func (usr *UploadSessionResource) GetPaymentStatus(c buffalo.Context) error {
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

	res := paymentStatusCreateRes{
		ID:            session.ID.String(),
		PaymentStatus: session.GetPaymentStatus(),
	}

	return c.Render(200, r.JSON(res))
}

/*ProcessAndStoreChunkData process the request chunks and store them into data map. Public for unit test.*/
func ProcessAndStoreChunkData(chunks []chunkReq, genesisHash string, treasureIdxMap []int) {
	chunksMap, minChunkIdx, maxChunkIdx := convertToSQLKeyedMapForChunks(chunks, genesisHash, treasureIdxMap)

	// Query data_maps to get models.DataMap based on require chunk.
	var dms []models.DataMap
	//rawQuery := fmt.Sprintf("SELECT * FROM data_maps WHERE %s", strings.Join(sqlWhereClosures, " OR "))
	err := models.DB.RawQuery(
		"SELECT * FROM data_maps WHERE genesis_hash = ? AND chunk_idx >= ? AND chunk_idx <= ?",
		genesisHash, minChunkIdx, maxChunkIdx).All(&dms)
	oyster_utils.LogIfError(err, nil)

	dmsMap := convertToSQLKeyedMapForDataMap(dms)

	// Create Update operation for data_maps table.
	dbOperation, _ := oyster_utils.CreateDbUpdateOperation(&models.DataMap{})
	batchSetKvMap := services.KVPairs{} // Store chunk.Data into KVStore
	var updatedDms []string
	for key, chunk := range chunksMap {
		dm, hasKey := dmsMap[key]
		if !hasKey {
			continue
		}

		if dm.MsgID == "" {
			oyster_utils.LogIfError(errors.New("DataMap was not stored into data_maps table and MsgID is empty"), nil)
			break
		}

		if chunk.Hash == dm.GenesisHash {
			if services.IsKvStoreEnabled() {
				batchSetKvMap[dm.MsgID] = chunk.Data
				dm.Message = "" // Remove previous Message data.
				dm.MsgStatus = models.MsgStatusUploadedHaveNotEncoded
			} else {
				// TODO:pzhao, remove this and this should not be called.
				message, err := oyster_utils.ChunkMessageToTrytesWithStopper(chunk.Data)
				if err != nil {
					oyster_utils.LogIfError(err, nil)
					panic(err.Error())
				}
				dm.Message = string(message)
				dm.MsgStatus = models.MsgStatusUploadedNoNeedEncode
			}

			if oyster_utils.BrokerMode == oyster_utils.TestModeNoTreasure {
				dm.Status = models.Unassigned
			}
			vErr, _ := dm.Validate(nil)
			oyster_utils.LogIfValidationError("Unable to create data_maps for batch insertion.", vErr, nil)
			if !vErr.HasAny() {
				updatedDms = append(updatedDms, fmt.Sprintf("(%s)", dbOperation.GetUpdatedValue(dm)))
			}
		}
	}

	err = batchUpsertDataMaps(updatedDms, dbOperation.GetColumns())
	// Save Message field into KVStore
	if err == nil && services.IsKvStoreEnabled() {
		services.BatchSet(&batchSetKvMap)
	}
}

// convertToSQLKeyedMapForChunks converts chunkReq into sql keyed maps. Return minChunkId and maxChunkId.
func convertToSQLKeyedMapForChunks(chunks []chunkReq, genesisHash string, treasureIdxMap []int) (map[string]chunkReq, int, int) {
	chunksMap := make(map[string]chunkReq)
	minChunkIdx := 0
	maxChunkIdx := 0

	for _, chunk := range chunks {
		var chunkIdx int
		if oyster_utils.BrokerMode == oyster_utils.TestModeNoTreasure {
			chunkIdx = chunk.Idx
		} else {
			chunkIdx = oyster_utils.TransformIndexWithBuriedIndexes(chunk.Idx, treasureIdxMap)
		}

		key := sqlWhereForGenesisHashAndChunkIdx(genesisHash, chunkIdx)
		chunksMap[key] = chunk
		minChunkIdx = oyster_utils.IntMin(minChunkIdx, chunkIdx)
		maxChunkIdx = oyster_utils.IntMax(maxChunkIdx, chunkIdx)
	}
	return chunksMap, minChunkIdx, maxChunkIdx
}

// convertToSQLKeyedMapForDataMap converts dataMaps into sql keyed maps. Remove any duplicate keyed.
func convertToSQLKeyedMapForDataMap(dataMaps []models.DataMap) map[string]models.DataMap {
	dmsMap := make(map[string]models.DataMap)
	for _, dm := range dataMaps {
		key := sqlWhereForGenesisHashAndChunkIdx(dm.GenesisHash, dm.ChunkIdx)
		// Only use the first one
		if _, hasKey := dmsMap[key]; !hasKey {
			dmsMap[key] = dm
		}
	}
	return dmsMap
}

// batchUpsertDataMaps update data_maps table to override the values of column: message, status, updated_at, msg_status.
func batchUpsertDataMaps(serializedDataMapValue []string, serializedColumnNames string) error {
	numOfBatchRequest := int(math.Ceil(float64(len(serializedDataMapValue)) / float64(SQL_BATCH_SIZE)))

	// Batch Update data_maps table.
	remainder := len(serializedDataMapValue)
	for i := 0; i < numOfBatchRequest; i++ {
		lower := i * SQL_BATCH_SIZE
		upper := i*SQL_BATCH_SIZE + int(math.Min(float64(remainder), SQL_BATCH_SIZE))

		sectionUpdatedDms := serializedDataMapValue[lower:upper]

		// Do an insert operation and dup by primary key.
		rawQuery := fmt.Sprintf(`INSERT INTO data_maps (%s) VALUES %s ON DUPLICATE KEY 
		 	UPDATE message = VALUES(message), status = VALUES(status), updated_at = VALUES(updated_at), msg_status = VALUES(msg_status)`,
			serializedColumnNames, strings.Join(sectionUpdatedDms, ","))

		err := models.DB.RawQuery(rawQuery).All(&[]models.DataMap{})
		for err != nil {
			time.Sleep(300 * time.Millisecond)
			err = models.DB.RawQuery(rawQuery).All(&[]models.DataMap{})
		}

		remainder = remainder - SQL_BATCH_SIZE
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			return err
		}
	}
	return nil
}

func sqlWhereForGenesisHashAndChunkIdx(genesisHash string, chunkIdx int) string {
	return fmt.Sprintf("(genesis_hash = '%s' AND chunk_idx = %d)", genesisHash, chunkIdx)
}

func waitForTransferAndNotifyBeta(alphaEthAddr string, betaEthAddr string, uploadSessionId string) {

	if oyster_utils.BrokerMode != oyster_utils.ProdMode {
		return
	}

	transferAddr := services.StringToAddress(alphaEthAddr)
	balance, err := EthWrapper.WaitForTransfer(transferAddr, "prl")

	paymentStatus := models.PaymentStatusConfirmed
	if err != nil {
		paymentStatus = models.PaymentStatusError
	}

	session := models.UploadSession{}
	if err := models.DB.Find(&session, uploadSessionId); err != nil {
		oyster_utils.LogIfError(err, nil)
		return
	}

	if session.PaymentStatus != models.PaymentStatusConfirmed {
		session.PaymentStatus = paymentStatus
	}
	if err := models.DB.Save(&session); err != nil {
		oyster_utils.LogIfError(err, nil)
		return
	}

	// Alpha send half of it to Beta
	checkAndSendHalfPrlToBeta(session, balance)
}

func checkAndSendHalfPrlToBeta(session models.UploadSession, balance *big.Int) {
	if session.Type != models.SessionTypeAlpha ||
		session.PaymentStatus != models.PaymentStatusConfirmed ||
		session.ETHAddrBeta.String == "" {
		return
	}

	betaAddr := services.StringToAddress(session.ETHAddrBeta.String)
	betaBalance := EthWrapper.CheckPRLBalance(betaAddr)
	if betaBalance.Int64() > 0 {
		return
	}

	var splitAmount big.Int
	splitAmount.Set(balance)
	splitAmount.Div(balance, big.NewInt(2))
	callMsg := services.OysterCallMsg{
		From:   services.StringToAddress(session.ETHAddrAlpha.String),
		To:     betaAddr,
		Amount: splitAmount,
	}
	EthWrapper.SendPRL(callMsg)
}

package actions_v3

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/nulls"
	"github.com/oysterprotocol/brokernode/actions/utils"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/utils"
	"github.com/pkg/errors"
)

const (
	BatchSize = 25
)

type UploadSessionResourceV3 struct {
	buffalo.Resource
}

type uploadSessionUpdateReqV3 struct {
	Chunks []models.ChunkReq `json:"chunks"`
}

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

type uploadSessionCreateBetaResV3 struct {
	ID              string `json:"id"`
	TreasureIndexes []int  `json:"treasureIndexes"`
	ETHAddr         string `json:"ethAddr"`
}

type uploadSessionCreateResV3 struct {
	ID            string         `json:"id"`
	BetaSessionID string         `json:"betaSessionId"`
	BatchSize     int            `json:"batchSize"`
	Invoice       models.Invoice `json:"invoice"`
}

/*uploadSessionConfig represents the general configuration that other client could understand.*/
type uploadSessionConfig struct {
	BatchSize        int    `json:"batchSize"`        // Represent each data contain not more number of field.
	FileSizeBytes    uint64 `json:"fileSizeBytes"`    // Represent the total file size.
	NumChunks        int    `json:"numChunks"`        // Represent total number of chunks.
	ReverseIteration bool   `json:"reverseIteration"` // Represent whether iterate it from the beginning to end or end to the beginning.
}

var NumChunksLimit = -1 //unlimited

func init() {
	if v, err := strconv.Atoi(os.Getenv("NUM_CHUNKS_LIMIT")); err == nil {
		NumChunksLimit = v
	}
}

// Update uploads a chunk associated with an upload session.
func (usr *UploadSessionResourceV3) Update(c buffalo.Context) error {
	req, err := validateAndGetUpdateReq(c)
	if err != nil {
		return c.Error(400, err)
	}

	uploadSession := &models.UploadSession{}
	if err = models.DB.Find(uploadSession, c.Param("id")); err != nil {
		return c.Error(500, oyster_utils.LogIfError(err, nil))
	}
	if uploadSession == nil {
		return c.Error(400, fmt.Errorf("Error in finding session for id %v", c.Param("id")))
	}

	if uploadSession.StorageMethod != models.StorageMethodS3 {
		return c.Error(400, errors.New("Using the wrong endpoint. This endpoint is for V3 only"))
	}

	isReverseIteration := uploadSession.Type == models.SessionTypeBeta
	objectKey := oyster_utils.GetObjectKeyForData(uploadSession.GenesisHash, req.Chunks[0].Idx, uploadSession.NumChunks, isReverseIteration, BatchSize)

	var data []byte
	if data, err = json.Marshal(req.Chunks); err != nil {
		return c.Error(500, oyster_utils.LogIfError(fmt.Errorf("Unable to marshal ChunkReq to JSON with err %v", err), nil))
	}
	if err = setDefaultBucketObject(objectKey, string(data)); err != nil {
		return c.Error(500, oyster_utils.LogIfError(fmt.Errorf("Unable to store data to S3 with err: %v", err), nil))
	}

	return c.Render(202, actions_utils.Render.JSON(map[string]bool{"success": true}))
}

/* Create endpoint. */
func (usr *UploadSessionResourceV3) Create(c buffalo.Context) error {
	req, err := validateAndGetCreateReq(c)
	if err != nil {
		return c.Error(400, err)
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
		StorageMethod:        models.StorageMethodS3,
	}

	if vErr, err := alphaSession.StartUploadSession(); err != nil || vErr.HasAny() {
		return c.Error(400, fmt.Errorf("StartUploadSession error: %v and validation error: %v", err, vErr))
	}

	invoice := alphaSession.GetInvoice()

	// Mutates this because copying in golang sucks...
	req.Invoice = invoice
	req.AlphaTreasureIndexes = oyster_utils.GenerateInsertedIndexesForPearl(oyster_utils.ConvertToByte(req.FileSizeBytes))

	hasBeta := req.BetaIP != ""
	var betaSessionID = ""
	if hasBeta {
		betaSessionRes, err := sendBetaWithUploadRequest(req)
		if err != nil {
			return c.Error(400, err)
		}

		betaSessionID = betaSessionRes.ID
		alphaSession.ETHAddrBeta = nulls.NewString(betaSessionRes.ETHAddr)

		if err := saveTreasureMapForS3(&alphaSession, req.AlphaTreasureIndexes, betaSessionRes.TreasureIndexes); err != nil {
			return c.Error(500, err)
		}

		if err := models.DB.Save(&alphaSession); err != nil {
			return c.Error(400, oyster_utils.LogIfError(err, nil))
		}
	}

	if err := saveConfigForS3(alphaSession); err != nil {
		return c.Error(500, err)
	}

	models.NewBrokerBrokerTransaction(&alphaSession)

	res := uploadSessionCreateResV3{
		ID:            alphaSession.ID.String(),
		BetaSessionID: betaSessionID,
		BatchSize:     BatchSize,
		Invoice:       invoice,
	}

	return c.Render(200, actions_utils.Render.JSON(res))
}

/* CreateBeta endpoint. */
func (usr *UploadSessionResourceV3) CreateBeta(c buffalo.Context) error {
	req, err := validateAndGetCreateReq(c)
	if err != nil {
		return err
	}

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
		StorageMethod:        models.StorageMethodS3,
	}

	if vErr, err := u.StartUploadSession(); err != nil || vErr.HasAny() {
		return c.Error(400, fmt.Errorf("Can't startUploadSession with validation error: %v and err: %v", vErr, err))
	}

	betaTreasureIndexes := oyster_utils.GenerateInsertedIndexesForPearl(oyster_utils.ConvertToByte(req.FileSizeBytes))
	if err := saveTreasureMapForS3(&u, req.AlphaTreasureIndexes, betaTreasureIndexes); err != nil {
		return c.Error(500, err)
	}

	if err := models.DB.Save(&u); err != nil {
		return c.Error(500, err)
	}

	if err := saveConfigForS3(u); err != nil {
		return c.Error(500, err)
	}

	models.NewBrokerBrokerTransaction(&u)

	res := uploadSessionCreateBetaResV3{
		ID:              u.ID.String(),
		TreasureIndexes: betaTreasureIndexes,
		ETHAddr:         u.ETHAddrBeta.String,
	}

	return c.Render(200, actions_utils.Render.JSON(res))
}

func validateAndGetCreateReq(c buffalo.Context) (uploadSessionCreateReqV3, error) {
	req := uploadSessionCreateReqV3{}
	if err := oyster_utils.ParseReqBody(c.Request(), &req); err != nil {
		return req, fmt.Errorf("Invalid request, unable to parse request body: %v", err)
	}

	if NumChunksLimit != -1 && req.NumChunks > NumChunksLimit {
		return req, errors.New("This broker has a limit of " + fmt.Sprint(NumChunksLimit) + " file chunks.")
	}
	return req, nil
}

func validateAndGetUpdateReq(c buffalo.Context) (uploadSessionUpdateReqV3, error) {
	req := uploadSessionUpdateReqV3{}
	if err := oyster_utils.ParseReqBody(c.Request(), &req); err != nil {
		return req, fmt.Errorf("Invalid request, unable to parse request body: %v", err)
	}

	if len(req.Chunks) > BatchSize {
		return req, fmt.Errorf("Except chunks to be in a batch of size %v", BatchSize)
	}

	sort.Sort(models.ChunkReqs(req.Chunks))
	startValue := req.Chunks[0].Idx - 1
	isIDUniqueIncrease := true
	for _, chunk := range req.Chunks {
		if startValue != chunk.Idx-1 {
			isIDUniqueIncrease = false
			break
		}
		startValue = chunk.Idx
	}
	if !isIDUniqueIncrease {
		return req, errors.New("Provided Id should be consecutive")
	}
	return req, nil
}

func sendBetaWithUploadRequest(req uploadSessionCreateReqV3) (uploadSessionCreateBetaResV3, error) {
	betaSessionRes := uploadSessionCreateBetaResV3{}
	betaURL := req.BetaIP + ":3000/api/v3/upload-sessions/beta"
	err := oyster_utils.SendHttpReq(betaURL, req, betaSessionRes)
	return betaSessionRes, err
}

/*saveTreasureMapForS3 saves treasure keys as JSON format into S3.*/
func saveTreasureMapForS3(u *models.UploadSession, treasureIndexA []int, treasureIndexB []int) error {
	mergedIndexes, err := oyster_utils.MergeIndexes(treasureIndexA, treasureIndexB,
		oyster_utils.FileSectorInChunkSize, u.NumChunks)
	if err != nil {
		return err
	}

	if len(mergedIndexes) == 0 {
		if oyster_utils.BrokerMode != oyster_utils.TestModeNoTreasure {
			return oyster_utils.LogIfError(errors.New("no indexes selected for treasure"), nil)
		}
		return nil
	}

	for {
		privateKeys, err := EthWrapper.GenerateKeys(len(mergedIndexes))
		if err != nil {
			return oyster_utils.LogIfError(errors.New("Could not generate eth keys: "+err.Error()), nil)
		}
		if len(mergedIndexes) != len(privateKeys) {
			return oyster_utils.LogIfError(errors.New("privateKeys and mergedIndexes should have the same length"), nil)
		}
		// Update treasureId
		u.MakeTreasureIdxMap(mergedIndexes, privateKeys)

		// Verify that MakeTreasureIdxMap is correct. Otherwise, regenerate it again.
		treasureIndexes, _ := alphaSession.GetTreasureIndexes()
		if alphaSession.TreasureStatus == models.TreasureInDataMapPending &&
			alphaSession.TreasureIdxMap.Valid && alphaSession.TreasureIdxMap.String != "" &&
			len(treasureIndexes) == len(mergedIndexes) {
			break
		}
	}

	return setDefaultBucketObject(oyster_utils.GetObjectKeyForTreasure(u.GenesisHash), u.TreasureIdxMap.String)
}

/*saveConfigForS3 saves uploadSession config to S3 endpoint so that Lamdba function could read it.*/
func saveConfigForS3(u models.UploadSession) error {
	config := uploadSessionConfig{
		BatchSize:        BatchSize,
		FileSizeBytes:    u.FileSizeBytes,
		NumChunks:        u.NumChunks,
		ReverseIteration: u.Type == models.SessionTypeBeta,
	}
	data, err := json.Marshal(config)
	if err != nil {
		return oyster_utils.LogIfError(err, nil)
	}
	return setDefaultBucketObject(oyster_utils.GetObjectKeyForConfig(u.GenesisHash), string(data))
}

package actions

import (
	"bytes"
	"encoding/json"
	"net/http"

	raven "github.com/getsentry/raven-go"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/nulls"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"github.com/pkg/errors"
)

type UploadSessionResource struct {
	buffalo.Resource
}

// Request Response structs

type uploadSessionCreateReq struct {
	GenesisHash          string         `json:"genesisHash"`
	FileSizeBytes        int            `json:"fileSizeBytes"`
	BetaIP               string         `json:"betaIp"`
	StorageLengthInYears int            `json:"storageLengthInYears"`
	AlphaTreasureIndexes []int          `json:"alphaTreasureIndexes"`
	Invoice              models.Invoice `json:"invoice"`
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
	Hash string `json:"hash"`
}

type UploadSessionUpdateReq struct {
	Chunks []chunkReq `json:"chunks"`
}

type paymentStatusCreateRes struct {
	ID            string `json:"id"`
	PaymentStatus string `json:"paymentStatus"`
}

// Create creates an upload session.
func (usr *UploadSessionResource) Create(c buffalo.Context) error {
	req := uploadSessionCreateReq{}
	oyster_utils.ParseReqBody(c.Request(), &req)

	// Is this really what golang wants me to do do namespace a function?
	eth := services.Eth{}
	alphaEthAddr, privKey, _ := eth.GenerateEthAddr()

	// Start Alpha Session.
	alphaSession := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          req.GenesisHash,
		FileSizeBytes:        req.FileSizeBytes,
		StorageLengthInYears: req.StorageLengthInYears,
		ETHAddrAlpha:         nulls.NewString(alphaEthAddr),
		ETHPrivateKey:        privKey,
	}
	vErr, err := alphaSession.StartUploadSession()
	if err != nil {
		return err
	}

	invoice := alphaSession.GetInvoice()

	// Mutates this because copying in golang sucks...
	req.Invoice = invoice
	// TODO(philip): req.AlphaBuriedIndexes

	// Start Beta Session.

	req.AlphaTreasureIndexes = oyster_utils.GenerateInsertedIndexesForPearl(req.FileSizeBytes)
	var betaSessionID = ""
	var betaTreasureIndexes []int
	if req.BetaIP != "" {
		betaReq, err := json.Marshal(req)
		if err != nil {
			c.Render(400, r.JSON(map[string]string{"Error starting Beta": err.Error()}))
			return err
		}

		reqBetaBody := bytes.NewBuffer(betaReq)

		// Should we be hardcoding the port?
		betaURL := req.BetaIP + ":3000/api/v2/upload-sessions/beta"
		betaRes, err := http.Post(betaURL, "application/json", reqBetaBody)

		if err != nil {
			c.Render(400, r.JSON(map[string]string{"Error starting Beta": err.Error()}))
			return err
		}
		betaSessionRes := &uploadSessionCreateBetaRes{}
		oyster_utils.ParseResBody(betaRes, betaSessionRes)
		betaSessionID = betaSessionRes.ID
		betaTreasureIndexes = betaSessionRes.BetaTreasureIndexes
	}

	// Update alpha treasure idx map.
	alphaSession.TreasureIdxMap = oyster_utils.GetTreasureIdxMap(req.AlphaTreasureIndexes, betaTreasureIndexes)
	err = models.DB.Save(&alphaSession)
	if err != nil {
		return err
	}

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
	return c.Render(200, r.JSON(res))
}

// Update uploads a chunk associated with an upload session.
func (usr *UploadSessionResource) Update(c buffalo.Context) error {

	req := UploadSessionUpdateReq{}
	oyster_utils.ParseReqBody(c.Request(), &req)

	// Get session
	uploadSession := &models.UploadSession{}
	err := models.DB.Find(uploadSession, c.Param("id"))
	if err != nil || uploadSession == nil {
		c.Render(400, r.JSON(map[string]string{"Error finding session": errors.WithStack(err).Error()}))
		return err
	}

	// Update dMaps to have chunks async
	go func() {
		// Map over chunks from request
		// TODO: Batch processing DB upserts.
		dMaps := make([]models.DataMap, len(req.Chunks))
		for i, chunk := range req.Chunks {
			// Fetch DataMap
			dm := models.DataMap{}
			err := models.DB.RawQuery(
				"SELECT * from data_maps WHERE genesis_hash = ? AND chunk_idx = ?", uploadSession.GenesisHash, chunk.Idx).First(&dm)

			if err != nil {
				raven.CaptureError(err, nil)
			}

			// Simple check if hashes match.
			if chunk.Hash == dm.GenesisHash {
				// Updates dmap in DB.
				dm.Message = chunk.Data
				models.DB.ValidateAndSave(&dm)
			}

			dMaps[i] = dm
		}
	}()

	return c.Render(202, r.JSON(map[string]bool{"success": true}))
}

// CreateBeta creates an upload session on the beta broker.
func (usr *UploadSessionResource) CreateBeta(c buffalo.Context) error {
	req := uploadSessionCreateReq{}
	oyster_utils.ParseReqBody(c.Request(), &req)

	betaTreasureIndexes := oyster_utils.GenerateInsertedIndexesForPearl(req.FileSizeBytes)

	// Generates ETH address.
	eth := services.Eth{}
	betaEthAddr, privKey, _ := eth.GenerateEthAddr()

	u := models.UploadSession{
		Type:                 models.SessionTypeBeta,
		GenesisHash:          req.GenesisHash,
		FileSizeBytes:        req.FileSizeBytes,
		StorageLengthInYears: req.StorageLengthInYears,
		TreasureIdxMap:       oyster_utils.GetTreasureIdxMap(req.AlphaTreasureIndexes, betaTreasureIndexes),
		TotalCost:            req.Invoice.Cost,
		ETHAddrAlpha:         req.Invoice.EthAddress,
		ETHAddrBeta:          nulls.NewString(betaEthAddr),
		ETHPrivateKey:        privKey,
	}
	vErr, err := u.StartUploadSession()
	if err != nil {
		return err
	}

	if len(vErr.Errors) > 0 {
		c.Render(422, r.JSON(vErr.Errors))
		return err
	}

	res := uploadSessionCreateBetaRes{
		UploadSession:       u,
		ID:                  u.ID.String(),
		Invoice:             u.GetInvoice(),
		BetaTreasureIndexes: betaTreasureIndexes,
	}
	return c.Render(200, r.JSON(res))
}

func (usr *UploadSessionResource) GetPaymentStatus(c buffalo.Context) error {
	session := models.UploadSession{}
	err := models.DB.Find(&session, c.Param("id"))

	if (err != nil || session == models.UploadSession{}) {
		//TODO: Return better error response when ID does not exist
		return err
	}

	res := paymentStatusCreateRes{
		ID:            session.ID.String(),
		PaymentStatus: session.GetPaymentStatus(),
	}

	return c.Render(200, r.JSON(res))
}

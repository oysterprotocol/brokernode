package actions

import (
	"bytes"
	"encoding/json"
	"net/http"

	"fmt"
	raven "github.com/getsentry/raven-go"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/nulls"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"github.com/pkg/errors"
	"strings"
)

type UploadSessionResource struct {
	buffalo.Resource
}

// Request Response structs

type uploadSessionCreateReq struct {
	GenesisHash          string         `json:"genesisHash"`
	NumChunks            int            `json:"numChunks"`
	FileSizeBytes        int            `json:"fileSizeBytes"` // This is Trytes instead of Byte
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

	alphaEthAddr, privKey, _ := services.EthWrapper.GenerateEthAddr()

	// Start Alpha Session.
	alphaSession := models.UploadSession{
		Type:                 models.SessionTypeAlpha,
		GenesisHash:          req.GenesisHash,
		FileSizeBytes:        req.FileSizeBytes,
		NumChunks:            req.NumChunks,
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

	req.AlphaTreasureIndexes = oyster_utils.GenerateInsertedIndexesForPearl(oyster_utils.ConvertToByte(req.FileSizeBytes))

	// Start Beta Session.
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

	err = models.DB.Save(&alphaSession)
	// Update alpha treasure idx map.
	alphaSession.MakeTreasureIdxMap(req.AlphaTreasureIndexes, betaTreasureIndexes)
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
	treasureIdxMap, err := uploadSession.GetTreasureIndexes()

	// Update dMaps to have chunks async
	go func() {
		var sqlWhereClosures []string
		chunksMap := make(map[string]Chunks)
		for _, chunk := range req.Chunks {
			var chunkIdx int
			if oyster_utils.BrokerMode == oyster_utils.TestModeNoTreasure {
				chunkIdx = chunk.Idx
			} else {
				chunkIdx = oyster_utils.TransformIndexWithBuriedIndexes(chunk.Idx, treasureIdxMap)
			}

			key := sqlWhereForGenesisHashAndChunkIdx(uploadSession.GenesisHash, chunkIdx)
			sqlWhereClosures = append(sqlWhereClosures, key)
			chunksMap[key] = chunk
		}

		var dms []models.DataMap
		rawQuery := fmt.Sprintf("SELECT * from data_maps WHERE %s", strings.Join(sqlWhereClosure, " OR "))
		err := models.DB.RawQuery(rawQuery).All(&dms)

		if err != nil {
			raven.CaptureError(err, nil)
		}

		dmsMap := make(map[string]models.DataMap)
		for _, dm := range dms {
			key := sqlWhereForGenesisHashAndChunkIdx(dm.GenesisHash, dm.ChunkIdx)
			// Only use the first one
			if _, hasKey := dmsMap[key]; !hasKey {
				dmsMap[key] = dm
			}
		}

		dbOperation := oyster_utils.CreateDbUpdateOperation(&models.DataMap{})
		var updatedDms []string
		for key, chunk := range chunksMap {
			dm, hasKey := dmsMap[key]
			if !hasKey {
				continue
			}

			if chunk.Hash == dm.GenesisHash {
				dm.Message = chunk.Data
				if oyster_utils.BrokerMode == oyster_utils.TestModeNoTreasure {
					dm.Status = models.Unassigned
				}
				vErr, _ := dm.Validate(nil)
				if len(vErr.Error) == 0 {
					updatedDms = append(updatedDms, fmt.Sprintf("(%s)", dbOperation.GetUpdatedValue(dm)))
				}
			}
		}

		// Do an insert operation and dup by primary key.
		rawQuery = fmt.Sprintf("INSERT INTO data_maps (%s) VALUES %s ON DUPLICATE KEY UPDATE message = VALUES(message), status = VALUES(status), updated_at = VALUES(updated_at)",
			dbOperation.GetColumns(), strings.Join(updatedDms, ","))
		err = models.DB.RawQuery(rawQuery).All(&[]models.DataMap)

		if err != nil {
			raven.CaptureError(err, nil)
		}
	}()

	return c.Render(202, r.JSON(map[string]bool{"success": true}))
}

// CreateBeta creates an upload session on the beta broker.
func (usr *UploadSessionResource) CreateBeta(c buffalo.Context) error {
	req := uploadSessionCreateReq{}
	oyster_utils.ParseReqBody(c.Request(), &req)

	betaTreasureIndexes := oyster_utils.GenerateInsertedIndexesForPearl(oyster_utils.ConvertToByte(req.FileSizeBytes))

	// Generates ETH address.
	betaEthAddr, privKey, _ := services.EthWrapper.GenerateEthAddr()

	u := models.UploadSession{
		Type:                 models.SessionTypeBeta,
		GenesisHash:          req.GenesisHash,
		NumChunks:            req.NumChunks,
		FileSizeBytes:        req.FileSizeBytes,
		StorageLengthInYears: req.StorageLengthInYears,
		TotalCost:            req.Invoice.Cost,
		ETHAddrAlpha:         req.Invoice.EthAddress,
		ETHAddrBeta:          nulls.NewString(betaEthAddr),
		ETHPrivateKey:        privKey,
	}
	vErr, err := u.StartUploadSession()

	u.MakeTreasureIdxMap(req.AlphaTreasureIndexes, betaTreasureIndexes)
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

func sqlWhereForGenesisHashAndChunkIdx(string genesisHash, int chunkIdx) string {
	return fmt.Sprintf("(genesis_hash = %s AND chunk_idx = %d)", uploadSession.GenesisHash, chunkIdx)
}

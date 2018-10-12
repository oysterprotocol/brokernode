package actions_v3

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
	GenesisHash          string `json:"genesisHash"`
	NumChunks            int    `json:"numChunks"`
	FileSizeBytes        uint64 `json:"fileSizeBytes"` // This is Trytes instead of Byte
	BetaIP               string `json:"betaIp"`
	StorageLengthInYears int    `json:"storageLengthInYears"`
	Version              uint32 `json:"version"`
}

type uploadSessionCreateBetaResV3 struct {
	ID              string `json:"id"`
	TreasureIndexes []int  `json:"treasureIndexes"`
	ETHAddr         string `json:"ethAddr"`
}

type uploadSessionCreateResV3 struct {
	ID            string `json:"id"`
	BetaSessionID string `json:"betaSessionId"`
	BatchSize     int    `json:"batchSize"`
}

var NumChunksLimit = -1 //unlimited

func init() {
	if v, err := strconv.Atoi(os.Getenv("NUM_CHUNKS_LIMIT")); err == nil {
		NumChunksLimit = v
	}
}

// Update uploads a chunk associated with an upload session.
func (usr *UploadSessionResourceV3) Update(c buffalo.Context) error {
	return c.Render(202, actions_utils.Render.JSON(map[string]bool{"success": true}))
}

/* Create endpoint. */
func (usr *UploadSessionResourceV3) Create(c buffalo.Context) error {
	req, err := validateAndGetCreateReq(c)
	if err != nil {
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

	// Generate bucket_name for s3 and create such bucket_name

	hasBeta := req.BetaIP != ""
	var betaSessionID = ""
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
		alphaSession.ETHAddrBeta = nulls.NewString(betaSessionRes.ETHAddr)
	}

	res := uploadSessionCreateResV3{
		ID:            alphaSession.ID.String(),
		BetaSessionID: betaSessionID,
		BatchSize:     BatchSize,
	}

	return c.Render(200, actions_utils.Render.JSON(res))
}

/* CreateBeta endpoint. */
func (usr *UploadSessionResourceV3) CreateBeta(c buffalo.Context) error {
	res := uploadSessionCreateBetaResV3{}
	return c.Render(200, actions_utils.Render.JSON(res))
}

func validateAndGetCreateReq(c buffalo.Context) (uploadSessionCreateReqV3, error) {
	req := uploadSessionCreateReqV3{}
	if err := oyster_utils.ParseReqBody(c.Request(), &req); err != nil {
		err = fmt.Errorf("Invalid request, unable to parse request body  %v", err)
		c.Error(400, err)
		return req, err
	}

	if NumChunksLimit != -1 && req.NumChunks > NumChunksLimit {
		err := errors.New("This broker has a limit of " + fmt.Sprint(NumChunksLimit) + " file chunks.")
		fmt.Println(err)
		c.Error(400, err)
		return req, err
	}
	return req, nil
}

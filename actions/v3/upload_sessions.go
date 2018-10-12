package actions_v3

import (
	"fmt"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/oysterprotocol/brokernode/actions/utils"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/utils"
	"github.com/pkg/errors"
	"gopkg.in/segmentio/analytics-go.v3"
)

const (
	BatchSize = 25
)

type bucketRequest struct {
	Path    string            `json:"path"`
	Headers map[string]string `json:"headers"`
}

type uploadSessionCreateBetaResV3 struct {
}

type UploadSessionResourceV3 struct {
	buffalo.Resource
}

type UploadSessionUpdateReqV3 struct {
	Chunks []models.ChunkReq `json:"chunks"`
}

type UploadSessionCreateReqV3 struct {
	GenesisHash          string `json:"genesisHash"`
	NumChunks            int    `json:"numChunks"`
	FileSizeBytes        uint64 `json:"fileSizeBytes"` // This is Trytes instead of Byte
	BetaIP               string `json:"betaIp"`
	StorageLengthInYears int    `json:"storageLengthInYears"`
}

type UploadSessionCreateResV3 struct {
	BatchSize   int           `json:"batchSize"`
	UploadAlpha bucketRequest `json:"uploadAlpha"`
	UploadBeta  bucketRequest `json:"uploadBeta"`
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

func (usr *UploadSessionResourceV3) Create(c buffalo.Context) error {
	req, err := validateAndGetCreateReq(c)
	if err != nil {
		return err
	}

	alphaEthAddr, privKey, _ := EthWrapper.GenerateEthAddr()
	uploadBeta = bucketRequest{}
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
	}

	res := uploadSessionCreateResV3{
		BatchSize:   BatchSize,
		UploadAlpha: bucketRequest{},
		UploadBeta:  uploadBeta,
	}
	return c.Render(200, actions_utils.Render.JSON(res))
}

func (usr *UploadSessionResourceV3) CreateBeta(c buffalo.Context) error {
	res := uploadSessionCreateBetaResV3{}
	return c.Render(200, actions_utils.Render.JSON(res))
}

func validateAndGetCreateReq(c buffalo.Context) uploadSessionCreateReqV3, error {
	req := uploadSessionCreateReqV3{}
	if err := oyster_utils.ParseReqBody(c.Request(), &req); err != nil {
		err = fmt.Errorf("Invalid request, unable to parse request body  %v", err)
		c.Error(400, err)
		return nil, err
	}

	if NumChunksLimit != -1 && req.NumChunks > NumChunksLimit {
		err := errors.New("This broker has a limit of " + fmt.Sprint(NumChunksLimit) + " file chunks.")
		fmt.Println(err)
		c.Error(400, err)
		return nil, err
	}
	return req, nil
}

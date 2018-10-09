package actions_v3

import (
	"time"
	"fmt"

	"github.com/gobuffalo/buffalo"
	"github.com/oysterprotocol/brokernode/actions/utils"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/utils"
	"github.com/pkg/errors"
	"gopkg.in/segmentio/analytics-go.v3"
)

type UploadSessionResourceV3 struct {
	buffalo.Resource
}

type UploadSessionUpdateReqV3 struct {
	Chunks []models.ChunkReq `json:"chunks"`
}

func init() {

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

	return c.Render(202, actions_utils.Render.JSON(map[string]bool{"success": true}))
}

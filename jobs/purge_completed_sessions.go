package jobs

import (
	"errors"
	"fmt"
	"sync"

	"github.com/gobuffalo/pop"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"gopkg.in/segmentio/analytics-go.v3"
)

var purgeMutex = &sync.Mutex{}

func PurgeCompletedSessions(PrometheusWrapper services.PrometheusService) {

	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramPurgeCompletedSessions, start)

	purgeMutex.Lock()
	defer purgeMutex.Unlock()

	completeGenesisHashes, err := getAllCompletedGenesisHashes()
	if err != nil {
		oyster_utils.LogIfError(errors.New(err.Error()+" getting the completeGenesisHashes in "+
			"purge_completed_sessions"), nil)
	}

	for _, genesisHash := range completeGenesisHashes {
		purgeSessions(genesisHash)
	}
}

func getAllCompletedGenesisHashes() ([]string, error) {
	genesisHashes := []string{}
	sessions, err := models.GetCompletedSessions()
	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return []string{}, err
	}

	for _, session := range sessions {
		genesisHashes = append(genesisHashes, session.GenesisHash)
	}
	return genesisHashes, err
}

func createCompletedDataMapIfNeeded(genesisHash string) error {
	completedUploadedSession := []models.CompletedUpload{}
	if err := models.DB.RawQuery("SELECT * FROM completed_uploads WHERE genesis_hash = ?", genesisHash).All(&completedUploadedSession); err != nil {
		oyster_utils.LogIfError(err, nil)
		return err
	}
	if len(completedUploadedSession) > 0 {
		return nil
	}

	session := []models.UploadSession{}
	if err := models.DB.RawQuery("SELECT * FROM upload_sessions WHERE genesis_hash = ?", genesisHash).All(&session); err != nil {
		oyster_utils.LogIfError(err, nil)
		return err
	}

	if len(session) == 0 {
		return nil
	}

	err := models.NewCompletedUpload(session[0])
	oyster_utils.LogIfError(err, nil)
	return err
}

func purgeSessions(genesisHash string) {
	err := models.DB.Transaction(func(tx *pop.Connection) error {

		sessions := []models.UploadSession{}

		err := models.DB.Where("genesis_hash = ?", genesisHash).All(&sessions)
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			return err
		}

		if len(sessions) > 0 {

			fmt.Println("PURGING SESSION!!")
			fmt.Println(sessions[0])

			errMovingChunks := sessions[0].MoveAllChunksToCompleted()
			if errMovingChunks != nil {
				oyster_utils.LogIfError(err, nil)
				return errMovingChunks
			}
			err = createCompletedDataMapIfNeeded(genesisHash)
			if err == nil {
				if err := tx.RawQuery("DELETE FROM upload_sessions WHERE genesis_hash = ?",
					genesisHash).All(&[]models.UploadSession{}); err != nil {
					oyster_utils.LogIfError(errors.New(err.Error()+" while deleting upload_sessions in "+
						"purge_completed_sessions"), nil)
					return err
				}
			}
		}
		return nil
	})
	oyster_utils.LogToSegment("purge_completed_sessions: completed_session_purged", analytics.NewProperties().
		Set("genesis_hash", genesisHash))

	if err != nil {
		oyster_utils.LogIfError(err, nil)
	}
}

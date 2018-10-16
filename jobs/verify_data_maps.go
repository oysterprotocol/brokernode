package jobs

import (
	"errors"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"time"
)

/*VerifyDataMaps will check sessions for which attachment has been completed, to make sure the chunks
have been attached to the tangle.*/
func VerifyDataMaps(IotaWrapper services.IotaService, PrometheusWrapper services.PrometheusService, thresholdTime time.Time) {

	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramVerifyDataMaps, start)

	sessions, err := models.GetVerifiableSessions(thresholdTime)

	for _, session := range sessions {
		checkSessionChunks(IotaWrapper, session)
	}

	if err != nil {
		oyster_utils.LogIfError(errors.New(err.Error()+" while getting sessions in VerifyDataMaps"), nil)
	}
}

func checkSessionChunks(IotaWrapper services.IotaService, sessionParam models.UploadSession) {

	// Get session
	session := &models.UploadSession{}
	models.DB.Find(session, sessionParam.ID)

	adjustIndexes(session)

	if session.NextIdxToAttach == session.NextIdxToVerify {
		return
	}

	chunkData := getChunksToVerify(session)

	for ok, i := true, 0; ok; ok = i < len(chunkData) {
		end := i + services.MaxNumberOfAddressPerFindTransactionRequest

		if end > len(chunkData) {
			end = len(chunkData)
		}

		if i >= end {
			break
		}

		if len(chunkData[i:end]) > 0 {
			CheckChunks(IotaWrapper, chunkData[i:end], session)
		}
		i += services.MaxNumberOfAddressPerFindTransactionRequest
	}
}

func getChunksToVerify(session *models.UploadSession) []oyster_utils.ChunkData {
	offset := int64(1)
	if session.Type == models.SessionTypeAlpha {
		offset = -1
	}

	keys := oyster_utils.GenerateBulkKeys(session.GenesisHash, session.NextIdxToVerify, session.NextIdxToAttach+offset)

	chunkData := []oyster_utils.ChunkData{}

	for ok, i := true, 0; ok; ok = i < len(*keys) {
		end := i + services.MaxNumberOfAddressPerFindTransactionRequest

		if end > len(*keys) {
			end = len(*keys)
		}

		if i >= end {
			break
		}

		if len((*(keys))[i:end]) > 0 {
			keySlice := oyster_utils.KVKeys{}
			keySlice = append(keySlice, (*(keys))[i:end]...)
			chunks, err := models.GetMultiChunkDataFromAnyDB(session.GenesisHash, &keySlice)
			if err != nil {
				oyster_utils.LogIfError(errors.New(err.Error()+" getting chunk data in checkSessionChunks in "+
					"verify_data_maps"), nil)
				continue
			}
			chunkData = append(chunkData, chunks...)
		}
		i += services.MaxNumberOfAddressPerFindTransactionRequest
	}
	return chunkData
}

func adjustIndexes(session *models.UploadSession) {
	if session.NextIdxToAttach == session.NextIdxToVerify {
		chunks, _ := session.GetUnassignedChunksBySession(1)
		if len(chunks) == 0 {
			if session.Type == models.SessionTypeAlpha {
				session.NextIdxToAttach = int64(session.NumChunks)
			} else {
				session.NextIdxToAttach = models.BetaSessionStopIdx
			}
			models.DB.ValidateAndUpdate(session)
		}
	}
}

/*CheckChunks will make calls to verify the chunks and update the indexes of the session*/
func CheckChunks(IotaWrapper services.IotaService, unverifiedChunks []oyster_utils.ChunkData,
	session *models.UploadSession) {

	filteredChunks, err := IotaWrapper.VerifyChunkMessagesMatchRecord(unverifiedChunks)
	if err != nil {
		oyster_utils.LogIfError(errors.New(err.Error()+" verifying chunks match record in CheckChunks() "+
			"in verify_data_maps"), nil)
	}

	oyster_utils.LogIfError(err, nil)

	session.DownGradeIndexesOnUnattachedChunks(filteredChunks.DoesNotMatchTangle)
	session.DownGradeIndexesOnUnattachedChunks(filteredChunks.NotAttached)
	session.UpdateIndexWithVerifiedChunks(filteredChunks.MatchesTangle)
}

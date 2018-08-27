package jobs

import (
	"errors"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
)

func VerifyDataMaps(IotaWrapper services.IotaService, PrometheusWrapper services.PrometheusService) {

	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramVerifyDataMaps, start)

	sessions, err := models.GetSessionsByAge()

	for _, session := range sessions {
		checkSessionChunks(IotaWrapper, session)
	}

	if err != nil {
		oyster_utils.LogIfError(errors.New(err.Error()+" getting sessions in VerifyDataMaps"), nil)
	}
}

func checkSessionChunks(IotaWrapper services.IotaService, session models.UploadSession) {
	if session.NextIdxToAttach == session.NextIdxToVerify {
		return
	}

	offset := int64(1)
	if session.Type == models.SessionTypeAlpha {
		offset = -1
	}

	keys := oyster_utils.GenerateBulkKeys(session.GenesisHash, session.NextIdxToVerify, session.NextIdxToAttach+offset)

	chunkData, err := models.GetMultiChunkData(oyster_utils.InProgressDir, session.GenesisHash, keys)
	if err != nil {
		oyster_utils.LogIfError(errors.New(err.Error()+" getting chunk data in checkSessionChunks in "+
			"verify_data_maps"), nil)
		return
	}

	CheckChunks(IotaWrapper, chunkData, session)
}

/*CheckChunks will make calls to verify the chunks and update the indexes of the session*/
func CheckChunks(IotaWrapper services.IotaService, unverifiedChunks []oyster_utils.ChunkData,
	session models.UploadSession) {
	filteredChunks, err := IotaWrapper.VerifyChunkMessagesMatchRecord(unverifiedChunks)
	if err != nil {
		oyster_utils.LogIfError(errors.New(err.Error()+" verifying chunks match record in CheckChunks() "+
			"in verify_data_maps"), nil)
	}

	treasureChunks := []oyster_utils.ChunkData{}
	nonTreasureChunks := []oyster_utils.ChunkData{}

	if len(filteredChunks.DoesNotMatchTangle) > 0 {
		treasureIndexes, err := session.GetTreasureIndexes()
		oyster_utils.LogIfError(err, nil)

		treasureIdxMap := make(map[int64]bool)
		for _, index := range treasureIndexes {
			treasureIdxMap[int64(index)] = true
		}

		for _, chunk := range filteredChunks.DoesNotMatchTangle {
			if _, ok := treasureIdxMap[chunk.Idx]; ok {
				treasureChunks = append(treasureChunks, chunk)
			} else {
				nonTreasureChunks = append(nonTreasureChunks, chunk)
			}
		}
		session.DownGradeIndexesOnUnattachedChunks(nonTreasureChunks)
	}

	if len(filteredChunks.NotAttached) > 0 {
		session.DownGradeIndexesOnUnattachedChunks(filteredChunks.NotAttached)
	}

	if len(filteredChunks.MatchesTangle) > 0 {
		chunks := InsertTreasureChunks(filteredChunks.MatchesTangle, treasureChunks, session)

		session.MoveChunksToCompleted(chunks)
		session.UpdateIndexWithVerifiedChunks(chunks)
	}
}

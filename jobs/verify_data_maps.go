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

	unverifiedDataMaps := []models.DataMap{}

	err := models.DB.Where("status = ?", models.Unverified).All(&unverifiedDataMaps)
	oyster_utils.LogIfError(err, nil)

	if len(unverifiedDataMaps) > 0 {
		for i := 0; i < len(unverifiedDataMaps); i += BundleSize {
			end := i + BundleSize

			if end > len(unverifiedDataMaps) {
				end = len(unverifiedDataMaps)
			}

			CheckChunks(IotaWrapper, unverifiedDataMaps[i:end])
		}
	}
}

func CheckChunks(IotaWrapper services.IotaService, unverifiedDataMaps []models.DataMap) {
	filteredChunks, err := IotaWrapper.VerifyChunkMessagesMatchRecord(unverifiedDataMaps)
	oyster_utils.LogIfError(errors.New(err.Error()+" verifying chunks match record in CheckChunks() "+
		"in verify_data_maps"), nil)

	if len(filteredChunks.MatchesTangle) > 0 {

		var updatedDms []string
		dbOperation, _ := oyster_utils.CreateDbUpdateOperation(&models.DataMap{})

		for _, matchingChunk := range filteredChunks.MatchesTangle {
			matchingChunk.Status = models.Complete
			updatedDms = append(updatedDms, dbOperation.GetUpdatedValue(matchingChunk))
		}

		err := models.BatchUpsert(
			"data_maps",
			updatedDms,
			dbOperation.GetColumns(),
			[]string{"status"})

		oyster_utils.LogIfError(errors.New(err.Error()+" setting status to complete in CheckChunks() "+
			"in verify_data_maps"), nil)
	}

	if len(filteredChunks.DoesNotMatchTangle) > 0 {

		// when we bring back hooknodes, decrement their reputation here

		var updatedDms []string
		dbOperation, _ := oyster_utils.CreateDbUpdateOperation(&models.DataMap{})

		for _, notMatchingChunk := range filteredChunks.DoesNotMatchTangle {

			// if a chunk did not match the tangle in verify_data_maps
			// we mark it as "Error" and there is no reason to check the tangle
			// for it again while its status is still in an Error state

			// this is to prevent verifyChunkMessageMatchesTangle
			// from being executed on an Error'd chunk in process_unassigned_chunks
			notMatchingChunk.Status = models.Error
			notMatchingChunk.TrunkTx = ""
			notMatchingChunk.BranchTx = ""
			notMatchingChunk.NodeID = ""
			updatedDms = append(updatedDms, dbOperation.GetUpdatedValue(notMatchingChunk))
		}

		err := models.BatchUpsert(
			"data_maps",
			updatedDms,
			dbOperation.GetColumns(),
			[]string{"status", "trunk_tx", "branch_tx", "node_id"})

		oyster_utils.LogIfError(errors.New(err.Error()+" setting status to error in CheckChunks() "+
			"in verify_data_maps"), nil)
	}
}

package jobs

import (
	"fmt"
	"github.com/getsentry/raven-go"
	"github.com/oysterprotocol/brokernode/models"
)

func init() {
}

func ProcessPaidSessions() {

	BuryTreasureInPaidDataMaps()
	MarkBuriedMapsAsUnassigned()
}

func BuryTreasureInPaidDataMaps() {
	unburiedSessions := []models.UploadSession{}

	err := models.DB.Where("payment_status = ? AND treasure_status = ?",
		models.Paid, models.Unburied).All(&unburiedSessions)
	if err != nil {
		raven.CaptureError(err, nil)
	}

	treasureChunks := []models.DataMap{}
	treasureChunks2 := []models.DataMap{}

	for _, unburiedSession := range unburiedSessions {
		// when this issue is completed, https://github.com/oysterprotocol/brokernode/issues/207
		// build an array of the treasure locations here
		// for now just using an empty dummy array

		dummyTreasureLocations := []int{1}

		err = models.DB.RawQuery("SELECT * from data_maps WHERE genesis_hash = ? "+
			"AND chunk_idx in (?) ORDER BY chunk_idx asc",
			unburiedSession.GenesisHash,
			dummyTreasureLocations).All(&treasureChunks)

		err = models.DB.RawQuery("SELECT * from data_maps WHERE genesis_hash = ? ORDER BY chunk_idx asc",
			unburiedSession.GenesisHash).All(&treasureChunks2)

		fmt.Println("_______________________________")
		fmt.Println("_______________________________")
		fmt.Println("_______________________________")
		fmt.Println("_______________________________")
		fmt.Println(len(treasureChunks))
		fmt.Println(len(treasureChunks2))
		fmt.Println("_______________________________")
		fmt.Println("_______________________________")
		fmt.Println("_______________________________")
		fmt.Println("_______________________________")

		BuryTreasure(treasureChunks, &unburiedSession)

		if len(treasureChunks) != 0 {
			unburiedSession.TreasureStatus = models.Buried
			models.DB.ValidateAndSave(&unburiedSession)
		}
	}
}

func BuryTreasure(treasureChunks []models.DataMap, unburiedSession *models.UploadSession) {

	var err error

	// @TODO grab the ethereum seeds once we have ethereum functionality enabled.  For now just using a dummy seed.

	// @TODO do something in the event that the length of the seeds isn't the same as the length of the treasureChunks
	// @TODO seeds and treasure chunks should be in the same order

	dummyEthSeed := "1004444400000006780000000000000000000000000012345000000765430001"

	for _, treasureChunk := range treasureChunks {
		treasureChunk.Message, err = models.CreateTreasurePayload(dummyEthSeed, treasureChunk.Hash, models.MaxSideChainLength)
		if err != nil {
			raven.CaptureError(err, nil)
		}
		models.DB.ValidateAndSave(&treasureChunk)
	}
}

// marking the maps as "Unassigned" will mean they get processed
func MarkBuriedMapsAsUnassigned() {
	readySessions := []models.UploadSession{}

	err := models.DB.Where("payment_status = ? AND treasure_status = ?",
		models.Paid, models.Buried).All(&readySessions)
	if err != nil {
		raven.CaptureError(err, nil)
	}

	var dataMaps = []models.DataMap{}

	for _, readySession := range readySessions {
		err = models.DB.RawQuery("UPDATE data_maps SET status = ? WHERE genesis_hash = ? AND status = ?",
			models.Unassigned,
			readySession.GenesisHash,
			models.Pending).All(&dataMaps)
	}
}

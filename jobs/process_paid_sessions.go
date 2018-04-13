package jobs

import (
	"encoding/json"
	"github.com/getsentry/raven-go"
	"github.com/gobuffalo/pop/slices"
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
		models.PaymentStatusPaid, models.TreasureUnburied).All(&unburiedSessions)
	if err != nil {
		raven.CaptureError(err, nil)
	}

	treasureChunks := []models.DataMap{}

	for _, unburiedSession := range unburiedSessions {

		/*TODO
		pop queries for "where in" need a slice type to work (pop/slices)

		https://godoc.org/github.com/gobuffalo/pop/slices#Int

		But, I cannot seem to get it to work.
		treasureIndex is an array of ints but "where in" won't work with it, and cannot figure
		out how to convert a regular array of ints to type pop/slices.

		Even with just manually inputting some dummy ints into slices.Int{} the "where in" does not
		seem to work.
		*/

		// this is taking the json from treasureIdxMap and making it an array of ints.
		// tried to pass in slices.Int{} as the interface but that did not work.
		// https://play.golang.org/p/nQorh-mMYOw
		treasureIndex := []int{}
		if unburiedSession.TreasureIdxMap.Valid {
			// only do this if the string value is valid
			err := json.Unmarshal([]byte(unburiedSession.TreasureIdxMap.String), &treasureIndex)
			if err != nil {
				raven.CaptureError(err, nil)
			}
		}

		/*@TODO remove this and actually make a slices.Int object that haves the values from
		TreasureIdxMap and get the "where in" query to work
		*/
		treasureSlices := slices.Int{1, 220, 355}

		err = models.DB.RawQuery("SELECT * from data_maps WHERE genesis_hash = ? "+
			"AND chunk_idx in (?) ORDER BY chunk_idx asc",
			unburiedSession.GenesisHash,
			treasureSlices).All(&treasureChunks)

		BuryTreasure(treasureChunks, &unburiedSession)

		if len(treasureChunks) != 0 {
			unburiedSession.TreasureStatus = models.TreasureBuried
			models.DB.ValidateAndSave(&unburiedSession)
		}
	}
}

func BuryTreasure(treasureChunks []models.DataMap, unburiedSession *models.UploadSession) {

	var err error

	/*@TODO would be better to re-write this method such that we're passing in the seeds, so we can pass in
	@TODO dummy seeds for our unit tests.
	*/

	//@TODO grab the ethereum seeds once we have ethereum functionality enabled.  For now just using a dummy seed.

	//@TODO do something in the event that the length of the seeds isn't the same as the length of the treasureChunks

	/*@TODO remove this*/
	dummyEthSeed := "1004444400000006780000000000000000000000000012345000000765430001"

	for _, treasureChunk := range treasureChunks {
		treasureChunk.Message, err = models.CreateTreasurePayload(dummyEthSeed, treasureChunk.Hash, models.MaxSideChainLength)
		if err != nil {
			raven.CaptureError(err, nil)
		}
		models.DB.ValidateAndSave(&treasureChunk)
	}
}

// marking the maps as "Unassigned" will trigger them to get processed by the process_unassigned_chunks cron task.
func MarkBuriedMapsAsUnassigned() {
	readySessions := []models.UploadSession{}

	err := models.DB.Where("payment_status = ? AND treasure_status = ?",
		models.PaymentStatusPaid, models.TreasureBuried).All(&readySessions)
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

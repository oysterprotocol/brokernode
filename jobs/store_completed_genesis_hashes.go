package jobs

import (
	"errors"
	"github.com/gobuffalo/pop"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
)

/*StoreCompletedGenesisHashes will look for completed sessions and store their
genesis hashes to sell to webnodes*/
func StoreCompletedGenesisHashes(PrometheusWrapper services.PrometheusService) {

	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramStoreCompletedGenesisHashes, start)

	completeGenesisHashes, err := getAllCompletedGenesisHashes()
	if err != nil {
		oyster_utils.LogIfError(errors.New(err.Error()+" getting completeGenesisHashes in "+
			"store_complete_genesis_hashes"), nil)
	}

	for _, genesisHash := range completeGenesisHashes {

		err = models.DB.Transaction(func(tx *pop.Connection) error {

			session := []models.UploadSession{}
			if err := tx.RawQuery("SELECT * FROM upload_sessions WHERE genesis_hash = ?", genesisHash).All(&session); err != nil {
				oyster_utils.LogIfError(errors.New(err.Error()+" finding upload_sessions that match "+
					"genesis_hash in store_complete_genesis_hashes"), nil)
				return err
			}

			if len(session) > 0 {

				genesisHashExistsAlready, err := models.CheckIfGenesisHashExists(session[0].GenesisHash)
				if err != nil {
					oyster_utils.LogIfError(errors.New(err.Error()+" error checking if genesis_hash exists in "+
						"store_complete_genesis_hashes"), nil)
				}

				if !genesisHashExistsAlready {
					vErr, err := tx.ValidateAndSave(&models.StoredGenesisHash{
						GenesisHash:   session[0].GenesisHash,
						NumChunks:     session[0].NumChunks,
						FileSizeBytes: session[0].FileSizeBytes,
					})
					if vErr.HasAny() {
						oyster_utils.LogIfValidationError("StoredGenesisHash validation failed.", vErr, nil)
						return errors.New("Unable to validate StoredGenesisHash")
					}
					if err != nil {
						oyster_utils.LogIfError(errors.New(err.Error()+" saving stored_genesis_hash in "+
							"store_complete_genesis_hashes"), nil)
						return err
					}
				}
			}

			return nil
		})
	}
}

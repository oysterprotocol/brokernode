package jobs

import (
	"github.com/gobuffalo/pop"
	"github.com/oysterprotocol/brokernode/models"
	"log"
)

func init() {
}

func PurgeCompletedSessions() {

	var genesisHashesNotComplete = []models.DataMap{}
	var allGenesisHashesStruct = []models.DataMap{}

	err := models.DB.RawQuery("SELECT distinct genesis_hash FROM data_maps").All(&allGenesisHashesStruct)

	if err != nil {
		log.Panic(err)
	}

	allGenesisHashes := make([]string, 0, len(allGenesisHashesStruct))

	for _, genesisHash := range allGenesisHashesStruct {
		allGenesisHashes = append(allGenesisHashes, genesisHash.GenesisHash)
	}

	err = models.DB.RawQuery("SELECT distinct genesis_hash FROM data_maps WHERE status != ? AND status != ?",
		models.Complete,
		models.Confirmed).All(&genesisHashesNotComplete)

	if err != nil {
		log.Panic(err)
	}

	notComplete := map[string]bool{}

	for _, genesisHash := range genesisHashesNotComplete {
		notComplete[genesisHash.GenesisHash] = true
	}

	var moveToComplete = []models.DataMap{}

	for _, genesisHash := range allGenesisHashes {
		if !notComplete[genesisHash] {

			models.DB.Transaction(func(tx *pop.Connection) error {
				tx.RawQuery("SELECT * from data_maps WHERE genesis_hash = ?", genesisHash).All(&moveToComplete)
				MoveToComplete(tx, moveToComplete) // Passed in the connection

				err = tx.RawQuery("DELETE from data_maps WHERE genesis_hash = ?", genesisHash).All(&[]models.DataMap{})
				if err != nil {
					return err
				}

				session := []models.UploadSession{}

				err = tx.RawQuery("SELECT * from upload_sessions WHERE genesis_hash = ?", genesisHash).All(&session)
				if err != nil {
					return err
				}

				if len(session) > 0 {
					_, err = tx.ValidateAndSave(&models.StoredGenesisHash{GenesisHash: session[0].GenesisHash, FileSizeBytes: session[0].FileSizeBytes})
					if err != nil {
						return err
					}
				}

				err = tx.RawQuery("DELETE from upload_sessions WHERE genesis_hash = ?", genesisHash).All(&[]models.UploadSession{})
				if err != nil {
					return err
				}

				return nil
			})
		}
	}
}

func MoveToComplete(tx *pop.Connection, dataMaps []models.DataMap) {

	for _, dataMap := range dataMaps {

		completedDataMap := models.CompletedDataMap{

			Status:      dataMap.Status,
			Message:     dataMap.Message,
			NodeID:      dataMap.NodeID,
			NodeType:    dataMap.NodeType,
			TrunkTx:     dataMap.TrunkTx,
			BranchTx:    dataMap.BranchTx,
			GenesisHash: dataMap.GenesisHash,
			ChunkIdx:    dataMap.ChunkIdx,
			Hash:        dataMap.Hash,
			Address:     dataMap.Address,
		}

		_, err := tx.ValidateAndSave(&completedDataMap)
		if err != nil {
			log.Panic(err)
		}
	}
}

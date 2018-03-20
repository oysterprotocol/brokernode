package jobs

import (
	"github.com/oysterprotocol/brokernode/models"
	"strconv"
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
		strconv.Itoa(models.Complete),
		strconv.Itoa(models.Confirmed)).All(&genesisHashesNotComplete)

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
			models.DB.RawQuery("SELECT * from data_maps WHERE genesis_hash = ?", genesisHash).All(&moveToComplete)
			MoveToComplete(moveToComplete)
			models.DB.RawQuery("DELETE from data_maps WHERE genesis_hash = ?", genesisHash).All(&[]models.DataMap{})
			models.DB.RawQuery("DELETE from upload_sessions WHERE genesis_hash = ?", genesisHash).All(&[]models.UploadSession{})
		}
	}
}

func MoveToComplete(dataMaps []models.DataMap) {

	for _, dataMap := range dataMaps {

		completedDataMap := models.CompletedDataMap{
			Status:      dataMap.Status,
			HooknodeIP:  dataMap.HooknodeIP,
			Message:     dataMap.Message,
			TrunkTx:     dataMap.TrunkTx,
			BranchTx:    dataMap.BranchTx,
			GenesisHash: dataMap.GenesisHash,
			ChunkIdx:    dataMap.ChunkIdx,
			Hash:        dataMap.Hash,
			Address:     dataMap.Address,
		}
		_, err := models.DB.ValidateAndSave(&completedDataMap)
		if err != nil {
			log.Panic(err)
		}
	}
}

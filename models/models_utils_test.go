package models_test

import (
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/utils"
)

func (suite *ModelSuite) Test_InsertToEmptyDataMap() {
	dbOperation, _ := oyster_utils.CreateDbUpdateOperation(&models.DataMap{})

	var dms []string
	for i := 0; i < 2; i++ {
		dm := models.DataMap{
			MsgStatus:   models.MsgStatusUploadedHaveNotEncoded,
			GenesisHash: "Test_InsertToEmptyDataMap",
			ChunkIdx:    i,
		}
		dms = append(dms, dbOperation.GetNewInsertedValue(dm))
	}
	suite.Nil(models.BatchUpsert("data_maps", dms, dbOperation.GetColumns(), nil))

	dataMaps := []models.DataMap{}
	suite.DB.RawQuery("SELECT * FROM data_maps WHERE genesis_hash = ?", "Test_InsertToEmptyDataMap").All(&dataMaps)

	suite.Equal(2, len(dataMaps))
	for i := 0; i < 2; i++ {
		suite.Equal(models.MsgStatusUploadedHaveNotEncoded, dataMaps[i].MsgStatus)
		suite.True(dataMaps[i].ChunkIdx == 0 || dataMaps[i].ChunkIdx == 1)
	}
}

func (suite *ModelSuite) Test_UpdateDataMap() {
	dbOperation, _ := oyster_utils.CreateDbUpdateOperation(&models.DataMap{})

	var dms []string
	for i := 0; i < 2; i++ {
		dm := models.DataMap{
			MsgStatus:   models.MsgStatusUploadedHaveNotEncoded,
			GenesisHash: "Test_UpdateDataMap",
			ChunkIdx:    i,
		}
		dms = append(dms, dbOperation.GetNewInsertedValue(dm))
	}
	suite.Nil(models.BatchUpsert("data_maps", dms, dbOperation.GetColumns(), nil))

	dataMaps := []models.DataMap{}
	suite.DB.RawQuery("SELECT * FROM data_maps WHERE genesis_hash = ? LIMIT 1", "Test_UpdateDataMap").All(&dataMaps)

	dm := models.DataMap{
		ID:          dataMaps[0].ID,
		MsgStatus:   models.MsgStatusNotUploaded,
		GenesisHash: "Test_UpdateDataMap",
		ChunkIdx:    dataMaps[0].ChunkIdx,
	}
	dms = []string{dbOperation.GetUpdatedValue(dm)}

	suite.Nil(models.BatchUpsert("data_maps", dms, dbOperation.GetColumns(), []string{"msg_status"}))

	updatedDataMaps := []models.DataMap{}
	suite.DB.RawQuery("SELECT * FROM data_maps WHERE genesis_hash = ? AND chunk_idx = ?", "Test_UpdateDataMap", dataMaps[0].ChunkIdx).All(&updatedDataMaps)

	suite.Equal(models.MsgStatusNotUploaded, updatedDataMaps[0].MsgStatus)
}

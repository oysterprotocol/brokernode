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
			Hash:        i,
			GenesisHash: "Test_InsertToEmptyDataMap",
		}
		dms = append(dms, dbOperation.GetNewInsertedValue(dm))
	}
	suite.Nil(models.BatchUpsert("data_maps", dms, dbOperation.GetColumns(), nil))

	dataMaps := []models.DataMap{}
	suite.DB.RawQuery("SELECT * FROM data_maps WHERE genesis_hash = ?", "Test_InsertToEmptyDataMap").All(&dataMaps)

	suite.Equal(2, len(dataMaps))
	for i := 0; i < 2; i++ {
		suite.Equal(models.MsgStatusUploadedHaveNotEncoded, dataMaps[i].MsgStatus)
	}
}

func (suite *ModelSuite) Test_UpdateDataMap() {
	dbOperation, _ := oyster_utils.CreateDbUpdateOperation(&models.DataMap{})

	var dms []string
	for i := 0; i < 2; i++ {
		dm := models.DataMap{
			MsgStatus:   models.MsgStatusUploadedHaveNotEncoded,
			Hash:        i,
			GenesisHash: "Test_UpdateDataMap",
		}
		dms = append(dms, dbOperation.GetNewInsertedValue(dm))
	}
	suite.Nil(models.BatchUpsert("data_maps", dms, dbOperation.GetColumns(), nil))

	dataMaps := []models.DataMap{}
	suite.DB.RawQuery("SELECT * FROM data_maps WHERE genesis_hash = ? LIMIT 1", "Test_UpdateDataMap").All(&dataMaps)

	dm := models.DataMap{
		ID:        dataMaps[0].ID,
		MsgStatus: models.MsgStatusNotUploaded,
	}
	dms = []string{dbOperation.GetUpdatedValue(dm)}

	suite.Nil(models.BatchUpsert("data_maps", dms, dbOperation.GetColumns(), []string{"msg_status"}))

	suite.DB.RawQuery("SELECT * FROM data_maps WHERE gensis_hash = ? AND hash = ?", "Test_UpdateDataMap", dataMaps[0].Hash).All(&dataMaps)

	suite.Equal(models.MsgStatusNotUploaded, dataMaps[0].MsgStatus)
}

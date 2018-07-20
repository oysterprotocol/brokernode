package models_test

import (
	"github.com/oysterprotocol/brokernode/models"
	"time"
)

func (suite *ModelSuite) Test_GetGenesisHashForWebnode_no_new_genesis_hashes() {

	existingGenesisHashes := []string{
		"abcdef11",
		"abcdef12",
		"abcdef13",
	}

	storedGenesisHash1 := models.StoredGenesisHash{
		GenesisHash:    "abcdef11",
		FileSizeBytes:  5000,
		NumChunks:      5,
		WebnodeCount:   0,
		Status:         models.StoredGenesisHashUnassigned,
		TreasureStatus: models.TreasureBuried,
	}

	storedGenesisHash2 := models.StoredGenesisHash{
		GenesisHash:    "abcdef12",
		FileSizeBytes:  5000,
		NumChunks:      5,
		WebnodeCount:   0,
		Status:         models.StoredGenesisHashUnassigned,
		TreasureStatus: models.TreasureBuried,
	}

	vErr, err := suite.DB.ValidateAndCreate(&storedGenesisHash1)
	suite.Nil(err)
	suite.False(vErr.HasAny())

	vErr, err = suite.DB.ValidateAndCreate(&storedGenesisHash2)
	suite.Nil(err)
	suite.False(vErr.HasAny())

	storedGenesisHash, err := models.GetGenesisHashForWebnode(existingGenesisHashes)

	suite.Equal(models.StoredGenesisHash{}, storedGenesisHash)
	suite.Equal(models.NoGenHashesMessage, err.Error())
}

func (suite *ModelSuite) Test_GetGenesisHashForWebnode_none_unassigned() {

	existingGenesisHashes := []string{
		"abcdef11",
		"abcdef12",
		"abcdef13",
	}

	storedGenesisHash1 := models.StoredGenesisHash{
		GenesisHash:    "aaaaaa",
		FileSizeBytes:  5000,
		NumChunks:      5,
		WebnodeCount:   0,
		Status:         models.StoredGenesisHashAssigned, // assigned already
		TreasureStatus: models.TreasureBuried,
	}

	vErr, err := suite.DB.ValidateAndCreate(&storedGenesisHash1)
	suite.Nil(err)
	suite.False(vErr.HasAny())

	storedGenesisHash, err := models.GetGenesisHashForWebnode(existingGenesisHashes)

	suite.Equal(models.StoredGenesisHash{}, storedGenesisHash)
	suite.Equal(models.NoGenHashesMessage, err.Error())
}

func (suite *ModelSuite) Test_GetGenesisHashForWebnode_none_below_webnode_count_limit() {

	existingGenesisHashes := []string{
		"abcdef11",
		"abcdef12",
		"abcdef13",
	}

	storedGenesisHash1 := models.StoredGenesisHash{
		GenesisHash:    "aaaaaa",
		FileSizeBytes:  5000,
		NumChunks:      5,
		WebnodeCount:   models.WebnodeCountLimit + 1, // over the limit
		Status:         models.StoredGenesisHashUnassigned,
		TreasureStatus: models.TreasureBuried,
	}

	vErr, err := suite.DB.ValidateAndCreate(&storedGenesisHash1)
	suite.Nil(err)
	suite.False(vErr.HasAny())

	storedGenesisHash, err := models.GetGenesisHashForWebnode(existingGenesisHashes)

	suite.Equal(models.StoredGenesisHash{}, storedGenesisHash)
	suite.Equal(models.NoGenHashesMessage, err.Error())
}

func (suite *ModelSuite) Test_GetGenesisHashForWebnode_success() {

	existingGenesisHashes := []string{
		"abcdef11",
		"abcdef12",
		"abcdef13",
	}

	genesisHash := "aaaaaa"

	storedGenesisHash1 := models.StoredGenesisHash{
		GenesisHash:    genesisHash,
		FileSizeBytes:  5000,
		NumChunks:      5,
		WebnodeCount:   0,
		Status:         models.StoredGenesisHashUnassigned,
		TreasureStatus: models.TreasureBuried,
	}

	vErr, err := suite.DB.ValidateAndCreate(&storedGenesisHash1)
	suite.Nil(err)
	suite.False(vErr.HasAny())

	storedGenesisHash, err := models.GetGenesisHashForWebnode(existingGenesisHashes)

	suite.Equal(storedGenesisHash1.GenesisHash, storedGenesisHash.GenesisHash)
	suite.Nil(err)
}

func (suite *ModelSuite) Test_GetGenesisHashForWebnode_success_return_oldest() {

	existingGenesisHashes := []string{
		"abcdef11",
		"abcdef12",
		"abcdef13",
	}

	newerGenesisHash := "aaaaaa"
	olderGenesisHash := "bbbbbb"

	storedGenesisHash1 := models.StoredGenesisHash{
		GenesisHash:    newerGenesisHash,
		FileSizeBytes:  5000,
		NumChunks:      5,
		WebnodeCount:   0,
		Status:         models.StoredGenesisHashUnassigned,
		TreasureStatus: models.TreasureBuried,
	}

	storedGenesisHash2 := models.StoredGenesisHash{
		GenesisHash:    olderGenesisHash,
		FileSizeBytes:  5000,
		NumChunks:      5,
		WebnodeCount:   0,
		Status:         models.StoredGenesisHashUnassigned,
		TreasureStatus: models.TreasureBuried,
	}

	vErr, err := suite.DB.ValidateAndCreate(&storedGenesisHash1)
	suite.Nil(err)
	suite.False(vErr.HasAny())

	vErr, err = suite.DB.ValidateAndCreate(&storedGenesisHash2)
	suite.Nil(err)
	suite.False(vErr.HasAny())

	// forcibly updated the "created_at" time for the second stored_genesis_hash so it will be
	// older
	err = suite.DB.RawQuery("UPDATE stored_genesis_hashes SET created_at = ? WHERE genesis_hash = ?",
		time.Now().Add(-20*time.Second), storedGenesisHash2.GenesisHash).All(&[]models.StoredGenesisHash{})
	suite.Nil(err)

	storedGenesisHash, err := models.GetGenesisHashForWebnode(existingGenesisHashes)

	suite.Equal(olderGenesisHash, storedGenesisHash.GenesisHash)
	suite.Nil(err)
}

func (suite *ModelSuite) Test_CheckIfGenesisHashExists_exists() {

	storedGenesisHash1 := models.StoredGenesisHash{
		GenesisHash:    "aaaaaa",
		FileSizeBytes:  5000,
		NumChunks:      5,
		WebnodeCount:   0,
		Status:         models.StoredGenesisHashAssigned, // assigned already
		TreasureStatus: models.TreasureBuried,
	}

	vErr, err := suite.DB.ValidateAndCreate(&storedGenesisHash1)
	suite.Nil(err)
	suite.False(vErr.HasAny())

	genesisHashExists, err := models.CheckIfGenesisHashExists("aaaaaa")
	suite.Nil(err)
	suite.Equal(true, genesisHashExists)
}

func (suite *ModelSuite) Test_CheckIfGenesisHashExists_does_not_exist() {

	storedGenesisHash1 := models.StoredGenesisHash{
		GenesisHash:    "aaaaaa",
		FileSizeBytes:  5000,
		NumChunks:      5,
		WebnodeCount:   0,
		Status:         models.StoredGenesisHashAssigned, // assigned already
		TreasureStatus: models.TreasureBuried,
	}

	vErr, err := suite.DB.ValidateAndCreate(&storedGenesisHash1)
	suite.Nil(err)
	suite.False(vErr.HasAny())

	genesisHashExists, err := models.CheckIfGenesisHashExists("ffffff")
	suite.Nil(err)
	suite.Equal(false, genesisHashExists)
}

func (suite *ModelSuite) Test_CheckIfGenesisHashExistsAndIsBuried_not_exists() {

	genesisHashExists, buried, err := models.CheckIfGenesisHashExistsAndIsBuried("bbbbbb")
	suite.Nil(err)
	suite.Equal(false, genesisHashExists)
	suite.Equal(false, buried)
}

func (suite *ModelSuite) Test_CheckIfGenesisHashExistsAndIsBuried_exists_not_buried() {

	storedGenesisHash1 := models.StoredGenesisHash{
		GenesisHash:    "aaaaaa",
		FileSizeBytes:  5000,
		NumChunks:      5,
		WebnodeCount:   0,
		Status:         models.StoredGenesisHashAssigned, // assigned already
		TreasureStatus: models.TreasurePending,
	}

	vErr, err := suite.DB.ValidateAndCreate(&storedGenesisHash1)
	suite.Nil(err)
	suite.False(vErr.HasAny())

	genesisHashExists, buried, err := models.CheckIfGenesisHashExistsAndIsBuried("aaaaaa")
	suite.Nil(err)
	suite.Equal(true, genesisHashExists)
	suite.Equal(false, buried)
}

func (suite *ModelSuite) Test_CheckIfGenesisHashExistsAndIsBuried_exists_and_buried() {

	storedGenesisHash1 := models.StoredGenesisHash{
		GenesisHash:    "aaaaaa",
		FileSizeBytes:  5000,
		NumChunks:      5,
		WebnodeCount:   0,
		Status:         models.StoredGenesisHashAssigned, // assigned already
		TreasureStatus: models.TreasureBuried,
	}

	vErr, err := suite.DB.ValidateAndCreate(&storedGenesisHash1)
	suite.Nil(err)
	suite.False(vErr.HasAny())

	genesisHashExists, buried, err := models.CheckIfGenesisHashExistsAndIsBuried("aaaaaa")
	suite.Nil(err)
	suite.Equal(true, genesisHashExists)
	suite.Equal(true, buried)
}

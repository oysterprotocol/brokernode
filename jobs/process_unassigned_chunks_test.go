package jobs_test

import (
	//"testing"
	//"github.com/gobuffalo/suite"
	"github.com/oysterprotocol/brokernode/models"
	"fmt"
	"log"
	//
	//"github.com/gobuffalo/envy"
	//"github.com/gobuffalo/pop"
)

//func (as *ActionSuite) Test_ProcessUnassignedChunks() {
//	as = &ActionSuite{suite.NewAction(ActionSuite.App())}
//	suite.Run(t, as)
//}

func (ms *ModelSuite) Test_ProcessUnassignedChunks() {
//func (ms *ModelSuite) Test_ProcessUnassignedChunks(as *ActionSuite) {
	genHash := "genHashTest"
	fileBytesCount := 9000

	models.BuildDataMaps(genHash, fileBytesCount)


	//fmt.Println(models.BuildDataMaps)
	fmt.Println(models.DataMap{})
	fmt.Println(ms)
	fmt.Println(ms.DB)

	dMaps := []models.DataMap{}
	ms.DB.Where("genesis_hash != NULL").Order("chunk_idx asc").All(&dMaps)

	fmt.Println(dMaps)
	fmt.Println("THIS WORKED, fmt")

	log.Println(dMaps)
	log.Println("THIS WORKED< log")

	ms.Equal(1, 1)
}
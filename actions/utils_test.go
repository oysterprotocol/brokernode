package actions

func (as *ActionSuite) Test_GenerateInsertedIndexesForPearl_BadFileSize() {
	indexes := generateInsertedIndexesForPearl(-1)

	as.Nil(indexes)
}

func (as *ActionSuite) Test_GenerateInsertedIndexesForPearl_SmallFileSize() {
	indexes := generateInsertedIndexesForPearl(100)

	as.True(len(indexes) == 1)
	as.True(indexes[0] == 0 || indexes[0] == 1)
}

func (as *ActionSuite) Test_GenerateInsertedIndexesForPearl_LargeFileSize() {
	// Test on 2.6GB
	indexes := generateInsertedIndexesForPearl(int(2.6 * fileSectorInByte))

	as.True(len(indexes) == 3)
	as.True(indexes[0] > 0 && indexes[0] < int(fileSectorInChunkSize))
	as.True(indexes[1] > int(fileSectorInChunkSize) && indexes[1] < 2*int(fileSectorInChunkSize))
	as.True(indexes[2] > 2*int(fileSectorInChunkSize) && indexes[2] < int(2.6*fileSectorInChunkSize))
}

func (as *ActionSuite) Test_GenerateInsertedIndexesForPearl_MediumFileSize() {
	// Test on 2MB
	indexes := generateInsertedIndexesForPearl(int(2000 * fileChunkSizeInByte))

	as.True(len(indexes) == 1)
	as.True(indexes[0] > 0 && indexes[0] < 2000)
}

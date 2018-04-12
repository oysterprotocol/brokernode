package actions

func (as *ActionSuite) Test_GenerateInsertedIndexesForPearl_BadFileSize() {
	indexes := generateInsertedIndexesForPearl(-1)

	as.True(len(indexes) == 0)
}

func (as *ActionSuite) Test_GenerateInsertedIndexesForPearl_SmallFileSize() {
	indexes := generateInsertedIndexesForPearl(100)

	as.True(len(indexes) == 1)
	as.True(indexes[0] == 0 || indexes[0] == 1)
}

func (as *ActionSuite) Test_GenerateInsertedIndexesForPearl_LargeFileSize() {
	// Test on 2.6GB
	indexes := generateInsertedIndexesForPearl(int(2.6 * fileSectorInChunkSize * fileChunkSizeInByte))

	as.True(len(indexes) == 3)
	as.True(indexes[0] >= 0 && indexes[0] < fileSectorInChunkSize)
	as.True(indexes[1] >= 0 && indexes[1] < fileSectorInChunkSize)
	as.True(indexes[2] >= 0 && indexes[2] < int(0.6*fileSectorInChunkSize)+3)
}

func (as *ActionSuite) Test_GenerateInsertedIndexesForPearl_MediumFileSize() {
	// Test on 2MB
	indexes := generateInsertedIndexesForPearl(int(2000 * fileChunkSizeInByte))

	as.True(len(indexes) == 1)
	as.True(indexes[0] >= 0 && indexes[0] < 2001)
}

func (as *ActionSuite) Test_GenerateInsertIndexesForPearl_ExtendedToNextSector() {
	// Test on 2.999998GB, by adding Pearls, it will extend to 3.000001GB
	indexes := generateInsertedIndexesForPearl(int(2.999998 * fileSectorInChunkSize * fileChunkSizeInByte))

	as.True(len(indexes) == 4)
	as.True(indexes[3] == 0 || indexes[3] == 1)
}

func (as *ActionSuite) Test_GenerateInsertIndexesForPearl_NotNeedToExtendedToNextSector() {
	// Test on 2.999997GB, by adding Pearls, it will extend to 3GB
	indexes := generateInsertedIndexesForPearl(int(2.999997 * fileSectorInChunkSize * fileChunkSizeInByte))

	as.True(len(indexes) == 3)
	as.True(indexes[2] >= 0 && indexes[2] < fileSectorInChunkSize)
}

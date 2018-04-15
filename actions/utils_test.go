package actions

func (as *ActionSuite) Test_TransformIndexWithBuriedIndexes_NoBuriedIndexes() {
	index := TransformIndexWithBuriedIndexes(10, []int{})

	as.Equal(10, index)
}

func (as *ActionSuite) Test_TransformIndexWithBuriedIndexes_NoChangeOnIndex() {
	index := TransformIndexWithBuriedIndexes(10, []int{20, 1})

	as.Equal(10, index)
}

func (as *ActionSuite) Test_TransformIndexWithBuriedIndexes_EqualIndex() {
	index := TransformIndexWithBuriedIndexes(20, []int{20, 1})

	as.Equal(21, index)
}

func (as *ActionSuite) Test_TransformIndexWithBuriedIndexes_WithinFirstSector() {
	index := TransformIndexWithBuriedIndexes(FileSectorInChunkSize-2, []int{20, 0})

	as.Equal(FileSectorInChunkSize-1, index)
}

func (as *ActionSuite) Test_TransformIndexWithBuriedIndexes_ToAnotherSector() {
	index := TransformIndexWithBuriedIndexes(FileSectorInChunkSize-1, []int{20, 0})

	as.Equal(FileSectorInChunkSize+1, index)
}

func (as *ActionSuite) Test_TransformIndexWithBuriedIndexes_TreasureAsLastIndex() {
	index := TransformIndexWithBuriedIndexes(FileSectorInChunkSize-1, []int{FileSectorInChunkSize - 1, 0})

	as.Equal(FileSectorInChunkSize+1, index)
}

func (as *ActionSuite) Test_TransformIndexWithBuriedIndexes_LastSector() {
	index := TransformIndexWithBuriedIndexes(FileSectorInChunkSize*2-3, []int{0, 0})

	as.Equal(FileSectorInChunkSize*2-1, index)
}

func (as *ActionSuite) Test_GenerateInsertedIndexesForPearl_BadFileSize() {
	indexes := GenerateInsertedIndexesForPearl(-1)

	as.True(len(indexes) == 0)
}

func (as *ActionSuite) Test_GenerateInsertedIndexesForPearl_SmallFileSize() {
	indexes := GenerateInsertedIndexesForPearl(100)

	as.True(len(indexes) == 1)
	as.True(indexes[0] == 0 || indexes[0] == 1)
}

func (as *ActionSuite) Test_GenerateInsertedIndexesForPearl_LargeFileSize() {
	// Test on 2.6GB
	indexes := GenerateInsertedIndexesForPearl(int(2.6 * FileSectorInChunkSize * FileChunkSizeInByte))

	as.True(len(indexes) == 3)
	as.True(indexes[0] >= 0 && indexes[0] < FileSectorInChunkSize)
	as.True(indexes[1] >= 0 && indexes[1] < FileSectorInChunkSize)
	as.True(indexes[2] >= 0 && indexes[2] < int(0.6*FileSectorInChunkSize)+3)
}

func (as *ActionSuite) Test_GenerateInsertedIndexesForPearl_MediumFileSize() {
	// Test on 2MB
	indexes := GenerateInsertedIndexesForPearl(int(2000 * FileChunkSizeInByte))

	as.True(len(indexes) == 1)
	as.True(indexes[0] >= 0 && indexes[0] < 2001)
}

func (as *ActionSuite) Test_GenerateInsertIndexesForPearl_ExtendedToNextSector() {
	// Test on 2.999998GB, by adding Pearls, it will extend to 3.000001GB
	indexes := GenerateInsertedIndexesForPearl(int(2.999998 * FileSectorInChunkSize * FileChunkSizeInByte))

	as.True(len(indexes) == 4)
	as.True(indexes[3] == 0 || indexes[3] == 1)
}

func (as *ActionSuite) Test_GenerateInsertIndexesForPearl_NotNeedToExtendedToNextSector() {
	// Test on 2.999997GB, by adding Pearls, it will extend to 3GB
	indexes := GenerateInsertedIndexesForPearl(int(2.999997 * FileSectorInChunkSize * FileChunkSizeInByte))

	as.True(len(indexes) == 3)
	as.True(indexes[2] >= 0 && indexes[2] < FileSectorInChunkSize)
}

func (as *ActionSuite) Test_MergedIndexes_EmptyIndexes() {
	_, err := MergeIndexes([]int{}, nil)

	as.Error(err)
}

func (as *ActionSuite) Test_MergedIndexes_OneNonEmptyIndexes() {
	_, err := MergeIndexes(nil, []int{1, 2})

	as.Error(err)
}

func (as *ActionSuite) Test_MergeIndexes_SameSize() {
	indexes, _ := MergeIndexes([]int{1, 2, 3}, []int{1, 2, 3})

	as.True(len(indexes) == 3)
}

func (as *ActionSuite) Test_GetTreasureIdxMap_ValidInput() {
	idxMap := GetTreasureIdxMap([]int{1}, []int{2})

	as.True(idxMap.Valid)
}

func (as *ActionSuite) Test_GetTreasureIdxMap_InvalidInput() {
	idxMap := GetTreasureIdxMap([]int{1}, []int{1, 2})

	as.Equal("", idxMap.String)
	as.False(idxMap.Valid)
}

func (as *ActionSuite) Test_IntsJoin_NoInts() {
	v := IntsJoin(nil, " ")

	as.True(v == "")
}

func (as *ActionSuite) Test_IntsJoin_ValidInts() {
	v := IntsJoin([]int{1, 2, 3}, "_")

	as.True(v == "1_2_3")
}

func (as *ActionSuite) Test_IntsJoin_SingleInt() {
	v := IntsJoin([]int{1}, "_")

	as.True(v == "1")
}

func (as *ActionSuite) Test_IntsJoin_InvalidDelim() {
	v := IntsJoin([]int{1, 2, 3}, "")

	as.True(v == "123")
}

func (as *ActionSuite) Test_IntsSplit_InvalidString() {
	v := IntsSplit("abc", " ")

	as.True(v == nil)
}

func (as *ActionSuite) Test_IntsSplit_ValidInput() {
	v := IntsSplit("1_2_3", "_")

	compareIntsArray(as, v, []int{1, 2, 3})
}

func (as *ActionSuite) Test_IntsSplit_SingleInt() {
	v := IntsSplit("1", "_")

	compareIntsArray(as, v, []int{1})
}

func (as *ActionSuite) Test_IntsSplit_MixIntString() {
	v := IntsSplit("1_a_2", "_")

	compareIntsArray(as, v, []int{1, 2})
}

func (as *ActionSuite) Test_IntsSplit_EmptyString() {
	v := IntsSplit("", "_")

	compareIntsArray(as, v, []int{})
}

// Private helper methods
func compareIntsArray(as *ActionSuite, a []int, b []int) {
	as.True(len(a) == len(b))

	for i := 0; i < len(a); i++ {
		as.True(a[i] == b[i])
	}
}

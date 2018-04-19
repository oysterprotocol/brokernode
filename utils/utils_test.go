package oyster_utils

import (
	"testing"
)

func Test_ConvertToByte_1TryBtye(t *testing.T) {
	v := ConvertToByte(1)

	assertTrue(v == 1, t, "")
}

func Test_ConvertToByte_2TryBtye(t *testing.T) {
	v := ConvertToByte(2)

	assertTrue(v == 1, t, "")
}

func Test_ConvertToByte_2186TryByte(t *testing.T) {
	v := ConvertToByte(2186)

	assertTrue(v == 1000, t, "")
}

func Test_ConvertToTryByte_1Byte(t *testing.T) {
	v := ConvertToTryByte(1)

	assertTrue(v == 3, t, "")
}

func Test_ConvertToTryByte_1000Bytes(t *testing.T) {
	v := ConvertToTryByte(1000)

	assertTrue(v == 2187, t, "")
}

func Test_GetTotalFileChunkIncludingBuriedPearls_SmallFileSize(t *testing.T) {
	v := GetTotalFileChunkIncludingBuriedPearls(10)

	assertTrue(v == 2, t, "")
}

func Test_GetTotalFileChunkIncludingBuriedPearls_MediaFileSize(t *testing.T) {
	v := GetTotalFileChunkIncludingBuriedPearls(FileChunkSizeInByte)

	assertTrue(v == 2, t, "")
}

func Test_GetTotalFileChunkIncludingBuriedPearls_BigFileSize(t *testing.T) {
	v := GetTotalFileChunkIncludingBuriedPearls(FileChunkSizeInByte * FileSectorInChunkSize * 2)

	assertTrue(v == 2*FileSectorInChunkSize+3, t, "")
}

func Test_TransformIndexWithBuriedIndexes_NoBuriedIndexes(t *testing.T) {
	index := TransformIndexWithBuriedIndexes(10, []int{})

	assertTrue(index == 10, t, "No change on the index")
}

func Test_TransformIndexWithBuriedIndexes_NoChangeOnIndex(t *testing.T) {
	index := TransformIndexWithBuriedIndexes(10, []int{20, 1})

	assertTrue(index == 10, t, "No change on the index")
}

func Test_TransformIndexWithBuriedIndexes_EqualIndex(t *testing.T) {
	index := TransformIndexWithBuriedIndexes(20, []int{20, 1})

	assertTrue(index == 21, t, "Increase index by 1")
}

func Test_TransformIndexWithBuriedIndexes_WithinFirstSector(t *testing.T) {
	index := TransformIndexWithBuriedIndexes(FileSectorInChunkSize-2, []int{20, 0})

	assertTrue(index == FileSectorInChunkSize-1, t, "Increase index by 1")
}

func Test_TransformIndexWithBuriedIndexes_ToAnotherSector(t *testing.T) {
	index := TransformIndexWithBuriedIndexes(FileSectorInChunkSize-1, []int{20, 0})

	assertTrue(index == FileSectorInChunkSize+1, t, "Increasee index by 2")
}

func Test_TransformIndexWithBuriedIndexes_TreasureAsLastIndex(t *testing.T) {
	index := TransformIndexWithBuriedIndexes(FileSectorInChunkSize-1, []int{FileSectorInChunkSize - 1, 0})

	assertTrue(index == FileSectorInChunkSize+1, t, "Increase index by 2")
}

func Test_TransformIndexWithBuriedIndexes_LastSector(t *testing.T) {
	index := TransformIndexWithBuriedIndexes(FileSectorInChunkSize*2-3, []int{0, 0})

	assertTrue(index == FileSectorInChunkSize*2-1, t, "Increase index by 2")
}

func Test_GenerateInsertedIndexesForPearl_BadFileSize(t *testing.T) {
	indexes := GenerateInsertedIndexesForPearl(-1)

	assertTrue(len(indexes) == 0, t, "Len must equal to 0")
}

func Test_GenerateInsertedIndexesForPearl_SmallFileSize(t *testing.T) {
	indexes := GenerateInsertedIndexesForPearl(100)

	assertTrue(len(indexes) == 1, t, "Len must equal to 1")
	assertTrue(indexes[0] == 0 || indexes[0] == 1, t, "Value must be either 0 or 1")
}

func Test_GenerateInsertedIndexesForPearl_LargeFileSize(t *testing.T) {
	// Test on 2.6GB
	indexes := GenerateInsertedIndexesForPearl(int(2.6 * FileSectorInChunkSize * FileChunkSizeInByte))

	assertTrue(len(indexes) == 3, t, "")
	assertTrue(indexes[0] >= 0 && indexes[0] < FileSectorInChunkSize, t, "")
	assertTrue(indexes[1] >= 0 && indexes[1] < FileSectorInChunkSize, t, "")
	assertTrue(indexes[2] >= 0 && indexes[2] < int(0.6*FileSectorInChunkSize)+3, t, "")
}

func Test_GenerateInsertedIndexesForPearl_MediumFileSize(t *testing.T) {
	// Test on 2MB
	indexes := GenerateInsertedIndexesForPearl(int(2000 * FileChunkSizeInByte))

	assertTrue(len(indexes) == 1, t, "")
	assertTrue(indexes[0] >= 0 && indexes[0] < 2001, t, "Must within range of [0, 2001)")
}

func Test_GenerateInsertIndexesForPearl_ExtendedToNextSector(t *testing.T) {
	// Test on 2.999998GB, by adding Pearls, it will extend to 3.000001GB
	indexes := GenerateInsertedIndexesForPearl(int(2.999998 * FileSectorInChunkSize * FileChunkSizeInByte))

	assertTrue(len(indexes) == 4, t, "")
	assertTrue(indexes[3] == 0 || indexes[3] == 1, t, "Must either 0 or 1")
}

func Test_GenerateInsertIndexesForPearl_NotNeedToExtendedToNextSector(t *testing.T) {
	// Test on 2.999997GB, by adding Pearls, it will extend to 3GB
	indexes := GenerateInsertedIndexesForPearl(int(2.999997 * FileSectorInChunkSize * FileChunkSizeInByte))

	assertTrue(len(indexes) == 3, t, "")
	assertTrue(indexes[2] >= 0 && indexes[2] < FileSectorInChunkSize, t, "Must within range of [0, FileSectorInChunkSize)")
}

func Test_MergedIndexes_EmptyIndexes(t *testing.T) {
	_, err := mergeIndexes([]int{}, nil)

	assertTrue(err != nil, t, "")
}

func Test_MergedIndexes_OneNonEmptyIndexes(t *testing.T) {
	_, err := mergeIndexes(nil, []int{1, 2})

	assertTrue(err != nil, t, "Must result an error")
}

func Test_MergeIndexes_SameSize(t *testing.T) {
	indexes, _ := mergeIndexes([]int{1, 2, 3}, []int{1, 2, 3})

	assertTrue(len(indexes) == 3, t, "Must result an error")
}

func Test_GetTreasureIdxMap_ValidInput(t *testing.T) {
	idxMap := GetTreasureIdxMap([]int{1}, []int{2})

	assertTrue(idxMap.Valid, t, "")
}

func Test_GetTreasureIdxMap_InvalidInput(t *testing.T) {
	idxMap := GetTreasureIdxMap([]int{1}, []int{1, 2})

	assertTrue(idxMap.String == "", t, "")
	assertTrue(!idxMap.Valid, t, "")
}

func Test_IntsJoin_NoInts(t *testing.T) {
	v := IntsJoin(nil, " ")

	assertTrue(v == "", t, "")
}

func Test_IntsJoin_ValidInts(t *testing.T) {
	v := IntsJoin([]int{1, 2, 3}, "_")

	assertTrue(v == "1_2_3", t, "")
}

func Test_IntsJoin_SingleInt(t *testing.T) {
	v := IntsJoin([]int{1}, "_")

	assertTrue(v == "1", t, "")
}

func Test_IntsJoin_InvalidDelim(t *testing.T) {
	v := IntsJoin([]int{1, 2, 3}, "")

	assertTrue(v == "123", t, "")
}

func Test_IntsSplit_InvalidString(t *testing.T) {
	v := IntsSplit("abc", " ")

	assertTrue(v == nil, t, "Result as nil")
}

func Test_IntsSplit_ValidInput(t *testing.T) {
	v := IntsSplit("1_2_3", "_")

	compareIntsArray(t, v, []int{1, 2, 3})
}

func Test_IntsSplit_SingleInt(t *testing.T) {
	v := IntsSplit("1", "_")

	compareIntsArray(t, v, []int{1})
}

func Test_IntsSplit_MixIntString(t *testing.T) {
	v := IntsSplit("1_a_2", "_")

	compareIntsArray(t, v, []int{1, 2})
}

func Test_IntsSplit_EmptyString(t *testing.T) {
	v := IntsSplit("", "_")

	compareIntsArray(t, v, []int{})
}

// Private helper methods
func compareIntsArray(t *testing.T, a []int, b []int) {
	assertTrue(len(a) == len(b), t, "a and b must have the same len")

	for i := 0; i < len(a); i++ {
		assertTrue(a[i] == b[i], t, "a and b value are different")
	}
}

func assertTrue(v bool, t *testing.T, desc string) {
	if !v {
		t.Error(desc)
	}
}

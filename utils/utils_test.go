package oyster_utils_test

import (
	"math/big"
	"testing"

	"github.com/gobuffalo/pop/nulls"
	"github.com/oysterprotocol/brokernode/utils"
)

func Test_ConvertToByte_1Trytes(t *testing.T) {
	v := oyster_utils.ConvertToByte(1)

	oyster_utils.AssertTrue(v == 1, t, "")
}

func Test_ConvertToByte_2Trytes(t *testing.T) {
	v := oyster_utils.ConvertToByte(2)

	oyster_utils.AssertTrue(v == 1, t, "")
}

func Test_ConvertToTrytes_1Byte(t *testing.T) {
	v := oyster_utils.ConvertToTrytes(1)

	oyster_utils.AssertTrue(v == 2, t, "")
}

func Test_GetTotalFileChunkIncludingBuriedPearlsUsingFileSize_SmallFileSize(t *testing.T) {
	v := oyster_utils.GetTotalFileChunkIncludingBuriedPearlsUsingFileSize(10)

	oyster_utils.AssertTrue(v == 2, t, "")
}

func Test_GetTotalFileChunkIncludingBuriedPearlsUsingFileSize_MediaFileSize(t *testing.T) {
	v := oyster_utils.GetTotalFileChunkIncludingBuriedPearlsUsingFileSize(oyster_utils.FileChunkSizeInByte)

	oyster_utils.AssertTrue(v == 2, t, "")
}

func Test_GetTotalFileChunkIncludingBuriedPearlsUsingFileSize_BigFileSize(t *testing.T) {
	v := oyster_utils.GetTotalFileChunkIncludingBuriedPearlsUsingFileSize(oyster_utils.FileChunkSizeInByte * oyster_utils.FileSectorInChunkSize * 2)

	oyster_utils.AssertTrue(v == 2*oyster_utils.FileSectorInChunkSize+3, t, "")
}

func Test_GetTotalFileChunkIncludingBuriedPearlsUsingNumChunks_SmallFileSize(t *testing.T) {
	v := oyster_utils.GetTotalFileChunkIncludingBuriedPearlsUsingNumChunks(10)

	oyster_utils.AssertTrue(v == 11, t, "")
}

func Test_GetTotalFileChunkIncludingBuriedPearlsUsingNumChunks_LargeFileSize(t *testing.T) {
	v := oyster_utils.GetTotalFileChunkIncludingBuriedPearlsUsingNumChunks((oyster_utils.FileSectorInChunkSize * 10) + 1)

	oyster_utils.AssertTrue(v == (oyster_utils.FileSectorInChunkSize*10)+1+11, t, "")
}

func Test_GetTotalFileChunkIncludingBuriedPearlsUsingNumChunks_HugeFileSize(t *testing.T) {
	v := oyster_utils.GetTotalFileChunkIncludingBuriedPearlsUsingNumChunks((oyster_utils.FileSectorInChunkSize * 60) + 500)

	oyster_utils.AssertTrue(v == (oyster_utils.FileSectorInChunkSize*60)+500+61, t, "")
}

func Test_TransformIndexWithBuriedIndexes_NoBuriedIndexes(t *testing.T) {
	index := oyster_utils.TransformIndexWithBuriedIndexes(10, []int{})

	oyster_utils.AssertTrue(index == 10, t, "No change on the index")
}

func Test_TransformIndexWithBuriedIndexes_NoChangeOnIndex(t *testing.T) {
	index := oyster_utils.TransformIndexWithBuriedIndexes(10, []int{20, 1})

	oyster_utils.AssertTrue(index == 10, t, "No change on the index")
}

func Test_TransformIndexWithBuriedIndexes_EqualIndex(t *testing.T) {
	index := oyster_utils.TransformIndexWithBuriedIndexes(20, []int{20, 1})

	oyster_utils.AssertTrue(index == 21, t, "Increase index by 1")
}

func Test_TransformIndexWithBuriedIndexes_WithinFirstSector(t *testing.T) {
	index := oyster_utils.TransformIndexWithBuriedIndexes(oyster_utils.FileSectorInChunkSize-2, []int{20, 0})

	oyster_utils.AssertTrue(index == oyster_utils.FileSectorInChunkSize-1, t, "Increase index by 1")
}

func Test_TransformIndexWithBuriedIndexes_ToAnotherSector(t *testing.T) {
	index := oyster_utils.TransformIndexWithBuriedIndexes(oyster_utils.FileSectorInChunkSize-1, []int{20, 0})

	oyster_utils.AssertTrue(index == oyster_utils.FileSectorInChunkSize+1, t, "Increasee index by 2")
}

func Test_TransformIndexWithBuriedIndexes_TreasureAsLastIndex(t *testing.T) {
	index := oyster_utils.TransformIndexWithBuriedIndexes(oyster_utils.FileSectorInChunkSize-1, []int{oyster_utils.FileSectorInChunkSize - 1, 0})

	oyster_utils.AssertTrue(index == oyster_utils.FileSectorInChunkSize+1, t, "Increase index by 2")
}

func Test_TransformIndexWithBuriedIndexes_LastSector(t *testing.T) {
	index := oyster_utils.TransformIndexWithBuriedIndexes(oyster_utils.FileSectorInChunkSize*2-3, []int{0, 0})

	oyster_utils.AssertTrue(index == oyster_utils.FileSectorInChunkSize*2-1, t, "Increase index by 2")
}

func Test_GenerateInsertedIndexesForPearl_BadFileSize(t *testing.T) {
	indexes := oyster_utils.GenerateInsertedIndexesForPearl(-1)

	oyster_utils.AssertTrue(len(indexes) == 0, t, "Len must equal to 0")
}

func Test_GenerateInsertedIndexesForPearl_SmallFileSize(t *testing.T) {
	indexes := oyster_utils.GenerateInsertedIndexesForPearl(100)

	oyster_utils.AssertTrue(len(indexes) == 1, t, "Len must equal to 1")
	oyster_utils.AssertTrue(indexes[0] == 0 || indexes[0] == 1, t, "Value must be either 0 or 1")
}

func Test_GenerateInsertedIndexesForPearl_LargeFileSize(t *testing.T) {
	// Test on 2.6GB
	indexes := oyster_utils.GenerateInsertedIndexesForPearl(int(2.6 * oyster_utils.FileSectorInChunkSize * oyster_utils.FileChunkSizeInByte))

	oyster_utils.AssertTrue(len(indexes) == 3, t, "")
	oyster_utils.AssertTrue(indexes[0] >= 0 && indexes[0] < oyster_utils.FileSectorInChunkSize, t, "")
	oyster_utils.AssertTrue(indexes[1] >= 0 && indexes[1] < oyster_utils.FileSectorInChunkSize, t, "")
	oyster_utils.AssertTrue(indexes[2] >= 0 && indexes[2] < int(0.6*oyster_utils.FileSectorInChunkSize)+3, t, "")
}

func Test_GenerateInsertedIndexesForPearl_MediumFileSize(t *testing.T) {
	// Test on 2MB
	indexes := oyster_utils.GenerateInsertedIndexesForPearl(int(2000 * oyster_utils.FileChunkSizeInByte))

	oyster_utils.AssertTrue(len(indexes) == 1, t, "")
	oyster_utils.AssertTrue(indexes[0] >= 0 && indexes[0] < 2001, t, "Must within range of [0, 2001)")
}

func Test_GenerateInsertIndexesForPearl_ExtendedToNextSector(t *testing.T) {
	// Test on 2.999998GB, by adding Pearls, it will extend to 3.000001GB
	indexes := oyster_utils.GenerateInsertedIndexesForPearl(int(2.999998 * oyster_utils.FileSectorInChunkSize * oyster_utils.FileChunkSizeInByte))

	oyster_utils.AssertTrue(len(indexes) == 4, t, "")
	oyster_utils.AssertTrue(indexes[3] == 0 || indexes[3] == 1, t, "Must either 0 or 1")
}

func Test_GenerateInsertIndexesForPearl_NotNeedToExtendedToNextSector(t *testing.T) {
	// Test on 2.999997GB, by adding Pearls, it will extend to 3GB
	indexes := oyster_utils.GenerateInsertedIndexesForPearl(int(2.999997 * oyster_utils.FileSectorInChunkSize * oyster_utils.FileChunkSizeInByte))

	oyster_utils.AssertTrue(len(indexes) == 3, t, "")
	oyster_utils.AssertTrue(indexes[2] >= 0 && indexes[2] < oyster_utils.FileSectorInChunkSize, t, "Must within range of [0, FileSectorInChunkSize)")
}

func Test_MergedIndexes_EmptyIndexes(t *testing.T) {
	_, err := oyster_utils.MergeIndexes([]int{}, nil)

	oyster_utils.AssertError(err, t, "")
}

func Test_MergedIndexes_OneNonEmptyIndexes(t *testing.T) {
	_, err := oyster_utils.MergeIndexes(nil, []int{1, 2})

	oyster_utils.AssertError(err, t, "Must result an error")
}

func Test_MergeIndexes_SameSize(t *testing.T) {
	indexes, _ := oyster_utils.MergeIndexes([]int{1, 2, 3}, []int{1, 2, 3})

	oyster_utils.AssertTrue(len(indexes) == 3, t, "Must result an error")
}

func Test_GetTreasureIdxMap_ValidInput(t *testing.T) {
	idxMap := oyster_utils.GetTreasureIdxMap([]int{1}, []int{2})

	oyster_utils.AssertTrue(idxMap.Valid, t, "")
}

func Test_GetTreasureIdxMap_InvalidInput(t *testing.T) {
	idxMap := oyster_utils.GetTreasureIdxMap([]int{1}, []int{1, 2})

	oyster_utils.AssertStringEqual(idxMap.String, "", t)
	oyster_utils.AssertTrue(!idxMap.Valid, t, "")
}

func Test_GetTreasureIdxIndexes_InvalidInput(t *testing.T) {
	indexes := oyster_utils.GetTreasureIdxIndexes(nulls.String{"", false})

	oyster_utils.AssertTrue(len(indexes) == 0, t, "")
}

func Test_GetTreasureIdxIndexes_ValidInput(t *testing.T) {
	indexes := oyster_utils.GetTreasureIdxIndexes(nulls.String{"1_1_1", true})
	oyster_utils.AssertTrue(len(indexes) == 3, t, "")
	oyster_utils.AssertTrue(indexes[0] == 1, t, "")

	oyster_utils.AssertTrue(indexes[1] == 1, t, "")
	oyster_utils.AssertTrue(indexes[2] == 1, t, "")
}

func Test_IntsJoin_NoInts(t *testing.T) {
	v := oyster_utils.IntsJoin(nil, " ")

	oyster_utils.AssertStringEqual(v, "", t)
}

func Test_IntsJoin_ValidInts(t *testing.T) {
	v := oyster_utils.IntsJoin([]int{1, 2, 3}, "_")

	oyster_utils.AssertStringEqual(v, "1_2_3", t)
}

func Test_IntsJoin_SingleInt(t *testing.T) {
	v := oyster_utils.IntsJoin([]int{1}, "_")

	oyster_utils.AssertStringEqual(v, "1", t)
}

func Test_IntsJoin_InvalidDelim(t *testing.T) {
	v := oyster_utils.IntsJoin([]int{1, 2, 3}, "")

	oyster_utils.AssertStringEqual(v, "123", t)
}

func Test_IntsSplit_InvalidString(t *testing.T) {
	v := oyster_utils.IntsSplit("abc", " ")

	oyster_utils.AssertTrue(v == nil, t, "Result as nil")
}

func Test_IntsSplit_ValidInput(t *testing.T) {
	v := oyster_utils.IntsSplit("1_2_3", "_")

	compareIntsArray(t, v, []int{1, 2, 3})
}

func Test_IntsSplit_SingleInt(t *testing.T) {
	v := oyster_utils.IntsSplit("1", "_")

	compareIntsArray(t, v, []int{1})
}

func Test_IntsSplit_MixIntString(t *testing.T) {
	v := oyster_utils.IntsSplit("1_a_2", "_")

	compareIntsArray(t, v, []int{1, 2})
}

func Test_IntsSplit_EmptyString(t *testing.T) {
	v := oyster_utils.IntsSplit("", "_")

	compareIntsArray(t, v, []int{})
}

func Test_ConvertToWeiUnit(t *testing.T) {
	v := oyster_utils.ConvertToWeiUnit(big.NewFloat(0.2))

	oyster_utils.AssertTrue(v.Cmp(big.NewInt(200000000000000000)) == 0, t, "")
}

func Test_ConvertToWeiUnit_SmallValue(t *testing.T) {
	v := oyster_utils.ConvertToWeiUnit(big.NewFloat(0.000000000000000002))

	oyster_utils.AssertTrue(v.Cmp(big.NewInt(2)) == 0, t, "")
}

func Test_ConvertToWeiUnit_ConsiderAsZero(t *testing.T) {
	v := oyster_utils.ConvertToWeiUnit(big.NewFloat(0.0000000000000000002))

	oyster_utils.AssertTrue(v.Cmp(big.NewInt(0)) == 0, t, "")
}

func Test_ConvertToPrlUnit(t *testing.T) {
	v := oyster_utils.ConverFromWeiUnit(big.NewInt(200000000000000000))

	oyster_utils.AssertTrue(v.String() == big.NewFloat(.2).String(), t, "")
}

func Test_ConvertToPrlUnit_SmallValue(t *testing.T) {
	v := oyster_utils.ConverFromWeiUnit(big.NewInt(2))

	oyster_utils.AssertTrue(v.String() == big.NewFloat(.000000000000000002).String(), t, "")
}

// Private helper methods
func compareIntsArray(t *testing.T, a []int, b []int) {

	oyster_utils.AssertTrue(len(a) == len(b), t, "a and b must have the same len")
	for i := 0; i < len(a); i++ {
		oyster_utils.AssertTrue(a[i] == b[i], t, "a and b value are different")
	}
}

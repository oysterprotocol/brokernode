package oyster_utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gobuffalo/pop/nulls"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
)

const (
	// One chunk unit as represents as 1KB
	FileChunkSizeInByte = 1000

	// Number of 1KB chunk in one Sector
	FileSectorInChunkSize = 1000000

	// Separator to join []int array
	IntsJoinDelim = "_"

	// Separator to join []string array
	StringsJoinDelim = ", "
)

// ParseReqBody take a request and parses the body to the target interface.
func ParseReqBody(req *http.Request, dest interface{}) (err error) {
	body := req.Body
	defer body.Close()

	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return
	}
	err = json.Unmarshal(bodyBytes, dest)

	return
}

// ParseResBody take a request and parses the body to the target interface.
func ParseResBody(res *http.Response, dest interface{}) (err error) {
	body := res.Body
	defer body.Close()

	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return
	}
	err = json.Unmarshal(bodyBytes, dest)

	return
}

// Transforms index with correct position for insertion after considering the buried indexes.
func TransformIndexWithBuriedIndexes(index int, treasureIdxMap []int) int {
	if len(treasureIdxMap) == 0 {
		// TODO(pzhao5): Should log here about missing treasureIdxMap. Which should not happen.
		return index
	}

	// We needs to consider to each sector to save a space for Treasure, thus -1.
	sector := index / (FileSectorInChunkSize - 1)
	if (index - sector*(FileSectorInChunkSize-1)) >= treasureIdxMap[sector] {
		return index + sector + 1
	} else {
		return index + sector
	}
}

// Randomly generate a set of indexes in each sector
func GenerateInsertedIndexesForPearl(fileSizeInByte int) []int {
	var indexes []int
	if fileSizeInByte <= 0 {
		return indexes
	}

	fileSectorInByte := FileChunkSizeInByte * (FileSectorInChunkSize - 1)
	numOfSectors := int(math.Ceil(float64(fileSizeInByte) / float64(fileSectorInByte)))
	remainderOfChunks := math.Ceil(float64(fileSizeInByte)/FileChunkSizeInByte) + float64(numOfSectors)

	for i := 0; i < numOfSectors; i++ {
		rang := int(math.Min(FileSectorInChunkSize, remainderOfChunks))
		indexes = append(indexes, rand.Intn(rang))
		remainderOfChunks = remainderOfChunks - FileSectorInChunkSize
	}
	return indexes
}

// Return the IdxMap for treasure to burried
func GetTreasureIdxMap(alphaIndexes []int, betaIndexs []int) nulls.String {
	mergedIndexes, err := mergeIndexes(alphaIndexes, betaIndexs)
	var idxMap nulls.String
	if err == nil {
		idxMap = nulls.NewString(IntsJoin(mergedIndexes, IntsJoinDelim))
	} else {
		idxMap = nulls.String{"", false}
	}
	return idxMap
}

// Convert an []string array to a string.
func StringsJoin(A []string, delim string) string {
	var buffer bytes.Buffer
	for i := 0; i < len(A); i++ {
		buffer.WriteString(A[i])
		if i != len(A)-1 {
			buffer.WriteString(delim)
		}
	}

	return buffer.String()
}

// Convert an int array to a string.
func IntsJoin(a []int, delim string) string {
	var buffer bytes.Buffer
	for i := 0; i < len(a); i++ {
		buffer.WriteString(strconv.Itoa(a[i]))
		if i != len(a)-1 {
			buffer.WriteString(delim)
		}
	}
	return buffer.String()
}

// Convert an string back to int array
func IntsSplit(a string, delim string) []int {
	var ints []int
	substrings := strings.Split(a, delim)
	for i := 0; i < len(substrings); i++ {
		v, e := strconv.Atoi(substrings[i])
		if e == nil {
			ints = append(ints, v)
		}
	}
	return ints
}

// Return the total file chunk, including burying pearl
func GetTotalFileChunkIncludingBuriedPearls(fileSizeInByte int) int {
	fileSectorInByte := FileChunkSizeInByte * (FileSectorInChunkSize - 1)
	numOfSectors := int(math.Ceil(float64(fileSizeInByte) / float64(fileSectorInByte)))

	return numOfSectors + int(math.Ceil(float64(fileSizeInByte)/float64(FileChunkSizeInByte)))
}

// Private methods
// Merge 2 different indexes into 1 indexes. Computed Merged indexes
func mergeIndexes(a []int, b []int) ([]int, error) {
	var merged []int
	if len(a) == 0 && len(b) == 0 || len(a) != len(b) {
		return nil, errors.New("Invalid input")
	}

	for i := 0; i < len(a); i++ {
		// TODO(pzhao5): figure a better way to hash it.
		merged = append(merged, (a[i]+b[i])/2)
	}
	return merged, nil
}

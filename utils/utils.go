package oyster_utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/big"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/getsentry/raven-go"
	"github.com/gobuffalo/pop/nulls"
)

const (
	// 1 Byte = 2 Trytes
	ByteToTrytes = 2

	// One chunk unit as represents as 1KB
	FileChunkSizeInByte = 1024

	// Number of 1KB chunk in one Sector
	FileSectorInChunkSize = 1000000

	// Separator to join []int array
	IntsJoinDelim = "_"

	// Separator to join []string array
	StringsJoinDelim = ", "

	// These are the multipliers for PRL denominations.
	PrlInWeiUnit = 1e18
)

var logErrorTags map[string]string

func init() {
	isOysterPay := "enabled"
	if os.Getenv("OYSTER_PAYS") == "" {
		isOysterPay = "disabled"
	}
	logErrorTags = map[string]string{
		"mode":       os.Getenv("MODE"),
		"host_ip":    os.Getenv("HOST_IP"),
		"osyter_pay": isOysterPay,
	}
}

/*SetLogInfoForDatabaseUrl updates db_url for log info.*/
func SetLogInfoForDatabaseUrl(dbUrl string) {
	logErrorTags["db_url"] = dbUrl
}

// ParseReqBody take a request and parses the body to the target interface.
func ParseReqBody(req *http.Request, dest interface{}) (err error) {
	body := req.Body
	defer body.Close()

	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		fmt.Println(err)
		raven.CaptureError(err, nil)
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
		fmt.Println(err)
		raven.CaptureError(err, nil)
		return
	}
	err = json.Unmarshal(bodyBytes, dest)

	return
}

// Convert trytes to bytes

func ConvertToByte(trytes int) int {
	return int(math.Ceil(float64(trytes) / float64(ByteToTrytes)))
}

func ConvertToTrytes(bytes int) int {
	return bytes * ByteToTrytes
}

// Return the total file chunk, including burying pearl
func GetTotalFileChunkIncludingBuriedPearlsUsingFileSize(fileSizeInByte int) int {
	fileSectorInByte := FileChunkSizeInByte * (FileSectorInChunkSize - 1)
	numOfSectors := int(math.Ceil(float64(fileSizeInByte) / float64(fileSectorInByte)))

	return numOfSectors + int(math.Ceil(float64(fileSizeInByte)/float64(FileChunkSizeInByte)))
}

// Return the total file chunk, including burying pearl
func GetTotalFileChunkIncludingBuriedPearlsUsingNumChunks(numChunks int) int {
	return numChunks + int(math.Ceil(float64(numChunks)/float64(FileSectorInChunkSize)))
}

// Transforms index with correct position for insertion after considering the buried indexes.
func TransformIndexWithBuriedIndexes(index int, treasureIdxMap []int) int {
	if len(treasureIdxMap) == 0 {
		log.Println("TransformIndexWithBuriedIndexes(): treasureIdxMap as []int{} is empty")
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
	mergedIndexes, err := MergeIndexes(alphaIndexes, betaIndexs)
	var idxMap nulls.String
	if err == nil {
		idxMap = nulls.NewString(IntsJoin(mergedIndexes, IntsJoinDelim))
	} else {
		idxMap = nulls.String{"", false}
	}
	return idxMap
}

// Returns int[] for serialized nulls.String
func GetTreasureIdxIndexes(idxMap nulls.String) []int {
	if !idxMap.Valid {
		// TODO(pzhao5): add some logging here
		return []int{}
	}
	return IntsSplit(idxMap.String, IntsJoinDelim)
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

// Private methods
// Merge 2 different indexes into 1 indexes. Computed Merged indexes
func MergeIndexes(a []int, b []int) ([]int, error) {
	var merged []int
	if len(a) == 0 && len(b) == 0 || len(a) != len(b) {
		err := errors.New("Invalid input for utils.MergeIndexes. Both a []int and b []int must have the same length")
		raven.CaptureError(err, nil)
		return nil, err
	}

	for i := 0; i < len(a); i++ {
		// TODO(pzhao5): figure a better way to hash it.
		idx := (a[i] + b[i]) / 2
		if idx == 0 {
			idx = 1
		}
		merged = append(merged, idx)
	}
	return merged, nil
}

func RandSeq(length int, sequence []rune) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = sequence[rand.Intn(len(sequence))]
	}
	return string(b)
}

/* Convert PRL unit to wei unit. */
func ConvertToWeiUnit(prl *big.Float) *big.Int {
	f := new(big.Float).Mul(prl, big.NewFloat(float64(PrlInWeiUnit)))
	wei, _ := f.Int(new(big.Int)) // ignore the accuracy
	return wei
}

/* Convert wei unit to PRL unit */
func ConverFromWeiUnit(wei *big.Int) *big.Float {
	weiInFloat := new(big.Float).SetInt(wei)
	return new(big.Float).Quo(weiInFloat, big.NewFloat(float64(PrlInWeiUnit)))
}

/*LogIfError logs any error if it is not nil. Allow caller to provide additional freeform info.*/
func LogIfError(err error, extraInfo map[string]interface{}) {
	if err != nil {
		fmt.Println(err)

		if extraInfo != nil {
			raven.CaptureError(raven.WrapWithExtra(err, extraInfo), logErrorTags)
		} else {
			raven.CaptureError(err, logErrorTags)
		}
	}
}

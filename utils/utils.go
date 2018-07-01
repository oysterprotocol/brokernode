package oyster_utils

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/params"
	"github.com/gobuffalo/uuid"
	"golang.org/x/crypto/sha3"
	"io/ioutil"
	"log"
	"math"
	"math/big"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/getsentry/raven-go"
	"github.com/gobuffalo/pop/nulls"
	"github.com/gobuffalo/validate"
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

// Enable/Disalbe raven reporting.
var isRavenEnabled bool = true
var logErrorTags map[string]string

func init() {
	isRavenEnabled = os.Getenv("RAVEN_ENABLED") != "false"

	isOysterPay := "enabled"
	if os.Getenv("OYSTER_PAYS") == "" {
		isOysterPay = "disabled"
	}
	displayName := "Unknown"
	if v := os.Getenv("DISPLAY_NAME"); v != "" {
		displayName = v
	}

	logErrorTags = map[string]string{
		"mode":        os.Getenv("MODE"),
		"hostIp":      os.Getenv("HOST_IP"),
		"ethNodeUrl":  os.Getenv("ETH_NODE_URL"),
		"osyterPay":   isOysterPay,
		"displayName": displayName,
	}
}

/*IsInUnitTest returns true if it is running in test mode.*/
func IsInUnitTest() bool {
	// Check whether current is in unit test mode.
	// If provided -v, then it would be -test.v=true. By default, it would output to a log dir
	// with the following format: -test.testlogfile=/tmp/go-build797632719/b292/testlog.txt
	for _, v := range os.Args {
		if strings.HasPrefix(v, "-test.") {
			return true
		}
	}
	return false
}

/*IsRavenEnabled returns whether Raven logging is enabled or disabled.*/
func IsRavenEnabled() bool {
	return isRavenEnabled && !IsInUnitTest()
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
		LogIfError(err, map[string]interface{}{
			"HttpMethod":    req.Method,
			"Url":           req.URL,
			"Header":        req.Header,
			"ContentLength": req.ContentLength,
			"Body":          req.Body})
		return
	}
	err = json.Unmarshal(bodyBytes, dest)
	LogIfError(err, nil)
	return
}

// ParseResBody take a request and parses the body to the target interface.
func ParseResBody(res *http.Response, dest interface{}) (err error) {
	body := res.Body
	defer body.Close()

	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		LogIfError(err, nil)
		return
	}
	err = json.Unmarshal(bodyBytes, dest)
	LogIfError(err, nil)
	return
}

/*ConvertToByte converts trytes to bytes.*/
func ConvertToByte(trytes uint64) uint64 {
	return uint64(math.Ceil(float64(trytes) / float64(ByteToTrytes)))
}

/*ConvertToTrytes convert bytes to trytes.*/
func ConvertToTrytes(bytes uint64) uint64 {
	return bytes * ByteToTrytes
}

/*GetTotalFileChunkIncludingBuriedPearlsUsingFileSize returns the total file chunk, including burying pearl.*/
func GetTotalFileChunkIncludingBuriedPearlsUsingFileSize(fileSizeInByte uint64) int {
	fileSectorInByte := FileChunkSizeInByte * (FileSectorInChunkSize - 1)
	numOfSectors := int(math.Ceil(float64(fileSizeInByte) / float64(fileSectorInByte)))

	return numOfSectors + int(math.Ceil(float64(fileSizeInByte)/float64(FileChunkSizeInByte)))
}

/*GetTotalFileChunkIncludingBuriedPearlsUsingNumChunks returns the total file chunk, including burying pearl.*/
func GetTotalFileChunkIncludingBuriedPearlsUsingNumChunks(numChunks int) int {
	return numChunks + int(math.Ceil(float64(numChunks)/float64(FileSectorInChunkSize)))
}

/*TransformIndexWithBuriedIndexes transforms index with correct position for insertion after considering the buried indexes.*/
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

/*GenerateInsertedIndexesForPearl randomly generates a set of indexes in each sector.*/
func GenerateInsertedIndexesForPearl(fileSizeInByte uint64) []int {
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

/*GetTreasureIdxMap returns the IdxMap for treasure to burried.*/
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

/*GetTreasureIdxIndexes returns int[] for serialized nulls.String.*/
func GetTreasureIdxIndexes(idxMap nulls.String) []int {
	if !idxMap.Valid {
		// TODO(pzhao5): add some logging here
		return []int{}
	}
	return IntsSplit(idxMap.String, IntsJoinDelim)
}

/*StringsJoin converts an []string array to a string.*/
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

/*IntsJoin converts an int array to a string.*/
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

/*IntsSplit converts an string back to int array.*/
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
		LogIfError(err, map[string]interface{}{"aInputSize": len(a), "bInputSize": len(b)})
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

/*ConvertToWeiUnit converts PRL unit to wei unit. */
func ConvertToWeiUnit(prl *big.Float) *big.Int {
	f := new(big.Float).Mul(prl, big.NewFloat(float64(PrlInWeiUnit)))
	wei, _ := f.Int(new(big.Int)) // ignore the accuracy
	return wei
}

/*ConvertFromWeiUnit converts wei unit to PRL unit */
func ConvertFromWeiUnit(wei *big.Int) *big.Float {
	weiInFloat := new(big.Float).SetInt(wei)
	return new(big.Float).Quo(weiInFloat, big.NewFloat(float64(PrlInWeiUnit)))
}

func ConvertWeiToGwei(wei *big.Int) *big.Int {

	return new(big.Int).Quo(wei, big.NewInt(params.Shannon))
}

func ConvertGweiToWei(gwei *big.Int) *big.Int {

	return new(big.Int).Mul(gwei, big.NewInt(params.Shannon))
}

/*LogIfError logs any error if it is not nil. Allow caller to provide additional freeform info.*/
func LogIfError(err error, extraInfo map[string]interface{}) {
	if err == nil {
		return
	}

	fmt.Println(err)

	if IsRavenEnabled() {
		if extraInfo != nil {
			raven.CaptureError(raven.WrapWithExtra(err, extraInfo), logErrorTags)
		} else {
			raven.CaptureError(err, logErrorTags)
		}
	}
}

func ReturnEncryptedEthKey(id uuid.UUID, createdAt time.Time, rawPrivateKey string) string {
	hashedSessionID := HashHex(hex.EncodeToString([]byte(fmt.Sprint(id))), sha3.New256())
	hashedCreationTime := HashHex(hex.EncodeToString([]byte(fmt.Sprint(createdAt.Clock()))), sha3.New256())

	encryptedKey := Encrypt(hashedSessionID, rawPrivateKey, hashedCreationTime)

	return hex.EncodeToString(encryptedKey)
}

func ReturnDecryptedEthKey(id uuid.UUID, createdAt time.Time, encryptedPrivateKey string) string {
	hashedSessionID := HashHex(hex.EncodeToString([]byte(fmt.Sprint(id))), sha3.New256())
	hashedCreationTime := HashHex(hex.EncodeToString([]byte(fmt.Sprint(createdAt.Clock()))), sha3.New256())

	decryptedKey := Decrypt(hashedSessionID, encryptedPrivateKey, hashedCreationTime)

	return hex.EncodeToString(decryptedKey)
}

/*LogIfValidationError logs any validation error from database. */
func LogIfValidationError(msg string, err *validate.Errors, extraInfo map[string]interface{}) {
	if err == nil || err.Count() == 0 {
		return
	}

	fmt.Printf("%v: %v\n", msg, err.Errors)
	if IsRavenEnabled() {
		info := make(map[string]interface{})
		for k, v := range err.Errors {
			info[k] = v
		}
		if extraInfo != nil {
			for k, v := range extraInfo {
				info[k] = v
			}
		}

		raven.CaptureError(raven.WrapWithExtra(errors.New(msg), info), logErrorTags)
	}
}

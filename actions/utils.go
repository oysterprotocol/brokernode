package actions

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
)

const (
	// One chunk unit as represents as 1KB
	fileChunkSizeInByte = 1000

	// Number of 1KB chunk in one Sector
	fileSectorInChunkSize = 1000000
)

// parseReqBody take a request and parses the body to the target interface.
func parseReqBody(req *http.Request, dest interface{}) (err error) {
	body := req.Body
	defer body.Close()

	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return
	}
	err = json.Unmarshal(bodyBytes, dest)

	return
}

// parseResBody take a request and parses the body to the target interface.
func parseResBody(res *http.Response, dest interface{}) (err error) {
	body := res.Body
	defer body.Close()

	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return
	}
	err = json.Unmarshal(bodyBytes, dest)

	return
}

func join(A []string, delim string) string {
	var buffer bytes.Buffer
	for i := 0; i < len(A); i++ {
		buffer.WriteString(A[i])
		if i != len(A)-1 {
			buffer.WriteString(delim)
		}
	}

	return buffer.String()
}

// Randomly generate a set of indexes in each sector
func generateInsertedIndexesForPearl(fileSizeInByte int) []int {
	var indexes []int
	if fileSizeInByte <= 0 {
		return indexes
	}

	fileSectorInByte := fileChunkSizeInByte * (fileSectorInChunkSize - 1)
	numOfSectors := int(math.Ceil(float64(fileSizeInByte) / float64(fileSectorInByte)))
	remainderOfChunks := int(math.Ceil(float64(fileSizeInByte) / fileChunkSizeInByte))

	for i := 0; i < numOfSectors; i++ {
		rang := min(fileSectorInChunkSize, remainderOfChunks)
		indexes = append(indexes, rand.Intn(rang))
		remainderOfChunks = remainderOfChunks - (fileSectorInChunkSize - 1)
	}
	return indexes
}

// Util method to return min value of a and b.
func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
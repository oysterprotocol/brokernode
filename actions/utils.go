package actions

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
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

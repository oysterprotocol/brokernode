package oyster_utils

import (
	"errors"
	"github.com/iotaledger/giota"
	"strings"
)

var (
	TrytesAlphabet = []rune("9ABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

func init() {
}

func AsciiToTrytes(asciiString string) (string, error) {
	var err error
	trytes := ""
	for _, character := range asciiString {
		var charCode = character

		// If not recognizable ASCII character, return null
		if charCode > 255 {
			return trytes, err
		}

		var firstValue = charCode % 27
		var secondValue = (charCode - firstValue) / 27
		var trytesValue = string(TrytesAlphabet[firstValue]) + string(TrytesAlphabet[secondValue])
		trytes += string(trytesValue)
	}

	return trytes, err
}

func TrytesToAsciiTrimmed(inputTrytes string) (string, error) {
	notNineIndex := strings.LastIndexFunc(inputTrytes, func(rune rune) bool {
		return string(rune) != "9"
	})
	trimmedString := inputTrytes[0 : notNineIndex+1]

	if len(trimmedString)%2 != 0 {
		trimmedString += "9"
	}

	return TrytesToAscii(trimmedString)
}

func TrytesToAscii(inputTrytes string) (string, error) {
	// If input length is odd, return an error
	if len(inputTrytes)%2 != 0 {
		return "", errors.New("TrytesToAscii needs input with an even number of characters!")
	}

	outputString := ""
	for i := 0; i < len(inputTrytes); i += 2 {
		// get a trytes pair
		trytes := string(inputTrytes[i]) + string(inputTrytes[i+1])

		firstValue := strings.Index(string(TrytesAlphabet), (string(trytes[0])))
		secondValue := strings.Index(string(TrytesAlphabet), (string(trytes[1])))

		decimalValue := firstValue + secondValue*27
		character := string(decimalValue)
		outputString += character
	}

	return outputString, nil
}

//TrytesToBytes and BytesToTrytes written by Chris Warner, thanks!
func TrytesToBytes(t giota.Trytes) []byte {
	var output []byte
	trytesString := string(t)
	for i := 0; i < len(trytesString); i += 2 {
		v1 := strings.IndexRune(string(TrytesAlphabet), rune(trytesString[i]))
		v2 := strings.IndexRune(string(TrytesAlphabet), rune(trytesString[i+1]))
		decimal := v1 + v2*27
		c := byte(decimal)
		output = append(output, c)
	}
	return output
}

func BytesToTrytes(b []byte) giota.Trytes {
	var output string
	for _, c := range b {
		v1 := c % 27
		v2 := (c - v1) / 27
		output += string(TrytesAlphabet[v1]) + string(TrytesAlphabet[v2])
	}
	return giota.Trytes(output)
}

func MakeAddress(hashString string) string {
	result := string(BytesToTrytes([]byte(hashString)))

	if len(result) > 81 {
		return result[0:81]
	} else if len(result) < 81 {
		return PadWith9s(result, 81)
	}
	return result

}

func PadWith9s(stringToPad string, desiredLength int) string {
	padCountInt := desiredLength - len(stringToPad)
	var retStr = stringToPad + strings.Repeat("9", padCountInt)
	return retStr[0:desiredLength]
}

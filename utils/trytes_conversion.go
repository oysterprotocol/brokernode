package oyster_utils

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"github.com/iotaledger/giota"
	"strings"
)

var (
	TrytesAlphabet = []rune("9ABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

const StopperTryte = "A"

func AsciiToTrytes(asciiString string) (string, error) {
	var b strings.Builder

	for _, character := range asciiString {
		var charCode = character

		// If not recognizable ASCII character, return null
		if charCode > 255 {
			err := errors.New("asciiString is not ASCII char in AsciiToTrytes method")
			if err != nil {
				LogIfError(err, nil)
			}
			return "", err
		}

		var firstValue = charCode % 27
		var secondValue = (charCode - firstValue) / 27
		var trytesValue = string(TrytesAlphabet[firstValue]) + string(TrytesAlphabet[secondValue])
		b.WriteString(string(trytesValue))
	}

	return b.String(), nil
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
		err := errors.New("TrytesToAscii needs input with an even number of characters!")
		if err != nil {
			LogIfError(err, nil)
		}
		return "", err
	}

	var b strings.Builder
	for i := 0; i < len(inputTrytes); i += 2 {
		// get a trytes pair
		trytes := string(inputTrytes[i]) + string(inputTrytes[i+1])

		firstValue := strings.Index(string(TrytesAlphabet), (string(trytes[0])))
		secondValue := strings.Index(string(TrytesAlphabet), (string(trytes[1])))

		decimalValue := firstValue + secondValue*27
		character := string(decimalValue)
		b.WriteString(character)
	}

	return b.String(), nil
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

func ChunkMessageToTrytesWithStopper(messageString string) (giota.Trytes, error) {
	// messageString will be either a binary string or will already be in trytes
	trytes, err := giota.ToTrytes(messageString)
	if err == nil {
		// not capturing here since this isn't a "real" error
		return trytes, nil
	}
	trytes, err = giota.ToTrytes(RunesToTrytes([]rune(messageString)) + StopperTryte)
	LogIfError(err, nil)
	return trytes, err
}

func RunesToTrytes(r []rune) string {

	var output string
	for _, c := range r {
		v1 := c % 27
		v2 := (c - v1) / 27
		output += string(TrytesAlphabet[v1]) + string(TrytesAlphabet[v2])
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
	trytes, err := giota.ToTrytes(output)
	if err != nil {
		LogIfError(err, nil)
	}
	return trytes
}

func MakeAddress(hashString string) string {
	bytes, err := hex.DecodeString(hashString)
	if err != nil {
		LogIfError(err, nil)
		return ""
	}

	result := string(BytesToTrytes(bytes))

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

/*Sha256ToAddress wraps functionality to turn a sha256 hash into an address*/
func Sha256ToAddress(hashString string) string {
	obfuscatedHash := HashHex(hashString, sha512.New384())
	return string(MakeAddress(obfuscatedHash))
}

/*ComputeSectorDataMapAddress computes a particular sectorIdx addresses in term of DataMaps. Limit by maxNumbOfHashes.*/
func ComputeSectorDataMapAddress(genHash string, sectorIdx int, maxNumOfHashes int) []string {
	var addr []string

	currHash := genHash
	for i := 0; i < sectorIdx*FileSectorInChunkSize; i++ {
		currHash = HashHex(currHash, sha256.New())
	}

	for i := 0; i < maxNumOfHashes; i++ {
		obfuscatedHash := HashHex(currHash, sha512.New384())
		currAddr := string(MakeAddress(obfuscatedHash))

		addr = append(addr, currAddr)
		currHash = HashHex(currHash, sha256.New())
	}
	return addr
}

package oyster_utils

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"hash"
)

func Encrypt(key string, secret string, nonce string) []byte {

	keyInBytes, err := hex.DecodeString(key)
	panicOnErr(err)
	secretInBytes, err := hex.DecodeString(secret)
	panicOnErr(err)
	block, err := aes.NewCipher(keyInBytes)
	panicOnErr(err)
	gcm, err := cipher.NewGCM(block)
	panicOnErr(err)
	nonceInBytes, err := hex.DecodeString(nonce[0 : 2*gcm.NonceSize()])
	panicOnErr(err)
	data := gcm.Seal(nil, nonceInBytes, secretInBytes, nil)
	return data
}

func Decrypt(key string, cipherText string, nonce string) []byte {
	keyInBytes, err := hex.DecodeString(key)
	panicOnErr(err)
	data, err := hex.DecodeString(cipherText)
	panicOnErr(err)
	block, err := aes.NewCipher(keyInBytes)
	panicOnErr(err)
	gcm, err := cipher.NewGCM(block)
	panicOnErr(err)
	nonceInBytes, err := hex.DecodeString(nonce[0 : 2*gcm.NonceSize()])
	panicOnErr(err)
	data, err = gcm.Open(nil, nonceInBytes, data, nil)
	if err != nil {
		return nil
	}
	return data
}

func HashString(str string, shaAlg hash.Hash) (h string) {
	shaAlg.Write([]byte(str))
	h = hex.EncodeToString(shaAlg.Sum(nil))
	return
}

func HashHex(hexStr string, shaAlg hash.Hash) (h string) {
	input, err := hex.DecodeString(hexStr)
	panicOnErr(err)
	shaAlg.Write(input)
	h = hex.EncodeToString(shaAlg.Sum(nil))
	return
}

func panicOnErr(err error) {
	// this is just so that the same 3 lines aren't repeated
	// throughout the encrypt/decrypt functions
	if err != nil {
		LogIfError(err, nil)
		panic(err.Error())
	}
}

package oyster_utils_test

import (
	"encoding/hex"
	"github.com/oysterprotocol/brokernode/utils"
	"testing"
)

type EncryptionTestStruct struct {
	key    string
	secret string
	nonce  string
	result string
}

var encryptionTestCases = []EncryptionTestStruct{
	{
		key:    "64dc1ce4655554f514a4ce83e08c1d08372fdf02bd8c9b6dbecfc74b783d39d1",
		secret: "0000000000000000000000000000000000000000000000000000000000000001",
		nonce:  "948791aa5dfd8f71405da35c637ad58cc9f5fec7424dba3e97630921e130c5b6",
		result: "592d93e3bb89f8835c9949460a2b0195e8ea915e724a35b3c713a6201ce94002ae94b5546647db1ffa94a3002f500897",
	},
	{
		key:    "99577b266e77d07e364d0b87bf1bcef44c78e3668dfdc3881969b375c09d4fcd",
		secret: "1004444400000006780000000000000000000000000012345000000765430001",
		nonce:  "23384a8eabc4a4ba091cfdbcb3dbacdc27000c03e318fd52accb8e2380f11320",
		result: "73fb51882b7fdd04d1f92146fed5b198e820ea08d00dd7bb65cde4f8a0b0e00cfedb93317ef05d7d149371b4b6b2c272",
	},
	{
		key:    "7fb4ca9cc0032bafc2ebd0fda018a41f5adfcf441123de22ab736a42207933f7",
		secret: "7777777774444444777777744444447777777444444777777744444777777744",
		nonce:  "0d412fa10c9027b7163302e38c96a5c0904b1b04cb55c66162296d0be2e3caa2",
		result: "8a859e9a265f28d36153c5d3849f5e1ec7574431fb1af68b0bc74d928772edcce1ae50fae6c4634bdcc876eef85679a9",
	},
}

func Test_Encrypt(t *testing.T) {
	for _, tc := range encryptionTestCases {
		result := oyster_utils.Encrypt(tc.key, tc.secret, tc.nonce)
		if hex.EncodeToString(result) != tc.result {
			t.Fatalf("Encrypt() result should be %s but returned %s",
				tc.result, hex.EncodeToString(result))
		}
	}
}

func Test_Decrypt(t *testing.T) {
	for _, tc := range encryptionTestCases {
		secret := oyster_utils.Decrypt(tc.key, tc.result, tc.nonce)
		if hex.EncodeToString(secret) != tc.secret {
			t.Fatalf("Decrypt() should be %s but returned %s",
				tc.secret, hex.EncodeToString(secret))
		}
	}
}

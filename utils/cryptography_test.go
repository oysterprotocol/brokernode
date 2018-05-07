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
		result: "691da3d38bb9c8b36ca979763a1b31a5d8daa16e427a0583f72396102cd9703378244ef3a08191754ceb61e2959561d52a8db41fa620a3c2c77d24ae8725386f58c962712d6bd659860f53eab91a8328",
	},
	{
		key:    "99577b266e77d07e364d0b87bf1bcef44c78e3668dfdc3881969b375c09d4fcd",
		secret: "1004444400000006780000000000000000000000000012345000000765430001",
		nonce:  "23384a8eabc4a4ba091cfdbcb3dbacdc27000c03e318fd52accb8e2380f11320",
		result: "52cf25f81f4be93699c91176cee581aedf18da38e03df5bf05fdd4cff5c3d03d261c8f974452d9b78fc5aa392042ec0f52eea73caee1c9d8bc51db16088ac5e10e141162d52b3a165635c7522e07a7a4",
	},
	{
		key:    "7fb4ca9cc0032bafc2ebd0fda018a41f5adfcf441123de22ab736a42207933f7",
		secret: "7777777774444444777777744444447777777444444777777744444777777744",
		nonce:  "0d412fa10c9027b7163302e38c96a5c0904b1b04cb55c66162296d0be2e3caa2",
		result: "cac5deda652c5ba021108693f4ef2e5d87170742886ab6c848b73de1c431adbfdb6936a0b37d4089a96572d4a74707127b16d605bbd0c17dd12e91fd660806e2e394e9762883228c66866592d8c9e052",
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
		if string(secret) != tc.secret {
			t.Fatalf("Decrypt() should be %s but returned %s",
				tc.secret, string(secret))
		}
	}
}

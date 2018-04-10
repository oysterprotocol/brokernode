package models_test

import (
	"crypto/sha512"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/utils"

	"github.com/iotaledger/giota"
)

type hashAddressConversion struct {
	sha256Hash     string
	ethPrivateSeed string
}

var encryptedTreasureCases = []hashAddressConversion{
	{sha256Hash: "64dc1ce4655554f514a4ce83e08c1d08372fdf02bd8c9b6dbecfc74b783d39d1",
		ethPrivateSeed: "0000000000000000000000000000000000000000000000000000000000000001"},
	{sha256Hash: "99577b266e77d07e364d0b87bf1bcef44c78e3668dfdc3881969b375c09d4fcd",
		ethPrivateSeed: "1004444400000006780000000000000000000000000012345000000765430001"},
	{sha256Hash: "7fb4ca9cc0032bafc2ebd0fda018a41f5adfcf441123de22ab736a42207933f7",
		ethPrivateSeed: "7777777774444444777777744444447777777444444777777744444777777744"},
}

func (ms *ModelSuite) Test_BuildDataMaps() {
	genHash := "genHashTest"
	fileBytesCount := 9000

	vErr, err := models.BuildDataMaps(genHash, fileBytesCount)
	ms.Nil(err)
	ms.Equal(0, len(vErr.Errors))

	expectedHashes := []string{
		"ef7edf0decd95c9e094184dca8641b68bb3ca0f69fec086341893816c68f7d9d408131fa01a66cf95f05b2a038185db9",
		"86ad8449bd1b32bcd86d86cfe7b3b6453f391c0c0df57956a2dff53f55709af3cd43a983ef46263cf8e361ae15734b33",
		"fbb914b1ba9cc663be0eb7b2570209af5caccfe5b7bba65e832c683072a969715e1b23866ce97ddb765fefe9b991e652",
		"e697116fd36a697f327f4682fd6f72250933bc61184fc36ff89badf749779aadf643b2e4f3fcd22fa9c07a6ce89c99a5",
		"167b2e33d17a4a96c6ad7216cd49c664b056efd30c08d65a354d1a5eb9cc9dbcb2f639495269f7ef5e56b8e62777edfc",
		"73ad9b9ba83acbf49a714980e660ead44f3fb574ee807d05d4ab728cfc9ecd1cd2f2a0a608948ea968d383db037a6d6c",
		"173b5a6ced53b7a84aa9f789bab0485418e949a3571ed964dde9b54618d38f212d496831cf083cc0b46d6d51e78461c7",
	}

	dMaps := []models.DataMap{}
	ms.DB.Where("genesis_hash = ?", genHash).Order("chunk_idx asc").All(&dMaps)

	for i, dMap := range dMaps {
		ms.Equal(expectedHashes[i], dMap.Hash)
	}
}

func (ms *ModelSuite) Test_CreateTreasurePayload() {
	maxSideChainLength := 50
	matchesFound := 0

	for _, tc := range encryptedTreasureCases {
		payload, err := models.CreateTreasurePayload(tc.ethPrivateSeed, tc.sha256Hash, maxSideChainLength)
		ms.Nil(err)

		payloadNotTryted := oyster_utils.TrytesToBytes(giota.Trytes(payload))

		currentHash := tc.sha256Hash

		for i := 0; i < maxSideChainLength; i++ {
			currentHash = oyster_utils.HashString(currentHash, sha512.New())
			result := oyster_utils.Decrypt(currentHash, string(payloadNotTryted))
			if result != "" {
				ms.Equal(result, tc.ethPrivateSeed)
				matchesFound++
			}
		}
	}
	ms.Equal(3, matchesFound)
}

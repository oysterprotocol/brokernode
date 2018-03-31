package oyster_utils_test

import (
	"github.com/iotaledger/giota"
	"github.com/oysterprotocol/brokernode/utils"
	"testing"
)

type tryteConversion struct {
	b []byte
	s string
	t giota.Trytes
}

type hashAddressConversion struct {
	hash    string
	address string
}

var stringConvCases = []tryteConversion{
	{b: []byte("Z"), s: "Z", t: giota.Trytes("IC")},
	{b: []byte("this is a test"), s: "this is a test", t: giota.Trytes("HDWCXCGDEAXCGDEAPCEAHDTCGDHD")},
	{b: []byte("Golang is the best lang!"), s: "Golang is the best lang!",
		t: giota.Trytes("QBCD9DPCBDVCEAXCGDEAHDWCTCEAQCTCGDHDEA9DPCBDVCFA")},
	{b: []byte(""), s: "", t: ""},
}

var hashAddressConvCases = []hashAddressConversion{
	{hash: "5804c3157e3de4e4a8b1f2417d8c61454e368883ec05e32f234386690e7c9696",
		address: "ZABBUAYARCXAVAZAABTCXASCTCYATCYAPCBBQCVAUCWAYAVAABSCBBRC9BVAYAZAYATCXA9BBBBBBBXAT"},
	{hash: "080779a63f5822c2606bfdd2801b5c4429918efcecffbaa34c2daadd51bc5748",
		address: "UABBUAABABCBPC9BXAUCZABBWAWARCWA9BUA9BQCUCSCSCWABBUAVAQCZARCYAYAWACBCBVABBTCUCRCT"},
	{hash: "d0199d3bd44c9301299de4d9d7054adb9c7fa11ac175cdee302794130b081681",
		address: "SCUAVACBCBSCXAQCSCYAYARCCBXAUAVAWACBCBSCTCYASCCBSCABUAZAYAPCSCQCCBRCABUCPCVAVAPCR"},
	{hash: "e512f80fa0e0c2872e0e29e621c40cf1693e112e020a708a619e7b87d421bf9c",
		address: "TCZAVAWAUCBBUAUCPCUATCUARCWABBABWATCUATCWACBTC9BWAVARCYAUARCUCVA9BCBXATCVAVAWATCU"},
	{hash: "cca31d69bcddfdd0ecd53d98c3daeca17ed61e04bf456ebd56b9ddbaf660091a",
		address: "RCRCPCXAVASC9BCBQCRCSCSCUCSCSCUATCRCSCZAXASCCBBBRCXASCPCTCRCPCVAABTCSC9BVATCUAYAQ"},
}

func Test_BytesToTrytes(t *testing.T) {
	for _, tc := range stringConvCases {
		result := oyster_utils.BytesToTrytes([]byte(tc.b))
		if result != tc.t {
			t.Fatalf("BytesToTrytes(%q) should be %#v but returned %s",
				tc.b, tc.t, result)
		}
	}
}

func Test_TrytesToBytes(t *testing.T) {
	for _, tc := range stringConvCases {
		if string(oyster_utils.TrytesToBytes(tc.t)) != string(tc.b) {
			t.Fatalf("TrytesToBytes(%q) should be %#v but returned %s",
				tc.t, tc.b, oyster_utils.TrytesToBytes(tc.t))
		}
	}
}

func Test_TrytesToAsciiTrimmed(t *testing.T) {
	for _, tc := range stringConvCases {
		result, _ := oyster_utils.TrytesToAsciiTrimmed(string(tc.t))
		if result != string(tc.s) {
			t.Fatalf("TrytesToAsciiTrimmed(%q) should be %#v but returned %s",
				tc.t, tc.s, result)
		}
	}
}

func Test_AsciiToTrytes(t *testing.T) {
	for _, tc := range stringConvCases {
		result, _ := oyster_utils.AsciiToTrytes(tc.s)
		if result != string(tc.t) {
			t.Fatalf("AsciiToTrytes(%q) should be %#v but returned %s",
				tc.s, tc.t, result)
		}
	}
}

func Test_MakeAddress(t *testing.T) {
	for _, tc := range hashAddressConvCases {
		result := oyster_utils.MakeAddress(tc.hash)
		if result != string(tc.address) {
			t.Fatalf("MakeAddress(%q) should be %#v but returned %s",
				tc.hash, tc.address, result)
		}
	}
}

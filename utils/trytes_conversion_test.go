package oyster_utils_test

import (
	"github.com/oysterprotocol/brokernode/utils"
	"github.com/iotaledger/giota"
	"testing"
)

type tryteByteConversion struct {
	s []byte
	t giota.Trytes
}

var stringConvCases = []tryteByteConversion{
	{s: []byte("Z"), t: giota.Trytes("IC")},
	{s: []byte("this is a test"), t: giota.Trytes("HDWCXCGDEAXCGDEAPCEAHDTCGDHD")},
	{s: []byte("Golang is the best lang!"),
		t: giota.Trytes("QBCD9DPCBDVCEAXCGDEAHDWCTCEAQCTCGDHDEA9DPCBDVCFA")},
	{s: []byte("Quizdeltagerne spiste jordbær med fløde, mens cirkusklovnen"),
		t: "9CIDXCNDSCTC9DHDPCVCTCFDBDTCEAGDDDXCGDHDTCEAYCCDFDSCQCFGDFFDEAADTCSCEAUC9DFGVFSCTCQAEAADTCBDGDEARCXCFDZCIDGDZC9DCDJDBDTCBD"},
	{s: []byte(""), t: ""},
}

func TestValidBytesToTrytes(t *testing.T) {
	for _, tc := range stringConvCases {
		result := oyster_utils.BytesToTrytes([]byte(tc.s))
		if result != tc.t {
			t.Fatalf("BytesToTrytes(%q) should be %#v but returned %s",
				tc.s, tc.t, result)
		}
	}
}

func TestValidTrytesToBytes(t *testing.T) {
	for _, tc := range stringConvCases {
		if string(oyster_utils.TrytesToBytes(tc.t)) != string(tc.s) {
			t.Fatalf("TrytesToBytes(%q) should be %#v but returned %s",
				tc.t, tc.s, oyster_utils.TrytesToBytes(tc.t))
		}
	}
}

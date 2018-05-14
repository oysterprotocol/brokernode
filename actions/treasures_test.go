package actions

import (
	"encoding/json"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/pkg/errors"
	"io/ioutil"
)

// Record data for VerifyTreasure method
type mockVerifyTreasure struct {
	input_addr   []string
	output_bool  bool
	output_error error
}

func (as *ActionSuite) Test_VerifyAndClaim_NoError() {
	res := as.JSON("/api/v2/treasures").Post(map[string]interface{}{
		"receiverEthAddr": "receiverEthAddr",
		"genesisHash":     "123",
	})

	as.Equal(200, res.Code)
}

func (as *ActionSuite) Test_VerifyTreasure_Success() {
	m := mockVerifyTreasure{
		output_bool:  true,
		output_error: nil,
	}
	IotaWrapper = services.IotaService{
		VerifyTreasure: m.verifyTreasure,
	}

	res := as.JSON("/api/v2/treasures").Post(map[string]interface{}{
		"receiverEthAddr": "receiverEthAddr",
		"genesisHash":     "123",
		"sectorIdx":       1,
		"numChunks":       5,
	})

	as.Equal(5, len(m.input_addr))
	as.Equal(200, res.Code)

	// Parse response
	resParsed := treasureRes{}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	as.Nil(err)
	err = json.Unmarshal(bodyBytes, &resParsed)
	as.Nil(err)

	as.Equal(true, resParsed.Success)
}

func (as *ActionSuite) Test_VerifyTreasure_FailureWithError() {
	m := mockVerifyTreasure{
		output_bool:  false,
		output_error: errors.New("Invalid address"),
	}
	IotaWrapper = services.IotaService{
		VerifyTreasure: m.verifyTreasure,
	}

	res := as.JSON("/api/v2/treasures").Post(map[string]interface{}{
		"receiverEthAddr": "receiverEthAddr",
		"genesisHash":     "123",
		"sectorIdx":       1,
		"numChunks":       5,
	})

	as.Equal(5, len(m.input_addr))

	// Parse response
	resParsed := treasureRes{}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	as.Nil(err)
	err = json.Unmarshal(bodyBytes, &resParsed)
	as.Nil(err)

	as.Equal(false, resParsed.Success)
}

// For mocking VerifyTreasure method
func (v *mockVerifyTreasure) verifyTreasure(addr []string) (bool, error) {
	v.input_addr = addr
	return v.output_bool, v.output_error
}

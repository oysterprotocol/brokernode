package actions

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/pkg/errors"
	"io/ioutil"
)

// Record data for VerifyTreasure method
type mockVerifyTreasure struct {
	hasCalled    bool
	input_addr   []string
	output_bool  bool
	output_error error
}

// Record data for ClaimPRL method
type mockClaimPrl struct {
	hasCalled           bool
	input_receiver_addr common.Address
	input_treasure_addr common.Address
	input_treasure_key  string
	output_bool         bool
}

func (as *ActionSuite) Test_VerifyTreasureAndClaim_Success() {
	mockVerifyTreasure := mockVerifyTreasure{
		output_bool:  true,
		output_error: nil,
	}
	IotaWrapper = services.IotaService{
		VerifyTreasure: mockVerifyTreasure.verifyTreasure,
	}
	mockClaimPrl := mockClaimPrl{
		output_bool:         true,
		input_receiver_addr: common.HexToAddress("0x0D8e461687b7D06f86EC348E0c270b0F279855F0"),
	}
	EthWrapper = services.Eth{
		ClaimPRL:                      mockClaimPrl.claimPRL,
		GenerateEthAddrFromPrivateKey: EthWrapper.GenerateEthAddrFromPrivateKey,
	}

	ethKey := "9999999999999999999999999999999999999999999999999999999999999999"

	res := as.JSON("/api/v2/treasures").Post(map[string]interface{}{
		"receiverEthAddr": "receiverEthAddr",
		"genesisHash":     "123",
		"sectorIdx":       1,
		"numChunks":       5,
		"ethKey":          ethKey,
	})

	as.Equal(200, res.Code)

	// Check mockVerifyTreasure
	as.True(mockVerifyTreasure.hasCalled)
	as.Equal(5, len(mockVerifyTreasure.input_addr))
	// Check mockClaimPrl
	as.True(mockClaimPrl.hasCalled)
	as.Equal(services.StringToAddress("receiverEthAddr"), mockClaimPrl.input_receiver_addr)
	address, err := EthWrapper.GenerateEthAddrFromPrivateKey(ethKey)
	as.Nil(err)
	as.Equal(address, mockClaimPrl.input_treasure_addr)
	as.Equal(ethKey, mockClaimPrl.input_treasure_key)

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

	as.True(m.hasCalled)
	as.Equal(5, len(m.input_addr))

	// Parse response
	resParsed := treasureRes{}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	as.Nil(err)
	err = json.Unmarshal(bodyBytes, &resParsed)
	as.Nil(err)

	as.Equal(false, resParsed.Success)
}

func (as *ActionSuite) Test_Claim_Failure() {
	mockVerifyTreasure := mockVerifyTreasure{
		output_bool:  true,
		output_error: nil,
	}
	IotaWrapper = services.IotaService{
		VerifyTreasure: mockVerifyTreasure.verifyTreasure,
	}
	mockClaimPrl := mockClaimPrl{
		output_bool: false,
	}
	EthWrapper = services.Eth{
		ClaimPRL:                      mockClaimPrl.claimPRL,
		GenerateEthAddrFromPrivateKey: EthWrapper.GenerateEthAddrFromPrivateKey,
	}

	res := as.JSON("/api/v2/treasures").Post(map[string]interface{}{
		"receiverEthAddr": "receiverEthAddr",
		"genesisHash":     "123",
		"sectorIdx":       1,
		"numChunks":       5,
		"ethKey":          "9999999999999999999999999999999999999999999999999999999999999999",
	})

	as.Equal(200, res.Code)

	as.True(mockVerifyTreasure.hasCalled)
	as.True(mockClaimPrl.hasCalled)

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
	v.hasCalled = true
	v.input_addr = addr
	return v.output_bool, v.output_error
}

// For mocking ClaimPRL method
func (v *mockClaimPrl) claimPRL(receiverAddress common.Address, treasureAddress common.Address, treasureKey string) bool {
	v.hasCalled = true
	v.input_receiver_addr = receiverAddress
	v.input_treasure_addr = treasureAddress
	v.input_treasure_key = treasureKey
	return v.output_bool
}

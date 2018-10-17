package actions_v2

import (
	"encoding/json"
	"errors"
	"github.com/oysterprotocol/brokernode/utils/eth_gateway"
	"io/ioutil"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/oysterprotocol/brokernode/services"
)

// Record data for VerifyTreasure method
type mockVerifyTreasure struct {
	hasCalled    bool
	input_addr   []string
	output_bool  bool
	output_error error
}

var checkClaimClockCalled = false
var ethAddressCalledWithCheckClaimClock common.Address

func (suite *ActionSuite) Test_VerifyTreasureAndClaim_Success() {
	checkClaimClockCalled = false

	mockVerifyTreasure := mockVerifyTreasure{
		output_bool:  true,
		output_error: nil,
	}
	IotaWrapper = services.IotaService{
		VerifyTreasure: mockVerifyTreasure.verifyTreasure,
	}
	EthWrapper = eth_gateway.Eth{
		GenerateEthAddrFromPrivateKey: EthWrapper.GenerateEthAddrFromPrivateKey,
		CheckClaimClock: func(address common.Address) (*big.Int, error) {
			checkClaimClockCalled = true
			ethAddressCalledWithCheckClaimClock = address
			return big.NewInt(1), nil
		},
	}

	ethKey := "9999999999999999999999999999999999999999999999999999999999999991"
	addr := eth_gateway.EthWrapper.GenerateEthAddrFromPrivateKey(ethKey)

	res := suite.JSON("/api/v2/treasures").Post(map[string]interface{}{
		"receiverEthAddr": addr,
		"genesisHash":     "1234",
		"sectorIdx":       1,
		"numChunks":       5,
		"ethKey":          ethKey,
	})

	suite.Equal(200, res.Code)

	// Check mockVerifyTreasure
	suite.True(mockVerifyTreasure.hasCalled)
	suite.Equal(5, len(mockVerifyTreasure.input_addr))

	suite.Equal(addr, ethAddressCalledWithCheckClaimClock)
	suite.Equal(true, checkClaimClockCalled)

	// Parse response
	resParsed := treasureRes{}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	suite.Nil(err)
	err = json.Unmarshal(bodyBytes, &resParsed)
	suite.Nil(err)

	suite.Equal(true, resParsed.Success)
}

func (suite *ActionSuite) Test_VerifyTreasure_FailureWithError() {

	checkClaimClockCalled = false

	m := mockVerifyTreasure{
		output_bool:  false,
		output_error: errors.New("Invalid address"),
	}
	IotaWrapper = services.IotaService{
		VerifyTreasure: m.verifyTreasure,
	}

	ethKey := "9999999999999999999999999999999999999999999999999999999999999991"
	addr := eth_gateway.EthWrapper.GenerateEthAddrFromPrivateKey(ethKey)

	res := suite.JSON("/api/v2/treasures").Post(map[string]interface{}{
		"receiverEthAddr": addr,
		"genesisHash":     "1234",
		"sectorIdx":       1,
		"numChunks":       5,
	})

	suite.True(m.hasCalled)
	suite.Equal(5, len(m.input_addr))

	// Parse response
	resParsed := treasureRes{}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	suite.Nil(err)
	err = json.Unmarshal(bodyBytes, &resParsed)
	suite.Nil(err)

	suite.Equal(false, resParsed.Success)
}

func (suite *ActionSuite) Test_Check_Claim_Clock_Error() {

	checkClaimClockCalled = false

	mockVerifyTreasure := mockVerifyTreasure{
		output_bool:  true,
		output_error: nil,
	}
	IotaWrapper = services.IotaService{
		VerifyTreasure: mockVerifyTreasure.verifyTreasure,
	}
	EthWrapper = eth_gateway.Eth{
		GenerateEthAddrFromPrivateKey: EthWrapper.GenerateEthAddrFromPrivateKey,
		CheckClaimClock: func(address common.Address) (*big.Int, error) {
			ethAddressCalledWithCheckClaimClock = address
			checkClaimClockCalled = true
			return big.NewInt(-1), errors.New("error")
		},
	}

	ethKey := "9999999999999999999999999999999999999999999999999999999999999991"
	addr := eth_gateway.EthWrapper.GenerateEthAddrFromPrivateKey(ethKey)

	res := suite.JSON("/api/v2/treasures").Post(map[string]interface{}{
		"receiverEthAddr": addr,
		"genesisHash":     "1234",
		"sectorIdx":       1,
		"numChunks":       5,
		"ethKey":          ethKey,
	})

	suite.Equal(200, res.Code)

	suite.True(mockVerifyTreasure.hasCalled)
	suite.True(checkClaimClockCalled)

	// Parse response
	resParsed := treasureRes{}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	suite.Nil(err)
	err = json.Unmarshal(bodyBytes, &resParsed)
	suite.Nil(err)

	suite.Equal(false, resParsed.Success)
}

// For mocking VerifyTreasure method
func (v *mockVerifyTreasure) verifyTreasure(addr []string) (bool, error) {
	v.hasCalled = true
	v.input_addr = addr
	return v.output_bool, v.output_error
}

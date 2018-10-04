package actions

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/oysterprotocol/brokernode/services"
)

type mockWaitForTransfer struct {
	hasCalled        bool
	input_brokerAddr common.Address
	output_int       *big.Int
	output_error     error
}

type mockSendPrl struct {
	hasCalled   bool
	input_msg   services.OysterCallMsg
	output_bool bool
}

type mockCheckPRLBalance struct {
	hasCalled  bool
	input_addr common.Address
	output_int *big.Int
}

func (v *mockWaitForTransfer) waitForTransfer(brokerAddr common.Address, transferType string) (*big.Int, error) {
	v.hasCalled = true
	v.input_brokerAddr = brokerAddr
	return v.output_int, v.output_error
}

func (v *mockSendPrl) sendPrl(msg services.OysterCallMsg) bool {
	v.hasCalled = true
	v.input_msg = msg
	return v.output_bool
}

func (v *mockCheckPRLBalance) checkPRLBalance(addr common.Address) *big.Int {
	v.hasCalled = true
	v.input_addr = addr
	return v.output_int
}

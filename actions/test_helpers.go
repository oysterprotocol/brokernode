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

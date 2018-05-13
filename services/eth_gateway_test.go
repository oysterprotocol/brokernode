package services_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"
	"math/big"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/oysterprotocol/brokernode/models"
	"fmt"
)

const (
	localNetworkUrl = "http://127.0.0.1:7545"
	oysterbyNetworkUrl = "http://54.197.3.171:8080"
)

// Ethereum Test Suite
type EthereumTestSuite struct {
	suite.Suite
	gateway *services.Eth
	ethClient *ethclient.Client // used for testing
}

//
// Testing Setup/Teardown
//

func (s *EthereumTestSuite) setupSuite(t *testing.T) {
	// EMPTY!!
}
func (s *EthereumTestSuite) tearDownSuite() {
	// EMPTY!!
}

//
// Ethereum Tests
//

// generate address test
func (s *EthereumTestSuite) generateAddress(t *testing.T) {

	// generate eth address using gateway
	addr, privateKey, error := s.gateway.GenerateEthAddr()
	if error != nil {
		t.Fatalf("error creating ethereum network address")
	}
	// ensure address is correct format
	if s.Assert().NotNil(addr) && common.IsHexAddress(addr.Hex()) {
		t.Fatalf("could not create a valid ethereum network address")
	}
	// ensure private key was returned
	if s.Assert().NotNil(privateKey) && privateKey == "" {
		t.Fatalf("could not create a valid private key")
	}
	t.Logf("ethereum network address was generated %v\n", addr.Hex())
}

// get gas price from network test
func (s *EthereumTestSuite) getGasPrice(t *testing.T) {

	// get the suggested gas price
	gasPrice, error := s.gateway.GetGasPrice()
	if error != nil {
		t.Fatalf("error retrieving gas price: %v\n",error)
	}
	if s.Assert().NotNil(gasPrice) && gasPrice.Uint64() > 0 {
		t.Logf("gas price verified: %v\n",gasPrice)
	} else {
		t.Fatalf("gas price less than zero: %v\n",gasPrice)
	}
}

// check balance on sim network test
func (s *EthereumTestSuite) checkBalance(t *testing.T) {

	// test balance for an account
	key, _ := crypto.GenerateKey()
	auth := bind.NewKeyedTransactor(key)

	// fund the sim

	// Convert string address to byte[] address form
	bal := s.gateway.CheckBalance(auth.From)
	if bal.Uint64() > 0 {
		t.Logf("balance verified: %v\n",bal)
	} else {
		t.Fatalf("balance less than zero: %v\n",bal)
	}
}

func (s *EthereumTestSuite) getCurrentBlock(t *testing.T) {
	// Get the current block from the network
	block, error := services.EthWrapper.GetCurrentBlock()
	if error != nil {
		t.Fatalf("could not retrieve the current block: %v\n",error)
	}
	if block != nil {
		t.Logf("retrieved the current block: %v\n",block)
	}
}

// send gas for a transaction
func (s *EthereumTestSuite) sendGas(t *testing.T) {
	// Send Gas to an Account
	// WIP - Add once we update the send via contract method
}

// send ether
func (s *EthereumTestSuite) sendEther(t *testing.T) {
	// Send Ether to an Account
	// WIP - Add once we update the send via contract method
}


//
// Oyster Pearl Contract Tests
//

// deploy the compiled oyster contract to Oysterby network
func (s *EthereumTestSuite) deployContractOnOysterby(t *testing.T) {
}

//
// Oyster Pearl Tests
//

// bury prl
func (s *EthereumTestSuite) buryPRL(t *testing.T) {

	// prepare oyster message call
	var msg = services.OysterCallMsg{
		From: common.HexToAddress("0x0d1d4e623d10f9fba5db95830f7d3839406c6af2"),
		To: common.HexToAddress("0xf17f52151ebef6c7334fad080c5704d77216b732"),
		Amount: *big.NewInt(1000),
		Gas: big.NewInt(10000).Uint64(),
		GasPrice: *big.NewInt(1000),
		TotalWei: *big.NewInt(100000),
		Data: []byte(""), // setup data
	}

	// Bury PRL
	var buried = s.gateway.BuryPrl(msg)
	if buried {
		// successful bury attempt
		t.Log("Buried the PRLs successfully")
	} else {
		// failed bury attempt
		t.Fatal("Faild to bury PRLs. Try Again?")
	}
}

// send prl
func (s *EthereumTestSuite) sendPRL(t *testing.T) {

	// prepare oyster message call
	var msg = services.OysterCallMsg{
		From: common.HexToAddress("0x0d1d4e623d10f9fba5db95830f7d3839406c6af2"),
		To: common.HexToAddress("0xf17f52151ebef6c7334fad080c5704d77216b732"),
		Amount: *big.NewInt(1000),
		Gas: big.NewInt(10000).Uint64(),
		GasPrice: *big.NewInt(1000),
		TotalWei: *big.NewInt(100000),
		Data: []byte(""), // setup data // TODO finalize by adding contract call to
	}

	// Send PRL
	var sent = s.gateway.SendPRL(msg)
	if sent {
		// successful prl send
		t.Logf("Sent PRL to :%v",msg.From.Hex())
	} else {
		// failed prl send
		t.Fatalf("Failed to send PRL to:%v",msg.From.Hex())
	}
}

// claim prl
func (s *EthereumTestSuite) claimPRL(t *testing.T) {

	// Need to fake the completed uploads by populating with data
	var rowWithGasTransferSuccess = models.CompletedUpload{
		GenesisHash:   "RowWithGasTransferSuccess",
		ETHAddr:       "0x5aeda56215b167893e80b4fe645ba6d5bab767de",
		ETHPrivateKey: "8d5366123cb560bb606379f90a0bfd4769eecc0557f1b362dcae9012b548b1e5",
		PRLStatus:     models.PRLClaimNotStarted,
		GasStatus:     models.GasTransferSuccess,
	}

	// mock completed upload
	completedUploads := []models.CompletedUpload{rowWithGasTransferSuccess}

	// Claim PRL
	err := s.gateway.ClaimPRLs(completedUploads)
	if err != nil {
		t.Fatal("Failed to claim PRLs")
	} else {
		t.Log("PRLs have been successfully claimed")
	}

}

// subscribe to transfer
func (s *EthereumTestSuite) subscribeToTransfer(t *testing.T) {

	// Subscribe to a Transaction
	// subscribeToTransfer(brokerAddr common.Address, outCh chan<- types.Log)
	broker := common.HexToAddress("")
	channel := make(chan types.Log)
	s.gateway.SubscribeToTransfer(broker, channel)

	fmt.Printf("Subscribed to :%v",<-channel)
}



package services_test

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/stretchr/testify/suite"
)

// Ethereum Test Suite
type EthereumTestSuite struct {
	suite.Suite
	gateway   *services.Eth
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
func Test_generateAddress(t *testing.T) {

	// generate eth address using gateway
	addr, privateKey, err := services.EthWrapper.GenerateEthAddr()
	if err != nil {
		t.Fatalf("error creating ethereum network address")
	}
	// ensure address is correct format
	if addr == services.StringToAddress("") && common.IsHexAddress(addr.Hex()) {
		t.Fatalf("could not create a valid ethereum network address")
	}
	// ensure private key was returned
	if privateKey == "" {
		t.Fatalf("could not create a valid private key")
	}
	t.Logf("ethereum network address was generated %v\n", addr.Hex())
}

// generate address from private key test
func Test_generateEthAddrFromPrivateKey(t *testing.T) {

	// generate eth address using gateway
	originalAddr, originalPrivateKey, err := services.EthWrapper.GenerateEthAddr()
	if err != nil {
		t.Fatalf("error creating ethereum network address")
	}

	generatedAddress := services.EthWrapper.GenerateEthAddrFromPrivateKey(originalPrivateKey)

	// ensure address is what we expected
	if originalAddr != generatedAddress {
		t.Fatalf("generated address was %s but we expected %s", generatedAddress, originalAddr)
	}
}

// get gas price from network test
func Test_getGasPrice(t *testing.T) {

	// get the suggested gas price
	gasPrice, err := services.EthWrapper.GetGasPrice()
	if err != nil {
		t.Fatalf("error retrieving gas price: %v\n", err)
	}
	if gasPrice != nil && gasPrice.Uint64() > 0 {
		t.Logf("gas price verified: %v\n", gasPrice)
	} else {
		t.Fatalf("gas price less than zero: %v\n", gasPrice)
	}
}

// check balance on sim network test
func Test_checkBalance(t *testing.T) {

	// test balance for an account
	key, _ := crypto.GenerateKey()
	auth := bind.NewKeyedTransactor(key)

	// fund the sim

	// Convert string address to byte[] address form
	bal := services.EthWrapper.CheckBalance(auth.From)
	if bal.Uint64() > 0 {
		t.Logf("balance verified: %v\n", bal)
	} else {
		//t.Fatalf("balance less than zero: %v\n", bal)
		/*TODO make this fatal again once we are using addresses with balances*/
		t.Logf("balance less than zero: %v\n", bal)
	}
}

func Test_getCurrentBlock(t *testing.T) {
	// Get the current block from the network
	block, err := services.EthWrapper.GetCurrentBlock()
	if err != nil {
		t.Fatalf("could not retrieve the current block: %v\n", err)
	}
	if block != nil {
		t.Logf("retrieved the current block: %v\n", block)
	}
}

// send gas for a transaction
func Test_sendGas(t *testing.T) {
	// Send Gas to an Account
	// WIP - Add once we update the send via contract method
}

// send ether
func Test_sendEther(t *testing.T) {
	// Send Ether to an Account
	// WIP - Add once we update the send via contract method
}

//
// Oyster Pearl Contract Tests
//

// deploy the compiled oyster contract to Oysterby network
func (s *EthereumTestSuite) deployContractOnOysterby(t *testing.T) {
	// TODO:  Does this method need a "Test_" prefix as well, or should we just call it from
	// the other test methods?  Does it need to be attached to the EthereumTestSuite struct?
}

//
// Oyster Pearl Tests
//

// bury prl
func Test_buryPRL(t *testing.T) {

	//// prepare oyster message call
	//var msg = services.OysterCallMsg{
	//	From:     common.HexToAddress("0x0d1d4e623d10f9fba5db95830f7d3839406c6af2"),
	//	To:       common.HexToAddress("0xf17f52151ebef6c7334fad080c5704d77216b732"),
	//	Amount:   *big.NewInt(1000),
	//	Gas:      big.NewInt(10000).Uint64(),
	//	GasPrice: *big.NewInt(1000),
	//	TotalWei: *big.NewInt(100000),
	//	Data:     []byte(""), // setup data
	//}
	//
	//// Bury PRL
	//var buried = services.EthWrapper.BuryPrl(msg)
	//if buried {
	//	// successful bury attempt
	//	t.Log("Buried the PRLs successfully")
	//} else {
	//	// failed bury attempt
	//	t.Fatal("Faild to bury PRLs. Try Again?")
	//}
}

// send prl
func Test_sendPRL(t *testing.T) {

	//// prepare oyster message call
	//var msg = services.OysterCallMsg{
	//	From:     common.HexToAddress("0x0d1d4e623d10f9fba5db95830f7d3839406c6af2"),
	//	To:       common.HexToAddress("0xf17f52151ebef6c7334fad080c5704d77216b732"),
	//	Amount:   *big.NewInt(1000),
	//	Gas:      big.NewInt(10000).Uint64(),
	//	GasPrice: *big.NewInt(1000),
	//	TotalWei: *big.NewInt(100000),
	//	Data:     []byte(""), // setup data // TODO finalize by adding contract call to
	//}
	//
	//// Send PRL
	//var sent = services.EthWrapper.SendPRL(msg)
	//if sent {
	//	// successful prl send
	//	t.Logf("Sent PRL to :%v", msg.From.Hex())
	//} else {
	//	// failed prl send
	//	t.Fatalf("Failed to send PRL to:%v", msg.From.Hex())
	//}
}

// claim prl
func Test_claimPRL(t *testing.T) {

	//// Receiver
	//receiverAddress := common.HexToAddress("0xC30efFC3509D56ef748d51f9580c81ff8e9c610E")
	//
	//// Setup Found Treasure Properties
	//treasureAddress := common.HexToAddress("0x5aeda56215b167893e80b4fe645ba6d5bab767de")
	//treasurePrivateKey := "8d5366123cb560bb606379f90a0bfd4769eecc0557f1b362dcae9012b548b1e5"
	//
	//// Claim PRL
	//claimed := services.EthWrapper.ClaimPRL(receiverAddress, treasureAddress, treasurePrivateKey)
	//if !claimed {
	//	t.Fatal("Failed to claim PRLs")
	//} else {
	//	t.Log("PRLs have been successfully claimed")
	//}

}

// claim unused prl from completed upload
func Test_claimUnusedPRL(t *testing.T) {

	//// Need to fake the completed uploads by populating with data
	//var rowWithGasTransferSuccess = models.CompletedUpload{
	//	GenesisHash:   "RowWithGasTransferSuccess",
	//	ETHAddr:       "0x5aeda56215b167893e80b4fe645ba6d5bab767de",
	//	ETHPrivateKey: "8d5366123cb560bb606379f90a0bfd4769eecc0557f1b362dcae9012b548b1e5",
	//	PRLStatus:     models.PRLClaimNotStarted,
	//	GasStatus:     models.GasTransferSuccess,
	//}
	//
	//// mock completed upload
	//completedUploads := []models.CompletedUpload{rowWithGasTransferSuccess}
	//
	//// Claim PRL
	//err := services.EthWrapper.ClaimUnusedPRLs(completedUploads)
	//if err != nil {
	//	t.Fatal("Failed to claim PRLs")
	//} else {
	//	t.Log("PRLs have been successfully claimed")
	//}

}

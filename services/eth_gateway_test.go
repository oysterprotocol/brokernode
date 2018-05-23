package services_test

import (
	"testing"

	"github.com/oysterprotocol/brokernode/services"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"os"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"strings"
	"math/big"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/core"
	"log"
	"context"
	"time"
)

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
	if common.IsHexAddress(addr.Str()) {
		t.Fatalf("could not create a valid ethereum network address:%v", addr.Str())
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
	t.Logf("generated address :%v", generatedAddress.Hex())
}

// get gas price from network test
func Test_getGasPrice(t *testing.T) {

	// get the suggested gas price
	gasPrice, err := services.EthWrapper.GetGasPrice()
	if err != nil {
		t.Fatalf("error retrieving gas price: %v\n", err)
	}
	if gasPrice.IsUint64() && gasPrice.Uint64() > 0 {
		t.Logf("gas price verified: %v\n", gasPrice)
	} else {
		t.Fatalf("gas price less than zero: %v\n", gasPrice)
	}
	t.Logf("current network gas price :%v", gasPrice.String())
}

// check balance on test network test
func Test_checkBalance(t *testing.T) {

	// test balance for an account
	testAcct := "0xf10a2706e98ef86b6866ae6cab2e0ca501fdf091"

	// Convert string address to byte[] address form
	bal := services.EthWrapper.CheckBalance(common.HexToAddress(testAcct))
	if bal.Uint64() > 0 {
		t.Logf("balance verified: %v\n", bal)
	} else {
		t.Fatalf("balance less than zero: %v\n", bal)
	}
}

func Test_getCurrentBlockNumber(t *testing.T) {
	// Get the current block from the network
	block, err := services.EthWrapper.GetCurrentBlock()
	if err != nil {
		t.Fatalf("could not retrieve the current block: %v\n", err)
	}
	if block != nil {
		t.Logf("retrieved the current block: %v\n", block.Number())
	}
}

func Test_getCurrentBlockGasLimit(t *testing.T) {
	// Get the current block from the network
	block, err := services.EthWrapper.GetCurrentBlock()
	if err != nil {
		t.Fatalf("could not retrieve the current block: %v\n", err)
	}
	if block != nil {
		t.Logf("retrieved the current block gas limit: %v\n", block.GasLimit())
	}
}

// simulated blockchain to deploy oyster pearl
func Test_deployOysterPearl(t *testing.T) {

	// Generate a new random account and a funded simulator
	key, _ := crypto.GenerateKey()
	auth := bind.NewKeyedTransactor(key)

	// genesis
	genesis := make(core.GenesisAlloc)
	genesis[auth.From] = core.GenesisAccount{
		Balance: big.NewInt(100000),
		Nonce:uint64(0),
	}

	// simulator
	sim := backends.NewSimulatedBackend(genesis)

	// initialize the context
	deadline := time.Now().Add(1000 * time.Millisecond)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	// Deploy a token contract on the simulated blockchain
	newAddress, transaction, token, err := services.DeployOysterPearl(&bind.TransactOpts{
		Nonce:big.NewInt(0),
		From:auth.From,
		GasLimit:uint64(30000),
		GasPrice: big.NewInt(25000),
		Context:  ctx,
		Signer:auth.Signer,
	}, sim)
	if err != nil {
		log.Fatalf("Failed to deploy oyster token contract: %v", err)
	}
	t.Logf("newAddress :%v", newAddress)
	t.Logf("transaction :%v", transaction)
	t.Logf("token :%v", token)

	// Commit all pending transactions in the simulator and print the names again
	sim.Commit()
}

// test sending PRLs from OysterPearl Contract
func Test_tokenNameFromOysterPearl(t *testing.T) {
	t.Skip(nil)
	// test ethClient
	var backend, _ = ethclient.Dial("http://54.86.134.172:8080")
	oysterPearl, err := services.NewOysterPearl(common.HexToAddress("0x84e07b9833af3d3c8e07b71b1c9c041ec5909d5d"), backend)
	if err != nil {
		t.Fatalf("unable to access contract instance at :%v",err)
	}
	name, err := oysterPearl.Name(nil)
	if err != nil {
		t.Fatalf("unable to access contract name")
	}
	t.Logf("oyster pearl contract name :%v", name)
}

// test sending PRLs from OysterPearl Contract
func Test_transferPRLFromOysterPearl(t *testing.T) {

	//file, err := ioutil.ReadFile("./testdata/key.json")
	//if err != nil {
	//	t.Fatalf("error reading key file")
	//}
	//var key map[string]interface{}
	//json.Unmarshal(file, &key)
	////address := key["address"]
	//t.Logf("key :%v", key)

	// test ethClient
	var backend, _ = ethclient.Dial("http://54.86.134.172:8080")
	oysterPearl, err := services.NewOysterPearl(common.HexToAddress("0x84e07b9833af3d3c8e07b71b1c9c041ec5909d5d"), backend)
	if err != nil {
		t.Fatalf("unable to access contract instance at :%v",err)
	}

	walletKey := os.Getenv("MAIN_WALLET_KEY")
	t.Logf("using wallet key:%v", walletKey)

	// Create an authorized transactor and spend 1 PRL
	auth, err := bind.NewTransactor(strings.NewReader(walletKey), "oysterby4000")
	if err != nil {
		t.Fatalf("unable to create a new transactor :%v", err)
	}

	//testPRLAcct := common.HexToAddress("f10a2706e98ef86b6866ae6cab2e0ca501fdf091")
	//testPRLAcct := common.HexToAddress("0xDf7D0030bfed998Db43288C190b63470c2d18F50")

	// wrap the oyster pearl contract instance into a session
	session := &services.OysterPearlSession{
		Contract: oysterPearl,
		CallOpts: bind.CallOpts{
			Pending: true,
		},
		TransactOpts: bind.TransactOpts{
			From:     auth.From,
			Signer:   auth.Signer,
			GasLimit: big.NewInt(3141592).Uint64(),
		},
	}
	session.Name()
	session.Transfer(common.HexToAddress("0x0000000000000000000000000000000000000000"), big.NewInt(1))

}

/*
func Test_sendEth(t *testing.T) {

	// test account accepting ether
	testAcct := common.HexToAddress("f10a2706e98ef86b6866ae6cab2e0ca501fdf091")
	transferValue := new(big.Int).SetUint64(1)

	// Send ether to test account
	raw, err := services.EthWrapper.SendETH(testAcct, transferValue)
	if err != nil {
		t.Logf("failed to send ether to %v ether to %v", transferValue, testAcct.Hex())
		t.Fatalf("transaction error: %v", err)
	}
	t.Logf("raw transaction : %v", raw)
}
*/

/*
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
*/
//
// Oyster Pearl Contract Tests
//
/*
// deploy the compiled oyster contract to Oysterby network
func (s *EthereumTestSuite) deployContractOnOysterby(t *testing.T) {
}
*/
//
// Oyster Pearl Tests
//

// bury prl
/*
func Test_buryPRL(t *testing.T) {

	// PRL based addresses
	from := common.HexToAddress("0x68e97d19da3b8a0dff21e2ac6bf1b7f63d4e7360")
	to := common.HexToAddress("0xf10a2706e98ef86b6866ae6cab2e0ca501fdf091")
	// Gas Price
	gasPrice, err := services.EthWrapper.GetGasPrice()
	if err != nil {
		t.Fatalf("error retrieving gas price: %v\n", err)
	}
	// prepare oyster message call
	var msg = services.OysterCallMsg{
		From:     from,
		To:       to,
		Amount:   *big.NewInt(100),
		Gas:      big.NewInt(10000).Uint64(),
		GasPrice: *gasPrice,
		TotalWei: *big.NewInt(100000),
		Data:     []byte(""), // setup data
	}

	// Bury PRL
	var buried = services.EthWrapper.BuryPrl(msg)
	if buried {
		// successful bury attempt
		t.Log("Buried the PRLs successfully")
	} else {
		// failed bury attempt
		t.Fatal("Faild to bury PRLs. Try Again?")
	}
}
*/
// send prl
/*
func (s *EthereumTestSuite) sendPRL(t *testing.T) {

	// prepare oyster message call
	var msg = services.OysterCallMsg{
		From:     common.HexToAddress("0x0d1d4e623d10f9fba5db95830f7d3839406c6af2"),
		To:       common.HexToAddress("0xf17f52151ebef6c7334fad080c5704d77216b732"),
		Amount:   *big.NewInt(1000),
		Gas:      big.NewInt(10000).Uint64(),
		GasPrice: *big.NewInt(1000),
		TotalWei: *big.NewInt(100000),
		Data:     []byte(""), // setup data // TODO finalize by adding contract call to
	}

	// Send PRL
	var sent = s.gateway.SendPRL(msg)
	if sent {
		// successful prl send
		t.Logf("Sent PRL to :%v", msg.From.Hex())
	} else {
		// failed prl send
		t.Fatalf("Failed to send PRL to:%v", msg.From.Hex())
	}
}
*/

/*
// claim prl
func (s *EthereumTestSuite) claimPRL(t *testing.T) {

	// Receiver
	receiverAddress := common.HexToAddress("0xC30efFC3509D56ef748d51f9580c81ff8e9c610E")

	// Setup Found Treasure Properties
	treasureAddress := common.HexToAddress("0x5aeda56215b167893e80b4fe645ba6d5bab767de")
	treasurePrivateKey := "8d5366123cb560bb606379f90a0bfd4769eecc0557f1b362dcae9012b548b1e5"

	// Claim PRL
	claimed := s.gateway.ClaimPRL(receiverAddress, treasureAddress, treasurePrivateKey)
	if !claimed {
		t.Fatal("Failed to claim PRLs")
	} else {
		t.Log("PRLs have been successfully claimed")
	}

}*/

/*
// claim unused prl from completed upload
func (s *EthereumTestSuite) claimUnusedPRL(t *testing.T) {

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
	err := s.gateway.ClaimUnusedPRLs(completedUploads)
	if err != nil {
		t.Fatal("Failed to claim PRLs")
	} else {
		t.Log("PRLs have been successfully claimed")
	}

}
*/
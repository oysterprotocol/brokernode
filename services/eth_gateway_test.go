package services_test

import (
	"testing"

	"github.com/oysterprotocol/brokernode/services"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/core"
	"time"
	"github.com/ethereum/go-ethereum/params"
	"log"
	"github.com/ethereum/go-ethereum/ethclient"
	"os"
	"strings"
	"context"
	"github.com/oysterprotocol/brokernode/models"
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


// wip - send gas(ether) to an address for a transaction
func Test_sendEth(t *testing.T) {
	
	// test account accepting ether
	testAcct := common.HexToAddress("0xf10a2706e98ef86b6866ae6cab2e0ca501fdf091")
	// transfer 1 ether
	transferValue := big.NewInt(1000000000000000000)

	// Send ether to test account
	txs, err := services.EthWrapper.SendETH(testAcct, transferValue)
	if err != nil {
		t.Logf("failed to send ether to %v ether to %v", transferValue, testAcct.Hex())
		t.Fatalf("transaction error: %v", err)
	}
	for tx := range txs {
		transaction := txs[tx]
		t.Logf("tx to     : %v", transaction.To().Hash().String())
		// subscribe to transaction hash
		t.Logf("tx hash   : %v", transaction.Hash().String())
		t.Logf("tx amount : %v", transaction.Value())
		t.Logf("tx cost   : %v", transaction.Cost())
	}
}

//
// Oyster Pearl Contract Tests
//

// simulated blockchain to deploy oyster pearl
func Test_deployOysterPearl(t *testing.T) {
	t.Skip(nil)
	// generate a new random account and a funded simulator
	key, _ := crypto.GenerateKey()
	auth := bind.NewKeyedTransactor(key)
	sim := backends.NewSimulatedBackend(core.GenesisAlloc{auth.From: {Balance: big.NewInt(990000000000)}})

	// initialize the context
	deadline := time.Now().Add(1000 * time.Millisecond)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	// deploy a token contract on the simulated blockchain
	_, _, token, err := services.DeployOysterPearl(&bind.TransactOpts{
		Nonce:big.NewInt(0),
		From:auth.From,
		GasLimit:params.GenesisGasLimit,
		GasPrice: big.NewInt(1),
		Context:  ctx,
		Signer:auth.Signer,
	}, sim)
	if err != nil {
		log.Fatalf("Failed to deploy oyster token contract: %v", err)
	}
	t.Logf("token :%v", token)

	// Commit all pending transactions in the simulator
	sim.Commit()


}

// testing token name access from OysterPearl Contract
// basic test which validates the existence of the contract on the network
func Test_tokenNameFromOysterPearl(t *testing.T) {
	t.Skip(nil)
	// test ethClient
	var backend, _ = ethclient.Dial("http://54.86.134.172:8080")
	// prl test 1 0x1be77862769ab791c4f95f8a2cbd0d3e07a3fd1f
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
	t.Skip(nil)
	// test ethClient
	var backend, _ = ethclient.Dial("http://54.86.134.172:8080")
	test1PRLAcct := common.HexToAddress("0x1be77862769ab791c4f95f8a2cbd0d3e07a3fd1f")
	//test2PRLAcct := common.HexToAddress("0x74ad69b41e71e311304564611434ddd59ee5d1f8")
	passPhrase := "oysterby4000"
	oysterContract := common.HexToAddress("0x84e07b9833af3d3c8e07b71b1c9c041ec5909d5d")
	oysterPearl, err := services.NewOysterPearl(oysterContract, backend)
	if err != nil {
		t.Fatalf("unable to access contract instance at : %v",err)
	}

	walletKey := os.Getenv("MAIN_WALLET_KEY")
	t.Logf("using wallet key store: %v", walletKey)

	// Create an authorized transactor and spend 1 PRL
	auth, err := bind.NewTransactor(strings.NewReader(walletKey), passPhrase)
	if err != nil {
		t.Fatalf("unable to create a new transactor : %v", err)
	}

	t.Logf("authorized transactor : %v", auth.From.Hex())

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

	// transfer single prl
	tx, err := session.Transfer(test1PRLAcct, big.NewInt(1))

	if err != nil {
		t.Fatalf("transfer failed : %v", err)
	}
	t.Logf("tx sent : %v", tx.Value())
	bal, err := session.BalanceOf(test1PRLAcct)
	if err != nil {
		t.Fatalf("balance of failed : %v", err)
	}

	t.Logf("new balance: %v", bal.Uint64())
}

//
// Oyster Pearl Tests
// all tests below will exercise the contract methods in the simulator and the private oysterby network.


// bury prl TODO provide simulated backend test and private network test
func Test_buryPRL(t *testing.T) {
	t.Skip(nil)
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

// send prl TODO provide simulated backend test and private network test
func Test_sendPRL(t *testing.T) {
	t.Skip(nil)
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
	var sent = services.EthWrapper.SendPRL(msg)
	if sent {
		// successful prl send
		t.Logf("Sent PRL to :%v", msg.From.Hex())
	} else {
		// failed prl send
		t.Fatalf("Failed to send PRL to:%v", msg.From.Hex())
	}
}

// claim prl TODO provide simulated backend test and private network test
func Test_claimPRL(t *testing.T) {
	t.Skip(nil)
	// Receiver
	receiverAddress := common.HexToAddress("0xC30efFC3509D56ef748d51f9580c81ff8e9c610E")

	// Setup Found Treasure Properties
	treasureAddress := common.HexToAddress("0x5aeda56215b167893e80b4fe645ba6d5bab767de")
	treasurePrivateKey := "8d5366123cb560bb606379f90a0bfd4769eecc0557f1b362dcae9012b548b1e5"

	// Claim PRL
	claimed := services.EthWrapper.ClaimPRL(receiverAddress, treasureAddress, treasurePrivateKey)
	if !claimed {
		t.Fatal("Failed to claim PRLs")
	} else {
		t.Log("PRLs have been successfully claimed")
	}

}


// claim unused prl from completed upload TODO provide simulated backend test and private network test
func Test_claimUnusedPRL(t *testing.T) {
	t.Skip(nil)
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
	err := services.EthWrapper.ClaimUnusedPRLs(completedUploads)
	if err != nil {
		t.Fatal("Failed to claim PRLs")
	} else {
		t.Log("PRLs have been successfully claimed")
	}

}
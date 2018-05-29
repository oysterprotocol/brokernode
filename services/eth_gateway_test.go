package services_test

import (
	"testing"

	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"log"
	"math/big"
	"time"
)

//
// Ethereum Constants
//
var oneEther = big.NewInt(1000000000000000000)
var onePrl = big.NewInt(1)

// TX_HASH = 0xd87157b84528173156a1ff06fd452300c6efaee6ea4f3d53730f190dad630327
//curl -H "Content-Type: application/json" -X POST --data \
//'{"jsonrpc":"2.0","method":"eth_getRawTransactionByHash","params":["<TX_HASH>"],"id":1}' http://54.86.134.172:8080

// Private Network
var oysterbyNetwork = "ws://54.86.134.172:8547"

//
// Ethereum Addresses
//
var ethAddress01 = common.HexToAddress("0xf10a2706e98ef86b6866ae6cab2e0ca501fdf091")
var ethAddress02 = common.HexToAddress("0x6abf0cdedd33e2bd6b574e0d81cdcaca817148c8")

//
// PRL Addresses
//
var prlBankAddress = common.HexToAddress("0x1BE77862769AB791C4f95f8a2CBD0d3E07a3FD1f")
var prlAddress01 = common.HexToAddress("0x73da066d94fc41f11c2672ed9ecd39127da30976")
var prlAddress02 = common.HexToAddress("0x74ad69b41e71e311304564611434ddd59ee5d1f8")

// Oysterby PRL Contract
var oysterContract = common.HexToAddress("0xb7baab5cad2d2ebfe75a500c288a4c02b74bc12c")

// PRL Distribution Contract
var prlDistribution = common.HexToAddress("0x08e6c245e21799c43970cd3193c59c1f6f2469ca")

//
// Utility
//

// print transaction
func printTx(tx types.Transaction) {
	fmt.Printf("tx to     : %v\n", tx.To().Hash().String())
	fmt.Printf("tx hash   : %v\n", tx.Hash().String())
	fmt.Printf("tx amount : %v\n", tx.Value())
	fmt.Printf("tx cost   : %v\n", tx.Cost())
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
func Test_checkETHBalance(t *testing.T) {

	// test balance for an ether account
	// Convert string address to byte[] address form
	bal := services.EthWrapper.CheckETHBalance(ethAddress01)
	if bal.Uint64() > 0 {
		t.Logf("balance verified: %v\n", bal)
	} else {
		t.Fatalf("balance less than zero: %v\n", bal)
	}
}

// get current block number
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

// get current block gas limit
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

// send gas(ether) to an address for a transaction
func Test_sendEth(t *testing.T) {
	t.Skip(nil)
	// transfer 1/5 of ether
	transferValue := oneEther.Div(oneEther, big.NewInt(5))

	// Send ether to test account
	txs, err := services.EthWrapper.SendETH(ethAddress01, transferValue)
	if err != nil {
		t.Logf("failed to send ether to %v ether to %v", transferValue, ethAddress01.Hex())
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

// ensure confirmation is over 12
func Test_confirmTransaction(t *testing.T) {
	t.Skip(nil)
	// tx hash
	txHash := common.HexToHash("0x718ee11d52445d72dd8f10b2349a9300a8988e46d29b49c83f8848b988f10522")
	// tx count
	txCount, err := services.EthWrapper.GetConfirmationCount(txHash)
	if err != nil {
		t.Fatalf("unable to get transaction confirmation with hash : %v", txHash)
	}
	//if txCount.Uint64() > 12 {
	//	t.Logf("confirmation have reached the golden number")
	//}
	t.Logf("txCount : %v", txCount.Uint64())
}

// send ether to an address and wait for transaction confirmation returning the new balance
func Test_sendEthAndWaitForTransfer(t *testing.T) {
	t.Skip(nil)
	// transfer 1/3 of an ether
	transferValue := oneEther

	// Send ether to test account
	txs, err := services.EthWrapper.SendETH(ethAddress02, transferValue)
	if err != nil {
		t.Logf("failed to send ether to %v ether to %v", transferValue, ethAddress02.Hex())
		t.Fatalf("transaction error: %v", err)
	}
	for tx := range txs {
		transaction := txs[tx]
		// trace out success tx
		t.Logf("tx to     : %v", transaction.To().Hash().String())
		// subscribe to transaction hash
		t.Logf("tx hash   : %v", transaction.Hash().String())
		t.Logf("tx amount : %v", transaction.Value())
		t.Logf("tx cost   : %v", transaction.Cost())
	}

	newBal, waitErr := services.EthWrapper.WaitForTransfer(ethAddress02, "eth")
	if waitErr != nil {
		t.Fatalf("wait for transfer error : %v", newBal)
	}
	t.Logf("new balance post confirmation : %v", newBal)
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
		Nonce:    big.NewInt(0),
		From:     auth.From,
		GasLimit: params.GenesisGasLimit,
		GasPrice: big.NewInt(1),
		Context:  ctx,
		Signer:   auth.Signer,
	}, sim)
	if err != nil {
		log.Fatalf("failed to deploy oyster token contract: %v", err)
	}
	t.Logf("token :%v", token)

	// Commit all pending transactions in the simulator
	sim.Commit()
}

// testing token name access from OysterPearl Contract
// basic test which validates the existence of the contract on the network
func Test_tokenNameFromOysterPearl(t *testing.T) {
	// test ethClient
	var backend, _ = ethclient.Dial(oysterbyNetwork)
	oysterPearl, err := services.NewOysterPearl(oysterContract, backend)
	if err != nil {
		t.Fatalf("unable to access contract instance at :%v", err)
	}
	name, err := oysterPearl.Name(nil)
	if err != nil {
		t.Fatalf("unable to access contract name : %v", err)
	}
	t.Logf("oyster pearl contract name :%v", name)
}

// testing token balanceOf from OysterPearl Contract account
// basic test which validates the balanceOf a PRL address
func Test_stakePRLFromOysterPearl(t *testing.T) {
	t.Skip(nil)
	// contract
	// test ethClient
	var backend, _ = ethclient.Dial(oysterbyNetwork)
	// instance of the oyster pearl contract
	pearlDistribute, err := services.NewPearlDistributeOysterby(prlDistribution, backend)

	// authentication
	walletKey := services.EthWrapper.GetWallet()
	walletAddress := walletKey.Address

	t.Logf("using wallet key store from: %v\n", walletAddress.Hex())

	block, _ := services.EthWrapper.GetCurrentBlock()

	// Create an authorized transactor and spend 1 PRL
	auth := bind.NewKeyedTransactor(walletKey.PrivateKey)
	if err != nil {
		t.Fatalf("unable to create a new transactor : %v", err)
	}
	t.Logf("authorized transactor : %v", auth.From.Hex())
	if err != nil {
		t.Fatalf("unable to access contract instance at :%v", err)
	}
	//
	// transact
	opts := bind.TransactOpts{
		From:     auth.From,
		Signer:   auth.Signer,
		GasLimit: block.GasLimit(),
	}

	// stake
	tx, err := pearlDistribute.Distribute(&opts)

	if err != nil {
		t.Fatalf("unable to access call distribute : %v", err)
	}
	t.Logf("oyster pearl distribute called :%v", tx)
	printTx(*tx)

}

// test sending PRLs from OysterPearl Contract
// issue > transfer failed : replacement transaction underpriced
// solution > increase gasPrice by 10% minimum will work.
func Test_transferPRLFromOysterPearl(t *testing.T) {

	t.Skip(nil)

	// test ethClient
	var backend, _ = ethclient.Dial(oysterbyNetwork)
	//passPhrase := "oysterby4000"
	oysterPearl, err := services.NewOysterPearl(oysterContract, backend)

	if err != nil {
		t.Fatalf("unable to access contract instance at : %v", err)
	}

	t.Logf("oysterPearl : %v", oysterPearl)

	walletKey := services.EthWrapper.GetWallet()
	walletAddress := walletKey.Address

	t.Logf("using wallet key store from: %v", walletAddress)

	// Create an authorized transactor and spend 1 PRL
	auth := bind.NewKeyedTransactor(walletKey.PrivateKey)
	if err != nil {
		t.Fatalf("unable to create a new transactor : %v", err)
	}

	t.Logf("authorized transactor : %v", auth.From.Hex())

	block, _ := services.EthWrapper.GetCurrentBlock()

	opts := bind.TransactOpts{
		From:     auth.From,
		Signer:   auth.Signer,
		GasLimit: block.GasLimit(),
	}
	tx, err := oysterPearl.Transfer(&opts, prlAddress02, big.NewInt(1))
	if err != nil {
		t.Fatalf("transfer failed : %v", err)
	}

	t.Logf("new transaction posted: %v", tx.Hash().Hex())

	printTx(*tx)

}

//
// Oyster Pearl Tests
// all tests below will exercise the contract methods in the simulator and the private oysterby network.

// bury prl
func Test_buryPRL(t *testing.T) {
	t.Skip(nil)
	// PRL based addresses
	from := prlAddress01
	to := prlAddress02

	// Gas Price
	gasPrice, _ := services.EthWrapper.GetGasPrice()
	block, _ := services.EthWrapper.GetCurrentBlock()

	var msg = services.OysterCallMsg{
		From:     from,
		To:       to,
		Amount:   *onePrl,
		Gas:      block.GasLimit(),
		GasPrice: *gasPrice,
		// this unit needs conversion
		TotalWei: *onePrl.Mul(onePrl, big.NewInt(18)),
		Data:     nil, // setup data
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

// send prl
func Test_sendPRL(t *testing.T) {

	gasPrice, _ := services.EthWrapper.GetGasPrice()
	block, _ := services.EthWrapper.GetCurrentBlock()

	var msg = services.OysterCallMsg{
		From:     prlBankAddress,
		To:       prlAddress01,
		Amount:   *big.NewInt(100),
		Gas:      block.GasLimit(),
		GasPrice: *gasPrice,
		//TotalWei: *big.NewInt(100).Mul(big.NewInt(100), big.NewInt(18)),
		Data: nil,
	}
	// Send PRL
	var sent = services.EthWrapper.SendPRL(msg)
	if sent {
		// successful prl send
		t.Logf("Sent PRL to :%v", msg.To.Hex())
	} else {
		// failed prl send
		t.Fatalf("Failed to send PRL to:%v", msg.To.Hex())
	}
}

// claim prl
func Test_claimPRL(t *testing.T) {
	t.Skip(nil)
	// Receiver
	receiverAddress := prlAddress02

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

// claim unused prl from completed upload
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

// testing token balanceOf from OysterPearl Contract account
// basic test which validates the balanceOf a PRL address
func Test_balanceOfFromOysterPearl(t *testing.T) {

	// test ethClient
	var backend, _ = ethclient.Dial(oysterbyNetwork)
	// instance of the oyster pearl contract
	oysterPearl, err := services.NewOysterPearl(oysterContract, backend)
	if err != nil {
		t.Fatalf("unable to access contract instance at :%v", err)
	}

	// oysterby balances
	bankBalance, _ := oysterPearl.BalanceOf(&bind.CallOpts{Pending: false}, prlBankAddress)
	prl01Balance, _ := oysterPearl.BalanceOf(&bind.CallOpts{Pending: false}, prlAddress01)
	prl02Balance, _ := oysterPearl.BalanceOf(&bind.CallOpts{Pending: false}, prlAddress02)

	t.Logf("oyster pearl bank address balance :%v", bankBalance)
	t.Logf("oyster pearl 01 address balance :%v", prl01Balance)
	t.Logf("oyster pearl 02 address balance :%v", prl02Balance)

}

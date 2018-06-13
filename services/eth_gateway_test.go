package services_test

import (
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
	"testing"
	"time"
	"reflect"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"io/ioutil"
	"os"
)

//
// Ethereum Constants
//
var oneEther = big.NewInt(1)
var onePrl = big.NewInt(1)
var oneWei = big.NewInt(1000000000000000000)

// TX_HASH = 0xd87157b84528173156a1ff06fd452300c6efaee6ea4f3d53730f190dad630327
//curl -H "Content-Type: application/json" -X POST --data \
//'{"jsonrpc":"2.0","method":"eth_getRawTransactionByHash","params":["<TX_HASH>"],"id":1}' http://54.86.134.172:8080

// Private Network
var oysterbyNetwork = "ws://54.86.134.172:8547"

//
// Ethereum Addresses
//
var ethCoinbase  = common.HexToAddress("0x919410005b53d6497517b9ad58c23c6a30207747")
//
// PRL Addresses
//
var prlBankAddress = common.HexToAddress("0x1BE77862769AB791C4f95f8a2CBD0d3E07a3FD1f")

var ethAccounts = map[string]interface{}{
	"ethAddress01": "0xf10a2706e98ef86b6866ae6cab2e0ca501fdf091",
	"ethAddress02": "0x6abf0cdedd33e2bd6b574e0d81cdcaca817148c8",
}

var prlAccounts = map[string]interface{}{
	"prlAddress01": "0x73da066d94fc41f11c2672ed9ecd39127da30976",
	"prlAddress02": "0x73da066d94fc41f11c2672ed9ecd39127da30976",
}

var ethAddress01 = common.HexToAddress(fmt.Sprint(ethAccounts["ethAddress01"]))
var ethAddress02 = common.HexToAddress(fmt.Sprint(ethAccounts["ethAddress02"]))
var prlAddress01 = common.HexToAddress(fmt.Sprint(ethAccounts["prlAddress01"]))
var prlAddress02 = common.HexToAddress(fmt.Sprint(ethAccounts["prlAddress02"]))

// Oysterby PRL Contract
var oysterContract = common.HexToAddress("0xb7baab5cad2d2ebfe75a500c288a4c02b74bc12c")

// PRL Distribution Contract
var prlDistribution = common.HexToAddress("0x08e6c245e21799c43970cd3193c59c1f6f2469ca")

// Transaction Placeholders
var knownTransaction = false
var lastTransaction = types.Transaction{}

//
// Utility
//
// ether to wei
func toWei(value int64) (uint64) {
	transferValue := big.NewInt(value)
	transferValueInWei := new(big.Int).Mul(transferValue, oneWei)
	return transferValueInWei.Uint64()
}
// print transaction
func printTx(tx *types.Transaction) {
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
	t.Skip(nil)
	//services.RunOnTestNet()
	
	// generate eth address using gateway
	addr, privateKey, err := services.EthWrapper.GenerateEthAddr()
	if err != nil {
		t.Fatalf("error creating ethereum network address")
	}
	// ensure address is correct format
	if common.IsHexAddress(addr.Str()) {
		t.Fatalf("could not create a valid ethereum network address:%v", addr.Hex())
	}
	// ensure private key was returned
	if privateKey == "" {
		t.Fatalf("could not create a valid private key")
	}
	t.Logf("ethereum network address was generated %v\n", addr.Hex())
}

// generate address from private key test
func Test_generateEthAddrFromPrivateKey(t *testing.T) {
	//services.RunOnTestNet()
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
	//services.RunOnTestNet()
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
	//services.RunOnTestNet()
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
	//services.RunOnTestNet()
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
	//services.RunOnTestNet()
	// Get the current block from the network
	block, err := services.EthWrapper.GetCurrentBlock()
	if err != nil {
		t.Fatalf("could not retrieve the current block: %v\n", err)
	}
	if block != nil {
		t.Logf("retrieved the current block gas limit: %v\n", block.GasLimit())
	}
}

func Test_getNonceForAccount(t *testing.T) {
	//services.RunOnTestNet()
	// Get the nonce for the given account
	nonce, err := services.EthWrapper.GetNonce(context.Background(), ethCoinbase)
	if err != nil {
		t.Fatalf("unable to access the account nonce : %v", err)
	}
	if nonce > 0 {
		t.Logf("valid account nonce : %v", nonce)
	} else {
		t.Fatalf("invalid account nonce : %v", nonce)
	}
}

// send gas(ether) to an address for a transaction
func Test_sendEth(t *testing.T) {
	t.Skip(nil)
	//services.RunOnTestNet()
	// transfer 1/5 of ether
	transferValue := big.NewInt(10)
	//transferValueInWei := new(big.Int).Mul(transferValue, oneWei)
	// Send ether to test account
	txs, err := services.EthWrapper.SendETH(ethAddress02, transferValue)
	if err != nil {
		t.Logf("failed to send ether to %v ether to %v\n", transferValue, ethAddress02.Hex())
		t.Fatalf("transaction error: %v\n", err)
	}
	for tx := range txs {
		transaction := txs[tx]
		// Store For Next Test
		lastTransaction = *transaction
		printTx(transaction)
		
		// wait for confirmation
		confirmed := services.EthWrapper.WaitForConfirmation(lastTransaction.Hash())
		if confirmed == 1 {
			t.Logf("confirmed ether was sent to : %v", ethAddress02.Hex())
		} else if confirmed == 0 {
			t.Logf("failed to confirm sending ether")
		}
	}
}

// ensure the transaction is stored in the transactions table
// it is accessed with the lastTransactionHash from the previous test
func Test_ensureTransactionStoredInPool(t *testing.T) {
	t.Skip(nil)
	// get item by txHash and ensure its in the table
	txWithBlockNumber := services.EthWrapper.GetTransaction(lastTransaction.Hash())
	if txWithBlockNumber.Transaction != nil {
		// compare transaction hash
		if reflect.DeepEqual(txWithBlockNumber.Transaction.Hash(), lastTransaction.Hash()) {
			t.Log("transaction is stored on the transactions table")
		} else {
			t.Fatal("transaction should be stored in the transaction table, post sendEth")
		}
	}
}


// ensure confirmation is made with last transaction hash from sendEth
func Test_confirmTransactionStatus(t *testing.T) {
	t.Skip(nil)
	//services.RunOnTestNet()
	
	txStatus := services.EthWrapper.WaitForConfirmation(lastTransaction.Hash())
	if txStatus == 0 {
		t.Logf("transaction failure")
	} else if txStatus == 1 {
		t.Logf("confirmation completed")
		
		bal := services.EthWrapper.CheckETHBalance(ethAddress02)
		t.Logf("balance updated : %v", bal)
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

// simulated blockchain to test bury,
// claim with a contract with buried address
func Test_simOysterPearlBury(t *testing.T) {
	t.Skip(nil)
	
	// generate a new random account and a funded simulator
	key, _ := crypto.GenerateKey()
	auth := bind.NewKeyedTransactor(key)
	fundedSupply := toWei(900)
	sim := backends.NewSimulatedBackend(core.GenesisAlloc{auth.From: {Balance: big.NewInt(0).SetUint64(fundedSupply)}})
	
	sim.Commit()
	
	// initialize the context
	deadline := time.Now().Add(3000 * time.Millisecond)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()
	//
	// deploy a token contract on the simulated blockchain
	//
	_, _, token, err := services.DeployOysterPearl(&bind.TransactOpts{
		Nonce:    big.NewInt(0),
		From:     auth.From,
		GasLimit: params.GenesisGasLimit,
		GasPrice: big.NewInt(0).SetUint64(params.TxGasContractCreation),
		Context:  ctx,
		Signer:   auth.Signer,
	}, sim)
	if err != nil {
		log.Fatalf("failed to deploy oyster token contract: %v", err)
	}
	t.Logf("token :%v", token)
	
	// commit all pending transactions in the simulator
	sim.Commit()
	
	// claim
	gasPrice, _ := sim.SuggestGasPrice(ctx)
	claimKey, _ := crypto.GenerateKey()
	claimAuth := bind.NewKeyedTransactor(claimKey)
	payoutAddress := claimAuth.From
	fee := auth.From
	tx, err := token.Claim(&bind.TransactOpts{
		Nonce: big.NewInt(0),
		From: claimAuth.From,
		GasLimit: params.GenesisGasLimit,
		GasPrice: gasPrice,
		Context:  ctx,
		Signer: claimAuth.Signer,
	}, payoutAddress, fee)
	t.Logf("claim transaction submitted : %v", tx.Hash())
	
	// send
	sendErr := sim.SendTransaction(ctx, tx)
	if sendErr != nil {
		t.Logf("failed to send transaction for claim : %v", sendErr)
	}
	
	sim.Commit()
	
	// create ethereum address and key
	buryKey, _ := crypto.GenerateKey()
	buryAuth := bind.NewKeyedTransactor(buryKey)
	
	// bury
	buryTx, err := token.Bury(&bind.TransactOpts{
		Nonce: big.NewInt(0),
		From: buryAuth.From,
		GasLimit: params.GenesisGasLimit,
		GasPrice: gasPrice,
		Context:  ctx,
		Signer:buryAuth.Signer,
	})
	
	// send
	buryErr := sim.SendTransaction(ctx, buryTx)
	if buryErr != nil {
		t.Logf("failed to send transaction for bury : %v", buryErr)
	}
	
}

// testing token name access from OysterPearl Contract
// basic test which validates the existence of the contract on the network
func Test_tokenNameFromOysterPearl(t *testing.T) {
	t.Skip(nil)
	//services.RunOnTestNet()

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
	walletAddress := services.MainWalletAddress

	t.Logf("using wallet key store from: %v\n", walletAddress.Hex())

	block, _ := services.EthWrapper.GetCurrentBlock()

	// Create an authorized transactor and spend 1 PRL
	auth := bind.NewKeyedTransactor(services.MainWalletPrivateKey)
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
	// printTx(tx)

}

// test sending PRLs from OysterPearl Contract
// issue > transfer failed : replacement transaction underpriced
// solution > increase gasPrice by 10% minimum will work.
func Test_transferPRLFromOysterPearl(t *testing.T) {
	t.Skip(nil)
	
	prlValue := big.NewInt(0).SetUint64(toWei(5))
	sent := services.EthWrapper.SendPRL(services.OysterCallMsg{
		Amount: *prlValue,
	})
	
	if sent {
		fmt.Printf("sent the transaction to the ")
	}
	
	// test ethClient
	var backend, _ = ethclient.Dial(oysterbyNetwork)
	oysterPearl, err := services.NewOysterPearl(oysterContract, backend)

	if err != nil {
		t.Fatalf("unable to access contract instance at : %v", err)
	}
	
	// access prl wallet
	prlWallet := getWallet()
	walletAddress := prlWallet.Address

	t.Logf("using wallet key store from: %v", walletAddress)

	// Create an authorized transactor and spend 1 PRL
	auth := bind.NewKeyedTransactor(prlWallet.PrivateKey)
	if err != nil {
		t.Fatalf("unable to create a new transactor : %v", err)
	}

	t.Logf("authorized transactor : %v", auth.From.Hex())
	
	block, _ := services.EthWrapper.GetCurrentBlock()
	gasPrice, err := backend.SuggestGasPrice(context.Background())
	
	opts := bind.TransactOpts{
		From:     auth.From,
		Signer:   auth.Signer,
		GasLimit: block.GasLimit(),
		GasPrice: gasPrice,
		Value: prlValue,
		Nonce: auth.Nonce,
	}
	
	tx, err := oysterPearl.Transfer(&opts,prlAddress02, prlValue)
	if err != nil {
		t.Fatalf("transfer failed : %v", err)
	}

	t.Logf("transfer pending: 0x%x\n", tx.Hash())
	
	printTx(tx)

}

//
// Oyster Pearl Tests
// all tests below will exercise the contract methods in the simulator and the private oysterby network.

// bury prl
func Test_buryPRL(t *testing.T) {
	t.Skip(nil)
	// PRL based addresses
	
	prlWallet := getPRLWallet()
	prlValue := big.NewInt(5)

	// Gas Price
	gasPrice, _ := services.EthWrapper.GetGasPrice()
	block, _ := services.EthWrapper.GetCurrentBlock()

	buryMsg := services.OysterCallMsg {
		From:     prlWallet.Address,
		To:       prlAddress02,
		Amount:   *prlValue,
		Gas:      block.GasLimit(),
		GasPrice: *gasPrice,
		TotalWei: *big.NewInt(0).SetUint64(toWei(prlValue.Int64())),
		PrivateKey: *prlWallet.PrivateKey,
		Data:     nil,
	}

	// Bury PRL
	var buried = services.EthWrapper.BuryPrl(buryMsg)
	if buried {
		// successful bury attempt
		t.Log("Buried the PRLs successfully")
	} else {
		// failed bury attempt
		t.Fatal("Faild to bury PRLs. Try Again?")
	}
}

// send prl from main wallet address to another address
func Test_sendPRL(t *testing.T) {
	//t.Skip(nil)
	
	// same as the prl bank address
	prlWallet := getPRLWallet()
	prlValue := big.NewInt(5)
	
	// compose message
	var sendMsg = services.OysterCallMsg{
		From:     prlWallet.Address,
		To:       prlAddress02,
		Amount:   *prlValue,
		TotalWei: *big.NewInt(0).SetUint64(toWei(prlValue.Int64())),
		PrivateKey: *prlWallet.PrivateKey,
	}
	
	// Send PRL is a blocking call which will send the new transaction to the network
	// then wait for the confirmation to return true or false
	var confirmed = services.EthWrapper.SendPRL(sendMsg)
	if confirmed {
		// successful prl send
		t.Logf("Sent PRL to :%v", sendMsg.To.Hex())
	} else {
		// failed prl send
		t.Fatalf("Failed to send PRL to:%v", sendMsg.To.Hex())
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

// testing balance of the prl account for a given address
func Test_balanceOfFromOysterPearl(t *testing.T) {
	//t.Skip(nil)
	
	// pearl balances
	bankBalance := services.EthWrapper.CheckPRLBalance(prlBankAddress)
	t.Logf("oyster pearl bank address balance :%v", bankBalance)
	
	prl00Balance := services.EthWrapper.CheckPRLBalance(common.HexToAddress("0x0000000000000000000000000000000000000000"))
	t.Logf("oyster pearl 00   address balance :%v", prl00Balance)
	
	prl01Balance := services.EthWrapper.CheckPRLBalance(prlAddress01)
	prl02Balance := services.EthWrapper.CheckPRLBalance(prlAddress02)

	t.Logf("oyster pearl 01   address balance :%v", prl01Balance)
	t.Logf("oyster pearl 02   address balance :%v", prl02Balance)
}

// utility to access the return the PRL wallet keystore
func getPRLWallet() *keystore.Key {
	
	// load local test wallet key, may need to pull ahead vs on-demand
	walletKeyJSON, err := ioutil.ReadFile("testdata/prl.prv")
	
	if err != nil {
		fmt.Printf("error loading the walletKey : %v", err)
	}
	// decrypt wallet
	walletKey, err := keystore.DecryptKey(walletKeyJSON, os.Getenv("MAIN_WALLET_PW"))
	if err != nil {
		fmt.Printf("walletKey err : %v", err)
	}
	return walletKey
}

// utility to access the return the PRL wallet keystore
func getWallet() *keystore.Key {
	
	// load local test wallet key, may need to pull ahead vs on-demand
	walletKeyJSON, err := ioutil.ReadFile("testdata/key.prv")
	
	if err != nil {
		fmt.Printf("error loading the walletKey : %v", err)
	}
	// decrypt wallet
	walletKey, err := keystore.DecryptKey(walletKeyJSON, os.Getenv("MAIN_WALLET_PW"))
	if err != nil {
		fmt.Printf("walletKey err : %v", err)
	}
	return walletKey
}
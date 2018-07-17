package services_test

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/oysterprotocol/brokernode/services"
	"io/ioutil"
	"math/big"
	"os"
	"reflect"
	"testing"
)

//
// Ethereum Constants
//
var oneEther = big.NewInt(1)
var oneWei = big.NewInt(1000000000000000000)

// TX_HASH = 0xd87157b84528173156a1ff06fd452300c6efaee6ea4f3d53730f190dad630327
//curl -H "Content-Type: application/json" -X POST --data \
//'{"jsonrpc":"2.0","method":"eth_getRawTransactionByHash","params":["<TX_HASH>"],"id":1}' http://54.86.134.172:8080

// Private Network
var oysterbyNetwork = os.Getenv("ETH_NODE_URL")

//
// Ethereum Addresses
//
var ethCoinbase = common.HexToAddress("0x919410005b53d6497517b9ad58c23c6a30207747")

//
// PRL Addresses
//
var prlBankAddress = common.HexToAddress("0x1BE77862769AB791C4f95f8a2CBD0d3E07a3FD1f")

var ethAccounts = map[string]interface{}{
	"ethAddress01": "0xf10a2706e98ef86b6866ae6cab2e0ca501fdf091",
	"ethAddress02": "0x6abf0cdedd33e2bd6b574e0d81cdcaca817148c8",
}

var prlAccounts = map[string]interface{}{
	"prlAddress01": "0x1BE77862769AB791C4f95f8a2CBD0d3E07a3FD1f",
	"prlAddress02": "0x73da066d94FC41f11C2672ed9ecD39127DA30976",
	"prlAddress03": "0x74aD69B41e71E311304564611434dDD59Ee5d1F8",
}

var ethAddress01 = common.HexToAddress(fmt.Sprint(ethAccounts["ethAddress01"]))
var ethAddress02 = common.HexToAddress(fmt.Sprint(ethAccounts["ethAddress02"]))
var prlAddress01 = common.HexToAddress(fmt.Sprint(ethAccounts["prlAddress01"]))
var prlAddress02 = common.HexToAddress(fmt.Sprint(ethAccounts["prlAddress02"]))
var prlAddress03 = common.HexToAddress(fmt.Sprint(ethAccounts["prlAddress03"]))

var ethFile = "testdata/prl.prv"
var prl1File = "testdata/prl.prv"
var prl2File = "testdata/prl2.prv"
var prl3File = "testdata/prl3.prv"

// PRL Distribution Contract
var prlDistribution = common.HexToAddress("0x08e6c245e21799c43970cd3193c59c1f6f2469ca")

// Transaction Placeholders
var knownTransaction = false
var lastTransaction = types.Transaction{}
var lastTransactionHash = common.HexToHash("0x91676473dad2d9e4c777ac1bbbc1a2d8548ab97d8332d9ca551de1df79c958b3")

//
// Utility
//
// ether to wei
func toWei(value int64) uint64 {
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
	// generate eth address using gateway
	addr, privateKey, err := services.EthWrapper.GenerateEthAddr()
	if err != nil {
		t.Fatalf("error creating ethereum network address")
	}
	// ensure address is correct format

	if common.IsHexAddress(addr.Hex()[1:]) {
		t.Fatalf("could not create a valid ethereum network address:%v", addr.Hex()[1:])
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
	//t.Skip(nil)
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

// check if it's worth it to try and reclaiming eth from an address
func Test_checkIfWorthReclaimingGas(t *testing.T) {
	worthIt, amountToReclaim, err := services.EthWrapper.CheckIfWorthReclaimingGas(ethAddress01, services.GasLimitETHSend)

	if worthIt {
		t.Logf("Should try to reclaim gas: %v\n", "true")
	} else {
		t.Logf("Should try to reclaim gas: %v\n", "false")
	}

	t.Logf("Will attempt to reclaim this much: %v\n", amountToReclaim.String())

	if err != nil {
		t.Fatalf("Received an error: %v\n", err.Error())
	}
}

// reclaim leftover gas
func Test_reclaimGas(t *testing.T) {

	gasToReclaim := big.NewInt(5000)
	prlWallet := getWallet(prl2File)

	success := services.EthWrapper.ReclaimGas(prlWallet.Address, prlWallet.PrivateKey, gasToReclaim)

	// This isn't a very good test because it succeeds regardless of the outcome?

	if success {
		t.Logf("Reclaim gas success: %v\n", "true")
	} else {
		t.Logf("Reclaim gas success: %v\n", "false")
	}
}

// get gas price from network test
func Test_calculateGasNeeded(t *testing.T) {

	gasPrice, err := services.EthWrapper.GetGasPrice()
	gasLimitToUse := services.GasLimitETHSend

	expectedGasToSend := new(big.Int).Mul(gasPrice, big.NewInt(int64(gasLimitToUse)))

	gasToSend, err := services.EthWrapper.CalculateGasNeeded(gasLimitToUse)
	if expectedGasToSend.Int64() != gasToSend.Int64() {
		t.Fatalf("failed to calculate the gas to send: %v\n", err)
	}
	if expectedGasToSend.Int64() == gasToSend.Int64() && gasToSend.Int64() > 0 {
		t.Logf("successfully calculated gas to send: %v\n", gasToSend)
	} else if gasToSend.Int64() <= 0 {
		t.Fatalf("calculated a gas to send amount less than zero: %v\n", gasPrice)
	}
}

// check balance on test network test
func Test_checkETHBalance(t *testing.T) {
	//services.RunOnTestNet()
	//t.Skip(nil)
	// test balance for an ether account
	// Convert string address to byte[] address form
	bal := services.EthWrapper.CheckETHBalance(ethAddress01)
	if bal.Int64() != -1 {
		t.Logf("balance verified: %v\n", bal)
	} else {
		t.Fatalf("could not get balance")
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
	} else {
		t.Logf("valid account nonce : %v", nonce)
	}
}

// send gas(ether) to an address for a transaction
func Test_sendEth(t *testing.T) {
	t.Skip(nil)
	// services.RunOnTestNet()
	// transfer
	transferValue := big.NewInt(1)
	transferValueInWei := new(big.Int).Mul(transferValue, oneWei)
	// Send ether to test account
	// wallet := getWallet(ethFile)
	to := common.HexToAddress("0xf10a2706e98ef86b6866ae6cab2e0ca501fdf091")
	txs, _, _, err := services.EthWrapper.SendETH(services.MainWalletAddress, services.MainWalletPrivateKey, to, transferValueInWei)
	//txs, _, _, err := services.EthWrapper.SendETH(from, services.MainWalletPrivateKey, to, transferValueInWei)
	if err != nil {
		t.Logf("failed to send ether to %v ether to %v\n", transferValue, to.Hex())
		t.Fatalf("transaction error: %v\n", err)
	}
	for tx := range txs {
		transaction := txs[tx]
		// Store For Next Test
		lastTransaction = *transaction
		printTx(transaction)

		// wait for confirmation
		confirmed := services.EthWrapper.WaitForConfirmation(lastTransaction.Hash(), 3)
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
	txHash := lastTransaction.Hash()
	if len(txHash) <= 0 {
		// set an existing tx hash
		txHash = lastTransactionHash
	}
	// check pending
	isPending := services.EthWrapper.PendingConfirmation(txHash)
	if isPending {
		// get item by txHash and ensure its in the table
		txWithBlockNumber := services.EthWrapper.GetTransaction(txHash)
		if txWithBlockNumber.Transaction != nil {
			// compare transaction hash
			if reflect.DeepEqual(txWithBlockNumber.Transaction.Hash(), txHash) {
				t.Log("transaction is stored on the transactions table")
			} else {
				t.Fatal("transaction should be stored in the transaction table, post sendEth")
			}
		}
	}
}

// ensure confirmation is made with last transaction hash from sendEth
func Test_confirmTransactionStatus(t *testing.T) {

	//services.RunOnTestNet()
	txHash := lastTransaction.Hash()
	if len(txHash) <= 0 {
		// set an existing tx hash
		txHash = lastTransactionHash
	}
	// check pending
	isPending := services.EthWrapper.PendingConfirmation(txHash)
	if isPending {
		// check confirmation
		txStatus := services.EthWrapper.WaitForConfirmation(txHash, 3)
		if txStatus == 0 {
			t.Logf("transaction failure")
		} else if txStatus == 1 {
			t.Logf("confirmation completed")

			bal := services.EthWrapper.CheckETHBalance(ethAddress02)
			t.Logf("balance updated : %v", bal)
		}
	}
}

// utility to access the return the PRL wallet keystore
func getWallet(fileName string) *keystore.Key {

	// load local test wallet key, may need to pull ahead vs on-demand
	walletKeyJSON, err := ioutil.ReadFile(fileName)

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

package services_test

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/oysterprotocol/brokernode/services"
	"math/big"
	"testing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/crypto/sha3"
)

//
// Oyster Pearl Contract and Services Tests
//

var onePrlWei = big.NewInt(1000000000000000000)

// Oysterby PRL Contract
var oysterContract = common.HexToAddress("0xB7baaB5caD2D2ebfE75A500c288A4c02B74bC12c")

// utility to generate account
func generateAuthKey() (*bind.TransactOpts) {
	// generate a new random account and a funded simulator
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		fmt.Errorf("%v", err)
	}
	return bind.NewKeyedTransactor(privateKey)
}

// utility to create simulator
func createSimulator(auth *bind.TransactOpts, key *ecdsa.PrivateKey) *backends.SimulatedBackend {
	return backends.NewSimulatedBackend(core.GenesisAlloc{
		auth.From: {
			Balance:    big.NewInt(23230000000000000),
			PrivateKey: crypto.FromECDSA(key),
			Nonce:      0,
		},
	})
}


// utility to deploy the oyster pearl on the simulated blockchain
func deployOysterPearl(auth *bind.TransactOpts, sim *backends.SimulatedBackend) (common.Address, *services.OysterPearl, error) {
	// deploy a token contract on the simulated blockchain
	// common.Address, *types.Transaction, *OysterPearl, error)
	contractAddress, _, token, err := services.DeployOysterPearl(auth, sim)

	if err != nil {
		fmt.Errorf("failed to deploy oyster token contract: %v\n", err)
		return common.HexToAddress(""), nil, err
	}

	return contractAddress, token, err
}

// transfer tokens in the simulator to setup the transfers
func transferTokens(sim *backends.SimulatedBackend) {
	
	privateKey, err := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19")
	if err != nil {
		fmt.Errorf("%s",err)
	}
	
	auth := bind.NewKeyedTransactor(privateKey)
	
	balance := new(big.Int)
	balance.SetString("10000000000000000000", 10) // 10 eth in wei
	
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		fmt.Errorf("error casting public key to ECDSA")
	}
	
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := sim.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		fmt.Errorf("%s",err)
	}
	
	value := big.NewInt(0) // 0 for token transfer wei (0 eth)
	gasPrice, err := sim.SuggestGasPrice(context.Background())
	if err != nil {
		fmt.Errorf("%s",err)
	}
	
	toAddress := common.HexToAddress("0x4592d8f8d7b001e72cb26a73e4fa1806a51ac79d")
	tokenAddress := common.HexToAddress("0x28b149020d2152179873ec60bed6bf7cd705775d")
	
	transferFnSignature := []byte("transfer(address,uint256)")
	hash := sha3.NewKeccak256()
	hash.Write(transferFnSignature)
	methodID := hash.Sum(nil)[:4]
	fmt.Println(hexutil.Encode(methodID)) // 0xa9059cbb
	
	paddedAddress := common.LeftPadBytes(toAddress.Bytes(), 32)
	fmt.Println(hexutil.Encode(paddedAddress)) // 0x0000000000000000000000004592d8f8d7b001e72cb26a73e4fa1806a51ac79d
	
	amount := new(big.Int)
	amount.SetString("1000000000000000000000", 10) // 1000 tokens
	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)
	fmt.Println(hexutil.Encode(paddedAmount)) // 0x00000000000000000000000000000000000000000000003635c9adc5dea00000
	
	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)
	
	gasLimit, err := sim.EstimateGas(context.Background(), ethereum.CallMsg{
		To:   &toAddress,
		Data: data,
	})
	if err != nil {
		fmt.Errorf("%s",err)
	}
	fmt.Println(gasLimit) // 23256
	
	tx := types.NewTransaction(nonce, tokenAddress, value, gasLimit, gasPrice, data)
	signedTx, err := types.SignTx(tx, types.HomesteadSigner{}, privateKey)
	if err != nil {
		fmt.Errorf("%s",err)
	}
	
	err = sim.SendTransaction(context.Background(), signedTx)
	if err != nil {
		fmt.Errorf("%s",err)
	}
	
	fmt.Printf("tx sent: %s", signedTx.Hash().Hex()) // tx sent: 0xa56316b637a94c4cc0331c73ef26389d6c097506d581073f927275e7a6ece0bc
	
	sim.Commit()
	
	receipt, err := sim.TransactionReceipt(auth.Context, signedTx.Hash())
	if err != nil {
		fmt.Errorf("failed to send transaction for transfer : %v", err)
	}
	//// 0 = failed, 1 = success
	if receipt.Status == 0 {
		fmt.Errorf("tx receipt error : %v", receipt.Status)
	} else if receipt.Status == 1 {
		fmt.Printf("tx receipt success : %v", receipt.Status)
	}
	
}

// get receipt for a given transfer
func getReceipt(t *testing.T, sim *backends.SimulatedBackend, tx *types.Transaction) {
	// get transaction receipt
	receipt, err := sim.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		t.Fatalf("failed to send transaction for transfer : %v", err)
	}
	// 0 = failed, 1 = success
	if receipt.Status == 0 {
		t.Fatalf("tx receipt error : %v", receipt.Status)
	} else if receipt.Status == 1 {
		t.Logf("tx receipt success : %v", receipt.Status)
	}
}

// get balance from an account address
func getBalance(t *testing.T, sim *backends.SimulatedBackend, address common.Address) {
	// get balance
	balance, err := sim.BalanceAt(context.Background(), address, nil)
	if err != nil {
		t.Fatalf("balance check failed")
	}
	fmt.Printf("new balance for address : %v", balance.Uint64())
}

type OysterAgent struct {
	sim *backends.SimulatedBackend
	currency string
}

func (e OysterAgent) setSimulator () {
	// generate a new random account and a funded simulator
	key, _ := crypto.GenerateKey()
	auth := bind.NewKeyedTransactor(key)
	// set simulator
	e.sim = createSimulator(auth, key)
}

// simulated blockchain to deploy oyster pearl
func Test_deployOysterPearl(t *testing.T) {
	
	// generate a new random account and a funded simulator
	key, _ := crypto.GenerateKey()
	auth := bind.NewKeyedTransactor(key)

	// initialize simulator
	sim := createSimulator(auth, key)

	// deploy a token contract on the simulated blockchain
	contractAddress, _, _, err := services.DeployOysterPearl(auth, sim)

	if err != nil {
		t.Fatalf("error deploying oyster pearl : %v", err)
	}
	// contract address
	if common.IsHexAddress(contractAddress.Hex()) {
		t.Logf("deployed contract address : %v", contractAddress.Hex())
	}
	
	// Commit to start a new state
	sim.Commit()
	
	// token instance by contract address
	token, err := services.NewOysterPearl(contractAddress, sim)
	// token name test
	tokenName, err := token.Name(nil)
	if err != nil {
		t.Fatalf("error getting oyster pearl name : %v", err)
	}
	t.Logf("deployed contract name : %v", tokenName)

	// total supply value from token deployment
	totalSupply, err := token.TotalSupply(nil)
	if err != nil {
		t.Fatalf("error getting oyster pearl totalSupply : %v", err)
	}
	t.Logf("deployed contract totalSupply : %v", totalSupply)
	
}

// bury failure due to criteria not met
// An address must have at least 'claimAmount' to be buried
// > solidity require(balances[msg.sender] >= claimAmount);
func Test_simOysterPearlBuryFail(t *testing.T) 	{
	
	// generate a new random account and a funded simulator
	key, _ := crypto.GenerateKey()
	auth := bind.NewKeyedTransactor(key)
	
	// initialize simulator
	sim := createSimulator(auth, key)
	
	// deploy a token contract on the simulated blockchain
	contractAddress, _, _, err := services.DeployOysterPearl(auth, sim)
	
	if err != nil {
		t.Fatalf("error deploying oyster pearl : %v", err)
	}
	
	// Commit to start a new state
	sim.Commit()
	
	// token instance by contract address
	token, err := services.NewOysterPearl(contractAddress, sim)
	
	// setup tx properties
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000)     // in units
	
	gasPrice, err := sim.SuggestGasPrice(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("gasPrice : %v", gasPrice)
	
	auth.GasPrice = gasPrice
	
	// token bury test
	tx, err := token.Bury(auth)
	if err != nil {
		t.Fatalf("error getting oyster pearl name : %v", err)
	}
	
	sim.Commit()
	
	t.Logf("deployed contract bury : %v", tx.Hash())
	
	printTx(tx)
	
	// tx receipt for opposite effect
	receipt, err := sim.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		t.Fatal(err)
	}
	
	// tx receipt status should be failure since there is no balance to allow bury
	// @see (balances[msg.sender] >= claimAmount)
	if receipt.Status == 1 {
		t.Fatalf("bury success: %v", receipt.Status) // status: 1
	} else {
		t.Logf("bury failed: %v", receipt.Status) // status: 0
	}
	
}

// simulated blockchain to test ether transfer
func Test_simTransferEth(t *testing.T) {
	
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	
	auth := bind.NewKeyedTransactor(privateKey)
	
	balance := new(big.Int)
	balance.SetString("10000000000000000000", 10) // 10 eth in wei
	
	address := auth.From
	genesisAlloc := map[common.Address]core.GenesisAccount{
		address: {
			Balance: balance,
		},
	}
	
	sim := backends.NewSimulatedBackend(genesisAlloc)
	
	fromAddress := auth.From
	nonce, err := sim.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		t.Fatal(err)
	}
	
	value := big.NewInt(400)
	gasLimit := uint64(21000)
	gasPrice, err := sim.SuggestGasPrice(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	// mock address
	toAddress := common.HexToAddress("0x4592d8f8d7b001e72cb26a73e4fa1806a51ac79d")
	var data []byte
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)
	signedTx, err := types.SignTx(tx, types.HomesteadSigner{}, privateKey)
	if err != nil {
		t.Fatal(err)
	}
	
	err = sim.SendTransaction(context.Background(), signedTx)
	if err != nil {
		t.Fatal(err)
	}
	
	fmt.Printf("tx sent: %s\n", signedTx.Hash().Hex())
	
	// mine transaction
	sim.Commit()
	
	// mined check receipt
	getReceipt(t, sim, signedTx)
	
	// get balance for toAddress from simulator blockchain
	bal, err := sim.BalanceAt(context.Background(), toAddress, nil)
	if err != nil {
		t.Fatalf("balance not available : %v", err)
	}
	t.Logf("balance post transfer : %v", bal.Uint64())
}

//
// Oyster Pearl Tests
// all tests below will exercise the contract methods in the simulator and the private oysterby network.

// send prl from main oyster pearl to another address
func Test_sendPRL(t *testing.T) {
	
	// send PRL via transaction from sim OysterPearl token
	// generate a new random account and a funded simulator
	key, _ := crypto.GenerateKey()
	auth := bind.NewKeyedTransactor(key)
	
	// initialize simulator
	sim := createSimulator(auth, key)
	
	// deploy a token contract on the simulated blockchain
	contractAddress, _, _, err := services.DeployOysterPearl(auth, sim)
	if err != nil {
		t.Fatalf("error deploying oyster pearl : %v", err)
	}
	// contract address
	if common.IsHexAddress(contractAddress.Hex()) {
		t.Logf("deployed contract address : %v", contractAddress.Hex())
	}
	
	// Commit to start a new state
	sim.Commit()
	
	// token instance by contract address
	token, err := services.NewOysterPearl(contractAddress, sim)
	
	// create toAddress to get the tokens
	toKey, _ := crypto.GenerateKey()
	toAuth := bind.NewKeyedTransactor(toKey)
	toAddress := toAuth.From
	
	fmt.Printf("transferring tokens")
	
	//transferTokens(sim)
	value := big.NewInt(1000)
	auth.GasLimit = uint64(300000)     // in units
	
	gasPrice, err := sim.SuggestGasPrice(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("gasPrice : %v", gasPrice)
	
	auth.GasPrice = gasPrice
	
	tx, err := token.Transfer(auth, toAddress, value)
	if err != nil {
		t.Fatalf("error transferring from oyster pearl : %v", err)
	}
	
	printTx(tx)
	
	sim.Commit()
	
	fmt.Println("transferring prl tokens completed")
	
	// print receipt status
	getReceipt(t, sim, tx)
	
	// get balance for toAddress from token
	balance, err := token.BalanceOf(&bind.CallOpts{From:auth.From,Pending:true}, toAddress)
	if err != nil {
		t.Fatalf("balance check failed")
	}
	t.Logf("new balance for address : %v", balance.Uint64())
	
}

// simulated blockchain to test bury,
// Contract Level Logic
//
// An address must have at least 'claimAmount' to be buried
// > solidity require(balances[msg.sender] >= claimAmount);
// Prevent addresses with large balances from getting buried
// > solidity require(balances[msg.sender] <= retentionMax);
func Test_simOysterPearlBurySuccess(t *testing.T) {
	
	t.Skip(nil)
	
	// generate a new random account and a funded simulator
	key, _ := crypto.GenerateKey()
	auth := bind.NewKeyedTransactor(key)
	
	// Bury an address in the contract
	// 0.5 prl - lower bound
	// claimAmount
	// 40 prl  - upper bound
	// retentionMax
	
	// create toAddress to get the tokens
	toKey, _ := crypto.GenerateKey()
	toAuth := bind.NewKeyedTransactor(toKey)
	toAddress := toAuth.From
	
	// initialize simulator with two ether funded accounts
	defaultNonce := uint64(0)
	defaultBalance := big.NewInt(93230000000000000)
	sim := backends.NewSimulatedBackend(core.GenesisAlloc{
		auth.From: {
			Balance:    defaultBalance,
			PrivateKey: crypto.FromECDSA(key),
			Nonce:      defaultNonce,
		},
		toAuth.From: {
			Balance:    defaultBalance,
			PrivateKey: crypto.FromECDSA(toKey),
			Nonce:      defaultNonce,
		},
	})
	
	// deploy a token contract on the simulated blockchain
	contractAddress, _, _, err := services.DeployOysterPearl(auth, sim)
	if err != nil {
		t.Fatalf("error deploying oyster pearl : %v", err)
	}
	// contract address
	if common.IsHexAddress(contractAddress.Hex()) {
		t.Logf("deployed contract address : %v", contractAddress.Hex())
	}
	
	// Commit to start a new state
	sim.Commit()
	
	// token instance by contract address
	token, err := services.NewOysterPearl(contractAddress, sim)
	
	fmt.Printf("transferring tokens")
	
	//transferTokens(sim) 20 prl
	value := onePrlWei.Mul(onePrlWei, big.NewInt(10))
	t.Logf("transfer %s tokens", value)
	
	auth.GasLimit = uint64(300000)     // in units
	
	gasPrice, err := sim.SuggestGasPrice(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("gasPrice : %v", gasPrice)
	
	auth.GasPrice = gasPrice
	// transfer tokens
	tx, err := token.Transfer(auth, toAddress, value)
	if err != nil {
		t.Fatalf("error transferring from oyster pearl : %v", err)
	}
	
	printTx(tx)
	
	sim.Commit()
	
	fmt.Println("transferring prl tokens completed")
	
	// print receipt status
	getReceipt(t, sim, tx)
	
	sim.Commit()
	
	// get balance for toAddress
	balance, err := token.BalanceOf(&bind.CallOpts{From:auth.From,Pending:true}, toAddress)
	if err != nil {
		t.Fatalf("balance check failed")
	}
	t.Logf("new balance for toAddress : %v", balance.Uint64())
	
	// bury the auth
	toAuth.GasPrice = gasPrice
	toAuth.GasLimit = uint64(300000)
	buryTx, err := token.Bury(toAuth)
	
	if err != nil {
		t.Fatalf("bury failed : %v", err)
	}
	
	printTx(buryTx)
	
	sim.Commit()
	
	// print receipt status
	getReceipt(t, sim, buryTx)
	
	t.Logf("bury completed")
}

// claim prl transfer treasure to the receiver address
// Contract Level Logic
//
// The claimed address must have already been buried
// The payout and fee addresses must be different
// The claimed address cannot pay itself
// It must be either the first time this address is being claimed
// claimed[msg.sender] = 1, occurs during bury address process, bury first
// Check if the buried address has enough:
// balances[msg.sender] >= claimAmount
// Check if claimed
func Test_simOysterPearlBuryClaim(t *testing.T) {
	
	// we are skipping here because when the previous test runs with success
	// and this test runs it fails, but if only one of them runs it works
	t.Skip(nil)
	
	// generate a new random account and a funded simulator
	key, _ := crypto.GenerateKey()
	auth := bind.NewKeyedTransactor(key)
	
	// Bury an address in the contract
	// 0.5 prl - lower bound
	// claimAmount
	// 40 prl  - upper bound
	// retentionMax
	
	// create toAddress to get the tokens
	toKey, _ := crypto.GenerateKey()
	toAuth := bind.NewKeyedTransactor(toKey)
	toAddress := toAuth.From
	
	feeKey, _ := crypto.GenerateKey()
	feeAuth := bind.NewKeyedTransactor(feeKey)
	feeAddress := feeAuth.From
	
	
	// initialize simulator with two ether funded accounts
	defaultNonce := uint64(0)
	defaultBalance := onePrlWei.Mul(onePrlWei, big.NewInt(11))
	sim := backends.NewSimulatedBackend(core.GenesisAlloc{
		auth.From: {
			Balance:    defaultBalance,
			PrivateKey: crypto.FromECDSA(key),
			Nonce:      defaultNonce,
		},
		toAuth.From: {
			Balance:    defaultBalance,
			PrivateKey: crypto.FromECDSA(toKey),
			Nonce:      defaultNonce,
		},
		feeAuth.From: {
			Balance:    defaultBalance,
			PrivateKey: crypto.FromECDSA(toKey),
			Nonce:      defaultNonce,
		},
	})
	
	// deploy a token contract on the simulated blockchain
	contractAddress, _, _, err := services.DeployOysterPearl(auth, sim)
	if err != nil {
		t.Fatalf("error deploying oyster pearl : %v", err)
	}
	// contract address
	if common.IsHexAddress(contractAddress.Hex()) {
		t.Logf("deployed contract address : %v", contractAddress.Hex())
	}
	
	// Commit to start a new state
	sim.Commit()
	
	// token instance by contract address
	token, err := services.NewOysterPearl(contractAddress, sim)
	
	fmt.Printf("transferring tokens")
	
	//transferTokens(sim) 20 prl
	value := onePrlWei.Mul(onePrlWei, big.NewInt(10))
	t.Logf("transfer %s tokens", value)
	
	auth.GasLimit = uint64(300000)     // in units
	
	gasPrice, err := sim.SuggestGasPrice(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("gasPrice : %v", gasPrice)
	
	auth.GasPrice = gasPrice
	// transfer tokens
	tx, err := token.Transfer(auth, toAddress, value)
	if err != nil {
		t.Fatalf("error transferring from oyster pearl : %v", err)
	}
	
	printTx(tx)
	
	sim.Commit()
	
	fmt.Println("transferring prl tokens completed")
	
	// print receipt status
	getReceipt(t, sim, tx)
	
	sim.Commit()
	
	// get balance for toAddress
	balance, err := token.BalanceOf(&bind.CallOpts{From:auth.From,Pending:true}, toAddress)
	if err != nil {
		t.Fatalf("balance check failed")
	}
	t.Logf("new balance for toAddress : %v", balance.Uint64())
	
	// bury the auth
	toAuth.GasPrice = gasPrice
	toAuth.GasLimit = uint64(300000)
	buryTx, err := token.Bury(toAuth)
	
	if err != nil {
		t.Fatalf("bury failed : %v", err)
	}
	
	printTx(buryTx)
	
	sim.Commit()
	
	// print receipt status
	getReceipt(t, sim, buryTx)
	
	t.Logf("bury completed")
	
	auth.GasPrice = gasPrice
	auth.GasLimit = uint64(300000)
	
	feeAuth.GasPrice = gasPrice
	feeAuth.GasLimit = uint64(300000)
	
	// claim
	claimTx, err := token.Claim(auth, toAddress, feeAddress)
	
	if err != nil {
		t.Fatalf("claim failed : %v", err)
	}
	t.Logf("claim ...")
	printTx(claimTx)
	
	sim.Commit()
	
	// print receipt status
	getReceipt(t, sim, claimTx)
	
	t.Logf("claim completed")
}

// claim prl insufficient funds
func Test_claimPRLInsufficientFunds(t *testing.T) {
	
	// claim PRL failure(insufficient funds) from sim OysterPearl token
	
	// generate a new random account and a funded simulator
	key, _ := crypto.GenerateKey()
	auth := bind.NewKeyedTransactor(key)
	
	// initialize simulator
	sim := createSimulator(auth, key)
	
	// deploy a token contract on the simulated blockchain
	contractAddress, _, token, err := services.DeployOysterPearl(auth, sim)
	if err != nil {
		t.Fatalf("error deploying oyster pearl : %v", err)
	}
	// contract address
	if common.IsHexAddress(contractAddress.Hex()) {
		t.Logf("deployed contract address : %v", contractAddress.Hex())
	}
	
	// Commit to start a new state
	sim.Commit()
	
	
	if err != nil {
		t.Fatalf("error deploying oyster pearl : %v", err)
	}
	t.Logf("contract address : %v", contractAddress.Hex())
	
	// claimant
	claimKey, _ := crypto.GenerateKey()
	claimAuth := bind.NewKeyedTransactor(claimKey)
	// to (swap to create failed state)
	payoutAddress := auth.From
	// from pays fee (swapped with above to create failed state) since
	// claimAuth has a zero balance
	fee := claimAuth.From
	
	// claim
	tx, err := token.Claim(claimAuth, payoutAddress, fee)
	
	if err != nil && tx == nil {
		t.Logf("insufficient funds to call claim oyster token contract method: %v", err)
	}
	
	sim.Commit()
	
}

// check if an address is in a buried state returns false
func Test_checkBuriedState(t *testing.T) {
	
	// check claim clock state from sim Oyster Pearl token
	
	// generate a new random account and a funded simulator
	key, _ := crypto.GenerateKey()
	auth := bind.NewKeyedTransactor(key)
	
	// create toAddress to get the tokens
	toKey, _ := crypto.GenerateKey()
	toAuth := bind.NewKeyedTransactor(toKey)
	toAddress := toAuth.From
	
	// initialize simulator
	sim := createSimulator(auth, key)
	buriedContext := context.Background()
	
	// deploy a token contract on the simulated blockchain
	contractAddress, _, _, err := services.DeployOysterPearl(auth, sim)
	if err != nil {
		t.Fatalf("error deploying oyster pearl : %v", err)
	}
	// contract address
	if common.IsHexAddress(contractAddress.Hex()) {
		t.Logf("deployed contract address : %v", contractAddress.Hex())
	}
	
	// token instance by contract address
	token, err := services.NewOysterPearl(contractAddress, sim)
	
	fmt.Printf("transferring tokens")
	
	//transferTokens(sim) 20 prl
	value := onePrlWei.Mul(onePrlWei, big.NewInt(10))
	t.Logf("transfer %s tokens", value)
	
	auth.GasLimit = uint64(300000)     // in units
	
	gasPrice, err := sim.SuggestGasPrice(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("gasPrice : %v", gasPrice)
	
	auth.GasPrice = gasPrice
	// transfer tokens
	tx, err := token.Transfer(auth, toAddress, value)
	if err != nil {
		t.Fatalf("error transferring from oyster pearl : %v", err)
	}
	
	printTx(tx)
	
	sim.Commit()
	
	fmt.Println("transferring prl tokens completed")
	
	// print receipt status
	getReceipt(t, sim, tx)
	
	t.Logf("gasPrice : %v", gasPrice)
	
	auth.GasLimit = uint64(300000)
	auth.GasPrice = gasPrice
	// check buried state post bury
	buried, err := token.Buried(&bind.CallOpts{
		From:auth.From,
		Pending:true,
		Context:buriedContext,
	}, auth.From)
	
	if err != nil {
		t.Fatalf("bury failed : %v", err)
	}
	
	// buried success
	t.Logf("%v successfully buried : %v", auth.From.Hex(), buried)
	
	sim.Commit()
}

// testing balance of the prl account for a given address
func Test_balanceOfFromOysterPearl(t *testing.T) {
	// check balance of PRLs from sim Oyster Pearl token
	
	// generate a new random account and a funded simulator
	key, _ := crypto.GenerateKey()
	auth := bind.NewKeyedTransactor(key)
	
	// initialize simulator
	sim := createSimulator(auth, key)
	
	// deploy a token contract on the simulated blockchain
	_, token, err := deployOysterPearl(auth, sim)
	
	// transfer tokens
	//transferTokens(sim) 10 prl
	value := onePrlWei.Mul(onePrlWei, big.NewInt(10))
	t.Logf("transfer %s tokens", value)
	
	tx, err := token.Transfer(auth, auth.From, value)
	if err != nil {
		t.Fatalf("error transferring from oyster pearl : %v", err)
	}
	
	printTx(tx)
	
	sim.Commit()
	
	fmt.Println("transferring prl tokens completed")
	
	// print receipt status
	getReceipt(t, sim, tx)
	
	// check balance from transfer
	bal, err := token.Balances(&bind.CallOpts{
		From:auth.From,
		Context:context.Background(),
		Pending:false,
	}, auth.From)
	
	if err != nil {
		t.Fatalf("balance check failed : %v", err)
	}
	
	// balance should equal the value set in the simulator initialization
	t.Logf("post transfer balance : %v", bal)

}

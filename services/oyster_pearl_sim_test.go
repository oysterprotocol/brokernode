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
)

//
// Oyster Pearl Contract and Services Tests
//

var onePrlWei = big.NewInt(1000000000000000000)

// Oysterby PRL Contract
var oysterContract = common.HexToAddress("0xB7baaB5caD2D2ebfE75A500c288A4c02B74bC12c")

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
	contractAddress, _, token, err := services.DeployOysterPearl(&bind.TransactOpts{
		From:     auth.From,
		Signer:   auth.Signer,
		Nonce:    auth.Nonce,
		GasLimit: auth.GasLimit,
		GasPrice: auth.GasPrice,
		Context:  auth.Context,
	}, sim)

	if err != nil {
		fmt.Errorf("failed to deploy oyster token contract: %v\n", err)
		return common.HexToAddress(""), nil, err
	}

	return contractAddress, token, err
}


// simulated blockchain to deploy oyster pearl
func Test_deployOysterPearl(t *testing.T) {

	// generate a new random account and a funded simulator
	key, _ := crypto.GenerateKey()
	auth := bind.NewKeyedTransactor(key)

	// initialize simulator
	sim := createSimulator(auth, key)

	// deploy a token contract on the simulated blockchain
	addr, _, err := deployOysterPearl(auth, sim)

	if err != nil {
		t.Fatalf("error deploying oyster pearl : %v", err)
	}
	
	if common.IsHexAddress(addr.Hex()) {
		t.Logf("deployed contract address : %v", addr.Hex())
	}

	// Commit to start a new state
	sim.Commit()
}

// simulated blockchain to test bury,
// claim with a contract with buried address
// Contract Level Logic
// An address must have at least 'claimAmount' to be buried
// > solidity require(balances[msg.sender] >= claimAmount);
// Prevent addresses with large balances from getting buried
// > solidity require(balances[msg.sender] <= retentionMax);
func Test_simOysterPearlBury(t *testing.T) {

	// generate a new random account and a funded simulator
	key, _ := crypto.GenerateKey()
	auth := bind.NewKeyedTransactor(key)

	// initialize simulator
	sim := createSimulator(auth, key)
	buryContext := context.Background()

	// deploy a token contract on the simulated blockchain
	addr, token, err := deployOysterPearl(auth, sim)

	if err != nil {
		t.Fatalf("error deploying oyster pearl : %v", err)
	}
	t.Logf("contract address : %v", addr.Hex())
	
	// retention should be less than
	// retentionMax = 40 * 10 ** 18 = 7200
	claimAmount := big.NewInt(0).SetUint64(toWei(7200))

	nonce, _ := sim.PendingNonceAt(buryContext, auth.From)

	// gas pricing from sim
	gasPrice, _ := sim.SuggestGasPrice(buryContext)
	t.Logf("gasPrice : %v", gasPrice)

	/**
	 * Transfer Tokens from Oyster Pearl to Eth Address
	 */
	transferTx, err := token.TransferFrom(&bind.TransactOpts{
		From:     auth.From,
		Signer:   auth.Signer,
		GasLimit: services.GasLimitPRLSend,
		Nonce:    big.NewInt(0).SetUint64(nonce),
		GasPrice: gasPrice,
		Context:  buryContext,
		Value:    claimAmount,
	}, auth.From, auth.From, claimAmount)

	if err != nil {
		t.Fatalf("transfer failed : %v", err)
	}

	printTx(transferTx)

	sim.Commit()

	receipt, err := sim.TransactionReceipt(buryContext, transferTx.Hash())
	if err != nil {
		t.Fatalf("failed to send transaction for bury : %v", err)
	}
	// 0 = failed, 1 = success
	if receipt.Status == 0 {
		t.Fatalf("tx receipt error : %v", receipt.Status)
	} else if receipt.Status == 1 {
		t.Logf("tx receipt success : %v", receipt.Status)
	}
	
	sim.Commit()
	
}

// test sending PRLs from OysterPearl Contract
// issue > transfer failed : replacement transaction underpriced
// solution > increase gasPrice by 10% minimum will work.
func Test_transferPRLFromOysterPearl(t *testing.T) {

	t.Skip(nil)

	// transfer PRL from sim OysterPearl token
	// generate a new random account and a funded simulator
	key, _ := crypto.GenerateKey()
	auth := bind.NewKeyedTransactor(key)
	
	// initialize simulator
	sim := createSimulator(auth, key)
	
	// deploy a token contract on the simulated blockchain
	addr, token, err := deployOysterPearl(auth, sim)
	
	if err != nil {
		t.Fatalf("error deploying oyster pearl : %v", err)
	}
	t.Logf("contract address : %v", addr.Hex())
	
	// transfer PRL receiver address
	receiverKey, _ := crypto.GenerateKey()
	receiverAuth := bind.NewKeyedTransactor(receiverKey)
	
	// transfer
	tx, err := token.Transfer(&bind.TransactOpts{
		Nonce:    receiverAuth.Nonce,
		From:     receiverAuth.From,
		GasLimit: receiverAuth.GasLimit,
		GasPrice: receiverAuth.GasPrice,
		Context:  receiverAuth.Context,
		Signer:   receiverAuth.Signer,
	}, receiverAuth.From, big.NewInt(0).SetUint64(toWei(75)))
	
	if err != nil {
		t.Fatalf("failed to call transfer oyster token contract method: %v", err)
	}
	
	t.Logf("transfer transaction submitted : %v", tx.Hash())
	
	sim.Commit()
}

//
// Oyster Pearl Tests
// all tests below will exercise the contract methods in the simulator and the private oysterby network.

// send prl from main wallet address to another address
func Test_sendPRL(t *testing.T) {

	t.Skip(nil)
	
	// send PRL via transaction from sim OysterPearl token
	
}

// claim prl transfer treasure to the receiver address
func Test_claimPRL(t *testing.T) {

	t.Skip(nil)
	
	// claim PRL from sim OysterPearl token
	
	// generate a new random account and a funded simulator
	key, _ := crypto.GenerateKey()
	auth := bind.NewKeyedTransactor(key)
	
	// initialize simulator
	sim := createSimulator(auth, key)
	
	// deploy a token contract on the simulated blockchain
	addr, token, err := deployOysterPearl(auth, sim)
	
	if err != nil {
		t.Fatalf("error deploying oyster pearl : %v", err)
	}
	t.Logf("contract address : %v", addr.Hex())
	
	// claimant
	claimKey, _ := crypto.GenerateKey()
	claimAuth := bind.NewKeyedTransactor(claimKey)
	// to
	payoutAddress := claimAuth.From
	// from pays fee
	fee := auth.From
	
	// claim
	tx, err := token.Claim(&bind.TransactOpts{
		Nonce:    claimAuth.Nonce,
		From:     claimAuth.From,
		GasLimit: claimAuth.GasLimit,
		GasPrice: claimAuth.GasPrice,
		Context:  claimAuth.Context,
		Signer:   claimAuth.Signer,
		Value: big.NewInt(0).SetUint64(toWei(7200)), // needs to have 7200 buried in balances to claim
	}, payoutAddress, fee)
	
	if err != nil {
		t.Fatalf("failed to call claim oyster token contract method: %v", err)
	}
	
	t.Logf("claim transaction submitted : %v", tx.Hash())
	
	sim.Commit()
}

// claim prl insufficient funds
func Test_claimPRLInsufficientFunds(t *testing.T) {

	t.Skip(nil)
	
	// claim PRL failure(insufficient funds) from sim OysterPearl token
	
	// generate a new random account and a funded simulator
	key, _ := crypto.GenerateKey()
	auth := bind.NewKeyedTransactor(key)
	
	// initialize simulator
	sim := createSimulator(auth, key)
	
	// deploy a token contract on the simulated blockchain
	addr, token, err := deployOysterPearl(auth, sim)
	
	if err != nil {
		t.Fatalf("error deploying oyster pearl : %v", err)
	}
	t.Logf("contract address : %v", addr.Hex())
	
	// claimant
	claimKey, _ := crypto.GenerateKey()
	claimAuth := bind.NewKeyedTransactor(claimKey)
	// to (swap to create failed state)
	payoutAddress := auth.From
	// from pays fee (swapped with above to create failed state) since
	// claimAuth has a zero balance
	fee := claimAuth.From
	
	// claim
	tx, err := token.Claim(&bind.TransactOpts{
		Nonce:    claimAuth.Nonce,
		From:     claimAuth.From,
		GasLimit: claimAuth.GasLimit,
		GasPrice: claimAuth.GasPrice,
		Context:  claimAuth.Context,
		Signer:   claimAuth.Signer,
	}, payoutAddress, fee)
	
	if err != nil {
		t.Logf("insufficient funds to call claim oyster token contract method: %v", err)
	}
	
	t.Logf("claim transaction submitted : %v", tx.Hash())
	
	sim.Commit()
	
}

// bury prl test comes after we set a claim
func Test_buryPRL(t *testing.T) {

	t.Skip(nil)

	// bury PRL from sim Oyster Pearl token
	
	// generate a new random account and a funded simulator
	key, _ := crypto.GenerateKey()
	auth := bind.NewKeyedTransactor(key)
	
	// initialize simulator
	sim := createSimulator(auth, key)
	
	// deploy a token contract on the simulated blockchain
	_, token, err := deployOysterPearl(auth, sim)
	
	// gas pricing from sim
	gasPrice, _ := sim.SuggestGasPrice(context.Background())
	t.Logf("gasPrice : %v", gasPrice)
	
	// bury
	buryTx, err := token.Bury(&bind.TransactOpts{
		Nonce:    auth.Nonce,
		From:     auth.From,
		GasLimit: auth.GasLimit,
		GasPrice: gasPrice,
		Context:  auth.Context,
		Signer:   auth.Signer,
		Value:    big.NewInt(0).SetUint64(toWei(75)),
	})
	
	if err != nil {
		t.Fatalf("bury failed : %v", err)
	}
	
	// send bury transaction
	printTx(buryTx)
	
	sim.Commit()
	
}

// check if an address is in a buried state
func Test_checkBuriedState(t *testing.T) {

	t.Skip(nil)
	
	// check claim clock state from sim Oyster Pearl token
	
	// generate a new random account and a funded simulator
	key, _ := crypto.GenerateKey()
	auth := bind.NewKeyedTransactor(key)
	
	// initialize simulator
	sim := createSimulator(auth, key)
	buriedContext := context.Background()
	
	// deploy a token contract on the simulated blockchain
	_, token, err := deployOysterPearl(auth, sim)
	
	// gas pricing from sim
	gasPrice, _ := sim.SuggestGasPrice(buriedContext)
	t.Logf("gasPrice : %v", gasPrice)
	
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

// check the claim clock value of an address
func Test_checkClaimClock(t *testing.T) {
	
	t.Skip(nil)
	
	// check buried state from sim Oyster Pearl token
	
}

// testing balance of the prl account for a given address
func Test_balanceOfFromOysterPearl(t *testing.T) {

	t.Skip(nil)
	
	// check balance of PRLs from sim Oyster Pearl token
	
	// generate a new random account and a funded simulator
	key, _ := crypto.GenerateKey()
	auth := bind.NewKeyedTransactor(key)
	
	// initialize simulator
	sim := createSimulator(auth, key)
	
	// deploy a token contract on the simulated blockchain
	_, token, err := deployOysterPearl(auth, sim)
	
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

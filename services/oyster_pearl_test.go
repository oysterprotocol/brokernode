package services_test

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/oysterprotocol/brokernode/services"
	"math/big"
	"testing"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"context"
	"github.com/ethereum/go-ethereum/crypto"
	"time"
)

//
// Oyster Pearl Contract and Services Tests
// All tests below will exercise the contract methods on private Oysterby network.
// Running Ethereum tests causes issues with posting transactions and confirmations.
// Due to this issue we have had to add time, or 2 seconds.
// Sleep(2 * time.Second) between certain transactions to ensure they will transact as expected.
// When the timeout was not added, the transactions fail randomly.
//


// testing token name access from OysterPearl Contract
// basic test which validates the existence of the contract on the network
func Test_oysterPearlTokenName(t *testing.T) {
	
	var backend, _ = ethclient.Dial(oysterbyNetwork)
	oysterPearl, err := services.NewOysterPearl(oysterContract, backend)
	if err != nil {
		t.Fatalf("unable to access contract instance at :%v", err)
	}
	name, err := oysterPearl.Name(nil)
	if err != nil {
		t.Fatalf("unable to access contract name : %v", err)
	}
	t.Logf("oyster pearl contract name from oysterby network :%v", name)
}

// testing prl balance
func Test_oysterPearlBalanceOf(t *testing.T) {
	
	// toWallet
	toWallet := getWallet(prl2File)
	
	// get balance from service
	balance := services.EthWrapper.CheckPRLBalance(toWallet.Address)
	
	// balance for prl2File account has prl
	if balance.Uint64() <= 0 {
		t.Fatalf("failed to get balance : %v", balance)
	} else {
		t.Logf("raw balance : %v", balance)
	}
	
	prlBalance := balance.Div(balance,onePrlWei)
	
	t.Logf("oyster pearl balance for address : %v", prlBalance)
}

// test sending PRLs from OysterPearl Contract
func Test_oysterPearlTransferPRL(t *testing.T) {
	
	var backend, _ = ethclient.Dial(oysterbyNetwork)
	
	// PRLs
	prlValue := big.NewInt(0).SetUint64(toWei(10))
	
	// oyster pearl bank
	fromWallet := getWallet(prlBank)
	auth := bind.NewKeyedTransactor(fromWallet.PrivateKey)
	// toWallet
	toWallet := getWallet(prl2File)
	
	// transfer tokens toWallet
	//value := prlValue // onePrlWei.Mul(onePrlWei, big.NewInt(1)) // from prl wei
	t.Logf("transferring %s tokens", prlValue)
	
	gasPrice, err := backend.SuggestGasPrice(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("gasPrice : %v", gasPrice)
	
	auth.GasPrice = gasPrice
	
	// transfer tokens via send prl
	sent, _, _ := services.EthWrapper.SendPRLFromOyster(services.OysterCallMsg{
		Amount: *prlValue,
		PrivateKey: *services.MainWalletPrivateKey,
		To: toWallet.Address,
	})
	
	if sent {
		t.Logf("sent the transaction successfully : %v", sent)
	} else {
		t.Fatalf("transaction failure")
	}
	
}


// oysterby blockchain to test bury success,
// issue > transfer failed : replacement transaction underpriced
// solution > increase gasPrice by 10% minimum will work.
//
// Contract Level Logic
// An address must have at least 'claimAmount' to be buried
// > solidity require(balances[msg.sender] >= claimAmount);
// Prevent addresses with large balances from getting buried
// > solidity require(balances[msg.sender] <= retentionMax);
func Test_oysterPearlBurySuccess(t *testing.T) {
	
	// this method runs ok, but with the suite fails, with tx failure
	// successful bury achieved
	//t.Skip(nil)
	
	time.Sleep(2 * time.Second)
	
	var backend, _ = ethclient.Dial(oysterbyNetwork)
	oysterPearl, err := services.NewOysterPearl(oysterContract, backend)
	if err != nil {
		t.Fatalf("unable to access contract instance at :%v", err)
	}
	
	// PRLs
	prlValue := big.NewInt(0).SetUint64(toWei(1))
	t.Logf("transferring %s tokens", prlValue)
	
	// oyster pearl bank
	fromWallet := getWallet(prlBank)
	auth := bind.NewKeyedTransactor(fromWallet.PrivateKey)
	
	//toWallet
	toKey, err := crypto.GenerateKey()
	toAuth := bind.NewKeyedTransactor(toKey)
	
	gasPrice, err := backend.SuggestGasPrice(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("gasPrice : %v", gasPrice)
	auth.GasPrice = gasPrice
	
	// transfer tokens via send prl
	sent, _, _ := services.EthWrapper.SendPRLFromOyster(services.OysterCallMsg{
		Amount: *prlValue,
		PrivateKey: *services.MainWalletPrivateKey,
		To: toAuth.From,
	})
	
	if sent {
		t.Logf("sent the transaction successfully : %v", sent)
		
		t.Log("starting bury process...")
		
		time.Sleep(3 * time.Second)
		
		// bury
		toAuth.GasPrice = gasPrice
		toAuth.GasLimit = services.GasLimitPRLBury
		buryTx, err := oysterPearl.Bury(toAuth)
		if err != nil {
			t.Fatalf("bury attempt failed : %v", err)
		}
		
		printTx(buryTx)
		
		buried, err := oysterPearl.Buried(&bind.CallOpts{From:toAuth.From,Pending:true}, toAuth.From)
		if err != nil {
			// failed bury attempt
			t.Fatalf("failed to get bury state : %v", err)
		}
		
		if buried {
			// successful bury attempt
			t.Logf("buried address successfully")
		} else {
			// failed bury attempt
			t.Fatal("failed to bury address.")
		}
		
	} else {
		t.Fatalf("transaction failure")
	}
	
	t.Log("bury process completed.")
	
}

// oysterby blockchain to test bury failure,
// No balance will fail this test as there is none in the random wallet key/pair we generate
//
// Contract Level Logic
// An address must have at least 'claimAmount' to be buried
// > solidity require(balances[msg.sender] >= claimAmount);
// Prevent addresses with large balances from getting buried
// > solidity require(balances[msg.sender] <= retentionMax);
func Test_oysterPearlBuryFailed(t *testing.T) {
	
	var backend, _ = ethclient.Dial(oysterbyNetwork)
	oysterPearl, err := services.NewOysterPearl(oysterContract, backend)
	if err != nil {
		t.Fatalf("unable to access contract instance at :%v", err)
	}
	
	// create random wallet key/pair to prove success for a bury call
	toKey, err := crypto.GenerateKey()
	toAuth := bind.NewKeyedTransactor(toKey)
	
	gasPrice, err := backend.SuggestGasPrice(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("gasPrice : %v", gasPrice)
	
	toAuth.GasPrice = gasPrice
	
	// setup auth
	toAuth.GasPrice = gasPrice
	toAuth.GasLimit = uint64(300000)
	
	t.Log("starting bury process...")
	
	// bury
	buryTx, err := oysterPearl.Bury(toAuth)
	if err != nil {
		t.Fatalf("bury attempt failed : %v", err)
	}
	
	printTx(buryTx)
	
	// check for a failed bury since there was no balance for the account
	buried, err := oysterPearl.Buried(&bind.CallOpts{From:toAuth.From,Pending:true}, toAuth.From)
	if err != nil {
		// failed bury attempt
		t.Fatalf("failed to get bury state : %v", err)
	}
	
	if buried {
		// successful bury attempt
		t.Fatalf("buried the PRLs successfully")
	} else {
		// failed bury attempt - expected result
		t.Logf("failed to bury PRLs.")
	}
	
	t.Log("bury process completed.")
}

// testing burying PRLs without enough funds to transact thereby returning a failure.
func Test_oysterPearlBuryInsufficientFunds(t *testing.T) {
	
	// retention should be less than
	// retentionMax = 40 * 10 ** 18 = 7200
	claimAmount := big.NewInt(7200)
	
	// load wallet key/pair
	wallet := getWallet(prl2File)
	
	// only configure to and amount
	buryMsg := services.OysterCallMsg{
		To:     wallet.Address,
		Amount: *claimAmount,
		PrivateKey: *wallet.PrivateKey,
	}
	
	// Bury PRL
	buried, _, _ := services.EthWrapper.BuryPrl(buryMsg)
	// Not Buried
	if !buried {
		// failed bury attempt
		t.Logf("Failed to bury PRLs due to insufficient funds")
	}
}


// claim prl transfer treasure to the receiver address
// claim prl works when there is available balance in treasure address
// therefore we need to fund the account prior
func Test_oysterPearlClaimPRL(t *testing.T) {
	
	time.Sleep(5 * time.Second)
	
	var backend, _ = ethclient.Dial(oysterbyNetwork)
	oysterPearl, _ := services.NewOysterPearl(oysterContract, backend)
	
	// PRLs
	prlValue := big.NewInt(0).SetUint64(toWei(2))
	
	// oyster pearl bank
	fromWallet := getWallet(prlBank)
	auth := bind.NewKeyedTransactor(fromWallet.PrivateKey)
	// toWallet
	toKey, err := crypto.GenerateKey()
	toAuth := bind.NewKeyedTransactor(toKey)
	
	// transfer tokens toWallet
	t.Logf("transferring %s tokens", prlValue)
	
	gasPrice, err := backend.SuggestGasPrice(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("gasPrice : %v", gasPrice)
	
	auth.GasPrice = gasPrice
	auth.GasLimit = services.GasLimitPRLSend
	
	// transfer tokens via send prl
	sent, _, _ := services.EthWrapper.SendPRLFromOyster(services.OysterCallMsg{
		Amount: *prlValue,
		PrivateKey: *fromWallet.PrivateKey,
		To: toAuth.From,
		GasPrice: *gasPrice,
	})
	
	if sent {
		t.Logf("sent the transaction successfully : %v", sent)
		
		time.Sleep(3 * time.Second)
		
		// bury
		toAuth.GasPrice = gasPrice
		toAuth.GasLimit = services.GasLimitPRLBury
		
		buryTx, err := oysterPearl.Bury(toAuth)
		if err != nil {
			t.Fatalf("bury attempt failed : %v", err)
		}
		
		printTx(buryTx)
		
		
	} else {
		t.Fatalf("transaction failure")
	}
	
	
	time.Sleep(5 * time.Second)
	
	// check buried
	toAuth.GasLimit = services.GasLimitPRLSend
	buried, err := oysterPearl.Buried(&bind.CallOpts{From:toAuth.From,Pending:true}, toAuth.From)
	if err != nil {
		// failed bury attempt
		t.Fatalf("failed to get bury state : %v", err)
	}
	
	if buried {
		// successful bury attempt
		t.Logf("buried address successfully")
	} else {
		// failed bury attempt
		t.Fatal("failed to bury address.")
	}
	
	time.Sleep(2 * time.Second)
	
	// setup receiver
	receiverKey, err := crypto.GenerateKey()
	receiverAuth := bind.NewKeyedTransactor(receiverKey)
	treasureAddress := receiverAuth.From
	treasurePrivateKey := receiverKey
	// Claim PRL
	claimed := services.EthWrapper.ClaimPRL(receiverAuth.From, treasureAddress, treasurePrivateKey)
	if !claimed {
		t.Fatal("Failed to claim PRLs")
	} else {
		t.Log("PRLs have been successfully claimed")
	}
	
}

// claim prl insufficient funds
func Test_oysterPearlClaimPRLInsufficientFunds(t *testing.T) {
	
	t.Skip(nil)
	
	// Receiver
	receiverAddress := prlAddress03
	
	// Treasure Wallet
	prlWallet := getWallet(prl2File)
	treasureAddress := prlWallet.Address
	treasurePrivateKey := prlWallet.PrivateKey
	
	// Claim PRL
	claimed := services.EthWrapper.ClaimPRL(receiverAddress, treasureAddress, treasurePrivateKey)
	if !claimed {
		t.Log("Failed to claim PRLs") // expected result
	} else {
		t.Log("PRLs have been successfully claimed") // not expected
	}
	
}

// check if an address is in a buried state
func Test_oysterPearlCheckBuriedState(t *testing.T) {
	
	t.Skip(nil)
	
	addr, _, _ := services.EthWrapper.GenerateEthAddr()
	
	buried, err := services.EthWrapper.CheckBuriedState(addr)
	
	if err != nil {
		t.Fatal("Failed to check the bury state of the given address.")
	} else {
		result := "false"
		if buried {
			result = "true"
		}
		t.Log("Successfully checked bury state: " + result)
	}
}

// check the claim clock value of an address
func Test_oysterPearlCheckClaimClock(t *testing.T) {
	
	t.Skip(nil)
	
	addr, _, _ := services.EthWrapper.GenerateEthAddr()
	
	claimClock, err := services.EthWrapper.CheckClaimClock(addr)
	
	if err != nil {
		t.Fatal("Failed to check the claim of the given address.")
	} else {
		t.Log("Successfully checked claim clock of the address: " + claimClock.String())
	}
}


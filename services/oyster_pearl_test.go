package services_test

import (
	"context"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/oysterprotocol/brokernode/services"
	"math/big"
	"testing"
)

//
// Oyster Pearl Contract and Services Tests
// All tests below will exercise the contract methods on private Oysterby network.
//


// testing token name access from OysterPearl Contract
// basic test which validates the existence of the contract on the network
func Test_oysterPearlTokenName(t *testing.T) {
	
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
	t.Logf("oyster pearl contract name from oysterby network :%v", name)
}

// testing balance of the prl account for a given address
func Test_oysterPearlBalanceOf(t *testing.T) {
	
	// working pulls the balance from Oyster PRL on test net prl balances
	bankBalance := services.EthWrapper.CheckPRLBalance(prlBankAddress)
	t.Logf("oyster pearl bank address balance :%v", bankBalance)
}


// testing token balanceOf from OysterPearl Contract account
// basic test which validates the balanceOf a PRL address
func Test_oysterPearlStakePRL(t *testing.T) {
	t.Skip(nil) // QA Method should not be run on regular testing runs
	// contract
	// test ethClient
	var backend, _ = ethclient.Dial(oysterbyNetwork)
	// instance of the oyster pearl contract
	pearlDistribute, err := services.NewPearlDistributeOysterby(prlDistribution, backend)
	
	// authentication
	walletAddress := services.MainWalletAddress
	
	t.Logf("using wallet key store from: %v\n", walletAddress.Hex())
	
	gasPrice, _ := services.EthWrapper.GetGasPrice()
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
	
	prlValue := big.NewInt(50)
	
	// transact
	opts := bind.TransactOpts{
		From:     auth.From,
		Signer:   auth.Signer,
		GasLimit: block.GasLimit(),
		Context:  context.Background(),
		Nonce:    auth.Nonce,
		GasPrice: gasPrice,
		Value:    prlValue,
	}
	t.Logf(opts.From.Hex())
	
	// stake
	tx, err := pearlDistribute.Stake(&opts, prlAddress02, prlValue)
	
	if err != nil {
		t.Fatalf("unable to access call distribute : %v", err)
	}
	t.Logf("oyster pearl distribute stake call :%v", tx)
}

// test sending PRLs from OysterPearl Contract
// issue > transfer failed : replacement transaction underpriced
// solution > increase gasPrice by 10% minimum will work.
func Test_oysterPearlTransferPRL(t *testing.T) {
	
	// PRLs
	prlValue := big.NewInt(0).SetUint64(toWei(15))
	
	// sendPRL
	sent, _, _ := services.EthWrapper.SendPRLFromOyster(services.OysterCallMsg{
		Amount: *prlValue,
		PrivateKey: *services.MainWalletPrivateKey,
		To: prlAddress02,
	})
	
	if sent {
		t.Logf("sent the transaction successfully : %v", sent)
	} else {
		t.Fatalf("transaction failure")
	}
	
}

// oysterby blockchain to test bury,
// claim with a contract with buried address
// Contract Level Logic
// An address must have at least 'claimAmount' to be buried
// > solidity require(balances[msg.sender] >= claimAmount);
// Prevent addresses with large balances from getting buried
// > solidity require(balances[msg.sender] <= retentionMax);
func Test_oysterPearlBury(t *testing.T) {
	
	//t.Skip(nil)
	
	// retention should be less than
	// retentionMax = 40 * 10 ** 18 = 7200
	claimAmount := big.NewInt(7200)
	
	// load wallet key/pair
	wallet := getWallet(prl3File)
	
	// only configure to and amount
	buryMsg := services.OysterCallMsg{
		To:     wallet.Address,
		Amount: *claimAmount,
		PrivateKey: *wallet.PrivateKey,
	}
	
	// Bury PRL
	buried, _, _ := services.EthWrapper.BuryPrl(buryMsg)
	if buried {
		// successful bury attempt
		t.Log("Buried the PRLs successfully")
	} else {
		// failed bury attempt
		t.Fatal("Faild to bury PRLs. Try Again?")
	}
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

// send prl from main wallet address to another address
func Test_oysterPearlSendPRL(t *testing.T) {
	
	// Wallet PRL bank address
	prlWallet := getWallet(prl1File)
	prlValue := big.NewInt(0).SetUint64(toWei(1))
	
	// Send PRL is a blocking call which will send the new transaction to the network
	// then wait for the confirmation to return true or false
	confirmed := services.EthWrapper.SendPRL(services.OysterCallMsg{
		To:         prlAddress02,
		From:       prlBankAddress,
		Amount:     *prlValue,
		PrivateKey: *prlWallet.PrivateKey,
	})
	if confirmed {
		// successful prl send
		t.Logf("Sent PRL to :%v", prlAddress02.Hex())
	} else {
		// failed prl send
		t.Fatalf("Failed to send PRL to:%v", prlAddress02.Hex())
	}
}

// claim prl transfer treasure to the receiver address
func Test_oysterPearlClaimPRL(t *testing.T) {
	
	t.Skip(nil)
	
	// Receiver
	receiverAddress := prlAddress02
	
	// Treasure Wallet
	prlWallet := getWallet(prl1File)
	treasureAddress := prlWallet.Address
	treasurePrivateKey := prlWallet.PrivateKey
	
	// Claim PRL
	claimed := services.EthWrapper.ClaimPRL(receiverAddress, treasureAddress, treasurePrivateKey)
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
	
	addr, _, _ := services.EthWrapper.GenerateEthAddr()
	
	claimClock, err := services.EthWrapper.CheckClaimClock(addr)
	
	if err != nil {
		t.Fatal("Failed to check the claim of the given address.")
	} else {
		t.Log("Successfully checked claim clock of the address: " + claimClock.String())
	}
}


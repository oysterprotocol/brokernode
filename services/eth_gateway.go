package services

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/oysterprotocol/brokernode/utils"

	"errors"
	"io/ioutil"
	"time"

	"github.com/ethereum/go-ethereum/params"
	"github.com/joho/godotenv"
)

type Eth struct {
	CalculateGasNeeded
	CheckIfWorthReclaimingGas
	ReclaimGas
	ClaimPRL
	GenerateEthAddr
	GenerateKeys
	GenerateEthAddrFromPrivateKey
	GeneratePublicKeyFromPrivateKey
	BuryPrl
	CheckBuriedState
	CheckClaimClock
	SendETH
	CreateSendPRLMessage
	SendPRL
	SendPRLFromOyster
	GetGasPrice
	WaitForTransfer
	CheckETHBalance
	CheckPRLBalance
	GetCurrentBlock
	GetConfirmationStatus
	WaitForConfirmation
	PendingConfirmation
	GetTransactionTable
	GetTransaction
	GetNonce
	GetTestWallet
	OysterCallMsg
}

// TxType Transaction Type Flag
type TxType int

const (
	PRLTransfer TxType = iota + 1
	EthTransfer
	PRLBury
	PRLClaim
)

const (
	// SecondsDelayForETHPolling Delay used during transaction confirmation(s)
	SecondsDelayForETHPolling = 10
)

// OysterPearlCallMsg
type OysterCallMsg struct {
	From       common.Address
	To         common.Address
	Amount     big.Int
	PrivateKey ecdsa.PrivateKey
	Gas        uint64
	GasPrice   big.Int
	TotalWei   big.Int
	Data       []byte
}

// OysterPearlTransactionType represents a transaction type used to determine which type of
// transaction was enacted (SEND_PRL, SEND_ETH, SEND_GAS)
type OysterPearlTransactionType struct {
	Type  string
	From  common.Address
	Value *big.Int
	Raw   types.Log // raw log object
}

// TransactionWithBlockNumber represents the data which confirms the transaction is completed
type TransactionWithBlockNumber struct {
	BlockNumber *big.Int
	Transaction *types.Transaction
	Confirmed   *bool
}

// CalculateGasNeeded Gas Calculation for Sending Transaction
type CalculateGasNeeded func(desiredGasLimit uint64) (*big.Int, error)

// CheckIfWorthReclaimingGas Accept an address and desired gas limit and decide if it is worth reclaiming the leftover ETH
type CheckIfWorthReclaimingGas func(address common.Address, desiredGasLimit uint64) (bool, *big.Int, error)

// ReclaimGas is called by various jobs methods to reclaim gas from a transaction address or treasure address
type ReclaimGas func(address common.Address, privateKey *ecdsa.PrivateKey, gasToReclaim *big.Int) bool

// CalculateGasToSend Gas Calculation for Sending Ether
type CalculateGasToSend func(desiredGasLimit uint64) (*big.Int, error)

// GenerateEthAddr Generate Valid Ethereum Network Address
type GenerateEthAddr func() (addr common.Address, privateKey string, err error)

// GenerateKeys Generate Private Keys W/O Address
type GenerateKeys func(int) (privateKeys []string, err error)

// GenerateEthAddrFromPrivateKey Generate Ethereum Address from Private Keys
type GenerateEthAddrFromPrivateKey func(privateKey string) (addr common.Address)

// GeneratePublicKeyFromPrivateKey Generate Ethereum Public Key from Private Key
type GeneratePublicKeyFromPrivateKey func(c elliptic.Curve, k *big.Int) *ecdsa.PrivateKey

// GetGasPrice Gas Price On Ethereum Network
type GetGasPrice func() (*big.Int, error)

// WaitForTransfer Wait For Transaction to Complete
type WaitForTransfer func(brokerAddr common.Address, transferType string) (*big.Int, error)

// CheckETHBalance Check Ethereum Balance
type CheckETHBalance func(common.Address) /*In Wei Unit*/ *big.Int

// CheckPRLBalance Check PRL Balance on Oyster Pearl
type CheckPRLBalance func(common.Address) /*In Wei Unit*/ *big.Int

// GetCurrentBlock Get Current(Latest) Block from Ethereum Network
type GetCurrentBlock func() (*types.Block, error)

// SendETH Send Ether To Valid Ethereum Network Address
type SendETH func(fromAddress common.Address, fromPrivateKey *ecdsa.PrivateKey, toAddr common.Address, amount *big.Int) (types.Transactions, string, int64, error)

// GetConfirmationStatus Get Transaction Confirmation Status
type GetConfirmationStatus func(txHash common.Hash) (*big.Int, error)

// WaitForConfirmation Wait For Transaction Confirmation to Complete
type WaitForConfirmation func(txHash common.Hash, pollingDelayInSeconds int) uint

// PendingConfirmation Check Transaction Pool For Pending Confirmation
type PendingConfirmation func(txHash common.Hash) bool

// GetNonce Return Nonce For Ethereum Account
type GetNonce func(ctx context.Context, address common.Address) (uint64, error)

// GetTransactionTable Return Transactions Table with Transactions Waiting To Confirm
type GetTransactionTable func() map[common.Hash]TransactionWithBlockNumber

// GetTransaction Return Transaction By Hash
type GetTransaction func(txHash common.Hash) TransactionWithBlockNumber

// GetTestWallet Utility to Access the Internal Test Wallet
type GetTestWallet func() *keystore.Key

// BuryPrl Bury Pearl With Oyster Pearl
type BuryPrl func(msg OysterCallMsg) (bool, string, int64)

// CheckBuriedState Check if an Address Is Buried
type CheckBuriedState func(addressToCheck common.Address) (bool, error)

// CheckClaimClock Call the smart contrat to check the claim clock status of an address
type CheckClaimClock func(addressToCheck common.Address) (*big.Int, error)

// CreateSendPRLMessage Utility to Send PRLs
type CreateSendPRLMessage func(from common.Address, privateKey *ecdsa.PrivateKey, to common.Address, prlAmount big.Int) (OysterCallMsg, error)

// SendPRL Send PRL from Account to Account
type SendPRL func(msg OysterCallMsg) bool

// SendPRLFromOyster Send PRL from Oyster Pearl
type SendPRLFromOyster func(msg OysterCallMsg) (bool, string, int64)

// ClaimPRL Claim Buried PRLs from Oyster Pearl
type ClaimPRL func(receiverAddress common.Address, treasureAddress common.Address, treasurePrivateKey *ecdsa.PrivateKey) bool

// Singleton client
var (
	EthUrl               string
	OysterPearlContract  string
	chainId              *big.Int
	MainWalletAddress    common.Address
	MainWalletPrivateKey *ecdsa.PrivateKey
	client               *ethclient.Client
	mtx                  sync.Mutex
	EthWrapper           Eth
)

// Limits selected based on actual transactions from etherscan
const (
	// PRL Gas Limit
	GasLimitPRLSend uint64 = 60000
	// ETH Gas Limit
	GasLimitETHSend uint64 = 21000
	// PRL Bury Gas Limit
	GasLimitPRLBury uint64 = 66000
	// PRL Claim Gas Limit
	GasLimitPRLClaim uint64 = 85000
)

func init() {

	RunOnMainETHNetwork()

	EthWrapper = Eth{
		CalculateGasNeeded:              calculateGasNeeded,
		CheckIfWorthReclaimingGas:       checkIfWorthReclaimingGas,
		ReclaimGas:                      reclaimGas,
		CreateSendPRLMessage:            createSendPRLMessage,
		SendPRL:                         sendPRL,
		SendPRLFromOyster:               sendPRLFromOyster,
		SendETH:                         sendETH,
		BuryPrl:                         buryPrl,
		CheckBuriedState:                checkBuriedState,
		CheckClaimClock:                 checkClaimClock,
		ClaimPRL:                        claimPRLs,
		GeneratePublicKeyFromPrivateKey: generatePublicKeyFromPrivateKey,
		GenerateEthAddr:                 generateEthAddr,
		GenerateKeys:                    generateKeys,
		GenerateEthAddrFromPrivateKey:   generateEthAddrFromPrivateKey,
		GetGasPrice:                     getGasPrice,
		WaitForTransfer:                 waitForTransfer,
		PendingConfirmation:             isPending,
		CheckETHBalance:                 checkETHBalance,
		CheckPRLBalance:                 checkPRLBalance,
		GetCurrentBlock:                 getCurrentBlock,
		GetConfirmationStatus:           getConfirmationStatus,
		WaitForConfirmation:             waitForConfirmation,
		GetNonce:                        getNonce,
		GetTransactionTable:             getTransactionTable,
		GetTransaction:                  getTransaction,
		GetTestWallet:                   getTestWallet,
	}
}

// Shared client provides access to the underlying Ethereum client
func sharedClient() (c *ethclient.Client, err error) {
	if client != nil {
		return client, nil
	}
	// check-lock-check pattern to avoid excessive locking.
	mtx.Lock()
	defer mtx.Unlock()

	if client != nil {
		return client, nil
	}

	c, err = ethclient.Dial(os.Getenv("ETH_NODE_URL"))
	if err != nil {
		fmt.Println("Failed to dial in to Ethereum node.")
		oyster_utils.LogIfError(err, nil)
		return
	}
	// Sets Singleton
	client = c
	return
}

// returns all the transactions in the transaction table
// table contains the sent transactions we are awaiting confirmation for
// post confirmation, we need to store the nonces for these transactions in the treasures table in the DB
// this post process will be responsible for clearing the table
func getTransactionTable() map[common.Hash]TransactionWithBlockNumber {
	return transactions
}

// transactions we search for transactions we have sent used to confirm
// size limitations? clear once item is found or flush to db
var transactions = make(map[common.Hash]TransactionWithBlockNumber)

// initialize subscription to access the latest blocks
func initializeSubscription() {
	client, _ := sharedClient()
	subscriptionChannel := make(chan types.Block)

	go func() {
		for {
			time.Sleep(3 * time.Second)
			subscribeToNewBlocks(client, subscriptionChannel)
		}
	}()

	for block := range subscriptionChannel {
		blockNumber := block.Number()
		// find transactions
		subTxs := block.Transactions()
		for subTx := range subTxs {
			transaction := subTxs[subTx]
			txHash := transaction.Hash()
			if txWithBlockNumber, ok := transactions[transaction.Hash()]; ok {
				fmt.Printf("transaction found : %v", txWithBlockNumber.BlockNumber)
				fmt.Printf("transaction found : %v", txWithBlockNumber.Transaction)
				updateTransaction(txHash, TransactionWithBlockNumber{Transaction: transaction, BlockNumber: blockNumber})
			}
		}
		// store the transaction with block number
		fmt.Printf("lastest block: %v", block.Number())
	}

}

func subscribeToNewBlocks(client *ethclient.Client, subscriptionChannel chan types.Block) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Subscribe to new blocks.
	//sub, err := client.EthSubscribe(ctx, subscriptionChannel, "newBlocks")
	// topicsHash := common.BytesToHash([]byte("newBlocks"))
	logCh := make(chan types.Log)

	sub, err := client.SubscribeFilterLogs(ctx, ethereum.FilterQuery{Topics: nil}, logCh)
	if err != nil {
		fmt.Println("subscribe error:", err)
		return
	}

	// Update the channel with the current block.
	lastBlock, err := getCurrentBlock()
	if err != nil {
		fmt.Println("can't get latest block:", err)
		return
	}
	subscriptionChannel <- *lastBlock

	// The subscription will deliver events to the channel. Wait for the
	// subscription to end for any reason, then loop around to re-establish
	// the connection.
	fmt.Println("connection lost: ", <-sub.Err())
}

// Generate an Ethereum address
func generateEthAddr() (addr common.Address, privateKey string, err error) {
	ethAccount, err := crypto.GenerateKey()
	if err != nil {
		oyster_utils.LogIfError(fmt.Errorf("Could not generate eth key: %v\n", err), nil)
		return addr, "", err
	}
	addr = crypto.PubkeyToAddress(ethAccount.PublicKey)
	privateKey = hex.EncodeToString(ethAccount.D.Bytes())
	if privateKey[0] == '0' || len(privateKey) != 64 || len(addr) != 20 {
		return generateEthAddr()
	}
	return addr, privateKey, err
}

// Generate an array of eth keys
func generateKeys(numKeys int) ([]string, error) {
	var keys []string
	var err error
	for i := 0; i < numKeys; i++ {
		key := ""
		if oyster_utils.BrokerMode == oyster_utils.TestModeDummyTreasure {
			key = os.Getenv("TEST_MODE_WALLET_KEY")
		} else {
			_, key, err = generateEthAddr()
			if err != nil {
				return keys, err
			}
		}
		keys = append(keys, key)
	}
	return keys, nil
}

// Generate an Ethereum address from a private key
func generateEthAddrFromPrivateKey(privateKey string) (addr common.Address) {
	privateKey = normalizePrivateKeyString(privateKey)
	privateKeyBigInt := hexutil.MustDecodeBig(privateKey)
	ethAccount := generatePublicKeyFromPrivateKey(crypto.S256(), privateKeyBigInt)
	addr = crypto.PubkeyToAddress(ethAccount.PublicKey)
	return addr
}

// GenerateKey generates a public and private key pair.
func generatePublicKeyFromPrivateKey(c elliptic.Curve, k *big.Int) *ecdsa.PrivateKey {
	privateKey := new(ecdsa.PrivateKey)
	privateKey.PublicKey.Curve = c
	privateKey.D = k
	privateKey.PublicKey.X, privateKey.PublicKey.Y = c.ScalarBaseMult(k.Bytes())
	return privateKey
}

func normalizePrivateKeyString(privateKey string) string {
	if !strings.HasPrefix(privateKey, "0x") && !strings.HasPrefix(privateKey, "0X") {
		return "0x" + privateKey
	} else {
		return privateKey
	}
}

// returns represents the 20 byte address of an ethereum account.
func StringToAddress(address string) common.Address {
	return common.HexToAddress(address)
}

// StringToPrivateKey Utility HexToECDSA parses a secp256k1 private key
func StringToPrivateKey(hexPrivateKey string) (*ecdsa.PrivateKey, error) {
	return crypto.HexToECDSA(hexPrivateKey)
}

// StringToTxHash Utility to parse transaction hash string to common.Hash
func StringToTxHash(txHash string) common.Hash {
	return common.HexToHash(txHash)
}

// SuggestGasPrice retrieves the currently suggested gas price to allow a timely
// execution for new transaction
func getGasPrice() (*big.Int, error) {
	// if QAing, un-comment out the line immediately below to hard-code a high gwei value for fast txs
	// return oyster_utils.ConvertGweiToWei(big.NewInt(70)), nil

	// connect ethereum client
	client, err := sharedClient()
	if err != nil {
		log.Fatal("Could not get gas price from network")
		oyster_utils.LogIfError(err, nil)
	}

	// there is no guarantee with estimate gas price
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal("Client could not get gas price from network")
		oyster_utils.LogIfError(err, nil)
	}
	return gasPrice, nil
}

// Get Estimated Gas Price for a Transaction
func getEstimatedGasPrice(to common.Address, from common.Address, gas uint64, gasPrice big.Int, value big.Int) (uint64, error) {
	// connect ethereum client
	client, err := sharedClient()
	if err != nil {
		log.Fatal("Could not get gas price from network")
		return 0, nil
	}

	// compose eth message
	msg := new(ethereum.CallMsg)
	msg.To = &to
	msg.From = from
	msg.Data = nil
	msg.Gas = gas
	msg.GasPrice = &gasPrice
	msg.Value = &value

	// estimated gas price
	estimatedGasPrice, err := client.EstimateGas(context.Background(), *msg)
	if err != nil {
		log.Fatal("Client could not get gas price estimate from network")
		oyster_utils.LogIfError(err, nil)
	}
	return estimatedGasPrice, nil
}

// Check balance from a valid Ethereum network address
func checkETHBalance(addr common.Address) *big.Int {
	// connect ethereum client
	client, err := sharedClient()
	if err != nil {
		log.Fatal("Could not initialize shared client")
	}

	balance, err := client.BalanceAt(context.Background(), addr, nil)
	if err != nil {
		oyster_utils.LogIfError(fmt.Errorf("Client could not retrieve balance: %v", err), nil)
		return big.NewInt(-1)
	}
	return balance
}

// Check balance from a valid PRL address
func checkPRLBalance(addr common.Address) *big.Int {
	// connect ethereum client
	client, err := sharedClient()
	if err != nil {
		log.Fatal("Could not initialize shared client")
		return big.NewInt(-1)
	}

	// instance of the oyster pearl contract
	OysterPearlAddress := common.HexToAddress(OysterPearlContract)
	oysterPearl, err := NewOysterPearl(OysterPearlAddress, client)
	if err != nil {
		fmt.Printf("unable to access contract instance at :%v", err)
		return big.NewInt(-1)
	}
	callOpts := bind.CallOpts{Pending: true, From: OysterPearlAddress}
	balance, err := oysterPearl.BalanceOf(&callOpts, addr)
	if err != nil {
		oyster_utils.LogIfError(fmt.Errorf("Client could not retrieve balance: %v", err), nil)
		return big.NewInt(-1)
	}
	return balance
}

// Get current block from blockchain
func getCurrentBlock() (*types.Block, error) {
	// connect ethereum client
	client, err := sharedClient()
	if err != nil {
		log.Fatal("Could not connect to Ethereum network", err)
		return nil, err
	}

	// latest block number is nil to get the latest block
	currentBlock, err := client.BlockByNumber(context.Background(), nil)
	if err != nil {
		oyster_utils.LogIfError(fmt.Errorf("Could not get last block: %v", err), nil)
		return nil, err
	}

	// latest block event
	fmt.Printf("latest block: %v\n", currentBlock.Number())
	return currentBlock, nil
}

// Utility to subscribe for notifications about the current blockchain head
func subscribeNewHead(tx common.Hash) (ethereum.Subscription, error) {
	client, _ := sharedClient()
	head := make(chan *types.Header)
	sub, err := client.SubscribeNewHead(context.Background(), head)
	if err != nil {
		return nil, err
	}
	return sub, nil
}

// determine if a transaction is pending
func isPending(txHash common.Hash) bool {
	client, _ := sharedClient()
	// get transaction
	_, isPending, err := client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		fmt.Printf("Could not get transaction by hash\n")
		return false
	}
	if isPending {
		fmt.Printf("transaction is pending\n")
	} else {
		fmt.Printf("transaction is not pending. Confirmed\n")
	}
	return isPending
}

// get transaction from transaction pool
func getTransaction(txHash common.Hash) TransactionWithBlockNumber {
	//check the transaction hash to find tx
	if tx, ok := transactions[txHash]; ok {
		return tx
	}
	// nil if not found
	return TransactionWithBlockNumber{}
}

// store initial transaction with nil block number
func storeTransaction(tx *types.Transaction) {
	// store in transactions table
	transactions[tx.Hash()] = TransactionWithBlockNumber{Transaction: tx, BlockNumber: nil}
}

// update transaction with block number from transactions table
func updateTransaction(txHash common.Hash, txWithBlockNumber TransactionWithBlockNumber) bool {
	//check the transaction hash to find tx
	if tx, ok := transactions[txHash]; ok {
		tx.BlockNumber = txWithBlockNumber.BlockNumber
		tx.Transaction = txWithBlockNumber.Transaction
		return true
	}
	return false
}

// Stub to write transactions to the the DB
func flushTransaction(txHash common.Hash) {
	// TODO should flush to database
	transactions[txHash] = TransactionWithBlockNumber{}
}

/**
 * Get number of confirmation status for a given transaction hash
 *
 * @return
 * (0) Failed is the status code of a transaction if execution failed.
 * (1) Successful is the status code of a transaction if execution succeeded.
 */
func getConfirmationStatus(txHash common.Hash) (*big.Int, error) {
	client, _ := sharedClient()

	block, _ := getCurrentBlock()
	blockNumber := block.Number()

	fmt.Printf("current block number : %v\n", blockNumber)

	//// get transaction
	tx, isPending, err := client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		fmt.Printf("isPending : %v", err)
		fmt.Println("Could not get transaction by hash")
		return big.NewInt(0), err
	}
	printTx(tx)
	if isPending {
		fmt.Println("transaction is pending...")
	} else {
		fmt.Println("transaction is no longer pending. confirmed!")
		// flush to db
		flushTransaction(txHash)
		return big.NewInt(1), nil
	}

	// get transaction receipt
	receipt, err := getTransactionReceipt(txHash)
	if err != nil {
		fmt.Printf("unable to get transaction receipt : %v\n", err)
		return big.NewInt(0), err
	}
	fmt.Printf("receipt status: %v\n", receipt.Status)
	return big.NewInt(int64(receipt.Status)), nil
}

// Wait For Transaction Confirmation
func waitForConfirmation(txHash common.Hash, pollingDelayInSeconds int) uint {
	var status uint
	pollingDelayDuration := time.Duration(pollingDelayInSeconds) * time.Second
	for {
		// access confirmation status until we get the correct response
		// need to be aware of possible *known transaction* error
		txStatus, err := getConfirmationStatus(txHash)
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			status = 0
		} else {
			if txStatus.Uint64() == 0 {
				oyster_utils.LogIfError(err, nil)
				status = 0
				break
			} else if txStatus.Uint64() == 1 {
				fmt.Println("confirmation completed")
				status = 1
				break
			}
		}
		time.Sleep(pollingDelayDuration)
	}
	return status
}

// Utility to access the nonce for a given account
func getNonce(ctx context.Context, address common.Address) (uint64, error) {
	client, err := sharedClient()
	if err != nil {
		return 0, err
	}
	return client.NonceAt(ctx, address, nil)
}

// Utility to get the transaction count in the pending tx pool
func getPendingTransactions() (uint, error) {
	client, _ := sharedClient()
	return client.PendingTransactionCount(context.Background())
}

// Utility to get check if the transaction is in the block
func getTransactionInBlock(ctx context.Context, blockHash common.Hash, index uint) (*types.Transaction, error) {
	client, _ := sharedClient()
	return client.TransactionInBlock(ctx, blockHash, index)
}

// Utility to get the transaction receipt
func getTransactionReceipt(txHash common.Hash) (*types.Receipt, error) {
	client, _ := sharedClient()
	ctx, cancel := createContext()
	defer cancel()
	return client.TransactionReceipt(ctx, txHash)
}

// WaitForTransfer is blocking call that will observe on brokerAddr on transfer of PRL or ETH
func waitForTransfer(brokerAddr common.Address, transferType string) (*big.Int, error) {
	balance := checkPRLBalance(brokerAddr)
	if balance.Int64() > 0 {
		// Has balance already, don't need to wait for it.
		return balance, nil
	}

	client, err := sharedClient()
	if err != nil {
		return big.NewInt(0), err
	}

	query := ethereum.FilterQuery{
		FromBlock: nil, // beginning of the queried range, nil means genesis block
		ToBlock:   nil, // end of the range, nil means latest block
		Addresses: []common.Address{brokerAddr},
		Topics:    nil, // matches any topic list
	}

	// initialize the context with an hour deadline before the channel closes
	deadline := time.Now().Add(time.Hour)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	// setup logs channel
	//var wg sync.WaitGroup

	logs := make(chan types.Log)
	sub, subErr := client.SubscribeFilterLogs(ctx, query, logs)
	if subErr != nil {
		oyster_utils.LogIfError(fmt.Errorf("error subscribing to logs: %v", subErr), nil)
		return big.NewInt(-1), subErr
	}
	defer sub.Unsubscribe()
	for {
		select {
		case err := <-sub.Err():
			log.Fatal(err)
			oyster_utils.LogIfError(err, nil)
			return big.NewInt(0), err
			//case <-time.After(1 * time.Minute):
			//	log.Print("Timeout to wait for brokerAddr\n")
			//	// Wait for 1 hr to receive payment before timeout
			//	return big.NewInt(0), errors.New("subscription timed out")
		case log := <-logs:
			fmt.Print(log)
			fmt.Printf("Log Data:%v", string(log.Data))
			fmt.Println("Confirmed Address:", log.Address.Hex())

			// OysterPearlTransactionType will hold what the action was, SEND_GAS,SEND_PRL
			// ensure confirmation type from "sendGas" or "sendPRL"
			// recordTransaction(log.Address, "")

			if transferType == "eth" {
				return checkETHBalance(brokerAddr), nil
			} else if transferType == "prl" {
				return checkPRLBalance(brokerAddr), nil
			} else if transferType == "bury" {
				// TODO do something for this, or remove it
				// If removing it, how will we monitor for when bury() has been invoked?
			}

		}
	}

}

func checkIfWorthReclaimingGas(address common.Address, desiredGasLimit uint64) (bool, *big.Int, error) {
	ethBalance := checkETHBalance(address)

	if ethBalance.Int64() > 0 {
		gasNeededToReclaimETH, err := calculateGasNeeded(desiredGasLimit)
		if err != nil {
			return false, big.NewInt(0), err
		}
		if gasNeededToReclaimETH.Int64() >= ethBalance.Int64() {
			return false, big.NewInt(0), nil
		}
		return true, new(big.Int).Sub(ethBalance, gasNeededToReclaimETH), nil
	} else {
		var balErr error
		if ethBalance.Int64() == -1 {
			balErr = errors.New("error checking if worth " +
				"reclaiming gas because could not get ethBalance")
		}
		return false, big.NewInt(0), balErr
	}
}

/* reclaimGas attempts to take any leftover ETH from the address */
func reclaimGas(address common.Address, privateKey *ecdsa.PrivateKey, gasToReclaim *big.Int) bool {

	_, _, _, err := EthWrapper.SendETH(
		address,
		privateKey,
		MainWalletAddress,
		gasToReclaim)

	if err != nil {
		fmt.Println("Could not reclaim leftover ETH from " + address.Hex())
		return false
	} else {
		fmt.Println("Reclaiming leftover ETH from " + address.Hex())
		return true
	}
}

func calculateGasNeeded(desiredGasLimit uint64) (*big.Int, error) {
	gasPrice, err := getGasPrice()
	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return big.NewInt(0), err
	}
	gasToSend := new(big.Int).Mul(gasPrice, big.NewInt(int64(desiredGasLimit)))
	return gasToSend, nil
}

// Transfer funds from main wallet
func sendETH(fromAddress common.Address, fromPrivKey *ecdsa.PrivateKey, toAddr common.Address, amount *big.Int) (types.Transactions, string, int64, error) {

	client, err := sharedClient()
	if err != nil {
		return types.Transactions{}, "", -1, err
	}

	// initialize the context
	ctx, cancel := createContext()
	defer cancel()

	// generate nonce
	nonce, _ := client.PendingNonceAt(ctx, fromAddress)

	// default gasLimit on oysterby 4294967295
	gasPrice, _ := getGasPrice()

	// estimation
	estimate, failedEstimate := getEstimatedGasPrice(toAddr, fromAddress, GasLimitETHSend, *gasPrice, *amount)
	if failedEstimate != nil {
		fmt.Printf("failed to get estimated network price : %v\n", failedEstimate)
		return types.Transactions{}, "", -1, failedEstimate
	}
	estimatedGas := new(big.Int).SetUint64(estimate)
	fmt.Printf("estimatedGas : %v\n", estimatedGas)

	balance := checkETHBalance(fromAddress)
	fmt.Printf("balance : %v\n", balance)

	// amount is greater than balance, return error
	if amount.Uint64() > balance.Uint64() {
		return types.Transactions{}, "", -1, errors.New("balance too low to proceed")
	}

	// create new transaction
	tx := types.NewTransaction(nonce, toAddr, amount, GasLimitETHSend, gasPrice, nil)

	// signer
	signer := types.NewEIP155Signer(chainId)

	// sign transaction
	signedTx, err := types.SignTx(tx, signer, fromPrivKey)
	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return types.Transactions{}, "", -1, err
	}

	// send transaction
	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		oyster_utils.LogIfError(fmt.Errorf("error sending transaction : %v", err), nil)
		return types.Transactions{}, "", -1, err
	}

	// pull signed transaction(s)
	signedTxs := types.Transactions{signedTx}
	for tx := range signedTxs {
		transaction := signedTxs[tx]
		// store in broker transaction pool
		storeTransaction(transaction)
		printTx(transaction)
	}

	// return signed transactions
	return signedTxs, signedTx.Hash().Hex(), int64(signedTx.Nonce()), nil
}

// Bury PRLs
func buryPrl(msg OysterCallMsg) (bool, string, int64) {

	// shared client
	client, _ := sharedClient()

	// Create an authorized transactor
	auth := bind.NewKeyedTransactor(&msg.PrivateKey)
	if auth == nil {
		fmt.Printf("unable to create a new transactor")
	}
	fmt.Printf("authorized transactor : %v\n", auth.From.Hex())

	// transact
	oysterPearl, err := NewOysterPearl(common.HexToAddress(OysterPearlContract), client)
	if err != nil {
		fmt.Print("Unable to instantiate OysterPearl")
	}

	ethBalance := checkETHBalance(auth.From)

	// determine the gas price we are willing to pay by the gas price we settled
	// upon when we sent the eth earlier in the sequence
	gasPrice := new(big.Int).Quo(ethBalance, big.NewInt(int64(GasLimitPRLBury)))

	// call bury on oyster pearl
	tx, err := oysterPearl.Bury(&bind.TransactOpts{
		From:     auth.From,
		Signer:   auth.Signer,
		GasLimit: GasLimitPRLBury,
		Nonce:    auth.Nonce,
		GasPrice: gasPrice,
		Context:  auth.Context,
	})

	printTx(tx)

	return tx != nil, tx.Hash().Hex(), int64(tx.Nonce())
}

// Check if an address is buried
func checkBuriedState(address common.Address) (bool, error) {

	// shared client
	client, _ := sharedClient()

	// transact
	OysterPearlAddress := common.HexToAddress(OysterPearlContract)
	oysterPearl, err := NewOysterPearl(OysterPearlAddress, client)
	if err != nil {
		fmt.Printf("unable to access contract instance at :%v", err)
		return false, err
	}

	// create callOpts
	callOpts := bind.CallOpts{
		From:    OysterPearlAddress,
		Pending: true,
	}

	// check buried state
	buried, err := oysterPearl.Buried(&callOpts, address)
	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return false, err
	}

	return buried, err
}

// Check the claim clock value of an address
func checkClaimClock(address common.Address) (*big.Int, error) {

	// shared client
	client, _ := sharedClient()

	// transact
	OysterPearlAddress := common.HexToAddress(OysterPearlContract)
	oysterPearl, err := NewOysterPearl(OysterPearlAddress, client)
	if err != nil {
		fmt.Printf("unable to access contract instance at :%v", err)
		return big.NewInt(-1), err
	}

	// create callOpts
	callOpts := bind.CallOpts{
		From:    OysterPearlAddress,
		Pending: true,
	}

	// check buried state
	claimClock, err := oysterPearl.Claimed(&callOpts, address)
	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return big.NewInt(-1), err
	}

	return claimClock, err
}

// Claim PRL allows the receiver to unlock the treasure address and private key to enable the transfer
func claimPRLs(receiverAddress common.Address, treasureAddress common.Address, treasurePrivateKey *ecdsa.PrivateKey) bool {

	// shared client
	client, _ := sharedClient()
	treasureBalance := checkPRLBalance(treasureAddress)
	fmt.Printf("treasure balance : %v\n", treasureBalance)

	if treasureBalance.Uint64() <= 0 {
		err := errors.New("treasure balance insufficient")
		oyster_utils.LogIfError(err, nil)
		return false
	}

	buried, err := checkBuriedState(treasureAddress)
	if err != nil {
		fmt.Println("cannot claim PRLs in claimPRL due to: " + err.Error())
		return false
	}
	if !buried {
		err = errors.New("treasure address is not in a buried state")
		oyster_utils.LogIfError(err, nil)
		return false
	}

	// Create an authorized transactor
	auth := bind.NewKeyedTransactor(treasurePrivateKey)
	if auth == nil {
		err := errors.New("unable to create a new transactor")
		oyster_utils.LogIfError(err, nil)
	}

	fmt.Printf("authorized transactor : %v\n", auth.From.Hex())

	// transact with oyster pearl instance
	oysterPearl, err := NewOysterPearl(common.HexToAddress(OysterPearlContract), client)
	if err != nil {
		err := errors.New("unable to instantiate OysterPearl")
		oyster_utils.LogIfError(err, nil)
		return false
	}

	ethBalance := checkETHBalance(auth.From)

	if ethBalance.Int64() == -1 {
		return false
	}

	// determine the gas price we are willing to pay by the gas price we settled
	// upon when we sent the eth earlier in the sequence
	gasPrice := new(big.Int).Quo(ethBalance, big.NewInt(int64(GasLimitPRLClaim)))

	// setup transaction options
	claimOpts := bind.TransactOpts{
		From:     auth.From,
		Signer:   auth.Signer,
		GasLimit: GasLimitPRLClaim,
		GasPrice: gasPrice,
		Nonce:    auth.Nonce,
		Context:  auth.Context,
	}
	// call claim, receiver is payout, fee coming from the treasure address and private key
	tx, err := oysterPearl.Claim(&claimOpts, receiverAddress, MainWalletAddress)

	if err != nil {
		fmt.Printf("unable to call claim with transactor : %v", err)
		return false
	}

	// store in broker transaction pool
	storeTransaction(tx)
	printTx(tx)

	//var status = false
	//
	//// confirm status of transaction
	//txStatus := waitForConfirmation(tx.Hash(), SecondsDelayForETHPolling)
	//
	//if txStatus == 0 {
	//	fmt.Printf("transaction failure")
	//	status = false
	//} else if txStatus == 1 {
	//	fmt.Printf("confirmation completed")
	//	flushTransaction(tx.Hash())
	//	status = true
	//}

	return true
}

func createSendPRLMessage(from common.Address, privateKey *ecdsa.PrivateKey, to common.Address, prlAmount big.Int) (OysterCallMsg, error) {

	callMsg := OysterCallMsg{
		From:       from,
		To:         to,
		Amount:     prlAmount,
		Gas:        GasLimitPRLSend,
		PrivateKey: *privateKey,
		TotalWei:   *big.NewInt(0).SetUint64(uint64(prlAmount.Int64())),
	}

	return callMsg, nil
}

/**
sendPrl
When a user uploads a file, we create an upload session on the broker.

For each "upload session", we generate a new wallet (we do this so we can associate a session to PRLs sent).
The broker responds to the uploader with an invoice to send X PRLs to Y eth address
The broker then listens for a transfer event so it knows when payment has happened,

Once the payment is received, the brokers will split up the PRLs and begin work to attach the file to the IOTA tangle
so to answer your question, there won't be a main PRL wallet, the address is different for each session
however there will be a "main" ETH wallet, which is used to pay gas fees
*/
func sendPRL(msg OysterCallMsg) bool {

	client, _ := sharedClient()
	// initialize the context
	ctx, cancel := createContext()
	defer cancel()

	// generate nonce
	nonce, _ := client.PendingNonceAt(ctx, msg.From)

	balance := checkPRLBalance(msg.From)
	fmt.Printf("balance : %v\n", balance)

	// amount is greater than balance, return error
	if msg.Amount.Uint64() > balance.Uint64() {
		fmt.Println("balance too low to proceed")
		return false
	}
	fmt.Printf("sending prl to : %v\n", msg.To.Hex())

	// default gasLimit on oysterby 4294967295
	gasPrice, _ := getGasPrice()

	// create new transaction
	tx := types.NewTransaction(nonce, msg.To, &msg.Amount, msg.Gas, gasPrice, nil)

	// signer
	signer := types.NewEIP155Signer(chainId)

	// sign transaction
	signedTx, err := types.SignTx(tx, signer, &msg.PrivateKey)
	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return false
	}

	// send transaction
	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		// given we have a "known transaction" error we need to respond
		oyster_utils.LogIfError(err, nil)
		return false
	}

	// pull signed transaction(s)
	signedTxs := types.Transactions{signedTx}
	var confirmTx types.Transaction
	for tx := range signedTxs {
		transaction := signedTxs[tx]

		// store in broker transaction pool
		storeTransaction(transaction)
		printTx(transaction)

		confirmTx = *transaction
	}

	// confirm status of transaction
	txStatus := waitForConfirmation(confirmTx.Hash(), SecondsDelayForETHPolling)

	return processTxStatus(txStatus)
}

// send prl from oyster via contract transfer method
func sendPRLFromOyster(msg OysterCallMsg) (bool, string, int64) {

	client, _ := sharedClient()
	oysterPearl, err := NewOysterPearl(common.HexToAddress(OysterPearlContract), client)

	if err != nil {
		log.Printf("unable to access contract instance at : %v\n", err)
	}

	// initialize transactor // may need to move this to a session based transactor
	auth := bind.NewKeyedTransactor(&msg.PrivateKey)
	if err != nil {
		log.Printf("unable to create a new transactor : %v", err)
	}

	log.Printf("authorized transactor : %v\n", auth.From.Hex())

	// use this when in production:
	gasPrice, err := getGasPrice()

	opts := bind.TransactOpts{
		From:     auth.From,
		Signer:   auth.Signer,
		GasLimit: GasLimitPRLSend,
		GasPrice: gasPrice,
		Nonce:    auth.Nonce,
		Context:  auth.Context,
	}

	tx, err := oysterPearl.Transfer(&opts, msg.To, &msg.Amount)
	if err != nil {
		log.Printf("transfer failed : %v", err)
		return false, "", int64(-1)
	}

	log.Printf("transfer pending: 0x%x\n", tx.Hash())

	printTx(tx)

	return true, tx.Hash().Hex(), int64(tx.Nonce())
}

// utility to process the transaction status
func processTxStatus(txStatus uint) bool {
	status := false
	if txStatus == 0 {
		fmt.Println("transaction failure")
		status = false
	} else if txStatus == 1 {
		fmt.Println("confirmation completed")
		status = true
	}
	return status
}

// utility to access the test wallet keystore
func getTestWallet() *keystore.Key {

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

// utility context helper to include the deadline initialization
func createContext() (ctx context.Context, cancel context.CancelFunc) {
	deadline := time.Now().Add(5000 * time.Millisecond)
	return context.WithDeadline(context.Background(), deadline)
}

var MAIN = "mainnet"
var TEST = "testnet"
var DEBUG = false

func configureGateway(network string) {
	// Load ENV variables
	//err := godotenv.Load()
	err := godotenv.Load()
	if err != nil {
		godotenv.Load()
		log.Printf(".env error: %v", err)
		oyster_utils.LogIfError(err, nil)
	}
	// Resolve network type
	switch network {
	case MAIN:
		// ethereum main net chain id
		chainId = params.MainnetChainConfig.ChainID
		break
	case TEST:
		// oysterby test net chain id
		chainId = big.NewInt(0)
		chainId.SetString(os.Getenv("CHAIN_ID"), 10)
		break
	}
	// ethereum network node
	EthUrl = os.Getenv("ETH_NODE_URL")
	// smart contract
	OysterPearlContract = os.Getenv("OYSTER_PEARL")
	// wallet address configuration
	MainWalletAddress = common.HexToAddress(os.Getenv("MAIN_WALLET_ADDRESS"))
	// wallet private key configuration
	MainWalletPrivateKey, err = crypto.HexToECDSA(os.Getenv("MAIN_WALLET_KEY"))

	if DEBUG {
		// Print Configuration
		printConfig()
	}

	if err != nil && os.Getenv("MODE") == "PROD_MODE" {
		// Panic and force releaser to figure out.
		panic(fmt.Errorf("unable to load key: %v", err))
	}
}

func RunOnMainETHNetwork() {
	configureGateway(MAIN)
}

func RunOnTestNet() {
	configureGateway(TEST)
}

// utility to print
func printTx(tx *types.Transaction) {
	fmt.Printf("tx to     : %v\n", tx.To().Hash().String())
	fmt.Printf("tx hash   : %v\n", tx.Hash().String())
	fmt.Printf("tx amount : %v\n", tx.Value())
	fmt.Printf("tx cost   : %v\n", tx.Cost())
}

// utility to print gateway configuration
func printConfig() {
	fmt.Println("Using main wallet address: ")
	fmt.Println(MainWalletAddress.Hex())
	fmt.Println("Using main wallet private key: ")
	privateKeyString := hex.EncodeToString(crypto.FromECDSA(MainWalletPrivateKey))
	fmt.Println(privateKeyString)
	fmt.Println("Using eth node url: ")
	fmt.Println(EthUrl)
	fmt.Println("Using oyster pearl contract: ")
	fmt.Println(OysterPearlContract)
}

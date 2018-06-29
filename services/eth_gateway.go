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
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/utils"

	"errors"
	"github.com/ethereum/go-ethereum/params"
	"github.com/joho/godotenv"
	"io/ioutil"
	"time"
)

type Eth struct {
	CalculateGasToSend
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

type TxType int

const (
	PRLTransfer TxType = iota + 1
	EthTransfer
	PRLBury
	PRLClaim
)

const (
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

type CalculateGasToSend func(desiredGasLimit uint64) (*big.Int, error)
type GenerateEthAddr func() (addr common.Address, privateKey string, err error)
type GenerateKeys func(int) (privateKeys []string, err error)
type GenerateEthAddrFromPrivateKey func(privateKey string) (addr common.Address)
type GeneratePublicKeyFromPrivateKey func(c elliptic.Curve, k *big.Int) *ecdsa.PrivateKey
type GetGasPrice func() (*big.Int, error)
type WaitForTransfer func(brokerAddr common.Address, transferType string) (*big.Int, error)
type CheckETHBalance func(common.Address) /*In Wei Unit*/ *big.Int
type CheckPRLBalance func(common.Address) /*In Wei Unit*/ *big.Int
type GetCurrentBlock func() (*types.Block, error)
type SendETH func(fromAddr common.Address, fromPrivKey *ecdsa.PrivateKey, toAddr common.Address, amount *big.Int) (types.Transactions, string, int64, error)
type GetConfirmationStatus func(txHash common.Hash) (*big.Int, error)
type WaitForConfirmation func(txHash common.Hash, pollingDelayInSeconds int) uint
type PendingConfirmation func(txHash common.Hash) bool
type GetNonce func(ctx context.Context, address common.Address) (uint64, error)
type GetTransactionTable func() map[common.Hash]TransactionWithBlockNumber
type GetTransaction func(txHash common.Hash) TransactionWithBlockNumber
type GetTestWallet func() *keystore.Key

type BuryPrl func(msg OysterCallMsg) (bool, string, int64)
type CheckBuriedState func(addressToCheck common.Address) (bool, error)
type CheckClaimClock func(addressToCheck common.Address) (*big.Int, error)
type CreateSendPRLMessage func(from common.Address, privateKey *ecdsa.PrivateKey, to common.Address, prlAmount big.Int) (OysterCallMsg, error)
type SendPRL func(msg OysterCallMsg) bool
type SendPRLFromOyster func(msg OysterCallMsg) (bool, string, int64)
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

const (
	GasLimitPRLSend  uint64 = 60000
	GasLimitETHSend  uint64 = 21000
	GasLimitPRLBury  uint64 = 66000
	GasLimitPRLClaim uint64 = 85000
	// from actual transactions on etherscan, this tends to be 53898
)

func init() {

	RunOnMainETHNetwork()

	EthWrapper = Eth{
		CalculateGasToSend:              calculateGasToSend,
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
	if privateKey[0] == '0' {
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

func StringToPrivateKey(hexPrivateKey string) (*ecdsa.PrivateKey, error) {
	return crypto.HexToECDSA(hexPrivateKey)
}

func StringToTxHash(txHash string) common.Hash {
	return common.HexToHash(txHash)
}

// SuggestGasPrice retrieves the currently suggested gas price to allow a timely
// execution for new transaction
func getGasPrice() (*big.Int, error) {
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

	// if QAing, comment out all lines above and un-comment out the lines below
	// for faster transactions

	//return oyster_utils.ConvertGweiToWei(big.NewInt(3)), nil
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
	}

	// instance of the oyster pearl contract
	OysterPearlAddress := common.HexToAddress(OysterPearlContract)
	oysterPearl, err := NewOysterPearl(OysterPearlAddress, client)
	if err != nil {
		fmt.Printf("unable to access contract instance at :%v", err)
	}
	callOpts := bind.CallOpts{Pending: true, From: OysterPearlAddress}
	balance, err := oysterPearl.BalanceOf(&callOpts, addr)
	if err != nil {
		oyster_utils.LogIfError(fmt.Errorf("Client could not retrieve balance: %v", err), nil)
		return big.NewInt(0)
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
		fmt.Errorf("unable to get transaction receipt : %v\n", err)
		return big.NewInt(0), err
	}
	fmt.Printf("receipt status: %v\n", receipt.Status)
	return big.NewInt(int64(receipt.Status)), nil
}

// Wait For Confirmation
// TODO add a channel output to return result via subscription
//func waitForConfirmation(txHash common.Hash, status chan<- uint) (ethereum.Subscription, error) {
func waitForConfirmation(txHash common.Hash, pollingDelayInSeconds int) uint {
	var status uint
	pollingDelayDuration := time.Duration(pollingDelayInSeconds) * time.Second
	for {
		// access confirmation status until we get the correct response
		// need to be aware of possible *known transaction* error
		// to prevent from calling this method unless we have a pending tx
		pendingCount, _ := getPendingTransactions()
		if pendingCount > 0 {
			continue
		}
		// need to break, need to provide older hash from previous txs to enact the error
		// could be done and ready for confirmation
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

// WaitForTransfer is blocking call that will observe on brokerAddr on transfer of PRL or ETH.
// If it is completed return number of PRL.
// TODO:  Does this need to return both PRL and ETH?
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
		Addresses: nil, //[]common.Address{MainWalletAddress},
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

func calculateGasToSend(desiredGasLimit uint64) (*big.Int, error) {
	gasPrice, err := getGasPrice()
	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return big.NewInt(0), err
	}
	gasToSend := new(big.Int).Mul(gasPrice, big.NewInt(int64(desiredGasLimit)))
	return gasToSend, nil
}

// Transfer funds from main wallet
func sendETH(from common.Address, fromPrivKey *ecdsa.PrivateKey, toAddr common.Address, amount *big.Int) (types.Transactions, string, int64, error) {

	client, err := sharedClient()
	if err != nil {
		return types.Transactions{}, "", -1, err
	}

	// initialize the context
	ctx, cancel := createContext()
	defer cancel()

	// generate nonce
	nonce, _ := client.PendingNonceAt(ctx, from)

	// default gasLimit on oysterby 4294967295
	gasPrice, _ := getGasPrice()

	// estimation
	estimate, failedEstimate := getEstimatedGasPrice(toAddr, from, GasLimitETHSend, *gasPrice, *amount)
	if failedEstimate != nil {
		fmt.Printf("failed to get estimated network price : %v\n", failedEstimate)
		return types.Transactions{}, "", -1, failedEstimate
	}
	estimatedGas := new(big.Int).SetUint64(estimate)
	fmt.Printf("estimatedGas : %v\n", estimatedGas)

	balance := checkETHBalance(from)
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

	var status = false

	// confirm status of transaction
	txStatus := waitForConfirmation(tx.Hash(), SecondsDelayForETHPolling)

	if txStatus == 0 {
		fmt.Printf("transaction failure")
		status = false
	} else if txStatus == 1 {
		fmt.Printf("confirmation completed")
		flushTransaction(tx.Hash())
		status = true
	}

	return status
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

	// default gasLimit on oysterby 4294967295
	gasPrice, _ := getGasPrice()
	currentBlock, _ := getCurrentBlock()
	gasLimit := currentBlock.GasLimit()

	// estimation
	estimate, failedEstimate := getEstimatedGasPrice(msg.To, msg.From, gasLimit, *gasPrice, msg.Amount)
	if failedEstimate != nil {
		fmt.Printf("failed to get estimated network price : %v\n", failedEstimate)
		return false
	}
	estimatedGas := new(big.Int).SetUint64(estimate)
	fmt.Printf("estimatedGas : %v\n", estimatedGas)

	balance := checkPRLBalance(msg.From)
	fmt.Printf("balance : %v\n", balance)

	// amount is greater than balance, return error
	if msg.Amount.Uint64() > balance.Uint64() {
		fmt.Println("balance too low to proceed")
		return false
	}
	fmt.Printf("sending prl to : %v\n", msg.To.Hex())
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

	var status = false

	// confirm status of transaction
	txStatus := waitForConfirmation(confirmTx.Hash(), SecondsDelayForETHPolling)

	if txStatus == 0 {
		fmt.Println("transaction failure")
		status = false
	} else if txStatus == 1 {
		fmt.Println("confirmation completed")
		status = true
	}

	return status
}

// send prl from oyster via contract transfer method
func sendPRLFromOyster(msg OysterCallMsg) (bool, string, int64) {

	client, _ := sharedClient()
	oysterPearl, err := NewOysterPearl(common.HexToAddress(OysterPearlContract), client)

	if err != nil {
		log.Printf("unable to access contract instance at : %v", err)
	}

	log.Printf("using wallet key store from: %v", msg.From)
	// initialize transactor // may need to move this to a session based transactor
	auth := bind.NewKeyedTransactor(&msg.PrivateKey)
	if err != nil {
		log.Printf("unable to create a new transactor : %v", err)
	}

	log.Printf("authorized transactor : %v", auth.From.Hex())

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

// record transaction
func recordTransaction(address common.Address, status string) {
	// when a successful transaction event, will need to change the status
	// of the correct row in the completed_uploads table.
	// expect "address" to be the "to" address of the gas transaction or
	// the "from" address of the PRL transaction.
	// do *not* use the broker's main wallet address
	switch status {
	case "sendGas":
		// gas transfers succeeded, call this method:
		models.SetGasStatusByAddress(address.Hex(), models.GasTransferSuccess)
	case "sendPRL":
		// PRL transfer succeeded, call this:
		models.SetPRLStatusByAddress(address.Hex(), models.PRLClaimSuccess)
	}
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

	if err != nil {
		oyster_utils.LogIfError(fmt.Errorf("unable to load key: %v", err), nil)
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

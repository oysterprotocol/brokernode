package services

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"fmt"
	"github.com/oysterprotocol/brokernode/utils"
	"log"
	"math/big"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/getsentry/raven-go"
	"github.com/joho/godotenv"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/pkg/errors"
)

type Eth struct {
	SendGas
	ClaimPRL
	ClaimUnusedPRLs
	GenerateEthAddr
	GenerateKeys
	GenerateEthAddrFromPrivateKey
	BuryPrl
	SendETH
	SendPRL
	GetGasPrice
	WaitForTransfer
	CheckBalance
	GetCurrentBlock
	OysterCallMsg
}

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

type SendGas func([]models.CompletedUpload) error
type GenerateEthAddr func() (addr common.Address, privateKey string, err error)
type GenerateKeys func(int) (privateKeys []string, err error)
type GenerateEthAddrFromPrivateKey func(privateKey string) (addr common.Address)
type GetGasPrice func() (*big.Int, error)
type WaitForTransfer func(brokerAddr common.Address) (*big.Int, error)
type CheckBalance func(common.Address) *big.Int
type GetCurrentBlock func() (*types.Block, error)
type SendETH func(toAddr common.Address, amount *big.Int) (rawTransaction string, err error)

type BuryPrl func(msg OysterCallMsg) bool
type SendPRL func(msg OysterCallMsg) bool
type ClaimPRL func(receiverAddress common.Address, treasureAddress common.Address, treasurePrivateKey string) bool
type ClaimUnusedPRLs func(uploadsWithUnclaimedPRLs []models.CompletedUpload) error

// Singleton client
var (
	oysterPearlContract string
	chainId             int64
	MainWalletAddress   common.Address
	MainWalletKey       string
	client              *ethclient.Client
	mtx                 sync.Mutex
	EthWrapper          Eth
)


func init() {
	// Load ENV variables
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println("Error loading .env file")
		raven.CaptureError(err, nil)
	}

	MainWalletAddress := common.HexToAddress(os.Getenv("MAIN_WALLET_ADDRESS"))
	MainWalletKey := os.Getenv("MAIN_WALLET_KEY")
	chainId := os.Getenv("CHAIN_ID")
	oysterPearlContract := os.Getenv("OYSTER_PEARL")
	ethUrl := os.Getenv("ETH_NODE_URL")

	fmt.Printf("Ethereum Network Address: %v\n", ethUrl)
	fmt.Printf("Ethereum Network Chain ID: %v\n", chainId)
	fmt.Printf("MainWallet Address: %v\n", MainWalletAddress)
	fmt.Printf("MainWallet Key: %v\n", MainWalletKey)
	fmt.Printf("Oyster Pearl Contract: %v\n", oysterPearlContract)

	EthWrapper = Eth{
		SendGas:                       sendGas,
		ClaimPRL:                      claimPRLs,
		ClaimUnusedPRLs:               claimUnusedPRLs,
		GenerateEthAddr:               generateEthAddr,
		GenerateKeys:                  generateKeys,
		GenerateEthAddrFromPrivateKey: generateEthAddrFromPrivateKey,
		BuryPrl:         buryPrl,
		SendETH:         sendETH,
		SendPRL:         sendPRL,
		GetGasPrice:     getGasPrice,
		WaitForTransfer: waitForTransfer,
		CheckBalance:    checkBalance,
		GetCurrentBlock: getCurrentBlock,
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
		raven.CaptureError(err, nil)
		return
	}
	// Sets Singleton
	client = c
	return
}

// Generate an Ethereum address
func generateEthAddr() (addr common.Address, privateKey string, err error) {
	ethAccount, err := crypto.GenerateKey()
	if err != nil {
		fmt.Printf("Could not generate eth key: %v\n", err)
		raven.CaptureError(err, nil)
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
	if privateKey[0:2] != "0x" && privateKey[0:2] != "0X" {
		privateKey = "0x" + privateKey
	}
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

// returns represents the 20 byte address of an ethereum account.
func StringToAddress(address string) common.Address {
	return common.HexToAddress(address)
}

// SuggestGasPrice retrieves the currently suggested gas price to allow a timely
// execution for new transaction
func getGasPrice() (*big.Int, error) {
	// connect ethereum client
	client, err := sharedClient()
	if err != nil {
		log.Fatal("Could not get gas price from network")
	}

	// there is no guarantee with estimate gas price
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal("Client could not get gas price from network")
		raven.CaptureError(err, nil)
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
	msg.Value =  &value

	// estimated gas price
	estimatedGasPrice, err := client.EstimateGas(context.Background(), *msg)
	if err != nil {
		log.Fatal("Client could not get gas price estimate from network")
		raven.CaptureError(err, nil)
	}
	return estimatedGasPrice, nil
}

// Check balance from a valid Ethereum network address
func checkBalance(addr common.Address) *big.Int {
	// connect ethereum client
	client, err := sharedClient()
	if err != nil {
		log.Fatal("Could not initialize shared client")
	}

	balance, err := client.BalanceAt(context.Background(), addr, nil) //Call(&bal, "eth_getBalance", addr, "latest")
	if err != nil {
		fmt.Println("Client could not retrieve balance:", err)
		raven.CaptureError(err, nil)
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
		fmt.Printf("Could not get last block: %v\n", err)
		raven.CaptureError(err, nil)
		return nil, err
	}

	// latest block event
	fmt.Printf("latest block: %v\n", currentBlock.Number())
	return currentBlock, nil
}

// WaitForTransfer is blocking call that will observe on brokerAddr on transfer on ETH.
// If it is completed return number of PRL.
func waitForTransfer(brokerAddr common.Address) (*big.Int, error) {
	client, err := sharedClient()
	if err != nil {
		return big.NewInt(0), err
	}

	currentBlock, err := getCurrentBlock()
	if err != nil {
		return big.NewInt(0), err
	}

	q := ethereum.FilterQuery{
		FromBlock: currentBlock.Number(), // beginning of the queried range, nil means genesis block
		ToBlock:   nil,                   // end of the range, nil means latest block
		Addresses: []common.Address{brokerAddr},
		Topics:    nil, // matches any topic list
	}

	logChan := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(context.Background(), q, logChan)
	if err != nil {
		raven.CaptureError(err, nil)
		return big.NewInt(0), err
	}

	defer sub.Unsubscribe()
	for {
		select {
		case err := <-sub.Err():
			log.Fatal(err)
			raven.CaptureError(err, nil)
			return big.NewInt(0), err
		case <-time.After(1 * time.Hour):
			log.Print("Timeout to wait for brokerAddr\n")
			// Wait for 1 hr to receive payment before timeout
			return big.NewInt(0), errors.New("Timeout")
			// TODO(astor): listen to the event and return true/false
			/*
				case log := <- outCh:
					fmt.Printf("Log Data:%v", log.Data)

					// need to add unpack abi result if
					// the method call is for the contract
					if err != nil {
						fmt.Println("Failed to unpack:", err)
					}

					fmt.Println("Confirmed Address:", log.Address.Hex())

					sub.Unsubscribe()

					// TODO ensure confirmation type from "sendGas" or "sendPRL"
					recordTransaction(log.Address, "")*/
		}
	}
}

// Send gas to the completed upload Ethereum account
func sendGas(completedUploads []models.CompletedUpload) error {
	for _, completedUpload := range completedUploads {
		// returns a raw transaction, we may need to store them to verify all transactions are completed
		// mock value need to get amount, not in completed upload object
		gasPrice, _ := getGasPrice()
		sendETH(StringToAddress(completedUpload.ETHAddr), gasPrice)
	}
	return nil
}

// Transfer funds from one Ethereum account to another.
// We need to pass in the credentials, to allow the transaction to execute.
func sendETH(toAddr common.Address, amount *big.Int) (rawTransaction string, err error) {

	client, err := sharedClient()
	if err != nil {
		return "", err
	}

	// initialize the context
	deadline := time.Now().Add(1000 * time.Millisecond)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	// generate nonce
	nonce, _ := client.NonceAt(ctx, MainWalletAddress, nil)

	// default gasLimit on oysterby 4294967295
	// gas * price + value

	gasPrice := big.NewInt(36505)//getGasPrice()
	currentBlock, _ := getCurrentBlock()
	gasLimit := currentBlock.GasLimit()

	estimate, failedEstimate := getEstimatedGasPrice(toAddr, MainWalletAddress, 0, *gasPrice, *amount)
	if failedEstimate != nil {
		return "", failedEstimate
	}

	estimatedGas := new(big.Int).SetUint64(estimate)

	// create new transaction
	tx := types.NewTransaction(nonce, toAddr, amount, gasLimit, estimatedGas, nil)

	// oysterby chainId 559966
	chainId := big.NewInt(chainId)
	// MainWalletKey
	privateKey, _ := crypto.HexToECDSA(MainWalletKey)
	signer := types.NewEIP155Signer(chainId)
	signedTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		raven.CaptureError(err, nil)
		return "", err
	}

	// send transaction
	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		raven.CaptureError(err, nil)
		return "", err
	}

	// pull signed transaction
	ts := types.Transactions{signedTx}
	// return raw transaction
	rawTransaction = string(ts.GetRlp(0))

	return
}

// Bury PRLs
func buryPrl(msg OysterCallMsg) bool {

	// dispense PRLs from the transaction address to each 'treasure' address
	rawTransaction, err := sendETH(msg.To, &msg.Amount)

	if err != nil || len(rawTransaction) == 0 {
		// sending eth has failed
		return false
	}

	// initialize the context
	ctx, cancel := createContext()
	defer cancel()
	// shared client
	client, err := sharedClient()
	if err != nil {
		return false
	}

	// abi
	oysterABI, err := abi.JSON(strings.NewReader(OysterPearlABI))
	// oyster contract method bury() no args
	buryPRL, _ := oysterABI.Pack("bury")
	// build transaction and sign
	signedTx, err := callOysterPearl(ctx, buryPRL)
	// send transaction
	err = client.SendTransaction(ctx, signedTx)

	if err != nil {
		raven.CaptureError(err, nil)
		return false
	}
	// pull signed transaction
	ts := types.Transactions{signedTx}
	// return raw transaction
	rawTransaction = string(ts.GetRlp(0))

	// successful contract message call
	return len(rawTransaction) > 0
}

// ClaimUnusedPRLs parses the completedUploads and sends PRL to the MainWalletAddress
func claimUnusedPRLs(completedUploads []models.CompletedUpload) error {
	// Contract claim(address _payout, address _fee) public returns (bool success)
	for _, completedUpload := range completedUploads {
		//	for each completed upload, get its PRL balance from its ETH
		//	address (completedUpload.ETHAddr) by calling CheckBalance.
		ethAddr := StringToAddress(completedUpload.ETHAddr)
		balance := checkBalance(ethAddr)
		if balance.Int64() <= 0 {
			// need to log this error to apply a retry
			err := errors.New("could not complete transaction due to zero balance for:" + completedUpload.ETHAddr)
			raven.CaptureError(err, nil)
			return err
		}
		//	Then, using SendPRL, create a transaction with each
		//	completedUpload.ETHAddr as the "fromAddr" address, the broker's
		//	main wallet (MainWalletAddress) as the "toAddr" address,
		from := ethAddr
		to := MainWalletAddress
		// privateKey := completedUpload.ETHPrivateKey
		//3.
		// 	and the PRL balance of completedUpload.ETHAddr as the "amt" to send,
		// 	and subscribe to the event with SubscribeToTransfer.
		var amountToSend = balance
		var gas = uint64(vm.GAS) // TODO get gas source are we pulling from ETHAddr?
		gasPrice, _ := getGasPrice()

		// prepare oyster message call
		var oysterMsg = OysterCallMsg{
			From:     from,
			To:       to,
			Amount:   *amountToSend,
			Gas:      gas,
			GasPrice: *gasPrice,
			TotalWei: *big.NewInt(1), // TODO finish wei
			Data:     []byte(""),     // setup data
		}

		// claimed := claimPRLs(to, from, privateKey)

		// send transaction from completed upload eth addr to main wallet
		// we may just do a straight transfer with network vs from contract
		if !sendPRL(oysterMsg) {
			// TODO more detailed error message
			err := errors.New("unable to send prl to main wallet")
			raven.CaptureError(err, nil)
			return err
		}
	}

	return nil
}

// Claim PRL allows the receiver to unlock the treasure address and private key to enable the transfer
func claimPRLs(receiverAddress common.Address, treasureAddress common.Address, treasurePrivateKey string) bool {
	// initialize the context
	ctx, cancel := createContext()
	defer cancel()
	// shared client
	client, err := sharedClient()
	if err != nil {
		return false
	}

	// abi
	oysterABI, err := abi.JSON(strings.NewReader(OysterPearlABI))
	// oyster contract method bury() no args
	claimPRL, _ := oysterABI.Pack("claim")
	// build transaction and sign
	signedTx, err := callOysterPearl(ctx, claimPRL)
	// send transaction
	err = client.SendTransaction(ctx, signedTx)

	if err != nil {
		raven.CaptureError(err, nil)
		return false
	}
	// pull signed transaction
	ts := types.Transactions{signedTx}

	// successful contract message call
	return ts.Len() > 0
}

/*
	When a user uploads a file, we create an upload session on the broker.

	For each "upload session", we generate a new wallet (we do this so we can associate a session to PRLs sent).
	The broker responds to the uploader with an invoice to send X PRLs to Y eth address
	The broker then listens for a transfer event so it knows when payment has happened,

	Once the payment is received, the brokers will split up the PRLs and begin work to attach the file to the IOTA tangle
	so to answer your question, there won't be a main PRL wallet, the address is different for each session
	however there will be a "main" ETH wallet, which is used to pay gas fees
*/
func sendPRL(msg OysterCallMsg) bool {

	// initialize the context
	ctx, cancel := createContext()
	defer cancel()

	// shared client
	client, err := sharedClient()
	if err != nil {
		return false
	}

	// abi
	oysterABI, err := abi.JSON(strings.NewReader(OysterPearlABI))
	// oyster contract method transfer(address _to, uint256 _value)
	sendPRL, _ := oysterABI.Pack("transfer", msg.To.Hex(), msg.Amount)
	// build transaction and sign
	signedTx, err := callOysterPearl(ctx, sendPRL)
	// send transaction
	err = client.SendTransaction(ctx, signedTx)

	if err != nil {
		raven.CaptureError(err, nil)
		return false
	}
	// pull signed transaction
	ts := types.Transactions{signedTx}
	// return raw transaction
	rawTransaction := string(ts.GetRlp(0))

	// successful contract message call
	return len(rawTransaction) > 0
}

// utility to call a method on OysterPearl contract
func callOysterPearl(ctx context.Context, data []byte) (*types.Transaction, error) {

	// invoke the smart contract bury() function with 'treasure'
	// Oyster Pearl on Ethereum Network
	contractAddress := common.HexToAddress(oysterPearlContract)

	// oysterby chainId 559966 - env
	chainId := big.NewInt(559966)
	privateKey, err := crypto.HexToECDSA(MainWalletKey)
	if err != nil {
		fmt.Printf("Failed to parse secp256k1 private key")
		raven.CaptureError(err, nil)
		return nil, err
	}
	client, err := sharedClient()
	if err != nil {
		return nil, err
	}
	token, err := NewOysterPearl(contractAddress, client)
	if err != nil {
		fmt.Print("Unable to instantiate OysterPearl")
	}
	name, err := token.Name(nil)
	fmt.Printf("OysterPearl :%v",name)

	nonce, _ := client.NonceAt(ctx, MainWalletAddress, nil)

	// get latest gas limit & price - current default gasLimit on oysterby 21000
	gasLimit := uint64(vm.GASLIMIT) // may pull gas limit from estimate gas price
	gasPrice, _ := getGasPrice()

	// create new transaction with 0 amount
	tx := types.NewTransaction(nonce, contractAddress, big.NewInt(0), gasLimit, gasPrice, data)

	signer := types.NewEIP155Signer(chainId)
	signedTx, _ := types.SignTx(tx, signer, privateKey)

	return signedTx, nil
}

// context helper to include the deadline initialization
func createContext() (ctx context.Context, cancel context.CancelFunc) {
	deadline := time.Now().Add(1000 * time.Millisecond)
	return context.WithDeadline(context.Background(), deadline)
}

// TODO will be Use channels/workers for subscribe to transaction events
// There is an example of a channel/worker in iota_wrappers.go
// These methods live in models/completed_uploads.go
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

package services

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/getsentry/raven-go"
	"github.com/joho/godotenv"
	"github.com/oysterprotocol/brokernode/models"
	"log"
	"os"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	"crypto/ecdsa"
	"errors"
)

type Eth struct {
	SendGas             SendGas
	ClaimPRLs           ClaimPRLs
	GenerateEthAddr     GenerateEthAddr
	BuryPrl             BuryPrl
	SendETH             SendETH
	SendPRL             SendPRL
	GetGasPrice         GetGasPrice
	SubscribeToTransfer SubscribeToTransfer
	CheckBalance        CheckBalance
	GetCurrentBlock     GetCurrentBlock
}

type SendGas func([]models.CompletedUpload) error
type ClaimPRLs func([]models.CompletedUpload) error
type GenerateEthAddr func() (addr string, privKey string, err error)
type BuryPrl func()
type SendETH func(fromAddr common.Address, toAddr common.Address, amt *big.Int, privateKey *ecdsa.PrivateKey) (rawTransaction string)
type SendPRL func(fromAddr common.Address, toAddr common.Address, amt float64)
type GetGasPrice func() (*big.Int, error)
type SubscribeToTransfer func(brokerAddr common.Address, outCh chan<- types.Log)
type CheckBalance func(common.Address) (*big.Int)
type GetCurrentBlock func() (*types.Block, error)

// Singleton client
var (
	ethUrl            string
	MainWalletAddress common.Address
	MainWalletKey     string
	client            *ethclient.Client
	mtx               sync.Mutex
	EthWrapper        Eth
)

func init() {
	// Load ENV variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
		raven.CaptureError(err, nil)
	}

	MainWalletAddress := os.Getenv("MAIN_WALLET_ADDRESS")
	MainWalletKey := os.Getenv("MAIN_WALLET_KEY")
	ethUrl := os.Getenv("ETH_NODE_URL")

	fmt.Println(MainWalletAddress)
	fmt.Println(MainWalletKey)
	fmt.Println(ethUrl)

	EthWrapper = Eth{
		SendGas:         sendGas,
		ClaimPRLs:       claimPRLs,
		GenerateEthAddr: generateEthAddr,
		BuryPrl:         buryPrl,
		SendETH:         sendETH,
		SendPRL:         sendPRL,
		GetGasPrice:     getGasPrice,
		SubscribeToTransfer: subscribeToTransfer,
		CheckBalance:        checkBalance,
		GetCurrentBlock:     getCurrentBlock,
	}
}

// Shared client provides access to the underlying Ethereum client
func sharedClient(netUrl string) (c *ethclient.Client, err error) {
	if client != nil {
		return client, nil
	}
	// check-lock-check pattern to avoid excessive locking.
	mtx.Lock()
	defer mtx.Unlock()

	if client != nil {
		// override to allow custom node url
		if len(netUrl)>0 {
			ethUrl = netUrl
		}
		c, err = ethclient.Dial(ethUrl)
		if err != nil {
			fmt.Println("Failed to dial in to Ethereum node.")
			return
		}
		// Sets Singleton
		client = c
	}
	return client, err
}

// Generate an Ethereum address
func generateEthAddr() (addr common.Address, privateKey string, err error) {
	ethAccount, err := crypto.GenerateKey()
	if err != nil {
		return
	}
	addr = crypto.PubkeyToAddress(ethAccount.PublicKey)
	privateKey = hex.EncodeToString(ethAccount.D.Bytes())
	return addr, privateKey, err
}

// returns represents the 20 byte address of an ethereum account.
func stringToAddress(address string) common.Address {
	var stringByte [20]byte
	decodedBytes, _ :=  hex.DecodeString(address[2:])
	copy(stringByte[:], decodedBytes)
	return common.Address([20]byte(stringByte))
}

// SuggestGasPrice retrieves the currently suggested gas price to allow a timely
// execution for new transaction
func getGasPrice() (*big.Int, error) {
	// connect ethereum client
	client, err := sharedClient("")
	if err != nil {
		log.Fatal("Could not get gas price from network")
	}

	// there is no guarantee with estimate gas price
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal("Client could not get gas price from network")
	}
	return gasPrice, nil
}


// Check balance from a valid Ethereum network address
func checkBalance(addr common.Address) (*big.Int) {
	// connect ethereum client
	client, err := sharedClient("")
	if err != nil {
		log.Fatal("Could not initialize shared client")
	}

	balance, err := client.BalanceAt(context.Background(),addr, nil)  //Call(&bal, "eth_getBalance", addr, "latest")
	if err != nil {
		fmt.Println("Client could not retrieve balance:", err)
		return big.NewInt(0)
	}
	return balance
}

// Get current block from blockchain
func getCurrentBlock() (*types.Block, error) {
	// connect ethereum client
	client, err := sharedClient("")
	if err != nil {
		log.Fatal("Could not connect to Ethereum network", err)
		return nil, err
	}

	// latest block number is nil to get the latest block
	currentBlock, err := client.BlockByNumber(context.Background(), nil)
	if err != nil {
		fmt.Printf("Could not get last block: %v\n", err)
		return nil, err
	}

	// latest block event
	fmt.Printf("latest block: %v\n", currentBlock.Number)
	return currentBlock, nil
}

// SubscribeToTransfer will subscribe to transfer events
// sending PRL to the brokerAddr given. Notifications
// will be sent in the out channel provided.
func subscribeToTransfer(brokerAddr common.Address, outCh chan<- types.Log) {
	client, _ := sharedClient("")
	currentBlock, _ := getCurrentBlock()
	q := ethereum.FilterQuery{
		FromBlock: currentBlock.Number(), // beginning of the queried range, nil means genesis block
		ToBlock: nil, // end of the range, nil means latest block
		Addresses: []common.Address{brokerAddr},
		Topics:    nil, // TODO: scope this to just oyster's contract.
	}
	// subscribe before passing it to outCh.
	client.SubscribeFilterLogs(context.Background(), q, outCh)
}

// Send gas to the completed upload Ethereum account
func sendGas(completedUploads []models.CompletedUpload) (error) {

	// retrieve current gas price
	gas, err := getGasPrice()
	if err != nil {
		return err
	}
	// collection of raw transactions
	var completedTransactions []string

	for _, completedUpload := range completedUploads {
		var privateKey = ecdsa.PrivateKey{} // completedUpload.ETHPrivateKey
		// returns a raw transaction, we may need to store them to verify all transactions are completed
		sendETH(MainWalletAddress, stringToAddress(completedUpload.ETHAddr), gas, &privateKey)
	}

	return nil
}

// Transfer funds from one Ethereum account to another.
// We need to pass in the credentials, to allow the transaction to execute.
func sendETH(fromAddr common.Address, toAddr common.Address, amt *big.Int, privateKey *ecdsa.PrivateKey) (rawTransaction string){

	data := []byte("") // setup data
	gasLimit := uint64(10000) // gasLimit
	gasPrice, _ := getGasPrice() // get gas price
	tx := types.NewTransaction(99, toAddr, amt, gasLimit, gasPrice, data)

	// 1999 temp chainId identifier, will need to pass in env
	signer := types.NewEIP155Signer(big.NewInt(1999))
	hash := tx.Hash().Bytes()

	signature, _ := crypto.Sign(hash, privateKey)
	signedTx, _ := tx.WithSignature(signer, signature)

	ts := types.Transactions{signedTx}
	rawTransaction = string(ts.GetRlp(0))

	return
}

type OysterCallMsg struct {
	From common.Address
	To common.Address
	Amount big.Int
	PrivateKey ecdsa.PrivateKey
	Gas uint64
	GasPrice big.Int
	TotalWei big.Int
	Data []byte
}

// Bury PRLs
func buryPrl(msg OysterCallMsg) (bool) {

	// dispense PRLs from the transaction address to each 'treasure' address
	sendETH(msg.From, msg.To, &msg.Amount, &msg.PrivateKey)

	// invoke the smart contract bury() function with 'treasure'
	// contract bury() public returns (bool success)
	// initialize call msg for oyster prl contract
	var callMsg = ethereum.CallMsg {
		From:msg.From,         		// the sender of the 'transaction'
		To:&msg.To,            		// the destination contract (nil for contract creation)
		Gas:msg.Gas,           		// if 0, the call executes with near-infinite gas
		GasPrice:&msg.GasPrice,		// wei <-> gas exchange ratio
		Value:&msg.TotalWei,   		// amount of wei sent along with the call
		Data:msg.Data, 				// ABI-encoded contract method bury
	}
	// call  byte[] or error
	contractResponse, error := client.CallContract(context.Background(), callMsg, big.NewInt(1))
	if error != nil {
		// TODO More descriptive contract error response
		return false
	}
	// successful contract message call
	if contractResponse != nil {
		return true
	} else {
		return false
	}
}

// Claim PRLs from OysterPearl contract
func claimPRLs(completedUploads []models.CompletedUpload) error {

	// Contract claim(address _payout, address _fee) public returns (bool success)

	for _, completedUpload := range completedUploads {
		//1
		//	    for each completed upload, get its PRL balance from its ETH
		//		address (completedUpload.ETHAddr) by calling CheckBalance.
		var ethAddr = stringToAddress(completedUpload.ETHAddr)
		var balance = checkBalance(ethAddr)
		if balance.Int64() <= 0 {
			// need to log this error to apply a retry
			return errors.New("could not complete transaction due to zero balance for:"+completedUpload.ETHAddr)
		}
		//2.
		//	    Then, using SendPRL, create a transaction with each
		//	    completedUpload.ETHAddr as the "fromAddr" address, the broker's
		//	    main wallet (MainWalletAddress) as the "toAddr" address,
		var from = ethAddr
		var to = MainWalletAddress
		//3.
		// 		and the PRL balance of completedUpload.ETHAddr as the "amt" to send,
		// 		and subscribe to the event with SubscribeToTransfer.
		var amountToSend = balance
		var gas = uint64(0)
		gasPrice, _ := getGasPrice()

		// prepare oyster message call
		var oysterMsg = OysterCallMsg{
			From: from,
			To: to,
			Amount: *amountToSend,
			Gas: gas,
			GasPrice: *gasPrice,
			TotalWei: *big.NewInt(1),
			Data: []byte,
		}
		// send transaction from completed upload eth addr to main wallet
		if !sendPRL(oysterMsg) {
			// TODO more detailed error message
			return errors.New("unable to send prl to main wallet")
		}
	}

	return nil
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
func sendPRL(msg OysterCallMsg) (bool) {

	// send PRL from Oyster contract call
	var callMsg = ethereum.CallMsg {
		From:msg.From,         		// the sender of the 'transaction'
		To:&msg.To,            		// the destination contract (nil for contract creation)
		Gas:msg.Gas,           		// if 0, the call executes with near-infinite gas
		GasPrice:&msg.GasPrice,		// wei <-> gas exchange ratio
		Value:&msg.TotalWei,   		// amount of wei sent along with the call
		Data:msg.Data, 				// ABI-encoded contract method bury
	}
	// call  byte[] or error
	contractResponse, error := client.CallContract(context.Background(), callMsg, big.NewInt(1))
	if error != nil {
		// TODO More descriptive contract error response for failed send
		return false
	}
	// successful contract message call
	if contractResponse != nil {
		return true
	} else {
		return false
	}
}



/* TODO will be using channels/workers for subscribe to transaction events

	-there is an example of a channel/worker in iota_wrappers.go

	-when you see a successful transaction event, will need to change the status
    of the correct row in the completed_uploads table.

	-to indicate that a gas transfers has succeedded, call this method:

    models.SetGasStatusByAddress(address, models.GasTransferSuccess)

    -to indicate that a PRL transfer succeeded, call this:

    models.SetPRLStatusByAddress(address, models.PRLClaimSuccess)

    Both methods currently expect address to be a string (change that if you want)
    and expect "address" to be the "to" address of the gas transaction or the "from"
    address of the PRL transaction.  In other words, *not* the broker's main wallet
    address.

    These methods live in models/completed_uploads.go
*/

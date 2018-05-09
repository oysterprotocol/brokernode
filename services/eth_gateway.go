package services

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/getsentry/raven-go"
	"github.com/joho/godotenv"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/utils"
	"gopkg.in/segmentio/analytics-go.v3"
	"log"
	"os"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Eth struct {
	SendGas             SendGas
	ClaimUnusedPRLs     ClaimUnusedPRLs
	GenerateEthAddr     GenerateEthAddr
	BuryPrl             BuryPrl
	SendETH             SendETH
	SendPRL             SendPRL
	GetGasPrice         GetGasPrice
	SubscribeToTransfer SubscribeToTransfer
	CheckBalance        CheckBalance
	GetCurrentBlock     GetCurrentBlock
	Claim               Claim
}

type SendGas func([]models.CompletedUpload) error
type ClaimUnusedPRLs func([]models.CompletedUpload) error
type GenerateEthAddr func() (addr string, privKey string, err error)
type BuryPrl func()
type SendETH func(fromAddr common.Address, toAddr common.Address, amt float64)
type SendPRL func(fromAddr common.Address, toAddr common.Address, amt float64)
type GetGasPrice func() (uint64, error)
type SubscribeToTransfer func(brokerAddr common.Address, outCh chan<- types.Log)
type CheckBalance func(common.Address)
type GetCurrentBlock func()
type Claim func(from string, to string) (result bool)

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
		ClaimUnusedPRLs: claimUnusedPRLs,
		GenerateEthAddr: generateEthAddr,
		BuryPrl:         buryPrl,
		SendETH:         sendETH,
		SendPRL:         sendPRL,
		//GetGasPrice:     getGasPrice,
		SubscribeToTransfer: subscribeToTransfer,
		CheckBalance:        checkBalance,
		GetCurrentBlock:     getCurrentBlock,
		Claim:               claim,
	}
}

func sharedClient() (c *ethclient.Client, err error) {
	if client != nil {
		return client, nil
	}

	// check-lock-check pattern to avoid excessive locking.
	mtx.Lock()
	defer mtx.Unlock()

	if client != nil {
		c, err = ethclient.Dial(ethUrl)
		if err != nil {
			fmt.Println("Failed to connect to Etherum node.")
			return
		}

		// Sets Singleton
		client = c
	}

	return client, err
}

func generateEthAddr() (addr string, privKey string, err error) {
	ethAccount, err := crypto.GenerateKey()
	if err != nil {
		return
	}

	addr = crypto.PubkeyToAddress(ethAccount.PublicKey).Hex()
	privKey = hex.EncodeToString(ethAccount.D.Bytes())

	oyster_utils.LogToSegment("generated_new_eth_address", analytics.NewProperties().
		Set("eth_address", fmt.Sprint(addr)))

	return
}

func buryPrl() {
	// TODO

	/*
		At a high level, what this method will have to do is send PRLs from the transaction address to each 'treasure'
		address, and invoke the smart contract bury() function on them.
	*/

	//oyster_utils.LogToSegment("bury_invoked", analytics.NewProperties().
	//	Set("eth_address", fmt.Sprint(addr)))
}

func sendGas(completedUploads []models.CompletedUpload) error {
	//gas, err := GetGasPrice()
	//for _, completedUpload := range completedUploads {

	//oyster_utils.LogToSegment("sending_gas", analytics.NewProperties().
	//	Set("eth_address_from", MainWalletAddress).
	//	Set("eth_address_to", completedUpload.ETHAddr).
	//	Set("genesis_hash", completedUpload.GenesisHash))

	//	Eth.SendETH(MainWalletAddress, completedUpload.ETHAddr, float64(gas))
	//}
	//if err != nil {
	//	return err
	//}
	return nil
}

// TODO: Don't use floats for money transactions!
func sendETH(fromAddr common.Address, toAddr common.Address, amt float64) {
	// TODO
}

func claimUnusedPRLs(completedUploads []models.CompletedUpload) error {

	//for _, completedUpload := range completedUploads {
	//	/* TODO:
	//
	//	    for each completed upload, get its PRL balance from its ETH
	//		address (completedUpload.ETHAddr) by calling CheckBalance.
	//
	//	    Then, using SendPRL, create a transaction with each
	//	    completedUpload.ETHAddr as the "fromAddr" address, the broker's
	//	    main wallet (MainWalletAddress) as the "toAddr" address, and the PRL
	//	    balance of completedUpload.ETHAddr as the "amt" to send, and
	//	    subscribe to the event with SubscribeToTransfer.
	//	*/

	//oyster_utils.LogToSegment("send_unused_prls_back_to_broker", analytics.NewProperties().
	//	Set("eth_address_from", completedUpload.ETHAddr).
	//	Set("eth_address_to", MainWalletAddress).
	//	Set("genesis_hash", completedUpload.GenesisHash))

	//}
	//if err != nil {
	//	return err
	//}

	return nil
}

func sendPRL(fromAddr common.Address, toAddr common.Address, amt float64) {
	// TODO
}

//func GetGasPrice() (uint64, error) {
//	//ctx := context.Background() // TODO: Should we have some timeout or cancel?
//
//	//TODO:  Decide whether 'SuggestGasPrice' or 'EstimateGasPrice' is the better method, then
//	// return the gas and the error
//	//gas, errEstimator := client.SuggestGasPrice(ctx)
//}

// SubscribeToTransfer will subscribe to transfer events
// sending PRL to the brokerAddr given. Notifications
// will be sent in the out channel provided.
func subscribeToTransfer(brokerAddr common.Address, outCh chan<- types.Log) {

	/*TODO:  Add segment logging of a detected transfer event */

	ethCl, _ := sharedClient()

	ctx := context.Background() // TODO: Should we have some timeout or cancel?

	q := ethereum.FilterQuery{
		FromBlock: nil, // TODO: Figure out how to get current block in GetCurrentBlock
		Addresses: []common.Address{brokerAddr},
		Topics:    nil, // TODO: scope this to just oyster's contract.
	}
	// Adapt msg sent from the subscription before passing it to outCh.
	ethCl.SubscribeFilterLogs(ctx, q, outCh)
}

func checkBalance(addr common.Address) {
	//ctx := context.Background() // TODO: Should we have some timeout or cancel?
	//client.BalanceAt(ctx, addr)
}

func getCurrentBlock() {

}

func claim(from string, to string) (result bool) {
	result = true
	return result
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

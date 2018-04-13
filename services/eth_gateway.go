package services

import (
	"context"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// TODO: Find etherum node to connect to.
const ethUrl = "ws://0.0.0.0"

// Singleton client
var (
	client *ethclient.Client
	mtx    sync.Mutex
)

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

func GenerateEthAddr() (addr string, privKey string, err error) {
	ethAccount, err := crypto.GenerateKey()
	if err != nil {
		return
	}

	addr = crypto.PubkeyToAddress(ethAccount.PublicKey).Hex()
	privKey = hex.EncodeToString(ethAccount.D.Bytes())

	return
}

func BuryPrl() {
	// TODO
}

// TODO: Don't use floats for money transactions!
func SendTransaction(fromAddr common.Address, toAddr common.Address, amt float64) {
	// TODO
}

// SubscribeToTransfer will subscribe to transfer events
// sending PRL to the brokerAddr given. Notifications
// will be sent in the out channel provided.
func SubscribeToTransfer(brokerAddr common.Address, outCh chan<- types.Log) {
	ethCl, _ := sharedClient()
	ctx := context.Background() // TODO: Should we have some timeout or cancel?

	q := ethereum.FilterQuery{
		FromBlock: nil, // TODO: Figure out how to get current block
		Addresses: []common.Address{brokerAddr},
		Topics:    nil, // TODO: scope this to just oyster's contract.
	}
	// Adapt msg sent from the subscription before passing it to outCh.
	ethCl.SubscribeFilterLogs(ctx, q, outCh)
}

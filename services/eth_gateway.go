package services

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/ethclient"
)

const ethUrl = "ws://0.0.0.0"

// Singleton client
var client *ethclient.Client
var mtx sync.Mutex

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

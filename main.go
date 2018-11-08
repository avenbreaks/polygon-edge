package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/umbracle/minimal/protocol"
	"github.com/umbracle/minimal/protocol/ethereum"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/umbracle/minimal/network"
)

// mainnet nodes
var peers = []string{}

func main() {
	fmt.Println("## Minimal ##")

	logger := log.New(os.Stderr, "", log.LstdFlags)

	privateKey := "b4c65ef6b82e96fb5f26dc10a79c929985217c078584721e9157c238d1690b22"

	key, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		panic(err)
	}

	// Start network protocol

	config := network.DefaultConfig()
	config.Bootnodes = readFile("./foundation.txt")

	server, err := network.NewServer("minimal", key, config, logger)
	if err != nil {
		panic(err)
	}

	// register protocols

	// mainnet status
	status := func() (*ethereum.Status, error) {
		s := &ethereum.Status{ // mainnet status
			ProtocolVersion: 63,
			NetworkID:       1,
			TD:              big.NewInt(17179869184),
			CurrentBlock:    common.HexToHash("0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3"),
			GenesisBlock:    common.HexToHash("0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3"),
		}
		return s, nil
	}

	callback := func(conn network.Conn, peer *network.Peer) protocol.Handler {
		return ethereum.NewEthereumProtocol(conn, peer, status)
	}

	server.RegisterProtocol(protocol.ETH63, callback)

	// connect to some peers

	for _, i := range config.Bootnodes {
		server.Dial(i)
	}

	go func() {
		for {
			select {
			case evnt := <-server.EventCh:
				if evnt.Type == network.NodeJoin {
					fmt.Printf("Node joined: %s\n", evnt.Peer.ID)
				}
			}
		}
	}()

	handleSignals(server)
}

func handleSignals(s *network.Server) int {
	signalCh := make(chan os.Signal, 4)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	var sig os.Signal
	select {
	case sig = <-signalCh:
	}

	fmt.Printf("Caught signal: %v\n", sig)
	fmt.Printf("Gracefully shutting down agent...\n")

	gracefulCh := make(chan struct{})
	go func() {
		s.Close()
		close(gracefulCh)
	}()

	select {
	case <-signalCh:
		return 1
	case <-time.After(5 * time.Second):
		return 1
	case <-gracefulCh:
		return 0
	}
}

func readFile(s string) []string {
	data, err := ioutil.ReadFile(s)
	if err != nil {
		panic(err)
	}
	return strings.Split(string(data), "\n")
}
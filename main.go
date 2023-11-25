package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	golog "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p/core/peer"
	_ "github.com/libp2p/go-libp2p/p2p/host/peerstore"
	ma "github.com/multiformats/go-multiaddr"
	gologging "github.com/whyrusleeping/go-logging"
	"log"
	"math"
	"myBlockchain/chain"
)

func main() {

	genesisBlock := chain.CreateGenesisBlock(0, 0)
	chain.Blockchain = append(chain.Blockchain, genesisBlock)

	golog.SetAllLoggers(golog.LogLevel(gologging.INFO)) // Change to DEBUG for extra info
	listenF := flag.Int("l", 0, "wait for incoming connections")
	target := flag.String("d", "", "target peer to dial")
	flag.Parse()

	if *listenF == 0 {
		log.Fatal("Please provide a port to bind on with -l")
	}

	host, err := chain.CreateHost(*listenF)
	if err != nil {
		log.Fatal(err)
	}
	memTransactions := make([]chain.Transaction, 0)
	stream := chain.PeerStream{Host: host, MemTransactions: memTransactions}

	if *target == "" {
		log.Println("listening for connections")
		host.SetStreamHandler("/p2p/1.0.0", stream.HandleStream)
	} else {
		host.SetStreamHandler("/p2p/1.0.0", stream.HandleStream)
		ipfsAddr, err := ma.NewMultiaddr(*target)
		if err != nil {
			log.Fatalln(err)
		}

		pid, err := ipfsAddr.ValueForProtocol(ma.P_IPFS)
		if err != nil {
			log.Fatalln(err)
		}

		peerId, err := peer.Decode(pid)
		if err != nil {
			log.Fatalln(err)
		}

		targetPeerAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", peerId.String()))
		targetAddr := ipfsAddr.Decapsulate(targetPeerAddr)
		host.Peerstore().AddAddr(peerId, targetAddr, math.MaxInt64)
		log.Println("opening stream")

		s, err := host.NewStream(context.Background(), peerId, "/p2p/1.0.0")
		if err != nil {
			log.Fatalln(err)
		}
		rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
		go stream.ReadStream(rw)
		go stream.HandleCli(rw)
	}

	select {}
}

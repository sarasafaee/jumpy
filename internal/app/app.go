package app

import (
	"bufio"
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p/core/peer"
	_ "github.com/libp2p/go-libp2p/p2p/host/peerstore"
	ma "github.com/multiformats/go-multiaddr"
	"log"
	"math"
	"myBlockchain/chain"
)

func Start(listenPort int, targetPeer string) {

	genesisBlock := chain.CreateGenesisBlock(0, 0)
	chain.Blockchain = append(chain.Blockchain, genesisBlock)

	host, err := chain.CreateHost(listenPort)
	if err != nil {
		log.Fatal(err)
	}
	memTransactions := make([]chain.Transaction, 0)
	stream := chain.PeerStream{Host: host, MemTransactions: memTransactions}

	if targetPeer == "" {
		log.Println("listening for connections")
		host.SetStreamHandler("/p2p/1.0.0", stream.HandleStream)
	} else {
		host.SetStreamHandler("/p2p/1.0.0", stream.HandleStream)
		ipfsAddr, err := ma.NewMultiaddr(targetPeer)
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

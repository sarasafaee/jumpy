package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	golog "github.com/ipfs/go-log"
	net "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	_ "github.com/libp2p/go-libp2p/p2p/host/peerstore"
	ma "github.com/multiformats/go-multiaddr"
	gologging "github.com/whyrusleeping/go-logging"
	"log"
	"math"
	"myBlockchain/chain"
	"myBlockchain/http"
	"sync"
)

// Blockchain is a series of validated Blocks

var mutex sync.Mutex
var rw *bufio.ReadWriter

/*var rw bufio.ReadWriter

func writeToNetwork(data []byte) {
	mutex.Lock()
	rw.WriteString(fmt.Sprintf("%s\n", string(data)))
	rw.Flush()
	mutex.Unlock()
}*/

func handleStream(s net.Stream) {
	log.Println("Got a new stream!")
	//rw = bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	readData(rw)
}

func readData(rw *bufio.ReadWriter) {

	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(str)
		/*
			if str == "" {
				return
			}


			if str != "\n" {

				chain := make([]Block, 0)
				if err := json.Unmarshal([]byte(str), &chain); err != nil {
					log.Fatal(err)
				}

				mutex.Lock()
				if len(chain) > len(Blockchain) {
					Blockchain = chain
					bytes, err := json.MarshalIndent(Blockchain, "", "  ")
					if err != nil {

						log.Fatal(err)
					}
					// Green console color: 	\x1b[32m
					// Reset console color: 	\x1b[0m
					fmt.Printf("\x1b[32m%s\x1b[0m> ", string(bytes))
				}
				mutex.Unlock()
			}
		*/
	}
}

func main() {
	//TODO: take x,y positions from CLI
	genesisBlock := chain.CreateGenesisBlock(0, 0)
	chain.Blockchain = append(chain.Blockchain, genesisBlock)

	// LibP2P code uses golog to log messages. They log with different
	// string IDs (i.e. "swarm"). We can control the verbosity level for
	// all loggers with:
	golog.SetAllLoggers(golog.LogLevel(gologging.INFO)) // Change to DEBUG for extra info

	// Parse options from the command line
	listenF := flag.Int("l", 0, "wait for incoming connections")
	muxPort := flag.Int("p", 0, "wait for incoming transactions")
	target := flag.String("d", "", "target peer to dial")
	secio := flag.Bool("secio", false, "enable secio")
	seed := flag.Int64("seed", 0, "set random seed for id generation")
	flag.Parse()

	if *listenF == 0 {
		log.Fatal("Please provide a port to bind on with -l")
	}

	// Make a host that listens on the given multiaddress
	host, err := chain.CreateHost(*listenF, *muxPort, *secio, *seed)
	if err != nil {
		log.Fatal(err)
	}

	if *target == "" {
		log.Println("listening for connections")
		host.SetStreamHandler("/p2p/1.0.0", handleStream)
	} else {

		// The following code extracts target's peer ID from the
		// given multiaddress
		ipfsaddr, err := ma.NewMultiaddr(*target)
		if err != nil {
			log.Fatalln(err)
		}

		pid, err := ipfsaddr.ValueForProtocol(ma.P_IPFS)
		if err != nil {
			log.Fatalln(err)
		}

		peerid, err := peer.Decode(pid)
		if err != nil {
			log.Fatalln(err)
		}

		// Decapsulate the /ipfs/<peerID> part from the target
		// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>
		targetPeerAddr, _ := ma.NewMultiaddr(
			fmt.Sprintf("/ipfs/%s", peerid.String()))
		targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)

		// We have a peer ID and a targetAddr so we add it to the peerstore
		// so LibP2P knows how to contact it
		// ha.Peerstore().AddAddr(peerid, targetAddr, pstore.PermanentAddrTTL)
		host.Peerstore().AddAddr(peerid, targetAddr, math.MaxInt64)
		log.Println("opening stream")
		// make a new stream from host B to host A
		// it should be handled on host A by the handler we set above because
		// we use the same /p2p/1.0.0 protocol
		s, err := host.NewStream(context.Background(), peerid, "/p2p/1.0.0")
		if err != nil {
			log.Fatalln(err)
		}
		// Create a buffered stream so that read and writes are non blocking.
		rw = bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
		host.SetStreamHandler("/p2p/1.0.0", handleStream)
	}

	httpServer := http.HttpServer{Host: host, RW: rw}
	if *muxPort != 0 {
		if err := httpServer.RunHttpServer(*muxPort); err != nil {
			log.Fatal(err)
		}
	}

	select {}
}

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
	"os"
	"strings"
	"sync"
)

var mutex sync.Mutex

func handleStream(s net.Stream) {
	log.Println("Got a new stream!")
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	go readData(rw)
	go handleCli(rw)
}

func readData(rw *bufio.ReadWriter) {

	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(fmt.Sprintf("Received -> %s", str))
	}
}

func handleCli(rw *bufio.ReadWriter) {
	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		transaction := strings.Replace(sendData, "\n", "", -1)
		//mutex.Lock()
		rw.WriteString(fmt.Sprintf("%s\n", transaction))
		rw.Flush()
		//mutex.Unlock()
	}
}

func main() {
	//TODO: take x,y positions from CLI
	genesisBlock := chain.CreateGenesisBlock(0, 0)
	chain.Blockchain = append(chain.Blockchain, genesisBlock)

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

	host, err := chain.CreateHost(*listenF, *muxPort, *secio, *seed)
	if err != nil {
		log.Fatal(err)
	}

	if *target == "" {
		log.Println("listening for connections")
		host.SetStreamHandler("/p2p/1.0.0", handleStream)
	} else {
		host.SetStreamHandler("/p2p/1.0.0", handleStream)
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

		targetPeerAddr, _ := ma.NewMultiaddr(
			fmt.Sprintf("/ipfs/%s", peerid.String()))
		targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)
		host.Peerstore().AddAddr(peerid, targetAddr, math.MaxInt64)
		log.Println("opening stream")

		s, err := host.NewStream(context.Background(), peerid, "/p2p/1.0.0")
		if err != nil {
			log.Fatalln(err)
		}
		rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

		/*mtx := &sync.Mutex{}
		httpServer := http.HttpServer{Host: host, RW: rw, Mutex: mtx}
		if *muxPort != 0 {
			if err := httpServer.RunHttpServer(*muxPort); err != nil {
				log.Fatal(err)
			}
		}*/

		go readData(rw)
		go handleCli(rw)
	}

	select {}
}

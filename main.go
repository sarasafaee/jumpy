package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	golog "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p/core/host"
	net "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	_ "github.com/libp2p/go-libp2p/p2p/host/peerstore"
	ma "github.com/multiformats/go-multiaddr"
	gologging "github.com/whyrusleeping/go-logging"
	"log"
	"math"
	mrand "math/rand"
	"myBlockchain/chain"
	"os"
	"strings"
	"sync"
)

var mutex sync.Mutex
var memTransactions []chain.Transaction

type PeerStream struct {
	host host.Host
}

func (ps PeerStream) handleStream(s net.Stream) {
	log.Println("Got a new stream!")
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	go readData(ps.host.ID().String(), rw)
	go handleCli(ps.host, rw)
}

func readData(hostId string, rw *bufio.ReadWriter) {

	for {
		str, err := rw.ReadString('#')
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(fmt.Sprintf("Received -> %s", str))
		message := strings.Split(str, ":")
		switch message[0] {
		case chain.PULL_BLOCK: //message = [command,receiver,sender]
			if message[1] == hostId {
				randomBlockHash := chain.GetRandomBlockHash()
				if len(message[2]) == 0 {
					fmt.Println(errors.New("sender ID is empty"))
				}
				senderID := strings.ReplaceAll(message[2], "#", "")
				msg := fmt.Sprintf("%s:%s:%s:%s#", chain.PUSH_BLOCK, randomBlockHash, senderID, hostId)
				rw.WriteString(msg)
				rw.Flush()
			}
		case chain.PUSH_BLOCK: //message = [command,randomBlockHash,receiver,sender]
			oldBlock := chain.Blockchain[len(chain.Blockchain)-1]
			if len(message[2]) == 0 {
				fmt.Println(errors.New("sender ID is empty"))
			}
			senderID := strings.ReplaceAll(message[2], "#", "")
			if senderID == hostId {
				chain.GenerateBlock(hostId, oldBlock, message[1], message[3], memTransactions)
			}
		}
	}
}

func handleCli(host host.Host, rw *bufio.ReadWriter) {
	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			log.Println(err)
			continue
		}

		fmt.Println(sendData)
		transaction := strings.Replace(sendData, "\n", "", -1)
		/*sbyte, err := json.Marshal(transaction)
		if err != nil {
			log.Println(err)
			continue
		}*/
		pos := chain.Position{}
		err = json.Unmarshal([]byte(transaction), &pos)
		if err != nil {
			log.Println(err)
			continue
		}
		memTransactions = append(memTransactions, chain.Transaction{
			Position: pos,
		})
		var randomPeer peer.ID
		for {
			randomPeer = host.Peerstore().Peers()[mrand.Intn(host.Peerstore().Peers().Len())]
			if randomPeer.String() == host.ID().String() {
				continue
			}
			break
		}

		//mutex.Lock()
		rw.WriteString(fmt.Sprintf("%s:%s:%s#", chain.PULL_BLOCK, randomPeer, host.ID().String()))
		rw.Flush()
		//mutex.Unlock()
	}
}

func main() {
	//TODO: take x,y positions from CLI
	genesisBlock := chain.CreateGenesisBlock(0, 0)
	chain.Blockchain = append(chain.Blockchain, genesisBlock)
	memTransactions = make([]chain.Transaction, 10)
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
	stream := PeerStream{host: host}

	if *target == "" {
		log.Println("listening for connections")
		host.SetStreamHandler("/p2p/1.0.0", stream.handleStream)
	} else {
		host.SetStreamHandler("/p2p/1.0.0", stream.handleStream)
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
		go readData(host.ID().String(), rw)
		go handleCli(host, rw)
	}

	select {}
}

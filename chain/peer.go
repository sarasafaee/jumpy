package chain

import (
	"bufio"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	net "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"io"
	"log"
	mrand "math/rand"
	"os"
	"strings"
)

type PeerStream struct {
	Host            host.Host
	MemTransactions []Transaction
}

func (ps *PeerStream) HandleStream(s net.Stream) {
	log.Println("Got a new stream!")
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	go ps.ReadStream(rw)
}

func (ps *PeerStream) ReadStream(rw *bufio.ReadWriter) {

	for {
		str, err := rw.ReadString('#')
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(fmt.Sprintf("Received -> %s", str))
		message := strings.Split(str, ":")
		switch message[0] {
		case PULL_BLOCK: //message = [command,receiver,sender]
			if message[1] == ps.Host.ID().String() {
				randomBlockHash := GetRandomBlockHash()
				if len(message[2]) == 0 {
					fmt.Println(errors.New("sender ID is empty"))
					continue
				}
				senderID := strings.ReplaceAll(message[2], "#", "")
				msg := fmt.Sprintf("%s:%s:%s:%s#", PUSH_BLOCK, randomBlockHash, senderID, ps.Host.ID().String())
				if _, err = rw.WriteString(msg); err != nil {
					fmt.Println(err)
					continue
				}
				if err = rw.Flush(); err != nil {
					fmt.Println(err)
					continue
				}
			}
		case PUSH_BLOCK: //message = [command,randomBlockHash,receiver,sender]
			oldBlock := Blockchain[len(Blockchain)-1]
			if len(message[2]) == 0 {
				fmt.Println(errors.New("sender ID is empty"))
				continue
			}

			senderID := strings.ReplaceAll(message[2], "#", "")
			if senderID == ps.Host.ID().String() {
				b := GenerateBlock(ps.Host.ID().String(), oldBlock, message[1], message[3], ps.MemTransactions)
				Blockchain = append(Blockchain, b)
				ps.MemTransactions = make([]Transaction, 0)
			} else {
				fmt.Println(errors.New("sender ID is not equal to my ID"))
				continue
			}

		default:
			continue
		}
	}
}

func (ps *PeerStream) HandleCli(rw *bufio.ReadWriter) {
	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			log.Println(err)
			continue
		}
		sendData = strings.Replace(sendData, "\n", "", -1)

		if sendData == "log" {
			printBlockChain()
			continue
		}

		pos := Position{}
		err = json.Unmarshal([]byte(sendData), &pos)
		if err != nil {
			log.Println(err)
			continue
		}
		ps.MemTransactions = append(ps.MemTransactions, Transaction{
			Position: pos,
		})
		var randomPeer peer.ID
		for {
			randomPeer = ps.Host.Peerstore().Peers()[mrand.Intn(ps.Host.Peerstore().Peers().Len())]
			if randomPeer.String() == ps.Host.ID().String() {
				continue
			}
			break
		}

		if _, err = rw.WriteString(fmt.Sprintf("%s:%s:%s#", PULL_BLOCK, randomPeer, ps.Host.ID().String())); err != nil {
			log.Println(err)
			continue
		}
		if err = rw.Flush(); err != nil {
			log.Println(err)
			continue
		}
	}
}

func CreateHost(listenPort int, muxPort int, secio bool, randseed int64) (host.Host, error) {
	// If the seed is zero, use real cryptographic randomness. Otherwise, use a
	// deterministic randomness source to make generated keys stay the same
	// across multiple runs
	var r io.Reader
	if randseed == 0 {
		r = rand.Reader
	} else {
		r = mrand.New(mrand.NewSource(randseed))
	}

	// Generate a key pair for this host. We will use it
	// to obtain a valid host ID.
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		return nil, err
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", listenPort)),
		libp2p.Identity(priv),
	}

	host, err := libp2p.New(opts...)
	if err != nil {
		return nil, err
	}

	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", host.ID()))

	// Now we can build a full multiaddress to reach this host
	// by encapsulating both addresses:
	addrs := host.Addrs()
	var addr ma.Multiaddr
	// select the address starting with "ip4"
	for _, i := range addrs {
		if strings.HasPrefix(i.String(), "/ip4") {
			addr = i
			break
		}
	}
	fullAddr := addr.Encapsulate(hostAddr)
	log.Printf("I am %s\n", fullAddr)
	if secio {
		log.Printf("Now run \"go run main.go -l %d -p %d -d %s -secio\" on a different terminal\n", listenPort+1, muxPort+1, fullAddr)
	} else {
		log.Printf("Now run \"go run main.go -l %d -p %d -d %s\" on a different terminal\n", listenPort+1, muxPort+1, fullAddr)
	}

	return host, nil
}

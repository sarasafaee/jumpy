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
		fmt.Printf("Received -> %s\n", str)
		receivedMessage := strings.Split(str, ":")
		switch receivedMessage[0] {
		case PULL_BLOCK: //message = [command,receiver,sender]
			if receivedMessage[1] == ps.Host.ID().String() {
				if len(receivedMessage[2]) == 0 {
					fmt.Println(errors.New("sender ID is empty"))
					continue
				}
				senderID := strings.ReplaceAll(receivedMessage[2], "#", "")

				lastBlock := GetLastBlock()
				if lastBlock == nil {
					fmt.Println(errors.New("no block founded in chain"))
					continue
				}
				message := fmt.Sprintf("%s:%s:%s:%s#", PUSH_BLOCK, lastBlock.Hash, senderID, ps.Host.ID().String())
				if err = ps.writeStringToStream(rw, message); err != nil {
					log.Println(err)
					continue
				}
			}
		case PUSH_BLOCK: //message = [command,lastBlockHash,receiver,sender]
			lastBlock := GetLastBlock()
			if len(receivedMessage[2]) == 0 {
				fmt.Println(errors.New("sender ID is empty"))
				continue
			}

			senderID := strings.ReplaceAll(receivedMessage[2], "#", "")
			if senderID == ps.Host.ID().String() {
				blockNode := strings.ReplaceAll(receivedMessage[3], "#", "")
				blockHash := receivedMessage[1]
				b := GenerateBlock(ps.Host.ID().String(), lastBlock, blockNode, blockHash, ps.MemTransactions)
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
		command, err := stdReader.ReadString('\n')
		if err != nil {
			log.Println(err)
			continue
		}
		command = strings.Replace(command, "\n", "", -1)

		if command == "log" {
			printBlockChain()
			continue
		}

		pos := Position{}
		err = json.Unmarshal([]byte(command), &pos)
		if err != nil {
			log.Println(err)
			continue
		}
		ps.MemTransactions = append(ps.MemTransactions, Transaction{
			Position: pos,
		})

		randomPeer := ps.getRandomPeer()
		message := fmt.Sprintf("%s:%s:%s#", PULL_BLOCK, randomPeer, ps.Host.ID().String())
		if err = ps.writeStringToStream(rw, message); err != nil {
			log.Println(err)
			continue
		}
	}
}

func (ps *PeerStream) writeStringToStream(rw *bufio.ReadWriter, message string) error {
	if _, err := rw.WriteString(message); err != nil {
		return err
	}
	return rw.Flush()
}

func (ps *PeerStream) getRandomPeer() peer.ID {
	var randomPeer peer.ID
	for {
		peersLen := ps.Host.Peerstore().Peers().Len()
		randomIndex := mrand.Intn(peersLen)
		randomPeer = ps.Host.Peerstore().Peers()[randomIndex]
		if randomPeer.String() == ps.Host.ID().String() {
			continue
		}
		break
	}

	return randomPeer
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

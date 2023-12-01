package p2p

import (
	"bufio"
	"context"
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
	"log"
	mrand "math/rand"
	"myBlockchain/chain"
	"os"
	"strings"
)

type PeerStream struct {
	Host            host.Host
	MemTransactions []chain.Transaction
}

func (ps *PeerStream) HandleStream(s net.Stream) {
	//log.Println("connected to: ", s.)
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
		case chain.PULL_BLOCK: //message = [command,receiver,sender]
			if receivedMessage[1] == ps.Host.ID().String() {
				if len(receivedMessage[2]) == 0 {
					fmt.Println(errors.New("sender ID is empty"))
					continue
				}
				senderID := strings.ReplaceAll(receivedMessage[2], "#", "")

				lastBlock := chain.GetLastBlock()
				if lastBlock == nil {
					fmt.Println(errors.New("no block founded in chain"))
					continue
				}
				message := fmt.Sprintf("%s:%s:%s:%s#", chain.PUSH_BLOCK, lastBlock.Hash, senderID, ps.Host.ID().String())
				if err = writeStringToStream(rw, message); err != nil {
					log.Println(err)
					continue
				}
			}
		case chain.PUSH_BLOCK: //message = [command,lastBlockHash,receiver,sender]
			lastBlock := chain.GetLastBlock()
			if len(receivedMessage[2]) == 0 {
				fmt.Println(errors.New("sender ID is empty"))
				continue
			}

			senderID := strings.ReplaceAll(receivedMessage[2], "#", "")
			if senderID == ps.Host.ID().String() {
				blockNode := strings.ReplaceAll(receivedMessage[3], "#", "")
				blockHash := receivedMessage[1]
				b := chain.GenerateBlock(ps.Host.ID().String(), lastBlock, blockNode, blockHash, ps.MemTransactions)
				chain.Blockchain = append(chain.Blockchain, b)
				ps.MemTransactions = make([]chain.Transaction, 0)
			} else {
				fmt.Println(errors.New("sender ID is not equal to my ID"))
			}
			continue
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
			chain.PrintBlockChain()
			continue
		}

		pos := chain.Position{}
		err = json.Unmarshal([]byte(command), &pos)
		if err != nil {
			log.Println(err)
			continue
		}
		ps.MemTransactions = append(ps.MemTransactions, chain.Transaction{
			Position: pos,
		})

		randomPeer := ps.getRandomPeer()
		message := fmt.Sprintf("%s:%s:%s#", chain.PULL_BLOCK, randomPeer, ps.Host.ID().String())
		if err = writeStringToStream(rw, message); err != nil {
			log.Println(err)
			continue
		}
	}
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

func (ps *PeerStream) getPeerFullAddr() ma.Multiaddr {
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", ps.Host.ID()))

	addrs := ps.Host.Addrs()
	var addr ma.Multiaddr
	for _, i := range addrs {
		if strings.HasPrefix(i.String(), "/ip4") {
			addr = i
			break
		}
	}
	return addr.Encapsulate(hostAddr)
}

func Run(ctx context.Context, listenPort int, chainGroupName string) {
	h, err := createHost(listenPort)
	if err != nil {
		log.Fatal(err)
	}
	memTransactions := make([]chain.Transaction, 0)
	stream := &PeerStream{Host: h, MemTransactions: memTransactions}
	peerAddr := stream.getPeerFullAddr()
	log.Printf("my address: %s\n", peerAddr)

	// connect to other peers
	h.SetStreamHandler("/p2p/1.0.0", stream.HandleStream)
	log.Println("listening for connections")
	peerChan := InitMDNS(h, chainGroupName)
	go func(ctx context.Context, stream *PeerStream) {
		for {
			peer := <-peerChan
			if err := stream.Host.Connect(ctx, peer); err != nil {
				fmt.Println("connection failed:", err)
				continue
			}

			fmt.Println("connected to: ", peer)
			s, err := stream.Host.NewStream(ctx, peer.ID, "/p2p/1.0.0")
			if err != nil {
				fmt.Println("stream open failed", err)
			} else {
				rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
				go stream.ReadStream(rw)
				go stream.HandleCli(rw)
			}
		}
	}(ctx, stream)
}

func createHost(listenPort int) (host.Host, error) {

	r := rand.Reader
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		return nil, err
	}

	sourceMultiAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", listenPort))
	opts := []libp2p.Option{
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(priv),
	}

	host, err := libp2p.New(opts...)
	if err != nil {
		return nil, err
	}
	return host, nil
}

func writeStringToStream(rw *bufio.ReadWriter, message string) error {
	if _, err := rw.WriteString(message); err != nil {
		return err
	}
	return rw.Flush()
}

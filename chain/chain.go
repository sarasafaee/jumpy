package chain

import (
	"bufio"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	ma "github.com/multiformats/go-multiaddr"
	"io"
	"log"
	mrand "math/rand"
	"strings"
)

const (
	PULL_BLOCK = "pull_block\n"
	PUSH_BLOCK = "push_block"
)

var Blockchain []Block

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

func PullBlockConnection(rw *bufio.ReadWriter) {

	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		if str == "" {
			return
		}
		if str != "\n" {

			chain := make([]Block, 0)
			if err := json.Unmarshal([]byte(str), &chain); err != nil {
				log.Fatal(err)
			}

			if len(chain) > len(Blockchain) {
				Blockchain = chain
				bytes, err := json.MarshalIndent(Blockchain, "", "  ")
				if err != nil {

					log.Fatal(err)
				}
				// Green console color: 	\x1b[32m
				// Reset console color: 	\x1b[0m
				fmt.Printf("\x1b[32m%s\x1b[0m> ", PULL_BLOCK, string(bytes))
			}
		}
	}
}

func PushBlockConnection(rw *bufio.ReadWriter) {

	/*go func() {
		time.Sleep(5 * time.Second)
		mutex.Lock()
		bytes, err := json.Marshal(Blockchain)
		if err != nil {
			log.Println(err)
		}
		mutex.Unlock()

		mutex.Lock()
		rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
		rw.Flush()
		mutex.Unlock()
	}()*/

	/*stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		sendData = strings.Replace(sendData, "\n", "", -1)
		bpm, err := strconv.Atoi(sendData)
		if err != nil {
			log.Fatal(err)
		}
		newBlock := generateBlock(Blockchain[len(Blockchain)-1], bpm)
		mutex.Lock()
		Blockchain = append(Blockchain, newBlock)
		mutex.Unlock()
	*/
	/*if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {

	}*/
	/*
				bytes, err := json.Marshal(Blockchain)
				if err != nil {
					log.Println(err)
				}

			spew.Dump(Blockchain)

		}
	*/
}

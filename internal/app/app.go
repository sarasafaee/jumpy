package app

import (
	"context"
	_ "github.com/libp2p/go-libp2p/p2p/host/peerstore"
	"myBlockchain/chain"
	"myBlockchain/p2p"
)

const hostGroupName = "jumpy"

func Start(listenPort int) {
	ctx := context.Background()

	//init genesis block
	genesisBlock := chain.CreateGenesisBlock(0, 0)
	chain.Blockchain = append(chain.Blockchain, genesisBlock)

	p2p.Run(ctx, listenPort, hostGroupName)
	select {}
}

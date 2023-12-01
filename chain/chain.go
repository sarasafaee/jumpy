package chain

import (
	"fmt"
	mrand "math/rand"
)

const (
	SuccessColor = "\033[1;32m%s\033[0m"
	ErrorColor   = "\033[1;31m%s\033[0m"
	InfoColor    = "\033[1;34m%s\033[0m"
	WarningColor = "\033[1;33m%s\033[0m"
	NoticeColor  = "\033[1;36m%s\033[0m"
	DebugColor   = "\033[0;36m%s\033[0m"
)

const (
	PULL_BLOCK = "pull_block"
	PUSH_BLOCK = "push_block"
)

var Blockchain []Block

func GetLastBlock() *Block {
	return &Blockchain[len(Blockchain)-1]
}

func GetRandomBlock() *Block {
	return &Blockchain[mrand.Intn(len(Blockchain))]
}

func PrintBlockChain() {
	for _, b := range Blockchain {
		fmt.Println("-----------------------------------------------------------------------")
		fmt.Println(fmt.Sprintf("Index: %d\nHash:%s", b.Index, b.Hash))
		fmt.Printf(SuccessColor, "Transactions:\n")
		for i, t := range b.Transaction {
			fmt.Println(fmt.Sprintf("%d: X = %d Y = %d", i, t.Position.X, t.Position.Y))
		}
		fmt.Printf(ErrorColor, "Conenctions:\n")
		for _, c := range b.Connections {
			fmt.Println()
			fmt.Println(fmt.Sprintf("Node = %s", c.PeerID))
			fmt.Println(fmt.Sprintf("Block = %s", c.BlockHash))
		}
	}
}

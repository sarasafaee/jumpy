package chain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Transaction struct {
	Position Position `json:"position"`
}

type BlockConnection struct {
	NodePublicKey string
	BlockHash     string
}

type Block struct {
	Index       int
	Timestamp   string
	Transaction []Transaction
	Hash        string
	Connections []BlockConnection
}

func CreateGenesisBlock(xPos, yPos int) Block {
	t := time.Now()
	genesisBlock := Block{}
	connections := make([]BlockConnection, 2)
	return Block{0, t.String(), []Transaction{{Position: Position{xPos, yPos}}}, genesisBlock.calculateHash(), connections}
}

func (b Block) calculateHash() string {
	record := b.toString()
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

func (b Block) toString() string {
	transactionStr := ""
	for _, t := range b.Transaction {
		transactionStr = fmt.Sprintf("%s%s", transactionStr, t.toString())
	}

	connectionStr := ""
	for _, c := range b.Connections {
		connectionStr = fmt.Sprintf("%s%s%s", connectionStr, c.NodePublicKey, c.BlockHash)
	}

	return fmt.Sprintf("%d%s%s%s", b.Index, b.Timestamp, transactionStr, connectionStr)
}

// GenerateBlock will create a new block using previous block's hash
func GenerateBlock(myPublicKey string, lastBlock *Block, targetBlockNode, targetBlockHash string, transaction []Transaction) Block {

	var newBlock Block

	newBlock.Index = lastBlock.Index + 1
	newBlock.Transaction = transaction
	//add last block hash to connections
	connections := make([]BlockConnection, 0)
	connections = append(connections, BlockConnection{NodePublicKey: myPublicKey, BlockHash: lastBlock.Hash})
	connections = append(connections, BlockConnection{NodePublicKey: targetBlockNode, BlockHash: targetBlockHash})
	newBlock.Connections = connections
	newBlock.Hash = newBlock.calculateHash()

	t := time.Now()
	newBlock.Timestamp = t.String()

	return newBlock
}

func (t Transaction) toString() string {
	return fmt.Sprintf("%d,%d", t.Position.X, t.Position.Y)
}

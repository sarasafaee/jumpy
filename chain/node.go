package chain

import (
	"time"
)

var Node *NodeMetaData

type NodeMetaData struct {
	PublicKey string
}

// GenerateBlock will create a new block using previous block's hash
func (n *NodeMetaData) GenerateBlock(oldBlock, targetBlock Block, transaction Transaction) Block {

	var newBlock Block
	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.Transaction = transaction
	connections := make([]BlockConnection, 2)
	//add last block hash to connections
	connections = append(connections, BlockConnection{NodePublicKey: n.PublicKey, BlockHash: oldBlock.Hash})
	//TODO: get random block
	newBlock.Connections = connections
	newBlock.Hash = newBlock.calculateHash()

	return newBlock
}

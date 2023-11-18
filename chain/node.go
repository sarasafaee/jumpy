package chain

import (
	"time"
)

// GenerateBlock will create a new block using previous block's hash
func GenerateBlock(myPublicKey string, oldBlock Block, targetBlockHash, targetBlockNode string, transaction []Transaction) Block {

	var newBlock Block
	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.Transaction = transaction
	connections := make([]BlockConnection, 2)
	//add last block hash to connections
	connections = append(connections, BlockConnection{NodePublicKey: myPublicKey, BlockHash: oldBlock.Hash})
	connections = append(connections, BlockConnection{NodePublicKey: targetBlockNode, BlockHash: targetBlockHash})
	newBlock.Connections = connections
	newBlock.Hash = newBlock.calculateHash()

	return newBlock
}

package cfxMysql

import (
	"github.com/openweb3/web3go/types"
)

type Block struct {
	ID         uint   `gorm:"primaryKey"`
	Author     string `json:"author,omitempty"` // Address as string
	Hash       string `json:"hash"`             // Hash as string
	ParentHash string `json:"parentHash"`       // Hash as string
	Timestamp  uint64 `json:"timestamp"`
}

func ConvertBlock(block *types.Block) Block {
	return Block{
		ID:         uint(block.Number.Uint64()),
		Author:     block.Author.Hex(),
		Hash:       block.Hash.Hex(),
		ParentHash: block.ParentHash.Hex(),
		Timestamp:  block.Timestamp,
	}
}

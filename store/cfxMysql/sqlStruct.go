package cfxMysql

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/openweb3/web3go/types"
)

type Block struct {
	ID         uint   `gorm:"primaryKey"`
	AuthorID   uint   `json:"author,omitempty"` // Address as string
	Hash       string `json:"hash"`             // Hash as string
	ParentHash string `json:"parentHash"`       // Hash as string
	Timestamp  uint64 `json:"timestamp"`
}

func ConvertBlockWithoutAuthor(block *types.Block) Block {
	return Block{
		ID:         uint(block.Number.Uint64()),
		Hash:       block.Hash.Hex(),
		ParentHash: block.ParentHash.Hex(),
		Timestamp:  block.Timestamp,
	}
}

func ConvertAuthorToString(author *common.Address) string {
	return author.Hex()
}

type Author struct {
	ID     uint   `gorm:"primaryKey"`
	Author string `gorm:"index" json:"author,omitempty"`
}

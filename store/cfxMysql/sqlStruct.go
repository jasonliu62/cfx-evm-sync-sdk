package cfxMysql

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/openweb3/web3go/types"
)

type Block struct {
	BlockNumber uint   `gorm:"primaryKey"`
	AuthorID    uint   `json:"author,omitempty"` // Address from address table as uint
	Hash        string `json:"hash"`             // Hash as string
	ParentHash  string `json:"parentHash"`       // Hash as string
	Timestamp   uint64 `json:"timestamp"`
}

type Address struct {
	ID      uint   `gorm:"primaryKey"`
	Address string `gorm:"index" json:"author,omitempty"`
}

type TransactionDetail struct {
	ID               uint   `gorm:"primaryKey"`
	BlockNumber      uint   `json:"block_number"`
	TransactionIndex uint   `json:"transaction_index"`
	TxHash           string `json:"tx_hash"`
	FromAddress      uint   `json:"from_address"` // Address from address table as uint
	ToAddress        uint   `json:"to_address"`   // Address from address table as uint
	Value            int64  `json:"value"`
	Gas              int64  `json:"price"`
	GasPrice         int64  `json:"gas_price"`
	Nonce            int64  `json:"nonce"`
	Input            string `json:"input"`
}

type Hash struct {
	ID   uint   `gorm:"primaryKey"`
	Hash string `gorm:"index" json:"tx_hash"`
}

func ConvertBlockWithoutAuthor(block *types.Block) Block {
	return Block{
		BlockNumber: uint(block.Number.Uint64()),
		Hash:        block.Hash.Hex(),
		ParentHash:  block.ParentHash.Hex(),
		Timestamp:   block.Timestamp,
	}
}

func ConvertTransactionDetail(index uint, tx *types.TransactionDetail) TransactionDetail {
	return TransactionDetail{
		BlockNumber:      uint(tx.BlockNumber.Uint64()),
		TransactionIndex: index,
		TxHash:           tx.Hash.Hex(),
		Value:            int64(tx.Value.Uint64()),
		Gas:              int64(tx.Gas),
		GasPrice:         int64(tx.GasPrice.Uint64()),
		Nonce:            int64(tx.Nonce),
		Input:            string(tx.Input),
	}
}

type BlockDataMySQL struct {
	Block              Block
	TransactionDetails []TransactionDetail
}

func ConvertAddressToString(author *common.Address) string {
	return author.Hex()
}

func ConvertHashToString(hash common.Hash) string {
	return hash.Hex()
}

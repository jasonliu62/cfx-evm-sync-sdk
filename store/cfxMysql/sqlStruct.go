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
	ID               uint   `gorm:"primaryKey;autoIncrement"`
	BlockNumber      uint   `json:"block_number"`
	TransactionIndex uint   `json:"transaction_index"`
	TxHash           string `json:"tx_hash"`
	FromAddress      uint   `json:"from_address"` // Address from address table as uint
	ToAddress        uint   `json:"to_address"`   // Address from address table as uint
	Value            int64  `json:"value"`
	Gas              int64  `json:"price"`
	GasPrice         int64  `json:"gas_price"`
	Nonce            int64  `json:"nonce"`
	Input            []byte `json:"input"`
}

type Hash struct {
	ID   uint   `gorm:"primaryKey"`
	Hash string `gorm:"index" json:"tx_hash"`
}

type Log struct {
	ID          uint   `gorm:"primaryKey;autoIncrement"`
	Address     uint   `json:"address"`
	BlockNumber uint64 `json:"block_number"`
	Data        []byte `json:"data"`
	LogIndex    uint   `json:"log_index"`
	Topic0      uint   `json:"topics0"`
	Topic1      string `json:"topics1"`
	Topic2      string `json:"topics2"`
	Topic3      string `json:"topics3"`
	TxIndex     uint   `json:"transactionIndex"`
}

type Erc20Transfer struct {
	ID          uint   `gorm:"primaryKey;autoIncrement"`
	Address     uint   `json:"address"`
	BlockNumber uint64 `json:"block_number"`
	LogIndex    uint   `json:"log_index"`
	Src         uint   `json:"topics1"`
	Dst         uint   `json:"topics2"`
	Wad         string `json:"data[0]"`
	TxIndex     uint   `json:"transactionIndex"`
}

func ConvertLogWithoutTopic(log *types.Log) Log {
	return Log{
		BlockNumber: log.BlockNumber,
		Data:        log.Data,
		LogIndex:    log.Index,
		TxIndex:     log.TxIndex,
	}
}

func ConvertErc20Transfer(log *types.Log) Erc20Transfer {
	return Erc20Transfer{
		BlockNumber: log.BlockNumber,
		LogIndex:    log.Index,
		Wad:         string(log.Data[0]),
		TxIndex:     log.TxIndex,
	}
}

func ConvertLogTopics(log Log, topics []common.Hash) Log {
	if len(topics) > 1 {
		log.Topic1 = topics[1].Hex()
	}
	if len(topics) > 2 {
		log.Topic2 = topics[2].Hex()
	}
	if len(topics) > 3 {
		log.Topic3 = topics[3].Hex()
	}
	return log
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
		Input:            tx.Input,
	}
}

type BlockDataMySQL struct {
	Block              Block
	TransactionDetails []TransactionDetail
	Logs               []Log
	Erc20Transfers     []Erc20Transfer
}

func ConvertAddressToString(author *common.Address) string {
	return author.Hex()
}

func ConvertHashToString(hash common.Hash) string {
	return hash.Hex()
}

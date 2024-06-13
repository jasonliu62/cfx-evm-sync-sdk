package data

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/openweb3/web3go"
	"github.com/openweb3/web3go/types"
)

type GetFunc func(*web3go.Client, BlockNumberOrHash) (interface{}, error)

type DataWrap struct {
	Error error
	Value interface{}
}

type BlockNumberOrHash struct {
	BlockNumber uint64
	Hash        common.Hash
}

type BlockData struct {
	Block              *types.Block
	TransactionDetails []*types.TransactionDetail
	Logs               []*types.Log
}

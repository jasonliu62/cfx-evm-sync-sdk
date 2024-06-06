package data

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/openweb3/web3go"
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

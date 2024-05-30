package blockNumberBiz

import (
	"cfx-evm-sync-sdk/sync/blockNumberSync"
	"github.com/openweb3/web3go"
	"github.com/openweb3/web3go/types"
)

// Support:
// BlockByNumber: *types.Block
// BlockReceipts: []*types.Receipt
// BlockTransactionCountByNumber: *big.Int
// BlockUnclesCountByNumber: *big.Int

type GetFunc func(*web3go.Client, uint64) (interface{}, error)

func BlockByNumber(nodes []string, startBlock, endBlock uint64) {
	GetFunc := func(w3client *web3go.Client, blockNumber uint64) (interface{}, error) {
		return w3client.Eth.BlockByNumber(types.BlockNumber(blockNumber), false)
	}
	blockNumberSync.InitConcurrentGet(nodes, startBlock, endBlock, GetFunc)
}

func MultiTask(nodes []string, startBlock, endBlock uint64) {
	funcs := []GetFunc{
		func(w3client *web3go.Client, blockNumber uint64) (interface{}, error) {
			return w3client.Eth.BlockByNumber(types.BlockNumber(blockNumber), false)
		},
		func(w3client *web3go.Client, blockNumber uint64) (interface{}, error) {
			blk := types.BlockNumberOrHashWithNumber(types.BlockNumber(blockNumber))
			return w3client.Eth.BlockReceipts(&blk)
		},
		func(w3client *web3go.Client, blockNumber uint64) (interface{}, error) {
			return w3client.Eth.BlockTransactionCountByNumber(types.BlockNumber(blockNumber))
		},
	}

	blockNumberSync.InitMultiTask(nodes, startBlock, endBlock, funcs)
}

func BlockReceipts(nodes []string, startBlock, endBlock uint64) {
	GetFunc := func(w3client *web3go.Client, blockNumber uint64) (interface{}, error) {
		blk := types.BlockNumberOrHashWithNumber(types.BlockNumber(blockNumber))
		return w3client.Eth.BlockReceipts(&blk)
	}
	blockNumberSync.InitConcurrentGet(nodes, startBlock, endBlock, GetFunc)
}

func BlockTransactionCountByNumber(nodes []string, startBlock, endBlock uint64) {
	GetFunc := func(w3client *web3go.Client, blockNumber uint64) (interface{}, error) {
		return w3client.Eth.BlockTransactionCountByNumber(types.BlockNumber(blockNumber))
	}
	blockNumberSync.InitConcurrentGet(nodes, startBlock, endBlock, GetFunc)
}

//func BlockUnclesCountByNumber(nodes []string, startBlock, endBlock uint64) {
//	blockNumberSync.InitConcurrentGet(nodes, startBlock, endBlock, "BlockUnclesCountByNumber")
//}

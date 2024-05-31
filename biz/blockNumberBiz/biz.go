package blockNumberBiz

import (
	"cfx-evm-sync-sdk/sync/blockNumberSync"
	"github.com/ethereum/go-ethereum/common"
	"github.com/openweb3/web3go"
	"github.com/openweb3/web3go/types"
	"github.com/spf13/viper"
)

// Support:
// BlockByNumber: *types.Block
// BlockReceipts: []*types.Receipt
// BlockTransactionCountByNumber: *big.Int
// BlockUnclesCountByNumber: *big.Int

type GetFunc func(*web3go.Client, interface{}) (interface{}, error)

func BlockByNumber(nodes []string, startBlock, endBlock uint64) {
	// init sdk in sync.gp
	var result []blockNumberSync.DataWrap
	s := blockNumberSync.NewSdk(nodes, viper.GetUint64("preload.count"), viper.GetUint64("preload.min"), result)
	GetFunc := func(w3client *web3go.Client, blockNumber interface{}) (interface{}, error) {
		blockNumberUint := blockNumber.(uint64)
		block, err := w3client.Eth.BlockByNumber(types.BlockNumber(blockNumberUint), false)
		return block.Hash, err
	}
	s.InitConcurrentGet(startBlock, endBlock, GetFunc)
	var result02 []blockNumberSync.DataWrap
	s = blockNumberSync.NewSdk(nodes, viper.GetUint64("preload.count"), viper.GetUint64("preload.min"), result02)
	GetFunc = func(w3client *web3go.Client, blockHash interface{}) (interface{}, error) {
		blockHashHash := blockHash.(common.Hash)
		blk := types.BlockNumberOrHashWithHash(blockHashHash, false)
		return w3client.Eth.BlockReceipts(&blk)
	}
	s.InitConcurrentGet(result[0].Value.(uint64), result[len(result)-1].Value.(uint64), GetFunc)
}

//func BlockReceipts(nodes []string, startBlock, endBlock uint64) {
//	GetFunc := func(w3client *web3go.Client, blockNumber uint64) (interface{}, error) {
//		blk := types.BlockNumberOrHashWithNumber(types.BlockNumber(blockNumber))
//		return w3client.Eth.BlockReceipts(&blk)
//	}
//	blockNumberSync.InitConcurrentGet(nodes, startBlock, endBlock, GetFunc)
//}
//
//func BlockTransactionCountByNumber(nodes []string, startBlock, endBlock uint64) {
//	GetFunc := func(w3client *web3go.Client, blockNumber uint64) (interface{}, error) {
//		return w3client.Eth.BlockTransactionCountByNumber(types.BlockNumber(blockNumber))
//	}
//	blockNumberSync.InitConcurrentGet(nodes, startBlock, endBlock, GetFunc)
//}

//func BlockUnclesCountByNumber(nodes []string, startBlock, endBlock uint64) {
//	blockNumberSync.InitConcurrentGet(nodes, startBlock, endBlock, "BlockUnclesCountByNumber")
//}

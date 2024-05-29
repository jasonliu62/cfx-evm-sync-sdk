package blockNumberBiz

import "cfx-evm-sync-sdk/sync/blockNumberSync"

// Support:
// BlockByNumber: *types.Block
// BlockReceipts: []*types.Receipt
// BlockTransactionCountByNumber: *big.Int
// BlockUnclesCountByNumber: *big.Int

func BlockByNumber(nodes []string, startBlock, endBlock uint64) {
	blockNumberSync.InitConcurrentGet(nodes, startBlock, endBlock, "BlockByNumber")
}

func BlockReceipts(nodes []string, startBlock, endBlock uint64) {
	blockNumberSync.InitConcurrentGet(nodes, startBlock, endBlock, "BlockReceipts")
}

func BlockTransactionCountByNumber(nodes []string, startBlock, endBlock uint64) {
	blockNumberSync.InitConcurrentGet(nodes, startBlock, endBlock, "BlockTransactionCountByNumber")
}

func BlockUnclesCountByNumber(nodes []string, startBlock, endBlock uint64) {
	blockNumberSync.InitConcurrentGet(nodes, startBlock, endBlock, "BlockUnclesCountByNumber")
}

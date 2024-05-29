package blockNumberSync

import (
	"cfx-evm-sync-sdk/rpc"
	"fmt"
	"github.com/openweb3/web3go"
	"github.com/openweb3/web3go/types"
	"github.com/spf13/viper"
	"log"
	"math/big"
	"sync"
)

type Data struct {
	block        *types.Block
	receipt      []*types.Receipt
	transactions *big.Int
	uncleCounts  *big.Int
}

var poolMutex sync.Mutex
var dataPool map[uint64]Data
var nextNum uint64

// TODO: 增加一个缓存错误blockNumber用于重发的池子
// 具体什么时候发送？

func initDataPool() {
	dataPool = make(map[uint64]Data)
}

func addBlockToDataPool(id uint64, block *types.Block) {
	dataPool[id] = Data{block: block}
}

func getBlock(w3client *web3go.Client, blockNumber uint64, nodeUrl string) {
	block, err := w3client.Eth.BlockByNumber(types.BlockNumber(blockNumber), false)
	// TODO: 错误处理
	if err != nil {
		log.Printf("Failed to get block %d from %s: %v", blockNumber, nodeUrl, err)
		return
	}
	fmt.Printf("We have block %d.\n", blockNumber)
	poolMutex.Lock()
	addBlockToDataPool(blockNumber, block)
	poolMutex.Unlock()
}

func addReceiptToDataPool(id uint64, receipt []*types.Receipt) {
	dataPool[id] = Data{receipt: receipt}
}

func getReceipt(w3client *web3go.Client, blockNumber uint64, nodeUrl string) {
	blk := types.BlockNumberOrHashWithNumber(types.BlockNumber(blockNumber))
	receipt, err := w3client.Eth.BlockReceipts(&blk)
	// TODO: 错误处理
	if err != nil {
		log.Printf("Failed to get receipt %d from %s: %v", blockNumber, nodeUrl, err)
		return
	}
	fmt.Printf("We have receipt from block %d.\n", blockNumber)
	poolMutex.Lock()
	addReceiptToDataPool(blockNumber, receipt)
	poolMutex.Unlock()
}

func addTransactionToDataPool(id uint64, transactions *big.Int) {
	dataPool[id] = Data{transactions: transactions}
}

func getTransaction(w3client *web3go.Client, blockNumber uint64, nodeUrl string) {
	trans, err := w3client.Eth.BlockTransactionCountByNumber(types.BlockNumber(blockNumber))
	// TODO: 错误处理
	if err != nil {
		log.Printf("Failed to get transaction %d from %s: %v", blockNumber, nodeUrl, err)
		return
	}
	fmt.Printf("We have transaction from block %d.\n", blockNumber)
	poolMutex.Lock()
	addTransactionToDataPool(blockNumber, trans)
	poolMutex.Unlock()
}

func addUncleCountToDataPool(id uint64, uncleCount *big.Int) {
	dataPool[id] = Data{uncleCounts: uncleCount}
}

func getUncleCount(w3client *web3go.Client, blockNumber uint64, nodeUrl string) {
	trans, err := w3client.Eth.BlockUnclesCountByNumber(types.BlockNumber(blockNumber))
	// TODO: 错误处理
	if err != nil {
		log.Printf("Failed to get uncle count %d from %s: %v", blockNumber, nodeUrl, err)
		return
	}
	fmt.Printf("We have uncle count from block %d.\n", blockNumber)
	poolMutex.Lock()
	addUncleCountToDataPool(blockNumber, trans)
	poolMutex.Unlock()
}

func PreloadPool(nodes []string, startBlock uint64, instruction string) {
	initDataPool()
	preloadCount := viper.GetUint64("preload.count")
	concurrentFetchData(nodes, startBlock, preloadCount, instruction)
}

func concurrentFetchData(nodes []string, startBlock, preloadNewGet uint64, instruction string) {
	var wg sync.WaitGroup
	wg.Add(len(nodes))

	for index, node := range nodes {
		go func(nodeUrl string, index int) {
			defer wg.Done()
			w3client := rpc.NewClient(nodeUrl)

			blocksPerNode := preloadNewGet / uint64(len(nodes))
			nodeStartBlock := startBlock + (blocksPerNode * uint64(index))
			var nodeEndBlock uint64
			if nodeStartBlock+blocksPerNode-1 <= startBlock+preloadNewGet-1 {
				nodeEndBlock = nodeStartBlock + blocksPerNode - 1
			} else {
				nodeEndBlock = startBlock + preloadNewGet - 1
			}
			for blockNumber := nodeStartBlock; blockNumber <= nodeEndBlock; blockNumber++ {
				if instruction == "BlockByNumber" {
					getBlock(w3client, blockNumber, nodeUrl)
				} else if instruction == "BlockReceipts" {
					getReceipt(w3client, blockNumber, nodeUrl)
				} else if instruction == "BlockTransactionCountByNumber" {
					getTransaction(w3client, blockNumber, nodeUrl)
				} else if instruction == "BlockUnclesCountByNumber" {
					getUncleCount(w3client, blockNumber, nodeUrl)
				} else {
					log.Printf("Fail to receive correct instruction.")
				}
			}
		}(node, index)
	}

	wg.Wait()
	nextNum = startBlock + preloadNewGet
}

func ConcurrentGetPool(nodes []string, startBlock, endBlock uint64, instruction string) {

	var wg sync.WaitGroup
	wg.Add(len(nodes))

	var fetchLock sync.Mutex

	for index := range nodes {
		go func(index int) {
			defer wg.Done()
			blocksPerNode := (endBlock - startBlock + 1) / uint64(len(nodes))
			nodeStartBlock := startBlock + (blocksPerNode * uint64(index))
			var nodeEndBlock uint64
			if nodeStartBlock+blocksPerNode-1 <= endBlock {
				nodeEndBlock = nodeStartBlock + blocksPerNode - 1
			} else {
				nodeEndBlock = endBlock
			}

			for blockNumber := nodeStartBlock; blockNumber <= nodeEndBlock; blockNumber++ {
				poolMutex.Lock()
				// data, ok := dataPool[blockNumber]
				_, ok := dataPool[blockNumber]
				if ok {
					delete(dataPool, blockNumber) // 从池中删除已取出的数据
				}
				poolMutex.Unlock()
				// TODO: 每个收到的data要具体做什么操作？
				//fmt.Printf("We get pool of block %d.\n", data.block.Number)
				//_, err := json.ConvertToJSON(data.block)
				//if err != nil {
				//	log.Printf("Failed to convert block %d to JSON: %v", blockNumber, err)
				//	continue
				//}
				fetchLock.Lock()
				if uint64(len(dataPool)) < viper.GetUint64("preload.min") {
					preloadCount := viper.GetUint64("preload.count")
					preloadMin := viper.GetUint64("preload.min")
					concurrentFetchData(nodes, nextNum, preloadCount-preloadMin, instruction)
				}
				fmt.Printf("现在pool长度是： %d.\n", len(dataPool))
				fetchLock.Unlock()
			}
		}(index)
	}

	wg.Wait()
}

func InitConcurrentGet(nodes []string, startBlock, endBlock uint64, instruction string) {

	// 预加载池中的数据
	PreloadPool(nodes, startBlock, instruction)

	start := startBlock
	end := viper.GetUint64("preload.count")
	if endBlock < end {
		end = endBlock
	}

	for start <= endBlock {
		ConcurrentGetPool(nodes, start, end, instruction)
		start = end + 1
		end = end + viper.GetUint64("preload.count")
		if endBlock < end {
			end = endBlock
		}

	}
}

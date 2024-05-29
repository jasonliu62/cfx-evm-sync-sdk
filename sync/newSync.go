package sync

import (
	"cfx-evm-sync-sdk/json"
	"cfx-evm-sync-sdk/rpc"
	"fmt"
	"github.com/openweb3/web3go/types"
	"github.com/spf13/viper"
	"log"
	"sync"
)

type Data struct {
	block *types.Block
}

var poolMutex sync.Mutex
var dataPool map[uint64]Data
var nextNum uint64

func initDataPool() {
	dataPool = make(map[uint64]Data)
}

func addToDataPool(id uint64, block *types.Block) {
	dataPool[id] = Data{block: block}
}

func PreloadPool(nodes []string, startBlock uint64) {
	initDataPool()
	preloadCount := viper.GetUint64("preload.count")
	var wg sync.WaitGroup
	wg.Add(len(nodes))
	for index, node := range nodes {
		go func(nodeUrl string, index int) {
			defer wg.Done()
			w3client := rpc.NewClient(nodeUrl)

			blocksPerNode := preloadCount / uint64(len(nodes))
			nodeStartBlock := startBlock + (blocksPerNode * uint64(index))
			var nodeEndBlock uint64
			if nodeStartBlock+blocksPerNode-1 <= startBlock+preloadCount-1 {
				nodeEndBlock = nodeStartBlock + blocksPerNode - 1
			} else {
				nodeEndBlock = startBlock + preloadCount - 1
			}
			// 一次预加载60个区块
			for blockNumber := nodeStartBlock; blockNumber <= nodeEndBlock; blockNumber++ {
				block, err := w3client.Eth.BlockByNumber(types.BlockNumber(blockNumber), false)
				// TODO: 错误处理
				if err != nil {
					log.Printf("Failed to get block %d from %s: %v", blockNumber, nodeUrl, err)
					continue
				}
				fmt.Printf("We have block %d.", blockNumber)
				poolMutex.Lock()
				addToDataPool(blockNumber, block)
				poolMutex.Unlock()
			}
		}(node, index)
	}

	wg.Wait()
	// TODO: 改成config里的参数
	nextNum = startBlock + 60

}

func concurrentFetchData(nodes []string, startBlock, preloadNewGet uint64) {
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
			// 一次预加载30个区块
			for blockNumber := nodeStartBlock; blockNumber <= nodeEndBlock; blockNumber++ {
				block, err := w3client.Eth.BlockByNumber(types.BlockNumber(blockNumber), false)
				// TODO: 错误处理
				if err != nil {
					log.Printf("Failed to get block %d from %s: %v", blockNumber, nodeUrl, err)
					continue
				}
				fmt.Printf("We have block %d.\n", blockNumber)
				poolMutex.Lock()
				addToDataPool(blockNumber, block)
				poolMutex.Unlock()
			}
		}(node, index)
	}

	wg.Wait()
	// TODO: 改成config里的参数
	nextNum = startBlock + 30
}

func ConcurrentGetPool(nodes []string, startBlock, endBlock uint64) {

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
				data, ok := dataPool[blockNumber]
				if ok {
					delete(dataPool, blockNumber) // 从池中删除已取出的数据
				}
				poolMutex.Unlock()
				fmt.Printf("We get pool of block %d.\n", data.block.Number)
				_, err := json.ConvertToJSON(data.block)
				if err != nil {
					log.Printf("Failed to convert block %d to JSON: %v", blockNumber, err)
					continue
				}
				fetchLock.Lock()
				if uint64(len(dataPool)) < viper.GetUint64("preload.min") {
					preloadCount := viper.GetUint64("preload.count")
					preloadMin := viper.GetUint64("preload.min")
					concurrentFetchData(nodes, nextNum, preloadCount-preloadMin)
				}
				fmt.Printf("现在pool长度是： %d.\n", len(dataPool))
				fetchLock.Unlock()
			}
		}(index)
	}

	wg.Wait()
}

func InitConcurrentGet(nodes []string, startBlock, endBlock uint64) {

	// 预加载池中的数据
	PreloadPool(nodes, startBlock)

	start := startBlock
	end := viper.GetUint64("preload.count")
	if endBlock < end {
		end = endBlock
	}

	for start <= endBlock {
		ConcurrentGetPool(nodes, start, end)
		start = end + 1
		end = end + viper.GetUint64("preload.count")
		if endBlock < end {
			end = endBlock
		}

	}
}

package sync

import (
	"cfx-evm-sync-sdk/json"
	"cfx-evm-sync-sdk/rpc"
	"fmt"
	"github.com/openweb3/web3go/types"
	"github.com/spf13/viper"
	"log"
	"sync"
	"time"
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
	// preloadMin := viper.GetUint64("preload.min")
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
	nextNum = startBlock + 60

	// 检查池的大小，如果少于 30 个，则启动一个 goroutine 去从网上获取数据并放入池
	// TODO: 需要在后面ConcurrentGetPool里面get并删除pool里东西的时候，一直检查如果池子大小少于30个，则去取额外的30个数据，具体function是：concurrentFetchData
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
}

func ConcurrentGetPool(nodes []string, startBlock, endBlock uint64) {

	var wg sync.WaitGroup
	wg.Add(len(nodes))

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
			}
			// 检查池的大小，如果少于 30 个，则启动一个 goroutine 去从网上获取数据并放入池
			// TODO: 需要在后面ConcurrentGetPool里面get并删除pool里东西的时候，一直检查如果池子大小少于30个，则去取额外的30个数据，具体function是：concurrentFetchData
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
		for uint64(len(dataPool)) < end {
			time.Sleep(time.Second) // 等待一秒钟后再次检查
		}
		ConcurrentGetPool(nodes, start, end)
		start = end + 1
		end = viper.GetUint64("preload.count")
		if endBlock < end {
			end = endBlock
		}

	}
}

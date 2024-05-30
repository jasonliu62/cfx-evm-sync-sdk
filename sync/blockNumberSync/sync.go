package blockNumberSync

import (
	"cfx-evm-sync-sdk/biz/blockNumberBiz"
	"cfx-evm-sync-sdk/rpc"
	"fmt"
	"github.com/openweb3/web3go"
	"github.com/spf13/viper"
	"log"
	"sync"
)

type DataWrap struct {
	value interface{}
}

var poolMutex sync.Mutex
var dataPool map[uint64]DataWrap
var nextNum uint64

// TODO: 增加一个缓存错误blockNumber用于重发的池子
// 具体什么时候发送？

func initDataPool() {
	dataPool = make(map[uint64]DataWrap)
}

func addDataToPool(id uint64, value interface{}) {
	dataPool[id] = DataWrap{value: value}
}

func dataTrim(w3client *web3go.Client, blockNumber uint64, nodeUrl string, getFunc blockNumberBiz.GetFunc) {
	result, err := getFunc(w3client, blockNumber)
	// TODO: 错误处理
	if err != nil {
		log.Printf("Failed to get data from block %d from %s: %v", blockNumber, nodeUrl, err)
		return
	}
	addDataToPool(blockNumber, result)
}

func PreloadPool(nodes []string, startBlock uint64, getFunc blockNumberBiz.GetFunc) {
	initDataPool()
	preloadCount := viper.GetUint64("preload.count")
	concurrentFetchData(nodes, startBlock, preloadCount, getFunc)
}

func concurrentFetchData(nodes []string, startBlock, preloadNewGet uint64, getFunc blockNumberBiz.GetFunc) {
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
				dataTrim(w3client, blockNumber, nodeUrl, getFunc)
			}
		}(node, index)
	}

	wg.Wait()
	nextNum = startBlock + preloadNewGet
}

func ConcurrentGetPool(nodes []string, startBlock, endBlock uint64, getFunc blockNumberBiz.GetFunc) {

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
				// poolMutex.Unlock()
				// TODO: 每个收到的data要具体做什么操作？
				// fetchLock.Lock()
				// 每扣一个都检测剩余池子还能不能满足30个，如果不满足，先填满再接着扣
				if uint64(len(dataPool)) < viper.GetUint64("preload.min") {
					preloadCount := viper.GetUint64("preload.count")
					preloadMin := viper.GetUint64("preload.min")
					concurrentFetchData(nodes, nextNum, preloadCount-preloadMin, getFunc)
				}
				fmt.Printf("现在pool长度是： %d.\n", len(dataPool))
				fetchLock.Unlock()
			}
		}(index)
	}

	wg.Wait()
}

func InitConcurrentGet(nodes []string, startBlock, endBlock uint64, getFunc blockNumberBiz.GetFunc) {

	// 预加载池中的数据
	PreloadPool(nodes, startBlock, getFunc)

	start := startBlock
	end := viper.GetUint64("preload.count")
	if endBlock < end {
		end = endBlock
	}

	for start <= endBlock {
		ConcurrentGetPool(nodes, start, end, getFunc)
		start = end + 1
		end = end + viper.GetUint64("preload.count")
		if endBlock < end {
			end = endBlock
		}

	}
}

func InitMultiTask(nodes []string, startBlock, endBlock uint64, getFuncs []blockNumberBiz.GetFunc) {
	if len(getFuncs) < len(nodes) {
		log.Printf("Functions exceeds node limits")
		return
	}

}

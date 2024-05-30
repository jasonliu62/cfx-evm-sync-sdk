package blockNumberSync

import (
	"cfx-evm-sync-sdk/biz/blockNumberBiz"
	"cfx-evm-sync-sdk/rpc"
	"fmt"
	"github.com/openweb3/web3go"
	"log"
	"sync"
)

type DataWrap struct {
	value interface{}
}

var poolMutex sync.Mutex
var dataPool map[uint64]DataWrap
var nextNum uint64

// Sdk 包含节点信息
type Sdk struct {
	Nodes        []string
	preloadCount uint64
	preloadMin   uint64
	Result       []DataWrap
}

// TODO: 增加一个缓存错误blockNumber用于重发的池子
// 具体什么时候发送？

func NewSdk(nodes []string, preloadC, preloadM uint64, res []DataWrap) *Sdk {
	return &Sdk{
		Nodes:        nodes,
		preloadCount: preloadC,
		preloadMin:   preloadM,
		Result:       res,
	}
}

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

func (s *Sdk) PreloadPool(startBlock uint64, getFunc blockNumberBiz.GetFunc) {
	initDataPool()
	s.concurrentFetchData(startBlock, s.preloadCount, getFunc)
}

func (s *Sdk) concurrentFetchData(startBlock, preloadNewGet uint64, getFunc blockNumberBiz.GetFunc) {
	var wg sync.WaitGroup
	wg.Add(len(s.Nodes))
	for index, node := range s.Nodes {
		go func(nodeUrl string, index int) {
			defer wg.Done()
			w3client := rpc.NewClient(nodeUrl)
			blocksPerNode := preloadNewGet / uint64(len(s.Nodes))
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

func (s *Sdk) ConcurrentGetPool(startBlock, endBlock uint64, getFunc blockNumberBiz.GetFunc) {

	var wg sync.WaitGroup
	wg.Add(len(s.Nodes))

	var fetchLock sync.Mutex

	for index := range s.Nodes {
		go func(index int) {
			defer wg.Done()
			blocksPerNode := (endBlock - startBlock + 1) / uint64(len(s.Nodes))
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
				data, ok := dataPool[blockNumber]
				if ok {
					delete(dataPool, blockNumber) // 从池中删除已取出的数据
				}
				// 输出的result
				s.Result = append(s.Result, data)
				// 每扣一个都检测剩余池子还能不能满足30个，如果不满足，先填满再接着扣
				if uint64(len(dataPool)) < s.preloadMin {
					s.concurrentFetchData(nextNum, s.preloadCount-s.preloadMin, getFunc)
				}
				fmt.Printf("现在pool长度是： %d.\n", len(dataPool))
				fetchLock.Unlock()
			}
		}(index)
	}

	wg.Wait()
}

func (s *Sdk) InitConcurrentGet(startBlock, endBlock uint64, getFunc blockNumberBiz.GetFunc) {

	// 预加载池中的数据
	s.PreloadPool(startBlock, getFunc)

	start := startBlock
	end := s.preloadCount
	if endBlock < end {
		end = endBlock
	}

	for start <= endBlock {
		s.ConcurrentGetPool(start, end, getFunc)
		start = end + 1
		end = end + s.preloadCount
		if endBlock < end {
			end = endBlock
		}

	}
}

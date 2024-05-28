package sync

import (
	"cfx-evm-sync-sdk/json"
	"cfx-evm-sync-sdk/rpc"
	"github.com/openweb3/web3go/types"
	"log"
	"sync"
	"time"
)

func SimpleGet(node string, startBlock, endBlock uint64) {

	var (
		totalReadDuration    time.Duration
		totalMarshalDuration time.Duration
		//totalWriteDuration   time.Duration
	)
	totalStartTime := time.Now()
	w3client := rpc.NewClient(node)

	// 循环获取并保存每个区块的数据
	for blockNumber := startBlock; blockNumber <= endBlock; blockNumber++ {

		// 记录读取区块开始时间
		readStartTime := time.Now()

		// 获取当前区块数据
		block, err := w3client.Eth.BlockByNumber(types.BlockNumber(blockNumber), false)
		if err != nil {
			log.Printf("Failed to get block %d: %v", blockNumber, err)
			continue
		}

		// 计算读取区块花费的时间
		readDuration := time.Since(readStartTime)
		totalReadDuration += readDuration

		// 记录 JSON 转换开始时间
		jsonStartTime := time.Now()

		// 转换为 JSON 格式并写入文件
		//jsonData, err := convertToJSON(block)
		_, err = json.ConvertToJSON(block)
		if err != nil {
			log.Printf("Failed to convert block %d to JSON: %v", blockNumber, err)
			continue
		}

		// 计算 JSON 转换花费的时间
		jsonDuration := time.Since(jsonStartTime)
		totalMarshalDuration += jsonDuration

		// 记录写入文件开始时间
		//writeStartTime := time.Now()

		//// 指定文件名
		//filename := fmt.Sprintf("block_%d.json", blockNumber)
		//
		//// 将 JSON 数据写入文件
		//err = writeJSONToFile(jsonData, filename)
		//if err != nil {
		//	log.Printf("Failed to write block %d JSON to file: %v", blockNumber, err)
		//	continue
		//}
		//
		//// 计算写入文件花费的时间
		//writeDuration := time.Since(writeStartTime)
		//totalWriteDuration += writeDuration

		// 输出每一步花费的时间
		//log.Printf("Block %d data processing times - Read: %s, JSON Marshal: %s, Write to file: %s", blockNumber, readDuration, jsonDuration, writeDuration)
		// log.Printf("Block %d data processing times - Read: %s, JSON Marshal: %s", blockNumber, readDuration, jsonDuration)
	}
	totalTime := time.Since(totalStartTime)
	log.Printf("The real time for single node processing 100 blocks: %s", totalTime)
	//log.Printf("Total read time for 100 blocks: %s", totalReadDuration)
	//log.Printf("Total json marshal time for 100 blocks: %s", totalMarshalDuration)
	//log.Printf("Total write time for 100 blocks: %s", totalWriteDuration)

}

func ConcurrentGet(nodes []string, startBlock, endBlock uint64) {

	var (
		totalReadDuration    time.Duration
		totalMarshalDuration time.Duration
		//totalWriteDuration   time.Duration
	)

	//nodes := []string{
	//	"https://evmtestnet.confluxrpc.com/09K6dkMRr9xyz6suyNgqTR",
	//	"https://evmtestnet.confluxrpc.com/09K6dkMRr9xyz6suyNgqTR",
	//}

	totalStartTime := time.Now()
	var wg sync.WaitGroup
	wg.Add(len(nodes))
	for index, node := range nodes {
		go func(nodeUrl string, index int) {
			defer wg.Done()
			w3client := rpc.NewClient(nodeUrl)
			// 计算每个节点负责的区块范围
			blocksPerNode := (endBlock - startBlock + 1) / uint64(len(nodes))
			nodeStartBlock := startBlock + (blocksPerNode * uint64(index))
			var nodeEndBlock uint64
			if nodeStartBlock+blocksPerNode-1 <= endBlock {
				nodeEndBlock = nodeStartBlock + blocksPerNode - 1
			} else {
				nodeEndBlock = endBlock
			}

			for blockNumber := nodeStartBlock; blockNumber <= nodeEndBlock; blockNumber++ {
				readStartTime := time.Now()
				block, err := w3client.Eth.BlockByNumber(types.BlockNumber(blockNumber), false)
				if err != nil {
					log.Printf("Failed to get block %d: %v", blockNumber, err)
					continue
				}
				readDuration := time.Since(readStartTime)
				jsonStartTime := time.Now()
				_, err = json.ConvertToJSON(block)
				if err != nil {
					log.Printf("Failed to convert block %d to JSON: %v", blockNumber, err)
					continue
				}
				jsonDuration := time.Since(jsonStartTime)
				totalReadDuration += readDuration
				totalMarshalDuration += jsonDuration
				// log.Printf("Block %d data processing times - Read: %s, JSON Marshal: %s", blockNumber, readDuration, jsonDuration)
			}
		}(node, index)
	}
	wg.Wait()
	totalTime := time.Since(totalStartTime)
	log.Printf("The real time for concrrently processing 100 blocks: %s", totalTime)

}

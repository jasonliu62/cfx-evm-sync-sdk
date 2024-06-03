package simpleSync

import (
	"cfx-evm-sync-sdk/common"
	"cfx-evm-sync-sdk/rpc"
	"context"
	"fmt"
	"log"
	"time"
)

type Sdk struct {
	Node   string
	Result map[uint64]common.DataWrap
}

func NewSdk(node string, res map[uint64]common.DataWrap) *Sdk {
	return &Sdk{
		Node:   node,
		Result: res,
	}
}

func (s *Sdk) SimpleGet(startBlock, endBlock uint64, getFunc common.GetFunc) {
	w3client := rpc.NewClient(s.Node)
	// 循环获取并保存每个区块的数据
	for blockNumber := startBlock; blockNumber <= endBlock; blockNumber++ {
		// 获取当前区块数据
		result, err := getFunc(w3client, blockNumber)
		// TODO: 错误处理
		for err != nil {
			s.Result[blockNumber] = common.DataWrap{Error: err}
			log.Printf("Failed to get data from block %d from %s: %v", blockNumber, s.Node, err)
			time.Sleep(1 * time.Second)
			result, err = getFunc(w3client, blockNumber)
		}
		s.Result[blockNumber] = common.DataWrap{Value: result}
	}
}

func (s *Sdk) ContinueGet(ctx context.Context, startBlock uint64, getFunc common.GetFunc) {
	w3client := rpc.NewClient(s.Node)
	currentBlock := startBlock
	// 循环获取并保存每个区块的数据
	for {
		select {
		case <-ctx.Done():
			log.Printf("ContinueGet terminated")
			return
		default:
			for {
				result, err := getFunc(w3client, currentBlock)
				// TODO: 错误处理
				for err != nil {
					s.Result[currentBlock] = common.DataWrap{Error: err}
					log.Printf("Failed to get data from block %d from %s: %v", currentBlock, s.Node, err)
					time.Sleep(1 * time.Second)
					result, err = getFunc(w3client, currentBlock)
				}
				s.Result[currentBlock] = common.DataWrap{Value: result}
				fmt.Printf("We have block %d.", currentBlock)
				break
			}
		}
		currentBlock++
	}
}

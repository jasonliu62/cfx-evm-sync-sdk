package simpleSync

import (
	"cfx-evm-sync-sdk/common"
	"cfx-evm-sync-sdk/rpc"
	"log"
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
		if err != nil {
			log.Printf("Failed to get data from block %d from %s: %v", blockNumber, s.Node, err)
			return
		}

		s.Result[blockNumber] = common.DataWrap{Value: result}
	}
}

package simpleSync

import (
	"cfx-evm-sync-sdk/data"
	"cfx-evm-sync-sdk/rpc"
	"context"
	"fmt"
	"github.com/openweb3/web3go"
	"log"
	"time"
)

type Sdk struct {
	// w3client *web3go.Client
	Result  map[uint64]data.DataWrap
	GetFunc data.GetFunc
}

func NewSdk(getFunc data.GetFunc) *Sdk {
	return &Sdk{
		// w3client: rpc.NewClient(node),
		Result:  make(map[uint64]data.DataWrap),
		GetFunc: getFunc,
	}
}

func (s *Sdk) SimpleGet(w3client *web3go.Client, startBlock, endBlock uint64) map[uint64]data.DataWrap {
	// 循环获取并保存每个区块的数据
	for blockNumber := startBlock; blockNumber <= endBlock; blockNumber++ {
		// 获取当前区块数据
		result, err := s.GetFunc(w3client, blockNumber)
		// TODO: 错误处理
		for err != nil {
			s.Result[blockNumber] = data.DataWrap{Error: err}
			log.Printf("Failed to get data from block %d: %v", blockNumber, err)
			time.Sleep(1 * time.Second)
			result, err = s.GetFunc(w3client, blockNumber)
		}
		s.Result[blockNumber] = data.DataWrap{Value: result}
	}
	return s.Result
}

func (s *Sdk) ContinueGet(w3client *web3go.Client, ctx context.Context, startBlock uint64) {
	currentBlock := startBlock
	// 循环获取并保存每个区块的数据
	for {
		select {
		case <-ctx.Done():
			log.Printf("ContinueGet terminated")
			return
		default:
			for {
				result, err := s.GetFunc(w3client, currentBlock)
				// TODO: 错误处理需要放在biz层面。后续需要修改
				for err != nil {
					s.Result[currentBlock] = data.DataWrap{Error: err}
					log.Printf("Failed to get data from block %d: %v", currentBlock, err)
					time.Sleep(1 * time.Second)
					result, err = s.GetFunc(w3client, currentBlock)
				}
				s.Result[currentBlock] = data.DataWrap{Value: result}
				fmt.Printf("We have block %d.", currentBlock)
				break
			}
		}
		currentBlock++
	}
}

func (s *Sdk) Get(w3client *web3go.Client, blockNumberOrHash data.BlockNumberOrHash) data.DataWrap {
	result, err := s.GetFunc(w3client, blockNumberOrHash)
	s.Result[blockNumberOrHash.BlockNumber] = data.DataWrap{
		Value: result,
		Error: err,
	}
	return s.Result[blockNumberOrHash.BlockNumber]
}

func GetRpcClient(node string) *web3go.Client {
	return rpc.NewClient(node)
}

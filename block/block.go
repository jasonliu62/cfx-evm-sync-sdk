package block

import (
	"github.com/openweb3/web3go"
	"github.com/openweb3/web3go/types"
)

// GetBlockByNumber 从指定的节点获取区块数据
func GetBlockByNumber(w3client *web3go.Client, blockNumber uint64) (*types.Block, error) {
	// 获取当前区块数据
	block, err := w3client.Eth.BlockByNumber(types.BlockNumber(blockNumber), false)
	if err != nil {
		return nil, err
	}
	return block, nil
}

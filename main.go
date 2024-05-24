package main

import (
	"cfx-evm-sync-sdk/config"
	"cfx-evm-sync-sdk/sync"
	"github.com/spf13/viper"
)

func main() {

	// 定义要获取的区块范围
	//startBlock := uint64(1)
	//endBlock := uint64(100)
	config.InitConfig()
	// 从配置文件读取区块范围
	startBlock := viper.GetUint64("block.start")
	endBlock := viper.GetUint64("block.end")

	// 从配置文件读取节点地址
	nodes := viper.GetStringSlice("nodes")

	sync.SimpleGet(nodes[0], startBlock, endBlock)

	// 并发访问节点
	sync.ConcurrentGet(nodes, startBlock, endBlock)

}

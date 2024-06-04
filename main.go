package main

import (
	"cfx-evm-sync-sdk/biz/simpleBiz"
	"cfx-evm-sync-sdk/config"
	"cfx-evm-sync-sdk/store/cfxMysql"
	"github.com/spf13/viper"
)

func main() {

	//// 定义要获取的区块范围
	////startBlock := uint64(1)
	////endBlock := uint64(100)
	config.InitConfig()
	//// 从配置文件读取区块范围
	////startBlock := viper.GetUint64("block.start")
	////endBlock := viper.GetUint64("block.end")
	//
	//// 从配置文件读取节点地址
	nodeUrl := viper.GetStringSlice("nodes")[0]
	//
	//// sync.SimpleGet(nodes[0], startBlock, endBlock)
	//
	db := cfxMysql.Start()
	simpleBiz.ContinueBlockByNumber(nodeUrl, uint64(10), db)
	//for key, dataWrap := range res {
	//	fmt.Printf("Key: %d, Value: %v, Type: %T\n", key, dataWrap.Value, dataWrap.Value)
	//}
}

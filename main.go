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
	res := simpleBiz.BlockByNumber(nodeUrl, uint64(10), uint64(12))
	//for key, dataWrap := range res {
	//	fmt.Printf("Key: %d, Value: %v, Type: %T\n", key, dataWrap.Value, dataWrap.Value)
	//}
	err := simpleBiz.StoreBlock(res, db)
	if err != nil {
		return
	}
}

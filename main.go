package main

import (
	"cfx-evm-sync-sdk/biz/simpleBiz"
	"cfx-evm-sync-sdk/config"
	"cfx-evm-sync-sdk/store/cfxMysql"
	"github.com/spf13/viper"
)

func main() {
	config.InitConfig()
	nodeUrl := viper.GetStringSlice("nodes")[0]
	db := cfxMysql.Start()
	simpleBiz.ContinueBlockByNumber(nodeUrl, uint64(97971351), db)
}

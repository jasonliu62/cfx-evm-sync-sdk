package main

import (
	"cfx-evm-sync-sdk/biz/simpleBiz"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
)

func main() {
	config.InitConfig()
	nodeUrl := viper.GetStringSlice("nodes")[0]
	db := cfxMysql.Start()
	simpleBiz.ContinueBlockByNumber(nodeUrl, uint64(97971351), db)
}

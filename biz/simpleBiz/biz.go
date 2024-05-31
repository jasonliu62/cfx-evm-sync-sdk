package simpleBiz

import (
	"cfx-evm-sync-sdk/common"
	"cfx-evm-sync-sdk/store/cfxMysql"
	"cfx-evm-sync-sdk/sync/simpleSync"
	"fmt"
	"github.com/openweb3/web3go"
	"github.com/openweb3/web3go/types"
	"gorm.io/gorm"
)

func BlockByNumber(node string, startBlock, endBlock uint64) map[uint64]common.DataWrap {
	result := make(map[uint64]common.DataWrap)
	s := simpleSync.NewSdk(node, result)
	GetFunc := func(w3client *web3go.Client, blockNumber interface{}) (interface{}, error) {
		blockNumberUint := blockNumber.(uint64)
		block, err := w3client.Eth.BlockByNumber(types.BlockNumber(blockNumberUint), false)
		return block, err
	}
	s.SimpleGet(startBlock, endBlock, GetFunc)
	return result
}

func StoreBlock(res map[uint64]common.DataWrap, db *gorm.DB) error {
	for key, dataWrap := range res {
		block, ok := dataWrap.Value.(*types.Block)
		if !ok {
			return fmt.Errorf("invalid type for block at key %d", key)
		}
		// 转换并存储块数据
		dbBlock := cfxMysql.ConvertBlock(block)
		if err := db.Create(&dbBlock).Error; err != nil {
			return err
		}
	}
	return nil
}

package simpleBiz

import (
	"cfx-evm-sync-sdk/common"
	"cfx-evm-sync-sdk/rpc"
	"cfx-evm-sync-sdk/store/cfxMysql"
	"cfx-evm-sync-sdk/sync/simpleSync"
	"context"
	"fmt"
	"github.com/openweb3/web3go"
	"github.com/openweb3/web3go/types"
	"gorm.io/gorm"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
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

func ContinueBlockByNumber(node string, startBlock uint64, db *gorm.DB) {
	result := make(map[uint64]common.DataWrap)
	s := simpleSync.NewSdk(node, result)
	GetFunc := func(w3client *web3go.Client, blockNumber interface{}) (interface{}, error) {
		blockNumberUint := blockNumber.(uint64)
		block, err := w3client.Eth.BlockByNumber(types.BlockNumber(blockNumberUint), false)
		return block, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigs := make(chan os.Signal)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigs
		log.Println("Received termination signal")
		cancel()
	}()

	w3client := rpc.NewClient(node)
	currentBlock := startBlock
	//s.ContinueGet(ctx, startBlock, GetFunc)
	for {
		select {
		case <-ctx.Done():
			log.Printf("ContinueGet terminated")
			return
		default:
			for {
				s.Get(w3client, currentBlock, GetFunc)
				err := result[currentBlock].Error
				if err != nil {
					log.Println("Get err:", err)
					time.Sleep(1 * time.Second)
					continue
				}
				fmt.Printf("We have block %d.", currentBlock)
				err = StoreBlock(result[currentBlock].Value.(*types.Block), db)
				if err != nil {
					return
				}
				break
			}
		}
		currentBlock++
	}
}

func StoreBlockFromMap(res map[uint64]common.DataWrap, db *gorm.DB) error {
	for key, dataWrap := range res {
		block, ok := dataWrap.Value.(*types.Block)
		if !ok {
			return fmt.Errorf("invalid type for block at key %d", key)
		}
		// 转换并存储块数据
		dbBlock := cfxMysql.ConvertBlockWithoutAuthor(block)
		authorName := cfxMysql.ConvertAuthorToString(block.Author)
		if err := cfxMysql.StoreBlock(db, dbBlock, authorName); err != nil {
			return err
		}
	}
	return nil
}

func StoreBlock(block *types.Block, db *gorm.DB) error {
	dbBlock := cfxMysql.ConvertBlockWithoutAuthor(block)
	authorName := cfxMysql.ConvertAuthorToString(block.Author)
	if err := cfxMysql.StoreBlock(db, dbBlock, authorName); err != nil {
		return err
	}
	return nil
}

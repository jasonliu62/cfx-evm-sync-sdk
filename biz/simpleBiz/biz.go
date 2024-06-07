package simpleBiz

import (
	"cfx-evm-sync-sdk/data"
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

//func BlockByNumber(node string, startBlock, endBlock uint64) map[uint64]data.DataWrap {
//	GetFunc := func(w3client *web3go.Client, blockNumber interface{}) (interface{}, error) {
//		blockNumberUint := blockNumber.(uint64)
//		block, err := w3client.Eth.BlockByNumber(types.BlockNumber(blockNumberUint), false)
//		return block, err
//	}
//	s := simpleSync.NewSdk(GetFunc)
//	w3client := simpleSync.GetRpcClient(node)
//	return s.SimpleGet(startBlock, endBlock)
//
//}

func ContinueBlockByNumber(node string, startBlock uint64, db *gorm.DB) {
	GetFuncBlock := func(w3client *web3go.Client, blockNumberOrHash data.BlockNumberOrHash) (interface{}, error) {
		blockNumberUint := blockNumberOrHash.BlockNumber
		block, err := w3client.Eth.BlockByNumber(types.BlockNumber(blockNumberUint), false)
		blockData := data.BlockData{
			Block:              block,
			TransactionDetails: []*types.TransactionDetail{},
		}
		blockHashList := block.Transactions.Hashes()
		var transactionDetail *types.TransactionDetail
		for _, hash := range blockHashList {
			transactionDetail, err = w3client.Eth.TransactionByHash(hash)
			blockData.TransactionDetails = append(blockData.TransactionDetails, transactionDetail)
		}
		return blockData, err
	}
	s := simpleSync.NewSdk(GetFuncBlock, node)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigs := make(chan os.Signal)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigs
		log.Println("Received termination signal")
		cancel()
	}()

	currentBlock := startBlock
	//s.ContinueGet(ctx, startBlock, GetFunc)
	for {
		select {
		case <-ctx.Done():
			log.Printf("ContinueGet terminated")
			return
		default:
			for {
				blockNumberOrHash := data.BlockNumberOrHash{
					BlockNumber: currentBlock,
				}
				result := s.Get(blockNumberOrHash)
				err := result.Error
				if err != nil {
					log.Println("Get err:", err)
					time.Sleep(1 * time.Second)
					continue
				}
				fmt.Printf("We have block %d.", currentBlock)
				err = cfxMysql.StoreBlockAndTransactions(db, result.Value.(data.BlockData).Block, result.Value.(data.BlockData).TransactionDetails)
				if err != nil {
					log.Printf("Failed to store block %d and transactions: %v", currentBlock, err)
					return
				}
				break
			}
		}
		currentBlock++
	}
}

//func StoreBlockFromMap(res map[uint64]data.DataWrap, db *gorm.DB) error {
//	for key, dataWrap := range res {
//		block, ok := dataWrap.Value.(*types.Block)
//		if !ok {
//			return fmt.Errorf("invalid type for block at key %d", key)
//		}
//		// 转换并存储块数据
//		dbBlock := cfxMysql.ConvertBlockWithoutAuthor(block)
//		authorName := cfxMysql.ConvertAddressToString(block.Author)
//		if err := cfxMysql.StoreBlock(db, dbBlock, authorName); err != nil {
//			return err
//		}
//	}
//	return nil
//}

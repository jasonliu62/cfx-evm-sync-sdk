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
			Logs:               []*types.Log{},
		}
		blockHashList := block.Transactions.Hashes()
		var transactionDetail *types.TransactionDetail
		for _, hash := range blockHashList {
			transactionDetail, err = w3client.Eth.TransactionByHash(hash)
			blockData.TransactionDetails = append(blockData.TransactionDetails, transactionDetail)
			txHash := transactionDetail.Hash
			receipt, err := w3client.Eth.TransactionReceipt(txHash)
			if err != nil {
			}
			if receipt.Logs == nil {
				continue
			}
			blockData.Logs = append(blockData.Logs, receipt.Logs...)
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

	currentBlock, err := cfxMysql.GetInitBlockNumber(db, startBlock)
	if err != nil {
		log.Println("GetInitBlockNumber err:", err)
	}
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
				blkDataSQL, err := convertBlockAndTransactionDetails(result.Value.(data.BlockData).Block, result.Value.(data.BlockData).TransactionDetails,
					result.Value.(data.BlockData).Logs, db)
				if err != nil {
					log.Printf("Failed to convert blockData %d: %v", currentBlock, err)
					return
				}
				err = cfxMysql.StoreBlockTransactionsAndLogs(db, blkDataSQL)
				if err != nil {
					log.Printf("Failed to store blockData %d to MySQL: %v", currentBlock, err)
					return
				}
				break
			}
		}
		currentBlock++
	}
}

func convertBlockAndTransactionDetails(block *types.Block, transactionDetails []*types.TransactionDetail,
	logs []*types.Log, db *gorm.DB) (cfxMysql.BlockDataMySQL, error) {
	dbBlock := cfxMysql.ConvertBlockWithoutAuthor(block)
	authorName := cfxMysql.ConvertAddressToString(block.Author)
	author, err := cfxMysql.FindOrCreateAddress(db, authorName)
	if err != nil {
		return cfxMysql.BlockDataMySQL{}, fmt.Errorf("failed to find or create author address: %w", err)
	}
	dbBlock.AuthorID = author.ID
	var dbTransactionDetailList []cfxMysql.TransactionDetail
	var dbLogList []cfxMysql.Log
	for index, transactionDetail := range transactionDetails {
		dbTransactionDetail := cfxMysql.ConvertTransactionDetail(uint(index), transactionDetail)
		from, err := cfxMysql.FindOrCreateAddress(db, cfxMysql.ConvertAddressToString(&transactionDetail.From))
		if err != nil {
			return cfxMysql.BlockDataMySQL{}, fmt.Errorf("failed to find or create from address: %w", err)
		}
		to, err := cfxMysql.FindOrCreateAddress(db, cfxMysql.ConvertAddressToString(transactionDetail.To))
		if err != nil {
			return cfxMysql.BlockDataMySQL{}, fmt.Errorf("failed to find or create to address: %w", err)
		}
		dbTransactionDetail.FromAddress = from.ID
		dbTransactionDetail.ToAddress = to.ID
		dbTransactionDetailList = append(dbTransactionDetailList, dbTransactionDetail)
	}
	for _, l := range logs {
		topics := l.Topics
		dbLog := cfxMysql.ConvertLogWithoutTopic(l)
		topic0 := l.Topics[0]
		t0Hash, err := cfxMysql.FindOrCreateHash(db, cfxMysql.ConvertHashToString(topic0))
		address, err := cfxMysql.FindOrCreateAddress(db, cfxMysql.ConvertAddressToString(&l.Address))
		if err != nil {
			return cfxMysql.BlockDataMySQL{}, fmt.Errorf("failed to find or create author address: %w", err)
		}
		dbLog.Topic0 = t0Hash.ID
		dbLog.Address = address.ID
		dbLogList = append(dbLogList, cfxMysql.ConvertLogTopics(dbLog, topics))
	}
	return cfxMysql.BlockDataMySQL{
		Block:              dbBlock,
		TransactionDetails: dbTransactionDetailList,
		Logs:               dbLogList,
	}, err
}

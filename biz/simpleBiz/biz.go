package simpleBiz

import (
	erc20 "cfx-evm-sync-sdk/abi"
	"cfx-evm-sync-sdk/data"
	"cfx-evm-sync-sdk/store/cfxMysql"
	"cfx-evm-sync-sdk/sync/simpleSync"
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
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
			Erc20Transfer:      []*types.Log{},
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
		clientForContract, _ := w3client.ToClientForContract()
		transfer := "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
		for _, lg := range blockData.Logs {
			if IsErc20(lg.Address, clientForContract) {
				// hardcode：topic[0]: 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef
				if lg.Topics[0].Hex() != transfer {
					continue
				}
				blockData.Erc20Transfer = append(blockData.Erc20Transfer, lg)
			} else {
				continue
			}
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
				blkDataSQL, err := convertToBlkData(result.Value.(data.BlockData).Block, result.Value.(data.BlockData).TransactionDetails,
					result.Value.(data.BlockData).Logs, result.Value.(data.BlockData).Erc20Transfer, db)
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

func convertToBlkData(block *types.Block, transactionDetails []*types.TransactionDetail,
	logs []*types.Log, erc20Transfers []*types.Log, db *gorm.DB) (cfxMysql.BlockDataMySQL, error) {
	dbBlock := cfxMysql.ConvertBlockWithoutAuthor(block)
	authorName := cfxMysql.ConvertAddressToString(block.Author)
	author, err := cfxMysql.FindOrCreateAddress(db, authorName)
	if err != nil {
		return cfxMysql.BlockDataMySQL{}, fmt.Errorf("failed to find or create author address: %w", err)
	}
	dbBlock.AuthorID = author.ID
	var dbTransactionDetailList []cfxMysql.TransactionDetail
	var dbLogList []cfxMysql.Log
	var dbErc20List []cfxMysql.Erc20Transfer
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
			return cfxMysql.BlockDataMySQL{}, fmt.Errorf("failed to find or create address: %w", err)
		}
		dbLog.Topic0 = t0Hash.ID
		dbLog.Address = address.ID
		dbLogList = append(dbLogList, cfxMysql.ConvertLogTopics(dbLog, topics))
	}
	for _, t := range erc20Transfers {
		dbErc20 := cfxMysql.ConvertErc20Transfer(t)
		address, err := cfxMysql.FindOrCreateAddress(db, cfxMysql.ConvertAddressToString(&t.Address))
		src, err := cfxMysql.FindOrCreateAddress(db, cfxMysql.ConvertHashToString(t.Topics[1]))
		dst, err := cfxMysql.FindOrCreateAddress(db, cfxMysql.ConvertHashToString(t.Topics[2]))
		if err != nil {
			return cfxMysql.BlockDataMySQL{}, fmt.Errorf("failed to find or create address: %w", err)
		}
		dbErc20.Address = address.ID
		dbErc20.Src = src.ID
		dbErc20.Dst = dst.ID
		dbErc20List = append(dbErc20List, dbErc20)
	}
	return cfxMysql.BlockDataMySQL{
		Block:              dbBlock,
		TransactionDetails: dbTransactionDetailList,
		Logs:               dbLogList,
		Erc20Transfers:     dbErc20List,
	}, err
}

func IsErc20(contractAddress common.Address, backend bind.ContractBackend) bool {
	erc20Contract, err := erc20.NewErc20(contractAddress, backend)
	if err != nil {
		log.Printf("Cannot make new erc20 contract: %v", err)
		return false
	}
	name, err := erc20Contract.Name(&bind.CallOpts{Context: context.Background()})
	if err != nil {
		log.Printf("Cannot get contract name: %v", err)
		return false
	}
	symbol, err := erc20Contract.Symbol(&bind.CallOpts{Context: context.Background()})
	if err != nil {
		log.Printf("Cannot get contract symbol: %v", err)
		return false
	}
	decimals, err := erc20Contract.Decimals(&bind.CallOpts{Context: context.Background()})
	if err != nil {
		log.Printf("Cannot get contract decimal: %v", err)
		return false
	}
	if name != "" && symbol != "" && decimals != 0 {
		return true
	}
	return false
}

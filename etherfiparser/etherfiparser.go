package etherfiparser

import (
	"context"
	"etherfi/cache"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"log"
	"math/big"
	"sync"
)

// Asset 表示用户资产的数据结构
type Asset struct {
	Symbol   string   // 资产符号，例如ETH、USDC
	Amount   *big.Int // 资产数量
	Decimals int      // 资产的小数位数
	Protocol string   // 协议名称
}

type AssetParser interface {
	ParseAssetDetails(tx *types.Transaction) (string, *big.Int, uint8, error)
}

// ProtocolParser 定义了解析用户资产的接口
type ProtocolParser interface {
	ParseBlock(chain, protocol string, blockNumber uint64, txIndex uint) (bool, error)
	GetLatestBlock() (uint64, error)
}

type EtherFiParser struct {
	client    *ethclient.Client
	rpcClient *rpc.Client
	cache     cache.Cache
}

func NewEtherFiParser(client *ethclient.Client, rpcClient *rpc.Client, cache cache.Cache) *EtherFiParser {
	return &EtherFiParser{
		client:    client,
		rpcClient: rpcClient,
		cache:     cache,
	}
}

// GetLatestBlock 获取最新区块高度
func (e *EtherFiParser) GetLatestBlock() (uint64, error) {
	latestBlock, err := e.client.BlockByNumber(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	return latestBlock.NumberU64(), nil
}

// ParseBlock 解析区块并返回指定用户的资产信息
func (e *EtherFiParser) ParseBlock(chain, protocol string, blockNumber uint64, txIndex uint) (bool, error) {
	block, err := e.client.BlockByNumber(context.Background(), big.NewInt(int64(blockNumber)))
	if err != nil {
		return false, err
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(block.Transactions()))

	// 遍历区块中的交易
	for i := int(txIndex); i < len(block.Transactions()); i++ {
		wg.Add(1)
		go func(tx *types.Transaction, index int) {
			defer wg.Done()
			etherFiContract := common.HexToAddress("0xYourEtherFiContractAddress")
			if isEtherFiTransaction(tx.To(), etherFiContract) {
				// 获取合约内部调用的详细信息
				internalCalls, err := e.traceTransaction(tx.Hash())
				if err != nil {
					errChan <- fmt.Errorf("traceTransaction error: %w", err)
					return
				}

				for _, call := range internalCalls {
					symbol, amount, decimals := getAssetDetailsFromCall(call)
					log.Printf(" 类型=%s, symbol=%s, amount=%s, decimals=%d", call.Type, symbol, amount.String(), decimals)
				}
			}

			// 更新解析进度
			err := e.cache.SetLastParsedBlock(chain, protocol, cache.BlockProgress{
				BlockHeight: blockNumber,
				TxIndex:     uint(index),
			})
			if err != nil {
				errChan <- err
			}
		}(block.Transactions()[i], i)
	}

	// 等待所有并发任务完成
	wg.Wait()
	close(errChan)

	// 检查是否有错误发生
	for err := range errChan {
		if err != nil {
			log.Printf("解析时出错: %v", err)
			return false, err
		}
	}

	// 完成解析该区块的所有交易
	return true, nil
}

// 辅助函数：检查是否为 EtherFi 交易
func isEtherFiTransaction(to *common.Address, contractAddress common.Address) bool {
	return to != nil && *to == contractAddress
}

// traceTransaction 使用 debug_traceTransaction 解析交易内部调用
func (e *EtherFiParser) traceTransaction(txHash common.Hash) ([]Call, error) {
	var result map[string]interface{}
	err := e.rpcClient.CallContext(context.Background(), &result, "debug_traceTransaction", txHash)
	if err != nil {
		return nil, err
	}

	// 解析内部调用
	calls, err := parseInternalCalls(result)
	if err != nil {
		return nil, err
	}
	return calls, nil
}

// Call 表示合约的内部调用结构
type Call struct {
	Type     string   // 调用类型，例如 "transfer"
	Symbol   string   // 资产符号
	Amount   *big.Int // 资产数量
	Decimals int      // 小数位数
}

// parseInternalCalls 从 traceTransaction 的结果解析出内部调用信息
func parseInternalCalls(data map[string]interface{}) ([]Call, error) {
	var calls []Call
	// 依据 trace 数据结构解析内部调用
	// 示例代码，这部分需要根据具体 trace 数据结构进行解析
	for _, item := range data["calls"].([]interface{}) {
		call := Call{
			Type:     item.(map[string]interface{})["type"].(string),
			Symbol:   "ETH",                           // 假设为 ETH，根据实际情况解析
			Amount:   big.NewInt(1000000000000000000), // 假设金额为 1 ETH
			Decimals: 18,
		}
		calls = append(calls, call)
	}
	return calls, nil
}

// getAssetDetailsFromCall 从内部调用获取资产信息
func getAssetDetailsFromCall(call Call) (string, *big.Int, uint8) {
	return call.Symbol, call.Amount, uint8(call.Decimals)
}

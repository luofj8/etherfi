package etherfiparser

import (
	"context"
	"etherfi/cache"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
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
	client *ethclient.Client
	cache  cache.Cache
}

func NewEtherFiParser(client *ethclient.Client, cache cache.Cache) *EtherFiParser {
	return &EtherFiParser{
		client: client,
		cache:  cache,
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

// 解析区块并返回指定用户的资产信息
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
				assetType := getAssetType(tx.Data())
				symbol, amount, decimals := getAssetDetails(tx)
				log.Printf(" 类型=%s, symbol=%s, amount=%s, decimals=%d", assetType, symbol, amount.String(), decimals)
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
// 在配置文件中加载合约地址，替代硬编码
func isEtherFiTransaction(to *common.Address, contractAddress common.Address) bool {
	return to != nil && *to == contractAddress
}

// 辅助函数：解析交易类型
func getAssetType(data []byte) string {
	// 依据 EtherFi 协议解析资产类型
	return "Stake" // 示例返回
}

// 辅助函数：解析资产详情
func getAssetDetails(tx *types.Transaction) (string, *big.Int, uint8) {
	return "ETH", big.NewInt(1000000000000000000), 18 // 示例返回
}

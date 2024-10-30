package scheduler

import (
	"etherfi/cache"
	"etherfi/etherfiparser"
	"log"
	"time"
)

type ChainScheduler struct {
	parsers map[string]map[string]etherfiparser.ProtocolParser // 链-协议的解析器映射
	cache   cache.Cache                                        // 缓存进度
}

func NewChainScheduler(cache cache.Cache) *ChainScheduler {
	return &ChainScheduler{
		parsers: make(map[string]map[string]etherfiparser.ProtocolParser),
		cache:   cache,
	}
}

// 注册链与协议
func (s *ChainScheduler) RegisterProtocol(chainName, protocolName string, parser etherfiparser.ProtocolParser) {
	if _, ok := s.parsers[chainName]; !ok {
		s.parsers[chainName] = make(map[string]etherfiparser.ProtocolParser)
	}
	s.parsers[chainName][protocolName] = parser
}

// 启动调度
func (s *ChainScheduler) Start() {
	for chainName, protocols := range s.parsers {
		for protocolName, parser := range protocols {
			go func(chain, protocol string, p etherfiparser.ProtocolParser) {
				for {
					// 获取解析进度
					progress, err := s.cache.GetLastParsedBlock(chain, protocol)
					if err != nil {
						log.Fatalf("无法获取 %s/%s 的解析进度: %v", chain, protocol, err)
					}

					// 获取链上最新区块
					latestBlock, err := p.GetLatestBlock()
					if err != nil {
						log.Fatalf("获取 %s/%s 最新区块时出错: %v", chain, protocol, err)
					}

					// 按顺序解析区块
					for progress.BlockHeight <= latestBlock {
						done, err := p.ParseBlock(chain, protocol, progress.BlockHeight, progress.TxIndex)
						if err != nil {
							log.Printf("解析 %s/%s 的区块 %d 时出错: %v", chain, protocol, progress.BlockHeight, err)
						}

						// 如果区块解析完成
						if done {
							progress.BlockHeight++
							progress.TxIndex = 0
						}
					}

					// 等待片刻，继续解析新块
					time.Sleep(10 * time.Second)
				}
			}(chainName, protocolName, parser)
		}
	}
}

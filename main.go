package main

import (
	"etherfi/cache"
	"etherfi/config"
	"etherfi/etherfiparser"
	"etherfi/scheduler"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"log"
)

func main() {
	config.LoadConfig()

	// 初始化Redis缓存
	redisCache := cache.NewRedisCache(config.AppConfig.Redis.Address)

	// 初始化以太坊客户端
	ethClient, err := ethclient.Dial(config.AppConfig.Ethereum.RpcURL)
	if err != nil {
		log.Fatalf("连接以太坊客户端出错: %v", err)
	}

	rpcClient, err := rpc.Dial(config.AppConfig.Ethereum.RpcURL)
	if err != nil {
		log.Fatalf("连接 RPC 客户端出错: %v", err)
	}

	// 初始化 EtherFi 解析器，传入 ethClient 和 rpcClient
	ethParser := etherfiparser.NewEtherFiParser(ethClient, rpcClient, redisCache)
	// 初始化 EtherFi 解析器

	// 创建调度器
	scheduler := scheduler.NewChainScheduler(redisCache)

	// 注册 EtherFi 协议
	scheduler.RegisterProtocol("Ethereum", "EtherFi", ethParser)

	// 启动调度器，开始解析任务
	scheduler.Start()

	// 阻止主程序退出
	select {}
}

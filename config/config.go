package config

import (
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
)

type Config struct {
	Ethereum struct {
		RpcURL string `yaml:"rpc_url"`
	} `yaml:"ethereum"`
	Arbitrum struct {
		RpcURL string `yaml:"rpc_url"`
	} `yaml:"arbitrum"`
	Redis struct {
		Address string `yaml:"address"`
	} `yaml:"redis"`
}

var AppConfig Config

func LoadConfig() {
	// 加载 .env 文件（可选）
	err := godotenv.Load()
	if err != nil {
		log.Println("未找到 .env 文件，跳过环境变量加载")
	}

	// 从 config.yaml 加载配置
	data, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("读取配置文件错误: %v", err)
	}

	err = yaml.Unmarshal(data, &AppConfig)
	if err != nil {
		log.Fatalf("解析配置文件错误: %v", err)
	}

	// 优先从环境变量中读取配置
	if os.Getenv("ETHEREUM_RPC_URL") != "" {
		AppConfig.Ethereum.RpcURL = os.Getenv("ETHEREUM_RPC_URL")
	}
	if os.Getenv("ARBITRUM_RPC_URL") != "" {
		AppConfig.Arbitrum.RpcURL = os.Getenv("ARBITRUM_RPC_URL")
	}
	if os.Getenv("REDIS_ADDRESS") != "" {
		AppConfig.Redis.Address = os.Getenv("REDIS_ADDRESS")
	}
}

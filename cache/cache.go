package cache

import (
	"context"
	"fmt"
	redis "github.com/go-redis/redis/v8"
	"strconv"
)

type BlockProgress struct {
	BlockHeight uint64 // 区块高度
	TxIndex     uint   // 当前区块中的交易索引
}

type Cache interface {
	GetLastParsedBlock(chain, protocol string) (BlockProgress, error)
	SetLastParsedBlock(chain, protocol string, progress BlockProgress) error
}

type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisCache(redisAddr string) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	return &RedisCache{
		client: client,
		ctx:    context.Background(),
	}
}

// 从Redis获取最后解析的区块进度
func (r *RedisCache) GetLastParsedBlock(chain, protocol string) (BlockProgress, error) {
	key := fmt.Sprintf("%s:%s", chain, protocol)
	result, err := r.client.HMGet(r.ctx, key, "blockHeight", "txIndex").Result()
	if err != nil || len(result) != 2 {
		return BlockProgress{BlockHeight: 0, TxIndex: 0}, err
	}

	blockHeight, _ := strconv.ParseUint(fmt.Sprintf("%v", result[0]), 10, 64)
	txIndex, _ := strconv.ParseUint(fmt.Sprintf("%v", result[1]), 10, 64)

	return BlockProgress{
		BlockHeight: blockHeight,
		TxIndex:     uint(txIndex),
	}, nil
}

// 将当前的解析进度存入Redis
func (r *RedisCache) SetLastParsedBlock(chain, protocol string, progress BlockProgress) error {
	key := fmt.Sprintf("%s:%s", chain, protocol)
	_, err := r.client.HMSet(r.ctx, key, map[string]interface{}{
		"blockHeight": progress.BlockHeight,
		"txIndex":     progress.TxIndex,
	}).Result()
	return err
}

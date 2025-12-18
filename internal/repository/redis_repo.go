package repository

import (
	"blockchain-asset-api/config"
	"blockchain-asset-api/internal/util"
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	_ "time"
)

var RedisClient *redis.Client
var ctx = context.Background()

// 初始化 Redis 客户端
func InitRedis() error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     config.Cfg.Redis.Addr,
		Password: config.Cfg.Redis.Password,
		DB:       config.Cfg.Redis.DB,
	})

	// 测试连接
	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("Redis 连接失败: %v", err)
	}
	util.Log.Info("Redis 客户端初始化成功")
	return nil
}

// 缓存 ETH 余额
func SetEthBalanceCache(address, balance string) error {
	key := fmt.Sprintf("eth:balance:%s", address)
	return RedisClient.Set(ctx, key, balance, config.Cfg.Redis.Expire).Err()
}

// 获取缓存的 ETH 余额
func GetEthBalanceCache(address string) (string, error) {
	key := fmt.Sprintf("eth:balance:%s", address)
	return RedisClient.Get(ctx, key).Result()
}

// 缓存 ERC20 代币余额
func SetErc20BalanceCache(address, contractAddress, balance string) error {
	key := fmt.Sprintf("erc20:balance:%s:%s", contractAddress, address)
	// 设置缓存有效期
	return RedisClient.Set(ctx, key, balance, config.Cfg.Redis.Expire).Err()
}

// 获取缓存的 ERC20 代币余额
func GetErc20BalanceCache(address, contractAddress string) (string, error) {
	key := fmt.Sprintf("erc20:balance:%s:%s", contractAddress, address)
	return RedisClient.Get(ctx, key).Result()
}

// 缓存区块信息
func SetBlockCache(blockNum string, blockData string) error {
	key := fmt.Sprintf("block:%s", blockNum)
	return RedisClient.Set(ctx, key, blockData, config.Cfg.Redis.Expire).Err()
}

// 获取缓存的区块信息
func GetBlockCache(blockNum string) (string, error) {
	key := fmt.Sprintf("block:%s", blockNum)
	return RedisClient.Get(ctx, key).Result()
}

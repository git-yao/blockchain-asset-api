package config

import (
	"github.com/spf13/viper"
	"log"
	"time"
)

type Config struct {
	Server ServerConfig
	Eth    EthConfig
	Redis  RedisConfig
	MySQL  MySQLConfig
}

type ServerConfig struct {
	Port    string
	Timeout time.Duration
}

type EthConfig struct {
	NodeURL string // 本地 GETH 节点：http://localhost:8545 或 Infura：https://mainnet.infura.io/v3/your-api-key
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	Expire   time.Duration // 缓存过期时间
}

type MySQLConfig struct {
	DSN string // 格式：user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local
}

var Cfg Config

// 初始化配置（读取本地 config.yaml，若不存在则用默认值）
func Init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	// 默认配置（你可根据实际环境修改）
	viper.SetDefault("server.port", ":8080")
	viper.SetDefault("server.timeout", 10*time.Second)
	viper.SetDefault("eth.nodeURL", "http://localhost:8545")
	viper.SetDefault("redis.addr", "127.0.0.1:6379")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.expire", 5*time.Minute)
	viper.SetDefault("mysql.dsn", "root:123456@tcp(127.0.0.1:3306)/blockchain_asset?charset=utf8mb4&parseTime=True&loc=Local")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("警告：未找到配置文件，使用默认配置: %v", err)
	}

	if err := viper.Unmarshal(&Cfg); err != nil {
		log.Fatalf("配置解析失败: %v", err)
	}
}

// @title Blockchain Asset API
// @version 1.0
// @description 区块链资产查询API
// @host localhost:8080
// @BasePath /api/v1
package main

import (
	_ "blockchain-asset-api/cmd/api/docs"
	"blockchain-asset-api/config"
	"blockchain-asset-api/internal/handler"
	"blockchain-asset-api/internal/repository"
	"blockchain-asset-api/internal/util"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"golang.org/x/time/rate"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func rateLimitMiddleware() gin.HandlerFunc {
	// 使用map存储每个IP的限流器
	ipLimiters := make(map[string]*rate.Limiter)
	var mutex sync.RWMutex

	return func(c *gin.Context) {
		// 获取客户端IP
		clientIP := c.ClientIP()

		// 获取或创建该IP对应的限流器
		mutex.Lock()
		limiter, exists := ipLimiters[clientIP]
		if !exists {
			// 为每个IP创建独立的限流器：每分钟100个请求
			limiter = rate.NewLimiter(rate.Every(time.Minute/100), 100)
			ipLimiters[clientIP] = limiter
		}
		mutex.Unlock()

		// 检查是否允许该请求
		if !limiter.Allow() {
			util.Log.Warnf("IP %s 请求过于频繁，请稍后再试", clientIP)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "请求过于频繁，请稍后再试",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func main() {
	// 1. 初始化配置
	config.Init()

	// 2. 初始化日志
	util.InitLog()

	// 3. 初始化依赖客户端
	if err := util.InitEthClient(config.Cfg.Eth.NodeURL); err != nil {
		util.Log.Fatalf("初始化以太坊客户端失败: %v", err)
	}
	if err := repository.InitRedis(); err != nil {
		util.Log.Fatalf("初始化 Redis 失败: %v", err)
	}
	if err := repository.InitMySQL(); err != nil {
		util.Log.Fatalf("初始化 MySQL 失败: %v", err)
	}

	// 4. 初始化 Gin 引擎
	r := gin.Default()

	// 5. 全局中间件：创建限流器：每分钟100个请求，桶容量100
	//limiter1 := rate.NewLimiter(rate.Every(time.Minute/100), 100)
	// 使用gin-contrib/ratelimit中间件包装限流器
	r.Use(rateLimitMiddleware())
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 提供静态文件服务（用于前端页面）
	r.Static("/static", "../../web/static")

	// 在 LoadHTMLGlob 前添加检查
	if _, err := os.Stat("../../web/templates"); err == nil {
		r.LoadHTMLGlob("../../web/templates/*")
	} else {
		util.Log.Warn("templates目录不存在，跳过HTML模板加载")
	}

	//printAllHTMLFiles()

	// 区块浏览器页面
	r.GET("/blocks", func(c *gin.Context) {
		c.HTML(http.StatusOK, "blocks.html", nil)
	})

	// 6. 路由注册
	v1 := r.Group("/api/v1")
	{

		//查询ETH余额
		v1.GET("/address/:addr/balance", GetEthBalanceHandler)

		//查询ERC20代币余额
		v1.GET("/address/:addr/tokens", GetErc20BalanceHandler)

		// 查询交易详情
		v1.GET("/transaction/:txhash", GetTransactionHandler)

		// 查询区块信息
		v1.GET("/block/:blocknum", GetBlockHandler)

		// 扫块
		v1.GET("/scan", ScanBlock)

		// 获取交易列表
		v1.GET("/transactions", handler.GetTransactionsHandler)

	}

	// 7. 启动服务
	util.Log.Infof("服务启动成功，监听端口: %s", config.Cfg.Server.Port)
	if err := r.Run(config.Cfg.Server.Port); err != nil {
		util.Log.Fatalf("服务启动失败: %v", err)
	}
}

// ScanBlockHandler godoc
// @Summary 扫描区块
// @Description 从指定区块开始扫描并将数据存储到数据库
// @Tags scan
// @Accept json
// @Produce json
// @Param from_block query int false "起始区块号（不指定则从数据库最新区块+1开始）"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /scan [get]
func ScanBlock(c *gin.Context) {
	handler.ScanBlock(c)
}

// @Summary 查询区块信息
// @Description 根据区块号查询区块详细信息
// @Tags block
// @Accept json
// @Produce json
// @Param blocknum path string true "区块号"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /block/{blocknum} [get]
func GetBlockHandler(c *gin.Context) {
	handler.GetBlockHandler(c)
}

// @Summary 查询交易详情
// @Description  根据交易哈希查询交易详细信息
// @Tags  transaction
// @Accept json
// @Produce json
// @Param txhash path string true "交易哈希"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /transaction/{txhash} [get]
func GetTransactionHandler(c *gin.Context) {
	handler.GetTransactionHandler(c)
}

// @Summary 查询ERC20代币余额
// @Description 根据地址和合约地址查询ERC20代币余额
// @Tags balance
// @Accept json
// @Produce json
// @Param addr path string true "以太坊地址"
// @Param contract query string true "ERC20合约地址"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /address/{addr}/tokens [get]
func GetErc20BalanceHandler(c *gin.Context) {
	handler.GetErc20BalanceHandler(c)
}

// @Summary 查询ETH余额
// @Description 根据以太坊地址查询ETH余额
// @Tags balance
// @accept json
// @Produce json
// @Param addr path string true "以太坊地址"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /address/{addr}/balance [get]
func GetEthBalanceHandler(c *gin.Context) {
	handler.GetEthBalanceHandler(c)
}

// 添加一个函数来扫描和打印HTML文件
// 改进的函数：扫描整个项目根目录下的所有HTML文件
func printAllHTMLFiles() {
	// 设置为项目根目录
	rootDir := "../../"

	// 检查根目录是否存在
	if _, err := os.Stat(rootDir); os.IsNotExist(err) {
		util.Log.Warnf("项目根目录不存在: %s", rootDir)
		return
	}

	util.Log.Info("开始扫描项目中的所有HTML文件...")

	// 使用Walk函数遍历目录
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// 记录无法访问的路径，但继续扫描其他路径
			util.Log.Debugf("无法访问路径 %s: %v", path, err)
			return nil
		}

		util.Log.Info("name:" + info.Name())
		// 检查是否为HTML文件
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".html") {
			// 获取相对路径
			relPath, _ := filepath.Rel(rootDir, path)
			util.Log.Infof("发现HTML文件: %s (完整路径: %s)", relPath, path)
		}

		return nil
	})

	if err != nil {
		util.Log.Errorf("扫描HTML文件时发生错误: %v", err)
	} else {
		util.Log.Info("HTML文件扫描完成")
	}
}

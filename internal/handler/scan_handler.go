package handler

import (
	"blockchain-asset-api/internal/service"
	"blockchain-asset-api/internal/util"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

var scanner *service.BlockScanner

// 初始化扫描器
func initScanner() {
	if scanner == nil {
		scanner = service.NewBlockScanner()
	}
}

// 扫描区块
func ScanBlock(c *gin.Context) {
	initScanner()

	// 获取起始区块号参数
	fromBlockStr := c.Query("from_block")
	var fromBlock int64 = 0

	if fromBlockStr != "" {
		var err error
		fromBlock, err = strconv.ParseInt(fromBlockStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "无效的起始区块号",
			})
			return
		}
	}

	// 启动扫描（在goroutine中执行以避免阻塞HTTP请求）
	go func() {
		if err := scanner.StartScan(fromBlock); err != nil {
			util.Log.Errorf("区块扫描失败: %v", err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"message":    "区块扫描已启动",
		"from_block": fromBlock,
	})
}

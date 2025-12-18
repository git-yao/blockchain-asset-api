package handler

import (
	"blockchain-asset-api/internal/service"
	"blockchain-asset-api/internal/util"
	"github.com/gin-gonic/gin"
	"net/http"
)

// Response 统一响应格式
type Response struct {
	Code    int         `json:"code"` // 0: 成功，非0: 失败
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// 成功响应
func success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// 失败响应
func fail(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
	})
}

// 查询ETH余额
func GetEthBalanceHandler(c *gin.Context) {
	address := c.Param("addr")
	if address == "" {
		fail(c, 400, "地址不能为空")
		return
	}

	balance, err := service.GetEthBalance(address)
	if err != nil {
		util.Log.Errorf("查询 ETH 余额失败: address=%s, err=%v", address, err)
		fail(c, 500, err.Error())
		return
	}

	success(c, gin.H{"address": address, "eth_balance": balance})
}

// 查询ERC20代币余额
func GetErc20BalanceHandler(c *gin.Context) {
	address := c.Param("addr")
	contractAddress := c.Query("contract") // 从查询参数获取合约地址
	if address == "" || contractAddress == "" {
		fail(c, 400, "地址和合约地址不能为空")
		return
	}

	balance, err := service.GetErc20Balance(address, contractAddress)
	if err != nil {
		util.Log.Errorf("查询 ERC20 余额失败: address=%s, contract=%s, err=%v", address, contractAddress, err)
		fail(c, 500, err.Error())
		return
	}

	success(c, gin.H{
		"address":          address,
		"contract_address": contractAddress,
		"token_balance":    balance,
	})
}

// 查询交易详情
func GetTransactionHandler(c *gin.Context) {
	txHash := c.Param("txhash")
	if txHash == "" {
		fail(c, 400, "交易哈希不能为空")
		return
	}

	detail, err := service.GetTransactionDetail(txHash)
	if err != nil {
		util.Log.Errorf("查询交易详情失败: txHash=%s, err=%v", txHash, err)
		fail(c, 500, err.Error())
		return
	}

	success(c, detail)
}

// 查询区块信息
func GetBlockHandler(c *gin.Context) {
	blockNum := c.Param("blocknum")
	if blockNum == "" {
		fail(c, 400, "区块号不能为空（支持 latest 或数字）")
		return
	}

	blockInfo, err := service.GetBlockInfo(blockNum)
	if err != nil {
		util.Log.Errorf("查询区块信息失败: blockNum=%s, err=%v", blockNum, err)
		fail(c, 500, err.Error())
		return
	}

	success(c, blockInfo)
}

package handler

import (
	"blockchain-asset-api/internal/service"
	"blockchain-asset-api/internal/util"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

// TransactionListResponse 交易列表响应
type TransactionListResponse struct {
	Transactions []TransactionResponse `json:"transactions"`
	Total        int64                 `json:"total"`
	Page         int                   `json:"page"`
	Pages        int                   `json:"pages"`
}

// TransactionResponse 交易响应
type TransactionResponse struct {
	ID          int64  `json:"id"`
	TxHash      string `json:"tx_hash"`
	BlockNumber int64  `json:"block_number"`
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
	Value       string `json:"value"`
	GasLimit    int64  `json:"gas_limit"`
	GasPrice    string `json:"gas_price"`
	GasUsed     *int64 `json:"gas_used"`
	TxType      string `json:"tx_type"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	ERC20Amount string `json:"erc20_amount"`
}

// GetTransactionsHandler godoc
// @Summary 获取交易列表
// @Description 获取扫描到的交易列表，支持分页和筛选
// @Tags transaction
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Param tx_type query string false "交易类型"
// @Param address query string false "地址筛选"
// @Param block_number query int false "区块号"
// @Success 200 {object} TransactionListResponse
// @Failure 500 {object} map[string]interface{}
// @Router /transactions [get]
func GetTransactionsHandler(c *gin.Context) {
	// 获取查询参数
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "10")
	txType := c.Query("tx_type")
	address := c.Query("address")
	blockNumberStr := c.Query("block_number")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 || size > 100 {
		size = 10
	}

	var blockNumber *int64
	if blockNumberStr != "" {
		if num, err := strconv.ParseInt(blockNumberStr, 10, 64); err == nil {
			blockNumber = &num
		}
	}

	// 调用服务获取数据
	transactions, total, err := service.GetTransactions(page, size, txType, address, blockNumber)
	if err != nil {
		util.Log.Errorf("获取交易列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取交易列表失败"})
		return
	}

	// 转换为响应格式
	var responseTxs []TransactionResponse
	for _, tx := range transactions {
		if tx.ERC20Amount != "" {
			tx.ERC20Amount = tx.ERC20Amount[:4]
		}
		responseTx := TransactionResponse{
			ID:          tx.ID,
			TxHash:      tx.TxHash,
			BlockNumber: tx.BlockNumber,
			FromAddress: tx.FromAddress,
			ToAddress:   tx.ToAddress,
			Value:       tx.Value[:4],
			GasLimit:    tx.GasLimit,
			GasPrice:    tx.GasPrice,
			GasUsed:     tx.GasUsed,
			TxType:      tx.TxType,
			Status:      tx.Status,
			CreatedAt:   tx.CreatedAt.Format("2006-01-02 15:04:05"),
			//判断不为空 保留4位小数
			ERC20Amount: tx.ERC20Amount,
		}
		responseTxs = append(responseTxs, responseTx)
	}

	pages := int((total + int64(size) - 1) / int64(size))

	c.JSON(http.StatusOK, TransactionListResponse{
		Transactions: responseTxs,
		Total:        total,
		Page:         page,
		Pages:        pages,
	})
}

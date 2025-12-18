package service

import (
	"blockchain-asset-api/internal/model"
	"blockchain-asset-api/internal/repository"
	"blockchain-asset-api/internal/util"
	"encoding/json"
	"fmt"
	_ "fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	_ "github.com/ethereum/go-ethereum/core/types"
	"log"
	"math/big"
	"strconv"
	"time"
)

// 查询 ETH 余额（优先查缓存，缓存未命中则查区块链）
func GetEthBalance(address string) (string, error) {
	// 1. 查缓存
	cacheBalance, err := repository.GetEthBalanceCache(address)
	if err == nil && cacheBalance != "" {
		util.Log.Infof("从缓存获取 ETH 余额: address=%s, balance=%s", address, cacheBalance)
		return cacheBalance, nil
	}

	// 2. 查区块链
	balance, err := util.GetEthBalance(address)
	if err != nil {
		return "", err
	}

	// 3. 写入缓存
	if err := repository.SetEthBalanceCache(address, balance); err != nil {
		util.Log.Warnf("缓存 ETH 余额失败: address=%s, err=%v", address, err)
	}

	// 4. 保存查询记录
	_ = repository.SaveQueryRecord(model.QueryRecord{
		Address:   address,
		QueryType: "eth_balance",
		CreatedAt: time.Now(),
	})

	return balance, nil
}

// 查询 ERC20 代币余额
func GetErc20Balance(address, contractAddress string) (string, error) {
	// 1. 查缓存
	cacheBalance, err := repository.GetErc20BalanceCache(address, contractAddress)
	if err == nil && cacheBalance != "" {
		util.Log.Infof("从缓存获取 ERC20 余额: address=%s, contract=%s, balance=%s", address, contractAddress, cacheBalance)
		return cacheBalance, nil
	}

	// 2. 查区块链
	balance, err := util.GetErc20Balance(address, contractAddress)
	if err != nil {
		return "", err
	}

	// 3. 写入缓存
	if err := repository.SetErc20BalanceCache(address, contractAddress, balance); err != nil {
		util.Log.Warnf("缓存 ERC20 余额失败: address=%s, contract=%s, err=%v", address, contractAddress, err)
	}

	// 4. 保存查询记录
	_ = repository.SaveQueryRecord(model.QueryRecord{
		Address:    address,
		QueryType:  "erc20_balance",
		QueryParam: contractAddress,
		CreatedAt:  time.Now(),
	})

	return balance, nil
}

// 查询交易详情（返回结构化数据）
type TransactionDetail struct {
	TxHash      string `json:"tx_hash"`
	From        string `json:"from"`
	To          string `json:"to"`
	Value       string `json:"value_eth"`
	GasUsed     uint64 `json:"gas_used"`
	GasPrice    string `json:"gas_price_gwei"`
	BlockNumber uint64 `json:"block_number"`
	Status      string `json:"status"` // success / failed
}

func GetTransactionDetail(txHash string) (*TransactionDetail, error) {
	tx, receipt, err := util.GetTransactionByHash(txHash)
	if err != nil {
		return nil, err
	}

	//signer := types.LatestSignerForChainID(nil)
	var signer types.Signer
	if tx.ChainId() != nil && tx.ChainId().Cmp(big.NewInt(0)) > 0 {
		// 如果交易包含链ID，使用对应链的签名者
		signer = types.LatestSignerForChainID(tx.ChainId())
	} else {
		// 否则使用默认的最新签名者
		signer = types.NewLondonSigner(tx.ChainId())
	}

	fromAddr, err := types.Sender(signer, tx)
	if err != nil {
		return nil, fmt.Errorf("无法恢复交易发送方地址: %v", err)
	}

	// 转换单位（Wei -> Gwei 用于 gasPrice）
	gasPriceGwei := new(big.Float).Quo(new(big.Float).SetInt(tx.GasPrice()), big.NewFloat(1e9))
	gasPriceStr, _ := gasPriceGwei.MarshalText()

	// 交易状态（receipt.Status == 1 表示成功）
	status := "failed"
	if receipt.Status == 1 {
		status = "success"
	}

	util.Log.Infof("交易详情: %+v", tx)
	// 解析LogDta
	for i, vlog := range receipt.Logs {
		if err != nil {
			log.Printf("日志序列化失败 - Index: %d, Error: %v", i, err)
			continue
		}
		var num string
		err = util.LogDtaUnpack(0, 64, &num, []byte(common.Bytes2Hex(vlog.Data)))
		if err != nil {
			log.Printf("LogDtaUnpack解析失败 - Index: %d, Error: %v", i, err)
			continue
		}
		atoi, err := strconv.ParseInt(num, 16, 64)
		if err != nil {
			log.Printf("数值转换失败 - Value: %s, Error: %v", num, err)
			continue
		}
		fmt.Println("TransactionToken === ", atoi)
	}

	detail := &TransactionDetail{
		TxHash:      tx.Hash().Hex(),
		From:        fromAddr.Hex(),
		To:          tx.To().Hex(),
		Value:       util.WeiToEth(tx.Value()),
		GasUsed:     receipt.GasUsed,
		GasPrice:    string(gasPriceStr),
		BlockNumber: receipt.BlockNumber.Uint64(),
		Status:      status,
	}

	// 保存查询记录
	_ = repository.SaveQueryRecord(model.QueryRecord{
		Address:    fromAddr.Hex(),
		QueryType:  "transaction",
		QueryParam: txHash,
		CreatedAt:  time.Now(),
	})

	return detail, nil
}

// 查询区块信息（返回结构化数据）
type BlockInfo struct {
	BlockNumber  uint64 `json:"block_number"`
	Hash         string `json:"hash"`
	Timestamp    string `json:"timestamp"`
	Transactions int    `json:"transactions_count"`
	GasUsed      uint64 `json:"gas_used"`
	GasLimit     uint64 `json:"gas_limit"`
	Miner        string `json:"miner"`
}

func GetBlockInfo(blockNum string) (*BlockInfo, error) {
	// 1. 查缓存（区块数据序列化后存储）
	cacheBlock, err := repository.GetBlockCache(blockNum)
	if err == nil && cacheBlock != "" {
		var blockInfo BlockInfo
		if err := json.Unmarshal([]byte(cacheBlock), &blockInfo); err == nil {
			util.Log.Infof("从缓存获取区块信息: blockNum=%s", blockNum)
			return &blockInfo, nil
		}
	}

	// 2. 查区块链
	block, err := util.GetBlockByNumber(blockNum)
	if err != nil {
		return nil, err
	}

	// 3. 结构化数据
	blockInfo := &BlockInfo{
		BlockNumber:  block.NumberU64(),
		Hash:         block.Hash().Hex(),
		Timestamp:    time.Unix(int64(block.Time()), 0).Format("2006-01-02 15:04:05"),
		Transactions: len(block.Transactions()),
		GasUsed:      block.GasUsed(),
		GasLimit:     block.GasLimit(),
		Miner:        block.Coinbase().Hex(),
	}

	// 4. 写入缓存（序列化后存储）
	blockData, _ := json.Marshal(blockInfo)
	if err := repository.SetBlockCache(blockNum, string(blockData)); err != nil {
		util.Log.Warnf("缓存区块信息失败: blockNum=%s, err=%v", blockNum, err)
	}

	// 5. 保存查询记录
	_ = repository.SaveQueryRecord(model.QueryRecord{
		Address:    "", // 区块查询无地址，留空
		QueryType:  "block",
		QueryParam: blockNum,
		CreatedAt:  time.Now(),
	})

	return blockInfo, nil
}

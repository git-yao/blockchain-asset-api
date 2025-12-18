package service

import (
	"blockchain-asset-api/internal/model"
	"blockchain-asset-api/internal/repository"
	"blockchain-asset-api/internal/util"
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
	"time"
)

const (
	TxTypeEthTransfer   = "eth_transfer"
	TxTypeERC20Transfer = "erc20_transfer"
	TxTypeContractCall  = "contract_call"
)

type BlockScanner struct {
	blockRepo *repository.BlockRepository
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewBlockScanner() *BlockScanner {
	ctx, cancel := context.WithCancel(context.Background())
	return &BlockScanner{
		blockRepo: repository.NewBlockRepository(),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// 开始扫描区块
func (s *BlockScanner) StartScan(fromBlock int64) error {
	latestBlockNumber, err := s.blockRepo.GetLatestBlockNumber()
	if err != nil {
		return fmt.Errorf("获取最新区块号失败: %v", err)
	}

	// 如果指定了起始区块，则使用指定的；否则从数据库中最新的区块+1开始
	startBlock := fromBlock
	if startBlock <= 0 {
		startBlock = latestBlockNumber + 1
	}

	// 获取当前最新区块
	header, err := util.EthClient.HeaderByNumber(s.ctx, nil)
	if err != nil {
		return fmt.Errorf("获取最新区块头失败: %v", err)
	}

	currentBlock := header.Number.Int64()

	util.Log.Infof("开始扫描区块: %d 到 %d", startBlock, currentBlock)

	for i := startBlock; i <= currentBlock; i++ {
		select {
		case <-s.ctx.Done():
			util.Log.Info("区块扫描已停止")
			return nil
		default:
			if err := s.scanBlock(i); err != nil {
				util.Log.Errorf("扫描区块 %d 失败: %v", i, err)
				continue
			}
			util.Log.Infof("成功扫描区块: %d", i)
			time.Sleep(100 * time.Millisecond) // 避免请求过于频繁
		}
	}

	util.Log.Info("区块扫描完成")
	return nil
}

// 扫描单个区块
func (s *BlockScanner) scanBlock(blockNumber int64) error {
	// 获取区块信息
	block, err := util.EthClient.BlockByNumber(context.Background(), big.NewInt(blockNumber))
	if err != nil {
		return fmt.Errorf("获取区块失败: %v", err)
	}

	// 保存区块信息
	blockModel := &model.Block{
		BlockNumber:       blockNumber,
		BlockHash:         block.Hash().Hex(),
		Timestamp:         time.Unix(int64(block.Time()), 0),
		TransactionsCount: len(block.Transactions()),
		GasUsed:           int64(block.GasUsed()),
		GasLimit:          int64(block.GasLimit()),
		Miner:             block.Coinbase().Hex(),
		CreatedAt:         time.Now(),
	}

	if err := s.blockRepo.SaveBlock(blockModel); err != nil {
		return fmt.Errorf("保存区块信息失败: %v", err)
	}

	// 处理区块中的交易
	for _, tx := range block.Transactions() {
		if err := s.processTransaction(tx, blockNumber); err != nil {
			util.Log.Errorf("处理交易 %s 失败: %v", tx.Hash().Hex(), err)
		}
	}

	return nil
}

// 处理交易
func (s *BlockScanner) processTransaction(tx *types.Transaction, blockNumber int64) error {
	// 获取交易回执
	receipt, err := util.EthClient.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		return fmt.Errorf("获取交易回执失败: %v", err)
	}

	// 恢复发送方地址
	signer := types.LatestSignerForChainID(tx.ChainId())
	fromAddr, err := types.Sender(signer, tx)
	if err != nil {
		return fmt.Errorf("恢复发送方地址失败: %v", err)
	}

	// 确定交易类型
	txType := s.determineTxType(tx, receipt)

	// 交易状态
	status := "failed"
	if receipt.Status == 1 {
		status = "success"
	}

	// 创建交易模型
	txModel := &model.Transaction{
		TxHash:      tx.Hash().Hex(),
		BlockNumber: blockNumber,
		FromAddress: fromAddr.Hex(),
		Value:       util.WeiToEth(tx.Value()),
		GasLimit:    int64(tx.Gas()),
		GasPrice:    util.WeiToEth(tx.GasPrice()),
		TxType:      txType,
		Status:      status,
		CreatedAt:   time.Now(),
	}

	// 设置接收方地址和Gas使用量
	if tx.To() != nil {
		txModel.ToAddress = tx.To().Hex()
	}

	gasUsed := int64(receipt.GasUsed)
	txModel.GasUsed = &gasUsed

	// 保存交易
	if err := s.blockRepo.SaveTransaction(txModel); err != nil {
		return fmt.Errorf("保存交易失败: %v", err)
	}

	// 如果是ERC20交易，处理Transfer事件
	if txType == TxTypeERC20Transfer {
		if err := s.processERC20Transfers(tx, receipt); err != nil {
			util.Log.Errorf("处理ERC20转移事件失败: %v", err)
		}
	}

	return nil
}

// 确定交易类型
func (s *BlockScanner) determineTxType(tx *types.Transaction, receipt *types.Receipt) string {
	// 如果有输入数据且不是简单的ETH转账，则可能是合约调用
	if len(tx.Data()) > 0 {
		// 检查是否是ERC20 Transfer事件
		for _, log := range receipt.Logs {
			if len(log.Topics) > 0 && log.Topics[0].Hex() == "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef" {
				// 这是ERC20 Transfer事件的签名哈希
				return TxTypeERC20Transfer
			}
		}
		return TxTypeContractCall
	}

	// 简单的ETH转账
	return TxTypeEthTransfer
}

// 处理ERC20转移事件
func (s *BlockScanner) processERC20Transfers(tx *types.Transaction, receipt *types.Receipt) error {
	// ERC20 Transfer事件的topic0签名哈希
	transferTopic := common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")

	for _, log := range receipt.Logs {
		// 检查是否是Transfer事件
		if len(log.Topics) == 3 && log.Topics[0] == transferTopic {
			// 解析Transfer事件数据
			fromAddr := common.BytesToAddress(log.Topics[1].Bytes()).Hex()
			toAddr := common.BytesToAddress(log.Topics[2].Bytes()).Hex()

			// 解析金额（value在data中）
			if len(log.Data) >= 32 {
				value := new(big.Int).SetBytes(log.Data[0:32])

				transfer := &model.ERC20Transfer{
					TxHash:          tx.Hash().Hex(),
					FromAddress:     fromAddr,
					ToAddress:       toAddr,
					ContractAddress: log.Address.Hex(),
					Amount:          value.String(),
					CreatedAt:       time.Now(),
				}

				if err := s.blockRepo.SaveERC20Transfer(transfer); err != nil {
					util.Log.Errorf("保存ERC20转移记录失败: %v", err)
				}
			}
		}
	}

	return nil
}

// 停止扫描
func (s *BlockScanner) Stop() {
	s.cancel()
}

package service

import (
	"blockchain-asset-api/internal/model"
	"blockchain-asset-api/internal/repository"
	"gorm.io/gorm"
	_ "gorm.io/gorm"
)

// GetTransactions 获取交易列表
func GetTransactions(page, size int, txType, address string, blockNumber *int64) ([]model.Transaction, int64, error) {
	db := repository.GetDB()
	var transactions []model.Transaction
	var total int64

	// 构建查询条件
	query := db.Model(&model.Transaction{})

	// 交易类型筛选
	if txType != "" {
		query = query.Where("tx_type = ?", txType)
	}

	// 地址筛选（发送方或接收方）
	if address != "" {
		query = query.Where("from_address = ? OR to_address = ?", address, address)
	}

	// 区块号筛选
	if blockNumber != nil {
		query = query.Where("block_number = ?", *blockNumber)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * size
	if err := query.Offset(offset).Limit(size).Order("id DESC").Find(&transactions).Error; err != nil {
		return nil, 0, err
	}

	// 收集所有ERC20转账交易的哈希
	erc20TxHashes := make([]string, 0)

	for i := range transactions {
		if transactions[i].TxType == "erc20_transfer" {
			erc20TxHashes = append(erc20TxHashes, transactions[i].TxHash)
		}
	}

	// 批量查询ERC20转账记录
	if len(erc20TxHashes) > 0 {
		var erc20Transfers []model.ERC20Transfer
		if err := db.Model(&model.ERC20Transfer{}).
			Where("tx_hash IN ?", erc20TxHashes).
			Find(&erc20Transfers).Error; err != nil && err != gorm.ErrRecordNotFound {
			return nil, 0, err
		}

		// 将ERC20转账金额关联到对应交易
		erc20TransferMap := make(map[string]string)
		for _, transfer := range erc20Transfers {
			erc20TransferMap[transfer.TxHash] = transfer.Amount
		}

		// 循环transactions为erc20_transfer类型的交易赋值Amount
		for i := range transactions {
			if transactions[i].TxType == "erc20_transfer" {
				if amount, exists := erc20TransferMap[transactions[i].TxHash]; exists && amount != "" {
					transactions[i].ERC20Amount = amount
				} else {
					transactions[i].ERC20Amount = "0.00000" // 空值时设为0
				}
			}
		}

	}

	return transactions, total, nil
}

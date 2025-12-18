package repository

import (
	"blockchain-asset-api/internal/model"
	"gorm.io/gorm"
)

type BlockRepository struct {
	db *gorm.DB
}

func NewBlockRepository() *BlockRepository {
	return &BlockRepository{db: GetDB()}
}

// 获取最新区块号
func (r *BlockRepository) GetLatestBlockNumber() (int64, error) {
	var block model.Block
	err := r.db.Order("block_number desc").First(&block).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil // 返回0表示从创世区块开始
		}
		return 0, err
	}
	return block.BlockNumber, nil
}

// 保存区块信息
func (r *BlockRepository) SaveBlock(block *model.Block) error {
	return r.db.Create(block).Error
}

// 保存交易信息
func (r *BlockRepository) SaveTransaction(tx *model.Transaction) error {
	return r.db.Create(tx).Error
}

// 保存ERC20转移记录
func (r *BlockRepository) SaveERC20Transfer(transfer *model.ERC20Transfer) error {
	return r.db.Create(transfer).Error
}

package model

import (
	"time"
)

// 查询记录模型
type QueryRecord struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Address    string    `gorm:"column:address;type:varchar(64);comment:查询地址" json:"address"`
	QueryType  string    `gorm:"column:query_type;type:varchar(32);comment:查询类型" json:"query_type"`
	QueryParam string    `gorm:"column:query_param;type:varchar(128);default:'';comment:查询参数" json:"query_param"`
	CreatedAt  time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
}

// 表名
func (QueryRecord) TableName() string {
	return "query_records"
}

// 区块模型
type Block struct {
	ID                int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	BlockNumber       int64     `gorm:"column:block_number;uniqueIndex" json:"block_number"`
	BlockHash         string    `gorm:"column:block_hash;type:varchar(66)" json:"block_hash"`
	Timestamp         time.Time `gorm:"column:timestamp" json:"timestamp"`
	TransactionsCount int       `gorm:"column:transactions_count" json:"transactions_count"`
	GasUsed           int64     `gorm:"column:gas_used" json:"gas_used"`
	GasLimit          int64     `gorm:"column:gas_limit" json:"gas_limit"`
	Miner             string    `gorm:"column:miner;type:varchar(42)" json:"miner"`
	CreatedAt         time.Time `gorm:"column:created_at" json:"-"`
}

// 表名
func (Block) TableName() string {
	return "blocks"
}

// 交易模型
type Transaction struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TxHash      string    `gorm:"column:tx_hash;type:varchar(66);uniqueIndex" json:"tx_hash"`
	BlockNumber int64     `gorm:"column:block_number;index" json:"block_number"`
	FromAddress string    `gorm:"column:from_address;type:varchar(42);index" json:"from_address"`
	ToAddress   string    `gorm:"column:to_address;type:varchar(42);index" json:"to_address"`
	Value       string    `gorm:"column:value;type:decimal(65,30)" json:"value"`
	GasLimit    int64     `gorm:"column:gas_limit" json:"gas_limit"`
	GasPrice    string    `gorm:"column:gas_price;type:decimal(65,30)" json:"gas_price"`
	GasUsed     *int64    `gorm:"column:gas_used" json:"gas_used"`
	TxType      string    `gorm:"column:tx_type;type:varchar(20)" json:"tx_type"`
	Status      string    `gorm:"column:status;type:varchar(10)" json:"status"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"-"`
	// 新增字段用于存储ERC20转账金额
	ERC20Amount string `gorm:"-" json:"erc20_amount"`
}

func (Transaction) TableName() string {
	return "transactions"
}

// ERC20代币转移模型
type ERC20Transfer struct {
	ID              int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TxHash          string    `gorm:"column:tx_hash;type:varchar(66);index" json:"tx_hash"`
	FromAddress     string    `gorm:"column:from_address;type:varchar(42)" json:"from_address"`
	ToAddress       string    `gorm:"column:to_address;type:varchar(42)" json:"to_address"`
	ContractAddress string    `gorm:"column:contract_address;type:varchar(42);index" json:"contract_address"`
	Amount          string    `gorm:"column:amount;type:decimal(65,30)" json:"amount"`
	CreatedAt       time.Time `gorm:"column:created_at" json:"-"`
}

func (ERC20Transfer) TableName() string {
	return "erc20_transfers"
}

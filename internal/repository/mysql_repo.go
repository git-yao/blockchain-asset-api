package repository

import (
	"blockchain-asset-api/config"
	"blockchain-asset-api/internal/model"
	"blockchain-asset-api/internal/util"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

var DB *gorm.DB

func GetDB() *gorm.DB {
	return DB
}

// 初始化 MySQL 客户端
func InitMySQL() error {
	var err error
	DB, err = gorm.Open(mysql.Open(config.Cfg.MySQL.DSN), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("MySQL 连接失败: %v", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("获取数据库实例失败: %v", err)
	}

	// 配置连接池
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	// 自动迁移数据库表
	err = DB.AutoMigrate(
		&model.QueryRecord{},
		&model.Block{},
		&model.Transaction{},
		&model.ERC20Transfer{},
	)
	if err != nil {
		return fmt.Errorf("数据库迁移失败: %v", err)
	}

	util.Log.Info("MySQL 客户端初始化成功")
	return nil
}

// 保存查询记录
func SaveQueryRecord(record model.QueryRecord) error {
	err := DB.Create(&record).Error
	if err != nil {
		util.Log.Errorf("保存查询记录失败: %v, record=%+v", err, record)
	}
	return err
}

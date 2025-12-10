package database

import (
	"fmt"
	"zk-sync-go-pool/internal/config"
	"zk-sync-go-pool/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB // 全局数据库连接对象

func InitMySQL(cfg *config.DatabaseConfig) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Dbname,
	)

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("数据库连接失败: %v", err)
	}

	// 自动迁移（创建表）
	err = db.AutoMigrate(
		&models.Pool{},
		&models.Token{},
		&models.SwapEvent{},
		&models.ScanProgress{},
	)
	if err != nil {
		return fmt.Errorf("表迁移失败: %v", err)
	}

	DB = db
	fmt.Println("MySQL连接成功")

	return nil
}

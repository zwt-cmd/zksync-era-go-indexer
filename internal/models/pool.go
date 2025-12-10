package models

import "time"

// 定义pool结构体

type Pool struct {
	ID             int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	PoolAddress    string     `gorm:"type:varchar(42);uniqueIndex;not null" json:"pool_address"`
	FactoryAddress string     `gorm:"type:varchar(42);not null" json:"factory_address"`
	PoolType       string     `gorm:"type:varchar(20);not null" json:"pool_type"`
	Version        string     `gorm:"type:varchar(10);not null" json:"version"`
	Token0         string     `gorm:"type:varchar(42);not null" json:"token0"`
	Token1         string     `gorm:"type:varchar(42);not null" json:"token1"`
	FeeRate        *int       `gorm:"type:int" json:"fee_rate"` // 可为空，用指针。因为我们不知道到底还是为空指针还是为0，所以用指针默认为空 兼容性好一些。
	CreatedAt      time.Time  `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	CreatedTx      string     `gorm:"type:varchar(66);not null" json:"created_tx"`
	CreatedBlock   uint64     `gorm:"type:bigint;not null" json:"created_block"`
	UpdatedAt      time.Time  `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updated_at"`
	DeletedAt      *time.Time `gorm:"type:timestamp" json:"deleted_at"` // 软删除字段，指针类型可为NULL
	Status         bool       `gorm:"type:boolean;default:true" json:"status"`
}

// TableName 指定表名
func (Pool) TableName() string {
	return "pools"
}

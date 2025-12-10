package models

import "time"

// 定义交换事件结构体

type SwapEvent struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	BlockNumber    uint64    `gorm:"type:bigint;not null" json:"block_number"`
	BlockTimeStamp int64     `gorm:"column:block_timestamp;type:bigint;not null" json:"block_timestamp"`
	TxHash         string    `gorm:"type:varchar(66);not null" json:"tx_hash"`
	LogIndex       int       `gorm:"type:int;not null" json:"log_index"`
	PoolAddress    string    `gorm:"type:varchar(42);not null" json:"pool_address"`
	Sender         string    `gorm:"type:varchar(42);not null" json:"sender"`
	Recipient      string    `gorm:"type:varchar(42);not null" json:"recipient"`
	TokenIn        string    `gorm:"type:varchar(42);not null" json:"token_in"`
	TokenOut       string    `gorm:"type:varchar(42);not null" json:"token_out"`
	AmountIn       string    `gorm:"type:varchar(78);not null" json:"amount_in"`
	AmountOut      string    `gorm:"type:varchar(78);not null" json:"amount_out"`
	FinalityStatus string    `gorm:"type:varchar(16);not null;default:'safe'" json:"finality_status"`
	CreatedAt      time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
}

// TableName 指定表名
func (SwapEvent) TableName() string {
	return "swap_events"
}

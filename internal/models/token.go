package models

import "time"

// 定义token结构体

type Token struct {
	ID        int        `gorm:"primaryKey;autoIncrement" json:"id"`
	Address   string     `gorm:"type:varchar(42);uniqueIndex;not null" json:"address"`
	Symbol    string     `gorm:"type:varchar(20);not null" json:"symbol"`
	Name      string     `gorm:"type:varchar(100);not null" json:"name"`
	Decimals  int        `gorm:"type:tinyint;not null" json:"decimals"`
	CreatedAt time.Time  `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time  `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updated_at"`
	DeletedAt *time.Time `gorm:"type:timestamp" json:"deleted_at"` // 软删除字段，指针类型可为NULL
	Status    bool       `gorm:"type:boolean;default:true" json:"status"`
}

// TableName 指定表名
func (Token) TableName() string {
	return "tokens"
}

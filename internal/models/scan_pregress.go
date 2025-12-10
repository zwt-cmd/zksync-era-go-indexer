package models

import "time"

// 定义扫描进度结构体

type ScanProgress struct {
	ID               int       `gorm:"primaryKey;autoIncrement" json:"id"`
	TaskName         string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"task_name"`
	LastScannedBlock uint64    `gorm:"type:bigint;not null" json:"last_scanned_block"`
	Status           string    `gorm:"type:varchar(20);default:'running'" json:"status"`
	ErrorMessage     string    `gorm:"type:text" json:"error_message"`
	CreatedAt        time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt        time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (ScanProgress) TableName() string {
	return "scan_progress"
}

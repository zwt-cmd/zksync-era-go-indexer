package repository

import (
	"errors"
	"fmt"
	"strings"
	"zk-sync-go-pool/internal/database"
	"zk-sync-go-pool/internal/models"

	"gorm.io/gorm"
)

type Repository struct {
}

// 创建Repository 仓库 专注于与数据库交互
// 就是写各种方法和调用各种方法，跟业务抽离出来。类似controller和service的关系。
func NewRepository() *Repository {
	return &Repository{}
}

// 获当前扫描进度
func (r *Repository) GetScanProgress(taskName string) (uint64, error) {
	var progress models.ScanProgress

	// 从数据库查询
	result := database.DB.Where("task_name = ?", taskName).First(&progress)

	if result.Error != nil {
		// 如果是"记录不存在"，返回 0（这不是错误）
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return 0, nil // ✅ 返回 0，表示首次运行
		}
		// 如果是其他错误（如数据库连接失败），才返回错误
		return 0, result.Error
	}

	return progress.LastScannedBlock, nil
}

// 初始化扫描进度
func (r *Repository) InitScanProgress(taskName string, startBlock uint64) error {
	progress := models.ScanProgress{
		TaskName:         taskName,
		LastScannedBlock: startBlock,
		Status:           "running",
	}
	// 插入数据库
	result := database.DB.Create(&progress)
	if result.Error != nil {
		return fmt.Errorf("初始化进度失败: %v", result.Error)
	}
	fmt.Printf("✅ 初始化进度记录: task=%s, block=%d\n", taskName, startBlock)
	return nil
}

// UpdateScanProgress 更新扫描进度
func (r *Repository) UpdateScanProgress(taskName string, blockNum uint64) error {
	result := database.DB.Model(&models.ScanProgress{}).
		Where("task_name = ?", taskName).
		Update("last_scanned_block", blockNum)

	if result.Error != nil {
		return fmt.Errorf("更新进度失败: %v", result.Error)
	}

	return nil
}

// 保存池子数据

func (r *Repository) SavePool(pool *models.Pool) error {
	result := database.DB.Create(pool)
	if result.Error != nil {
		// 利用UNIQUE索引，防止重复保存
		if strings.Contains(result.Error.Error(), "Duplicate entry") { // 如果数据库中已经存在该池子，则不进行保存，防止重复保存
			return nil
		}
		return fmt.Errorf("保存池子数据失败: %v", result.Error)
	}
	return nil
}

// 获取全部池子信息（用于初始化内存缓存）
func (s *Repository) GetAllPools() ([]*models.Pool, error) {
	var pool []*models.Pool
	result := database.DB.Find(&pool)
	if result.Error != nil {
		return nil, fmt.Errorf("获取全部池子信息失败: %v", result.Error)
	}
	return pool, nil

}

// 保存swap事件
func (s *Repository) SaveSwapEvent(swapEvent *models.SwapEvent) error {
	err := database.DB.Create(swapEvent).Error
	if err == nil {
		return nil
	}
	// 唯一约束冲突则更新
	if strings.Contains(err.Error(), "Duplicate entry") {
		return database.DB.Model(&models.SwapEvent{}).
			Where("tx_hash = ? AND log_index = ?", swapEvent.TxHash, swapEvent.LogIndex).
			Updates(map[string]interface{}{
				"block_number":    swapEvent.BlockNumber,
				"block_timestamp": swapEvent.BlockTimeStamp,
				"pool_address":    swapEvent.PoolAddress,
				"sender":          swapEvent.Sender,
				"recipient":       swapEvent.Recipient,
				"token_in":        swapEvent.TokenIn,
				"token_out":       swapEvent.TokenOut,
				"amount_in":       swapEvent.AmountIn,
				"amount_out":      swapEvent.AmountOut,
				"finality_status": swapEvent.FinalityStatus,
			}).Error
	}
	return nil

}

// 根据池子地址获取池子信息
func (s *Repository) GetPoolByAddress(poolAddress string) (*models.Pool, error) {
	var pool models.Pool
	result := database.DB.Where("pool_address = ?", poolAddress).First(&pool)
	if result.Error != nil {
		// 如果是"记录不存在"，返回 nil, nil（这不是错误）
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // ✅ 返回 nil，表示池子不存在
		}
		// 如果是其他错误（如数据库连接失败），才返回错误
		return nil, fmt.Errorf("根据池子地址获取池子信息失败: %v", result.Error)
	}
	return &pool, nil
}

/*
删除所有高度大于safe且状态为pending的swap事件
为什么要大于，不是小于safe呢？
假设上一次扫描到的safe高度是100，latest高度是110。入库的swap事件高度是101-110，状态都是pending。
这次扫描到的safe高度是105，latest高度是115。入库的swap事件高度是106-115，状态都是pending。
那么101-105的swap事件已经被确认了，状态应该改为safe，而106-110的swap事件仍然是pending状态。
所以我们需要删除所有高度大于105且状态为pending的swap事件，然后重新入库106-115的swap事件。
*/
func (r *Repository) DeletePendingAfter(safe uint64) error {
	return database.DB.
		Where("block_number > ? AND finality_status = ?", safe, "pending").
		Delete(&models.SwapEvent{}).Error
}

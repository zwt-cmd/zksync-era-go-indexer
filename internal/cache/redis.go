package cache

import (
	"fmt"
	"zk-sync-go-pool/internal/config"

	"github.com/go-redis/redis"
)

var RDB *redis.Client // 全局Redis连接对象

func InitRedis(cfg *config.RedisConfig) error {

	// 创建redis 客户端
	RDB = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.Db,
		PoolSize: cfg.PoolSize,
	})

	// 测试连接
	_, err := RDB.Ping().Result()
	if err != nil {
		return fmt.Errorf("Redis连接失败: %v", err)
	}
	fmt.Println("Redis连接成功")

	return nil
}

func CloseRedis() error {
	return RDB.Close()
}

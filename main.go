package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"zk-sync-go-pool/internal/abi"
	"zk-sync-go-pool/internal/blockchain"
	"zk-sync-go-pool/internal/cache"
	"zk-sync-go-pool/internal/config"
	"zk-sync-go-pool/internal/database"
	"zk-sync-go-pool/internal/repository"
	"zk-sync-go-pool/internal/scanner"
)

func main() {
	// 监听系统中断信号以优雅关闭
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 加载配置文件
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// 初始化数据库
	err = database.InitMySQL(&cfg.Database)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	// 初始化Redis
	err = cache.InitRedis(&cfg.Redis)
	if err != nil {
		log.Fatalf("初始化Redis失败: %v", err)
	}

	// 初始化ABI
	err = abi.DownloadABIs(&cfg.Abi)
	if err != nil {
		log.Fatalf("初始化ABI失败: %v", err)
	}

	// // 初始化区块链客户端
	err = blockchain.InitClient(&cfg.Blockchain)
	if err != nil {
		log.Fatalf("初始化区块链客户端失败: %v", err)
	}

	// 创建Repository 业务拆离，Repository层负责与数据库交互
	// 就是写各种方法和调用各种方法，跟业务抽离出来。类似controller和service的关系。
	repo := repository.NewRepository()

	// 创建Scanner 扫描器 专注于扫描事件和索引事件
	scanner := scanner.NewABIScanner(cfg, repo)

	// 启动扫描器
	if err := scanner.Start(ctx); err != nil {
		log.Fatal("扫描失败:", err)
	}

	fmt.Println("🎉 所有模块初始化成功！")

}

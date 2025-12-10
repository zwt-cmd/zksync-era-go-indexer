# SyncEra Indexer

SyncEra Indexer 是一个面向 zkSync Era 的 Web3 索引器，使用 Go 语言编写，强调「易上手」和「适合作为学习项目」。

项目通过连接 zkSync Era 的 RPC 节点，实时同步区块 / 交易 / 日志数据，解析常见 DeFi / DEX 合约事件（例如 SyncSwap），并将标准化后的结果写入 MySQL 和 Redis。代码结构尽量保持清晰、模块化，方便你阅读、调试和二次开发，用来理解区块链索引器的整体架构和实现方式。

如果你是：
- 想入门 Web3 后端 / 区块链数据索引；
- 想学习如何用 Go 连接 zkSync Era、消费区块 / 日志；
- 想为自己或团队搭建一套简单可扩展的链上数据服务，

都可以把这个仓库当作一个实践型脚手架，在此基础上继续扩展更多协议、更多指标和对外 API。



## 快速开始

### 1. 启动数据库

**Mac / Linux:**
```bash
make up
```

**Windows:**
```batch
start-windows.bat
```

### 2. 连接数据库

使用你的桌面工具连接：

**MySQL (Navicat):**
- 主机: `localhost`
- 端口: `3307`
- 用户: `scanner`
- 密码: `scannerpass`
- 数据库: `syncswap`

**Redis (Another Redis Desktop Manager):**
- 主机: `localhost`
- 端口: `6380`

### 3. 配置项目

```bash
cp config/config.yaml.example config/config.yaml
# 编辑 config.yaml，填入你的 RPC 地址
```

### 4. 运行程序

```bash
make deps  # 安装依赖
make run   # 运行
```

## 常用命令

```bash
make up      # 启动服务
make down    # 停止服务
make logs    # 查看日志
make db      # 进入 MySQL
make redis   # 进入 Redis
make clean   # 删除所有数据（危险）
```

## 端口说明

> 为避免与其他项目冲突，使用非默认端口：

- MySQL: `3307` (默认3306)
- Redis: `6380` (默认6379)

## 项目结构

```
scan-chain/
├── config/              # 配置文件
├── docker/              # Docker 配置
├── main.go              # 主程序
├── Makefile             # Mac/Linux 命令
└── start-windows.bat    # Windows 启动脚本
```

## 技术栈

- Go 1.24+
- MySQL 8.0
- Redis 7
- Docker

## License

MIT

## 联系作者

有工作或项目需求，可邮件联系：`austin.rate@foxmail.com`


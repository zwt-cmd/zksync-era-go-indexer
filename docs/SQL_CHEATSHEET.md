# SQL + Go 完全速查手册（前端转后端必看）

> 从前端角度理解MySQL数据类型和表设计，配合Go语言实战

## 📊 类型速查表（最常用的）

### 数字类型

| MySQL类型 | Go类型 | 前端类型 | 什么时候用 | 举例 |
|----------|-------|---------|-----------|------|
| `INT` | `int32` / `int` | `number` | 年龄、状态码、计数 | `age: 25` |
| `BIGINT` | `int64` / `uint64` | `number` | ID、时间戳、区块号 | `id: 123456789` |
| `VARCHAR(78)` | `string` | `string` | 超大数字（区块链） | `"1000000000000000000"` |
| `DECIMAL(10,2)` | `float64` | `number` | 价格、金额 | `price: 19.99` |
| `TINYINT` | `int8` / `uint8` | `number` | 很小的数字 | `decimals: 18` |

### 字符串类型

| MySQL类型 | Go类型 | 前端类型 | 什么时候用 | 举例 |
|----------|-------|---------|-----------|------|
| `VARCHAR(n)` | `string` | `string` | 短文本（知道长度） | `username: "alice"` |
| `TEXT` | `string` | `string` | 长文本（不知道长度） | 文章内容 |
| `CHAR(n)` | `string` | `string` | 固定长度 | 身份证号 |

### 时间类型

| MySQL类型 | Go类型 | 前端类型 | 什么时候用 | 举例 |
|----------|-------|---------|-----------|------|
| `TIMESTAMP` | `time.Time` | `Date` | 业务时间 | 创建时间、更新时间 |
| `BIGINT` | `int64` / `uint64` | `number` | 区块链时间 | Unix时间戳 |

### 布尔类型

| MySQL类型 | Go类型 | 前端类型 | 什么时候用 | 举例 |
|----------|-------|---------|-----------|------|
| `BOOLEAN` | `bool` | `boolean` | 开关状态 | `isActive: true` |

---

## 🎯 以太坊/区块链专用

| 用途 | MySQL类型 | Go类型 | 长度/范围 | 例子 |
|-----|----------|-------|----------|------|
| 以太坊地址 | `VARCHAR(42)` | `string` | 42字符 | `0x5aea5775959fbc2557cc8789bc1bf90a239d9a91` |
| 交易哈希 | `VARCHAR(66)` | `string` | 66字符 | `0x1234...cdef` |
| Wei金额 | `VARCHAR(78)` | `string` / `*big.Int` | 最大78位 | `1000000000000000000` |
| 区块号 | `BIGINT` | `uint64` | 0 ~ 2^64-1 | `18500000` |
| 区块时间 | `BIGINT` | `int64` / `uint64` | Unix秒 | `1698765432` |
| TokenID | `VARCHAR(78)` | `string` / `*big.Int` | 最大78位 | `123` |
| Log索引 | `INT` | `int32` / `uint` | 0 ~ 2^31-1 | `5` |

### Go语言特殊类型

```go
// 1. 以太坊地址
import "github.com/ethereum/go-ethereum/common"
address common.Address  // 自动处理0x格式

// 2. 大数（Wei、TokenID）
import "math/big"
amount *big.Int  // 可以存储任意大的数字

// 3. 区块哈希/交易哈希
txHash common.Hash  // 32字节的哈希值
```

---

## 📝 建表模板（直接抄）

### 模板1：基础信息表

```sql
CREATE TABLE 表名 (
    -- 1. 主键ID（必须）
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    
    -- 2. 核心字段
    name VARCHAR(100) NOT NULL COMMENT '名称',
    address VARCHAR(42) UNIQUE NOT NULL COMMENT '地址',
    
    -- 3. 状态字段
    status VARCHAR(20) DEFAULT 'active' COMMENT '状态',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    
    -- 4. 时间字段（标配）
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    -- 5. 索引（加速查询）
    INDEX idx_address (address),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='表说明';
```

### 模板2：事件记录表

```sql
CREATE TABLE 事件名_events (
    -- 1. 自增ID
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    
    -- 2. 区块信息
    block_number BIGINT NOT NULL COMMENT '区块号',
    block_timestamp BIGINT NOT NULL COMMENT '区块时间戳',
    
    -- 3. 交易信息
    tx_hash VARCHAR(66) NOT NULL COMMENT '交易哈希',
    log_index INT NOT NULL COMMENT '日志索引',
    
    -- 4. 业务字段
    sender VARCHAR(42) NOT NULL COMMENT '发送者',
    amount VARCHAR(78) NOT NULL COMMENT '金额',
    
    -- 5. 记录时间
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- 6. 唯一约束（防止重复）
    UNIQUE KEY uk_event (tx_hash, log_index),
    
    -- 7. 查询索引
    INDEX idx_block (block_number),
    INDEX idx_sender (sender),
    INDEX idx_time (block_timestamp)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='事件表';
```

---

## 🔧 字段修饰符（必会）

### NOT NULL vs NULL

```sql
-- NOT NULL：必填（像前端的required）
username VARCHAR(50) NOT NULL    -- 必须有值
email VARCHAR(100) NOT NULL      -- 必须有值

-- NULL：可选
nickname VARCHAR(50) NULL        -- 可以为空
phone VARCHAR(20)                -- 默认就是NULL
```

### DEFAULT（默认值）

```sql
-- 像前端的初始值
status VARCHAR(20) DEFAULT 'pending'              -- 默认pending
is_active BOOLEAN DEFAULT TRUE                    -- 默认true
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP    -- 默认当前时间
count INT DEFAULT 0                               -- 默认0
```

### UNIQUE（唯一）

```sql
-- 像前端的唯一性校验
email VARCHAR(100) UNIQUE        -- 邮箱不能重复
username VARCHAR(50) UNIQUE      -- 用户名不能重复
address VARCHAR(42) UNIQUE       -- 地址不能重复
```

### AUTO_INCREMENT（自增）

```sql
-- 像前端的自动生成ID
id BIGINT PRIMARY KEY AUTO_INCREMENT    -- 自动1,2,3,4...
```

### COMMENT（注释）

```sql
-- 给字段加说明
username VARCHAR(50) NOT NULL COMMENT '用户名'
amount VARCHAR(78) NOT NULL COMMENT '金额(wei)'
```

---

## 📑 索引（INDEX）速查

### 什么时候加索引？

```sql
-- ✅ 需要加索引的场景：
-- 1. 经常WHERE查询的字段
INDEX idx_username (username)        -- WHERE username = ?

-- 2. 经常JOIN的字段
INDEX idx_user_id (user_id)          -- JOIN ON user_id

-- 3. 经常排序的字段
INDEX idx_created_at (created_at)    -- ORDER BY created_at

-- 4. 外键字段
INDEX idx_pool_address (pool_address)
```

### 索引命名规则

```sql
-- 单字段索引
INDEX idx_字段名 (字段名)
INDEX idx_username (username)
INDEX idx_email (email)

-- 多字段索引（联合索引）
INDEX idx_字段1_字段2 (字段1, 字段2)
INDEX idx_token0_token1 (token0, token1)

-- 唯一索引
UNIQUE KEY uk_字段名 (字段名)
UNIQUE KEY uk_address (address)
```

---

## 💡 实战：4张表详解

### 表1：pools（池子信息）

#### SQL定义

```sql
CREATE TABLE pools (
    -- ID：自增主键
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    
    -- 地址：42字符，唯一，不能为空
    pool_address VARCHAR(42) UNIQUE NOT NULL COMMENT '池子地址',
    
    -- 工厂：创建这个池子的工厂地址
    factory_address VARCHAR(42) NOT NULL COMMENT '工厂地址',
    
    -- 类型：classic/stable/range等
    pool_type VARCHAR(20) NOT NULL COMMENT '池子类型',
    
    -- 版本：v1/v2/v2.1/v3
    version VARCHAR(10) NOT NULL COMMENT '版本',
    
    -- Token对：组成池子的两个代币
    token0 VARCHAR(42) NOT NULL COMMENT 'Token0地址',
    token1 VARCHAR(42) NOT NULL COMMENT 'Token1地址',
    
    -- 创建信息
    created_block BIGINT NOT NULL COMMENT '创建区块号',
    created_tx VARCHAR(66) NOT NULL COMMENT '创建交易',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- 索引：加速查询
    INDEX idx_tokens (token0, token1),       -- 查询某个代币对
    INDEX idx_type (pool_type, version),     -- 查询某种类型
    INDEX idx_factory (factory_address)      -- 查询某个工厂的池子
);
```

#### Go结构体

```go
type Pool struct {
    ID             uint64    `gorm:"primaryKey;autoIncrement"`
    PoolAddress    string    `gorm:"type:varchar(42);uniqueIndex;not null"`
    FactoryAddress string    `gorm:"type:varchar(42);not null"`
    PoolType       string    `gorm:"type:varchar(20);not null"`
    Version        string    `gorm:"type:varchar(10);not null"`
    Token0         string    `gorm:"type:varchar(42);not null"`
    Token1         string    `gorm:"type:varchar(42);not null"`
    CreatedBlock   uint64    `gorm:"not null"`
    CreatedTx      string    `gorm:"type:varchar(66);not null"`
    CreatedAt      time.Time `gorm:"autoCreateTime"`
}
```

**为什么这么设计？**
- `pool_address` 用UNIQUE：一个地址只能是一个池子
- `token0/token1` 用VARCHAR(42)：以太坊地址固定长度
- Go中用 `uint64`：因为ID和区块号不会是负数
- 加索引在`tokens`：因为经常查"有哪些USDC的池子"

### 表2：tokens（代币信息）

#### SQL定义

```sql
CREATE TABLE tokens (
    id INT PRIMARY KEY AUTO_INCREMENT,
    
    -- 地址：唯一标识
    address VARCHAR(42) UNIQUE NOT NULL COMMENT 'Token地址',
    
    -- 基础信息：从合约读取
    symbol VARCHAR(20) COMMENT '符号(ETH,USDC)',
    name VARCHAR(100) COMMENT '名称(Wrapped Ether)',
    decimals TINYINT COMMENT '精度(18,6,8)',
    
    -- 时间
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_symbol (symbol)
);
```

#### Go结构体

```go
type Token struct {
    ID        uint      `gorm:"primaryKey;autoIncrement"`
    Address   string    `gorm:"type:varchar(42);uniqueIndex;not null"`
    Symbol    string    `gorm:"type:varchar(20)"`
    Name      string    `gorm:"type:varchar(100)"`
    Decimals  uint8     `gorm:"type:tinyint"`  // uint8 = 0-255
    CreatedAt time.Time `gorm:"autoCreateTime"`
    UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
```

**为什么需要decimals？**
```go
// Go中转换金额
amountWei := "1000000000000000000"  // 从数据库读取
decimals := uint8(18)                // 从tokens表读取

// 转换为 big.Int
amount := new(big.Int)
amount.SetString(amountWei, 10)

// 计算实际金额
divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
realAmount := new(big.Float).Quo(new(big.Float).SetInt(amount), new(big.Float).SetInt(divisor))
// 结果: 1.0
```

### 表3：swap_events（交易事件）

#### SQL定义

```sql
CREATE TABLE swap_events (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    
    -- 区块信息
    block_number BIGINT NOT NULL COMMENT '区块号',
    block_timestamp BIGINT NOT NULL COMMENT '区块时间',
    
    -- 交易信息
    tx_hash VARCHAR(66) NOT NULL COMMENT '交易哈希',
    log_index INT NOT NULL COMMENT '日志索引',
    
    -- 池子和用户
    pool_address VARCHAR(42) NOT NULL COMMENT '池子地址',
    sender VARCHAR(42) NOT NULL COMMENT '发送者',
    recipient VARCHAR(42) COMMENT '接收者',
    
    -- 交易详情
    token_in VARCHAR(42) NOT NULL COMMENT '输入代币',
    token_out VARCHAR(42) NOT NULL COMMENT '输出代币',
    amount_in VARCHAR(78) NOT NULL COMMENT '输入数量(wei)',
    amount_out VARCHAR(78) NOT NULL COMMENT '输出数量(wei)',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- 防止重复：同一个交易的同一个日志只记录一次
    UNIQUE KEY uk_event (tx_hash, log_index),
    
    -- 查询优化
    INDEX idx_block (block_number),          -- 按区块查
    INDEX idx_pool (pool_address),           -- 按池子查
    INDEX idx_sender (sender),               -- 按用户查
    INDEX idx_tokens (token_in, token_out),  -- 按代币对查
    INDEX idx_time (block_timestamp)         -- 按时间查
);
```

#### Go结构体

```go
type SwapEvent struct {
    ID             uint64    `gorm:"primaryKey;autoIncrement"`
    BlockNumber    uint64    `gorm:"not null;index:idx_block"`
    BlockTimestamp int64     `gorm:"not null;index:idx_time"`
    TxHash         string    `gorm:"type:varchar(66);not null;uniqueIndex:uk_event"`
    LogIndex       uint      `gorm:"not null;uniqueIndex:uk_event"`
    PoolAddress    string    `gorm:"type:varchar(42);not null;index:idx_pool"`
    Sender         string    `gorm:"type:varchar(42);not null;index:idx_sender"`
    Recipient      string    `gorm:"type:varchar(42)"`
    TokenIn        string    `gorm:"type:varchar(42);not null;index:idx_tokens"`
    TokenOut       string    `gorm:"type:varchar(42);not null;index:idx_tokens"`
    AmountIn       string    `gorm:"type:varchar(78);not null"`  // 字符串存储大数
    AmountOut      string    `gorm:"type:varchar(78);not null"`
    CreatedAt      time.Time `gorm:"autoCreateTime"`
}
```

**为什么amount用VARCHAR(78)？**
```
以太坊最大值：2^256 - 1
换成十进制：约78位数字
BIGINT最大值：2^63 - 1（只有19位）
所以：用VARCHAR(78)字符串存储

// Go中处理大数
import "math/big"
amountBig := new(big.Int)
amountBig.SetString(swapEvent.AmountIn, 10)
```

### 表4：scan_progress（扫描进度）

#### SQL定义

```sql
CREATE TABLE scan_progress (
    id INT PRIMARY KEY AUTO_INCREMENT,
    
    -- 任务名称：唯一标识
    task_name VARCHAR(50) UNIQUE NOT NULL COMMENT '任务名',
    
    -- 进度：最后扫到哪个区块
    last_scanned_block BIGINT NOT NULL COMMENT '最后区块',
    
    -- 状态：运行中/暂停/错误
    status VARCHAR(20) DEFAULT 'running' COMMENT '状态',
    error_message TEXT COMMENT '错误信息',
    
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 初始化数据
INSERT INTO scan_progress (task_name, last_scanned_block) VALUES 
('factory_scan', 40000000),        -- 扫描工厂到4000万区块
('pool_events_scan', 40000000);    -- 扫描事件到4000万区块
```

#### Go结构体

```go
type ScanProgress struct {
    ID               uint      `gorm:"primaryKey;autoIncrement"`
    TaskName         string    `gorm:"type:varchar(50);uniqueIndex;not null"`
    LastScannedBlock uint64    `gorm:"not null"`
    Status           string    `gorm:"type:varchar(20);default:running"`
    ErrorMessage     string    `gorm:"type:text"`
    UpdatedAt        time.Time `gorm:"autoUpdateTime"`
}

// 使用示例
func GetLastBlock(taskName string) uint64 {
    var progress ScanProgress
    db.Where("task_name = ?", taskName).First(&progress)
    return progress.LastScannedBlock
}

func UpdateProgress(taskName string, blockNumber uint64) {
    db.Model(&ScanProgress{}).
        Where("task_name = ?", taskName).
        Update("last_scanned_block", blockNumber)
}
```

**为什么需要这个表？**
```go
// 程序重启时从上次的位置继续
lastBlock := GetLastBlock("factory_scan")
// 从 lastBlock + 1 继续扫描，不会重复

// 扫描完成后更新进度
UpdateProgress("factory_scan", currentBlock)
```

---

## 🔧 GORM 标签详解（Go专用）

### 常用标签速查

| 标签 | 作用 | 示例 |
|------|------|------|
| `primaryKey` | 主键 | `gorm:"primaryKey"` |
| `autoIncrement` | 自增 | `gorm:"autoIncrement"` |
| `not null` | 不能为空 | `gorm:"not null"` |
| `uniqueIndex` | 唯一索引 | `gorm:"uniqueIndex"` |
| `index` | 普通索引 | `gorm:"index:idx_name"` |
| `type:varchar(42)` | 指定类型 | `gorm:"type:varchar(42)"` |
| `default:value` | 默认值 | `gorm:"default:0"` |
| `autoCreateTime` | 自动创建时间 | `gorm:"autoCreateTime"` |
| `autoUpdateTime` | 自动更新时间 | `gorm:"autoUpdateTime"` |

### 标签组合使用

```go
// 多个标签用分号分隔
PoolAddress string `gorm:"type:varchar(42);uniqueIndex;not null" json:"pool_address"`

// 联合索引
Token0 string `gorm:"type:varchar(42);index:idx_tokens,priority:1"`
Token1 string `gorm:"type:varchar(42);index:idx_tokens,priority:2"`

// 联合唯一约束
TxHash   string `gorm:"uniqueIndex:uk_event"`
LogIndex uint   `gorm:"uniqueIndex:uk_event"`
```

---

## 💻 Go代码实战

### 1. 数据库连接

```go
import (
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

func InitDB(dsn string) (*gorm.DB, error) {
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        return nil, err
    }
    
    // 自动迁移（创建表）
    db.AutoMigrate(&Pool{}, &Token{}, &SwapEvent{}, &ScanProgress{})
    
    return db, nil
}

// DSN格式
// "scanner:scannerpass@tcp(localhost:3307)/syncswap?charset=utf8mb4&parseTime=True&loc=Local"
```

### 2. 插入数据

```go
// 插入单条
pool := Pool{
    PoolAddress:    "0x123...",
    FactoryAddress: "0xf2D...",
    PoolType:       "classic",
    Version:        "v1",
    Token0:         "0x5ae...",
    Token1:         "0x335...",
    CreatedBlock:   18500000,
    CreatedTx:      "0xabc...",
}
db.Create(&pool)

// 批量插入
events := []SwapEvent{event1, event2, event3}
db.CreateInBatches(events, 100)  // 每批100条
```

### 3. 查询数据

```go
// 根据ID查询
var pool Pool
db.First(&pool, 1)  // WHERE id = 1

// 根据条件查询
var pools []Pool
db.Where("pool_type = ?", "classic").Find(&pools)

// 复杂查询
db.Where("token0 = ? OR token1 = ?", tokenAddr, tokenAddr).
   Order("created_block DESC").
   Limit(10).
   Find(&pools)
```

### 4. 更新数据

```go
// 更新单个字段
db.Model(&ScanProgress{}).
   Where("task_name = ?", "factory_scan").
   Update("last_scanned_block", 18500000)

// 更新多个字段
db.Model(&pool).Updates(Pool{
    PoolType: "stable",
    Version:  "v2",
})
```

### 5. 事务处理

```go
err := db.Transaction(func(tx *gorm.DB) error {
    // 1. 插入Pool
    if err := tx.Create(&pool).Error; err != nil {
        return err
    }
    
    // 2. 插入Events
    if err := tx.CreateInBatches(events, 100).Error; err != nil {
        return err
    }
    
    // 3. 更新Progress
    if err := tx.Model(&ScanProgress{}).
        Where("task_name = ?", "pool_scan").
        Update("last_scanned_block", currentBlock).Error; err != nil {
        return err
    }
    
    return nil
})
```

### 6. 类型转换工具函数

```go
package utils

import (
    "math/big"
    "github.com/ethereum/go-ethereum/common"
)

// 地址转换
func AddressToString(addr common.Address) string {
    return addr.Hex()  // 0x...
}

func StringToAddress(s string) common.Address {
    return common.HexToAddress(s)
}

// 哈希转换
func HashToString(hash common.Hash) string {
    return hash.Hex()
}

// 大数转换
func BigIntToString(b *big.Int) string {
    if b == nil {
        return "0"
    }
    return b.String()
}

func StringToBigInt(s string) *big.Int {
    b := new(big.Int)
    b.SetString(s, 10)
    return b
}

// Wei转实际金额
func WeiToEther(wei *big.Int) *big.Float {
    ether := new(big.Float)
    ether.SetString(wei.String())
    return ether.Quo(ether, big.NewFloat(1e18))
}
```

---

## 🎨 完整示例：创建一个表

假设要创建一个"用户NFT持仓表"：

```sql
-- 第1步：想清楚要存什么
-- - 用户地址
-- - NFT合约地址
-- - TokenID
-- - 数量
-- - 获取时间

-- 第2步：选择类型
CREATE TABLE user_nfts (
    -- ID
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    
    -- 用户地址（42字符）
    user_address VARCHAR(42) NOT NULL COMMENT '用户地址',
    
    -- NFT合约（42字符）
    nft_contract VARCHAR(42) NOT NULL COMMENT 'NFT合约',
    
    -- TokenID（可能很大）
    token_id VARCHAR(78) NOT NULL COMMENT 'TokenID',
    
    -- 数量（ERC1155可以有多个）
    balance INT DEFAULT 1 COMMENT '持有数量',
    
    -- 获取信息
    acquired_block BIGINT NOT NULL COMMENT '获取区块',
    acquired_tx VARCHAR(66) NOT NULL COMMENT '获取交易',
    
    -- 时间
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- 第3步：加索引
    INDEX idx_user (user_address),                    -- 查某用户的NFT
    INDEX idx_nft (nft_contract, token_id),          -- 查某个NFT的持有者
    UNIQUE KEY uk_holding (user_address, nft_contract, token_id)  -- 防止重复
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户NFT持仓';
```

---

## ⚡ 常见错误

### ❌ 错误1：类型选择不当

```sql
-- 错误：区块号用INT
block_number INT          -- 会溢出！最大21亿

-- 正确：用BIGINT
block_number BIGINT       -- 可以到900万亿
```

### ❌ 错误2：忘记加索引

```sql
-- 慢查询：没有索引
SELECT * FROM swap_events WHERE pool_address = '0x...';  -- 全表扫描

-- 快查询：有索引
CREATE INDEX idx_pool ON swap_events(pool_address);      -- 秒查
```

### ❌ 错误3：字符串长度不够

```sql
-- 错误：以太坊地址不是40
address VARCHAR(40)       -- 少了'0x'前缀！

-- 正确：42字符
address VARCHAR(42)       -- 0x + 40位 = 42
```

### ❌ 错误4：忘记COMMENT

```sql
-- 难懂
amount VARCHAR(78)

-- 清楚
amount VARCHAR(78) COMMENT '金额(wei单位)'
```

---

## 📚 学习路径

### 第1天：理解类型
- [ ] 看懂VARCHAR vs TEXT
- [ ] 看懂INT vs BIGINT
- [ ] 看懂TIMESTAMP vs BIGINT

### 第2天：写基础表
- [ ] 抄模板创建一个表
- [ ] 理解每个字段的作用
- [ ] 加上合适的索引

### 第3天：优化表结构
- [ ] 学习什么时候加索引
- [ ] 理解UNIQUE的作用
- [ ] 会用DEFAULT设置默认值

---

## 🎯 总结：记住这些就够了

### SQL类型选择口诀
```
小数字用INT，大数字用BIGINT
字符串短用VARCHAR，长用TEXT
以太坊地址42，交易哈希66
Wei金额字符串，长度要78
时间戳用BIGINT，创建时间TIMESTAMP
```

### Go类型选择口诀
```
ID和区块号 → uint64（不会负数）
时间戳 → int64（可能负数）
地址和哈希 → string（存数据库）
大数Wei → string + *big.Int（转换）
小数字 → uint8（decimals）
时间 → time.Time（自动处理）
```

### 建表三板斧
```sql
-- 1. 主键ID
id BIGINT PRIMARY KEY AUTO_INCREMENT

-- 2. 核心字段 + NOT NULL + COMMENT
字段名 类型 NOT NULL COMMENT '说明'

-- 3. 时间字段
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
```

```go
// Go对应
type Model struct {
    ID        uint64    `gorm:"primaryKey;autoIncrement"`
    Field     string    `gorm:"type:varchar(42);not null"`
    CreatedAt time.Time `gorm:"autoCreateTime"`
}
```

### 索引加在哪
```
WHERE用到的字段 → 加索引
JOIN用到的字段 → 加索引
ORDER BY的字段 → 加索引
```

---

## 📚 快速参考

### MySQL → Go 类型映射

| MySQL | Go | 说明 |
|-------|-----|------|
| `INT` | `int32` / `int` | 小整数 |
| `BIGINT` | `uint64` / `int64` | 大整数 |
| `TINYINT` | `uint8` | 0-255 |
| `VARCHAR(n)` | `string` | 字符串 |
| `TEXT` | `string` | 长文本 |
| `TIMESTAMP` | `time.Time` | 时间 |
| `BOOLEAN` | `bool` | 布尔 |

### 以太坊专用

```go
import (
    "math/big"
    "github.com/ethereum/go-ethereum/common"
)

// 地址
addr := common.HexToAddress("0x...")
addrStr := addr.Hex()

// 大数
amount := new(big.Int)
amount.SetString("1000000000000000000", 10)
```

---

现在你有了**SQL+Go双语速查手册**，对照着 `scripts/init_tables.sql` 就能完全看懂并写出Go代码了！🎉


USE syncswap;

-- ========================================
-- SyncSwap 扫链数据库初始化
-- ========================================

USE syncswap;

-- 1. 池子信息表
CREATE TABLE IF NOT EXISTS pools (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    pool_address VARCHAR(42) UNIQUE NOT NULL COMMENT '池子地址',
    factory_address VARCHAR(42) NOT NULL COMMENT '工厂地址(基于哪个合约的池子)',
    pool_type VARCHAR(20) NOT NULL COMMENT '池子类型(classic/stable/range)',
    version VARCHAR(10) NOT NULL COMMENT '版本(v1/v2/v2.1/v3)',
    token0 VARCHAR(42) NOT NULL COMMENT 'token0地址(池子的对币地址，固定不会变)',
    token1 VARCHAR(42) NOT NULL COMMENT 'token1地址(池子的对币地址，固定不会变)',
    fee_rate INT COMMENT '滑点(30 基点 = 0.3% 、5 基点 = 0.05%。可为空，因为有些池子可能没有固定费率或者需要动态获取)',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    created_tx VARCHAR(66) NOT NULL COMMENT '创建交易哈希(创建池子也算交易，保证唯一性)',
    created_block BIGINT NOT NULL COMMENT '创建区块号',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at TIMESTAMP NULL COMMENT '删除时间',
    status BOOLEAN DEFAULT TRUE COMMENT '状态',

    INDEX idx_tokens (token0 , token1), -- 按照代币0地址和代币1地址查询
    INDEX idx_pool_type (pool_type , version), -- 按照类型和版本查询
    INDEX idx_factory_address (factory_address) -- 按照工厂地址查询
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='池子信息表';



CREATE TABLE IF NOT EXISTS tokens(
    id INT PRIMARY KEY AUTO_INCREMENT,
    address VARCHAR(42) UNIQUE NOT NULL COMMENT '代币地址',
    symbol varchar(20) NOT NULL COMMENT '代币符号',
    name VARCHAR(100) NOT NULL COMMENT '代币名称', 
    decimals TINYINT NOT NULL COMMENT '代币精度',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at TIMESTAMP NULL COMMENT '删除时间',
    status BOOLEAN DEFAULT TRUE COMMENT '状态',
 
    INDEX idx_symbol (symbol) -- 按照代币符号查询
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='代币信息表';


CREATE TABLE IF NOT EXISTS swap_events(
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    block_number BIGINT NOT NULL COMMENT '区块高度',
    block_timestamp BIGINT NOT NULL COMMENT '区块时间戳(Unix秒)',
    tx_hash VARCHAR(66) NOT NULL COMMENT '交易哈希',
    log_index INT NOT NULL COMMENT '日志索引（同一交易可能有多条日志，加上tx_hash和log_index联合唯一索引区分日志; 触发多个事件:Transfer->Transfer->Swap->Transfer->Transfer）',
    pool_address VARCHAR(42) NOT NULL COMMENT '池子地址',
    sender VARCHAR(42) NOT NULL COMMENT '发送者',
    recipient VARCHAR(42) NOT NULL COMMENT '接收者',
    token_in VARCHAR(42) NOT NULL COMMENT '输入代币地址(比如WETH,USDC,USDT,WBTC的地址)',
    token_out VARCHAR(42) NOT NULL COMMENT '输出代币地址(比如WETH,USDC,USDT,WBTC的地址)',
    amount_in VARCHAR(78) NOT NULL COMMENT '输入数量(Wei,字符串)',
    amount_out VARCHAR(78) NOT NULL COMMENT '输出数量(Wei,字符串)',
    finality_status VARCHAR(16) NOT NULL DEFAULT "safe" COMMENT '最终状态(pending/safe)',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',

    INDEX idx_block_number (block_number), -- 按照区块高度查询
    UNIQUE idx_tx_event (tx_hash , log_index), -- 交易哈希加日志索引联合唯一索引 防止重复记录
    INDEX idx_pool_address (pool_address), -- 按照池子地址查询
    INDEX idx_sender (sender), -- 按照发送者查询
    INDEX idx_recipient (recipient), -- 按照接收者查询
    INDEX idx_tokens (token_in , token_out), -- 按照输入代币地址和输出代币地址查询
    INDEX idx_time (block_timestamp) -- 按照区块时间戳查询
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='交易事件表';


CREATE TABLE IF NOT EXISTS scan_progress(
    id INT PRIMARY KEY AUTO_INCREMENT,
    task_name VARCHAR(50) UNIQUE NOT NULL COMMENT '任务名称',
    last_scanned_block BIGINT NOT NULL COMMENT '最后扫描的区块高度',
    status VARCHAR(20) DEFAULT 'running' COMMENT '状态',
    error_message TEXT COMMENT '错误信息',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='扫描进度表';



-- 预填充常用Token（zkSync Era主网）
INSERT IGNORE INTO tokens (address, symbol, name, decimals) VALUES
('0x5aea5775959fbc2557cc8789bc1bf90a239d9a91', 'WETH', 'Wrapped Ether', 18),
('0x3355df6D4c9C3035724Fd0e3914dE96A5a83aaf4', 'USDC', 'USD Coin', 6),
('0x493257fD37EDB34451f62EDf8D2a0C418852bA4C', 'USDT', 'Tether USD', 6),
('0xBBeB516fb02a01611cBBE0453Fe3c580D7281011', 'WBTC', 'Wrapped BTC', 8);

-- 查看表创建结果
SHOW TABLES;
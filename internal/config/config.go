package config

// Config全局配置,后面通用映射config.yaml文件赋值给结构体
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`     // 服务配置
	Blockchain BlockchainConfig `mapstructure:"blockchain"` // 区块链配置
	Syncswap   SyncswapConfig   `mapstructure:"syncswap"`   // Syncswap配置
	Scanner    ScannerConfig    `mapstructure:"scanner"`    // 扫描器配置
	Abi        AbiConfig        `mapstructure:"abi"`        // ABI配置
	Database   DatabaseConfig   `mapstructure:"database"`   // 数据库配置
	Redis      RedisConfig      `mapstructure:"redis"`      // Redis配置
	Log        LogConfig        `mapstructure:"log"`        // 日志配置
}

// ServerConfig子配置,映射server配置
type ServerConfig struct {
	Name        string `mapstructure:"name"`        // 服务名称
	Environment string `mapstructure:"environment"` // 环境
}

// BlockchainConfig子配置,映射blockchain配置
type BlockchainConfig struct {
	Network    string   `mapstructure:"network"`     // 网络
	ChainID    int      `mapstructure:"chain_id"`    // 链ID
	RPCURL     string   `mapstructure:"rpc_url"`     // RPC地址
	RPCBackups []string `mapstructure:"rpc_backups"` // 备用RPC地址
}

// SyncswapConfig子配置,映射syncswap配置
type SyncswapConfig struct {
	Factories   FactoriesConfig  `mapstructure:"factories"`    // 工厂配置
	PoolMasters PoolMasterConfig `mapstructure:"pool_masters"` // 池类型地址
	Routers     RoutersConfig    `mapstructure:"routers"`      // 路由器配置
}

// FactoriesConfig子配置,映射SyncswapConfig子配置
type FactoriesConfig struct {
	ClassicV1   string `mapstructure:"classic_v1"`   // 经典V1工厂
	StableV1    string `mapstructure:"stable_v1"`    // 稳定V1工厂
	ClassicV2   string `mapstructure:"classic_v2"`   // 经典V2工厂
	StableV2    string `mapstructure:"stable_v2"`    // 稳定V2工厂
	AquaV2      string `mapstructure:"aqua_v2"`      // AquaV2工厂
	ClassicV2_1 string `mapstructure:"classic_v2_1"` // 经典V2.1工厂
	StableV2_1  string `mapstructure:"stable_v2_1"`  // 稳定V2.1工厂
	AquaV2_1    string `mapstructure:"aqua_v2_1"`    // AquaV2.1工厂
	RangeV3     string `mapstructure:"range_v3"`     // 范围V3工厂
}

type PoolMasterConfig struct {
	ClassicV1   string `mapstructure:"classic_v1"`
	StableV1    string `mapstructure:"stable_v1"`
	ClassicV2   string `mapstructure:"classic_v2"`
	StableV2    string `mapstructure:"stable_v2"`
	AquaV2      string `mapstructure:"aqua_v2"`
	ClassicV2_1 string `mapstructure:"classic_v2_1"`
	StableV2_1  string `mapstructure:"stable_v2_1"`
	AquaV2_1    string `mapstructure:"aqua_v2_1"`
	RangeV3     string `mapstructure:"range_v3"`
}

// GetAllFactories作为FactoriesConfig结构体方法，返回所有工厂地址
func (f *FactoriesConfig) GetAllFactories() []string {
	return []string{
		f.ClassicV1,
		f.StableV1,
		f.ClassicV2,
		f.StableV2,
		f.AquaV2,
		f.ClassicV2_1,
		f.StableV2_1,
		f.AquaV2_1,
		f.RangeV3,
	}
}

// RoutersConfig子配置,映射SyncswapConfig
type RoutersConfig struct {
	V1 string `mapstructure:"v1"` // 路由器V1
	V2 string `mapstructure:"v2"` // 路由器V2
	V3 string `mapstructure:"v3"` // 路由器V3
}

type ScannerConfig struct {
	StartBlock        int    `mapstructure:"start_block"`         // 开始区块
	FetchMode         string `mapstructure:"fetch_mode"`          // 获取模式
	BatchSize         int    `mapstructure:"batch_size"`          // 批量大小
	BatchIntervarSize int    `mapstructure:"batch_interval_size"` // 批量间隔大小
	Workers           int    `mapstructure:"workers"`             // 工作线程数
}

type AbiConfig struct {
	AutoDownload   bool     `mapstructure:"auto_download"`   // 自动下载
	GetAbiEndpoint string   `mapstructure:"getabi_endpoint"` // 获取ABI端点
	SaveDir        string   `mapstructure:"save_dir"`        // 保存目录
	Addresses      []string `mapstructure:"addresses"`       // 地址
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`     // 主机
	Port     int    `mapstructure:"port"`     // 端口
	User     string `mapstructure:"user"`     // 用户
	Password string `mapstructure:"password"` // 密码
	Dbname   string `mapstructure:"dbname"`   // 数据库名称
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`      // 主机
	Port     int    `mapstructure:"port"`      // 端口
	Password string `mapstructure:"password"`  // 密码
	Db       int    `mapstructure:"db"`        // 数据库
	PoolSize int    `mapstructure:"pool_size"` // 连接池大小
}

type LogConfig struct {
	Level  string `mapstructure:"level"`  // 日志级别
	Format string `mapstructure:"format"` // 日志格式
}

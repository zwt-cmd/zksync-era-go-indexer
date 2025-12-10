package config

import (
	"fmt"

	"github.com/spf13/viper"
)

/*
  目的将config.yaml文件中的配置加载到Config结构体中,并配置全部变量，使全局可调用。
  配置步骤：
  1. 读取config.yaml文件
  2. 将配置加载到Config结构体中
  3. 配置全部变量，使全局可调用
  4. 返回Config结构体
  5. 使用Config结构体中的配置
*/

// 定义一个包级别的全局变量,类型为Config，外部可以xxx/cohfig引用
var GlobalConfig *Config

// 加载配置文件
func Load(configFile string) (*Config, error) {
	// 初始化viper，用于读取配置文件
	v := viper.New()

	// 设置配置文件名
	v.SetConfigFile(configFile)

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 将配置文件映射到Config结构体
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 验证配置可以全部验证也可以部分验证
	if cfg.Blockchain.RPCURL == "" {
		return nil, fmt.Errorf("RPCURL不能为空")
	}
	if cfg.Database.Host == "" {
		return nil, fmt.Errorf("数据库主机不能为空")
	}

	GlobalConfig = &cfg
	fmt.Println("配置文件加载成功")

	return &cfg, nil

}

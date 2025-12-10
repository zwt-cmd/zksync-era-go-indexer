package abi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"zk-sync-go-pool/internal/config"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

//

var ABIs = make(map[string]*abi.ABI) // 全局ABI映射

func DownloadABIs(cfg *config.AbiConfig) error {
	// 创建abi保存目录
	if err := os.MkdirAll(cfg.SaveDir, 0755); err != nil {
		return fmt.Errorf("创建abi保存目录失败: %v", err)
	}
	// fmt.Printf("📂 ABI 保存目录: %s\n", cfg.SaveDir)

	for _, address := range cfg.Addresses {
		address = strings.ToLower(address)                                     // 统一小写
		abiFile := filepath.Join(cfg.SaveDir, fmt.Sprintf("%s.json", address)) // 构建abi文件路径

		// 检查文件是否存在
		if _, err := os.Stat(abiFile); os.IsNotExist(err) {

			if cfg.AutoDownload { // 如果配置了自动下载，则下载abi，否则跳过下载
				// 下载abi
				if err := downloadABI(address, abiFile, cfg.GetAbiEndpoint); err != nil {
					fmt.Printf("下载ABI失败: %v\n", err)
					continue
				}
				fmt.Printf("🔍 ABI 下载成功: %s\n", abiFile)
			} else {
				// fmt.Printf("🔍 ABI 文件已存在: %s, 跳过下载\n", abiFile)
				continue
			}
		} else {
			// fmt.Printf("🔍 ABI 文件已存在: %s, 跳过下载\n", abiFile)
		}

		// 加载abi文件 到内存
		if err := loadABI(address, abiFile); err != nil {
			return fmt.Errorf("加载ABI失败: %v", err)
		}
		// fmt.Printf("🔍 ABI 加载成功: %s\n", abiFile)
	}
	// fmt.Printf("🎉 成功加载 %d 个 ABI\n", len(ABIs))
	return nil
}

// downloadABI 从区块浏览器下载 ABI
func downloadABI(address, savePath, endpoint string) error {
	url := endpoint + address

	// 发送 HTTP 请求
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("HTTP 请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析 API 响应
	var result struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  string `json:"result"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("解析 JSON 失败: %v", err)
	}

	if result.Status != "1" {
		return fmt.Errorf("API 返回错误: %s", result.Message)
	}

	// 保存到文件
	if err := os.WriteFile(savePath, []byte(result.Result), 0644); err != nil {
		return fmt.Errorf("保存文件失败: %v", err)
	}

	return nil
}

// loadABI 从文件加载 ABI 到内存,将abi解析存入全局ABI映射中，直接使用ABIs[address]获取
func loadABI(address, abiFile string) error {
	// 读取 ABI 文件
	abiJSON, err := os.ReadFile(abiFile)
	if err != nil {
		return fmt.Errorf("读取文件失败: %v", err)
	}

	// 解析 ABI
	contractABI, err := abi.JSON(strings.NewReader(string(abiJSON)))
	if err != nil {
		return fmt.Errorf("解析 ABI 失败: %v", err)
	}

	// 存入全局缓存
	ABIs[address] = &contractABI
	// fmt.Printf("🔍 ABI 加载成功: %s\n", abiFile)
	return nil
}

// GetABI 获取指定地址的 ABI
func GetABI(address string) *abi.ABI {
	return ABIs[strings.ToLower(address)]
}

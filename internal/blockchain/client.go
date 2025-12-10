package blockchain

import (
	"context"
	"fmt"
	"math/big"
	"time"
	"zk-sync-go-pool/internal/config"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

var Client *ethclient.Client // 全局区块链客户端 最终获得类似于https://zksync-mainnet.core.chainstack.com/a65bb3406867941f5537427dc0e05896 的RPC地址

func InitClient(cfg *config.BlockchainConfig) error {
	// 链接主RPC
	client, err := ethclient.Dial(cfg.RPCURL)
	if err != nil {
		return fmt.Errorf("创建区块链客户端失败: %v", err)
	}

	// 测试连接 - 获取chainID
	ctx := context.Background()         // 创建一个上下文,用于取消请求
	chainID, err := client.ChainID(ctx) // 内部可取消的请求
	if err != nil {
		return fmt.Errorf("获取chainID失败: %v", err)
	}

	// 验证chainID是否正确
	if chainID.Uint64() != uint64(cfg.ChainID) {
		return fmt.Errorf("链ID不正确: %d != %d", chainID.Uint64(), uint64(cfg.ChainID))
	}

	Client = client
	fmt.Printf("区块链客户端初始化成功: %s\n", cfg.RPCURL)
	return nil
}

// 获取最新区块
func GetLatestBlockNumber() (uint64, error) {
	ctx := context.Background()                 // 创建一个上下文,用于取消请求
	blockNumber, err := Client.BlockNumber(ctx) // 获取当然节点的最新区块号
	if err != nil {
		return 0, fmt.Errorf("获取最新区块失败: %v", err)
	}
	return blockNumber, nil
}

// 获取指定区块的详细信息
func GetBlockByNumber(blockNumber uint64) (*types.Block, error) {
	ctx := context.Background()                                             // 创建一个上下文,用于取消请求
	block, err := Client.BlockByNumber(ctx, big.NewInt(int64(blockNumber))) // 获取指定区块的详细信息,返回一个Block结构体
	if err != nil {
		return nil, fmt.Errorf("获取指定区块失败: %v", err)
	}
	return block, nil
}

// 获取指定区块的时间戳
func GetBlockTimestamp(blockNumber uint64) (int64, error) {
	ctx := context.Background()                                               // 创建一个上下文,用于取消请求
	header, err := Client.HeaderByNumber(ctx, big.NewInt(int64(blockNumber))) // 获取指定区块头部信息,返回一个Header结构体
	if err != nil {
		return 0, fmt.Errorf("获取指定区块头部信息失败: %v", err)
	}
	return int64(header.Time), nil // 返回指定区块的时间戳
}

// /*
// 获取指定区块的所有交易回执

// 由于zksync-era的交易有特殊的交易类型(EIP-712)。
// 所以不能直接使用eth_getBlockReceipts RPC 方法
// */
// func GetBlockReceipts(blockNumber uint64) ([]*types.Receipt, error) {
// 	ctx := context.Background() // 创建一个上下文,用于取消请求
// 	// 先获取区块
// 	block, err := GetBlockByNumber(blockNumber)
// 	if err != nil {
// 		return nil, fmt.Errorf("获取指定区块失败: %v", err)
// 	}
// 	// 再获取区块所有交易回执 - 遍历区块所有交易,获取每个交易的回执 -
// 	// 使用ethclient.TransactionReceipt函数 - 传入交易哈希,返回一个Receipt结构体 - 内部可取消的请求
// 	var receipts []*types.Receipt
// 	for _, tx := range block.Transactions() { // 遍历区块所有交易 - 返回一个Transaction结构体切片
// 		receipt, err := Client.TransactionReceipt(ctx, tx.Hash()) // 获取每个交易的回执 - 内部可取消的请求
// 		if err != nil {
// 			return nil, fmt.Errorf("获取指定区块的所有交易回执失败: %v", err)
// 		}
// 		// 将每个交易的回执添加到receipts数组中,
// 		// 交易回执类似于{"status": 1, "cumulativeGasUsed": 1000000, "logs": []}
// 		// 其中status表示交易是否成功,cumulativeGasUsed表示累计消耗的Gas,logs表示交易日志
// 		receipts = append(receipts, receipt)
// 	}
// 	return receipts, nil // 返回所有交易回执
// }

// 获取指定区块的所有交易回执，rpc重试机制
func GetBlockReceipts(blockNumber uint64) ([]*types.Receipt, error) {
	maxRetries := 3
	var receipts []*types.Receipt
	var err error
	for retries := 0; retries < maxRetries; retries++ {
		receipts, err = getBlockReceiptsOnce(blockNumber)
		if err == nil {
			return receipts, nil
		}
		if retries == maxRetries-1 {
			return nil, err
		}
		time.Sleep(time.Second * time.Duration(retries+1)) // 每次重试间隔1秒，2秒，3秒
	}
	return nil, err
}

/*
获取指定区块的所有交易回执 兼容性写法
*/
func getBlockReceiptsOnce(blockNumber uint64) ([]*types.Receipt, error) {
	ctx := context.Background() // 创建一个上下文,用于取消请求
	// Step 1: 获取区块信息（只获取交易哈希，不解析交易体）
	type BlockWithTxHashes struct {
		Transactions []common.Hash `json:"transactions"`
	}

	var block BlockWithTxHashes
	err := Client.Client().CallContext(ctx, &block, "eth_getBlockByNumber",
		fmt.Sprintf("0x%x", blockNumber), false) // false = 只返回交易哈希
	if err != nil {
		return nil, fmt.Errorf("获取区块信息失败: %v", err)
	}

	// Step 2: 逐个获取交易回执 加上重试机制
	var receipts []*types.Receipt
	for _, txHash := range block.Transactions {
		var receipt *types.Receipt
		var err error
		for retry := 0; retry < 3; retry++ {
			receipt, err = Client.TransactionReceipt(ctx, txHash)
			if err == nil {
				break
			}
			time.Sleep(time.Duration(retry+1) * time.Second)
		}
		if err != nil {
			return nil, fmt.Errorf("获取交易回执失败 %s: %w", txHash.Hex(), err)
		}
		receipts = append(receipts, receipt)
	}

	return receipts, nil
}

/*
获取safe头高度
*/

func GetSafeBlockNumber() (uint64, error) {
	ctx := context.Background() // 创建一个上下文,用于取消请求
	var block struct {
		Number string `json:"number"`
	}
	err := Client.Client().CallContext(ctx, &block, "eth_getBlockByNumber", "safe", false)
	if err != nil {
		return 0, fmt.Errorf("获取safe头高度失败: %v", err)
	}

	n, ok := new(big.Int).SetString(block.Number[2:], 16)
	if !ok {
		return 0, fmt.Errorf("解析safe头高度失败: %s", block.Number)
	}

	return n.Uint64(), nil
}

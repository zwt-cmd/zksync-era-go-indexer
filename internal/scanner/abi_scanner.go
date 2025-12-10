package scanner

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"
	"zk-sync-go-pool/internal/blockchain"
	"zk-sync-go-pool/internal/config"
	"zk-sync-go-pool/internal/models"
	"zk-sync-go-pool/internal/repository"

	"zk-sync-go-pool/internal/abi"

	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// 映射工厂地址
type factoryInfo struct {
	PoolType  string
	Version   string
	EventName string // 默认 PoolCreated 可为空
	FeeField  string // 某些事件会带fee/feeTier 可为空
}

/*
基于ABI扫描解析
*/
type ABIScanner struct {
	cfg            *config.Config          //引用config指针地址
	repo           *repository.Repository  // 引用repo方法集指针地址
	poolCache      map[string]*models.Pool // 池子地址集合
	poolCacheMu    sync.RWMutex
	factoryInfoMap map[string]factoryInfo
	poolABIMap     map[string]string
}

func NewABIScanner(cfg *config.Config, repo *repository.Repository) *ABIScanner {
	s := &ABIScanner{ //结构体赋值
		cfg:  cfg,
		repo: repo,
	}
	s.initFatoryInfo()
	s.initPoolABIMap()
	s.initPoolCache()
	return s
}

// 初始化映射版本+池类型地址
func (s *ABIScanner) initPoolABIMap() {
	s.poolABIMap = make(map[string]string)
	masters := s.cfg.Syncswap.PoolMasters
	s.poolABIMap["classic:v1"] = strings.ToLower(masters.ClassicV1)
	s.poolABIMap["stable:v1"] = strings.ToLower(masters.StableV1)
	s.poolABIMap["classic:v2"] = strings.ToLower(masters.ClassicV2)
	s.poolABIMap["stable:v2"] = strings.ToLower(masters.StableV2)
	s.poolABIMap["aqua:v2"] = strings.ToLower(masters.AquaV2)
	s.poolABIMap["classic:v2.1"] = strings.ToLower(masters.ClassicV2_1)
	s.poolABIMap["stable:v2.1"] = strings.ToLower(masters.StableV2_1)
	s.poolABIMap["aqua:v2.1"] = strings.ToLower(masters.AquaV2_1)
	s.poolABIMap["range:v3"] = strings.ToLower(masters.RangeV3)
}

// 初始化映射工厂合约地址
func (s *ABIScanner) initFatoryInfo() {
	s.factoryInfoMap = make(map[string]factoryInfo)
	factories := s.cfg.Syncswap.Factories
	s.factoryInfoMap[strings.ToLower(factories.ClassicV1)] = factoryInfo{PoolType: "classic", Version: "v1", EventName: "PoolCreated"}
	s.factoryInfoMap[strings.ToLower(factories.StableV1)] = factoryInfo{PoolType: "stable", Version: "v1", EventName: "PoolCreated"}
	s.factoryInfoMap[strings.ToLower(factories.ClassicV2)] = factoryInfo{PoolType: "classic", Version: "v2", EventName: "PoolCreated"}
	s.factoryInfoMap[strings.ToLower(factories.StableV2)] = factoryInfo{PoolType: "stable", Version: "v2", EventName: "PoolCreated"}
	s.factoryInfoMap[strings.ToLower(factories.AquaV2)] = factoryInfo{PoolType: "aqua", Version: "v2", EventName: "PoolCreated"}
	s.factoryInfoMap[strings.ToLower(factories.ClassicV2_1)] = factoryInfo{PoolType: "classic", Version: "v2.1", EventName: "PoolCreated"}
	s.factoryInfoMap[strings.ToLower(factories.StableV2_1)] = factoryInfo{PoolType: "stable", Version: "v2.1", EventName: "PoolCreated"}
	s.factoryInfoMap[strings.ToLower(factories.AquaV2_1)] = factoryInfo{PoolType: "aqua", Version: "v2.1", EventName: "PoolCreated"}
	s.factoryInfoMap[strings.ToLower(factories.RangeV3)] = factoryInfo{PoolType: "range", Version: "v3", EventName: "PoolCreated"}
}

// 初始化映射池子地址
func (s *ABIScanner) initPoolCache() {
	s.poolCache = make(map[string]*models.Pool)
	pools, err := s.repo.GetAllPools()
	if err != nil {
		fmt.Printf("加载历史池子失败%v", err)
		return
	}
	for _, pool := range pools {
		addr := strings.ToLower(pool.PoolAddress)
		s.poolCache[addr] = pool
	}

	fmt.Printf("初始化池子缓存: %d 条\n", len(s.poolCache))
}

/*
开始扫描
*/
func (s *ABIScanner) Start(ctx context.Context) error {
	fmt.Println("启动ABI扫描器...")
	stableCursor, err := s.repo.GetScanProgress("stable_scan")
	if err != nil {
		return err
	}
	if stableCursor == 0 {
		startBlock := uint64(s.cfg.Scanner.StartBlock)
		fmt.Printf("首次运行从配置起始块开始%v:", startBlock)
		if err := s.repo.InitScanProgress("stable_scan", startBlock); err != nil {
			return err
		}
		stableCursor = startBlock
	} else {
		fmt.Printf("从上次扫描的区块%v开始", stableCursor)
	}

	// 开启双worker模式
	go s.runStableWorker(ctx, stableCursor)
	go s.runLiveWorker(ctx)

	<-ctx.Done() //监听信号取消
	return nil
	// -----------------------------------------------------

	//单循环扫描模式
	//获取配置批量扫描数量
	// batchSize := uint64(s.cfg.Scanner.BatchSize)
	// // for循环不间断遍历区块链
	// for {
	// 	latest, err := blockchain.GetLatestBlockNumber()
	// 	if err != nil {
	// 		fmt.Printf("获取最新区块失败:%v，5s后重试", err)
	// 		time.Sleep(time.Second * 5)
	// 		continue
	// 	}

	// 	if stableCursor >= latest {
	// 		fmt.Printf("已扫描到最新的区块:%v,等待新的区块...", latest)
	// 		time.Sleep(time.Second * 2)
	// 		continue
	// 	}

	// 	endBlock := stableCursor + uint64(batchSize)
	// 	if endBlock > latest {
	// 		endBlock = latest
	// 	}

	// 	if err := s.scanRange(stableCursor+1, endBlock); err != nil {
	// 		fmt.Printf("扫描区块范围%v-%v", stableCursor+1, endBlock)
	// 	}

	// 	stableCursor = endBlock //更新stableCursor
	// 	// 一次批量之后更新进度一次
	// 	if err := s.repo.UpdateScanProgress("main_scan", endBlock); err != nil {
	// 		fmt.Printf("更新最新进度失败:%v", endBlock)
	// 	}
	// }

}

func (s *ABIScanner) runStableWorker(ctx context.Context, cursor uint64) {
	batchSize := uint64(s.cfg.Scanner.BatchSize) // 获取配置批量扫描数量

	for {
		select { //这里select用来监听ctx取消信号 不做其他用途
		case <-ctx.Done():
			return
		default:
		}
		safeHead, err := blockchain.GetSafeBlockNumber()
		if err != nil {
			fmt.Printf("获取safe头高度失败:%v，1s后重试", err)
			time.Sleep(time.Second) // 1s后重试
			continue
		}

		if cursor >= safeHead { // 已经扫描到最新的safe头高度
			time.Sleep(time.Second * 2)
			continue
		}
		form := cursor + 1
		to := form + batchSize - 1
		if to > safeHead { // 不超过safe头高度
			to = safeHead
		}

		if err := s.scanRange(form, to, "safe"); err != nil { // 扫描区块范围
			fmt.Printf("扫描区块范围%v-%v失败:%v\n", form, to, err)
			time.Sleep(time.Second) // 1s后重试
			continue
		}

		cursor = to //更新cursor
		// 一次批量之后更新进度一次
		if err := s.repo.UpdateScanProgress("stable_scan", to); err != nil {
			fmt.Printf("更新最新进度失败:%v", to)
		}

	}
}

/*
live worker 负责扫描最新区块，入库pending状态.
2s检查一次获取最新的区块高度和safe头高度，如果有新的区块就
*/
func (s *ABIScanner) runLiveWorker(ctx context.Context) {
	interval := time.Second * 5 // 每5秒检查一次

	for {
		select { //这里select用来监听ctx取消信号 不做其他用途
		case <-ctx.Done():
			return
		default:
		}

		// 获取两个头
		latest, err1 := blockchain.GetLatestBlockNumber()
		safeHead, err2 := blockchain.GetSafeBlockNumber()
		if err1 != nil || err2 != nil {
			fmt.Printf("获取最新区块或safe头高度失败:%v,%v，1s后重试", err1, err2)
			time.Sleep(interval)
			continue
		}

		if latest <= safeHead { //假设没有新的区块产生，等待2秒后继续执行for循环
			time.Sleep(interval)
			continue
		}

		from, to := safeHead+1, latest // 扫描区块范围 safeHead+1 到 latest

		// 先清理旧的也就是上次的pending数据
		if err := s.repo.DeletePendingAfter(safeHead); err != nil {
			fmt.Printf("清理pending状态Swap事件失败:%v", err)
		}

		// 重建当前的live区域
		if err := s.scanRangeLive(from, to, "pending"); err != nil {
			fmt.Printf("扫描区块范围%v-%v失败:%v\n", from, to, err)
		}

		time.Sleep(interval) // 2秒后继续执行for循环

	}
}

/*
批量区块扫描，多协程
例如1000个区块，那我们就将数据<-到通道中，然后for循环遍历开启5个协程，
一起来执行解析某个区块的任务。结合计数器(wg)、锁(mu),通道(channel)。
注意点:
1. 读写共享字段需要加锁。
2. 遵循消费者-> 生产者模写。
*/
func (s *ABIScanner) scanRange(start, end uint64, finality string) error {
	workers := s.cfg.Scanner.Workers
	if workers == 0 {
		workers = 5 // 获取不到则默认5个协程
	}

	tasks := make(chan uint64, workers*2) // 通道设置内存大小
	var wg sync.WaitGroup
	var mu sync.Mutex        //互斥锁
	var errorCount int       // 协程解析单个区块错误数量
	maxScannedBlock := start // 记录当前扫描最新的区块高度(因为多协程扫描不是按照顺序来的)
	if start > 0 {
		maxScannedBlock = start - 1
	}

	// 开启消费者（等待生产者生产数据）
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for blockNum := range tasks {
				if err := s.scanBlock(blockNum, finality); err != nil {
					fmt.Printf("扫描区块:%v失败\n", blockNum)
					mu.Lock()
					errorCount++
					mu.Unlock()
					continue
				}

				mu.Lock()
				if blockNum > maxScannedBlock {
					maxScannedBlock = blockNum // 取最大的高度
				}
				mu.Unlock()

			}
		}()
	}

	// 开启生产者
	go func() {
		defer close(tasks) // 协程结束前关闭通道
		for blockNum := start; blockNum <= end; blockNum++ {
			tasks <- blockNum
		}
	}()

	// 批量扫描断点记录 每隔一定数据区块记录一次，防止进程异常进度丢失
	batchIntervalSize := s.cfg.Scanner.BatchIntervarSize
	done := make(chan bool)
	var lastUpdatedBlock uint64 = start - 1 // 上次更新的区块高度

	go func() {
		ticker := time.NewTicker(time.Second * 5) // 每5秒检查一次
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C: // 5s触发
				mu.Lock()
				currectMax := maxScannedBlock
				mu.Unlock()
				// 当前的进度，大于一开始的进度+间隔，说明有新的进度需要更新
				if currectMax >= lastUpdatedBlock+uint64(batchIntervalSize) {
					// 更新数据库进度
					if finality == "safe" {
						// 只有safe扫描任务才更新进度
						if err := s.repo.UpdateScanProgress("stable_scan", currectMax); err != nil {
							fmt.Printf("定时更新扫描进度失败:%v", err)
						} else {
							fmt.Printf("定时更新扫描进度到区块:%d\n", currectMax)
							lastUpdatedBlock = currectMax
						}
					}
				}
			case <-done: // 收到停止信号
				return
			}
		}
	}()

	wg.Wait()    // 阻塞进程消费者完成才会走下面业务
	done <- true // 停止定时更新协程

	mu.Lock()
	finalBlock := maxScannedBlock
	finalErrors := errorCount
	mu.Unlock()

	fmt.Printf("%d区域，批次完成，扫描区块 %d-%d，错误数 %d\n", finality, start, finalBlock, finalErrors)
	return nil

}

func (s *ABIScanner) scanRangeLive(start, end uint64, finality string) error {
	workers := 5 // 固定5个协程

	tasks := make(chan uint64, workers*2) // 通道设置内存大小
	var wg sync.WaitGroup
	var mu sync.Mutex  //互斥锁
	var errorCount int // 协程解析单个区块错误数量

	// 开启消费者（等待生产者生产数据）
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for blockNum := range tasks {
				if err := s.scanBlock(blockNum, finality); err != nil {
					fmt.Printf("扫描区块:%v失败\n", blockNum)
					mu.Lock()
					errorCount++
					mu.Unlock()
					continue
				}
			}
		}()
	}

	// 开启生产者
	go func() {
		defer close(tasks) // 协程结束前关闭通道
		for blockNum := start; blockNum <= end; blockNum++ {
			tasks <- blockNum
		}
	}()

	wg.Wait() // 阻塞进程消费者完成才会走下面业务

	mu.Lock()
	finalErrors := errorCount
	mu.Unlock()

	fmt.Printf("%d区域，批次完成，扫描区块 %d-%d，错误数 %d\n", finality, start, end, finalErrors)
	return nil

}

/*
单区块开始解析日志
*/
func (s *ABIScanner) scanBlock(blockNum uint64, finality string) error {
	receipts, err := blockchain.GetBlockReceipts(blockNum)
	if err != nil {
		return err
	}
	// 获取区块时间戳
	blockTimestamp, err := blockchain.GetBlockTimestamp(blockNum)
	if err != nil {
		return err
	}

	var poolCount int
	var swapCount int
	for _, receipt := range receipts {
		for _, log := range receipt.Logs {
			if s.handlePoolLog(blockNum, receipt.TxHash.Hex(), log) {
				poolCount++
				continue
			}
			if s.handleSwapLog(blockNum, blockTimestamp, receipt.TxHash.Hex(), log, finality) {
				swapCount++
				continue
			}
		}
	}
	if poolCount > 0 || swapCount > 0 {
		fmt.Printf("✅ 扫描区块 %d: 发现 %d 个池子, %d 个Swap事件\n", blockNum, poolCount, swapCount)
	} else {
		fmt.Printf("  区块 %d: %d 笔交易\n", blockNum, len(receipts))
	}
	return nil

}

/*
解析Pool创建池类型日志
*/
func (s *ABIScanner) handlePoolLog(blockNum uint64, txHash string, log *types.Log) bool {
	factoryAddr := strings.ToLower(log.Address.Hex()) // 如果是创建池子，log.address为工厂地址
	info, ok := s.factoryInfoMap[factoryAddr]
	if !ok {
		return false // 不是我们跟踪的工厂
	}

	eventName := info.EventName
	if eventName == "" {
		eventName = "PoolCreated"
	}

	contracABI := s.getABI(factoryAddr) // 获取对应ABI解析的日志信息
	if contracABI == nil {
		return false
	}

	event, ok := contracABI.Events[eventName]
	if !ok || log.Topics[0] != event.ID {
		return false
	}

	indexedCount := 0
	for _, input := range event.Inputs {
		if input.Indexed {
			indexedCount++
		}
	}
	if len(log.Topics) < indexedCount+1 {
		return false
	}

	// 解析indexed参数 哈希截取创建池子的token0 token1代币类型地址
	token0 := common.BytesToAddress(log.Topics[1].Bytes()).Hex()
	token1 := common.BytesToAddress(log.Topics[2].Bytes()).Hex()

	// 解析非indexed数据
	data := make(map[string]interface{})
	if err := contracABI.UnpackIntoMap(data, eventName, log.Data); err != nil {
		fmt.Printf("解析PoolCreated失败:%v", err)
		return false
	}
	poolAddr, _ := data["pool"].(common.Address) //获取到池子地址（创建池类型，池子地址在data中）

	pool := &models.Pool{
		PoolAddress:    poolAddr.Hex(),
		FactoryAddress: log.Address.Hex(),
		PoolType:       info.PoolType,
		Version:        info.Version,
		Token0:         token0,
		Token1:         token1,
		CreatedTx:      txHash,
		CreatedBlock:   blockNum,
	}

	// 创建池子默认会带着这个池子手续费
	if info.FeeField != "" {
		if v, ok := data[info.FeeField].(*big.Int); ok && v != nil {
			fee := int(v.Int64())
			pool.FeeRate = &fee
		}
	}

	if err := s.repo.SavePool(pool); err != nil {
		fmt.Printf("保存池子失败：%v", err)
		return false
	}

	s.poolCacheMu.Lock()
	s.poolCache[strings.ToLower(pool.PoolAddress)] = pool // 将池子信息缓存到内存中
	s.poolCacheMu.Unlock()

	return true

}

/*
解析兑换swap类型日志
*/
func (s *ABIScanner) handleSwapLog(blockNum uint64, blockTimestamp int64, txHash string, log *types.Log, finality string) bool {
	poolAddress := strings.ToLower(log.Address.Hex()) //如果是swap类型，log.address为池子地址

	s.poolCacheMu.RLock()
	pool, ok := s.poolCache[poolAddress] // 判断缓存中是否记录该池子
	s.poolCacheMu.RUnlock()
	if !ok { // 缓存中没有则从数据库获取
		// poolFromDB, _ := s.repo.GetPoolByAddress(poolAddress)
		// if poolFromDB == nil {
		return false
		// }
		// s.poolCacheMu.Lock()
		// s.poolCache[poolAddress] = poolFromDB
		// s.poolCacheMu.Unlock()
		// pool = poolFromDB
	}

	// 找到对应的 pool master ABI
	key := fmt.Sprintf("%s:%s", pool.PoolType, pool.Version) // 如 classic:v2

	masterAddr, ok := s.poolABIMap[key]
	if !ok {
		return false // 不支持此类型
	}

	contractABI := s.getABI(masterAddr)
	if contractABI == nil {
		return false
	}

	// 3. 校验事件签名（这里默认事件名都是 "Swap"，不同版本可做映射）
	event, ok := contractABI.Events["Swap"]
	if !ok || log.Topics[0] != event.ID {
		return false
	}

	// 4. 解析 indexed & non-indexed 数据
	sender := common.BytesToAddress(log.Topics[1].Bytes()).Hex()
	recipient := common.BytesToAddress(log.Topics[2].Bytes()).Hex()

	fields := make(map[string]interface{}) //解析log.data
	if err := contractABI.UnpackIntoMap(fields, "Swap", log.Data); err != nil {
		fmt.Printf("解析 Swap 失败: %v\n", err)
		return false
	}

	var tokenIn, tokenOut, amountIn, amountOut string
	switch pool.PoolType {
	case "range":
		amt0, _ := fields["amount0"].(*big.Int)
		amt1, _ := fields["amount1"].(*big.Int)
		if amt0 == nil || amt1 == nil {
			return false
		}
		if amt0.Sign() < 0 {
			tokenIn, tokenOut = pool.Token1, pool.Token0
			amountIn, amountOut = new(big.Int).Abs(amt1).String(), new(big.Int).Abs(amt0).String()
		} else {
			tokenIn, tokenOut = pool.Token0, pool.Token1
			amountIn, amountOut = new(big.Int).Abs(amt0).String(), new(big.Int).Abs(amt1).String()
		}
	default:
		amt0In, _ := fields["amount0In"].(*big.Int)
		amt1In, _ := fields["amount1In"].(*big.Int)
		amt0Out, _ := fields["amount0Out"].(*big.Int)
		amt1Out, _ := fields["amount1Out"].(*big.Int)
		if amt0In == nil || amt1In == nil || amt0Out == nil || amt1Out == nil {
			return false
		}
		if amt0In.Sign() > 0 {
			tokenIn, tokenOut = pool.Token0, pool.Token1
			amountIn, amountOut = amt0In.String(), amt1Out.String()
		} else {
			tokenIn, tokenOut = pool.Token1, pool.Token0
			amountIn, amountOut = amt1In.String(), amt0Out.String()
		}
	}

	// 5. 落库
	swap := &models.SwapEvent{
		BlockNumber:    blockNum,
		BlockTimeStamp: blockTimestamp,
		TxHash:         txHash,
		LogIndex:       int(log.Index),
		PoolAddress:    pool.PoolAddress,
		Sender:         sender,
		Recipient:      recipient,
		TokenIn:        tokenIn,
		TokenOut:       tokenOut,
		AmountIn:       amountIn,
		AmountOut:      amountOut,
		FinalityStatus: finality,
	}

	if err := s.repo.SaveSwapEvent(swap); err != nil {
		fmt.Printf("保存 Swap 失败: %v\n", err)
		return false
	}

	return true

}

/*
根据工厂地址，poolMaster获取我们下载的abi文件
*/
func (s *ABIScanner) getABI(address string) *ethabi.ABI {
	return abi.GetABI(strings.ToLower(address))
}

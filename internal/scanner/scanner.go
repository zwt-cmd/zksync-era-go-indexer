package scanner

import (
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"
	"zk-sync-go-pool/internal/blockchain"
	"zk-sync-go-pool/internal/config"
	"zk-sync-go-pool/internal/models"
	"zk-sync-go-pool/internal/repository"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// 工厂信息结构体
type FactoryInfo struct {
	PoolType       string
	Version        string
	PoolCreatedSig common.Hash // 各类型创建池子事件哈希集合
}

type Scanner struct {
	cfg            *config.Config
	repo           *repository.Repository
	factoryInfoMap map[string]FactoryInfo // 池子信息映射 用于存储工厂信息

	poolCache      map[string]bool // 池子地址内存缓存
	swapSignatures []common.Hash   // 各类型swap事件哈希集合
}

// 创建Scanner 扫描器 专注于扫描事件和索引事件
func NewScanner(cfg *config.Config, repo *repository.Repository) *Scanner {
	s := &Scanner{cfg: cfg, repo: repo}
	s.initPoolInfoMap()    // 初始化映射工厂地址
	s.initSwapSignatures() // 初始化收集各类型swap事件哈希集合
	s.initPoolCache()      // 初始化池子地址内存缓存
	return s
}

// 初始化池子信息映射
func (s *Scanner) initPoolInfoMap() {
	s.factoryInfoMap = make(map[string]FactoryInfo) // map初始化，未分配内存，空map
	// zksync-era 稳定池跟经典池各版本事件哈希一致
	standardSig := common.HexToHash("0x9c5d829b9b23efc461f9aeef91979ec04bb903feb3bee4f26d22114abfc7335b")
	// zksync-era 范围池事件哈希
	rangeV3Sig := common.HexToHash("0xab0d57f0df537bb25e80245ef7748fa62353808c54d6e528a9dd20887aed9ac2")
	s.factoryInfoMap[strings.ToLower(s.cfg.Syncswap.Factories.ClassicV1)] = FactoryInfo{
		PoolType:       "classic",
		Version:        "v1",
		PoolCreatedSig: standardSig,
	}
	s.factoryInfoMap[strings.ToLower(s.cfg.Syncswap.Factories.StableV1)] = FactoryInfo{
		PoolType:       "stable",
		Version:        "v1",
		PoolCreatedSig: standardSig,
	}
	s.factoryInfoMap[strings.ToLower(s.cfg.Syncswap.Factories.ClassicV2)] = FactoryInfo{
		PoolType:       "classic",
		Version:        "v2",
		PoolCreatedSig: standardSig,
	}

	s.factoryInfoMap[strings.ToLower(s.cfg.Syncswap.Factories.StableV2)] = FactoryInfo{
		PoolType:       "stable",
		Version:        "v2",
		PoolCreatedSig: standardSig,
	}

	s.factoryInfoMap[strings.ToLower(s.cfg.Syncswap.Factories.AquaV2)] = FactoryInfo{
		PoolType:       "aqua",
		Version:        "v2",
		PoolCreatedSig: standardSig,
	}
	s.factoryInfoMap[strings.ToLower(s.cfg.Syncswap.Factories.ClassicV2_1)] = FactoryInfo{
		PoolType:       "classic",
		Version:        "v2.1",
		PoolCreatedSig: standardSig,
	}

	s.factoryInfoMap[strings.ToLower(s.cfg.Syncswap.Factories.StableV2_1)] = FactoryInfo{
		PoolType:       "stable",
		Version:        "v2.1",
		PoolCreatedSig: standardSig,
	}

	s.factoryInfoMap[strings.ToLower(s.cfg.Syncswap.Factories.AquaV2_1)] = FactoryInfo{
		PoolType:       "aqua",
		Version:        "v2.1",
		PoolCreatedSig: standardSig,
	}
	s.factoryInfoMap[strings.ToLower(s.cfg.Syncswap.Factories.RangeV3)] = FactoryInfo{
		PoolType:       "range",
		Version:        "v3",
		PoolCreatedSig: rangeV3Sig, // ✅ 使用不同的签名
	}

	fmt.Printf("✅ 已加载 %d 个工厂合约映射\n", len(s.factoryInfoMap))

}

// 初始化各类型swap事件哈希集合
func (s *Scanner) initSwapSignatures() {
	// classic/stable/Aaqa类型池子的swap事件签名一致
	classicSwapSig := common.HexToHash("0xd78ad95fa46c994b6551d0da85fc275fe613ce37657fb8d5e3d130840159d822")
	// range V3 得swap事件签名不同
	rangeV3Swapsig := common.HexToHash("0xc42079f94a6350d7e6235f29174924f928cc2ac818eb64fed8004e115fbcca67")
	s.swapSignatures = []common.Hash{classicSwapSig, rangeV3Swapsig}

}

// 初始化池子地址内存缓存
func (s *Scanner) initPoolCache() {
	s.poolCache = make(map[string]bool) // map初始化，未分配内存，空map
	pools, err := s.repo.GetAllPools()
	if err != nil {
		fmt.Printf("加载历史池子失败: %v\n", err)
		return
	}
	for _, pool := range pools {
		poolAddr := strings.ToLower(pool.PoolAddress)
		s.poolCache[poolAddr] = true
	}

}

/*
启动扫描器

	Start函数属于Scanner结构体的方法，属于Scanner的方法集。
	注意!! 不加前面*Scanner，则Start就是普通函数。
	interface只能匹配方法集，匹配不到普通函数，实现不了多态。
*/
func (s *Scanner) Start() error {
	fmt.Println("启动扫描器")
	// 1.先读取扫描进度，需要Repository 提供方法
	lastBlock, err := s.repo.GetScanProgress("main_scan")
	if err != nil {
		return err
	}
	fmt.Println("上次扫描到的区块高度:", lastBlock)

	// 2. 如果是首次运行(lastBlock == 0)，则从配置中的起始区块回填
	if lastBlock == 0 {
		startBlock := s.cfg.Scanner.StartBlock
		fmt.Println("首次运行，从配置中的起始区块回填:", startBlock)

		//首次运行，进度为空要初始化一下，需要Repository 提供方法
		if err := s.repo.InitScanProgress("main_scan", uint64(startBlock)); err != nil {
			return err
		}
		lastBlock = uint64(startBlock)
	} else {
		fmt.Printf("📖 从上次进度继续: %d\n", lastBlock)
	}

	// 3. 获取配置文件每批扫描的区块数
	batchSize := uint64(s.cfg.Scanner.BatchSize)

	// 4. for单独使用是无限循环，正常情况下内容执行完，自动再次执行，可以加条件跳出循环。
	for {
		latest, err := blockchain.GetLatestBlockNumber()
		if err != nil {
			fmt.Printf("⚠️ 获取最新区块失败: %v，5秒后重试\n", err)
			time.Sleep(5 * time.Second)
			continue
		}

		//如果已经扫描到最新 等待2s跳过继续轮询新的区块，不可以太长时间，交易状态不能及时更新。
		if lastBlock >= latest {
			fmt.Printf("已扫描到最新的区块:%d", latest)
			time.Sleep(2 * time.Second) // 等待2s 继续for循环
			continue
		}
		// 否则就是正常批量区块
		endBlock := lastBlock + batchSize
		if endBlock > latest {
			endBlock = latest
		}

		// 开始扫描
		if err := s.scanRange(lastBlock+1, endBlock); err != nil {
			fmt.Printf("扫描失败：%v，但是不退出，继续下一批", err)
		}

		// 失败/成功都要更新进度，保证继续走下去。
		lastBlock = endBlock
		if err := s.repo.UpdateScanProgress("main_scan", endBlock); err != nil {
			fmt.Printf("⚠️ 更新进度失败: %v\n", err)
		}
	}

	// // 4.先扫描10个区块测试
	// endBlock := lastBlock + 100000
	// if endBlock > latest {
	// 	endBlock = latest
	// }
	// fmt.Printf("📖 扫描区块范围: %d - %d\n", lastBlock, endBlock)

	// // 调用扫描方法（scanRange 内部会更新进度，这里不需要重复更新）
	// if err := s.scanRange(lastBlock+1, endBlock); err != nil {
	// 	return err
	// }

	// fmt.Printf("✅ 扫描完成\n")
	// return nil
}

// scanRange 扫描区块范围 for循环一次区块一个区块遍历
// func (s *Scanner) scanRange(start, end uint64) error {
// 	var updateInterval = uint64(100) // 暂定间隔100个区块更新一次进度，防止中途崩溃，进度丢失.
// 	var updateCount = start
// 	var errorCount int // 统计这个协程扫描统计区块错误数量
// 	for blockNum := start; blockNum <= end; blockNum++ {
// 		if err := s.scanBlock(blockNum); err != nil {
// 			// return err 这里return err 会导致整个协程退出，所以不能直接return err
// 			errorCount++
// 			fmt.Printf("⚠️ 扫描区块 %d 失败: %v\n", blockNum, err)
// 			continue // 继续扫描下一个区块
// 		}
// 		// 达到更新间隔，更新进度
// 		if blockNum-updateCount >= updateInterval {
// 			if err := s.repo.UpdateScanProgress("main_scan", blockNum); err != nil {
// 				// return err 这个也不要影响for循环遍历
// 				fmt.Printf("⚠️ 更新进度失败: %v\n", err)
// 				continue
// 			}
// 			updateCount = blockNum
// 		}
// 	}
// 	// 打印多少到区块，多少个错误
// 	fmt.Printf("✅ 扫描完成，扫描到区块: %d，错误数量: %d\n", updateCount, errorCount)
// 	errorCount = 0 // 清零，防止下次扫描时，错误数量不准确
// 	return nil
// }

func (s *Scanner) scanRange(start, end uint64) error {
	// 设定工作协程数量
	workers := s.cfg.Scanner.Workers

	// 既然用了协程，那数据势必要放在通道了，所以我们要定义通道
	// 通道缓冲区大小为工作协程数量的2倍，防止协程都在工作，新的数据无法进入通道，导致协程阻塞。
	tasks := make(chan uint64, workers*2)

	var errorCount int                     // 定义扫描的错误数量
	var wg sync.WaitGroup                  // 定义等待组，用于等待所有协程完成
	var mu sync.Mutex                      // 定义互斥锁，用于保护共享资源
	var maxScannedBlock uint64 = start - 1 // 记录扫描到的最大区块高度，初始值为起始区块-1，因为for循环会先加1再判断

	//启动消费者，准备好等待任务
	for i := 0; i < workers; i++ { // 协程三伙伴：计数器（wg）、锁（mu）、通道（channel）
		wg.Add(1) // 计数器加1，表示有一个协程要完成
		go func() {
			defer wg.Done() // 协程完成时，计数器减1，表示有一个协程完成了
			for blockNum := range tasks {
				if err := s.scanBlock(blockNum); err != nil {
					mu.Lock() // 锁住共享资源，防止多个协程同时修改errorCount
					errorCount++
					fmt.Printf("⚠️ 扫描区块 %d 失败: %v\n", blockNum, err)
					mu.Unlock()
					continue
				}
				mu.Lock() // 锁住共享资源，防止多个协程同时修改maxScannedBlock
				if blockNum > maxScannedBlock {
					// 因为多个协程异步执行，我们并不知道哪个协程在扫描到最高的区块。所以谁扫描到最高的区块，谁就更新maxScannedBlock
					maxScannedBlock = blockNum
				}
				mu.Unlock()
			}
		}()
	}

	//发布任务，生产者
	go func() {
		defer close(tasks) // 关闭通道，表示没有更多的任务了
		for blockNum := start; blockNum <= end; blockNum++ {
			tasks <- blockNum // 发布任务
		}
	}()

	// 定期更新进度到数据库
	updateInterval := uint64(100)           // 每隔100个区块更新一次进度
	done := make(chan bool)                 // 完成信号
	var lastUpdatedBlock uint64 = start - 1 // 记录上次更新的区块号

	go func() {
		ticker := time.NewTicker(time.Second * 5) // 每隔5秒更新一次进度
		defer ticker.Stop()                       // 协程完成前停止定时器
		for {                                     // for select用来监听信号，不管是channel还是定时器，其实就是不同的信号我就执行什么业务。
			select {
			case <-ticker.C: // 定时器触发
				mu.Lock()
				currentMax := maxScannedBlock
				mu.Unlock()

				// 达到更新间隔，更新进度
				// 检查是否达到更新间隔（距离上次更新 >= 100 个区块）
				if currentMax >= start && currentMax-lastUpdatedBlock >= updateInterval {
					if err := s.repo.UpdateScanProgress("main_scan", currentMax); err != nil {
						fmt.Printf("⚠️ 更新进度失败: %v\n", err)
						continue
					}
					fmt.Printf("✅ 进度更新到: %d (已扫描 %d 个区块)\n",
						currentMax, currentMax-start+1)
					lastUpdatedBlock = currentMax // 更新记录
				}
			case <-done: // 收到完成信号，退出循环
				return
			}
		}
	}()

	// 等待所有协程完成
	wg.Wait()    //阻塞等待所有工作协程完成,既然工作协程都完成了，那么相关联的更新进度协程也该停止了
	done <- true // 发送完成信号

	// 可能存在工作协程关闭，更新进度协程紧接着也关闭了，可能会存在最后一次没有进到更新协程中去，
	// 所以最后再更新一次（读取共享变量需要加锁）
	mu.Lock()
	finalBlock := maxScannedBlock
	finalErrors := errorCount
	mu.Unlock()

	if err := s.repo.UpdateScanProgress("main_scan", finalBlock); err != nil {
		fmt.Printf("⚠️ 更新进度失败: %v\n", err)
	}
	fmt.Printf("✅ 扫描完成，扫描到区块: %d，错误数量: %d\n", finalBlock, finalErrors)
	return nil
}

// scanBlock 扫描单个区块
func (s *Scanner) scanBlock(blockNum uint64) error {
	// 调用 blockchain 获取区块数据
	receipts, err := blockchain.GetBlockReceipts(blockNum)
	if err != nil {
		return err
	}
	// 获取区块时间戳
	blockTimestamp, err := blockchain.GetBlockTimestamp(blockNum)
	if err != nil {
		return err
	}

	// TODO:在这里解析全部类型的日志（Swap/Mint/Burn/Sync）
	var poolCount, swapCount int // 统计池子数量和Swap事件数量
	for _, receipt := range receipts {
		for _, log := range receipt.Logs {

			/*
			 Mint 事件
			 判断是否为工厂合约（过滤掉非工厂合约的log 粗过滤）
			*/
			if s.isFactoryContract(log.Address) {
				// 判断log是否是池子创建事件（过滤掉非池子创建事件的log 细过滤）
				if s.isPoolCreatedEvent(*log) {
					// fmt.Printf("✅ 扫描区块 %d: 发现池子创建事件\n", blockNum)
					// 解析池子创建事件
					pool := s.parsePoolCreatedEvent(*log, receipt.TxHash.Hex(), blockNum) // 解析池子创建事件
					// 存储池子数据
					if err := s.repo.SavePool(pool); err != nil {
						fmt.Printf("⚠️  保存失败: %v\n", err)
						continue
					}
					poolCount++
					// 将扫的池子更新到内存缓存中
					poolAddr := strings.ToLower(pool.PoolAddress)
					s.poolCache[poolAddr] = true
				}
			}
			/*
				Swap 事件
			*/
			if s.IsSwapEvent(*log) {
				// fmt.Printf("✅ 扫描区块 %d: 发现Swap事件\n", blockNum)
				swapEvent := s.parseSwapEvent(*log, receipt.TxHash.Hex(), blockNum, blockTimestamp)
				if err := s.repo.SaveSwapEvent(swapEvent); err != nil {
					fmt.Printf("⚠️  保存失败: %v\n", err)
					continue
				}
				swapCount++
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

// 判断是否为工厂合约
func (s *Scanner) isFactoryContract(address common.Address) bool {
	// factories := s.cfg.Syncswap.Factories.GetAllFactories()
	// addStr := strings.ToLower(address.Hex())
	// // 循环 判断当前的log.address 是否在工厂地址中
	// for _, factory := range factories {
	// 	if strings.ToLower(factory) == addStr {
	// 		return true
	// 	}
	// }

	// 既然我们已经做了映射了，那就不需要以上从配置中获取
	factoryAddr := strings.ToLower(address.Hex())
	_, ok := s.factoryInfoMap[factoryAddr] // 判断工厂地址是否在映射中
	return ok
}

// 判断当前的log是否为创建池子的事件
func (s *Scanner) isPoolCreatedEvent(log types.Log) bool {
	if len(log.Topics) < 3 {
		return false // 不是池子创建事件
	}

	factoryAddr := strings.ToLower(log.Address.Hex())
	info, ok := s.factoryInfoMap[factoryAddr]
	if !ok {
		return false // 不是我们监控的合约
	}
	// 对比所属工厂合约的事件哈希
	return log.Topics[0] == info.PoolCreatedSig

	/*	有逻辑，但是不够完善。
		每种事件都有唯一的哈希比如swap: 0x123abc...; poolcreated:0x0d3648bd...; Transfer:0xddf252ad...；
		但是不同类型的池子每种事件哈希不一定相同，所以我们需要选找出不同池子事件哈希，再遍历是否是我们监控的工厂合约的事件哈希
		所以我们可以通过判断log.Topics[0]事件类型哈希是否等于每种事件的唯一哈希来判断是否为池子创建事件
	*/
	// poolCreatedTopic := common.HexToHash("0x0d3648bd0f6ba80134a33ba9275ac585d9d315f0ad8355cddefde31afa28d0e9")
	// return log.Topics[0] == poolCreatedTopic // 判断log.Topics[0]事件类型哈希是否等于每种事件的唯一哈希
}

// 解析池子创建事件
func (s *Scanner) parsePoolCreatedEvent(log types.Log, txHash string, blockNum uint64) *models.Pool {
	/*
		zksync-era 池子创建在非indexed的log.Data中，所以需要从log.Data中解析出池子地址
		common.BytesToAddress 会自动处理填充，取最后20字节 这是solidity合约的地址格式
	*/
	poolAddress := common.BytesToAddress(log.Data).Hex()

	// 从映射的池子获取类型和版本
	factoryAddr := strings.ToLower(log.Address.Hex())
	info := s.factoryInfoMap[factoryAddr]
	return &models.Pool{
		PoolAddress:    poolAddress,
		FactoryAddress: log.Address.Hex(), //log.address根据topic[0]事件类型判断不同意义也不同，如果是transfer那就是代币合约地址，如果是swap就是池子合约地址，如果是poolcreated那么就是工厂合约地址
		Token0:         common.BytesToAddress(log.Topics[1].Bytes()).Hex(),
		Token1:         common.BytesToAddress(log.Topics[2].Bytes()).Hex(),
		CreatedTx:      txHash,
		CreatedBlock:   blockNum,
		PoolType:       info.PoolType,
		Version:        info.Version,
	}
}

/*
判断是否为Swap事件
硬编码方式判断
*/
func (s *Scanner) IsSwapEvent(log types.Log) bool {
	if len(log.Topics) < 3 { // 至少3个topic 才可能是swap事件
		return false
	}
	//先判断是否在swap事件哈希集合中，确定是swap事件
	isSwapSignatrue := false
	for _, sig := range s.swapSignatures {
		if log.Topics[0] == sig {
			isSwapSignatrue = true
			break
		}
	}
	if !isSwapSignatrue {
		return false
	}
	// return isSwapSignatrue

	// 以上判断了是swap事件，但是不一定是我们syncwap项目所需监控的池子，所以要再判断是我们池子中的swap事件。
	// 既然是swap事件，那么log.address就是池子地址，跟内存缓存的池子列表做对比
	poolAddr := strings.ToLower(log.Address.Hex())
	return s.poolCache[poolAddr] //处理命中的缓存的池子

}

// 解析Swap事件
func (s *Scanner) parseSwapEvent(log types.Log, txHash string, blockNum uint64, blockTimestamp int64) *models.SwapEvent {
	poolAddr := log.Address.Hex()
	sender := common.BytesToAddress(log.Topics[1].Bytes()).Hex()
	recipient := common.BytesToAddress(log.Topics[2].Bytes()).Hex()

	var token0, token1 string // 判断这笔交易的池子，谁是输入代币token0，谁是输出代币token1
	pool, err := s.repo.GetPoolByAddress(poolAddr)
	if err == nil && pool != nil {
		token0 = pool.Token0
		token1 = pool.Token1
	}

	// 解析 Data 字段获取交易金额
	// Data 包含：amount0In(32字节) + amount1In(32字节) + amount0Out(32字节) + amount1Out(32字节)
	var amount0In, amount1In, amount0Out, amount1Out string
	if len(log.Data) >= 128 {
		amount0In = new(big.Int).SetBytes(log.Data[0:32]).String()
		amount1In = new(big.Int).SetBytes(log.Data[32:64]).String()
		amount0Out = new(big.Int).SetBytes(log.Data[64:96]).String()
		amount1Out = new(big.Int).SetBytes(log.Data[96:128]).String()
	}

	// 判断这笔交易的方向，谁是输入代币，谁是输出代币
	var tokenIn, tokenOut, amountIn, amountOut string
	if amount0In != "0" && amount0In != "" { // 如果token0的amount0In输入金额不为0，则token0为输入代币，token1为输出代币，反之亦然
		tokenIn = token0
		tokenOut = token1
		amountIn = amount0In
		amountOut = amount1Out
	} else {
		tokenIn = token1
		tokenOut = token0
		amountIn = amount1In
		amountOut = amount0Out
	}

	return &models.SwapEvent{
		PoolAddress:    poolAddr,
		TxHash:         txHash,
		LogIndex:       int(log.Index),
		BlockNumber:    blockNum,
		BlockTimeStamp: blockTimestamp,
		Sender:         sender,
		Recipient:      recipient,
		TokenIn:        tokenIn,
		TokenOut:       tokenOut,
		AmountIn:       amountIn,
		AmountOut:      amountOut,
	}
}

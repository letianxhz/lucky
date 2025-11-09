package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	clog "github.com/cherry-game/cherry/logger"
	pomeloClient "github.com/cherry-game/cherry/net/parser/pomelo/client"
)

var (
	// 压测统计
	totalRequests   int64          // 总请求数
	successRequests int64          // 成功请求数
	failedRequests  int64          // 失败请求数
	totalLatency    int64          // 总延迟（毫秒）
	maxLatency      int64          // 最大延迟（毫秒）
	minLatency      int64 = 999999 // 最小延迟（毫秒）
	startTime       time.Time
	wg              sync.WaitGroup
)

// BenchmarkBuyItem 压测购买道具
func BenchmarkBuyItem(robotCount int, requestsPerRobot int) {
	url := "http://127.0.0.1:8081" // web node
	addr := "127.0.0.1:10011"      // 网关地址
	serverId := int32(10001)       // 测试的游戏服id
	pid := "2126001"               // 测试的sdk包id
	printLog := false              // 压测时不输出详细日志

	clog.Infof("========== 开始压测购买道具 ==========")
	clog.Infof("机器人数量: %d", robotCount)
	clog.Infof("每个机器人请求数: %d", requestsPerRobot)
	clog.Infof("总请求数: %d", robotCount*requestsPerRobot)

	startTime = time.Now()

	// 批量注册账号
	accounts := make(map[string]string)
	for i := 1; i <= robotCount; i++ {
		userName := fmt.Sprintf("bench%d", i)
		accounts[userName] = userName
	}
	RegisterDevAccount(url, accounts)
	time.Sleep(1 * time.Second)

	// 启动所有机器人
	for i := 1; i <= robotCount; i++ {
		userName := fmt.Sprintf("bench%d", i)
		password := userName
		wg.Add(1)
		go runBenchmarkRobot(url, pid, userName, password, addr, serverId, requestsPerRobot, printLog)
		// 错开启动时间，避免同时连接
		time.Sleep(10 * time.Millisecond)
	}

	// 等待所有机器人完成
	wg.Wait()

	// 输出统计结果
	printBenchmarkResults()
}

func runBenchmarkRobot(url, pid, userName, password, addr string, serverId int32, requestCount int, printLog bool) {
	defer wg.Done()

	// 创建客户端
	cli := New(
		pomeloClient.New(
			pomeloClient.WithRequestTimeout(10*time.Second),
			pomeloClient.WithErrorBreak(false), // 压测时不要因为错误中断
		),
	)
	cli.PrintLog = printLog

	// 1. 登录获取token
	if err := cli.GetToken(url, pid, userName, password); err != nil {
		atomic.AddInt64(&failedRequests, int64(requestCount))
		clog.Warnf("[%s] 获取 token 失败: %v", userName, err)
		return
	}

	// 2. 连接网关
	if err := cli.ConnectToTCP(addr); err != nil {
		atomic.AddInt64(&failedRequests, int64(requestCount))
		clog.Warnf("[%s] 连接网关失败: %v", userName, err)
		return
	}

	time.Sleep(100 * time.Millisecond)

	// 3. 用户登录
	if err := cli.UserLogin(serverId); err != nil {
		atomic.AddInt64(&failedRequests, int64(requestCount))
		clog.Warnf("[%s] 用户登录失败: %v", userName, err)
		return
	}

	time.Sleep(100 * time.Millisecond)

	// 4. 查看角色
	cli.PlayerSelect()

	time.Sleep(100 * time.Millisecond)

	// 5. 如果没有角色，创建角色
	if cli.PlayerId == 0 {
		if err := cli.ActorCreate(); err != nil {
			atomic.AddInt64(&failedRequests, int64(requestCount))
			clog.Warnf("[%s] 创建角色失败: %v", userName, err)
			return
		}
	}

	time.Sleep(100 * time.Millisecond)

	// 6. 角色进入游戏
	if err := cli.ActorEnter(); err != nil {
		atomic.AddInt64(&failedRequests, int64(requestCount))
		clog.Warnf("[%s] 角色进入游戏失败: %v", userName, err)
		return
	}

	time.Sleep(100 * time.Millisecond)

	// 7. 执行购买请求
	for i := 0; i < requestCount; i++ {
		reqStart := time.Now()

		// 随机选择道具和数量
		itemId := int32(1001)
		if i%2 == 0 {
			itemId = 1002
		}
		count := int32(1)
		payType := int32(1)

		err := cli.BuyItem(1, itemId, count, payType)

		latency := time.Since(reqStart).Milliseconds()
		atomic.AddInt64(&totalRequests, 1)
		atomic.AddInt64(&totalLatency, latency)

		// 更新最大最小延迟
		for {
			oldMax := atomic.LoadInt64(&maxLatency)
			if latency > oldMax {
				if atomic.CompareAndSwapInt64(&maxLatency, oldMax, latency) {
					break
				}
			} else {
				break
			}
		}

		for {
			oldMin := atomic.LoadInt64(&minLatency)
			if latency < oldMin {
				if atomic.CompareAndSwapInt64(&minLatency, oldMin, latency) {
					break
				}
			} else {
				break
			}
		}

		if err != nil {
			atomic.AddInt64(&failedRequests, 1)
		} else {
			atomic.AddInt64(&successRequests, 1)
		}

		// 控制请求频率，避免过快
		time.Sleep(50 * time.Millisecond)
	}

	cli.Disconnect()
}

func printBenchmarkResults() {
	total := atomic.LoadInt64(&totalRequests)
	success := atomic.LoadInt64(&successRequests)
	failed := atomic.LoadInt64(&failedRequests)
	totalLat := atomic.LoadInt64(&totalLatency)
	maxLat := atomic.LoadInt64(&maxLatency)
	minLat := atomic.LoadInt64(&minLatency)

	elapsed := time.Since(startTime).Seconds()
	avgLatency := float64(0)
	if total > 0 {
		avgLatency = float64(totalLat) / float64(total)
	}
	qps := float64(total) / elapsed

	clog.Infof("========== 压测结果统计 ==========")
	clog.Infof("总请求数: %d", total)
	clog.Infof("成功请求: %d (%.2f%%)", success, float64(success)/float64(total)*100)
	clog.Infof("失败请求: %d (%.2f%%)", failed, float64(failed)/float64(total)*100)
	clog.Infof("总耗时: %.2f 秒", elapsed)
	clog.Infof("QPS: %.2f 请求/秒", qps)
	clog.Infof("平均延迟: %.2f 毫秒", avgLatency)
	clog.Infof("最大延迟: %d 毫秒", maxLat)
	clog.Infof("最小延迟: %d 毫秒", minLat)
	clog.Infof("====================================")
}

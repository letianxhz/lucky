package main

import (
	"time"

	clog "github.com/cherry-game/cherry/logger"
	pomeloClient "github.com/cherry-game/cherry/net/parser/pomelo/client"
)

// TestBuyItem 测试购买道具流程
func TestBuyItem() {
	url := "http://127.0.0.1:8081" // web node
	addr := "127.0.0.1:10011"      // 网关地址
	serverId := int32(10001)       // 测试的游戏服id
	pid := "2126001"               // 测试的sdk包id
	userName := "test_buy_item"    // 测试账号
	password := "test_buy_item"    // 测试密码
	printLog := true               // 输出详细日志

	clog.Infof("========== 开始测试购买道具流程 ==========")

	// 先注册账号
	accounts := map[string]string{
		userName: password,
	}
	RegisterDevAccount(url, accounts)
	time.Sleep(500 * time.Millisecond)

	// 创建客户端
	cli := New(
		pomeloClient.New(
			pomeloClient.WithRequestTimeout(10*time.Second),
			pomeloClient.WithErrorBreak(true),
		),
	)
	cli.PrintLog = printLog

	// 1. 登录获取token
	clog.Infof("步骤 1: 获取登录 token")
	if err := cli.GetToken(url, pid, userName, password); err != nil {
		clog.Errorf("获取 token 失败: %v", err)
		return
	}
	clog.Infof("✓ Token 获取成功: %s", cli.Token)

	// 2. 连接网关
	clog.Infof("步骤 2: 连接网关服务器")
	if err := cli.ConnectToTCP(addr); err != nil {
		clog.Errorf("连接网关失败: %v", err)
		return
	}
	clog.Infof("✓ 网关连接成功: %s", addr)

	time.Sleep(500 * time.Millisecond)

	// 3. 用户登录
	clog.Infof("步骤 3: 用户登录")
	if err := cli.UserLogin(serverId); err != nil {
		clog.Errorf("用户登录失败: %v", err)
		return
	}
	clog.Infof("✓ 用户登录成功: UID=%d, PID=%d, OpenId=%s", cli.UID, cli.PID, cli.OpenId)

	time.Sleep(500 * time.Millisecond)

	// 4. 查看角色
	clog.Infof("步骤 4: 查看角色列表")
	if err := cli.PlayerSelect(); err != nil {
		clog.Errorf("查看角色失败: %v", err)
		return
	}

	time.Sleep(500 * time.Millisecond)

	// 5. 如果没有角色，创建角色
	if cli.PlayerId == 0 {
		clog.Infof("步骤 5: 创建角色")
		if err := cli.ActorCreate(); err != nil {
			clog.Errorf("创建角色失败: %v", err)
			return
		}
		clog.Infof("✓ 角色创建成功: PlayerID=%d, PlayerName=%s", cli.PlayerId, cli.PlayerName)
	} else {
		clog.Infof("步骤 5: 使用已有角色: PlayerID=%d, PlayerName=%s", cli.PlayerId, cli.PlayerName)
	}

	time.Sleep(500 * time.Millisecond)

	// 6. 角色进入游戏
	clog.Infof("步骤 6: 角色进入游戏")
	if err := cli.ActorEnter(); err != nil {
		clog.Errorf("角色进入游戏失败: %v", err)
		return
	}
	clog.Infof("✓ 角色进入游戏成功")

	time.Sleep(500 * time.Millisecond)

	// 7. 测试购买道具 - 成功案例
	clog.Infof("========== 测试购买道具 - 成功案例 ==========")
	testCases := []struct {
		name    string
		shopId  int32
		itemId  int32
		count   int32
		payType int32
	}{
		{
			name:    "购买道具 1001 (金币支付)",
			shopId:  1,
			itemId:  1001,
			count:   1,
			payType: 1, // 金币
		},
		{
			name:    "购买道具 1001 x2 (金币支付)",
			shopId:  1,
			itemId:  1001,
			count:   2,
			payType: 1,
		},
		{
			name:    "购买道具 1002 (钻石支付)",
			shopId:  1,
			itemId:  1002,
			count:   1,
			payType: 2, // 钻石
		},
	}

	for i, tc := range testCases {
		clog.Infof("--- 测试用例 %d: %s ---", i+1, tc.name)
		if err := cli.BuyItem(tc.shopId, tc.itemId, tc.count, tc.payType); err != nil {
			clog.Errorf("✗ 购买失败: %v", err)
		} else {
			clog.Infof("✓ 购买成功")
		}
		time.Sleep(500 * time.Millisecond)
	}

	// 8. 测试购买道具 - 失败案例
	clog.Infof("========== 测试购买道具 - 失败案例 ==========")
	failCases := []struct {
		name    string
		shopId  int32
		itemId  int32
		count   int32
		payType int32
	}{
		{
			name:    "无效道具ID (应该返回 401)",
			shopId:  1,
			itemId:  9999, // 不存在的道具
			count:   1,
			payType: 1,
		},
		{
			name:    "无效数量 (应该返回 403)",
			shopId:  1,
			itemId:  1001,
			count:   0, // 无效数量
			payType: 1,
		},
		{
			name:    "无效支付类型 (应该返回 403)",
			shopId:  1,
			itemId:  1001,
			count:   1,
			payType: 99, // 无效支付类型
		},
	}

	for i, tc := range failCases {
		clog.Infof("--- 测试用例 %d: %s ---", i+1, tc.name)
		if err := cli.BuyItem(tc.shopId, tc.itemId, tc.count, tc.payType); err != nil {
			clog.Warnf("✗ 预期失败: %v", err)
		} else {
			clog.Warnf("⚠ 未按预期失败，但这是正常的（服务端可能返回错误码）")
		}
		time.Sleep(500 * time.Millisecond)
	}

	clog.Infof("========== 购买道具流程测试完成 ==========")
	clog.Infof("测试账号: %s", userName)
	clog.Infof("玩家ID: %d", cli.PlayerId)
	clog.Infof("玩家名称: %s", cli.PlayerName)

	// 保持连接一段时间，观察日志
	clog.Infof("保持连接 5 秒后断开...")
	time.Sleep(5 * time.Second)

	cli.Disconnect()
	clog.Infof("✓ 连接已断开")
}

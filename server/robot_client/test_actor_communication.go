package main

import (
	"fmt"
	"time"

	clog "github.com/cherry-game/cherry/logger"
	pomeloClient "github.com/cherry-game/cherry/net/parser/pomelo/client"
	"lucky/server/gen/msg"
)

// TestActorCommunication 测试 Player Actor 和 Room Actor 之间的通信
func TestActorCommunication() {
	url := "http://127.0.0.1:8081" // web node
	addr := "127.0.0.1:10011"      // 网关地址
	serverId := int32(10001)       // 测试的游戏服id
	pid := "2126001"               // 测试的sdk包id
	// 使用固定账号，角色名使用时间戳确保唯一
	userName := "test_actor_comm" // 测试账号
	password := "test_actor_comm" // 测试密码
	printLog := true              // 输出详细日志

	clog.Infof("========== 开始测试 Actor 通信 ==========")

	// 先注册账号（如果不存在）
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
	clog.Infof("✓ Token 获取成功")

	// 2. 连接网关
	clog.Infof("步骤 2: 连接网关服务器")
	if err := cli.ConnectToTCP(addr); err != nil {
		clog.Errorf("连接网关失败: %v", err)
		return
	}
	clog.Infof("✓ 网关连接成功")

	time.Sleep(500 * time.Millisecond)

	// 3. 用户登录
	clog.Infof("步骤 3: 用户登录")
	if err := cli.UserLogin(serverId); err != nil {
		clog.Errorf("用户登录失败: %v", err)
		return
	}
	clog.Infof("✓ 用户登录成功: UID=%d", cli.UID)

	time.Sleep(500 * time.Millisecond)

	// 4. 查看角色
	clog.Infof("步骤 4: 查看角色列表")
	if err := cli.PlayerSelect(); err != nil {
		clog.Errorf("查看角色失败: %v", err)
		return
	}

	time.Sleep(500 * time.Millisecond)

	// 5. 如果没有角色，创建角色（使用时间戳生成唯一角色名）
	if cli.PlayerId == 0 {
		clog.Infof("步骤 5: 创建角色")
		// 使用时间戳生成唯一的角色名
		timestamp := time.Now().Unix()
		playerName := fmt.Sprintf("actor_test_%d", timestamp)
		if err := cli.ActorCreateWithName(playerName); err != nil {
			clog.Errorf("创建角色失败: %v", err)
			return
		}
		clog.Infof("✓ 角色创建成功: PlayerID=%d, PlayerName=%s", cli.PlayerId, cli.PlayerName)
	} else {
		clog.Infof("步骤 5: 使用已有角色: PlayerID=%d", cli.PlayerId)
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

	// 7. 测试 Player Actor 和 Room Actor 通信
	clog.Infof("========== 测试 Actor 通信 ==========")

	// 7.1 测试加入房间（Player Actor -> Room Actor）
	clog.Infof("--- 测试 1: Player Actor 调用 Room Actor 的 joinRoom ---")
	joinReq := &msg.JoinRoomRequest{
		RoomId: 1, // 房间ID
	}
	msg, err := cli.Request("game.player.joinRoom", joinReq)
	if err != nil {
		clog.Warnf("✗ 加入房间失败: %v", err)
	} else {
		var rsp msg.JoinRoomResponse
		if err := cli.Serializer().Unmarshal(msg.Data, &rsp); err == nil {
			clog.Infof("✓ 加入房间成功: roomId=%s, playerCount=%d, maxPlayers=%d, success=%v",
				rsp.RoomId, rsp.PlayerCount, rsp.MaxPlayers, rsp.Success)
		} else {
			clog.Infof("✓ 加入房间成功: %+v", msg)
		}
	}

	time.Sleep(500 * time.Millisecond)

	// 7.2 测试获取房间信息（Player Actor -> Room Actor）
	clog.Infof("--- 测试 2: Player Actor 调用 Room Actor 的 getRoomInfo ---")
	getInfoReq := &msg.GetRoomInfoRequest{}
	msg, err = cli.Request("game.player.getRoomInfo", getInfoReq)
	if err != nil {
		clog.Warnf("✗ 获取房间信息失败: %v", err)
	} else {
		var rsp msg.GetRoomInfoResponse
		if err := cli.Serializer().Unmarshal(msg.Data, &rsp); err == nil {
			clog.Infof("✓ 获取房间信息成功: roomId=%s, playerCount=%d, maxPlayers=%d, playerIds=%v",
				rsp.RoomId, rsp.PlayerCount, rsp.MaxPlayers, rsp.PlayerIds)
		} else {
			clog.Infof("✓ 获取房间信息成功: %+v", msg)
		}
	}

	time.Sleep(500 * time.Millisecond)

	// 7.3 测试离开房间（Player Actor -> Room Actor）
	clog.Infof("--- 测试 3: Player Actor 调用 Room Actor 的 leaveRoom ---")
	leaveReq := &msg.LeaveRoomRequest{
		PlayerId: int64(cli.UID),
	}
	msg, err = cli.Request("game.player.leaveRoom", leaveReq)
	if err != nil {
		clog.Warnf("✗ 离开房间失败: %v", err)
	} else {
		var rsp msg.LeaveRoomResponse
		if err := cli.Serializer().Unmarshal(msg.Data, &rsp); err == nil {
			clog.Infof("✓ 离开房间成功: success=%v, message=%s", rsp.Success, rsp.Message)
		} else {
			clog.Infof("✓ 离开房间成功: %+v", msg)
		}
	}

	clog.Infof("========== Actor 通信测试完成 ==========")
	clog.Infof("测试账号: %s", userName)
	clog.Infof("玩家ID: %d", cli.PlayerId)

	// 保持连接一段时间，观察日志
	clog.Infof("保持连接 3 秒后断开...")
	time.Sleep(3 * time.Second)

	cli.Disconnect()
	clog.Infof("✓ 连接已断开")
}

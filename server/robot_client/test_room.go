package main

import (
	"fmt"
	"sync"
	"time"

	clog "github.com/cherry-game/cherry/logger"
	pomeloClient "github.com/cherry-game/cherry/net/parser/pomelo/client"
	"lucky/server/gen/msg"
)

// TestRoom 测试房间功能
// 包括：创建房间、加入房间、获取房间信息、房间广播、离开房间
func TestRoom() {
	url := "http://127.0.0.1:8081" // web node
	addr := "127.0.0.1:10011"      // 网关地址
	serverId := int32(10001)       // 测试的游戏服id
	pid := "2126001"               // 测试的sdk包id
	printLog := true               // 输出详细日志

	clog.Infof("========== 开始测试 Room 功能 ==========")

	// 测试账号配置
	testAccounts := []struct {
		userName string
		password string
	}{
		{"test_room_1", "test_room_1"},
		{"test_room_2", "test_room_2"},
		{"test_room_3", "test_room_3"},
		{"test_room_4", "test_room_4"},
	}

	// 注册测试账号
	accounts := make(map[string]string)
	for _, acc := range testAccounts {
		accounts[acc.userName] = acc.password
	}
	RegisterDevAccount(url, accounts)
	time.Sleep(500 * time.Millisecond)

	// 创建多个机器人客户端
	robots := make([]*Robot, 0, len(testAccounts))
	for i, acc := range testAccounts {
		clog.Infof("========== 初始化机器人 %d: %s ==========", i+1, acc.userName)

		cli := New(
			pomeloClient.New(
				pomeloClient.WithRequestTimeout(10*time.Second),
				pomeloClient.WithErrorBreak(true),
			),
		)
		cli.PrintLog = printLog

		// 1. 获取登录 token
		if err := cli.GetToken(url, pid, acc.userName, acc.password); err != nil {
			clog.Errorf("机器人 %d 获取 token 失败: %v", i+1, err)
			continue
		}

		// 2. 连接网关
		if err := cli.ConnectToTCP(addr); err != nil {
			clog.Errorf("机器人 %d 连接网关失败: %v", i+1, err)
			continue
		}

		time.Sleep(200 * time.Millisecond)

		// 3. 用户登录
		if err := cli.UserLogin(serverId); err != nil {
			clog.Errorf("机器人 %d 用户登录失败: %v", i+1, err)
			continue
		}

		time.Sleep(200 * time.Millisecond)

		// 4. 查看角色列表
		if err := cli.PlayerSelect(); err != nil {
			clog.Warnf("机器人 %d 查看角色列表失败: %v", i+1, err)
		}

		time.Sleep(200 * time.Millisecond)

		// 5. 创建角色（如果不存在）
		if cli.PlayerId == 0 {
			timestamp := time.Now().Unix()
			playerName := fmt.Sprintf("room_test_%d_%d", i+1, timestamp)
			if err := cli.ActorCreateWithName(playerName); err != nil {
				clog.Errorf("机器人 %d 创建角色失败: %v", i+1, err)
				continue
			}
		}

		time.Sleep(200 * time.Millisecond)

		// 6. 角色进入游戏
		if err := cli.ActorEnter(); err != nil {
			clog.Errorf("机器人 %d 角色进入游戏失败: %v", i+1, err)
			continue
		}

		robots = append(robots, cli)
		clog.Infof("✓ 机器人 %d 初始化完成: UID=%d, PlayerID=%d", i+1, cli.UID, cli.PlayerId)
		time.Sleep(300 * time.Millisecond)
	}

	if len(robots) == 0 {
		clog.Errorf("没有成功初始化的机器人，退出测试")
		return
	}

	clog.Infof("========== 成功初始化 %d 个机器人 ==========", len(robots))
	time.Sleep(1 * time.Second)

	// 测试 1: 第一个机器人创建房间
	// 注意：Room Actor 是子 Actor，客户端无法直接访问
	// 需要通过 Player Actor 调用，或者直接跳过创建房间测试（房间会在加入时自动创建）
	clog.Infof("========== 测试 1: 创建房间（跳过，房间会在加入时自动创建） ==========")
	clog.Infof("注意：Room Actor 是子 Actor，客户端无法直接发送消息到 Room Actor")
	clog.Infof("房间会在第一个玩家加入时自动创建（通过 OnFindChild）")
	time.Sleep(500 * time.Millisecond)

	// 测试 2: 所有机器人加入房间
	clog.Infof("========== 测试 2: 多个机器人加入房间 ==========")
	roomId := int64(1) // 使用默认房间ID
	var wg sync.WaitGroup
	for i, robot := range robots {
		wg.Add(1)
		go func(idx int, r *Robot) {
			defer wg.Done()
			joinReq := &msg.JoinRoomRequest{
				RoomId: roomId,
			}
			msg, err := r.Request("game.player.joinRoom", joinReq)
			if err != nil {
				clog.Warnf("✗ 机器人 %d 加入房间失败: %v", idx+1, err)
			} else {
				var rsp msg.JoinRoomResponse
				if err := r.Serializer().Unmarshal(msg.Data, &rsp); err == nil {
					clog.Infof("✓ 机器人 %d 加入房间成功: roomId=%s, playerCount=%d, maxPlayers=%d, success=%v",
						idx+1, rsp.RoomId, rsp.PlayerCount, rsp.MaxPlayers, rsp.Success)
				} else {
					clog.Infof("✓ 机器人 %d 加入房间成功: %+v", idx+1, msg)
				}
			}
			time.Sleep(100 * time.Millisecond) // 避免并发冲突
		}(i, robot)
	}
	wg.Wait()
	time.Sleep(1 * time.Second)

	// 测试 3: 获取房间信息
	clog.Infof("========== 测试 3: 获取房间信息 ==========")
	if len(robots) > 0 {
		robot := robots[0]
		getInfoReq := &msg.GetRoomInfoRequest{}
		msg, err := robot.Request("game.player.getRoomInfo", getInfoReq)
		if err != nil {
			clog.Warnf("✗ 获取房间信息失败: %v", err)
		} else {
			var rsp msg.GetRoomInfoResponse
			if err := robot.Serializer().Unmarshal(msg.Data, &rsp); err == nil {
				clog.Infof("✓ 获取房间信息成功: roomId=%s, playerCount=%d, maxPlayers=%d, playerIds=%v",
					rsp.RoomId, rsp.PlayerCount, rsp.MaxPlayers, rsp.PlayerIds)
			} else {
				clog.Infof("✓ 获取房间信息成功: %+v", msg)
			}
		}
	}
	time.Sleep(500 * time.Millisecond)

	// 测试 4: 房间广播
	// 注意：Room Actor 是子 Actor，客户端无法直接访问
	// 需要通过 Player Actor 调用，或者暂时跳过
	clog.Infof("========== 测试 4: 房间广播（跳过，Room Actor 是子 Actor） ==========")
	clog.Infof("注意：房间广播功能需要通过 Player Actor 调用 Room Actor，或实现专门的 Player Handler")
	time.Sleep(500 * time.Millisecond)

	// 测试 5: 部分机器人离开房间
	clog.Infof("========== 测试 5: 部分机器人离开房间 ==========")
	leaveCount := len(robots) / 2
	if leaveCount == 0 {
		leaveCount = 1
	}
	for i := 0; i < leaveCount && i < len(robots); i++ {
		robot := robots[i]
		leaveReq := &msg.LeaveRoomRequest{
			PlayerId: int64(robot.UID),
		}
		msg, err := robot.Request("game.player.leaveRoom", leaveReq)
		if err != nil {
			clog.Warnf("✗ 机器人 %d 离开房间失败: %v", i+1, err)
		} else {
			var rsp msg.LeaveRoomResponse
			if err := robot.Serializer().Unmarshal(msg.Data, &rsp); err == nil {
				clog.Infof("✓ 机器人 %d 离开房间成功: success=%v, message=%s",
					i+1, rsp.Success, rsp.Message)
			} else {
				clog.Infof("✓ 机器人 %d 离开房间成功: %+v", i+1, msg)
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	time.Sleep(500 * time.Millisecond)

	// 测试 6: 再次获取房间信息，验证离开后的状态
	clog.Infof("========== 测试 6: 验证离开后的房间状态 ==========")
	if len(robots) > leaveCount {
		robot := robots[leaveCount]
		getInfoReq := &msg.GetRoomInfoRequest{}
		msg, err := robot.Request("game.player.getRoomInfo", getInfoReq)
		if err != nil {
			clog.Warnf("✗ 获取房间信息失败: %v", err)
		} else {
			var rsp msg.GetRoomInfoResponse
			if err := robot.Serializer().Unmarshal(msg.Data, &rsp); err == nil {
				clog.Infof("✓ 获取房间信息成功: roomId=%s, playerCount=%d, maxPlayers=%d, playerIds=%v",
					rsp.RoomId, rsp.PlayerCount, rsp.MaxPlayers, rsp.PlayerIds)
			} else {
				clog.Infof("✓ 获取房间信息成功: %+v", msg)
			}
		}
	}
	time.Sleep(500 * time.Millisecond)

	// 测试 7: 剩余机器人全部离开
	clog.Infof("========== 测试 7: 剩余机器人全部离开房间 ==========")
	for i := leaveCount; i < len(robots); i++ {
		robot := robots[i]
		leaveReq := &msg.LeaveRoomRequest{
			PlayerId: int64(robot.UID),
		}
		msg, err := robot.Request("game.player.leaveRoom", leaveReq)
		if err != nil {
			clog.Warnf("✗ 机器人 %d 离开房间失败: %v", i+1, err)
		} else {
			var rsp msg.LeaveRoomResponse
			if err := robot.Serializer().Unmarshal(msg.Data, &rsp); err == nil {
				clog.Infof("✓ 机器人 %d 离开房间成功: success=%v, message=%s",
					i+1, rsp.Success, rsp.Message)
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	time.Sleep(500 * time.Millisecond)

	clog.Infof("========== Room 功能测试完成 ==========")
	clog.Infof("测试了 %d 个机器人的房间功能", len(robots))

	// 保持连接一段时间，观察日志
	clog.Infof("保持连接 2 秒后断开...")
	time.Sleep(2 * time.Second)

	// 断开所有机器人连接
	for i, robot := range robots {
		robot.Disconnect()
		clog.Debugf("机器人 %d 已断开连接", i+1)
	}

	clog.Infof("✓ 所有机器人已断开连接")
}

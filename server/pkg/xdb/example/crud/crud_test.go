package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"lucky/server/pkg/xdb"
	"lucky/server/pkg/xdb/example/pb"
	_ "lucky/server/pkg/xdb/storage/mysql"
)

func main() {
	ctx := context.Background()

	// 从环境变量获取 MySQL 配置
	dbName := os.Getenv("MYSQL_DB")
	if dbName == "" {
		dbName = "test"
	}
	host := os.Getenv("MYSQL_HOST")
	if host == "" {
		host = "localhost"
	}
	port := int32(3306)
	if p := os.Getenv("MYSQL_PORT"); p != "" {
		fmt.Sscanf(p, "%d", &port)
	}
	username := os.Getenv("MYSQL_USER")
	if username == "" {
		username = "root"
	}
	password := os.Getenv("MYSQL_PASSWORD")
	if password == "" {
		password = ""
	}

	fmt.Println("=== xdb CRUD 操作测试 ===")
	fmt.Printf("MySQL 配置: %s@%s:%d/%s\n\n", username, host, port, dbName)

	// 1. 初始化 xdb
	fmt.Println("步骤 1: 初始化 xdb...")
	err := SetupXdbWithMySQL(ctx, dbName, host, port, username, password)
	if err != nil {
		fmt.Printf("   ✗ 初始化失败: %v\n", err)
		return
	}
	fmt.Println("   ✓ xdb 初始化成功")
	fmt.Println()

	// 等待初始化完成
	time.Sleep(200 * time.Millisecond)

	// 测试用的玩家ID
	testPlayerId := int64(10001)

	// 2. CREATE - 创建玩家记录
	fmt.Println("步骤 2: CREATE - 创建玩家记录...")
	player, err := xdb.Create[*PlayerModel](ctx, &msg.Player{
		PlayerId: testPlayerId,
		Name:     "测试玩家",
		Level:    1,
		Exp:      0,
		Ctime:    time.Now().Unix(),
		Mtime:    time.Now().Unix(),
	})
	if err != nil {
		fmt.Printf("   ✗ 创建失败: %v\n", err)
	} else {
		fmt.Printf("   ✓ 创建成功: ID=%d, Name=%s, Level=%d\n",
			player.PlayerId, player.Name, player.Level)
	}
	fmt.Println()

	// 同步保存到数据库
	fmt.Println("   同步保存到数据库...")
	err = xdb.Sync(ctx, player)
	if err != nil {
		fmt.Printf("   ✗ 同步失败: %v\n", err)
	} else {
		fmt.Println("   ✓ 同步成功")
	}
	fmt.Println()

	// 3. READ - 获取玩家记录
	fmt.Println("步骤 3: READ - 获取玩家记录...")
	player2, err := xdb.Get[*PlayerModel](ctx, testPlayerId)
	if err != nil {
		fmt.Printf("   ✗ 获取失败: %v\n", err)
	} else {
		fmt.Printf("   ✓ 获取成功: ID=%d, Name=%s, Level=%d, Exp=%d\n",
			player2.PlayerId, player2.Name, player2.Level, player2.Exp)
	}
	fmt.Println()

	// 4. UPDATE - 更新玩家记录
	fmt.Println("步骤 4: UPDATE - 更新玩家记录...")
	if player2 != nil {
		// 更新玩家信息
		player2.Name = "更新后的玩家"
		player2.Level = 10
		player2.Exp = 1000
		player2.Mtime = time.Now().Unix()

		// 标记变更的字段
		player2.GetHeader().SetChanged(
			msg.PlayerFieldName,
			msg.PlayerFieldLevel,
			msg.PlayerFieldExp,
			msg.PlayerFieldMtime,
		)

		// 保存变更
		xdb.Save(ctx, player2)
		fmt.Printf("   ✓ 更新成功: Name=%s, Level=%d, Exp=%d\n",
			player2.Name, player2.Level, player2.Exp)

		// 同步保存
		err = xdb.Sync(ctx, player2)
		if err != nil {
			fmt.Printf("   ✗ 同步失败: %v\n", err)
		} else {
			fmt.Println("   ✓ 同步成功")
		}
	}
	fmt.Println()

	// 5. 再次读取验证更新
	fmt.Println("步骤 5: 验证更新 - 再次读取玩家记录...")
	player3, err := xdb.Get[*pPlayerModel](ctx, testPlayerId)
	if err != nil {
		fmt.Printf("   ✗ 获取失败: %v\n", err)
	} else {
		fmt.Printf("   ✓ 验证成功: Name=%s, Level=%d, Exp=%d\n",
			player3.Name, player3.Level, player3.Exp)
		if player3.Name == "更新后的玩家" && player3.Level == 10 {
			fmt.Println("   ✓ 数据更新正确")
		} else {
			fmt.Println("   ✗ 数据更新不正确")
		}
	}
	fmt.Println()

	// 6. DELETE - 删除玩家记录
	fmt.Println("步骤 6: DELETE - 删除玩家记录...")
	if player3 != nil {
		// 标记为删除
		deleted := player3.Delete(ctx)
		if deleted {
			fmt.Println("   ✓ 标记删除成功")

			// 保存删除操作
			xdb.Save(ctx, player3)

			// 同步保存
			err = xdb.Sync(ctx, player3)
			if err != nil {
				fmt.Printf("   ✗ 同步失败: %v\n", err)
			} else {
				fmt.Println("   ✓ 删除同步成功")
			}
		} else {
			fmt.Println("   ✗ 标记删除失败")
		}
	}
	fmt.Println()

	// 7. 验证删除 - 尝试再次获取
	fmt.Println("步骤 7: 验证删除 - 尝试再次获取玩家记录...")
	player4, err := xdb.Get[*PlayerModel](ctx, testPlayerId)
	if err != nil {
		fmt.Printf("   ✓ 获取失败（预期）: %v\n", err)
		fmt.Println("   ✓ 删除验证成功：记录已不存在")
	} else if player4 == nil {
		fmt.Println("   ✓ 删除验证成功：记录为 nil")
	} else {
		fmt.Printf("   ⚠ 记录仍存在: %v\n", player4)
	}
	fmt.Println()

	// 8. 清理资源
	fmt.Println("步骤 8: 清理资源...")
	xdb.Stop(ctx)
	fmt.Println("   ✓ 清理完成")
	fmt.Println()

	fmt.Println("=== CRUD 测试完成 ===")
}

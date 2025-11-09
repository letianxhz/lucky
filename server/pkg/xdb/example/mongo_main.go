package main

import (
	"context"
	"fmt"
	"log"
	"lucky/server/pkg/xdb/example/pb"
	"os"
	"time"

	"lucky/server/pkg/xdb"
)

func main() {
	ctx := context.Background()

	// 从环境变量获取 MongoDB URI，默认为本地
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	fmt.Println("=== xdb MongoDB 测试 ===")
	fmt.Printf("MongoDB URI: %s\n\n", mongoURI)

	// 1. 初始化 xdb with MongoDB
	fmt.Println("步骤 1: 初始化 xdb 模块（使用 MongoDB）...")
	err := SetupXdbWithMongo(ctx, mongoURI)
	if err != nil {
		log.Fatalf("初始化 xdb 失败: %v", err)
	}
	fmt.Println("   ✓ xdb 初始化成功\n")

	// 等待初始化完成
	time.Sleep(200 * time.Millisecond)

	// 2. 检查 Source 注册
	fmt.Println("步骤 2: 检查 Source 注册...")
	sources := xdb.Sources()
	fmt.Printf("   已注册的 Source 数量: %d\n", len(sources))

	for _, src := range sources {
		fmt.Printf("   ✓ %s\n", src.Namespace)
		fmt.Printf("     - TableName: %s\n", src.TableName)
		fmt.Printf("     - DriverName: %s\n", src.DriverName)
		fmt.Printf("     - KeySize: %d\n", src.KeySize)
	}
	fmt.Println()

	// 3. 测试 Player Source
	fmt.Println("步骤 3: 测试 Player Source...")
	playerSource := xdb.GetSourceByNS("player")
	if playerSource == nil {
		log.Fatal("Player Source 未注册")
	}

	fmt.Printf("   ✓ Player Source 已注册\n")

	// 测试 PK 创建
	pk, err := playerSource.PKCreator([]interface{}{int64(1001)})
	if err != nil {
		log.Printf("     ✗ PK 创建失败: %v\n", err)
	} else {
		fmt.Printf("     ✓ PK 创建成功: %s\n", pk.String())
	}
	fmt.Println()

	// 4. 测试创建记录
	fmt.Println("步骤 4: 测试创建 Player 记录...")
	player, err := xdb.Create[*PlayerModel](ctx, &msg.Player{
		PlayerId: 1001,
		Name:     "TestPlayer",
		Level:    1,
		Exp:      0,
		Ctime:    time.Now().Unix(),
		Mtime:    time.Now().Unix(),
	})
	if err != nil {
		log.Printf("   创建失败: %v\n", err)
	} else {
		fmt.Printf("   ✓ 创建成功\n")
		fmt.Printf("     - XId: %s\n", player.XId())
		fmt.Printf("     - Name: %s\n", player.Name)
		fmt.Printf("     - Level: %d\n", player.Level)
	}
	fmt.Println()

	// 5. 测试保存
	if player != nil {
		fmt.Println("步骤 5: 测试保存到 MongoDB...")
		xdb.Save(ctx, player)
		fmt.Println("   ✓ 保存请求已提交")
		fmt.Println()
	}

	// 6. 测试同步
	if player != nil {
		fmt.Println("步骤 6: 测试同步保存...")
		err = xdb.Sync(ctx, player)
		if err != nil {
			log.Printf("   同步失败: %v\n", err)
		} else {
			fmt.Println("   ✓ 同步成功")
		}
		fmt.Println()
	}

	// 7. 测试获取记录
	fmt.Println("步骤 7: 测试从 MongoDB 获取记录...")
	player2, err := xdb.Get[*PlayerModel](ctx, int64(1001))
	if err != nil {
		log.Printf("   获取失败: %v\n", err)
	} else if player2 != nil {
		fmt.Printf("   ✓ 获取成功\n")
		fmt.Printf("     - XId: %s\n", player2.XId())
		fmt.Printf("     - Name: %s\n", player2.Name)
		fmt.Printf("     - Level: %d\n", player2.Level)
	} else {
		fmt.Println("   ⚠ 记录不存在")
	}
	fmt.Println()

	// 8. 清理
	fmt.Println("步骤 8: 清理资源...")
	xdb.Stop(ctx)
	fmt.Println("   ✓ 清理完成")
	fmt.Println()

	fmt.Println("=== 测试完成 ===")
	fmt.Println("\n注意:")
	fmt.Println("  - 确保 MongoDB 服务正在运行")
	fmt.Println("  - 可以通过环境变量 MONGO_URI 指定 MongoDB 连接地址")
	fmt.Println("  - 例如: MONGO_URI=mongodb://localhost:27017 go run mongo_test.go ...")
}

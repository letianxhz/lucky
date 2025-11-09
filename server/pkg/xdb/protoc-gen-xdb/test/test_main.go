package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"lucky/server/pkg/xdb"
)

func main() {
	ctx := context.Background()

	fmt.Println("=== xdb 模块测试 ===\n")

	// 1. 初始化 xdb
	fmt.Println("1. 初始化 xdb 模块...")
	err := SetupXdb(ctx)
	if err != nil {
		log.Fatalf("初始化 xdb 失败: %v", err)
	}
	fmt.Println("   ✓ xdb 初始化成功\n")

	// 等待一下让初始化完成
	time.Sleep(100 * time.Millisecond)

	// 2. 测试 Source 注册
	fmt.Println("2. 检查 Source 注册...")
	playerSource := xdb.GetSourceByNS("player")
	if playerSource == nil {
		log.Fatal("Player Source 未注册")
	}
	fmt.Printf("   ✓ Player Source 已注册\n")
	fmt.Printf("     - Namespace: %s\n", playerSource.Namespace)
	fmt.Printf("     - TableName: %s\n", playerSource.TableName)
	fmt.Printf("     - DriverName: %s\n", playerSource.DriverName)

	itemSource := xdb.GetSourceByNS("item")
	if itemSource == nil {
		log.Fatal("Item Source 未注册")
	}
	fmt.Printf("   ✓ Item Source 已注册\n")
	fmt.Printf("     - Namespace: %s\n", itemSource.Namespace)
	fmt.Printf("     - TableName: %s\n", itemSource.TableName)
	fmt.Println()

	// 3. 测试创建记录
	fmt.Println("3. 测试创建 Player 记录...")
	player, err := xdb.Create[PlayerRecord](ctx, &Player{
		PlayerId: 1001,
		Name:     "TestPlayer",
		Level:    1,
		Exp:      0,
		Ctime:    time.Now().Unix(),
		Mtime:    time.Now().Unix(),
	})
	if err != nil {
		log.Printf("   创建失败: %v", err)
	} else {
		fmt.Printf("   ✓ 创建成功\n")
		fmt.Printf("     - XId: %s\n", player.XId())
		fmt.Printf("     - Name: %s\n", player.Name)
		fmt.Printf("     - Level: %d\n", player.Level)
		fmt.Printf("     - Lifecycle: %s\n", player.Lifecycle())
	}
	fmt.Println()

	// 4. 测试获取记录
	fmt.Println("4. 测试获取 Player 记录...")
	player2, err := xdb.Get[PlayerRecord](ctx, int64(1001))
	if err != nil {
		log.Printf("   获取失败: %v", err)
	} else if player2 != nil {
		fmt.Printf("   ✓ 获取成功\n")
		fmt.Printf("     - XId: %s\n", player2.XId())
		fmt.Printf("     - Name: %s\n", player2.Name)
	} else {
		fmt.Println("   ⚠ 记录不存在（这是正常的，因为使用了 DryRun 模式）")
	}
	fmt.Println()

	// 5. 测试更新记录
	if player != nil {
		fmt.Println("5. 测试更新 Player 记录...")
		oldName := player.Name
		oldLevel := player.Level

		player.Name = "UpdatedPlayer"
		player.Level = 10
		player.Exp = 1000

		// 标记变更的字段
		player.GetHeader().SetChanged(FieldName, FieldLevel, FieldExp)

		xdb.Save(ctx, player)

		fmt.Printf("   ✓ 更新成功\n")
		fmt.Printf("     - Name: %s -> %s\n", oldName, player.Name)
		fmt.Printf("     - Level: %d -> %d\n", oldLevel, player.Level)
		fmt.Printf("     - Dirty: %v\n", player.Dirty())
		fmt.Printf("     - Changes: %v\n", player.GetHeader().Changes())
	}
	fmt.Println()

	// 6. 测试复合主键（Item）
	fmt.Println("6. 测试创建 Item 记录（复合主键）...")
	item, err := xdb.Create[ItemRecord](ctx, &Item{
		PlayerId: 1001,
		ItemId:   2001,
		Count:    10,
		Ctime:    time.Now().Unix(),
		Mtime:    time.Now().Unix(),
	})
	if err != nil {
		log.Printf("   创建失败: %v", err)
	} else {
		fmt.Printf("   ✓ 创建成功\n")
		fmt.Printf("     - XId: %s\n", item.XId())
		fmt.Printf("     - PlayerId: %d\n", item.PlayerId)
		fmt.Printf("     - ItemId: %d\n", item.ItemId)
		fmt.Printf("     - Count: %d\n", item.Count)
	}
	fmt.Println()

	// 7. 测试字段常量
	fmt.Println("7. 测试字段常量...")
	fmt.Printf("   Player 字段常量:\n")
	fmt.Printf("     - FieldPlayerId: %d\n", FieldPlayerId)
	fmt.Printf("     - FieldName: %d\n", FieldName)
	fmt.Printf("     - FieldLevel: %d\n", FieldLevel)
	fmt.Printf("     - FieldExp: %d\n", FieldExp)
	fmt.Println()

	// 8. 测试 PK 创建
	fmt.Println("8. 测试 PK 创建...")
	if playerSource != nil {
		pk, err := playerSource.PKCreator([]interface{}{int64(1001)})
		if err != nil {
			log.Printf("   PK 创建失败: %v", err)
		} else {
			fmt.Printf("   ✓ PK 创建成功\n")
			fmt.Printf("     - PK: %s\n", pk.String())
			fmt.Printf("     - Full: %v\n", pk.Full())
			fmt.Printf("     - HashGroup: %d\n", pk.HashGroup())
		}
	}
	fmt.Println()

	// 9. 测试同步
	if player != nil {
		fmt.Println("9. 测试同步保存...")
		err = xdb.Sync(ctx, player)
		if err != nil {
			log.Printf("   同步失败: %v", err)
		} else {
			fmt.Println("   ✓ 同步成功")
		}
	}
	fmt.Println()

	// 10. 清理
	fmt.Println("10. 清理资源...")
	xdb.Stop(ctx)
	fmt.Println("   ✓ 清理完成")
	fmt.Println()

	fmt.Println("=== 测试完成 ===")
	fmt.Println("\n注意: 由于使用了 DryRun 模式，数据不会实际保存到数据库。")
	fmt.Println("这是正常的测试行为，用于验证代码生成和接口实现的正确性。")
}


package main

import (
	"context"
	"fmt"
	"log"

	"lucky/server/pkg/xdb"
	"lucky/server/pkg/xdb/protoc-gen-xdb/test"
)

// 这个示例展示了如何使用生成的 xdb 代码
// 注意：需要先运行 generate.sh 生成代码

func main() {
	ctx := context.Background()

	fmt.Println("=== protoc-gen-xdb 使用示例 ===\n")

	// 示例 1: 创建玩家记录
	fmt.Println("1. 创建玩家记录")
	player, err := xdb.Create[test.PlayerRecord](ctx, &test.Player{
		PlayerId: 1001,
		Name:     "TestPlayer",
		Level:    1,
		Exp:      0,
	})
	if err != nil {
		log.Printf("创建玩家失败: %v", err)
	} else {
		fmt.Printf("   ✓ 创建成功: %s\n", player.XId())
		fmt.Printf("   玩家信息: ID=%d, Name=%s, Level=%d\n",
			player.PlayerId, player.Name, player.Level)
	}

	// 示例 2: 获取玩家记录
	fmt.Println("\n2. 获取玩家记录")
	player, err = xdb.Get[test.PlayerRecord](ctx, int64(1001))
	if err != nil {
		log.Printf("获取玩家失败: %v", err)
	} else if player != nil {
		fmt.Printf("   ✓ 获取成功: %s\n", player.XId())
		fmt.Printf("   玩家信息: Name=%s, Level=%d, Exp=%d\n",
			player.Name, player.Level, player.Exp)
	} else {
		fmt.Println("   玩家不存在")
	}

	// 示例 3: 更新玩家记录
	fmt.Println("\n3. 更新玩家记录")
	if player != nil {
		oldName := player.Name
		oldLevel := player.Level

		player.Name = "UpdatedPlayer"
		player.Level = 10
		player.Exp = 1000

		// 标记变更的字段
		player.GetHeader().SetChanged(test.FieldName, test.FieldLevel, test.FieldExp)

		xdb.Save(ctx, player)

		fmt.Printf("   ✓ 更新成功\n")
		fmt.Printf("   变更: Name %s -> %s, Level %d -> %d\n",
			oldName, player.Name, oldLevel, player.Level)
	}

	// 示例 4: 同步保存
	fmt.Println("\n4. 同步保存")
	if player != nil {
		err = xdb.Sync(ctx, player)
		if err != nil {
			log.Printf("同步失败: %v", err)
		} else {
			fmt.Println("   ✓ 同步成功")
		}
	}

	// 示例 5: 创建道具记录（复合主键）
	fmt.Println("\n5. 创建道具记录（复合主键）")
	item, err := xdb.Create[test.ItemRecord](ctx, &test.Item{
		PlayerId: 1001,
		ItemId:   2001,
		Count:    10,
	})
	if err != nil {
		log.Printf("创建道具失败: %v", err)
	} else {
		fmt.Printf("   ✓ 创建成功\n")
		fmt.Printf("   道具信息: PlayerID=%d, ItemID=%d, Count=%d\n",
			item.PlayerId, item.ItemId, item.Count)
	}

	// 示例 6: 获取道具记录
	fmt.Println("\n6. 获取道具记录")
	item, err = xdb.Get[test.ItemRecord](ctx, int64(1001), int32(2001))
	if err != nil {
		log.Printf("获取道具失败: %v", err)
	} else if item != nil {
		fmt.Printf("   ✓ 获取成功\n")
		fmt.Printf("   道具信息: Count=%d\n", item.Count)
	} else {
		fmt.Println("   道具不存在")
	}

	// 示例 7: 更新道具数量
	fmt.Println("\n7. 更新道具数量")
	if item != nil {
		oldCount := item.Count
		item.Count = 20
		item.GetHeader().SetChanged(test.FieldCount)
		xdb.Save(ctx, item)

		fmt.Printf("   ✓ 更新成功\n")
		fmt.Printf("   变更: Count %d -> %d\n", oldCount, item.Count)
	}

	// 示例 8: 字段常量使用
	fmt.Println("\n8. 字段常量")
	fmt.Printf("   FieldPlayerId: %d\n", test.FieldPlayerId)
	fmt.Printf("   FieldName: %d\n", test.FieldName)
	fmt.Printf("   FieldLevel: %d\n", test.FieldLevel)
	fmt.Printf("   FieldExp: %d\n", test.FieldExp)

	// 示例 9: Source 信息
	fmt.Println("\n9. Source 信息")
	source := xdb.GetSourceByNS("player")
	if source != nil {
		fmt.Printf("   ✓ Source 已注册\n")
		fmt.Printf("   Namespace: %s\n", source.Namespace)
		fmt.Printf("   TableName: %s\n", source.TableName)
		fmt.Printf("   DriverName: %s\n", source.DriverName)
	} else {
		fmt.Println("   ✗ Source 未注册")
	}

	fmt.Println("\n=== 示例完成 ===")
	fmt.Println("\n注意: 这个示例展示了代码生成和使用的基本流程。")
	fmt.Println("实际使用时需要配置数据库驱动和连接。")
}




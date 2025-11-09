package main

import (
	"context"
	"fmt"
	"time"

	"lucky/server/pkg/xdb"
)

// 简化的测试，不依赖生成的 proto 代码
func main() {
	ctx := context.Background()

	fmt.Println("=== xdb 模块配置和测试 ===\n")

	// 1. 初始化 xdb
	fmt.Println("1. 初始化 xdb 模块...")
	err := SetupXdb(ctx)
	if err != nil {
		fmt.Printf("   ✗ 初始化失败: %v\n", err)
		return
	}
	fmt.Println("   ✓ xdb 初始化成功\n")

	// 等待一下让初始化完成
	time.Sleep(100 * time.Millisecond)

	// 2. 测试 Source 注册
	fmt.Println("2. 检查 Source 注册...")
	
	// 获取所有已注册的 Source
	sources := xdb.Sources()
	fmt.Printf("   已注册的 Source 数量: %d\n", len(sources))
	
	for _, src := range sources {
		fmt.Printf("   ✓ Source: %s\n", src.Namespace)
		fmt.Printf("     - TableName: %s\n", src.TableName)
		fmt.Printf("     - DriverName: %s\n", src.DriverName)
		fmt.Printf("     - KeySize: %d\n", src.KeySize)
	}
	fmt.Println()

	// 3. 测试通过命名空间获取 Source
	fmt.Println("3. 测试通过命名空间获取 Source...")
	playerSource := xdb.GetSourceByNS("player")
	if playerSource != nil {
		fmt.Printf("   ✓ Player Source 获取成功\n")
		fmt.Printf("     - Namespace: %s\n", playerSource.Namespace)
		fmt.Printf("     - TableName: %s\n", playerSource.TableName)
	} else {
		fmt.Println("   ⚠ Player Source 未找到（可能未注册）")
	}

	itemSource := xdb.GetSourceByNS("item")
	if itemSource != nil {
		fmt.Printf("   ✓ Item Source 获取成功\n")
		fmt.Printf("     - Namespace: %s\n", itemSource.Namespace)
		fmt.Printf("     - TableName: %s\n", itemSource.TableName)
	} else {
		fmt.Println("   ⚠ Item Source 未找到（可能未注册）")
	}
	fmt.Println()

	// 4. 测试 PK 创建（如果 Source 存在）
	if playerSource != nil {
		fmt.Println("4. 测试 PK 创建...")
		pk, err := playerSource.PKCreator([]interface{}{int64(1001)})
		if err != nil {
			fmt.Printf("   ✗ PK 创建失败: %v\n", err)
		} else {
			fmt.Printf("   ✓ PK 创建成功\n")
			fmt.Printf("     - PK: %s\n", pk.String())
			fmt.Printf("     - Full: %v\n", pk.Full())
			fmt.Printf("     - HashGroup: %d\n", pk.HashGroup())
		}
		fmt.Println()
	}

	// 5. 测试配置器
	fmt.Println("5. 测试配置器...")
	configurator := &TestConfigurator{}
	redoOpts := configurator.RedoOptions()
	fmt.Printf("   ✓ RedoOptions:\n")
	fmt.Printf("     - Enabled: %v\n", redoOpts.Enabled)
	fmt.Printf("     - Dir: %s\n", redoOpts.Dir)
	
	tableOpts := configurator.TableOptions("none", "test")
	fmt.Printf("   ✓ TableOptions:\n")
	fmt.Printf("     - Concurrence: %d\n", tableOpts.Concurrence)
	fmt.Printf("     - SaveTimeout: %v\n", tableOpts.SaveTimeout)
	fmt.Printf("     - SyncInterval: %v\n", tableOpts.SyncInterval)
	fmt.Printf("     - DryRun: %v\n", configurator.DryRun())
	fmt.Println()

	// 6. 清理
	fmt.Println("6. 清理资源...")
	xdb.Stop(ctx)
	fmt.Println("   ✓ 清理完成")
	fmt.Println()

	fmt.Println("=== 测试完成 ===")
	fmt.Println("\n注意:")
	fmt.Println("  - 由于使用了 DryRun 模式，数据不会实际保存到数据库")
	fmt.Println("  - 这是正常的测试行为，用于验证配置和接口实现的正确性")
	fmt.Println("  - 要测试完整的 CRUD 功能，需要:")
	fmt.Println("    1. 生成完整的 proto Go 代码")
	fmt.Println("    2. 配置真实的数据库驱动")
	fmt.Println("    3. 设置 DryRun = false")
}


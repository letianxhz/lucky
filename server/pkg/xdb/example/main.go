package main

//func main() {
//	ctx := context.Background()
//
//	fmt.Println("=== xdb 完整功能示例 ===")
//	fmt.Println()
//
//	// 1. 初始化 xdb
//	fmt.Println("步骤 1: 初始化 xdb 模块...")
//	err := SetupXdb(ctx)
//	if err != nil {
//		log.Fatalf("初始化 xdb 失败: %v", err)
//	}
//	fmt.Println("   ✓ xdb 初始化成功")
//	fmt.Println()
//
//	// 等待初始化完成
//	time.Sleep(200 * time.Millisecond)
//
//	// 2. 检查 Source 注册
//	fmt.Println("步骤 2: 检查 Source 注册...")
//	sources := xdb.Sources()
//	fmt.Printf("   已注册的 Source 数量: %d\n", len(sources))
//
//	for _, src := range sources {
//		fmt.Printf("   ✓ %s\n", src.Namespace)
//		fmt.Printf("     - TableName: %s\n", src.TableName)
//		fmt.Printf("     - DriverName: %s\n", src.DriverName)
//		fmt.Printf("     - KeySize: %d\n", src.KeySize)
//	}
//	fmt.Println()
//
//	// 3. 测试 Player Source 和 CRUD 操作
//	fmt.Println("步骤 3: 测试 Player CRUD 操作...")
//	playerSource := xdb.GetSourceByNS("player")
//	if playerSource == nil {
//		fmt.Println("   ⚠ Player Source 未注册")
//		fmt.Println("   提示: 确保已导入 player_xdb.pb.go")
//	} else {
//		fmt.Printf("   ✓ Player Source 已注册\n")
//
//		// 测试 PK 创建
//		pk, err := playerSource.PKCreator([]interface{}{int64(1001)})
//		if err != nil {
//			fmt.Printf("     ✗ PK 创建失败: %v\n", err)
//		} else {
//			fmt.Printf("     ✓ PK 创建成功: %s\n", pk.String())
//			fmt.Printf("       - Full: %v\n", pk.Full())
//			fmt.Printf("       - HashGroup: %d\n", pk.HashGroup())
//		}
//
//		// 演示创建记录
//		fmt.Println("")
//		fmt.Println("   创建记录示例:")
//		fmt.Println("   player, err := xdb.Create[PlayerRecord](ctx, &Player{")
//		fmt.Println("       PlayerId: 1001,")
//		fmt.Println("       Name:     \"TestPlayer\",")
//		fmt.Println("       Level:    1,")
//		fmt.Println("       Exp:      0,")
//		fmt.Println("       Ctime:    time.Now().Unix(),")
//		fmt.Println("       Mtime:    time.Now().Unix(),")
//		fmt.Println("   })")
//	}
//	fmt.Println()
//
//	// 4. 测试 Item Source（复合主键）
//	fmt.Println("步骤 4: 测试 Item Source（复合主键）...")
//	itemSource := xdb.GetSourceByNS("item")
//	if itemSource == nil {
//		fmt.Println("   ⚠ Item Source 未注册")
//	} else {
//		fmt.Printf("   ✓ Item Source 已注册\n")
//		fmt.Printf("     - KeySize: %d (复合主键)\n", itemSource.KeySize)
//
//		// 测试复合主键创建
//		pk, err := itemSource.PKCreator([]interface{}{int64(1001), int32(2001)})
//		if err != nil {
//			fmt.Printf("     ✗ PK 创建失败: %v\n", err)
//		} else {
//			fmt.Printf("     ✓ 复合 PK 创建成功: %s\n", pk.String())
//			fmt.Printf("       - Full: %v\n", pk.Full())
//		}
//	}
//	fmt.Println()
//
//	// 5. 显示完整使用流程
//	fmt.Println("步骤 5: 完整使用流程...")
//	fmt.Println("")
//	fmt.Println("   // === 完整 CRUD 操作流程 ===")
//	fmt.Println("   ")
//	fmt.Println("   // 1. 创建记录")
//	fmt.Println("   player, err := xdb.Create[PlayerRecord](ctx, &Player{...})")
//	fmt.Println("   ")
//	fmt.Println("   // 2. 获取记录")
//	fmt.Println("   player, err = xdb.Get[PlayerRecord](ctx, int64(1001))")
//	fmt.Println("   ")
//	fmt.Println("   // 3. 更新记录")
//	fmt.Println("   player.Name = \"NewName\"")
//	fmt.Println("   player.GetHeader().SetChanged(FieldName)")
//	fmt.Println("   xdb.Save(ctx, player)")
//	fmt.Println("   ")
//	fmt.Println("   // 4. 同步保存")
//	fmt.Println("   err = xdb.Sync(ctx, player)")
//	fmt.Println()
//
//	// 6. 清理
//	fmt.Println("步骤 6: 清理资源...")
//	xdb.Stop(ctx)
//	fmt.Println("   ✓ 清理完成")
//	fmt.Println()
//
//	fmt.Println("=== 示例完成 ===")
//}

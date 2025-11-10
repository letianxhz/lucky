package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"lucky/server/pkg/xdb"
	_ "lucky/server/pkg/xdb/storage/mysql"
)

func main() {
	ctx := context.Background()

	// 从环境变量获取 MySQL 配置，默认为本地
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

	fmt.Println("=== xdb MySQL 测试 ===")
	fmt.Printf("MySQL 配置: %s@%s:%d/%s\n\n", username, host, port, dbName)

	// 1. 初始化 xdb with MySQL
	fmt.Println("步骤 1: 初始化 xdb 模块（使用 MySQL）...")
	err := SetupXdbWithMySQL(ctx, dbName, host, port, username, password)
	if err != nil {
		fmt.Printf("   ✗ 初始化失败: %v\n", err)
		return
	}
	fmt.Println("   ✓ xdb 初始化成功")
	fmt.Println()

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
		fmt.Println("   ⚠ Player Source 未注册")
	} else {
		fmt.Printf("   ✓ Player Source 已注册\n")
		fmt.Printf("     - DriverName: %s\n", playerSource.DriverName)

		// 测试 PK 创建
		pk, err := playerSource.PKCreator([]interface{}{int64(1001)})
		if err != nil {
			fmt.Printf("     ✗ PK 创建失败: %v\n", err)
		} else {
			fmt.Printf("     ✓ PK 创建成功: %s\n", pk.String())
		}
	}
	fmt.Println()

	// 4. 清理
	fmt.Println("步骤 4: 清理资源...")
	xdb.Stop(ctx)
	fmt.Println("   ✓ 清理完成")
	fmt.Println()

	fmt.Println("=== MySQL 测试完成 ===")
}

package main

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"time"

	_ "lucky/server/pkg/xdb/storage/mysql"
	"lucky/server/pkg/xdb"
	"lucky/server/pkg/xdb/example/pb"
)

func main() {
	ctx := context.Background()

	dbName := os.Getenv("MYSQL_DB")
	if dbName == "" {
		dbName = "test"
	}
	host := os.Getenv("MYSQL_HOST")
	if host == "" {
		host = "localhost"
	}
	port := int32(3306)
	username := os.Getenv("MYSQL_USER")
	if username == "" {
		username = "root"
	}
	password := os.Getenv("MYSQL_PASSWORD")
	if password == "" {
		password = ""
	}

	fmt.Println("=== 测试 Model 类型注册和获取 ===")

	err := SetupXdbWithMySQL(ctx, dbName, host, port, username, password)
	if err != nil {
		fmt.Printf("初始化失败: %v\n", err)
		return
	}

	time.Sleep(200 * time.Millisecond)

	testPlayerId := int64(10003)

	// 1. 使用 xdb.Create[*PlayerModel] 创建
	fmt.Println("\n1. 使用 xdb.Create[*PlayerModel] 创建...")
	player1, err := xdb.Create[*PlayerModel](ctx, &msg.Player{
		PlayerId: testPlayerId,
		Name:     "Model测试",
		Level:    5,
		Exp:      100,
		Ctime:    time.Now().Unix(),
		Mtime:    time.Now().Unix(),
	})
	if err != nil {
		fmt.Printf("   ✗ 创建失败: %v\n", err)
		return
	}
	fmt.Printf("   ✓ 创建成功，类型: %T\n", player1)
	xdb.Sync(ctx, player1)

	// 2. 使用 xdb.Get[*PlayerModel] 获取
	fmt.Println("\n2. 使用 xdb.Get[*PlayerModel] 获取...")
	player2, err := xdb.Get[*PlayerModel](ctx, testPlayerId)
	if err != nil {
		fmt.Printf("   ✗ 获取失败: %v\n", err)
		return
	}
	fmt.Printf("   ✓ 获取成功，类型: %T\n", player2)
	fmt.Printf("   ✓ 数据: ID=%d, Name=%s, Level=%d\n", player2.PlayerId, player2.Name, player2.Level)

	// 3. 验证类型
	fmt.Println("\n3. 验证类型...")
	if reflect.TypeOf(player2).String() == "*main.PlayerModel" {
		fmt.Println("   ✓ 类型正确：*main.PlayerModel")
	} else {
		fmt.Printf("   ✗ 类型错误：期望 *main.PlayerModel，实际 %T\n", player2)
	}

	// 4. 测试使用 *msg.PlayerRecord 会 panic（因为注册了 Model）
	fmt.Println("\n4. 测试使用 xdb.Get[*msg.PlayerRecord]（应该 panic）...")
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("   ✓ 正确 panic: %v\n", r)
			} else {
				fmt.Println("   ✗ 应该 panic 但没有")
			}
		}()
		_, _ = xdb.Get[*msg.PlayerRecord](ctx, testPlayerId)
	}()

	// 清理
	xdb.Stop(ctx)
	fmt.Println("\n=== 测试完成 ===")
}

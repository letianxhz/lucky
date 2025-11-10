package main

import (
	"context"
	"fmt"
	"os"
	"reflect"
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
	username := os.Getenv("MYSQL_USER")
	if username == "" {
		username = "root"
	}
	password := os.Getenv("MYSQL_PASSWORD")
	if password == "" {
		password = ""
	}

	fmt.Println("=== 测试 xdb.Get[*pb.PlayerRecord] 返回 PlayerModel ===")

	// 初始化 xdb
	err := SetupXdbWithMySQL(ctx, dbName, host, port, username, password)
	if err != nil {
		fmt.Printf("初始化失败: %v\n", err)
		return
	}

	time.Sleep(200 * time.Millisecond)

	testPlayerId := int64(10002)

	// 先创建一个记录
	player, err := xdb.Create[*PlayerModel](ctx, &pb.Player{
		PlayerId: testPlayerId,
		Name:     "测试玩家2",
		Level:    1,
		Exp:      0,
		Ctime:    time.Now().Unix(),
		Mtime:    time.Now().Unix(),
	})
	if err != nil {
		fmt.Printf("创建失败: %v\n", err)
		return
	}
	xdb.Sync(ctx, player)

	// 使用 xdb.Get[*pb.PlayerRecord] 获取
	player2, err := xdb.Get[*pb.PlayerRecord](ctx, testPlayerId)
	if err != nil {
		fmt.Printf("获取失败: %v\n", err)
		return
	}

	// 检查类型
	fmt.Printf("player2Any 类型: %T\n", player2Any)

	// 尝试转换为 *PlayerModel
	player2, ok := player2Any.(*PlayerModel)
	if !ok {
		fmt.Printf("✗ 无法转换为 *PlayerModel，实际类型: %T\n", player2Any)
		return
	}

	fmt.Printf("✓ 成功转换为 *PlayerModel: ID=%d, Name=%s\n", player2.PlayerId, player2.Name)

	// 由于 PlayerModel 嵌入了 PlayerRecord，可以通过反射获取
	val := reflect.ValueOf(player2)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// 检查是否是 PlayerModel 类型
	if val.Type().Name() == "PlayerModel" {
		fmt.Printf("✓ 返回的是 PlayerModel 类型\n")
		// 转换为 *PlayerModel
		pm := reflect.New(val.Type()).Elem()
		pm.Set(val)
		playerModel := pm.Addr().Interface().(*PlayerModel)
		fmt.Printf("✓ 成功转换为 *PlayerModel: ID=%d, Name=%s\n", playerModel.PlayerId, playerModel.Name)
	} else {
		fmt.Printf("✗ 返回的不是 PlayerModel 类型，实际类型: %T\n", player2)
	}

	xdb.Stop(ctx)
}

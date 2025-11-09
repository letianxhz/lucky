package test

import (
	"context"
	"fmt"
	"testing"

	"lucky/server/pkg/xdb"
)

// 这个测试展示了如何使用生成的 xdb 代码
// 注意：这需要先运行 generate.sh 生成代码

func ExamplePlayerRecord() {
	ctx := context.Background()

	// 1. 创建玩家记录
	player, err := xdb.Create[PlayerRecord](ctx, &Player{
		PlayerId: 1001,
		Name:     "TestPlayer",
		Level:    1,
		Exp:      0,
	})
	if err != nil {
		fmt.Printf("Error creating player: %v\n", err)
		return
	}

	fmt.Printf("Created player: %s\n", player.XId())

	// 2. 获取玩家记录
	player, err = xdb.Get[PlayerRecord](ctx, int64(1001))
	if err != nil {
		fmt.Printf("Error getting player: %v\n", err)
		return
	}

	if player != nil {
		fmt.Printf("Got player: %s, Level: %d\n", player.Name, player.Level)
	}

	// 3. 更新玩家记录
	player.Name = "UpdatedPlayer"
	player.Level = 10
	player.GetHeader().SetChanged(FieldName, FieldLevel)
	xdb.Save(ctx, player)

	fmt.Printf("Updated player: %s, Level: %d\n", player.Name, player.Level)

	// 4. 同步保存
	err = xdb.Sync(ctx, player)
	if err != nil {
		fmt.Printf("Error syncing player: %v\n", err)
		return
	}

	fmt.Println("Player synced successfully")
}

func ExampleItemRecord() {
	ctx := context.Background()

	// 创建道具记录（复合主键）
	item, err := xdb.Create[ItemRecord](ctx, &Item{
		PlayerId: 1001,
		ItemId:   2001,
		Count:    10,
	})
	if err != nil {
		fmt.Printf("Error creating item: %v\n", err)
		return
	}

	fmt.Printf("Created item: player_id=%d, item_id=%d, count=%d\n",
		item.PlayerId, item.ItemId, item.Count)

	// 获取道具记录
	item, err = xdb.Get[ItemRecord](ctx, int64(1001), int32(2001))
	if err != nil {
		fmt.Printf("Error getting item: %v\n", err)
		return
	}

	if item != nil {
		fmt.Printf("Got item: count=%d\n", item.Count)
	}

	// 更新道具数量
	item.Count = 20
	item.GetHeader().SetChanged(FieldCount)
	xdb.Save(ctx, item)

	fmt.Printf("Updated item count: %d\n", item.Count)
}

func TestFieldConstants(t *testing.T) {
	// 测试字段常量是否正确生成
	fmt.Println("Field constants:")
	fmt.Printf("  FieldPlayerId: %d\n", FieldPlayerId)
	fmt.Printf("  FieldName: %d\n", FieldName)
	fmt.Printf("  FieldLevel: %d\n", FieldLevel)
	fmt.Printf("  FieldExp: %d\n", FieldExp)
}

func TestPKCreation(t *testing.T) {
	// 测试 PK 创建
	pk, err := _PlayerSource.PKCreator([]interface{}{int64(1001)})
	if err != nil {
		t.Fatalf("Failed to create PK: %v", err)
	}

	playerPK := pk.(*PlayerPK)
	if playerPK.PlayerId != 1001 {
		t.Errorf("Expected PlayerId=1001, got %d", playerPK.PlayerId)
	}

	fmt.Printf("Created PK: %s\n", playerPK.String())
}

func TestSourceRegistration(t *testing.T) {
	// 测试 Source 是否正确注册
	source := xdb.GetSourceByNS("player")
	if source == nil {
		t.Fatal("Source not registered")
	}

	if source.Namespace != "player" {
		t.Errorf("Expected namespace 'player', got '%s'", source.Namespace)
	}

	if source.TableName != "player" {
		t.Errorf("Expected table name 'player', got '%s'", source.TableName)
	}

	fmt.Printf("Source registered: %s\n", source.Namespace)
}


package main

import (
	"context"
	"testing"
	"time"

	"lucky/server/pkg/xdb"
)

func TestXdbSetup(t *testing.T) {
	ctx := context.Background()

	// 初始化 xdb
	err := SetupXdb(ctx)
	if err != nil {
		t.Fatalf("初始化 xdb 失败: %v", err)
	}

	// 等待初始化完成
	time.Sleep(200 * time.Millisecond)

	// 检查 Source 注册
	sources := xdb.Sources()
	if len(sources) == 0 {
		t.Fatal("没有注册的 Source")
	}

	// 检查 Player Source
	playerSource := xdb.GetSourceByNS("player")
	if playerSource == nil {
		t.Fatal("Player Source 未注册")
	}

	if playerSource.Namespace != "player" {
		t.Errorf("Player Source Namespace 错误: 期望 'player', 得到 '%s'", playerSource.Namespace)
	}

	if playerSource.KeySize != 1 {
		t.Errorf("Player Source KeySize 错误: 期望 1, 得到 %d", playerSource.KeySize)
	}

	// 检查 Item Source
	itemSource := xdb.GetSourceByNS("item")
	if itemSource == nil {
		t.Fatal("Item Source 未注册")
	}

	if itemSource.KeySize != 2 {
		t.Errorf("Item Source KeySize 错误: 期望 2, 得到 %d", itemSource.KeySize)
	}

	// 清理
	xdb.Stop(ctx)
}

func TestPKCreation(t *testing.T) {
	ctx := context.Background()

	err := SetupXdb(ctx)
	if err != nil {
		t.Fatalf("初始化 xdb 失败: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	// 测试 Player PK
	playerSource := xdb.GetSourceByNS("player")
	if playerSource == nil {
		t.Fatal("Player Source 未注册")
	}

	pk, err := playerSource.PKCreator([]interface{}{int64(1001)})
	if err != nil {
		t.Fatalf("创建 Player PK 失败: %v", err)
	}

	if pk.String() != "player:1001" {
		t.Errorf("Player PK String() 错误: 期望 'player:1001', 得到 '%s'", pk.String())
	}

	if !pk.Full() {
		t.Error("Player PK Full() 应该返回 true")
	}

	// 测试 Item PK（复合主键）
	itemSource := xdb.GetSourceByNS("item")
	if itemSource == nil {
		t.Fatal("Item Source 未注册")
	}

	itemPK, err := itemSource.PKCreator([]interface{}{int64(1001), int32(2001)})
	if err != nil {
		t.Fatalf("创建 Item PK 失败: %v", err)
	}

	if itemPK.String() != "item:1001:2001" {
		t.Errorf("Item PK String() 错误: 期望 'item:1001:2001', 得到 '%s'", itemPK.String())
	}

	if !itemPK.Full() {
		t.Error("Item PK Full() 应该返回 true")
	}

	xdb.Stop(ctx)
}

func TestConfiguratorInterface(t *testing.T) {
	configurator := &TestConfigurator{}

	// 测试 RedoOptions
	redoOpts := configurator.RedoOptions()
	if redoOpts == nil {
		t.Fatal("RedoOptions 不应该为 nil")
	}

	if redoOpts.Enabled {
		t.Error("测试模式下 RedoOptions.Enabled 应该为 false")
	}

	// 测试 TableOptions
	tableOpts := configurator.TableOptions("none", "test")
	if tableOpts == nil {
		t.Fatal("TableOptions 不应该为 nil")
	}

	if tableOpts.Concurrence == 0 {
		t.Error("TableOptions.Concurrence 不应该为 0")
	}

	// 测试 DryRun
	if !configurator.DryRun() {
		t.Error("测试模式下 DryRun 应该返回 true")
	}
}


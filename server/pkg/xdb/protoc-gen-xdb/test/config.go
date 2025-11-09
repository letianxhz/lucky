package main

import (
	"context"
	"time"

	"lucky/server/pkg/xdb"
)

// TestConfigurator 测试用的配置器
type TestConfigurator struct{}

// RedoOptions 返回重做日志选项
func (c *TestConfigurator) RedoOptions() *xdb.RedoOptions {
	return &xdb.RedoOptions{
		Dir:          "./redo",
		Enabled:      false, // 测试时禁用重做日志
		SyncInterval: 100 * time.Millisecond,
	}
}

// DriverOptions 返回驱动选项
func (c *TestConfigurator) DriverOptions(driver string) interface{} {
	// 对于测试，返回简单的配置
	return map[string]interface{}{
		"driver": driver,
	}
}

// DaoOptions 返回 DAO 选项
func (c *TestConfigurator) DaoOptions(daoKey interface{}) interface{} {
	// 对于测试，返回简单的配置
	return map[string]interface{}{
		"daoKey": daoKey,
	}
}

// TableOptions 返回表选项
func (c *TestConfigurator) TableOptions(driver string, table string) *xdb.TableOptions {
	// 对于 "none" 驱动，返回 nil 或使用默认值
	if driver == "none" {
		return &xdb.TableOptions{
			DaoKey:      nil,
			Concurrence: 1,
			SaveTimeout: 5 * time.Second,
			SyncInterval: 100 * time.Millisecond,
		}
	}
	return &xdb.TableOptions{
		DaoKey:      "test",
		Concurrence: 2,                    // 并发数
		SaveTimeout: 5 * time.Second,      // 保存超时
		SyncInterval: 100 * time.Millisecond, // 同步间隔
	}
}

// DryRun 返回是否干运行（不实际保存）
func (c *TestConfigurator) DryRun() bool {
	return true // 测试时使用干运行模式，不实际保存到数据库
}

// SetupXdb 初始化 xdb
func SetupXdb(ctx context.Context) error {
	configurator := &TestConfigurator{}
	return xdb.Setup(ctx, configurator)
}


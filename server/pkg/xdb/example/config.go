package main

import (
	"context"
	"time"

	"lucky/server/pkg/xdb"
)

// TestConfigurator 测试用的配置器
type TestConfigurator struct{}

// InitializeDatabase 初始化数据库（可选实现）
// 如果使用真实数据库，可以在这里初始化数据库连接
func (c *TestConfigurator) InitializeDatabase() error {
	// 示例：如果需要初始化数据库，可以在这里调用
	// database.MustInitialize(config.Get().GetDatabase())
	// 测试模式下不需要初始化数据库
	return nil
}

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
	return map[string]interface{}{
		"driver": driver,
	}
}

// DaoOptions 返回 DAO 选项
func (c *TestConfigurator) DaoOptions(daoKey interface{}) interface{} {
	return map[string]interface{}{
		"daoKey": daoKey,
	}
}

// TableOptions 返回表选项
func (c *TestConfigurator) TableOptions(driver string, table string) *xdb.TableOptions {
	// 对于 "none" 驱动，返回默认值
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
		Concurrence: 2,
		SaveTimeout: 5 * time.Second,
		SyncInterval: 100 * time.Millisecond,
	}
}

// DryRun 返回是否干运行（不实际保存）
func (c *TestConfigurator) DryRun() bool {
	return true // 测试时使用干运行模式
}

// SetupXdb 初始化 xdb
func SetupXdb(ctx context.Context) error {
	configurator := &TestConfigurator{}
	return xdb.Setup(ctx, configurator)
}


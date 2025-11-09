package main

import (
	"context"
	"time"

	"lucky/server/pkg/xdb"
	"lucky/server/pkg/xdb/storage/mongo"
)

// MongoConfigurator MongoDB 配置器
type MongoConfigurator struct {
	MongoURI string // MongoDB 连接 URI
}

// InitializeDatabase 初始化数据库
func (c *MongoConfigurator) InitializeDatabase() error {
	// MongoDB 驱动会在 NewDao 时自动连接，这里可以添加其他初始化逻辑
	return nil
}

// RedoOptions 返回重做日志选项
func (c *MongoConfigurator) RedoOptions() *xdb.RedoOptions {
	return &xdb.RedoOptions{
		Dir:          "./redo",
		Enabled:      true, // 启用重做日志
		SyncInterval: 100 * time.Millisecond,
	}
}

// DriverOptions 返回驱动选项
func (c *MongoConfigurator) DriverOptions(driver string) interface{} {
	if driver == "mongo" {
		return &mongo.DriverOptions{}
	}
	return nil
}

// DaoOptions 返回 DAO 选项
func (c *MongoConfigurator) DaoOptions(daoKey interface{}) interface{} {
	if daoKey == "mongo" {
		uri := c.MongoURI
		if uri == "" {
			uri = "mongodb://localhost:27017" // 默认 URI
		}
		return &mongo.DaoOptions{
			URI:      uri,
			PoolSize: 10,
		}
	}
	return nil
}

// TableOptions 返回表选项
func (c *MongoConfigurator) TableOptions(driver string, table string) *xdb.TableOptions {
	if driver == "mongo" {
		return &xdb.TableOptions{
			DaoKey:      "mongo",
			Concurrence: 2,
			SaveTimeout: 5 * time.Second,
			SyncInterval: 100 * time.Millisecond,
		}
	}
	return nil
}

// DryRun 返回是否干运行
func (c *MongoConfigurator) DryRun() bool {
	return false // 使用真实数据库
}

// SetupXdbWithMongo 使用 MongoDB 初始化 xdb
func SetupXdbWithMongo(ctx context.Context, mongoURI string) error {
	configurator := &MongoConfigurator{
		MongoURI: mongoURI,
	}
	return xdb.Setup(ctx, configurator)
}


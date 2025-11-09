package main

import (
	"context"
	"time"

	"lucky/server/pkg/xdb"
	"lucky/server/pkg/xdb/storage/mysql"
)

// MySQLConfigurator MySQL 配置器
type MySQLConfigurator struct {
	dbName       string
	host         string
	port         int32
	username     string
	password     string
}

// NewMySQLConfigurator 创建 MySQL 配置器
func NewMySQLConfigurator(dbName, host string, port int32, username, password string) *MySQLConfigurator {
	return &MySQLConfigurator{
		dbName:   dbName,
		host:     host,
		port:     port,
		username: username,
		password: password,
	}
}

// InitializeDatabase 初始化数据库连接
func (c *MySQLConfigurator) InitializeDatabase() error {
	// MySQL 连接在 NewDao 时创建，这里可以添加数据库初始化逻辑
	// 例如：创建数据库、创建表等
	return nil
}

// RedoOptions 返回重做日志选项
func (c *MySQLConfigurator) RedoOptions() *xdb.RedoOptions {
	return &xdb.RedoOptions{
		Dir:          "./redo",
		Enabled:      false, // 测试时禁用重做日志
		SyncInterval: 100 * time.Millisecond,
	}
}

// DriverOptions 返回驱动选项
func (c *MySQLConfigurator) DriverOptions(driver string) interface{} {
	if driver == "mysql" {
		return &mysql.DriverOptions{}
	}
	return nil
}

// DaoOptions 返回 DAO 选项
func (c *MySQLConfigurator) DaoOptions(daoKey interface{}) interface{} {
	if daoKey == "mysql" {
		return &mysql.DaoOptions{
			DBName:       c.dbName,
			Host:         c.host,
			Port:         c.port,
			Username:     c.username,
			Password:     c.password,
			Charset:      "utf8mb4",
			MaxOpenConns: 10,
			QueryTimeout: 5 * time.Second,
		}
	}
	return nil
}

// TableOptions 返回表选项
func (c *MySQLConfigurator) TableOptions(driver string, table string) *xdb.TableOptions {
	if driver == "mysql" {
		return &xdb.TableOptions{
			DaoKey:      "mysql",
			Concurrence: 2,
			SaveTimeout: 5 * time.Second,
			SyncInterval: 100 * time.Millisecond,
		}
	}
	return nil
}

// DryRun 返回是否干运行
func (c *MySQLConfigurator) DryRun() bool {
	return false // 使用真实数据库
}

// SetupXdbWithMySQL 使用 MySQL 初始化 xdb
func SetupXdbWithMySQL(ctx context.Context, dbName, host string, port int32, username, password string) error {
	configurator := NewMySQLConfigurator(dbName, host, port, username, password)
	return xdb.Setup(ctx, configurator)
}


package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"

	"lucky/server/pkg/xdb"
)

// Dao MySQL 数据访问对象
type Dao struct {
	client   *sql.DB
	tableMgr *TableMgr
}

func newDao(ctx context.Context, opts *DaoOptions) (*Dao, error) {
	if opts.Charset == "" {
		opts.Charset = "utf8mb4"
	}
	if opts.Port == 0 {
		opts.Port = 3306
	}
	if opts.MaxOpenConns == 0 {
		opts.MaxOpenConns = 10
	}
	if opts.QueryTimeout == 0 {
		opts.QueryTimeout = 5 * time.Second
	}

	// 构建 DSN
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true&loc=Local",
		opts.Username,
		opts.Password,
		opts.Host,
		opts.Port,
		opts.DBName,
		opts.Charset,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open MySQL connection")
	}

	db.SetMaxOpenConns(int(opts.MaxOpenConns))
	db.SetMaxIdleConns(int(opts.MaxOpenConns) / 2)

	// 测试连接
	ctx, cancel := context.WithTimeout(ctx, opts.QueryTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, errors.Wrap(err, "failed to ping MySQL")
	}

	return &Dao{
		client:   db,
		tableMgr: newTableMgr(),
	}, nil
}

func (d *Dao) Table(src *xdb.Source) xdb.Table {
	return d.tableMgr.CreateTable(d, src)
}

// TableMgr 表管理器
type TableMgr struct {
	tables map[string]*Table
	mu     sync.Mutex
}

func newTableMgr() *TableMgr {
	return &TableMgr{
		tables: make(map[string]*Table),
	}
}

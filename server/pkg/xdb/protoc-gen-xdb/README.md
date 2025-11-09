# protoc-gen-xdb

protoc-gen-xdb 是一个 Protocol Buffers 代码生成插件，用于从 `.proto` 文件生成 xdb ORM 代码。

## 功能特性

- 自动生成 Record 结构体
- 自动生成 PK（主键）结构体
- 自动生成 Source 配置
- 自动生成 Commitment 对象
- 支持字段常量生成
- 支持 MutableRecord 接口实现

## 安装

### 方式一：从源码构建

```bash
cd lucky/server/pkg/xdb/protoc-gen-xdb
./build.sh
```

### 方式二：使用 go install

```bash
go install lucky/server/pkg/xdb/protoc-gen-xdb
```

## 使用方法

### 1. 定义 proto 文件

在 proto 文件中使用 xdb 扩展选项：

```protobuf
syntax = "proto3";

package example;

import "xdb/extension.proto";

// 定义数据表
message Player {
  option (xdb.table) = "player";
  option (xdb.driver) = DRIVER_MYSQL;
  
  int64 player_id = 1 [(xdb.pk) = true];
  string name = 2;
  int32 level = 3;
  int64 _version = 4 [(xdb.runtime) = true];
  int64 ctime = 5;
  int64 mtime = 6;
}
```

### 2. 生成代码

```bash
protoc \
  --proto_path=. \
  --proto_path=lucky/server/pkg/xdb \
  --xdb_out=. \
  example.proto
```

### 3. 使用生成的代码

```go
import (
    "context"
    "lucky/server/pkg/xdb"
    "example/pb"
)

func main() {
    ctx := context.Background()
    
    // 创建记录
    player, err := xdb.Create[pb.PlayerRecord](ctx, &pb.Player{
        PlayerId: 1001,
        Name:     "TestPlayer",
        Level:    1,
    })
    
    // 获取记录
    player, err := xdb.Get[pb.PlayerRecord](ctx, int64(1001))
    
    // 更新记录
    player.Name = "NewName"
    player.GetHeader().SetChanged(pb.FieldName)
    xdb.Save(ctx, player)
}
```

## Proto 选项说明

### Message 级别选项

- `xdb.table`: 表名（必需）
- `xdb.driver`: 驱动类型（DRIVER_MYSQL, DRIVER_MONGODB, DRIVER_NONE）
- `xdb.database`: 数据库名
- `xdb.layout`: 布局类型（LAYOUT_FLAT, LAYOUT_NESTED）
- `xdb.replica`: 是否为副本
- `xdb.lock_priority`: 锁优先级
- `xdb.lock_free_entire`: 整个消息无锁
- `xdb.cross_server`: 跨服务器

### Field 级别选项

- `xdb.pk`: 主键字段
- `xdb.gk`: 全局键字段
- `xdb.runtime`: 运行时字段（不持久化）
- `xdb.ticket`: 需要 ticket 生成 ID
- `xdb.readonly`: 只读字段
- `xdb.lock_free`: 无锁字段
- `xdb.critical`: 关键字段
- `xdb.shard_hint`: 分片提示
- `xdb.comment`: 字段注释

## 生成的文件

生成的文件名为 `{proto_file}_xdb.pb.go`，包含：

1. **字段常量**: 每个字段对应的 Field 常量
2. **PK 结构体**: 主键结构体，实现 `xdb.PK` 接口
3. **Record 结构体**: 记录结构体，实现 `xdb.Record` 和 `xdb.MutableRecord` 接口
4. **Commitment 结构体**: 提交对象，实现 `xdb.Commitment` 接口
5. **Source 配置**: 数据源配置对象
6. **初始化代码**: 自动注册 Source 的 init 函数

## 注意事项

1. 每个 message 必须至少有一个主键字段（使用 `xdb.pk = true`）
2. 建议包含 `_version` 字段用于版本控制
3. 建议包含 `ctime` 和 `mtime` 字段用于时间戳
4. 运行时字段（`xdb.runtime = true`）不会持久化到数据库

## 示例

完整示例请参考 `example.go` 文件。


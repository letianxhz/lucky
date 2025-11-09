# protoc-gen-xdb 使用指南

## 快速开始

### 1. 安装工具

```bash
cd lucky/server/pkg/xdb/protoc-gen-xdb
./build.sh
```

确保 `protoc-gen-xdb` 在 PATH 中，或者使用完整路径。

### 2. 定义 Proto 文件

创建 `player.proto`:

```protobuf
syntax = "proto3";

package game;

option go_package = "lucky/server/app/game/pb";

import "xdb/extension.proto";

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

### 3. 生成代码

```bash
protoc \
  --proto_path=. \
  --proto_path=lucky/server/pkg/xdb \
  --xdb_out=. \
  player.proto
```

这将生成 `player_xdb.pb.go` 文件。

### 4. 使用生成的代码

```go
package main

import (
    "context"
    "lucky/server/pkg/xdb"
    "lucky/server/app/game/pb"
)

func main() {
    ctx := context.Background()
    
    // 创建玩家
    player, err := xdb.Create[pb.PlayerRecord](ctx, &pb.Player{
        PlayerId: 1001,
        Name:     "TestPlayer",
        Level:    1,
    })
    if err != nil {
        panic(err)
    }
    
    // 获取玩家
    player, err = xdb.Get[pb.PlayerRecord](ctx, int64(1001))
    if err != nil {
        panic(err)
    }
    
    // 更新玩家
    player.Name = "NewName"
    player.GetHeader().SetChanged(pb.FieldName)
    xdb.Save(ctx, player)
    
    // 同步保存
    err = xdb.Sync(ctx, player)
    if err != nil {
        panic(err)
    }
}
```

## 选项说明

### Message 选项

| 选项 | 类型 | 说明 | 必需 |
|------|------|------|------|
| `xdb.table` | string | 表名 | 是 |
| `xdb.driver` | DriverType | 驱动类型 | 是 |
| `xdb.database` | string | 数据库名 | 否 |
| `xdb.layout` | LayoutType | 布局类型 | 否 |
| `xdb.replica` | bool | 是否为副本 | 否 |
| `xdb.lock_priority` | int32 | 锁优先级 | 否 |
| `xdb.lock_free_entire` | bool | 整个消息无锁 | 否 |
| `xdb.cross_server` | bool | 跨服务器 | 否 |

### Field 选项

| 选项 | 类型 | 说明 |
|------|------|------|
| `xdb.pk` | bool | 主键字段 |
| `xdb.gk` | bool | 全局键字段 |
| `xdb.runtime` | bool | 运行时字段（不持久化） |
| `xdb.ticket` | bool | 需要 ticket 生成 ID |
| `xdb.readonly` | bool | 只读字段 |
| `xdb.lock_free` | bool | 无锁字段 |
| `xdb.critical` | bool | 关键字段 |
| `xdb.shard_hint` | bool | 分片提示 |
| `xdb.comment` | string | 字段注释 |

## 最佳实践

### 1. 主键设计

- 每个表必须至少有一个主键字段
- 复合主键使用多个 `xdb.pk = true` 字段
- 主键字段建议使用 `_id` 或 `{entity}_id` 命名

### 2. 版本控制

- 建议包含 `_version` 字段用于乐观锁
- 使用 `xdb.runtime = true` 标记为运行时字段

### 3. 时间戳

- 建议包含 `ctime`（创建时间）和 `mtime`（修改时间）
- 这些字段通常不需要标记为运行时字段

### 4. 字段命名

- 使用驼峰命名法
- 运行时字段以 `_` 开头
- 主键字段建议以 `_id` 或 `Id` 结尾

## 常见问题

### Q: 如何定义复合主键？

A: 在多个字段上使用 `xdb.pk = true`:

```protobuf
message Item {
  int64 player_id = 1 [(xdb.pk) = true];
  int32 item_id = 2 [(xdb.pk) = true];
  int64 count = 3;
}
```

### Q: 如何标记不持久化的字段？

A: 使用 `xdb.runtime = true`:

```protobuf
int64 _version = 1 [(xdb.runtime) = true];
```

### Q: 如何设置锁优先级？

A: 在 message 级别设置:

```protobuf
message Player {
  option (xdb.lock_priority) = 1;
  // ...
}
```

### Q: 生成的代码在哪里？

A: 生成的文件名为 `{proto_file}_xdb.pb.go`，与 proto 文件在同一目录。

## 与 zplus-go/orm 的差异

1. **使用标准 protobuf**: 不再依赖 gogo/protobuf
2. **简化接口**: 移除了部分高级特性，保留核心功能
3. **适配 xdb**: 生成的代码直接适配 xdb 模块

## 故障排除

### 问题：找不到 extension.proto

**解决**: 确保在 `--proto_path` 中包含 `lucky/server/pkg/xdb`:

```bash
protoc --proto_path=lucky/server/pkg/xdb ...
```

### 问题：生成的代码编译错误

**解决**: 
1. 确保已安装 xdb 模块
2. 检查 proto 文件语法
3. 确保所有必需的字段都已定义

### 问题：主键字段未识别

**解决**: 确保字段使用了 `xdb.pk = true` 选项，或者字段名包含 `Id` 或 `ID`。


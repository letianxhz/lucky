# 数据库定义目录

这个目录用于存放定义数据库结构的 `.proto` 文件。

## 目录结构

```
db/
├── proto/          # 数据库定义的 proto 文件
│   ├── player.proto
│   ├── item.proto
│   └── ...
└── README.md
```

## 使用说明

1. **定义数据库表结构**：在 `proto/` 目录下创建 `.proto` 文件，定义数据库表结构
2. **生成代码**：运行 `../gen_all.sh` 脚本生成所有代码
   - PB 代码：生成到 `gen/pb/`
   - MSG 代码：生成到 `gen/msg/`
   - DB 脚本：生成到 `gen/db/`
   - Config 代码：生成到 `gen/config/`

## 示例

### player.proto

```protobuf
syntax = "proto3";

package db;

option go_package = "lucky/server/gen/pb";

import "extension.proto";

// Player 玩家数据表
message Player {
  option (xdb.table) = "player";
  option (xdb.driver) = DRIVER_MYSQL;
  
  int64 player_id = 1 [(xdb.pk) = true, (xdb.comment) = "玩家ID"];
  string name = 2 [(xdb.comment) = "玩家名称"];
  int32 level = 3 [(xdb.comment) = "玩家等级"];
  int64 exp = 4 [(xdb.comment) = "经验值"];
  int64 ctime = 5 [(xdb.comment) = "创建时间"];
  int64 mtime = 6 [(xdb.comment) = "修改时间"];
}
```

## 生成脚本

- `gen_all.sh` - 生成所有代码（pb、msg、db、config）
- `gen_msg.sh` - 仅生成 msg 代码
- `gen_db.sh` - 仅生成 db 脚本




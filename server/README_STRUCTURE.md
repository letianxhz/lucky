# Demo Cluster 目录结构说明

本目录结构参考 `cherry/game_server` 的设计，采用分层架构。

## 设计原则

**重要**：`server/` 是一个**项目目录**，类似整个 `server/` 项目。

- **服务工程**：放在 `server/app/` 下（类似 `game_server/app/`）
  - `app/gate/` - 网关服务工程
  - `app/game/` - 游戏服务工程
  - `app/center/` - 中心服务工程
  - `app/master/` - 主节点服务工程
  - `app/web/` - Web服务工程
- **服务入口**：放在 `server/cmd/` 下（类似 `game_server/cmd/`）
- **通用模块**：放在 `server/internal/` 下（共享的内部业务逻辑）
- **配置文件**：放在 `server/profiles/` 下

## 目录结构

```
server/
├── app/                  # 服务工程目录
│   ├── gate/            # 网关服务工程
│   │   ├── actor/       # Actor层（处理用户连接、登录等）
│   │   ├── config/      # 配置层
│   │   ├── controller/  # 控制器层
│   │   └── service/     # 服务层（路由、中间件）
│   ├── game/            # 游戏服务工程
│   │   ├── actor/       # Actor层（玩家管理、游戏逻辑）
│   │   ├── config/      # 配置层
│   │   ├── db/          # 数据库层
│   │   └── module/      # 模块层（在线管理等）
│   ├── center/          # 中心服务工程
│   │   ├── actor/       # Actor层（账号管理、运维等）
│   │   ├── config/      # 配置层
│   │   └── db/          # 数据库层
│   ├── master/          # 主节点服务工程
│   └── web/             # Web服务工程
│       ├── controller/  # 控制器层
│       ├── sdk/         # SDK层
│       ├── static/      # 静态文件
│       └── view/        # 视图文件
├── cmd/                 # 服务入口
│   ├── gate/
│   │   └── main.go
│   ├── game/
│   │   └── main.go
│   ├── center/
│   │   └── main.go
│   ├── master/
│   │   └── main.go
│   └── web/
│       └── main.go
├── profiles/            # 配置文件
│   ├── server.json
│   └── ...
├── internal/           # 共享的内部业务逻辑
│   ├── code/           # 业务状态码
│   ├── component/      # 组件
│   ├── data/           # 数据配置
│   ├── event/          # 游戏事件
│   ├── pb/             # protobuf生成的协议结构
│   ├── protocol/       # protobuf结构定义
│   ├── rpc/            # 跨节点rpc函数封装
│   └── ...
└── nodes/              # 旧目录（保留用于参考，后续可删除）
```

## 分层说明

### 1. app/gate/ - 网关服务工程
- `actor/` - Actor层，处理用户连接和登录
- `config/` - 配置层，网关特定配置
- `service/router/` - 路由服务，消息路由逻辑
- `service/middleware/` - 中间件服务

### 2. app/game/ - 游戏服务工程
- `actor/` - Actor层，玩家管理和游戏逻辑
- `config/` - 配置层，游戏特定配置
- `db/` - 数据库层，数据访问
- `module/` - 模块层，在线管理等

### 3. app/center/ - 中心服务工程
- `actor/` - Actor层，账号管理、运维等
- `config/` - 配置层
- `db/` - 数据库层

### 4. cmd/ - 服务入口
每个服务的 main.go 入口，负责：
- 加载配置文件
- 初始化服务
- 注册组件和Actor
- 启动服务

### 5. internal/ - 共享模块
所有服务共享的内部业务逻辑，包括：
- 业务状态码
- 数据配置
- 游戏事件
- RPC封装
- 协议定义

## 与旧结构的对应关系

| 新结构 | 旧结构 | 说明 |
|--------|--------|------|
| `app/gate/` | `nodes/gate/` | 网关服务 |
| `app/game/` | `nodes/game/` | 游戏服务 |
| `app/center/` | `nodes/center/` | 中心服务 |
| `app/master/` | `nodes/master/` | 主节点服务 |
| `app/web/` | `nodes/web/` | Web服务 |
| `cmd/{service}/main.go` | `nodes/main.go` | 服务入口 |
| `profiles/` | `../config/` | 配置文件 |
| `internal/` | `internal/` | 共享模块（保持不变） |

## 运行方式

### 旧方式（已废弃）
```bash
go run nodes/main.go gate --path=../../config/server.json --node=gc-gate-1
```

### 新方式
```bash
# Gate服务
cd cmd/gate && go run main.go

# Game服务
cd cmd/game && NODE_ID=10001 go run main.go

# Center服务
cd cmd/center && go run main.go

# Master服务
cd cmd/master && go run main.go

# Web服务
cd cmd/web && go run main.go
```

## 注意事项

1. **服务特定代码**：放在对应的 `app/{service}/` 下
2. **共享模块**：使用 `internal/` 下的模块
3. **配置管理**：每个服务有自己的 `config/` 目录，全局配置在 `profiles/` 下
4. **导入路径**：已更新为新的路径结构


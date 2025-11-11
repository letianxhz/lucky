# 按 Actor 划分 Entity 和 Module 层重构方案

## 重构目标

将 entity 层和 module 层按 actor 划分，使结构更清晰、更易维护。

## 当前结构

```
app/game/
├── actor/
│   ├── player/      # 玩家 Actor
│   ├── room/        # 房间 Actor
│   └── alliance/    # 联盟 Actor
├── module/
│   ├── login/       # 登录模块（按业务领域）
│   ├── item/        # 道具模块（按业务领域）
│   ├── room/        # 房间模块（按业务领域）
│   ├── alliance/    # 联盟模块（按业务领域）
│   ├── equipment/   # 装备模块（按业务领域）
│   ├── online/      # 在线状态模块（按业务领域）
│   └── player/      # 玩家模块（按业务领域）
└── entity/          # 空目录
```

## 重构后的结构

```
app/game/
├── actor/
│   ├── player/      # 玩家 Actor
│   ├── room/        # 房间 Actor
│   └── alliance/    # 联盟 Actor
│
├── module/
│   ├── player/      # 玩家相关的所有模块
│   │   ├── login/   # 登录模块
│   │   │   ├── login.go
│   │   │   ├── login_impl.go
│   │   │   └── handler.go
│   │   ├── item/    # 道具模块
│   │   │   ├── item.go
│   │   │   ├── item_impl.go
│   │   │   └── handler.go
│   │   ├── equipment/ # 装备模块
│   │   │   ├── equipment.go
│   │   │   └── equipment_impl.go
│   │   └── handler.go # 玩家相关的 handler（如果需要）
│   │
│   ├── room/        # 房间相关的所有模块
│   │   ├── room/     # 房间模块
│   │   │   ├── room.go
│   │   │   ├── room_impl.go
│   │   │   └── handler.go
│   │   └── handler.go # 房间相关的 handler（如果需要）
│   │
│   ├── alliance/    # 联盟相关的所有模块
│   │   ├── alliance/ # 联盟模块
│   │   │   ├── alliance.go
│   │   │   └── handler.go
│   │   └── handler.go # 联盟相关的 handler（如果需要）
│   │
│   └── shared/      # 共享模块（跨 actor）
│       ├── online/  # 在线状态模块
│       │   └── online.go
│       └── handler/ # Handler 基础设施
│           ├── actor_type.go
│           ├── receiver_v3.go
│           └── ...
│
└── entity/
    ├── player/      # 玩家相关的实体
    │   ├── player.go
    │   └── ...
    ├── room/        # 房间相关的实体
    │   ├── room.go
    │   └── ...
    ├── alliance/    # 联盟相关的实体
    │   ├── alliance.go
    │   └── ...
    └── shared/      # 共享实体
        └── ...
```

## 重构步骤

### 步骤 1: 创建新的目录结构

1. 创建 `module/player/` 目录
2. 创建 `module/room/` 目录
3. 创建 `module/alliance/` 目录
4. 创建 `module/shared/` 目录
5. 创建 `entity/player/` 目录
6. 创建 `entity/room/` 目录
7. 创建 `entity/alliance/` 目录
8. 创建 `entity/shared/` 目录

### 步骤 2: 移动模块文件

1. **Player 相关模块**：
   - `module/login/` → `module/player/login/`
   - `module/item/` → `module/player/item/`
   - `module/equipment/` → `module/player/equipment/`
   - `module/player/handler.go` → `module/player/handler.go`（保持不变）

2. **Room 相关模块**：
   - `module/room/` → `module/room/room/`（保持 room 子目录）

3. **Alliance 相关模块**：
   - `module/alliance/` → `module/alliance/alliance/`（保持 alliance 子目录）

4. **共享模块**：
   - `module/online/` → `module/shared/online/`
   - `module/handler/` → `module/shared/handler/`

### 步骤 3: 更新导入路径

更新所有文件中的导入路径：
- `lucky/server/app/game/module/login` → `lucky/server/app/game/module/player/login`
- `lucky/server/app/game/module/item` → `lucky/server/app/game/module/player/item`
- `lucky/server/app/game/module/equipment` → `lucky/server/app/game/module/player/equipment`
- `lucky/server/app/game/module/room` → `lucky/server/app/game/module/room/room`
- `lucky/server/app/game/module/alliance` → `lucky/server/app/game/module/alliance/alliance`
- `lucky/server/app/game/module/online` → `lucky/server/app/game/module/shared/online`
- `lucky/server/app/game/module/handler` → `lucky/server/app/game/module/shared/handler`

### 步骤 4: 更新 manager.go

```go
// module/manager.go
package module

import (
    // Player 相关模块
    _ "lucky/server/app/game/module/player/login"
    _ "lucky/server/app/game/module/player/item"
    _ "lucky/server/app/game/module/player/equipment"
    _ "lucky/server/app/game/module/player" // handler.go
    
    // Room 相关模块
    _ "lucky/server/app/game/module/room/room"
    _ "lucky/server/app/game/module/room" // handler.go
    
    // Alliance 相关模块
    _ "lucky/server/app/game/module/alliance/alliance"
    _ "lucky/server/app/game/module/alliance" // handler.go
    
    // 共享模块
    _ "lucky/server/app/game/module/shared/online"
)
```

## 优势

1. **结构更清晰**：每个 actor 相关的代码都在一个地方
2. **易于查找**：要找玩家相关的代码，直接看 `module/player/`
3. **边界明确**：每个 actor 的模块边界清晰
4. **易于扩展**：新增 actor 时，只需创建对应的目录
5. **减少冲突**：不同 actor 的模块不会混在一起

## 注意事项

1. **共享模块**：跨 actor 的模块放在 `shared/` 目录
2. **导入路径**：需要更新所有相关的导入路径
3. **测试文件**：测试文件也需要移动和更新
4. **文档**：需要更新相关文档

## 实施计划

1. ✅ 创建重构方案文档（本文档）
2. ⏳ 创建新的目录结构
3. ⏳ 移动模块文件
4. ⏳ 更新导入路径
5. ⏳ 更新 manager.go
6. ⏳ 运行测试确保一切正常
7. ⏳ 更新文档




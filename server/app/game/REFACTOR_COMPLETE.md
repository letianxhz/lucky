# 按 Actor 划分 Entity 和 Module 层重构完成

## 重构内容

已成功将 entity 层和 module 层按 actor 划分，使结构更清晰、更易维护。

## 新的目录结构

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
│   │   ├── item/    # 道具模块
│   │   ├── equipment/ # 装备模块
│   │   └── handler.go # 玩家相关的 handler
│   │
│   ├── room/        # 房间相关的所有模块
│   │   └── room/     # 房间模块
│   │
│   ├── alliance/    # 联盟相关的所有模块
│   │   └── alliance/ # 联盟模块
│   │
│   └── shared/      # 共享模块（跨 actor）
│       ├── online/  # 在线状态模块
│       └── handler/ # Handler 基础设施
│
└── entity/
    ├── player/      # 玩家相关的实体（待创建）
    ├── room/        # 房间相关的实体（待创建）
    ├── alliance/    # 联盟相关的实体（待创建）
    └── shared/      # 共享实体（待创建）
```

## 已完成的更改

### 1. 目录结构重组

- ✅ `module/login/` → `module/player/login/`
- ✅ `module/item/` → `module/player/item/`
- ✅ `module/equipment/` → `module/player/equipment/`
- ✅ `module/room/` → `module/room/room/`
- ✅ `module/alliance/` → `module/alliance/alliance/`
- ✅ `module/online/` → `module/shared/online/`
- ✅ `module/handler/` → `module/shared/handler/`

### 2. 导入路径更新

所有文件中的导入路径已更新：
- ✅ `lucky/server/app/game/module/login` → `lucky/server/app/game/module/player/login`
- ✅ `lucky/server/app/game/module/item` → `lucky/server/app/game/module/player/item`
- ✅ `lucky/server/app/game/module/equipment` → `lucky/server/app/game/module/player/equipment`
- ✅ `lucky/server/app/game/module/room` → `lucky/server/app/game/module/room/room`
- ✅ `lucky/server/app/game/module/alliance` → `lucky/server/app/game/module/alliance/alliance`
- ✅ `lucky/server/app/game/module/online` → `lucky/server/app/game/module/shared/online`
- ✅ `lucky/server/app/game/module/handler` → `lucky/server/app/game/module/shared/handler`

### 3. manager.go 更新

已更新 `module/manager.go`，使用新的导入路径。

## 优势

1. **结构更清晰**：每个 actor 相关的代码都在一个地方
2. **易于查找**：要找玩家相关的代码，直接看 `module/player/`
3. **边界明确**：每个 actor 的模块边界清晰
4. **易于扩展**：新增 actor 时，只需创建对应的目录
5. **减少冲突**：不同 actor 的模块不会混在一起

## 下一步

1. **创建 Entity 层**：按 actor 划分 entity
2. **更新文档**：更新相关架构文档
3. **测试验证**：确保所有功能正常

## 注意事项

- 共享模块（如 `online`、`handler`）放在 `shared/` 目录
- 每个 actor 的模块可以有自己的子模块（如 `room/room/`）
- 导入路径已全部更新，编译应该通过


# Handler 注册优化说明

## 优化目标

优化 `RegisterAllToActorByType`，使每种 actor 类型（如 player）的 handler 列表只遍历一次，后续同类型的 actor 实例直接使用缓存的列表。

## 优化前的问题

**问题**：
- 每个 actor 实例在 `OnInit()` 时都会调用 `RegisterAllToActorByType`
- 每次调用都会遍历所有 handler，找出属于该 actor 类型的 handler
- 如果有 100 个 player actor 实例，就会遍历 100 次 handler 列表

**性能影响**：
- 当 handler 数量较多时，每次遍历都有 O(n) 的时间复杂度
- 大量 actor 实例创建时，会有明显的性能开销

## 优化方案

### 1. Handler 列表缓存

为每种 actor 类型缓存其 handler 列表：

```go
// actorTypeHandlers 缓存每种 actor 类型的 handler 列表
// key: actorType, value: handler 列表
actorTypeHandlers = make(map[ActorType][]*msgHandlerInfoV3)
```

### 2. 首次构建缓存

第一次为某种 actor 类型注册 handler 时：
1. 遍历所有 handler，找出属于该类型的 handler
2. 构建 handler 列表
3. 缓存该列表

### 3. 后续直接使用缓存

后续同类型的 actor 实例创建时：
1. 直接从缓存获取 handler 列表
2. 遍历缓存的列表，注册到新的 actor 实例
3. 无需再次遍历所有 handler

## 优化效果

### 性能提升

**优化前**：
- 100 个 player actor 实例 × 遍历所有 handler = 100 次 O(n) 遍历

**优化后**：
- 第一次：遍历所有 handler，构建缓存 = 1 次 O(n) 遍历
- 后续 99 次：直接使用缓存 = 99 次 O(1) 查找 + O(m) 遍历（m 是该类型的 handler 数量，通常远小于 n）

**性能提升**：
- 时间复杂度：从 O(n × m) 降低到 O(n + m × k)，其中 n 是总 handler 数，m 是某类型的 handler 数，k 是该类型的 actor 实例数
- 实际效果：当 handler 数量较多时，性能提升明显

### 内存开销

- 每个 actor 类型缓存一个 handler 列表（指针数组）
- 内存开销很小，可以忽略不计

## 实现细节

### V3 版本（泛型，无反射）

```go
// getOrBuildActorTypeHandlers 获取或构建指定 actor 类型的 handler 列表
func getOrBuildActorTypeHandlers(actorType ActorType) []*msgHandlerInfoV3 {
    actorTypeHandlersLock.Lock()
    defer actorTypeHandlersLock.Unlock()

    // 如果已缓存，直接返回
    if cached, exists := actorTypeHandlers[actorType]; exists {
        return cached
    }

    // 第一次：遍历所有 handler，找出属于该 actor 类型的 handler
    msgHandlersV3Lock.Lock()
    handlerList := make([]*msgHandlerInfoV3, 0)
    for _, info := range msgHandlersV3 {
        if info.actorType == actorType {
            handlerList = append(handlerList, info)
        }
    }
    msgHandlersV3Lock.Unlock()

    // 缓存该 actor 类型的 handler 列表
    actorTypeHandlers[actorType] = handlerList

    return handlerList
}
```

### 旧版本（兼容）

同样的优化逻辑，使用 `actorTypeHandlersOld` 缓存。

## 使用示例

### 正常使用（无需修改代码）

```go
// actor_player.go
func (p *actorPlayer) OnInit() {
    // 第一次调用：构建 player 类型的 handler 缓存
    // 后续调用：直接使用缓存
    handler.RegisterAllToActorByType(handler.ActorTypePlayer, &p.ActorBase)
}
```

### 日志输出

**第一次注册**（构建缓存）：
```
[Handler] Built handler cache for actor type=player, count=5
[Handler] Registered 5 handler(s) to actor 10001.player.1001 (type=player, generic)
```

**后续注册**（使用缓存）：
```
[Handler] Registered 5 handler(s) to actor 10001.player.1002 (type=player, generic)
```

## 注意事项

1. **Handler 在 init() 中注册**：所有 handler 都在程序启动时的 `init()` 函数中注册，不会在运行时动态注册，所以缓存是安全的。

2. **线程安全**：使用 `actorTypeHandlersLock` 保护缓存，确保并发安全。

3. **内存占用**：缓存的是指针数组，内存占用很小。

4. **如果将来支持动态注册**：如果将来需要在运行时动态注册 handler，需要清除对应的缓存。

## 总结

通过缓存每种 actor 类型的 handler 列表，我们实现了：
- ✅ 每种 actor 类型只遍历一次 handler 列表
- ✅ 后续同类型的 actor 实例直接使用缓存
- ✅ 显著减少重复遍历的开销
- ✅ 保持代码简洁，无需修改调用方代码

这个优化在 handler 数量较多或 actor 实例较多时，性能提升明显。


# Actor 自动创建与 Handler 注册机制

## 问题说明

当消息发送到不存在的 actor 时，cherry 框架会自动创建该 actor。这引发了一个问题：handler 是在 actor 的 `OnInit()` 中注册的，如果消息在 actor 创建和 handler 注册之间到达，是否会丢失？

## 机制保证

### 1. Actor 创建和初始化流程

根据 cherry 框架的实现（`cherry/net/actor/actor.go`）：

```go
func (p *Actor) run() {
    p.onInit()  // 先调用 onInit()，注册 handler
    defer p.onStop()

    for {
        if p.loop() {  // 然后才开始处理消息队列
            break
        }
    }
}
```

**关键点**：
- `onInit()` 在消息处理循环之前调用
- 消息会被放入 `localMail` 队列，等待 `onInit()` 完成后处理
- 这确保了 handler 在消息处理之前就已经注册完成

### 2. 消息路由流程

当消息发送到不存在的 actor 时：

1. **消息到达** → 触发 `OnFindChild` → 创建 actor → 启动 `run()` goroutine
2. **`run()` 执行** → 先调用 `onInit()` → 注册 handler → 然后才开始处理消息队列
3. **消息处理** → 从队列中取出消息 → 调用已注册的 handler

### 3. 时序保证

```
时间线：
T1: 消息发送到不存在的 actor
T2: 触发 OnFindChild，创建 actor
T3: 启动 run() goroutine
T4: 调用 onInit()，注册 handler  ← handler 已注册
T5: 开始处理消息队列
T6: 处理消息，调用 handler        ← handler 可用
```

**结论**：消息不会丢失，handler 会在消息处理之前完成注册。

## 当前实现

### Handler 注册位置

所有 handler 在 actor 的 `OnInit()` 中注册：

```go
// actor_player.go
func (p *actorPlayer) OnInit() {
    // 注册所有模块的消息处理器
    handler.RegisterAllToActorByType(handler.ActorTypePlayer, &p.ActorBase)
}
```

### 自动创建机制

当消息发送到不存在的 actor 时，框架会自动创建：

```go
// actor_players.go
func (p *ActorPlayers) OnFindChild(msg *cfacade.Message) (cfacade.IActor, bool) {
    childID := msg.TargetPath().ChildID
    childActor, err := p.Child().Create(childID, &actorPlayer{
        isOnline: false,
    })
    // ...
    return childActor, true
}
```

## 优化建议

虽然框架已经保证了安全性，但为了更明确和可维护，建议：

1. **保持当前实现**：handler 在 `OnInit()` 中注册，这是最清晰的方式
2. **添加日志**：在 `OnInit()` 中添加日志，确认 handler 注册完成
3. **监控机制**：可以添加监控，确保所有 actor 实例都正确注册了 handler

## 注意事项

1. **不要在其他地方注册 handler**：handler 应该只在 `OnInit()` 中注册，确保一致性
2. **避免在 `OnFindChild` 中注册 handler**：这会导致重复注册，且不符合框架设计
3. **确保 `OnInit()` 快速完成**：虽然消息会等待，但 `OnInit()` 应该尽快完成，避免消息堆积

## 总结

当前的实现是安全的：
- ✅ Handler 在消息处理之前注册
- ✅ 消息会被缓冲在队列中，等待 handler 注册完成
- ✅ 框架保证了正确的执行顺序

无需额外的优化，当前机制已经能够正确处理 actor 自动创建和 handler 注册的场景。


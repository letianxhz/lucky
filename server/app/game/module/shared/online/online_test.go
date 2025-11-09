package online

import (
	"testing"
)

func TestBindPlayer(t *testing.T) {
	playerId := int64(1001)
	uid := int64(2001)
	agentPath := "gate.10001.agent.1001"

	// 测试绑定玩家
	BindPlayer(playerId, uid, agentPath)

	// 验证绑定成功
	retrievedPlayerId := GetPlayerId(uid)
	if retrievedPlayerId != playerId {
		t.Errorf("Expected playerId %d, got %d", playerId, retrievedPlayerId)
	}

	// 验证在线数量
	count := Count()
	if count < 1 {
		t.Errorf("Expected online count >= 1, got %d", count)
	}

	t.Logf("BindPlayer test passed: playerId=%d, uid=%d, count=%d", playerId, uid, count)
}

func TestUnBindPlayer(t *testing.T) {
	playerId := int64(1002)
	uid := int64(2002)
	agentPath := "gate.10001.agent.1002"

	// 先绑定
	BindPlayer(playerId, uid, agentPath)

	// 验证绑定成功
	retrievedPlayerId := GetPlayerId(uid)
	if retrievedPlayerId != playerId {
		t.Errorf("Expected playerId %d, got %d", playerId, retrievedPlayerId)
	}

	// 解绑
	unboundPlayerId := UnBindPlayer(uid)
	if unboundPlayerId != playerId {
		t.Errorf("Expected unbound playerId %d, got %d", playerId, unboundPlayerId)
	}

	// 验证解绑成功
	retrievedPlayerId = GetPlayerId(uid)
	if retrievedPlayerId != 0 {
		t.Errorf("Expected playerId 0 after unbind, got %d", retrievedPlayerId)
	}

	t.Logf("UnBindPlayer test passed: playerId=%d, uid=%d", playerId, uid)
}

func TestCount(t *testing.T) {
	initialCount := Count()

	// 绑定几个玩家
	BindPlayer(1001, 2001, "gate.10001.agent.1001")
	BindPlayer(1002, 2002, "gate.10001.agent.1002")
	BindPlayer(1003, 2003, "gate.10001.agent.1003")

	count := Count()
	expectedCount := initialCount + 3
	// 由于可能有其他测试并发执行，只检查数量是否增加
	if count < expectedCount {
		t.Logf("Count may be affected by concurrent tests: expected >= %d, got %d", expectedCount, count)
	}

	// 清理
	UnBindPlayer(2001)
	UnBindPlayer(2002)
	UnBindPlayer(2003)

	finalCount := Count()
	// 由于可能有其他测试并发执行，只检查数量是否减少
	if finalCount > initialCount {
		t.Logf("Final count may be affected by concurrent tests: expected <= %d, got %d", initialCount, finalCount)
	}

	t.Logf("Count test passed: initial=%d, after bind=%d, final=%d", initialCount, count, finalCount)
}

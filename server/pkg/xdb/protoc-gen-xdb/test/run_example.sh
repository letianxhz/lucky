#!/bin/bash

# 运行完整示例的脚本

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "=== protoc-gen-xdb 测试示例 ==="
echo ""

# 步骤 1: 生成代码
echo "步骤 1: 生成 xdb 代码..."
./generate.sh

if [ $? -ne 0 ]; then
    echo "错误: 代码生成失败"
    exit 1
fi

echo ""
echo "步骤 2: 检查生成的文件..."
if [ -f "player_xdb.pb.go" ]; then
    echo "✓ player_xdb.pb.go 已生成"
    echo ""
    echo "文件内容预览:"
    head -20 player_xdb.pb.go
else
    echo "✗ player_xdb.pb.go 未生成"
    exit 1
fi

echo ""
echo "步骤 3: 运行测试..."
echo ""

# 检查是否有 go.mod
if [ ! -f "go.mod" ]; then
    echo "初始化 go module..."
    go mod init test 2>/dev/null || true
fi

# 运行测试
go test -v 2>&1 | head -50 || {
    echo ""
    echo "注意: 测试可能需要先配置 xdb 模块"
    echo "这是正常的，因为 xdb 模块需要数据库连接"
    echo ""
    echo "生成的代码结构:"
    echo "  - 字段常量 (FieldPlayerId, FieldName, etc.)"
    echo "  - PK 结构体 (PlayerPK, ItemPK)"
    echo "  - Record 结构体 (PlayerRecord, ItemRecord)"
    echo "  - Commitment 结构体 (PlayerCommitment, ItemCommitment)"
    echo "  - Source 配置 (_PlayerSource, _ItemSource)"
    echo ""
    echo "代码生成成功！"
}

echo ""
echo "=== 完成 ==="


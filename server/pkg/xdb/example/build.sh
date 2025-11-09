#!/bin/bash

# 构建 xdb 示例程序

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

PB_DIR="pb"

echo "=== 构建 xdb 示例程序 ==="
echo ""

# 检查生成的文件是否存在
if [ ! -f "$PB_DIR/player.pb.go" ] || [ ! -f "$PB_DIR/player_xdb.pb.go" ]; then
    echo "错误: 缺少生成的代码文件"
    echo "请先运行 ./generate.sh 生成代码"
    exit 1
fi

# 构建
echo "构建中..."
# 清理可能存在的旧文件
rm -f player.pb.go player_xdb.pb.go

# 由于 Go 要求所有文件在同一目录，临时复制 pb 文件到当前目录
cp "$PB_DIR"/*.pb.go . 2>/dev/null || true

# 构建
go build -o xdb_example main.go config.go player.pb.go player_xdb.pb.go

# 清理临时文件（构建完成后删除，避免重复注册问题）
rm -f player.pb.go player_xdb.pb.go

if [ $? -eq 0 ]; then
    echo "✓ 构建成功: ./xdb_example"
    echo ""
    echo "运行示例:"
    echo "  ./xdb_example"
else
    echo "✗ 构建失败"
    exit 1
fi


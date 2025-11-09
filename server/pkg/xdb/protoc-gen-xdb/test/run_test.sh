#!/bin/bash
# 运行 xdb 配置和测试

set -e

echo "=== 运行 xdb 配置和测试 ==="
echo ""

# 确保生成的代码存在
if [ ! -f "player_xdb.pb.go" ]; then
    echo "错误: player_xdb.pb.go 不存在"
    echo "请先运行 ./generate.sh 生成代码"
    exit 1
fi

echo "1. 运行配置测试..."
go run simple_main.go config.go player_xdb.pb.go

echo ""
echo "=== 测试完成 ==="

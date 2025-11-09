#!/bin/bash

# MongoDB 测试脚本

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "=== MongoDB 测试准备 ==="
echo ""

# 1. 检查 MongoDB 驱动是否已编译
echo "1. 检查 MongoDB 驱动..."
if ! go build -o /dev/null ../storage/mongo/*.go 2>/dev/null; then
    echo "   编译 MongoDB 驱动..."
    cd ../storage/mongo
    go build .
    cd "$SCRIPT_DIR"
fi
echo "   ✓ MongoDB 驱动就绪"
echo ""

PB_DIR="pb"

# 2. 生成代码（如果还没有）
if [ ! -f "$PB_DIR/player_xdb.pb.go" ]; then
    echo "2. 生成代码..."
    ./generate.sh
    echo ""
fi

# 3. 修改生成的代码以使用 MongoDB
echo "3. 配置为使用 MongoDB..."
./fix_driver.sh

echo ""

# 4. 运行测试
echo "4. 运行 MongoDB 测试..."
echo "   注意: 确保 MongoDB 服务正在运行"
echo "   可以通过环境变量设置: export MONGO_URI=mongodb://localhost:27017"
echo ""

MONGO_URI=${MONGO_URI:-"mongodb://localhost:27017"}
export MONGO_URI

go run mongo_main.go mongo_config.go "$PB_DIR/player.pb.go" "$PB_DIR/player_xdb.pb.go"


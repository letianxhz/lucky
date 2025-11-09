#!/bin/bash

# MySQL 测试脚本

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "=== MySQL 测试准备 ==="
echo ""

PB_DIR="pb"

# 1. 检查 MySQL 驱动是否已编译
echo "1. 检查 MySQL 驱动..."
if ! go build -o /dev/null ../storage/mysql/*.go 2>/dev/null; then
    echo "   编译 MySQL 驱动..."
    cd ../storage/mysql
    go build .
    cd "$SCRIPT_DIR"
fi
echo "   ✓ MySQL 驱动就绪"
echo ""

# 2. 生成代码（如果还没有）
if [ ! -f "$PB_DIR/player_xdb.pb.go" ]; then
    echo "2. 生成代码..."
    ./generate.sh
    echo ""
fi

# 3. 修改生成的代码以使用 MySQL
echo "3. 配置为使用 MySQL..."
./fix_driver.sh mysql

echo ""

# 4. 运行测试
echo "4. 运行 MySQL 测试..."
echo "   注意: 确保 MySQL 服务正在运行"
echo "   可以通过环境变量设置:"
echo "     export MYSQL_DB=test"
echo "     export MYSQL_HOST=localhost"
echo "     export MYSQL_PORT=3306"
echo "     export MYSQL_USER=root"
echo "     export MYSQL_PASSWORD=your_password"
echo ""

MYSQL_DB=${MYSQL_DB:-"test"}
MYSQL_HOST=${MYSQL_HOST:-"localhost"}
MYSQL_PORT=${MYSQL_PORT:-"3306"}
MYSQL_USER=${MYSQL_USER:-"root"}
MYSQL_PASSWORD=${MYSQL_PASSWORD:-""}

export MYSQL_DB MYSQL_HOST MYSQL_PORT MYSQL_USER MYSQL_PASSWORD

go run mysql_main.go mysql_config.go "$PB_DIR/player.pb.go" "$PB_DIR/player_xdb.pb.go"


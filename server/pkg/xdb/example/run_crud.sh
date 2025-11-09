#!/bin/bash

# 运行 CRUD 测试

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "=== 运行 CRUD 测试 ==="
echo ""

# 检查是否已构建
if [ ! -f "./crud_test" ]; then
    echo "构建测试程序..."
    ./build_crud.sh
fi

# 设置 MySQL 环境变量（如果未设置）
export MYSQL_DB=${MYSQL_DB:-"test"}
export MYSQL_HOST=${MYSQL_HOST:-"localhost"}
export MYSQL_PORT=${MYSQL_PORT:-"3306"}
export MYSQL_USER=${MYSQL_USER:-"root"}
export MYSQL_PASSWORD=${MYSQL_PASSWORD:-""}

echo "MySQL 配置:"
echo "  DB: $MYSQL_DB"
echo "  Host: $MYSQL_HOST"
echo "  Port: $MYSQL_PORT"
echo "  User: $MYSQL_USER"
echo ""

echo "注意: 确保 MySQL 服务正在运行，并且数据库 '$MYSQL_DB' 已创建"
echo "      可以使用以下命令创建数据库和表:"
echo "      mysql -u$MYSQL_USER -p -e 'CREATE DATABASE IF NOT EXISTS $MYSQL_DB;'"
echo "      mysql -u$MYSQL_USER -p $MYSQL_DB < pb/player.sql"
echo ""

# 运行测试
./crud_test

echo ""
echo "=== 测试完成 ==="


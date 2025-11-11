#!/bin/bash

# 数据库初始化脚本

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

DB_NAME=${MYSQL_DB:-"test"}
MYSQL_USER=${MYSQL_USER:-"root"}
MYSQL_PASSWORD=${MYSQL_PASSWORD:-""}

echo "=== 初始化数据库 ==="
echo "数据库名: $DB_NAME"
echo "用户: $MYSQL_USER"
echo ""

# 构建 MySQL 命令
MYSQL_CMD="mysql -u$MYSQL_USER"
if [ -n "$MYSQL_PASSWORD" ]; then
    MYSQL_CMD="$MYSQL_CMD -p$MYSQL_PASSWORD"
else
    MYSQL_CMD="$MYSQL_CMD"
fi

# 1. 创建数据库
echo "1. 创建数据库 '$DB_NAME'..."
$MYSQL_CMD -e "CREATE DATABASE IF NOT EXISTS $DB_NAME;" 2>&1 || {
    echo "   ⚠ 自动创建失败，请手动执行:"
    echo "   mysql -u$MYSQL_USER -p -e 'CREATE DATABASE IF NOT EXISTS $DB_NAME;'"
    exit 1
}
echo "   ✓ 数据库创建成功"
echo ""

# 2. 创建表
if [ -f "pb/player.sql" ]; then
    echo "2. 创建表 'player'..."
    $MYSQL_CMD $DB_NAME < pb/player.sql 2>&1 || {
        echo "   ⚠ 自动创建表失败，请手动执行:"
        echo "   mysql -u$MYSQL_USER -p $DB_NAME < pb/player.sql"
        exit 1
    }
    echo "   ✓ 表创建成功"
    echo ""
else
    echo "2. ⚠ pb/player.sql 不存在，跳过表创建"
    echo ""
fi

# 3. 验证
echo "3. 验证数据库和表..."
$MYSQL_CMD -e "USE $DB_NAME; SHOW TABLES;" 2>&1 | grep -q "player" && {
    echo "   ✓ 表 'player' 存在"
} || {
    echo "   ⚠ 表 'player' 不存在"
}
echo ""

echo "=== 数据库初始化完成 ==="
echo ""
echo "现在可以运行测试:"
echo "  ./crud_test"




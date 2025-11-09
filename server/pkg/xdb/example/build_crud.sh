#!/bin/bash

# 构建 CRUD 测试程序

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "=== 构建 CRUD 测试程序 ==="
echo ""

# 使用 go build 构建（会自动处理包依赖）
echo "构建程序..."
go build -o crud_test crud_main.go mysql_config.go player_model.go

echo ""
echo "✓ 构建完成: ./crud_test"
echo ""


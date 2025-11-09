#!/bin/bash

# 生成 db 脚本
# 从 .proto 文件生成数据库脚本（SQL）

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "=== 生成 db 脚本 ==="
echo ""

# 创建输出目录
SQL_DIR="$SCRIPT_DIR/gen/sql"
mkdir -p "$SQL_DIR"

# 查找所有 .proto 文件（在 db/proto 目录下）
PROTO_DIR="$SCRIPT_DIR/db/proto"
PROTO_FILES=$(find "$PROTO_DIR" -maxdepth 1 -name "*.proto" -type f 2>/dev/null)

# 如果没有找到，尝试从 example 目录查找（兼容旧代码）
if [ -z "$PROTO_FILES" ]; then
    PROTO_DIR="$SCRIPT_DIR/pkg/xdb/example"
    PROTO_FILES=$(find "$PROTO_DIR" -maxdepth 1 -name "*.proto" -type f 2>/dev/null)
fi

if [ -z "$PROTO_FILES" ]; then
    echo "⚠ 未找到 .proto 文件在 $PROTO_DIR"
    exit 0
fi

XDB_DIR="$SCRIPT_DIR/pkg/xdb"
PROTOC_GEN_XDB="$XDB_DIR/protoc-gen-xdb/protoc-gen-xdb"

# 检查 protoc-gen-xdb 是否存在
if [ ! -f "$PROTOC_GEN_XDB" ]; then
    echo "构建 protoc-gen-xdb..."
    cd "$XDB_DIR/protoc-gen-xdb"
    ./build.sh
    cd "$SCRIPT_DIR"
fi

# 检查 protoc 是否安装
if ! command -v protoc &> /dev/null; then
    echo "错误: protoc 未安装"
    exit 1
fi

# 临时目录用于生成 SQL
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

# 处理每个 .proto 文件
for proto_file in $PROTO_FILES; do
    proto_name=$(basename "$proto_file" .proto)
    
    echo "处理: $proto_file"
    
    # 生成 xdb 代码（会同时生成 SQL）
    protoc \
      --proto_path="$PROTO_DIR" \
      --proto_path="$XDB_DIR" \
      --plugin=protoc-gen-xdb="$PROTOC_GEN_XDB" \
      --xdb_out="$TEMP_DIR" \
      "$proto_file"
    
    # 查找生成的 SQL 文件
    if [ -f "$TEMP_DIR/${proto_name}.sql" ]; then
        mv "$TEMP_DIR/${proto_name}.sql" "$SQL_DIR/"
        echo "  ✓ 生成: $SQL_DIR/${proto_name}.sql"
    elif [ -f "$PROTO_DIR/${proto_name}.sql" ]; then
        cp "$PROTO_DIR/${proto_name}.sql" "$SQL_DIR/"
        echo "  ✓ 复制: $SQL_DIR/${proto_name}.sql"
    fi
done

# 清理临时文件
rm -rf "$TEMP_DIR"

echo ""
echo "✓ SQL 脚本生成完成"
echo "  输出目录: $SQL_DIR"
echo ""
echo "生成的 SQL 文件:"
ls -lh "$SQL_DIR"/*.sql 2>/dev/null | awk '{print "  " $9}' || echo "  无 SQL 文件"
echo ""


#!/bin/bash

# 生成 proto 和 xdb 代码的脚本

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

XDB_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
PROTOC_GEN_XDB="$XDB_DIR/protoc-gen-xdb/protoc-gen-xdb"
PB_DIR="pb"

echo "=== 生成代码 ==="
echo "XDB 目录: $XDB_DIR"
echo "输出目录: $PB_DIR"
echo ""

# 创建 pb 目录
mkdir -p "$PB_DIR"

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
    echo "请安装 Protocol Buffers 编译器:"
    echo "  brew install protobuf  # macOS"
    exit 1
fi

echo "1. 生成 proto Go 代码..."
protoc \
  --go_out="$PB_DIR" \
  --go_opt=paths=source_relative \
  --proto_path=. \
  --proto_path="$XDB_DIR" \
  player.proto

echo "   ✓ player.pb.go 生成成功"

echo ""
echo "2. 生成 xdb 代码..."
protoc \
  --proto_path=. \
  --proto_path="$XDB_DIR" \
  --plugin=protoc-gen-xdb="$PROTOC_GEN_XDB" \
  --xdb_out="$PB_DIR" \
  player.proto

echo "   ✓ player_xdb.pb.go 生成成功"

# 移动生成的文件到 pb 目录（如果生成到了子目录）
if [ -d "$PB_DIR/lucky" ]; then
    # 查找所有生成的 .pb.go 文件并移动到 pb 目录
    find "$PB_DIR/lucky" -name "*.pb.go" -exec mv {} "$PB_DIR/" \; 2>/dev/null || true
    rm -rf "$PB_DIR/lucky" 2>/dev/null || true
fi

# 修改生成的代码包名为 pb
for f in "$PB_DIR"/*.pb.go; do
    if [ -f "$f" ]; then
        # 替换 package example 或 package main 为 package pb
        sed -i '' -e 's/^package example$/package pb/' -e 's/^package main$/package pb/' "$f" 2>/dev/null || \
        sed -i -e 's/^package example$/package pb/' -e 's/^package main$/package pb/' "$f"
    fi
done

# 移动 SQL 文件到 pb 目录（如果生成到了当前目录）
if [ -f "player.sql" ]; then
    mv player.sql "$PB_DIR/" 2>/dev/null || true
fi
if [ -f "item.sql" ]; then
    mv item.sql "$PB_DIR/" 2>/dev/null || true
fi

# 清理可能生成的临时文件
find . -name "*.pb.go" -not -path "./$PB_DIR/*" -exec rm -f {} \; 2>/dev/null || true
rm -rf lucky 2>/dev/null || true

echo ""
echo "生成的文件:"
ls -lh "$PB_DIR"/*.pb.go 2>/dev/null || echo "  无 .pb.go 文件生成"
if [ -f "$PB_DIR/player.sql" ] || [ -f "$PB_DIR/item.sql" ]; then
    echo ""
    echo "SQL 文件:"
    ls -lh "$PB_DIR"/*.sql 2>/dev/null || true
fi

echo ""
echo "=== 完成 ==="


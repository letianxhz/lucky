#!/bin/bash

# 生成 xdb 代码的脚本

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# 获取项目根目录
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../../.." && pwd)"
XDB_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
PROTOC_GEN_XDB="$XDB_DIR/protoc-gen-xdb/protoc-gen-xdb"

echo "Project root: $PROJECT_ROOT"
echo "XDB dir: $XDB_DIR"

# 检查 protoc-gen-xdb 是否存在
if [ ! -f "$PROTOC_GEN_XDB" ]; then
    echo "Building protoc-gen-xdb..."
    cd "$XDB_DIR/protoc-gen-xdb"
    ./build.sh
    cd "$SCRIPT_DIR"
fi

# 检查 protoc 是否安装
if ! command -v protoc &> /dev/null; then
    echo "Error: protoc is not installed"
    echo "Please install Protocol Buffers compiler:"
    echo "  brew install protobuf  # macOS"
    echo "  apt-get install protobuf-compiler  # Ubuntu"
    exit 1
fi

echo "Generating xdb code..."

# 生成代码
protoc \
  --proto_path=. \
  --proto_path="$XDB_DIR" \
  --plugin=protoc-gen-xdb="$PROTOC_GEN_XDB" \
  --xdb_out=. \
  player.proto

echo "Generated files:"
ls -la *_xdb.pb.go 2>/dev/null || echo "No files generated"

echo "Done!"


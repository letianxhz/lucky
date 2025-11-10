#!/bin/bash

# 构建 protoc-gen-xdb 工具

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "Building protoc-gen-xdb..."

# 检查 Go 环境
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed"
    exit 1
fi

# 构建（包含所有 .go 文件）
go build -o protoc-gen-xdb .

echo "Build complete: protoc-gen-xdb"

# 如果 PATH 中包含当前目录，可以创建符号链接
if [[ ":$PATH:" == *":$SCRIPT_DIR:"* ]]; then
    echo "protoc-gen-xdb is ready to use"
else
    echo "To use protoc-gen-xdb, add it to your PATH or use full path:"
    echo "  $SCRIPT_DIR/protoc-gen-xdb"
fi


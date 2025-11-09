#!/bin/bash
# 修复生成的代码中的驱动名称

PB_DIR="pb"
DRIVER="${1:-mongo}"  # 默认使用 mongo，可以通过参数指定：./fix_driver.sh mysql

for f in "$PB_DIR"/*_xdb.pb.go; do
    if [ -f "$f" ]; then
        sed -i '' "s/DriverName: \"none\"/DriverName: \"$DRIVER\"/g" "$f" 2>/dev/null || \
        sed -i "s/DriverName: \"none\"/DriverName: \"$DRIVER\"/g" "$f"
        if [ "$DRIVER" != "mongo" ]; then
            sed -i '' 's/DriverName: "mongo"/DriverName: "'"$DRIVER"'"/g' "$f" 2>/dev/null || \
            sed -i 's/DriverName: "mongo"/DriverName: "'"$DRIVER"'"/g' "$f"
        fi
        if [ "$DRIVER" != "mysql" ]; then
            sed -i '' 's/DriverName: "mysql"/DriverName: "'"$DRIVER"'"/g' "$f" 2>/dev/null || \
            sed -i 's/DriverName: "mysql"/DriverName: "'"$DRIVER"'"/g' "$f"
        fi
        echo "Fixed $f (driver: $DRIVER)"
    fi
done

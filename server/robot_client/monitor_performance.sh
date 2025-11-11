#!/bin/bash

# 性能监控脚本
# 在压测期间监控服务器性能

cd "$(dirname "$0")/.."

echo "========== 性能监控 =========="
echo "监控服务: game, gate, web"
echo "按 Ctrl+C 停止监控"
echo ""

# 创建监控目录
mkdir -p logs/monitor

# 监控函数
monitor_service() {
    SERVICE=$1
    PID=$(pgrep -f "bin/$SERVICE" | head -n 1)
    
    if [ -z "$PID" ]; then
        echo "[$SERVICE] 服务未运行"
        return
    fi

    LOG_FILE="logs/monitor/${SERVICE}_$(date +%Y%m%d_%H%M%S).log"
    
    echo "[$SERVICE] PID: $PID, 日志: $LOG_FILE"
    
    (
        echo "时间,CPU%,内存(MB),Goroutine数" > "$LOG_FILE"
        
        while true; do
            if ! ps -p $PID > /dev/null 2>&1; then
                break
            fi
            
            # CPU和内存
            CPU_MEM=$(ps -p $PID -o %cpu,rss= | tail -n 1)
            CPU=$(echo $CPU_MEM | awk '{print $1}')
            MEM_KB=$(echo $CPU_MEM | awk '{print $2}')
            MEM_MB=$((MEM_KB / 1024))
            
            # Goroutine数量（需要服务支持gops）
            GOROUTINES=$(curl -s http://localhost:$(grep -oP 'gops.*:\K[0-9]+' <(ps aux | grep $SERVICE))/debug/pprof/goroutine?debug=1 2>/dev/null | grep -c "goroutine" || echo "N/A")
            
            TIMESTAMP=$(date +"%Y-%m-%d %H:%M:%S")
            echo "$TIMESTAMP,$CPU,$MEM_MB,$GOROUTINES" >> "$LOG_FILE"
            
            printf "\r[$SERVICE] CPU: %5s%% | 内存: %6d MB | Goroutines: %s" "$CPU" "$MEM_MB" "$GOROUTINES"
            
            sleep 1
        done
    ) &
}

# 启动监控
monitor_service "game"
monitor_service "gate"
monitor_service "web"

# 等待
wait






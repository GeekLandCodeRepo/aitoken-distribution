#!/usr/bin/env bash
set -e

cd "$(dirname "$0")"

APP_DIR="$(pwd)"
RUNTIME_DIR="$APP_DIR/runtime"
PID_DIR="$RUNTIME_DIR/pid"
LOG_DIR="$RUNTIME_DIR/log"

# 加载 .env
if [ -f .env ]; then
    set -a
    source .env
    set +a
fi

mkdir -p "$PID_DIR" "$LOG_DIR"

if [ ! -x "$APP_DIR/bin/aitsd" ]; then
    echo "Missing executable: $APP_DIR/bin/aitsd"
    exit 1
fi

if [ ! -x "$APP_DIR/bin/aitsd-worker" ]; then
    echo "Missing executable: $APP_DIR/bin/aitsd-worker"
    exit 1
fi

# 关闭旧进程
stop_one() {
    local name="$1"
    local pid_file="$PID_DIR/${name}.pid"

    if [ -f "$pid_file" ]; then
        local pid
        pid=$(cat "$pid_file")
        if [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null; then
            echo "Stopping old $name (PID: $pid)..."
            kill "$pid"

            for i in $(seq 1 10); do
                if ! kill -0 "$pid" 2>/dev/null; then
                    echo "$name stopped."
                    break
                fi
                sleep 1
            done

            if kill -0 "$pid" 2>/dev/null; then
                echo "Timeout, force killing $name..."
                kill -9 "$pid" 2>/dev/null || true
            fi
        fi
        rm -f "$pid_file"
    fi
}

stop_one "aitsd"
stop_one "aitsd-worker"

start_one() {
    local name="$1"
    local cmd="$2"
    local log_file="$3"

    echo "Starting $name..."
    nohup "$cmd" >> "$log_file" 2>&1 &
    local pid=$!
    echo "$pid" > "$PID_DIR/${name}.pid"

    sleep 1
    if ! kill -0 "$pid" 2>/dev/null; then
        echo "$name failed to start. See log: $log_file"
        rm -f "$PID_DIR/${name}.pid"
        exit 1
    fi

    echo "  $name started (PID: $pid)"
}

# 先启动 worker，再启动 server
start_one "aitsd-worker" "$APP_DIR/bin/aitsd-worker" "$LOG_DIR/worker.log"
start_one "aitsd" "$APP_DIR/bin/aitsd" "$LOG_DIR/aitsd.log"

echo "All services started."

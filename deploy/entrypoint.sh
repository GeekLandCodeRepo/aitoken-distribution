#!/bin/sh
set -eu

MIGRATION_MARKER="${MIGRATION_MARKER:-/srv/data/.migrated}"
RUN_MIGRATE_ON_START="${RUN_MIGRATE_ON_START:-true}"
FORCE_MIGRATE="${FORCE_MIGRATE:-false}"
STATUS=0

shutdown() {
  if [ -n "${SERVER_PID:-}" ] && kill -0 "$SERVER_PID" 2>/dev/null; then
    kill "$SERVER_PID" 2>/dev/null || true
  fi
  if [ -n "${WORKER_PID:-}" ] && kill -0 "$WORKER_PID" 2>/dev/null; then
    kill "$WORKER_PID" 2>/dev/null || true
  fi
}

trap shutdown INT TERM

if [ "$RUN_MIGRATE_ON_START" = "true" ]; then
  if [ ! -f "$MIGRATION_MARKER" ] || [ "$FORCE_MIGRATE" = "true" ]; then
    echo "Running database migration..."
    aitsd-migrate --skip-env
    mkdir -p "$(dirname "$MIGRATION_MARKER")"
    date -u +%Y-%m-%dT%H:%M:%SZ > "$MIGRATION_MARKER"
    echo "Migration completed."
  else
    echo "Migration already completed, skipping."
  fi
fi

echo "Starting worker..."
aitsd-worker &
WORKER_PID=$!

echo "Starting API/Web server..."
aitsd &
SERVER_PID=$!

while true; do
  if ! kill -0 "$SERVER_PID" 2>/dev/null; then
    wait "$SERVER_PID" || STATUS=$?
    shutdown
    exit "${STATUS:-1}"
  fi
  if ! kill -0 "$WORKER_PID" 2>/dev/null; then
    wait "$WORKER_PID" || STATUS=$?
    shutdown
    exit "${STATUS:-1}"
  fi
  sleep 2
done

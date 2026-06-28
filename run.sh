set -e

cd "$(dirname "$0")"

# 加载 .env
if [ -f .env ]; then
    set -a
    source .env
    set +a
fi

export CGO_ENABLED=0

echo "Starting LLM Gateway..."
go run ./cmd/server/

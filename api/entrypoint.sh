#!/usr/bin/env sh
# Start gunicorn (HTTP :5000) and gRPC Retriever (:50051) in one container.
set -eu
PORT="${PYTHON_SERVICE_PORT:-5000}"
GRPC_PORT="${PYTHON_GRPC_PORT:-50051}"

python - <<'PY' &
import logging
import os
import time

logging.basicConfig(level=logging.INFO)
from api.grpc_retriever import serve

port = int(os.environ.get("PYTHON_GRPC_PORT", "50051"))
server = serve(port)
try:
    while True:
        time.sleep(3600)
except KeyboardInterrupt:
    server.stop(grace=5)
PY
GRPC_PID=$!

gunicorn -w "${GUNICORN_WORKERS:-2}" -b "0.0.0.0:${PORT}" \
  --timeout "${GUNICORN_TIMEOUT:-120}" --graceful-timeout 30 \
  --access-logfile - --error-logfile - api.app:app &
HTTP_PID=$!

term() {
  kill -TERM "$HTTP_PID" "$GRPC_PID" 2>/dev/null || true
  wait "$HTTP_PID" "$GRPC_PID" 2>/dev/null || true
}
trap term INT TERM

# dash/busybox lack `wait -n` — poll until either child exits
while kill -0 "$HTTP_PID" 2>/dev/null && kill -0 "$GRPC_PID" 2>/dev/null; do
  sleep 2
done
term
exit 1

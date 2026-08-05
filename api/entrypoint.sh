#!/usr/bin/env sh
# Supervisor: Gunicorn (HTTP) + gRPC Retriever in one container.
# Under Docker, tini is PID 1 and forwards SIGTERM/SIGINT here.
set -eu

export PYTHON_SERVICE_PORT="${PYTHON_SERVICE_PORT:-5000}"
export PYTHON_GRPC_PORT="${PYTHON_GRPC_PORT:-50051}"

python -m api.grpc &
GRPC_PID=$!

gunicorn -w "${GUNICORN_WORKERS:-2}" -b "0.0.0.0:${PYTHON_SERVICE_PORT}" \
  --timeout "${GUNICORN_TIMEOUT:-120}" --graceful-timeout 30 \
  --access-logfile - --error-logfile - api.http.app:app &
HTTP_PID=$!

shutdown() {
  kill -TERM "$HTTP_PID" "$GRPC_PID" 2>/dev/null || true
  wait "$HTTP_PID" 2>/dev/null || true
  wait "$GRPC_PID" 2>/dev/null || true
  exit 0
}
trap shutdown INT TERM

# dash/busybox lack `wait -n` — poll until either child exits
while kill -0 "$HTTP_PID" 2>/dev/null && kill -0 "$GRPC_PID" 2>/dev/null; do
  sleep 2
done

status=1
if ! kill -0 "$HTTP_PID" 2>/dev/null; then
  wait "$HTTP_PID" 2>/dev/null || status=$?
  kill -TERM "$GRPC_PID" 2>/dev/null || true
  wait "$GRPC_PID" 2>/dev/null || true
elif ! kill -0 "$GRPC_PID" 2>/dev/null; then
  wait "$GRPC_PID" 2>/dev/null || status=$?
  kill -TERM "$HTTP_PID" 2>/dev/null || true
  wait "$HTTP_PID" 2>/dev/null || true
fi

exit "$status"

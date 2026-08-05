#!/usr/bin/env bash
set -Eeuo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="${VINE_PROJECT_ROOT:-$(cd -- "$SCRIPT_DIR/.." && pwd)}"
WEB_DIR="${VINE_WEB_DIR:-$PROJECT_ROOT/web}"
SERVER_PACKAGE="${VINE_SERVER_PACKAGE:-./cmd/server}"
PREPARE_PACKAGE="${VINE_PREPARE_PACKAGE:-}"
HOST="${VINE_HOST:-127.0.0.1}"
VITE_PORT="${VINE_VITE_PORT:-5174}"
PUBLIC_PORT="${VINE_PUBLIC_PORT:-7288}"
DASHBOARD_PORT="${VINE_DASHBOARD_PORT:-7299}"
STARTUP_TIMEOUT="${VINE_STARTUP_TIMEOUT:-45}"

INSTALL=0
CHECK_ONLY=0
FRONTEND_PID=""
BACKEND_PID=""
FRONTEND_GROUP=0
BACKEND_GROUP=0
CLEANED_UP=0

usage() {
    cat <<'EOF'
Usage: ./scripts/start_vine_app.sh [--install] [--check]

Starts a browser-enabled Vine Standalone project:
  http://127.0.0.1:7288/            Portal WEBGW frontend
  http://127.0.0.1:7288/api/invoke  Portal RPCGW vRPC
  http://127.0.0.1:7299/            Standalone Dashboard
  http://127.0.0.1:5174/            Vite development upstream

Options:
  --install  Run pnpm install before startup.
  --check    Validate files, commands, dependencies, and ports without starting.
  -h, --help Show this help.

Optional environment overrides:
  VINE_PROJECT_ROOT, VINE_WEB_DIR, VINE_SERVER_PACKAGE,
  VINE_PREPARE_PACKAGE, VINE_HOST, VINE_VITE_PORT,
  VINE_PUBLIC_PORT, VINE_DASHBOARD_PORT, VINE_STARTUP_TIMEOUT

Set VINE_PREPARE_PACKAGE=./cmd/migrate only when the target project explicitly
requires that preparation step. The script never infers or runs migrations.
EOF
}

fail() {
    printf '[vine-start] error: %s\n' "$*" >&2
    exit 1
}

is_positive_integer() {
    [[ "$1" =~ ^[1-9][0-9]*$ ]]
}

validate_port() {
    local name="$1"
    local value="$2"
    is_positive_integer "$value" || fail "$name must be an integer between 1 and 65535"
    (( value <= 65535 )) || fail "$name must be between 1 and 65535"
}

command_exists() {
    command -v "$1" >/dev/null 2>&1
}

port_is_open() {
    local host="$1"
    local port="$2"
    (exec 3<>"/dev/tcp/$host/$port") >/dev/null 2>&1
}

require_available_port() {
    local port="$1"
    if port_is_open "$HOST" "$port"; then
        fail "required listener $HOST:$port is already in use"
    fi
}

start_child() {
    local workdir="$1"
    local __pid_var="$2"
    local __group_var="$3"
    shift 3

    local pid
    local grouped=0
    printf '[vine-start] start:'
    printf ' %q' "$@"
    printf '\n'

    if command_exists setsid; then
        (
            cd -- "$workdir"
            exec setsid "$@"
        ) &
        pid=$!
        grouped=1
    else
        (
            cd -- "$workdir"
            exec "$@"
        ) &
        pid=$!
    fi

    printf -v "$__pid_var" '%s' "$pid"
    printf -v "$__group_var" '%s' "$grouped"
}

stop_child() {
    local name="$1"
    local pid="$2"
    local grouped="$3"

    [[ -n "$pid" ]] || return 0
    [[ "$pid" =~ ^[1-9][0-9]*$ ]] || return 0

    printf '[vine-start] stop %s\n' "$name"
    if (( grouped == 1 )); then
        kill -TERM -- "-$pid" >/dev/null 2>&1 || true
    else
        kill -TERM -- "$pid" >/dev/null 2>&1 || true
    fi

    local deadline=$((SECONDS + 5))
    while kill -0 "$pid" >/dev/null 2>&1 && (( SECONDS < deadline )); do
        sleep 0.1
    done

    if kill -0 "$pid" >/dev/null 2>&1; then
        if (( grouped == 1 )); then
            kill -KILL -- "-$pid" >/dev/null 2>&1 || true
        else
            kill -KILL -- "$pid" >/dev/null 2>&1 || true
        fi
    fi
    wait "$pid" >/dev/null 2>&1 || true
}

cleanup() {
    (( CLEANED_UP == 0 )) || return 0
    CLEANED_UP=1
    stop_child "Vine server" "$BACKEND_PID" "$BACKEND_GROUP"
    stop_child "frontend" "$FRONTEND_PID" "$FRONTEND_GROUP"
}

wait_for_port() {
    local name="$1"
    local pid="$2"
    local port="$3"
    local deadline=$((SECONDS + STARTUP_TIMEOUT))

    while (( SECONDS < deadline )); do
        kill -0 "$pid" >/dev/null 2>&1 || fail "$name exited before readiness"
        if port_is_open "$HOST" "$port"; then
            printf '[vine-start] %s is listening on %s:%s\n' "$name" "$HOST" "$port"
            return 0
        fi
        sleep 0.2
    done
    fail "timed out waiting for $name on $HOST:$port"
}

http_status() {
    local host="$1"
    local port="$2"
    local status=""

    if ! exec 3<>"/dev/tcp/$host/$port"; then
        return 1
    fi
    printf 'GET / HTTP/1.0\r\nHost: %s:%s\r\nConnection: close\r\n\r\n'         "$host" "$port" >&3
    IFS=' ' read -r -t 1 _ status _ <&3 || true
    exec 3>&- 3<&-
    [[ "$status" =~ ^[0-9]{3}$ ]] || return 1
    printf '%s\n' "$status"
}

wait_for_public_page() {
    local pid="$1"
    local deadline=$((SECONDS + STARTUP_TIMEOUT))
    local status=""

    while (( SECONDS < deadline )); do
        kill -0 "$pid" >/dev/null 2>&1 || fail "Vine server exited before Portal readiness"
        status="$(http_status "$HOST" "$PUBLIC_PORT" 2>/dev/null || true)"
        if [[ "$status" =~ ^[23][0-9][0-9]$ ]]; then
            printf '[vine-start] Portal page is ready: http://%s:%s/ (HTTP %s)\n'                 "$HOST" "$PUBLIC_PORT" "$status"
            return 0
        fi
        sleep 0.2
    done
    fail "timed out waiting for Portal page on $HOST:$PUBLIC_PORT; last HTTP status: ${status:-none}"
}

trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

while (($# > 0)); do
    case "$1" in
        --install)
            INSTALL=1
            ;;
        --check)
            CHECK_ONLY=1
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            usage >&2
            fail "unknown argument: $1"
            ;;
    esac
    shift
done

validate_port "VINE_VITE_PORT" "$VITE_PORT"
validate_port "VINE_PUBLIC_PORT" "$PUBLIC_PORT"
validate_port "VINE_DASHBOARD_PORT" "$DASHBOARD_PORT"
is_positive_integer "$STARTUP_TIMEOUT" || fail "VINE_STARTUP_TIMEOUT must be a positive integer"

[[ "$VITE_PORT" != "$PUBLIC_PORT" ]] || fail "Vite and Portal ports must be distinct"
[[ "$VITE_PORT" != "$DASHBOARD_PORT" ]] || fail "Vite and Dashboard ports must be distinct"
[[ "$PUBLIC_PORT" != "$DASHBOARD_PORT" ]] || fail "Portal and Dashboard ports must be distinct"

[[ -f "$PROJECT_ROOT/go.mod" ]] || fail "go.mod is missing from project root: $PROJECT_ROOT"
[[ -f "$WEB_DIR/package.json" ]] || fail "frontend package.json is missing: $WEB_DIR/package.json"
command_exists go || fail "required command is unavailable on PATH: go"
command_exists pnpm || fail "required command is unavailable on PATH: pnpm"

if [[ ! -d "$WEB_DIR/node_modules" && "$INSTALL" -eq 0 ]]; then
    fail "frontend dependencies are missing; rerun with --install or run 'pnpm install' in $WEB_DIR"
fi

require_available_port "$VITE_PORT"
require_available_port "$PUBLIC_PORT"
require_available_port "$DASHBOARD_PORT"

if (( CHECK_ONLY == 1 )); then
    printf '[vine-start] preflight passed\n'
    exit 0
fi

if (( INSTALL == 1 )); then
    printf '[vine-start] run: pnpm install\n'
    (cd -- "$WEB_DIR" && pnpm install)
fi

if [[ -n "$PREPARE_PACKAGE" ]]; then
    printf '[vine-start] run: go run %q\n' "$PREPARE_PACKAGE"
    (cd -- "$PROJECT_ROOT" && go run "$PREPARE_PACKAGE")
fi

start_child "$WEB_DIR" FRONTEND_PID FRONTEND_GROUP pnpm run dev
wait_for_port "frontend" "$FRONTEND_PID" "$VITE_PORT"

start_child "$PROJECT_ROOT" BACKEND_PID BACKEND_GROUP go run "$SERVER_PACKAGE"
wait_for_public_page "$BACKEND_PID"

printf '[vine-start] application: http://%s:%s/\n' "$HOST" "$PUBLIC_PORT"
printf '[vine-start] vRPC: http://%s:%s/api/invoke\n' "$HOST" "$PUBLIC_PORT"
printf '[vine-start] Dashboard: http://%s:%s/\n' "$HOST" "$DASHBOARD_PORT"
printf '[vine-start] press Ctrl+C to stop both processes\n'

set +e
wait -n "$FRONTEND_PID" "$BACKEND_PID"
STATUS=$?
set -e
fail "a child process exited with status $STATUS; stopping the other process"


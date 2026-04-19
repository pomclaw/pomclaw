#!/bin/bash
# setup-postgres.sh — One-shot PostgreSQL setup for pomclaw
#
# Usage:
#   ./scripts/setup-postgres.sh                  # uses default password
#   ./scripts/setup-postgres.sh MyPassword123    # custom password
#
# Requirements: docker, pomclaw binary built (make build)

set -euo pipefail

POSTGRES_PWD="${1:-Pomclaw123}"
CONTAINER_NAME="postgres-pomclaw"
POSTGRES_IMAGE="pgvector/pgvector:pg16"
POM_BIN="./build/pomclaw-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')"
CONFIG_FILE="$HOME/.pomclaw/config.json"

# ── helpers ──────────────────────────────────────────────────────────────────
info()    { echo "  $*"; }
ok()      { echo "✓ $*"; }
fail()    { echo "✗ $*" >&2; exit 1; }
section() { echo; echo "── $* ─────────────────────────────────────────"; }

# ── preflight ─────────────────────────────────────────────────────────────────
section "Preflight"
command -v docker >/dev/null 2>&1 || fail "docker not found. Install Docker first."
[ -f "$POM_BIN" ] || fail "Binary not found at $POM_BIN. Run 'make build' first."
[ -f "$CONFIG_FILE" ] || fail "Config not found at $CONFIG_FILE. Run 'pomclaw onboard' first."
ok "Prerequisites met"

# ── step 1: start PostgreSQL container ───────────────────────────────────────
section "Step 1/4: PostgreSQL container with pgvector"
if docker ps --format '{{.Names}}' | grep -qx "$CONTAINER_NAME"; then
    ok "Container '$CONTAINER_NAME' already running — skipping"
elif docker ps -a --format '{{.Names}}' | grep -qx "$CONTAINER_NAME"; then
    info "Restarting stopped container '$CONTAINER_NAME'..."
    docker start "$CONTAINER_NAME"
    ok "Container started"
else
    info "Pulling and starting PostgreSQL with pgvector..."
    docker run -d --name "$CONTAINER_NAME" \
        -p 5432:5432 \
        -e POSTGRES_USER=pomclaw \
        -e POSTGRES_PASSWORD="$POSTGRES_PWD" \
        -e POSTGRES_DB=pomclaw \
        -v postgres-pomclaw-data:/var/lib/postgresql/data \
        "$POSTGRES_IMAGE"
    ok "Container launched"
fi

info "Waiting for database to be ready..."
TIMEOUT=60
ELAPSED=0
while ! docker exec "$CONTAINER_NAME" pg_isready -U pomclaw >/dev/null 2>&1; do
    sleep 2
    ELAPSED=$((ELAPSED + 2))
    printf "\r  Waiting... %ds" "$ELAPSED"
    [ "$ELAPSED" -ge "$TIMEOUT" ] && fail "Timed out after ${TIMEOUT}s. Check: docker logs $CONTAINER_NAME"
done
echo
ok "PostgreSQL is ready"

# ── step 2: enable pgvector extension ────────────────────────────────────────
section "Step 2/4: pgvector extension"
info "Enabling pgvector extension..."
docker exec "$CONTAINER_NAME" psql -U pomclaw -d pomclaw -c "CREATE EXTENSION IF NOT EXISTS vector;" >/dev/null 2>&1
ok "pgvector enabled"

# ── step 3: patch config ──────────────────────────────────────────────────────
section "Step 3/4: Config"
info "Patching $CONFIG_FILE with PostgreSQL settings..."
python3 - "$CONFIG_FILE" "$POSTGRES_PWD" <<'PYEOF'
import json, sys
path, pwd = sys.argv[1], sys.argv[2]
with open(path) as f:
    cfg = json.load(f)

# Update storage type
cfg["storage_type"] = "postgres"

# Update postgres config
cfg.setdefault("postgres", {}).update({
    "enabled": True,
    "host": "localhost",
    "port": 5432,
    "database": "pomclaw",
    "user": "pomclaw",
    "password": pwd,
    "ssl_mode": "disable",
    "pool_max_open": 25,
    "pool_max_idle": 5
})

# Disable oracle if present
if "oracle" in cfg:
    cfg["oracle"]["enabled"] = False

with open(path, "w") as f:
    json.dump(cfg, f, indent=2)
print("  patched successfully")
PYEOF
ok "Config updated"

# ── step 4: initialize schema ────────────────────────────────────────────────
section "Step 4/4: Schema initialization"
info "Running pomclaw setup-database..."
"$POM_BIN" setup-database

echo
echo "════════════════════════════════════════════════════════"
echo "  PostgreSQL setup complete!"
echo "  Connection: postgresql://pomclaw:${POSTGRES_PWD}@localhost:5432/pomclaw"
echo "  Test with:"
echo "    $POM_BIN agent -m \"Remember that I love Go\""
echo "    $POM_BIN agent -m \"What language do I like?\""
echo "════════════════════════════════════════════════════════"

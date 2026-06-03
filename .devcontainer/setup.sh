#!/bin/bash
set -e

ARCH=$(uname -m)
# aarch64 -> arm64 に正規化 (buf のバイナリ命名に合わせる)
if [ "$ARCH" = "aarch64" ]; then ARCH="arm64"; fi

echo "--- Installing Claude Code ---"
curl -fsSL https://claude.ai/install.sh | bash
# インストール先が PATH に含まれない場合に /usr/local/bin へ symlink
if ! command -v claude &>/dev/null; then
  CLAUDE_BIN=$(find /root /home -name "claude" -type f 2>/dev/null | head -1)
  if [ -n "$CLAUDE_BIN" ]; then
    ln -sf "$CLAUDE_BIN" /usr/local/bin/claude
    echo "Linked claude: $CLAUDE_BIN -> /usr/local/bin/claude"
  fi
fi
claude --version

echo "--- Installing buf CLI ---"
curl -sSL "https://github.com/bufbuild/buf/releases/latest/download/buf-$(uname -s)-${ARCH}" \
  -o /usr/local/bin/buf
chmod +x /usr/local/bin/buf
buf --version

echo "--- Installing sqlc ---"
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
sqlc version

echo "--- Installing golang-migrate ---"
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
migrate -version

echo "--- Installing go-arch-lint ---"
go install github.com/fe3dback/go-arch-lint@latest

echo "--- Downloading Go modules ---"
cd /workspace && go mod download

echo "--- Installing web dependencies ---"
cd /workspace/web && npm install

echo "--- Setup complete ---"

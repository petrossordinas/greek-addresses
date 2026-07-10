#!/usr/bin/env bash
# Installs graddresses as a systemd service, no Docker required.
#
#   curl -fsSL https://raw.githubusercontent.com/petrossordinas/greek-addresses/master/install.sh | sudo bash
#
# Re-running this script upgrades the binary and systemd unit in place. It
# never overwrites an existing database or config file.
set -euo pipefail

REPO="${GRADDRESSES_REPO:-petrossordinas/greek-addresses}"
INSTALL_BIN="/usr/local/bin/graddresses"
DATA_DIR="/var/lib/graddresses"
CONFIG_DIR="/etc/graddresses"
CONFIG_FILE="$CONFIG_DIR/graddresses.env"
SERVICE_FILE="/etc/systemd/system/graddresses.service"
SERVICE_USER="graddresses"

if [ "$(id -u)" -ne 0 ]; then
	echo "This installer must be run as root (e.g. via sudo)." >&2
	exit 1
fi

if [ "$(uname -s)" != "Linux" ]; then
	echo "This installer only supports Linux." >&2
	exit 1
fi

case "$(uname -m)" in
amd64 | x86_64) ARCH="amd64" ;;
arm64 | aarch64) ARCH="arm64" ;;
*)
	echo "Unsupported architecture: $(uname -m)" >&2
	exit 1
	;;
esac

WORK_DIR="$(mktemp -d)"
trap 'rm -rf "$WORK_DIR"' EXIT

ASSET="graddresses-linux-${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/latest/download/${ASSET}"

echo "Downloading $URL..."
curl -fsSL "$URL" -o "$WORK_DIR/graddresses.tar.gz"
tar -xzf "$WORK_DIR/graddresses.tar.gz" -C "$WORK_DIR"

if ! id "$SERVICE_USER" >/dev/null 2>&1; then
	echo "Creating system user $SERVICE_USER..."
	useradd --system --no-create-home --shell /usr/sbin/nologin "$SERVICE_USER"
fi

echo "Stopping existing service (if any)..."
systemctl stop graddresses.service 2>/dev/null || true

mkdir -p "$DATA_DIR" "$CONFIG_DIR"

echo "Installing binary to $INSTALL_BIN..."
install -m 755 "$WORK_DIR/graddresses" "$INSTALL_BIN"

if [ ! -f "$DATA_DIR/gr_addresses.db" ]; then
	echo "Installing database to $DATA_DIR/gr_addresses.db..."
	install -m 644 "$WORK_DIR/gr_addresses.db" "$DATA_DIR/gr_addresses.db"
else
	echo "Existing database found at $DATA_DIR/gr_addresses.db, leaving it in place."
fi
chown -R "$SERVICE_USER:$SERVICE_USER" "$DATA_DIR"

if [ ! -f "$CONFIG_FILE" ]; then
	echo "Writing default config to $CONFIG_FILE..."
	printf 'PORT=%s\n' "${PORT:-9013}" >"$CONFIG_FILE"
else
	echo "Existing config found at $CONFIG_FILE, leaving it in place."
fi

echo "Installing systemd unit to $SERVICE_FILE..."
cat >"$SERVICE_FILE" <<EOF
[Unit]
Description=GrAddresses - Greek address search API
After=network.target

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_USER
WorkingDirectory=$DATA_DIR
EnvironmentFile=-$CONFIG_FILE
ExecStart=$INSTALL_BIN -db=$DATA_DIR/gr_addresses.db
Restart=on-failure
RestartSec=2

NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$DATA_DIR
PrivateTmp=true

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable --now graddresses.service

echo
echo "graddresses is installed and running."
echo "  status:  systemctl status graddresses"
echo "  logs:    journalctl -u graddresses -f"
echo "  config:  $CONFIG_FILE (PORT, DB_PATH)"

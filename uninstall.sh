#!/usr/bin/env bash
# Uninstalls the graddresses systemd service installed by install.sh.
#
#   curl -fsSL https://raw.githubusercontent.com/petrossordinas/greek-addresses/master/uninstall.sh | sudo bash
#
# By default this keeps the database and config file (/var/lib/graddresses,
# /etc/graddresses) in case you reinstall later. Pass --purge to remove
# those too.
set -euo pipefail

INSTALL_BIN="/usr/local/bin/graddresses"
DATA_DIR="/var/lib/graddresses"
CONFIG_DIR="/etc/graddresses"
SERVICE_FILE="/etc/systemd/system/graddresses.service"
SERVICE_USER="graddresses"

PURGE=0
for arg in "$@"; do
	case "$arg" in
	--purge) PURGE=1 ;;
	*)
		echo "Unknown option: $arg" >&2
		exit 1
		;;
	esac
done

if [ "$(id -u)" -ne 0 ]; then
	echo "This uninstaller must be run as root (e.g. via sudo)." >&2
	exit 1
fi

echo "Stopping graddresses service..."
systemctl stop graddresses.service 2>/dev/null || true
systemctl disable graddresses.service 2>/dev/null || true

if [ -f "$SERVICE_FILE" ]; then
	echo "Removing systemd unit..."
	rm -f "$SERVICE_FILE"
	systemctl daemon-reload
fi

if [ -f "$INSTALL_BIN" ]; then
	echo "Removing binary $INSTALL_BIN..."
	rm -f "$INSTALL_BIN"
fi

if [ "$PURGE" -eq 1 ]; then
	echo "Purging data and config ($DATA_DIR, $CONFIG_DIR)..."
	rm -rf "$DATA_DIR" "$CONFIG_DIR"
else
	echo "Keeping $DATA_DIR and $CONFIG_DIR (pass --purge to remove them too)."
fi

if id "$SERVICE_USER" >/dev/null 2>&1; then
	echo "Removing system user $SERVICE_USER..."
	userdel "$SERVICE_USER" 2>/dev/null || true
fi

echo
echo "graddresses has been uninstalled."

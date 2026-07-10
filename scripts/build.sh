#!/usr/bin/env bash
# Cross-compiles static release binaries and packages each with the sqlite
# database into a tarball under dist/. Asset names are version-independent
# (graddresses-<os>-<arch>.tar.gz) so they work with GitHub's
# /releases/latest/download/<asset> URLs.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="$ROOT_DIR/dist"
TARGETS=("linux/amd64" "linux/arm64")

rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

for target in "${TARGETS[@]}"; do
	os="${target%/*}"
	arch="${target#*/}"
	name="graddresses-${os}-${arch}"
	build_dir="$DIST_DIR/$name"
	mkdir -p "$build_dir"

	echo "Building $name..."
	(
		cd "$ROOT_DIR"
		CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build -o "$build_dir/graddresses" .
	)
	cp "$ROOT_DIR/db/gr_addresses.db" "$build_dir/gr_addresses.db"

	tar -C "$DIST_DIR" -czf "$DIST_DIR/$name.tar.gz" "$name"
	rm -rf "$build_dir"
	echo "  -> dist/$name.tar.gz"
done

echo
echo "Artifacts:"
ls -la "$DIST_DIR"

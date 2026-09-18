#!/usr/bin/env sh
set -eu

backup_path=${1:?"Usage: scripts/restore.sh BACKUP_DIRECTORY --force"}
confirmation=${2:-}
database_path=${DATABASE_PATH:-./data/librearchive.db}
storage_path=${STORAGE_PATH:-./data/storage}

if [ "$confirmation" != "--force" ]; then
  echo "Restoring overwrites the configured database and storage. Pass --force to continue." >&2
  exit 1
fi
if [ ! -f "$backup_path/database.sqlite" ] || [ ! -d "$backup_path/storage" ]; then
  echo "Backup must contain database.sqlite and storage/." >&2
  exit 1
fi
case "$database_path" in ''|/|.) echo "Unsafe DATABASE_PATH." >&2; exit 1;; esac
case "$storage_path" in ''|/|.) echo "Unsafe STORAGE_PATH." >&2; exit 1;; esac

if [ -f "$backup_path/checksums.sha256" ] && command -v sha256sum >/dev/null 2>&1; then
  (
    cd "$backup_path"
    sha256sum -c checksums.sha256
  )
fi

mkdir -p "$(dirname "$database_path")"
rm -f "$database_path"
rm -rf "$storage_path"
cp "$backup_path/database.sqlite" "$database_path"
cp -R "$backup_path/storage" "$storage_path"

echo "Restore completed. Start LibreArchive with the same DATABASE_PATH and STORAGE_PATH values."

#!/usr/bin/env sh
set -eu

database_path=${DATABASE_PATH:-./data/librearchive.db}
storage_path=${STORAGE_PATH:-./data/storage}
backup_root=${1:?"Usage: scripts/backup.sh BACKUP_DIRECTORY"}
timestamp=$(date -u +%Y%m%dT%H%M%SZ)
backup_path="$backup_root/librearchive-$timestamp"

if ! command -v sqlite3 >/dev/null 2>&1; then
  echo "sqlite3 is required to create a consistent database backup." >&2
  exit 1
fi
if [ ! -f "$database_path" ]; then
  echo "Database file does not exist: $database_path" >&2
  exit 1
fi
if [ ! -d "$storage_path" ]; then
  echo "Storage directory does not exist: $storage_path" >&2
  exit 1
fi

mkdir -p "$backup_path"
sqlite3 "$database_path" ".backup '$backup_path/database.sqlite'"
cp -R "$storage_path" "$backup_path/storage"

printf 'created_at=%s\n' "$timestamp" > "$backup_path/manifest.env"
printf 'database_file=database.sqlite\n' >> "$backup_path/manifest.env"
printf 'storage_directory=storage\n' >> "$backup_path/manifest.env"

if command -v sha256sum >/dev/null 2>&1; then
  (
    cd "$backup_path"
    sha256sum database.sqlite > checksums.sha256
  )
fi

echo "Backup created at: $backup_path"

# LibreArchive API

LibreArchive is a self-hosted PDF library and reading platform. This repository currently contains the first API foundation: authentication, administrator and reader roles, user management, and protected PDF book catalog management.

## Requirements

- Go 1.27 or newer
- A `JWT_SECRET` value with enough entropy for production

SQLite is the default database. It is intentionally suitable for a portable, single-household deployment. The database file and media storage location can be changed with environment variables.

## Run locally

```sh
export JWT_SECRET='replace-with-a-long-random-secret'
go run ./cmd/api
```

The API automatically loads a `.env` file from the working directory when present. Existing environment variables take precedence over values in that file. Copy `.env.example` to `.env`, set a unique `JWT_SECRET`, and do not commit the resulting `.env` file.

The API listens on `:8080` by default. On first run it creates these persistent locations:

- `data/librearchive.db` — SQLite database
- `data/storage/books` — uploaded PDFs
- `data/storage/covers` — uploaded cover images

On its first start with no administrator in the database, the API creates an active administrator with username `admin` and name `admin`. It prints a generated twelve-word passphrase to the server console once; store it securely, because only its bcrypt hash is persisted. On later starts, no passphrase is printed and no account is recreated. Subsequent users must be created by an authenticated administrator and receive their generated passphrase once.

Access tokens last one day for every role. Refresh tokens are signed JWTs and rotate on use; they last one day for administrators and 365 days for readers. Logging out, resetting a passphrase, deactivating, or deleting an account revokes its active sessions.

## Configuration

| Variable | Default | Purpose |
| --- | --- | --- |
| `LISTEN_ADDRESS` | `:8080` | HTTP address to listen on. |
| `DATABASE_PATH` | `./data/librearchive.db` | SQLite database path. |
| `STORAGE_PATH` | `./data/storage` | Parent directory for managed PDFs and covers. |
| `JWT_SECRET` | none | Required secret used to sign access tokens. |
| `MAX_UPLOAD_BYTES` | `262144000` | Maximum total multipart upload size, in bytes. |
| `CORS_ALLOWED_ORIGINS` | none | Comma-separated origins allowed to call the API from a browser. |

## Current API endpoints

Public endpoints:

- `GET /health`
- `POST /api/v1/auth/login`

Authenticated endpoints:

- `GET /api/v1/me`
- `GET /api/v1/books?search=&page=&pageSize=`
- `GET /api/v1/books/{bookID}`
- `GET /api/v1/books/{bookID}/file`
- `GET /api/v1/books/{bookID}/cover`
- `GET`, `PUT /api/v1/books/{bookID}/reading-progress`
- `GET /api/v1/reading-progress?status=unread|in_progress|finished&downloaded=true|false`
- `GET /api/v1/favorites`
- `PUT`, `DELETE /api/v1/books/{bookID}/favorite`
- `GET`, `POST /api/v1/collections`
- `GET`, `PATCH`, `DELETE /api/v1/collections/{collectionID}`
- `PUT`, `DELETE /api/v1/collections/{collectionID}/books/{bookID}`
- `GET /api/v1/books/{bookID}/annotations?type=highlight|note&color=`
- `POST /api/v1/books/{bookID}/highlights`
- `PATCH`, `DELETE /api/v1/highlights/{highlightID}`
- `POST /api/v1/highlights/{highlightID}/notes`
- `PATCH`, `DELETE /api/v1/notes/{noteID}`

Administrator-only endpoints:

- `GET`, `POST /api/v1/users`
- `PATCH /api/v1/users/{userID}`
- `POST /api/v1/books` (multipart form with `title`, `author`, and PDF `file`; optional `cover`)
- `PATCH`, `DELETE /api/v1/books/{bookID}`

Protected routes require `Authorization: Bearer <token>`.

## Backup and migration

The included scripts require `sqlite3`. Stop the service before restore operations. A backup can run while the service is active because SQLite's `.backup` command creates a consistent database copy.

```sh
scripts/backup.sh /path/to/backups
scripts/restore.sh /path/to/backups/librearchive-YYYYMMDDTHHMMSSZ --force
```

Both scripts use `DATABASE_PATH` and `STORAGE_PATH`, falling back to the documented defaults. Each backup contains `database.sqlite`, the complete `storage/` directory, a manifest, and (when `sha256sum` is available) a database checksum. Restore verifies that checksum before replacing the configured persistent data. Database records only store asset filenames, not absolute server paths.

See [PROJECT_SPECIFICATION.md](../PROJECT_SPECIFICATION.md) for the complete product scope and planned features.

See [docs/api.md](docs/api.md) for the current endpoint reference and [.env.example](.env.example) for an environment-variable template.

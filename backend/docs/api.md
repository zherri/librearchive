# LibreArchive API Reference

All protected endpoints require `Authorization: Bearer <access-token>`. Responses and request bodies use JSON unless an endpoint is explicitly described as multipart.

## Authentication and account

- `POST /api/v1/auth/bootstrap` creates the first administrator and only succeeds while no user exists. It accepts `username` and `name`, then returns a one-time generated twelve-word `passphrase`.
- `POST /api/v1/auth/login` accepts `username` and `passphrase`, then returns an access token and the authenticated user.
- `POST /api/v1/auth/refresh` accepts `{ "refreshToken": "..." }` and rotates it, returning a new access token and refresh token.
- `POST /api/v1/auth/logout` revokes the current authenticated session.
- `GET /api/v1/me` returns the current user.
- `PATCH /api/v1/me` updates the current user's name: `{ "name": "..." }`.

## Reader library data

- `GET`, `PUT /api/v1/books/{bookID}/reading-progress`
- `GET /api/v1/reading-progress?status=unread|in_progress|finished&downloaded=true|false`
- `GET /api/v1/favorites`; `PUT`, `DELETE /api/v1/books/{bookID}/favorite`
- `GET`, `POST /api/v1/collections`
- `GET`, `PATCH`, `DELETE /api/v1/collections/{collectionID}`
- `PUT`, `DELETE /api/v1/collections/{collectionID}/books/{bookID}`
- `GET /api/v1/books/{bookID}/annotations?type=highlight|note&color=`
- `POST /api/v1/books/{bookID}/highlights`
- `PATCH`, `DELETE /api/v1/highlights/{highlightID}`
- `POST /api/v1/highlights/{highlightID}/notes`
- `PATCH`, `DELETE /api/v1/notes/{noteID}`

Reading-progress writes accept `status`, `currentPage`, `progressPercent`, and `isDownloaded`. Highlight writes accept `selectedText`, `color`, `page`, optional `startOffset`, `endOffset`, and `chapterRef`. Note writes accept `content`.

## Administrator endpoints

- `GET`, `POST /api/v1/users`
- `PATCH`, `DELETE /api/v1/users/{userID}`
- `POST /api/v1/users/{userID}/reset-passphrase` generates and returns a replacement twelve-word passphrase once.
- `POST /api/v1/books` accepts multipart fields `title`, `author`, PDF `file`, and optional `description`, `language`, `publishedYear`, `pageCount`, and image `cover`.
- `PATCH`, `DELETE /api/v1/books/{bookID}`
- `GET /api/v1/categories`, `GET /api/v1/tags` are available to authenticated readers.
- `POST /api/v1/categories`, `PATCH`, `DELETE /api/v1/categories/{categoryID}` manage categories; the equivalent tag routes use `/api/v1/tags` and `/api/v1/tags/{tagID}`.
- `PUT`, `DELETE /api/v1/books/{bookID}/categories/{categoryID}` and `/api/v1/books/{bookID}/tags/{tagID}` assign catalog metadata.

## Catalog

- `GET /api/v1/books?search=&categoryId=&tagId=&page=&pageSize=`
- `GET /api/v1/books/{bookID}`
- `GET /api/v1/books/{bookID}/file`
- `GET /api/v1/books/{bookID}/cover`

`GET /health` confirms that the HTTP process is running. `GET /ready` also checks database connectivity.

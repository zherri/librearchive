# LibreArchive Product Specification

## 1. Product Vision

LibreArchive is a portable, self-hosted digital-library and reading platform inspired by the core experience of a Kindle. It allows an individual or a family to store, organize, read, and annotate PDF books through a private server and a Flutter mobile application.

The system must be designed for user ownership. A user can run the backend on their own server, keep the database and book assets under their control, and move the complete installation—including the database, PDF files, and cover images—to another server without losing the library or reading data.

The initial implementation priority is the backend API. The API will be written in Go, using Chi for HTTP routing and GORM for database access. All source code, identifiers, API contracts, documentation, and user-facing backend messages must be written in English.

## 2. Goals

- Provide a private library for PDF books.
- Support one administrator and optional family members or other registered users.
- Preserve a user’s reading state, highlights, notes, favorites, collections, and downloaded-book records.
- Provide role-based administration for users and library content.
- Make deployment, backup, restore, and server migration practical for non-commercial self-hosted use.
- Expose a clean API that a Flutter application can consume.

## 3. User Roles and Authorization

### Administrator

Administrators are the only users allowed to manage the catalog and registered users. They can:

- Create, update, deactivate, and remove user accounts.
- Upload PDF books and cover images.
- Create, update, and remove book metadata and book files.
- View and manage the library catalog with pagination.
- Perform the backend operations required by the mobile administration panel.

### Reader

Readers may only access their own library activity and personal data. They can:

- Browse with pagination and search books available to them.
- Read PDF books provided by the library.
- Save reading progress.
- Mark books as favorites (Favorites is a collection).
- Organize books into personal named collections.
- Create, edit, and delete their own highlights and notes.
- Track whether a book is unread, in progress, or finished.
- Record whether a book has been downloaded to their device.

The API must enforce authorization server-side; the mobile application must not be trusted as the authorization boundary.

## 4. Core Domain Requirements

### Books and Library Management

The system must support a complete administrator-managed CRUD lifecycle for books:

- Add a book by uploading a PDF file.
- Store and serve a book cover image.
- Create and update book metadata, including title, description, one or more authors, publisher, publication year, categories or tags, and cover reference. Language is not part of the book model.
- Require an administrator to submit the PDF, metadata, and an optional cover image together through the mobile administration workflow. The API must calculate and persist `pageCount` from the uploaded PDF; clients must not submit or edit it.
- Remove a book and its associated managed assets according to a safe deletion policy.
- List books available to a reader with pagination.
- Search the catalog by relevant metadata, at minimum title and authors.

PDF is the required book format for the first version. The data model should allow additional formats in a future version, but no other reader format is required now.

### Reading State and Progress

Reading data is personal to each reader and book. The API must support:

- A reading status: `unread`, `in_progress`, or `finished`.
- A progress value suitable for restoring a reading session, such as page number, total page count, and/or a normalized percentage.
- The most recent reading position and last-read timestamp.
- Views or filters for unread, in-progress, and finished books with pagination.
- Views or filters for books marked as downloaded on the reader’s device with pagination.

The mobile application determines how a PDF is rendered and where navigation occurs. The API persists the reading position supplied by the client.

### Highlights and Notes

Readers must be able to save text they find meaningful and attach their own notes to it. The API must support:

- Creating, reading, updating, and deleting a reader’s highlights (Reading with pagination).
- Storing the selected text and enough PDF-location information to reopen the selection reliably when supported by the client (for example, page number and selection coordinates or offsets).
- Assigning a color to a highlight.
- Creating, reading, updating, and deleting notes associated with a highlight (Reading with pagination).
- Listing a reader’s highlights and notes for a book with pagination.
- Filtering annotations by type (highlight or note) and highlight color with pagination.

Filtering by PDF chapter or topic is a mobile-reader concern in the first version. The API should remain able to accept chapter/topic references later if the app can extract them, but it is not responsible for parsing a PDF table of contents.

### Favorites and Collections

Readers must be able to:

- Mark and unmark books as favorites.
- List their favorite books (As a collection).
- Create, rename, and delete personal collections.
- Add and remove books from their own named collections.
- List books in a collection with pagination.

Collections and favorite flags are personal; one reader’s organization must not affect another reader’s library.

### User Management

The API must provide administrator-only user management, including:

- Administrator-created user accounts. Public self-registration is not required.
- Authentication with a unique username and an automatically generated twelve-word passphrase. On the first API start with no administrator in the database, the backend must create an active administrator named `admin` with username `admin` and display its generated passphrase once in the server console. Email addresses, email verification, and email-based account recovery are intentionally out of scope.
- The generated passphrase must be shown only once at account creation or administrator-initiated reset. Only a bcrypt hash of the passphrase may be stored in the database.
- Secure session/token handling.
- User profile updates where permitted.
- Account deactivation and removal policies.
- Role assignment, with at least `admin` and `reader` roles.

## 5. Mobile Application Responsibilities

The Flutter application is the client for the API and is responsible for the reading interface. Its first-version responsibilities include:

- Rendering PDF files.
- Displaying a book’s table of contents or chapters and navigating to a selected location.
- Searching inside a downloaded or opened PDF.
- Applying font and font-size preferences where the reader technology supports those settings.
- Providing light and dark reading themes.
- Downloading books for offline use and reporting download state to the API when appropriate.
- Providing the administrator UI for user registration, user management, book upload, and book management. The corresponding authorization and business rules remain in the API.

The backend is not required to render PDFs, extract text, parse chapters, perform in-book search, or control presentation themes and font settings in the first version.

Flutter is the selected client technology. The mobile app should use Dart and Flutter's recommended layered architecture, keeping presentation, application logic, and data access separate by feature. It may use platform plugins for secure credential storage and local PDF files. If the selected PDF reader requires functionality not provided by a plugin, Flutter platform channels or a dedicated plugin may be used to integrate native Android and iOS code.

## 6. API Scope and Technical Direction

The initial backend must be a Go HTTP API built with:

- Go as the implementation language.
- Chi as the router and middleware framework.
- GORM as the ORM and database abstraction layer.
- Migrations made with GORM.
- A relational database, selected with portability in mind.

The API should follow conventional REST-style resource boundaries and return predictable JSON responses. It should include validation, authentication, authorization, error handling, pagination for list endpoints, and audit-friendly timestamps.

Likely top-level resources include:

- Authentication and current user.
- Users.
- Books and managed book assets.
- Reading progress.
- Favorites.
- Collections and collection items.
- Highlights and notes.

API endpoints and database schema will be designed in the next implementation stage; this document defines the product requirements rather than freezing those technical details prematurely.

## 7. Self-Hosting and Portability Requirements

The backend must be deployable on infrastructure controlled by the user. It must avoid requiring a proprietary cloud service for normal operation.

A deployment must keep all persistent state in clearly identified, exportable locations:

- Relational database data.
- Uploaded PDF book files.
- Book cover images.
- Any other user-generated or managed assets.
- Configuration needed to restore the service, excluding secrets that should be recreated securely.

The project must support a documented backup and migration procedure. Moving to a new server should consist of exporting the database, copying persistent asset storage, installing the application, configuring it, and restoring the data. Asset references must remain portable: do not depend on absolute paths or server-specific URLs in database records.

Container-based deployment may be provided, but the application should not depend on containers to run. Persistent files must be stored outside ephemeral application directories.

## 8. Security and Data Ownership Baseline

- Generated passphrases must be hashed with bcrypt; they must never be stored in plain text. The initial administrator passphrase may be displayed once to the server operator during first-start provisioning, but must not be written by application logging thereafter.
- Protected endpoints must require authentication.
- Authorization must be checked for every resource access, especially personal reading data and administrator operations.
- File uploads must validate file type, size, and ownership before being stored or served.
- Download and file-serving endpoints must prevent path traversal and unauthorized access.
- Backups contain private reading data and should be protected by the operator.

## 9. Out of Scope for the First API Iteration

- EPUB, MOBI, and other non-PDF formats.
- Server-side PDF rendering, text extraction, in-book search, or table-of-contents parsing.
- Social reading, public sharing, recommendations, a public marketplace, or DRM.
- Cloud synchronization operated by a third party.
- Billing, subscriptions, or e-commerce.

## 10. Definition of Initial Success

The first backend milestone is successful when an administrator can securely manage users and PDF books; a reader can browse and search the available catalog, download and access books, maintain personal reading progress, favorites, and collections, and create and manage highlights with notes; and the full installation can be backed up and restored on another self-hosted server with its database and media assets intact.

## 11. API Production-Readiness Requirements

The first backend milestone provides the functional domain API. Before declaring a production release, the API must also meet the following operational, security, and quality requirements.

### API Contract and Documentation

- Publish an OpenAPI specification for every supported endpoint, request body, response body, authentication requirement, and error response.
- Provide runnable request examples for administrator and reader workflows.
- Apply consistent pagination, filtering, sorting, and error-response conventions to every collection endpoint. Collection requests use offset pagination with `offset` (default `0`) and `limit` (default `20`, maximum `100`); responses use `{ "items": [...], "offset": 0, "limit": 20, "total": 0 }`. The mobile client must refresh from offset zero and offer progressive loading whenever `offset + limit < total`.
- Document API versioning and compatibility expectations.

### Authentication and Account Security

- Implement access-token renewal and secure token revocation, including logout from an individual device and all devices where applicable.
- Use usernames and automatically generated twelve-word passphrases; do not collect, require, verify, or use email addresses.
- Use role-based mobile refresh-token lifetimes: access tokens expire after one day for every role; administrator refresh tokens expire after one day, while reader refresh tokens expire after 365 days to support low-friction household reading devices. Both tokens must be signed JWTs and refresh tokens must rotate on use. Passphrase resets, deactivation, deletion, and logout must revoke active sessions.
- Normalize a passphrase and pre-hash it with SHA-256 before bcrypt to avoid bcrypt's 72-byte input limit; never persist a generated passphrase in plain text. The backend displays the initial administrator passphrase once in the first-start server console output; it must not emit it in normal application logs.
- Provide an administrator-initiated passphrase reset that generates a new twelve-word passphrase and returns it once to the administrator for secure delivery to the reader.
- Prevent an administrator from accidentally removing or deactivating the final active administrator account.
- Record security-relevant account events without exposing passphrases, tokens, or private book content.

### Book Metadata and Catalog Quality

- Add administrator-managed categories and tags, with book assignment and catalog filtering.
- Support richer optional metadata, including ISBN, publisher, publication date, contributors, and custom metadata where useful.
- Support safe replacement of a PDF file or cover image while preserving the book record and reader data when the administrator chooses to do so.
- Define and document a safe user-deletion policy, including the handling of that user’s reading data and personal annotations.

### Upload and File-Serving Hardening

- Validate uploaded content by inspecting file signatures and MIME types, not only filename extensions.
- Make upload-size limits configurable.
- Generate safe server-side filenames and prohibit path traversal in every asset operation.
- Define an optional malware-scanning integration point for operators who need it.
- Ensure file-serving endpoints authorize every request and set appropriate content headers.

### Reliability and Operations

- Add structured logging with request identifiers and configurable log levels.
- Add graceful HTTP shutdown and database connection cleanup.
- Add configurable CORS rules for the mobile client and administrative interfaces.
- Add rate limiting or equivalent protection for authentication and expensive endpoints.
- Provide health and readiness checks suitable for a reverse proxy or container orchestrator.
- Provide documented, repeatable backup, verification, restore, and migration commands or scripts.
- Provide configuration validation and a production deployment guide, including secret management and persistent-volume requirements.

### Testing and Quality Assurance

- Add unit tests for validation and domain logic.
- Add integration tests for authentication, authorization, ownership isolation, CRUD operations, file uploads, and database migrations.
- Add regression tests for administrator-only operations and for attempts to access another reader’s private data.
- Run formatting, static analysis, and automated tests in continuous integration.

### Explicit Non-Goals for This API

The API remains responsible for persistence, authorization, and secure file access. PDF rendering, in-book text search, chapter extraction, typography settings, and light/dark reading themes remain mobile-client responsibilities unless the project scope is intentionally expanded.

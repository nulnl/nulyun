**Overview**
- **Project**: Nul Yun — a self-hosted file service (backend + SPA) focused on file storage, sharing, previews, and TOTP authentication.
- **Purpose**: This doc helps integrators and client developers understand how to run the service, authenticate, call key APIs (REST, WebDAV, TUS), and build clients.
- **Code entrypoints**: main server startup at [main.go](main.go) and HTTP routing at [http/http.go](http/http.go).

**Quickstart**
- **Build**: `go build -o nulyun .`
- **Run (dev)**: `./nulyun -database ./nulyun.db -address 127.0.0.1 -port 8080`
- **Config file**: copy and edit [config.example.json](config.example.json) and pass with `-config`.

**Auth & Sessions**
- **Auth methods**: pluggable via `auth` implementations. See `auth/` for available strategies (proxy, none, etc.).
- **Login endpoint**: `POST /api/login` — returns JSON `{ "token": "<JWT>", "otp": false }` on success. Token is a JWT.
  - Example:

    curl -X POST -H "Content-Type: application/json" -d '{"username":"admin","password":"secret"}' http://localhost:8080/api/login

- **Sending token**: include JWT in request header `X-Auth: <token>`. For GET requests, a cookie named `auth` is also supported.
- **TOTP**: If enabled, login may return an OTP flow — see `/api/login/otp` in [http/http.go](http/http.go).

**API Overview**
- **Base path**: `/api` (see router in [http/http.go](http/http.go)).
- **Key endpoint groups**:
  - Authentication: `POST /api/login`, `POST /api/renew`, `POST /api/signup`.
  - Users: `/api/users` CRUD.
  - Files/resources: `/api/resources` — list, upload, delete, patch. Use `GET`, `POST`, `PUT`, `DELETE`, `PATCH`.
  - TUS (resumable upload): `/api/tus` — supports `POST` (create), `HEAD`/`GET`, `PATCH`, `DELETE`.
  - Preview & raw: `/api/preview/{size}/{path}` and `/api/raw/...` for binary access.
  - Shares: `/api/shares`, `/api/share`.
  - Settings & usage: `/api/settings`, `/api/usage`.
  - Search & subtitle: `/api/search`, `/api/subtitle`.
  - Public download/share: `/api/public/dl` and `/api/public/share`.

**File Upload: resources vs TUS**
- **Simple resource upload**: `POST /api/resources` accepts multi-part form or JSON payloads depending on client. Use for small files or server-side imports.
- **Resumable upload (TUS)**: Follow TUS protocol at `/api/tus`.
  - Use `tus-js-client` for browser clients, or `curl --request PATCH` for testing. The router handlers are in [http/http.go].

**WebDAV**
- WebDAV endpoints are wired into the mux (see `setupWebDAVRoutes` and `setupWebDAVHandler` in [http/http.go]).
- Mount the server as WebDAV at the configured base path and use standard WebDAV clients.

**Storage & Configuration**
- **Storage drivers**: default uses BoltDB for metadata and local filesystem for file storage. See [storage/bolt](storage/bolt) and [storage/storage.go](storage/storage.go).
- **Server settings**: default settings and global settings live under [settings/global](settings/global). Many runtime options are available via CLI flags (see [main.go](main.go)).

**Client Integration Patterns**
- **Authentication**:
  - Call `POST /api/login` and extract `token`.
  - Add `X-Auth: <token>` to subsequent requests.
- **List files**: call `GET /api/resources` (query parameters for path/offset/limit may apply depending on implementation).
- **Download**: use `GET /api/raw/...` or `GET /api/preview/...` for images/previews.
- **Upload (JS)**: prefer TUS for large files (use `tus-js-client`). For simple uploads, `multipart/form-data` to `/api/resources`.
- **Share creation**: `POST /api/share` with JSON describing targets and permissions.

**Development & Running**
- **Source**: server startup: [main.go](main.go); routing: [http/http.go](http/http.go).
- **Build**: requires Go modules. Run `go mod download` then `go build`.
- **Tests**: run `go test ./...` to execute Go unit tests. Check `files/service_test.go` and `users/storage_test.go` for examples.
- **Assets / Frontend**: built SPA assets are embedded from `www` and served by the handler (see `www` package and `assets.go`). For frontend development, run the SPA build separately in `www/`.

**Examples**
- Login and use token:

  curl -X POST -H "Content-Type: application/json" -d '{"username":"admin","password":"secret"}' http://localhost:8080/api/login

  curl -H "X-Auth: <token>" http://localhost:8080/api/resources

- Upload a file (simple multipart):

  curl -H "X-Auth: <token>" -F "file=@./myfile.txt" http://localhost:8080/api/resources

- TUS quick test (conceptual): use `tus-js-client` or CLI that implements TUS protocol against `http://localhost:8080/api/tus`.

**References**
- Server entrypoint: [main.go](main.go)
- HTTP routing & handlers: [http/http.go](http/http.go)
- Configuration example: [config.example.json](config.example.json)
- Storage drivers: [storage/bolt](storage/bolt)
- Authentication implementations: [auth](auth)

**Next Steps for Integrators**
- Try the Quickstart, get a token via `/api/login`, list and upload files.
- For browser clients implement JWT storage and add `X-Auth` header to API calls.

If you want, I can also:
- Add language-specific snippets (Node/Go/Python) for auth and uploads.
- Generate a minimal SDK example for uploads and downloads.

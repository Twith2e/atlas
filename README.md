# Atlas

Atlas is a Go API built with [Gin](https://github.com/gin-gonic/gin), backed by Postgres (Neon).

## Prerequisites

- Go 1.25+
- A Postgres database (this project targets [Neon](https://neon.tech))
- [`golang-migrate`](https://github.com/golang-migrate/migrate) CLI for running migrations
- [`swag`](https://github.com/swaggo/swag) CLI for regenerating API docs (`go install github.com/swaggo/swag/cmd/swag@latest`)

## Setup

1. Clone the repo and install dependencies:
   ```
   go mod download
   ```
2. Create a `.env` file in the repo root:
   ```
   PG_CONN_STRING=postgresql://user:password@host/dbname?sslmode=require
   PORT=8080
   TERMII_API_KEY=your_termii_api_key
   ACCESS_TOKEN_SECRET=your_access_token_secret
   REFRESH_TOKEN_SECRET=your_refresh_token_secret
   ```
3. Run migrations against your database:
   ```
   migrate -path ./migrations -database "$PG_CONN_STRING" up
   ```

## Running the app

```
go run ./cmd/api
```

The server starts on `PORT` (default `8080`).

## API docs

Swagger UI is served at `http://localhost:8080/swagger/index.html`.

After adding or changing `@`-annotated handler comments, regenerate the spec:
```
swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal
```

Both flags are required: the handlers document their responses as `response.APIResponse[T]`, and `swag` will not resolve that generic envelope without them.

## Auth: client requirements

The refresh token is stored in an `HttpOnly`, `SameSite=Strict` cookie scoped to `/api/v1/auth`. Access tokens live 15 minutes; sessions live 30 days of inactivity.

`POST /api/v1/auth/refresh` **rotates** the refresh token on every call, and revokes the entire session if a token that has already been rotated away is presented (stolen-token detection).

Because of that, clients **must serialize refresh attempts**: keep a single in-flight `/refresh` request and have all concurrent `401`s await it. Firing two refreshes in parallel means the second one presents an already-rotated token, which the server treats as reuse and logs the user out.

Logout is best-effort by design — clear local auth state and redirect regardless of the response status.

## Project structure

- `cmd/api` — application entry point
- `internal/server` — HTTP server and router setup
- `internal/modules/auth` — auth handlers, service, and routes
- `internal/database` — Postgres connection setup
- `internal/config` — environment-based configuration
- `migrations` — SQL migration files (up/down pairs)
- `docs` — Swagger spec

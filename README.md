## Wellness Tracker API (Go, Gin, Gorm, Postgres)

Minimalistic, production-minded HTTP API in Go for a frontend wellness-tracking app. It uses:

- **Gin** for HTTP routing and JSON handling
- **Gorm** with **Postgres** for persistence
- **godotenv** and environment variables for configuration
- **Graceful shutdown** and basic health checks
- **A small OpenAPI document** exposed at `/openapi.json` for documentation and tooling

---

### Project structure

- **`main.go`**: Application entrypoint, HTTP server setup, graceful shutdown
- **`config/`**: Configuration loading from env / `.env`
- **`internal/database/`**: Gorm + Postgres setup and migrations
- **`internal/models/`**: Data models (e.g. `Checkin`)
- **`internal/api/`**: Gin router and HTTP handlers

---

### Prerequisites

- Go 1.22 or newer
- A running Postgres instance

Create a database for the app (example):

```bash
createdb wellness_tracker
```

---

### Configuration

Copy the example environment file and adjust for your environment:

```bash
cp .env.example .env
```

Key variables:

- **`APP_ENV`**: Environment name (`development`, `staging`, `production`, ...)
- **`PORT`**: HTTP port (default: `8080`)
- **`DATABASE_URL`**: Postgres DSN, e.g. `postgres://user:password@localhost:5432/wellness_tracker?sslmode=disable`
- **`HTTP_READ_TIMEOUT`**, **`HTTP_WRITE_TIMEOUT`**, **`HTTP_SHUTDOWN_TIMEOUT`**: Go duration strings like `5s`, `1m`, etc.

`config.Load()` will automatically:

- Load variables from `.env` (if present, via `godotenv`)
- Validate required variables (e.g. `DATABASE_URL`)
- Parse timeout durations with safe fallbacks

---

### Installing dependencies

From the project root:

```bash
go mod tidy
```

This will pull in Gin, Gorm, and other dependencies based on imports.

---

### Running the API

From the project root:

```bash
go run ./...
```

The server will start on `:${PORT}` (default `:8080`).

Health check:

```bash
curl http://localhost:8080/health
```

You should get:

```json
{"status":"ok"}
```

---

### API overview

Base URL (local dev): `http://localhost:8080`

- **Health**
  - `GET /health` – basic service health check

- **Check-ins (v1)**
  - `GET /api/v1/checkins` – list check-ins (supports `page`, `page_size`)
  - `POST /api/v1/checkins` – create a new check-in
  - `GET /api/v1/checkins/:id` – retrieve a single check-in
  - `PATCH /api/v1/checkins/:id` – partially update a check-in
  - `DELETE /api/v1/checkins/:id` – delete a check-in


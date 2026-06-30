# review-api

## Architecture

The backend is structured with clear module/api boundary:
- `cmd/server/main.go` is the composition root only (DB connection, dependency wiring, HTTP server lifecycle).
- `internal/api` owns HTTP concerns: route registration, `/reviews` and `/reviews-async` handlers, JSON responses, and SSE framing.
- `internal/service` coordinates repository reads and feed sync behavior.
- `internal/repository` persists reviews in SQLite via GORM.
- `internal/sync` fetches new feed data and persists it.

## Design decisions and intent

- **JSON feed over XML**: the backend consumes the JSON App Store feed because it is simpler to parse and better aligned with the API payload shape we return.
- **Pagination strategy**: we detect total pages from feed links and iterate pages directly instead of following XML-oriented `next`/`last` links.
- **API-boundary testing first**: higher-value behavior (especially SSE sequencing and failure modes) is validated in `internal/api/api_test.go`; `internal/service/service_test.go` stays focused on service-level logic.
- **Lazy refresh model**: refresh from Apple is triggered by API requests. This keeps the design simple for the homework scope while still demonstrating sync + persistence behavior.
- **Clear composition boundary**: `cmd/server/main.go` stays focused on wiring and lifecycle, while API behavior lives in `internal/api`.

## API endpoints

### `GET /reviews`
Returns all reviews as JSON.

### `GET /reviews-async`
Streams Server-Sent Events (`text/event-stream`) with the canonical event contract:
- `event: data` with payload `{ "reviews": [...] }`
- `event: refresh_error` with payload `{ "error": "..." }`

The stream is designed for two-stage delivery:
1. initial DB snapshot (`data`)
2. refresh result from the RSS feed (`data` or `refresh_error`)

## Testing strategy
Tests are intentionally split by level:
- `internal/api/api_test.go` validates handler behavior and SSE sequencing at the API boundary.
- `internal/service/service_test.go` keeps focused unit coverage for service logic and async result sequencing.

Run all tests:
```bash
go test ./...
```

## Data model
Reviews are stored in SQLite table `reviews_apple` (`internal/repository/schema.go`) with fields:
- `source_id` (primary key)
- `title`
- `author`
- `content`
- `rating`
- `date`

Schema creation/update currently happens through GORM auto-migration at server startup.
This is a simple project, in practice for a real app I would advise managing/running migrations separately through a tool like Flyway.

## Run locally
Prerequisite: Go 1.26+

```bash
go run ./cmd/server
```

The server listens on `:8080` and uses local SQLite file `reviews.db`.

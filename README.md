# runway-itw
Monorepo for the review application:
- `review-api`: Go backend that fetches, stores, and serves Apple reviews.
- `review-ui`: React frontend that consumes the backend stream and renders reviews.

## Getting started
Run backend and frontend in separate terminals.

### 1) Start the API
Prerequisite: Go 1.26+

```bash
cd review-api
go run ./cmd/server
```

The API listens on `http://localhost:8080`.

### 2) Start the UI
Prerequisite: Node.js

```bash
cd review-ui
npm install
npm run dev
```

The Vite dev server proxies `/reviews` and `/reviews-async` to the backend on `localhost:8080`.

## Tests
Backend tests:
```bash
go test -C review-api ./...
```

## Package documentation
- Backend details: `review-api/README.md`
- Frontend details: `review-ui/README.md`

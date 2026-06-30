# review-ui
React + TypeScript + Vite frontend for viewing reviews from `review-api`.

## Design decisions and intent
- **Backend-first scope**: this project prioritizes backend/API work; the UI is intentionally lightweight and primarily a delivery surface for streamed review data.
- **Hook-centered state management**: `src/hooks/useReviewsStream.ts` is a deliberate abstraction to keep SSE parsing, sequencing, and error-state handling isolated from rendering.
- **Single-page composition**: the UI is intentionally not split into many smaller components because current complexity is low and additional decomposition would add ceremony without much payoff.
- 
## What it does

- Connects to `GET /reviews-async` with `EventSource`.
- Handles SSE events:
  - `data` for review payloads
  - `refresh_error` for refresh failures
- Renders loading, hard-error, and soft-warning states.
- 
## Local development
Prerequisite: Node.js
```bash
npm install
npm run dev
```

The dev server runs on Vite defaults and proxies:
- `/reviews` -> `http://localhost:8080`
- `/reviews-async` -> `http://localhost:8080`
- 
Make sure the backend is running first.

## Scripts
- `npm run dev` - start dev server
- `npm run build` - type-check and build
- `npm run lint` - run Oxlint
- `npm run preview` - preview production build

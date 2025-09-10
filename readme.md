# Chirpy

A minimal Twitter-like API with user auth, chirps, and simple sorting.

## Quick Start

- Requirements: Go 1.21+, SQLite (bundled), make (optional)
- Install deps: `go mod download`
- Run server: `go run .`
- Base URL: `http://localhost:8080`

## Endpoints

### Auth

- POST `/api/users`
  - Body: {"email":"string","password":"string"}
  - 201 -> {"id":number,"email":"string","is_chirpy_red":bool}

- POST `/api/login`
  - Body: {"email":"string","password":"string"}
  - 200 -> {"token":"JWT access token"}

- POST `/api/refresh`
  - Body: {"token":"refresh_token"}
  - 200 -> {"token":"new access JWT"}

Authorization: `Authorization: Bearer <access_token>` for protected routes.

### Chirps

- POST `/api/chirps`
  - Auth required
  - Body: {"body":"string (<= 280 chars)"}
  - 201 -> {"id":number,"author_id":number,"body":"string","created_at":"RFC3339"}

- GET `/api/chirps`
  - Query params:
    - `author_id` (optional, number): filter by author
    - `sort` (optional, "asc" | "desc", default "asc"): sort by `created_at`
  - 200 -> [
      {"id":number,"author_id":number,"body":"string","created_at":"RFC3339"},
      ...
    ]
  - Examples:
    - `/api/chirps`
    - `/api/chirps?sort=asc`
    - `/api/chirps?sort=desc`
    - `/api/chirps?author_id=12&sort=desc`

- GET `/api/chirps/{id}`
  - 200 -> {"id":number,"author_id":number,"body":"string","created_at":"RFC3339"}
  - 404 if not found

- DELETE `/api/chirps/{id}`
  - Auth required (must be author)
  - 204 on success

### Webhooks

- POST `/api/polka/webhooks`
  - Handles payment events (sets `is_chirpy_red` on users).

## Errors

JSON error shape:
- 400 -> {"error":"message"}
- 401 -> {"error":"unauthorized"}
- 404 -> {"error":"not found"}

## Notes on Sorting

- `sort` defaults to `"asc"` if missing or empty.
- `"asc"`: oldest first by `created_at`.
- `"desc"`: newest first.
# go-url-shortener

[![CI](https://github.com/Rafiur/go-url-shortener/actions/workflows/ci.yml/badge.svg)](https://github.com/Rafiur/go-url-shortener/actions/workflows/ci.yml)

**[Live demo](https://go-url-shortener-33xx.onrender.com)** — hosted free on
Render, so the first request after a quiet spell takes ~30s while the instance
wakes up.

A URL shortener written in Go, structured with clean architecture. Paste a long
URL, get a short one back; hitting the short link 302s to the original.

**Stack:** Go 1.23 · [Echo](https://echo.labstack.com) · PostgreSQL ([Bun](https://bun.uptrace.dev)) · Redis

## Design

The layers are wired one direction only — handlers depend on usecases, usecases
depend on repository *interfaces*, and the concrete Postgres/Redis
implementations are injected at startup in `main.go`. Swapping a datastore means
writing one new repository, not touching the business logic.

```
internal/
  delivery/handler/     HTTP handlers — bind, call a usecase, render a response
  usecase/              validation + short-code generation; datastore-agnostic
  domain/entity/        the URL and Request types
  infrastructure/
    repository/         repository interfaces (the contract usecases depend on)
    repo_postgres/      Bun implementation
    repo_redis/         go-redis implementation
    schema/             DB row model + mapping to entities
  config/               env config, Postgres/Redis connection setup, migration
web/                    single-file frontend, embedded into the binary
```

Reads on the redirect path are **cache-aside**: Redis is checked first, and on a
miss Postgres is queried and the result written back to Redis with a 24h TTL, so
repeat hits on a popular link never touch the database.

The cache is an optimisation, not a dependency. If Redis is unreachable at
startup the app logs a warning and runs without it — links still resolve from
Postgres, `/healthz` reports `"cache": "disabled"`, and only the cache-only
`/redis` endpoints return `503`.

Short codes are 7 characters from `crypto/rand`, base64url-encoded. A generated
code that collides with an existing one is retried with a fresh code (up to 4
attempts); a *caller-supplied* alias that clashes returns `409` instead, since
silently substituting someone's chosen alias would be worse than failing.

Every redirect increments a click counter. The write is issued from a detached
goroutine so it never delays the 302, and the increment happens in SQL
(`clicks = clicks + 1`) so concurrent redirects cannot lose a count.

Link creation is rate limited per client IP; reads and redirects are not.

## Testing

```bash
go test ./...
```

The repository interfaces exist so the business logic can be tested without a
database: the usecase and handler suites run entirely against in-memory fakes,
covering collision retries, taken aliases, cache hit/miss/backfill, and
degraded-mode behaviour. No container or network required.

## API

| Method | Path               | Description                                        |
| ------ | ------------------ | -------------------------------------------------- |
| `GET`  | `/`                | the frontend (embedded in the binary via `go:embed`) |
| `GET`  | `/healthz`         | liveness probe; reports whether the cache is live  |
| `GET`  | `/:shortcode`      | resolve and 302 — Redis first, Postgres on a miss  |
| `GET`  | `/stats/:shortcode`| click count and metadata for a link                |
| `POST` | `/pg`              | create a link, stored in Postgres                  |
| `GET`  | `/pg?shortcode=`   | look up a link in Postgres                         |
| `GET`  | `/pg/:shortcode`   | resolve and 302 from Postgres                      |
| `POST` | `/redis`           | create a link, stored in Redis                     |
| `GET`  | `/redis?shortcode=`| look up a link in Redis                            |
| `GET`  | `/redis/:shortcode`| resolve and 302 from Redis                         |

The `/pg` and `/redis` groups expose each datastore directly — useful for
comparing them side by side. The root `/:shortcode` route is the one real links
use.

```bash
curl -X POST https://go-url-shortener-33xx.onrender.com/pg \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://example.com/a/very/long/path","short":""}'
```

Swap the host for `http://localhost:8080` when running locally. Leave `short`
empty for a generated code, or set it to claim a custom alias — a taken alias
comes back as `409`.

```json
{
  "success": true,
  "data": {
    "id": 1,
    "short_code": "rdQfuhr",
    "original_url": "https://example.com/a/very/long/path",
    "clicks": 0,
    "created_at": "2026-08-25T18:40:53.874151Z"
  },
  "message": "Successfully Created PostgresURL"
}
```

## Running locally

```bash
cp config.env.example config.env
docker compose up -d          # Postgres on :5434, Redis on :6379
go run .                      # http://localhost:8080
```

The `urls` table is created on startup, so there is no separate migration step.

## Configuration

Read from `config.env` if present, otherwise from the process environment — so
local dev uses the file and a deployed instance uses injected env vars.
`config.env` is gitignored; see `config.env.example`.

| Variable          | Default   | Notes                                                  |
| ----------------- | --------- | ------------------------------------------------------ |
| `PORT`            | `8080`    | injected by most container hosts                       |
| `DBHOST` `DBPORT` `DBUSER` `DBPASS` `DBNAME` `DBSCHEMA` | — | Postgres connection |
| `DBSSLMODE`       | `disable` | must be `require` for managed Postgres (Neon)          |
| `REDIS_ADDRESS`   | —         | `host:port`                                            |
| `REDIS_PASSWORD`  | empty     | required by managed Redis (Upstash)                    |
| `REDIS_TLS`       | `false`   | must be `true` for managed Redis (Upstash)             |
| `ALLOWED_ORIGINS` | any       | comma-separated CORS allowlist                         |
| `CREATE_RATE_LIMIT` | `5`     | sustained link creations per second, per IP            |
| `CREATE_RATE_BURST` | `10`    | spike allowed above that rate                          |
| `DEBUG`           | `false`   | logs every SQL statement                               |

## Deploying

A `Dockerfile` builds a static binary into an Alpine image, and `render.yaml`
describes the service. The frontend is compiled into the binary, so one
container serves both the UI and the API.

The live instance runs on Render's free plan against Neon (Postgres) and Upstash
(Redis). To stand up your own:

1. **Postgres** — create a project on [Neon](https://neon.tech), take
   `DBHOST` / `DBUSER` / `DBPASS` / `DBNAME` from the connection string, set
   `DBPORT=5432` and `DBSSLMODE=require`.
2. **Redis** — create a database on [Upstash](https://upstash.com), set
   `REDIS_ADDRESS` to `<endpoint>:6379`, `REDIS_PASSWORD` to the token, and
   `REDIS_TLS=true`.
3. **App** — point [Render](https://render.com) at this repo; it picks up
   `render.yaml` and prompts for the secrets above.

### GitHub Pages (optional second front door)

A shortener needs a server to resolve `/:shortcode` and issue the 302, so Pages
cannot host it on its own — but it can serve the UI against the deployed API.
`.github/workflows/pages.yml` publishes `web/index.html`, rewriting its
`FALLBACK_API` constant to the repo variable `API_BASE_URL`.

1. Set `API_BASE_URL` to the Render URL under
   *Settings → Secrets and variables → Actions → Variables*.
2. Enable Pages with source *GitHub Actions* under *Settings → Pages*.
3. Set `ALLOWED_ORIGINS` on Render to your Pages origin, e.g.
   `https://<user>.github.io`.

The links the page produces still point at the backend, since that is what
performs the redirect.

## Possible next steps

- Expiring links and a delete endpoint
- Per-link referrer and user-agent breakdown
- Batch shortening
- Structured logging (the `Logger` config block is defined but unused)
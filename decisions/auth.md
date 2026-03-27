# Authentication

## What

Two authentication mechanisms, both accepted by the same middleware on every protected endpoint:

| Client | Mechanism | Header / Cookie |
|---|---|---|
| Browser (frontend) | httpOnly JWT cookie | Cookie: `session=<jwt>` |
| Console / programmatic | API key | `Authorization: Bearer <key>` |

## Protected endpoints

All endpoints except:
- `GET /health` — infrastructure health check
- `POST /auth/login` — issues the JWT cookie
- `POST /auth/logout` — clears the JWT cookie

## JWT design

- Algorithm: HS256
- Secret: 32 random bytes generated at startup (`crypto/rand`)
- Expiry: 1 hour
- Claims: `sub=admin`, `iat`, `exp`
- Stored in: httpOnly cookie named `session`, `SameSite=Lax`, `Path=/`

**Why auto-generated secret:** Zero extra config, tokens always expire within 1h anyway, restart-on-deploy is acceptable for a single-user secretary app.

## API key design

- Single key configured via `API_KEY` env var on the backend
- Injected by the console via `MEETING_API_TOKEN` env var (same value)
- Passed as `Authorization: Bearer <key>`
- No expiry — rotate by changing the env var and redeploying

## CORS

CORS middleware echoes the request `Origin` header (instead of `*`) to allow `Access-Control-Allow-Credentials: true`. This is equivalent in permissiveness to the previous `*` setup and is required for cookies to work in cross-origin dev mode.

## Env vars

| Var | Service | Purpose |
|---|---|---|
| `ADMIN_PASSWORD` | backend | Password for the `admin` user (required) |
| `API_KEY` | backend | Valid API key for programmatic access (required) |
| `MEETING_API_TOKEN` | console | Must match `API_KEY` on the backend (required) |

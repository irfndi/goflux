# GoFlux Telemetry Worker

Cloudflare Worker that receives opt-in anonymized telemetry from GoFlux library users.

## Endpoints

- `POST /v1/telemetry` — Submit telemetry payload
- `GET /v1/health` — Health check
- `GET /v1/stats/summary?hours=24` — Summary stats (requires auth)
- `GET /v1/stats/errors?hours=24` — Error breakdown (requires auth)
- `GET /v1/stats/usage?hours=24` — Usage breakdown (requires auth)

## Data Storage

Uses **Cloudflare D1** (SQLite-compatible edge database) for structured storage and easy SQL querying.

## Setup

### 1. Install dependencies

```bash
pnpm install
```

### 2. Create the D1 database

```bash
pnpm run db:create
```

Copy the `database_id` from the output into `wrangler.toml`:

```toml
[[d1_databases]]
binding = "DB"
database_name = "goflux_telemetry"
database_id = "your-database-id-here"
```

### 3. Apply schema

```bash
pnpm run db:migrate
```

For local development:
```bash
pnpm run db:migrate:local
```

### 4. Set secrets

```bash
npx wrangler secret put TELEMETRY_TOKEN
```

### 5. Deploy

```bash
pnpm run deploy
```

## Payload Schema

```json
{
  "v": 1,
  "ts": 1717200000000,
  "lib_version": "0.1.0",
  "go_version": "go1.21.0",
  "os": "linux",
  "arch": "amd64",
  "type": "error",
  "feature": "EMA",
  "error_type": "*errors.errorString",
  "error_hash": "a1b2c3d4",
  "context": {"window": "10"}
}
```

## Querying Data

You can query D1 directly via Wrangler:

```bash
npx wrangler d1 execute goflux_telemetry --command="SELECT * FROM telemetry_events ORDER BY received_at DESC LIMIT 10"
```

Or use the stats API:

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  https://goflux-telemetry.your-account.workers.dev/v1/stats/summary?hours=24
```

## Privacy

- No IP addresses are stored; they are hashed before processing.
- Error messages are hashed client-side; only the hash and error type are sent.
- Users must explicitly opt in via `telemetry.Enable(endpoint, token)`.

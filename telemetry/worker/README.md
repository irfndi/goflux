# GoFlux Telemetry Worker

Cloudflare Worker that receives opt-in anonymized telemetry from GoFlux library users.

## Endpoints

- `POST /v1/telemetry` — Submit telemetry payload
- `GET /v1/health` — Health check

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

## Deployment

1. Install dependencies:
   ```bash
   npm install
   ```

2. Configure secrets in Cloudflare dashboard:
   - `TELEMETRY_TOKEN` — optional Bearer token for basic auth
   - `ANALYTICS_DATASET` — optional Analytics Engine dataset name

3. Deploy:
   ```bash
   npx wrangler deploy
   ```

## Privacy

- No IP addresses are stored; they are hashed before processing.
- Error messages are hashed client-side; only the hash and error type are sent.
- Users must explicitly opt in via `telemetry.Enable(endpoint, token)`.

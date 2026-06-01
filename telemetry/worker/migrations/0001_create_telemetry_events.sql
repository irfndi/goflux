-- D1 schema for goflux telemetry
-- Run: wrangler d1 execute goflux_telemetry --file=./schema.sql

CREATE TABLE IF NOT EXISTS telemetry_events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  received_at INTEGER NOT NULL,
  ts INTEGER NOT NULL,
  lib_version TEXT NOT NULL,
  go_version TEXT NOT NULL,
  os TEXT NOT NULL,
  arch TEXT NOT NULL,
  type TEXT NOT NULL CHECK(type IN ('error', 'usage')),
  feature TEXT,
  error_type TEXT,
  error_hash TEXT,
  ip_hash TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_type ON telemetry_events(type);
CREATE INDEX IF NOT EXISTS idx_received_at ON telemetry_events(received_at);
CREATE INDEX IF NOT EXISTS idx_lib_version ON telemetry_events(lib_version);
CREATE INDEX IF NOT EXISTS idx_error_hash ON telemetry_events(error_hash);
CREATE INDEX IF NOT EXISTS idx_feature ON telemetry_events(feature);

export interface TelemetryPayload {
  v: number;
  ts: number;
  lib_version: string;
  go_version: string;
  os: string;
  arch: string;
  type: "error" | "usage";
  feature?: string;
  error_type?: string;
  error_hash?: string;
  context?: Record<string, string>;
}

export interface Env {
  TELEMETRY_TOKEN: string;
  DB: D1Database;
}

const MAX_BODY_BYTES = 64 * 1024;
const MAX_CONTEXT_ENTRIES = 32;
const MAX_CONTEXT_VALUE_LENGTH = 256;

function hashIP(ip: string): string {
  let hash = 5381;
  for (let i = 0; i < ip.length; i++) {
    hash = ((hash << 5) + hash) + ip.charCodeAt(i);
  }
  return (hash >>> 0).toString(16);
}

function jsonResponse(data: unknown, status = 200): Response {
  return new Response(JSON.stringify(data), {
    status,
    headers: {
      "Content-Type": "application/json",
      "Access-Control-Allow-Origin": "*",
      "Access-Control-Allow-Methods": "GET, POST, OPTIONS",
    },
  });
}

async function storeEvent(db: D1Database, payload: TelemetryPayload, ipHash: string): Promise<void> {
  await db
    .prepare(
      `INSERT INTO telemetry_events
       (received_at, ts, lib_version, go_version, os, arch, type, feature, error_type, error_hash, context_json, ip_hash)
       VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
    )
    .bind(
      Date.now(),
      payload.ts,
      payload.lib_version,
      payload.go_version,
      payload.os,
      payload.arch,
      payload.type,
      payload.feature ?? null,
      payload.error_type ?? null,
      payload.error_hash ?? null,
      payload.context ? JSON.stringify(payload.context) : null,
      ipHash
    )
    .run();
}

async function getSummary(db: D1Database, hours = 24): Promise<unknown> {
  const since = Date.now() - hours * 60 * 60 * 1000;

  const total = await db
    .prepare("SELECT COUNT(*) as count FROM telemetry_events WHERE received_at > ?")
    .bind(since)
    .first<{ count: number }>();

  const errors = await db
    .prepare("SELECT COUNT(*) as count FROM telemetry_events WHERE type = 'error' AND received_at > ?")
    .bind(since)
    .first<{ count: number }>();

  const usage = await db
    .prepare("SELECT COUNT(*) as count FROM telemetry_events WHERE type = 'usage' AND received_at > ?")
    .bind(since)
    .first<{ count: number }>();

  const topFeatures = await db
    .prepare(
      `SELECT feature, COUNT(*) as count
       FROM telemetry_events
       WHERE type = 'usage' AND received_at > ? AND feature IS NOT NULL
       GROUP BY feature
       ORDER BY count DESC
       LIMIT 10`
    )
    .bind(since)
    .all<{ feature: string; count: number }>();

  const topErrors = await db
    .prepare(
      `SELECT error_type, error_hash, COUNT(*) as count
       FROM telemetry_events
       WHERE type = 'error' AND received_at > ? AND error_type IS NOT NULL
       GROUP BY error_type, error_hash
       ORDER BY count DESC
       LIMIT 10`
    )
    .bind(since)
    .all<{ error_type: string; error_hash: string; count: number }>();

  const libVersions = await db
    .prepare(
      `SELECT lib_version, COUNT(*) as count
       FROM telemetry_events
       WHERE received_at > ?
       GROUP BY lib_version
       ORDER BY count DESC
       LIMIT 5`
    )
    .bind(since)
    .all<{ lib_version: string; count: number }>();

  return {
    period_hours: hours,
    total_events: total?.count ?? 0,
    error_events: errors?.count ?? 0,
    usage_events: usage?.count ?? 0,
    top_features: topFeatures.results ?? [],
    top_errors: topErrors.results ?? [],
    lib_versions: libVersions.results ?? [],
  };
}

async function getErrors(db: D1Database, hours = 24, limit = 50): Promise<unknown> {
  const since = Date.now() - hours * 60 * 60 * 1000;

  const results = await db
    .prepare(
      `SELECT error_type, error_hash, COUNT(*) as count, MAX(received_at) as last_seen
       FROM telemetry_events
       WHERE type = 'error' AND received_at > ? AND error_type IS NOT NULL
       GROUP BY error_type, error_hash
       ORDER BY count DESC
       LIMIT ?`
    )
    .bind(since, limit)
    .all<{ error_type: string; error_hash: string; count: number; last_seen: number }>();

  return {
    period_hours: hours,
    errors: results.results ?? [],
  };
}

async function getUsage(db: D1Database, hours = 24, limit = 50): Promise<unknown> {
  const since = Date.now() - hours * 60 * 60 * 1000;

  const results = await db
    .prepare(
      `SELECT feature, COUNT(*) as count, MAX(received_at) as last_seen
       FROM telemetry_events
       WHERE type = 'usage' AND received_at > ? AND feature IS NOT NULL
       GROUP BY feature
       ORDER BY count DESC
       LIMIT ?`
    )
    .bind(since, limit)
    .all<{ feature: string; count: number; last_seen: number }>();

  return {
    period_hours: hours,
    usage: results.results ?? [],
  };
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    if (request.method === "OPTIONS") {
      return new Response(null, {
        status: 204,
        headers: {
          "Access-Control-Allow-Origin": "*",
          "Access-Control-Allow-Methods": "GET, POST, OPTIONS",
          "Access-Control-Allow-Headers": "Content-Type, Authorization",
        },
      });
    }

    const url = new URL(request.url);
    const clientIP = request.headers.get("cf-connecting-ip") || "unknown";
    const hashedIP = hashIP(clientIP);

    // Health check
    if (url.pathname === "/v1/health") {
      return jsonResponse({ status: "ok", db: !!env.DB });
    }

    // Submit telemetry
    if (url.pathname === "/v1/telemetry" && request.method === "POST") {
      if (!env.TELEMETRY_TOKEN) {
        return jsonResponse({ error: "telemetry authentication is not configured" }, 503);
      }
      const auth = request.headers.get("Authorization");
      if (auth !== `Bearer ${env.TELEMETRY_TOKEN}`) {
        return jsonResponse({ error: "unauthorized" }, 401);
      }
      const contentLength = Number(request.headers.get("Content-Length") ?? 0);
      if (contentLength > MAX_BODY_BYTES) {
        return jsonResponse({ error: "payload too large" }, 413);
      }

      try {
        const body = await request.text();
        if (new TextEncoder().encode(body).byteLength > MAX_BODY_BYTES) {
          return jsonResponse({ error: "payload too large" }, 413);
        }
        const payload = JSON.parse(body) as TelemetryPayload;

        if (!isValidPayload(payload)) {
          return jsonResponse({ error: "missing required fields" }, 400);
        }

        await storeEvent(env.DB, payload, hashedIP);

        return jsonResponse({ ok: true }, 202);
      } catch {
        return jsonResponse({ error: "invalid json" }, 400);
      }
    }

    // Stats endpoints require the same token as ingestion.
    if (!env.TELEMETRY_TOKEN) {
      return jsonResponse({ error: "telemetry authentication is not configured" }, 503);
    }
    const auth = request.headers.get("Authorization");
    if (auth !== `Bearer ${env.TELEMETRY_TOKEN}`) {
      return jsonResponse({ error: "unauthorized" }, 401);
    }

    const hours = boundedInteger(url.searchParams.get("hours"), 24, 1, 24 * 365);

    if (url.pathname === "/v1/stats/summary") {
      const summary = await getSummary(env.DB, hours);
      return jsonResponse(summary);
    }

    if (url.pathname === "/v1/stats/errors") {
      const errors = await getErrors(env.DB, hours);
      return jsonResponse(errors);
    }

    if (url.pathname === "/v1/stats/usage") {
      const usage = await getUsage(env.DB, hours);
      return jsonResponse(usage);
    }

    return jsonResponse({ error: "not found" }, 404);
  },
};

function isValidPayload(payload: TelemetryPayload): boolean {
  if (!payload || typeof payload !== "object" || !Number.isInteger(payload.v) || payload.v < 1 || !Number.isFinite(payload.ts)) {
    return false;
  }
  if (
    typeof payload.lib_version !== "string" ||
    payload.lib_version.length === 0 ||
    payload.lib_version.length > 64 ||
    typeof payload.go_version !== "string" ||
    payload.go_version.length === 0 ||
    payload.go_version.length > 64
  ) {
    return false;
  }
  if (
    typeof payload.os !== "string" ||
    payload.os.length === 0 ||
    payload.os.length > 32 ||
    typeof payload.arch !== "string" ||
    payload.arch.length === 0 ||
    payload.arch.length > 32
  ) {
    return false;
  }
  if (payload.type !== "error" && payload.type !== "usage") {
    return false;
  }
  if (payload.feature !== undefined && (typeof payload.feature !== "string" || payload.feature.length > 128)) {
    return false;
  }
  if (payload.error_type !== undefined && (typeof payload.error_type !== "string" || payload.error_type.length > 256)) {
    return false;
  }
  if (payload.error_hash !== undefined && (typeof payload.error_hash !== "string" || payload.error_hash.length > 128)) {
    return false;
  }
  if (
    payload.context !== undefined &&
    (payload.context === null || typeof payload.context !== "object" || Array.isArray(payload.context))
  ) {
    return false;
  }
  if (payload.context && Object.keys(payload.context).length > MAX_CONTEXT_ENTRIES) {
    return false;
  }
  return !payload.context || Object.entries(payload.context).every(
    ([key, value]) => typeof value === "string" && key.length <= 64 && value.length <= MAX_CONTEXT_VALUE_LENGTH,
  );
}

function boundedInteger(value: string | null, fallback: number, min: number, max: number): number {
  const parsed = Number.parseInt(value ?? "", 10);
  if (!Number.isFinite(parsed)) {
    return fallback;
  }
  return Math.min(Math.max(parsed, min), max);
}

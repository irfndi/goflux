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
  TELEMETRY_TOKEN?: string;
  DB: D1Database;
}

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
       (received_at, ts, lib_version, go_version, os, arch, type, feature, error_type, error_hash, ip_hash)
       VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
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
      if (env.TELEMETRY_TOKEN) {
        const auth = request.headers.get("Authorization");
        if (auth !== `Bearer ${env.TELEMETRY_TOKEN}`) {
          return jsonResponse({ error: "unauthorized" }, 401);
        }
      }

      try {
        const payload = (await request.json()) as TelemetryPayload;

        if (!payload.v || !payload.ts || !payload.lib_version || !payload.go_version || !payload.type) {
          return jsonResponse({ error: "missing required fields" }, 400);
        }

        await storeEvent(env.DB, payload, hashedIP);

        return jsonResponse({ ok: true }, 202);
      } catch (e) {
        return jsonResponse({ error: "invalid json" }, 400);
      }
    }

    // Stats endpoints (optional token auth)
    if (env.TELEMETRY_TOKEN) {
      const auth = request.headers.get("Authorization");
      if (auth !== `Bearer ${env.TELEMETRY_TOKEN}`) {
        return jsonResponse({ error: "unauthorized" }, 401);
      }
    }

    const hours = parseInt(url.searchParams.get("hours") || "24", 10);

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

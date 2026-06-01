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
  ANALYTICS_DATASET?: string;
}

function hashIP(ip: string): string {
  // Simple djb2 hash to anonymize IP addresses
  let hash = 5381;
  for (let i = 0; i < ip.length; i++) {
    hash = ((hash << 5) + hash) + ip.charCodeAt(i);
  }
  return (hash >>> 0).toString(16);
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url);
    const clientIP = request.headers.get("cf-connecting-ip") || "unknown";
    const hashedIP = hashIP(clientIP);

    if (url.pathname === "/v1/health") {
      return new Response(JSON.stringify({ status: "ok" }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      });
    }

    if (url.pathname === "/v1/telemetry" && request.method === "POST") {
      // Optional token check for basic auth
      if (env.TELEMETRY_TOKEN) {
        const auth = request.headers.get("Authorization");
        if (auth !== `Bearer ${env.TELEMETRY_TOKEN}`) {
          return new Response(JSON.stringify({ error: "unauthorized" }), {
            status: 401,
            headers: { "Content-Type": "application/json" },
          });
        }
      }

      try {
        const payload = (await request.json()) as TelemetryPayload;

        // Validate required fields
        if (!payload.v || !payload.ts || !payload.lib_version || !payload.go_version || !payload.type) {
          return new Response(JSON.stringify({ error: "missing required fields" }), {
            status: 400,
            headers: { "Content-Type": "application/json" },
          });
        }

        // Enrich with hashed IP (for rate limiting / abuse detection without tracking identity)
        const enriched = {
          ...payload,
          ip_hash: hashedIP,
          received_at: Date.now(),
        };

        // Write to Analytics Engine if configured
        if (env.ANALYTICS_DATASET && (request as any).cf?.writeDataPoint) {
          (request as any).cf.writeDataPoint({
            blobs: [
              enriched.lib_version,
              enriched.go_version,
              enriched.os,
              enriched.arch,
              enriched.type,
              enriched.feature || "",
              enriched.error_type || "",
              enriched.error_hash || "",
              enriched.ip_hash,
            ],
            doubles: [enriched.ts, enriched.received_at],
            indexes: [enriched.type, enriched.lib_version],
          });
        }

        // Also log to console for debugging / alternative ingestion
        console.log(JSON.stringify(enriched));

        return new Response(JSON.stringify({ ok: true }), {
          status: 202,
          headers: { "Content-Type": "application/json" },
        });
      } catch (e) {
        return new Response(JSON.stringify({ error: "invalid json" }), {
          status: 400,
          headers: { "Content-Type": "application/json" },
        });
      }
    }

    return new Response(JSON.stringify({ error: "not found" }), {
      status: 404,
      headers: { "Content-Type": "application/json" },
    });
  },
};

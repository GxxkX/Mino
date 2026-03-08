import { NextRequest, NextResponse } from 'next/server';

/**
 * Internal proxy for MinIO audio files.
 * Usage: GET /api/minio?url=<encoded-minio-url>
 *
 * This keeps MinIO credentials and internal network addresses
 * off the client, and avoids CORS / mixed-content issues.
 */
export async function GET(req: NextRequest) {
  const raw = req.nextUrl.searchParams.get('url');
  if (!raw) {
    return NextResponse.json({ error: 'Missing url parameter' }, { status: 400 });
  }

  let target: URL;
  try {
    target = new URL(raw);
  } catch {
    return NextResponse.json({ error: 'Invalid url parameter' }, { status: 400 });
  }

  // SSRF guard: only allow requests to the configured MinIO host.
  // MINIO_INTERNAL_HOST should be the MinIO endpoint hostname (e.g. "minio" or "minio:9000").
  const allowedHost = process.env.MINIO_INTERNAL_HOST || '';
  if (allowedHost) {
    try {
      // Strip port from allowedHost for hostname comparison
      const allowedHostname = allowedHost.split(':')[0];
      if (target.hostname !== allowedHostname) {
        return NextResponse.json({ error: 'Forbidden host' }, { status: 403 });
      }
    } catch {
      // If parsing fails, skip the check
    }
  }

  try {
    const upstream = await fetch(target.toString(), {
      headers: {
        // Forward Range header to support audio seeking
        ...(req.headers.get('range') ? { Range: req.headers.get('range')! } : {}),
      },
    });

    const headers = new Headers();
    const forward = ['content-type', 'content-length', 'content-range', 'accept-ranges', 'last-modified', 'etag'];
    for (const h of forward) {
      const v = upstream.headers.get(h);
      if (v) headers.set(h, v);
    }
    // Allow browser to cache audio
    headers.set('cache-control', 'private, max-age=3600');

    return new NextResponse(upstream.body, {
      status: upstream.status,
      headers,
    });
  } catch (err) {
    console.error('[minio-proxy] fetch error:', err);
    return NextResponse.json({ error: 'Upstream fetch failed' }, { status: 502 });
  }
}

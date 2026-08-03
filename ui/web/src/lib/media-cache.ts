/** Client-side blob cache for media URLs.
 *  Prevents image flickering when switching chat sessions —
 *  backend generates fresh ?ft= tokens on each chat.history call,
 *  so the same file gets a different URL each time.
 *  Cache key = clean path (without ?ft=), value = blob ObjectURL with 5-min TTL. */

const CACHE_TTL_MS = 5 * 60 * 1000; // 5 minutes, matches backend FileTokenTTL

interface CacheEntry {
  objectUrl: string;
  expiresAt: number;
}

const cache = new Map<string, CacheEntry>();
const inflight = new Map<string, Promise<string>>();

/** Strip ?ft=... or &ft=... from URL, return clean path. */
export function stripFt(url: string): string {
  return url.replace(/[?&]ft=[^&\s)"'<>]*/g, "").replace(/[?&]$/, "");
}

/** Evict expired entries and revoke their ObjectURLs. */
function evictExpired() {
  const now = Date.now();
  for (const [key, entry] of cache) {
    if (entry.expiresAt <= now) {
      URL.revokeObjectURL(entry.objectUrl);
      cache.delete(key);
    }
  }
}

/** Whether a URL is an absolute external URL (not same-origin, not a /v1/files/ path).
 *  Such URLs are rendered directly by the browser — they must not go through the
 *  blob fetch path (CSP connect-src blocks cross-origin fetch, and caching remote
 *  blobs is unnecessary). */
function isExternalUrl(url: string): boolean {
  return /^https?:\/\//i.test(url) && !url.includes("/v1/files/");
}

/** Get cached ObjectURL or fetch blob and cache it.
 *  Returns the blob ObjectURL on success, or the original signed URL on failure. */
export async function getOrFetchUrl(signedUrl: string): Promise<string> {
  // External absolute URLs (e.g. OSS image host) bypass the blob cache —
  // CSP connect-src forbids the cross-origin fetch, and the <img> loads the
  // URL directly once img-src allows the host.
  if (isExternalUrl(signedUrl)) return signedUrl;

  evictExpired();

  const key = stripFt(signedUrl);
  const cached = cache.get(key);
  if (cached && cached.expiresAt > Date.now()) {
    return cached.objectUrl;
  }

  // Deduplicate concurrent fetches for the same key
  const existing = inflight.get(key);
  if (existing) return existing;

  const promise = (async () => {
    try {
      const res = await fetch(signedUrl);
      if (!res.ok) return signedUrl;
      const blob = await res.blob();
      const objectUrl = URL.createObjectURL(blob);
      cache.set(key, { objectUrl, expiresAt: Date.now() + CACHE_TTL_MS });
      return objectUrl;
    } catch {
      return signedUrl; // graceful degradation
    } finally {
      inflight.delete(key);
    }
  })();

  inflight.set(key, promise);
  return promise;
}

/** Revoke all cached ObjectURLs (call on logout/disconnect). */
export function revokeAll() {
  for (const entry of cache.values()) {
    URL.revokeObjectURL(entry.objectUrl);
  }
  cache.clear();
  inflight.clear();
}

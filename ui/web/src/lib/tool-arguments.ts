/**
 * Tool argument normalization.
 *
 * eino streams tool-call `arguments` as a JSON string (e.g. `"{\"name\":\"x\"}"`),
 * but the UI wants a parsed object to read individual fields. This helper
 * accepts either form and returns a parsed object, or undefined when there is
 * nothing usable. Used by both the live tool-call stream and history adapter.
 */
export function normalizeToolArguments(
  raw: unknown,
): Record<string, unknown> | undefined {
  if (!raw) return undefined;
  if (typeof raw === "object") return raw as Record<string, unknown>;
  if (typeof raw === "string") {
    const trimmed = raw.trim();
    if (!trimmed) return undefined;
    try {
      const parsed = JSON.parse(trimmed);
      if (parsed && typeof parsed === "object" && !Array.isArray(parsed)) {
        return parsed as Record<string, unknown>;
      }
    } catch {
      return undefined;
    }
  }
  return undefined;
}

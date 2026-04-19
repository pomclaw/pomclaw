// Stub file - agents utils
export function agentLogoUrl(basePath: string) {
  return `${basePath}/favicon.svg`;
}

export function resolveAgentAvatarUrl() {
  return "";
}

export function resolveAgentConfig() {
  return null;
}

export function resolveConfiguredCronModelSuggestions() {
  return [];
}

export function resolveEffectiveModelFallbacks() {
  return {};
}

export function resolveModelPrimary() {
  return null;
}

export function sortLocaleStrings(items: Set<string> | string[]): string[] {
  const arr = Array.isArray(items) ? items : Array.from(items);
  return arr.sort((a, b) => a.localeCompare(b));
}

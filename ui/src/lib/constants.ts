// Barrel re-exports for backward compatibility.
export { ROUTES } from "./routes";

export const LOCAL_STORAGE_KEYS = {
  TOKEN: "pomclaw:token",
  USER_ID: "pomclaw:userId",
  TENANT_ID: "pomclaw:tenant_id",
  THEME: "pomclaw:theme",
  SIDEBAR_COLLAPSED: "pomclaw:sidebarCollapsed",
  LANGUAGE: "pomclaw:language",
  TIMEZONE: "pomclaw:timezone",
} as const;

export const SUPPORTED_LANGUAGES = ["en", "zh"] as const;
export type Language = (typeof SUPPORTED_LANGUAGES)[number];

export const LANGUAGE_LABELS: Record<Language, string> = {
  en: "English",
  zh: "中文",
};

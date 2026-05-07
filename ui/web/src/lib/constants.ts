// Barrel re-exports for backward compatibility.
// Import directly from sub-modules for new code.
export { ROUTES } from "./routes";
export {
  TIMEZONE_OPTIONS,
  getAllIanaTimezones,
  isValidIanaTimezone,
} from "./timezone-utils";

export const LOCAL_STORAGE_KEYS = {
  TOKEN: "pomclaw:token",
  USER_ID: "pomclaw:userId",
  SENDER_ID: "pomclaw:senderID",
  TENANT_ID: "pomclaw:tenant_id",
  TENANT_HINT: "pomclaw:tenant_hint",
  SETUP_SKIPPED: "pomclaw:setup_skipped",
  THEME: "pomclaw:theme",
  SIDEBAR_COLLAPSED: "pomclaw:sidebarCollapsed",
  LANGUAGE: "pomclaw:language",
  TIMEZONE: "pomclaw:timezone",
} as const;

export const SUPPORTED_LANGUAGES = ["en", "vi", "zh"] as const;
export type Language = (typeof SUPPORTED_LANGUAGES)[number];

export const LANGUAGE_LABELS: Record<Language, string> = {
  en: "English",
  vi: "Tiếng Việt",
  zh: "中文",
};

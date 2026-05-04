import i18n from "i18next";
import { initReactI18next } from "react-i18next";

// --- EN namespaces ---
import enCommon from "./locales/en/common.json";
import enSidebar from "./locales/en/sidebar.json";
import enTopbar from "./locales/en/topbar.json";
import enLogin from "./locales/en/login.json";
import enOverview from "./locales/en/overview.json";
import enChat from "./locales/en/chat.json";
import enAgents from "./locales/en/agents.json";
import enSkills from "./locales/en/skills.json";
import enConfig from "./locales/en/config.json";
import enChannels from "./locales/en/channels.json";
import enProviders from "./locales/en/providers.json";
import enTools from "./locales/en/tools.json";
import enSetup from "./locales/en/setup.json";
import enV3Capabilities from "./locales/en/v3-capabilities.json";
import enHooks from "./locales/en/hooks.json";
import enTts from "./locales/en/tts.json";
import enTraces from "./locales/en/traces.json";
import enCron from "./locales/en/cron.json";
import enUsage from "./locales/en/usage.json";
import enPackages from "./locales/en/packages.json";

// --- VI namespaces ---
import viCommon from "./locales/vi/common.json";
import viSidebar from "./locales/vi/sidebar.json";
import viTopbar from "./locales/vi/topbar.json";
import viLogin from "./locales/vi/login.json";
import viOverview from "./locales/vi/overview.json";
import viChat from "./locales/vi/chat.json";
import viAgents from "./locales/vi/agents.json";
import viSkills from "./locales/vi/skills.json";
import viConfig from "./locales/vi/config.json";
import viChannels from "./locales/vi/channels.json";
import viProviders from "./locales/vi/providers.json";
import viTools from "./locales/vi/tools.json";
import viSetup from "./locales/vi/setup.json";
import viV3Capabilities from "./locales/vi/v3-capabilities.json";
import viHooks from "./locales/vi/hooks.json";
import viTts from "./locales/vi/tts.json";
import viTraces from "./locales/vi/traces.json";
import viCron from "./locales/vi/cron.json";
import viUsage from "./locales/vi/usage.json";
import viPackages from "./locales/vi/packages.json";

// --- ZH namespaces ---
import zhCommon from "./locales/zh/common.json";
import zhSidebar from "./locales/zh/sidebar.json";
import zhTopbar from "./locales/zh/topbar.json";
import zhLogin from "./locales/zh/login.json";
import zhOverview from "./locales/zh/overview.json";
import zhChat from "./locales/zh/chat.json";
import zhAgents from "./locales/zh/agents.json";
import zhSkills from "./locales/zh/skills.json";
import zhConfig from "./locales/zh/config.json";
import zhChannels from "./locales/zh/channels.json";
import zhProviders from "./locales/zh/providers.json";
import zhTools from "./locales/zh/tools.json";
import zhSetup from "./locales/zh/setup.json";
import zhV3Capabilities from "./locales/zh/v3-capabilities.json";
import zhHooks from "./locales/zh/hooks.json";
import zhTts from "./locales/zh/tts.json";
import zhTraces from "./locales/zh/traces.json";
import zhCron from "./locales/zh/cron.json";
import zhUsage from "./locales/zh/usage.json";
import zhPackages from "./locales/zh/packages.json";

const STORAGE_KEY = "goclaw:language";

function getInitialLanguage(): string {
  const stored = localStorage.getItem(STORAGE_KEY);
  if (stored === "en" || stored === "vi" || stored === "zh") return stored;
  const lang = navigator.language.toLowerCase();
  if (lang.startsWith("vi")) return "vi";
  if (lang.startsWith("zh")) return "zh";
  return "en";
}

const ns = [
  "common", "sidebar", "topbar", "login", "overview", "chat",
  "agents", "skills", "config", "channels", "providers", "tools",
  "setup", "v3-capabilities", "hooks", "tts",
  "traces", "cron", "usage", "packages",
] as const;

i18n.use(initReactI18next).init({
  resources: {
    en: {
      common: enCommon, sidebar: enSidebar, topbar: enTopbar, login: enLogin,
      overview: enOverview, chat: enChat, agents: enAgents, skills: enSkills,
      config: enConfig, channels: enChannels, providers: enProviders, tools: enTools,
      setup: enSetup, "v3-capabilities": enV3Capabilities, hooks: enHooks, tts: enTts,
      traces: enTraces, cron: enCron, usage: enUsage, packages: enPackages,
    },
    vi: {
      common: viCommon, sidebar: viSidebar, topbar: viTopbar, login: viLogin,
      overview: viOverview, chat: viChat, agents: viAgents, skills: viSkills,
      config: viConfig, channels: viChannels, providers: viProviders, tools: viTools,
      setup: viSetup, "v3-capabilities": viV3Capabilities, hooks: viHooks, tts: viTts,
      traces: viTraces, cron: viCron, usage: viUsage, packages: viPackages,
    },
    zh: {
      common: zhCommon, sidebar: zhSidebar, topbar: zhTopbar, login: zhLogin,
      overview: zhOverview, chat: zhChat, agents: zhAgents, skills: zhSkills,
      config: zhConfig, channels: zhChannels, providers: zhProviders, tools: zhTools,
      setup: zhSetup, "v3-capabilities": zhV3Capabilities, hooks: zhHooks, tts: zhTts,
      traces: zhTraces, cron: zhCron, usage: zhUsage, packages: zhPackages,
    },
  },
  ns: [...ns],
  defaultNS: "common",
  lng: getInitialLanguage(),
  fallbackLng: "en",
  interpolation: { escapeValue: false },
  missingKeyHandler: import.meta.env.DEV
    ? (_lngs, _ns, key) => console.warn(`[i18n] missing: ${key}`)
    : undefined,
});

i18n.on("languageChanged", (lng) => {
  localStorage.setItem(STORAGE_KEY, lng);
  document.documentElement.lang = lng;
});

export default i18n;

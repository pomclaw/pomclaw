import { getSafeLocalStorage } from "../local-storage.ts";

// 设备认证类型定义
type DeviceAuthEntry = {
  gatewayUrl: string;
  token: string;
  expiresAt?: number;
};

type DeviceAuthStore = {
  version: number;
  entries: Record<string, DeviceAuthEntry>;
};

// 从存储加载设备认证令牌
function loadDeviceAuthTokenFromStore(store: DeviceAuthStore | null, gatewayUrl: string): DeviceAuthEntry | null {
  if (!store || !store.entries) {
    return null;
  }
  return store.entries[gatewayUrl] || null;
}

// 在存储中保存设备认证令牌
function storeDeviceAuthTokenInStore(store: DeviceAuthStore | null, entry: DeviceAuthEntry): DeviceAuthStore {
  const entries = store?.entries || {};
  return {
    version: 1,
    entries: {
      ...entries,
      [entry.gatewayUrl]: entry,
    },
  };
}

// 从存储中清除设备认证令牌
function clearDeviceAuthTokenFromStore(store: DeviceAuthStore | null, gatewayUrl: string): DeviceAuthStore {
  if (!store || !store.entries) {
    return { version: 1, entries: {} };
  }
  const entries = { ...store.entries };
  delete entries[gatewayUrl];
  return { version: 1, entries };
}

const STORAGE_KEY = "openclaw.device.auth.v1";

function readStore(): DeviceAuthStore | null {
  try {
    const raw = getSafeLocalStorage()?.getItem(STORAGE_KEY);
    if (!raw) {
      return null;
    }
    const parsed = JSON.parse(raw) as DeviceAuthStore;
    if (!parsed || parsed.version !== 1) {
      return null;
    }
    if (!parsed.deviceId || typeof parsed.deviceId !== "string") {
      return null;
    }
    if (!parsed.tokens || typeof parsed.tokens !== "object") {
      return null;
    }
    return parsed;
  } catch {
    return null;
  }
}

function writeStore(store: DeviceAuthStore) {
  try {
    getSafeLocalStorage()?.setItem(STORAGE_KEY, JSON.stringify(store));
  } catch {
    // best-effort
  }
}

export function loadDeviceAuthToken(params: {
  deviceId: string;
  role: string;
}): DeviceAuthEntry | null {
  return loadDeviceAuthTokenFromStore({
    adapter: { readStore, writeStore },
    deviceId: params.deviceId,
    role: params.role,
  });
}

export function storeDeviceAuthToken(params: {
  deviceId: string;
  role: string;
  token: string;
  scopes?: string[];
}): DeviceAuthEntry {
  return storeDeviceAuthTokenInStore({
    adapter: { readStore, writeStore },
    deviceId: params.deviceId,
    role: params.role,
    token: params.token,
    scopes: params.scopes,
  });
}

export function clearDeviceAuthToken(params: { deviceId: string; role: string }) {
  clearDeviceAuthTokenFromStore({
    adapter: { readStore, writeStore },
    deviceId: params.deviceId,
    role: params.role,
  });
}

type AuthConfig = {
  token: string;
  userId: string;
  senderID: string;
  tenantId: string;
};

let _config: AuthConfig = { token: '', userId: '', senderID: '', tenantId: '' };
let _onAuthFailure: (() => void) | null = null;

export function setAuthConfig(config: AuthConfig) {
  _config = config;
}

export function getAuthHeaders(): Record<string, string> {
  return {
    Authorization: `Bearer ${_config.token}`,
    'X-User-Id': _config.userId,
    'X-Sender-Id': _config.senderID,
    'X-Tenant-Id': _config.tenantId,
  };
}

export function setOnAuthFailure(fn: (() => void) | null) {
  _onAuthFailure = fn;
}

export function triggerAuthFailure() {
  _onAuthFailure?.();
}

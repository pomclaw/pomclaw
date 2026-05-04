// Stub for removed channels page
export function getChannelAttentionPriority(_channel: any, _enabled: boolean) {
  return 0;
}

export function getChannelStatusFallback(_instance: any) {
  return null;
}

export function getChannelStatusMeta(_status: any, _enabled?: boolean, _t?: any) {
  return { icon: null, label: '', variant: 'default' as const, dotClass: 'bg-gray-400' };
}

export function ChannelStatusView() {
  return null;
}

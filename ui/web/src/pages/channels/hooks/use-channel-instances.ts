// Stub for removed channels page
export interface ChannelInstance {
  channel_type: string;
  config: Record<string, any>;
  [key: string]: any;
}

export function useChannelInstances(_params?: any) {
  return {
    instances: [] as ChannelInstance[],
    total: 0,
    loading: false,
    createInstance: (_data: any) => Promise.resolve()
  };
}

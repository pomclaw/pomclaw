// Stub for removed builtin-tools page
export interface BuiltinTool {
  name: string;
  display_name?: string;
  [key: string]: any;
}

export function useBuiltinTools() {
  return { tools: [] as BuiltinTool[], loading: false };
}

export interface SkillInfo {
  id?: string;
  name: string;
  slug?: string;
  description: string;
  source: string;
  visibility?: string;
  tags?: string[];
  version?: number;
  is_system?: boolean;
  status?: string;
  enabled?: boolean;
  tenant_enabled?: boolean | null;
  missing_deps?: string[];
  content_url?: string;
  is_shared?: boolean;
  created_by?: string;
  created_by_name?: string;
  agent_count?: number;
}

export interface SkillVersions {
  versions: number[];
  current: number;
}

export interface SkillWithGrant {
  id: string;
  name: string;
  slug: string;
  description: string;
  visibility: string;
  version: number;
  granted: boolean;
  pinned_version?: number;
  is_system: boolean;
}

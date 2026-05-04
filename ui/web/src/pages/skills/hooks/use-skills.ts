// Stub for removed skills page
export interface Skill {
  name: string;
  description?: string;
  [key: string]: any;
}

export function useSkills() {
  return { skills: [] as Skill[], loading: false, refresh: () => {} };
}

export function useSkillDetail(_id: string) {
  return { skill: null, loading: false };
}

// Stub
export function getChatGPTOAuthPoolOwnership(_provider?: any) {
  return {
    mode: 'none' as const,
    owners: [],
    ownerByMember: new Map<string, string>(),
    membersByOwner: new Map<string, Set<string>>()
  };
}
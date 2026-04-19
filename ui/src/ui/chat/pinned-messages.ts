// Stub file - pinned messages removed
export class PinnedMessages {
  indices = new Set<number>();

  constructor() {}

  pin() {}
  unpin() {}
  isPinned() {
    return false;
  }
  getPins() {
    return [];
  }
}

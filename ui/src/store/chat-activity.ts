// Chat activity atoms using Jotai
// Ported from GoClaw's chat-activity-store.ts (Zustand → Jotai)

import { atom } from 'jotai'

export interface Activity {
  phase: string // thinking, tool_exec, compacting, streaming, retrying, leader_processing
  tool?: string
  iteration?: number
}

// Base atoms
export const isRunningAtom = atom(false)
export const activityAtom = atom<Activity | null>(null)
export const currentRunIdAtom = atom<string | null>(null)

// Setter atoms for each action
export const startRunAtom = atom(
  null,
  (_get, set, runId: string) => {
    set(isRunningAtom, true)
    set(currentRunIdAtom, runId)
    set(activityAtom, { phase: 'thinking' })
  }
)

export const setActivityAtom = atom(
  null,
  (_get, set, activity: Activity | null) => {
    set(activityAtom, activity)
  }
)

export const completeRunAtom = atom(
  null,
  (_get, set) => {
    set(isRunningAtom, false)
    set(activityAtom, null)
    set(currentRunIdAtom, null)
  }
)

export const failRunAtom = atom(
  null,
  (_get, set) => {
    set(isRunningAtom, false)
    set(activityAtom, null)
    set(currentRunIdAtom, null)
  }
)

export const cancelRunAtom = atom(
  null,
  (_get, set) => {
    set(isRunningAtom, false)
    set(activityAtom, null)
    set(currentRunIdAtom, null)
  }
)

// Restore running state on session switch (without creating a new assistant message).
export const restoreRunningAtom = atom(
  null,
  (_get, set, activity?: Activity | null) => {
    set(isRunningAtom, true)
    set(activityAtom, activity ?? null)
  }
)

export const clearActivityAtom = atom(
  null,
  (_get, set) => {
    set(isRunningAtom, false)
    set(activityAtom, null)
    set(currentRunIdAtom, null)
  }
)

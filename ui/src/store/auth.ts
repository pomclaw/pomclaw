import { atom, getDefaultStore } from "jotai"

export interface AuthUser {
  id: string
  username: string
  email: string
}

export interface AuthState {
  user: AuthUser | null
  accessToken: string | null
  isLoading: boolean
  error: string | null
}

const DEFAULT_AUTH_STATE: AuthState = {
  user: null,
  accessToken: null,
  isLoading: false,
  error: null,
}

const STORAGE_KEY = "pomclaw_auth"

/**
 * 从 localStorage 恢复认证状态
 */
function restoreAuthFromStorage(): Partial<AuthState> {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored) {
      return JSON.parse(stored)
    }
  } catch {
    // 解析失败，忽略
  }
  return {}
}

/**
 * 将认证状态保存到 localStorage
 */
function persistAuthToStorage(state: AuthState) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify({
      user: state.user,
      accessToken: state.accessToken,
    }))
  } catch {
    // 存储失败，忽略
  }
}

// 初始化 auth 状态：尝试从 localStorage 恢复
const restoredAuth = restoreAuthFromStorage()
const initialState = {
  ...DEFAULT_AUTH_STATE,
  ...restoredAuth,
  // If we have a token but no user, we need to load the user (set isLoading)
  isLoading: restoredAuth.accessToken && !restoredAuth.user ? true : false,
}

export const authAtom = atom<AuthState>(initialState)

// Derived atoms for convenience
export const userAtom = atom(
  (get) => get(authAtom).user,
  (get, set, user: AuthUser | null) => {
    const state = get(authAtom)
    set(authAtom, { ...state, user })
  },
)

export const accessTokenAtom = atom(
  (get) => get(authAtom).accessToken,
  (get, set, token: string | null) => {
    const state = get(authAtom)
    set(authAtom, { ...state, accessToken: token })
  },
)

export const isAuthenticatedAtom = atom(
  (get) => get(authAtom).user !== null && get(authAtom).accessToken !== null,
)

export const isLoadingAtom = atom((get) => get(authAtom).isLoading)

export const errorAtom = atom((get) => get(authAtom).error)

const store = getDefaultStore()

export function getAuthState() {
  return store.get(authAtom)
}

export function updateAuthStore(
  patch:
    | Partial<AuthState>
    | ((prev: AuthState) => Partial<AuthState> | AuthState),
) {
  store.set(authAtom, (prev) => {
    const nextPatch = typeof patch === "function" ? patch(prev) : patch
    const next = { ...prev, ...nextPatch }
    // 每次更新都持久化到 localStorage
    persistAuthToStorage(next)
    return next
  })
}

import { useAtom, useAtomValue } from "jotai"
import { useCallback, useState } from "react"

import {
  getCurrentUser,
  loginUser,
  logoutUser,
  registerUser,
} from "@/api/gateway-auth"
import {
  accessTokenAtom,
  authAtom,
  errorAtom,
  getAuthState,
  isAuthenticatedAtom,
  isLoadingAtom,
  updateAuthStore,
} from "@/store/auth"

export function useAuth() {
  const [authState] = useAtom(authAtom)
  const isAuthenticated = useAtomValue(isAuthenticatedAtom)
  const isLoading = useAtomValue(isLoadingAtom)
  const error = useAtomValue(errorAtom)
  const accessToken = useAtomValue(accessTokenAtom)

  const [internalError, setInternalError] = useState<string | null>(null)

  const register = useCallback(
    async (username: string, email: string, password: string) => {
      setInternalError(null)
      updateAuthStore({ isLoading: true, error: null })

      try {
        const response = await registerUser(username, email, password)
        updateAuthStore({
          user: response.user,
          accessToken: response.access_token,
          isLoading: false,
        })
        return true
      } catch (err) {
        const msg = err instanceof Error ? err.message : "Registration failed"
        updateAuthStore({
          isLoading: false,
          error: msg,
        })
        setInternalError(msg)
        return false
      }
    },
    [],
  )

  const login = useCallback(
    async (username: string, password: string) => {
      setInternalError(null)
      updateAuthStore({ isLoading: true, error: null })

      try {
        const response = await loginUser(username, password)
        updateAuthStore({
          user: response.user,
          accessToken: response.access_token,
          isLoading: false,
        })
        return true
      } catch (err) {
        const msg = err instanceof Error ? err.message : "Login failed"
        updateAuthStore({
          isLoading: false,
          error: msg,
        })
        setInternalError(msg)
        return false
      }
    },
    [],
  )

  const logout = useCallback(async () => {
    setInternalError(null)
    updateAuthStore({
      user: null,
      accessToken: null,
      isLoading: false,
      error: null,
    })

    try {
      await logoutUser()
    } catch (err) {
      // Logout failed, but we've already cleared state
      console.error("Logout error:", err)
    }
  }, [])

  const restoreSession = useCallback(async () => {
    // Try to restore session from existing token/cookie
    // This is called on app load
    const authState = getAuthState()
    if (authState.accessToken && !authState.user) {
      updateAuthStore({ isLoading: true })
      try {
        const user = await getCurrentUser()
        updateAuthStore({
          user,
          isLoading: false,
        })
      } catch (err) {
        // Token invalid or expired
        updateAuthStore({
          user: null,
          accessToken: null,
          isLoading: false,
        })
      }
    } else {
      // No token to restore, mark loading as complete
      updateAuthStore({ isLoading: false })
    }
  }, [])

  return {
    user: authState.user,
    isAuthenticated,
    isLoading,
    error: error || internalError,
    accessToken,
    register,
    login,
    logout,
    restoreSession,
  }
}

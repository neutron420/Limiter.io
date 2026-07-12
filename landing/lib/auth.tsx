"use client"

import * as React from "react"

import { api, ApiError, tokens } from "./api"
import type { AuthResponse } from "./types"

interface AuthUser {
  email: string
  userId: string
}

interface AuthContextValue {
  user: AuthUser | null
  ready: boolean // finished reading persisted tokens
  login: (email: string, password: string) => Promise<void>
  register: (email: string, password: string) => Promise<void>
  logout: () => Promise<void>
}

const AuthContext = React.createContext<AuthContextValue | null>(null)

const USER_KEY = "limiter_user"

function readPersistedUser(): AuthUser | null {
  if (typeof window === "undefined") return null
  if (!tokens.access()) return null
  try {
    const raw = localStorage.getItem(USER_KEY)
    return raw ? (JSON.parse(raw) as AuthUser) : null
  } catch {
    return null
  }
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = React.useState<AuthUser | null>(null)
  const [ready, setReady] = React.useState(false)

  React.useEffect(() => {
    setUser(readPersistedUser())
    setReady(true)
  }, [])

  const persist = React.useCallback((data: AuthResponse) => {
    tokens.set(data.access_token, data.refresh_token)
    const u: AuthUser = { email: data.email, userId: data.user_id }
    localStorage.setItem(USER_KEY, JSON.stringify(u))
    setUser(u)
  }, [])

  const login = React.useCallback(
    async (email: string, password: string) => {
      const data = await api.post<AuthResponse>("/auth/login", { email, password }, { auth: false })
      persist(data)
    },
    [persist],
  )

  const register = React.useCallback(
    async (email: string, password: string) => {
      const data = await api.post<AuthResponse>("/auth/register", { email, password }, { auth: false })
      // Some backends return tokens on register; if not, fall back to login.
      if (data?.access_token) persist(data)
      else await login(email, password)
    },
    [persist, login],
  )

  const logout = React.useCallback(async () => {
    try {
      await api.post("/auth/logout")
    } catch (e) {
      if (!(e instanceof ApiError)) throw e // ignore auth errors on logout
    } finally {
      tokens.clear()
      localStorage.removeItem(USER_KEY)
      setUser(null)
      window.location.href = "/login"
    }
  }, [])

  const value = React.useMemo(
    () => ({ user, ready, login, register, logout }),
    [user, ready, login, register, logout],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const ctx = React.useContext(AuthContext)
  if (!ctx) throw new Error("useAuth must be used within an AuthProvider")
  return ctx
}

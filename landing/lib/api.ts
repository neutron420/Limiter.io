// Typed fetch client for the Limiter.io API.
// Handles: base URL, Bearer injection, one silent refresh on 401, JSON parse + errors.

import type { AuthResponse } from "./types"

const API_DOMAIN =
  process.env.NEXT_PUBLIC_API_URL?.replace(/\/$/, "") ?? "http://localhost:8080"
export const API_BASE = `${API_DOMAIN}/api/v1`

const ACCESS_KEY = "limiter_access_token"
const REFRESH_KEY = "limiter_refresh_token"

export class ApiError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.status = status
    this.name = "ApiError"
  }
}

// --- token storage (localStorage; SSR-safe guards) ---
export const tokens = {
  access: () => (typeof window === "undefined" ? null : localStorage.getItem(ACCESS_KEY)),
  refresh: () => (typeof window === "undefined" ? null : localStorage.getItem(REFRESH_KEY)),
  set(access: string, refresh: string) {
    if (typeof window === "undefined") return
    localStorage.setItem(ACCESS_KEY, access)
    localStorage.setItem(REFRESH_KEY, refresh)
  },
  clear() {
    if (typeof window === "undefined") return
    localStorage.removeItem(ACCESS_KEY)
    localStorage.removeItem(REFRESH_KEY)
  },
}

async function parse<T>(res: Response): Promise<T> {
  const text = await res.text()
  const body = text ? safeJson(text) : null
  if (!res.ok) {
    const msg = (body && (body.error || body.message)) || res.statusText || "Request failed"
    throw new ApiError(res.status, msg)
  }
  return body as T
}

function safeJson(text: string): any {
  try {
    return JSON.parse(text)
  } catch {
    return { message: text }
  }
}

async function refreshAccessToken(): Promise<boolean> {
  const refresh = tokens.refresh()
  if (!refresh) return false
  try {
    const res = await fetch(`${API_BASE}/auth/refresh`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refresh_token: refresh }),
    })
    if (!res.ok) return false
    const data = (await res.json()) as AuthResponse
    tokens.set(data.access_token, data.refresh_token)
    return true
  } catch {
    return false
  }
}

interface RequestOptions {
  auth?: boolean // default true — attach Bearer + auto-refresh
  headers?: Record<string, string>
}

async function request<T>(
  method: string,
  path: string,
  body?: unknown,
  opts: RequestOptions = {},
): Promise<T> {
  const withAuth = opts.auth !== false
  const doFetch = () => {
    const headers: Record<string, string> = { ...opts.headers }
    if (body !== undefined) headers["Content-Type"] = "application/json"
    const access = tokens.access()
    if (withAuth && access) headers["Authorization"] = `Bearer ${access}`
    return fetch(`${API_BASE}${path}`, {
      method,
      headers,
      body: body !== undefined ? JSON.stringify(body) : undefined,
    })
  }

  let res = await doFetch()
  if (res.status === 401 && withAuth && tokens.refresh()) {
    const ok = await refreshAccessToken()
    if (ok) {
      res = await doFetch()
    } else {
      tokens.clear()
      if (typeof window !== "undefined" && !window.location.pathname.startsWith("/login")) {
        window.location.href = "/login"
      }
    }
  }
  return parse<T>(res)
}

export const api = {
  get: <T>(path: string, opts?: RequestOptions) => request<T>("GET", path, undefined, opts),
  post: <T>(path: string, body?: unknown, opts?: RequestOptions) => request<T>("POST", path, body, opts),
  put: <T>(path: string, body?: unknown, opts?: RequestOptions) => request<T>("PUT", path, body, opts),
  patch: <T>(path: string, body?: unknown, opts?: RequestOptions) => request<T>("PATCH", path, body, opts),
  del: <T>(path: string, opts?: RequestOptions) => request<T>("DELETE", path, undefined, opts),
}

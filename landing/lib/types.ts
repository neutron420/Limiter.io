// TypeScript mirrors of the Go backend DTOs / models.
// Source of truth: internal/dto/dto.go and internal/models/models.go

export type Algorithm =
  | "token_bucket"
  | "fixed_window"
  | "sliding_window_counter"
  | "sliding_window_log"
  | "leaky_bucket"

export type PlanID = "free" | "pro" | "enterprise"
export type Decision = "allowed" | "blocked"

export interface AuthResponse {
  access_token: string
  refresh_token: string
  email: string
  user_id: string
}

export interface Project {
  id: string
  user_id: string
  name: string
  description: string
  created_at: string
  updated_at: string
}

export interface ApiKey {
  id: string
  project_id: string
  name: string
  prefix: string
  /** Full secret — only returned once, at creation time. */
  key?: string
  expires_at?: string | null
  revoked_at?: string | null
  last_used_at?: string | null
  created_at: string
}

export interface Rule {
  id: string
  project_id: string
  name: string
  route_pattern: string
  algorithm: Algorithm
  limit: number
  period: number
  burst: number
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface Subscription {
  user_id: string
  plan_id: PlanID
  status: string
  starts_at: string
  expires_at?: string | null
}

export interface TopBlocked {
  route: string
  count: number
}

export interface Stats {
  total_requests: number
  allowed_requests: number
  blocked_requests: number
  avg_latency_ms: number
  top_blocked: TopBlocked[] | null
}

export interface AnalyticsLog {
  id: string
  project_id: string
  api_key_id: string
  request_id: string
  client_ip: string
  route: string
  status_code: number
  latency_ms: number
  decision: Decision
  blocked_reason?: string
  timestamp: string
}

// Human-readable metadata for the 5 algorithms — used across policy UI.
export const ALGORITHMS: { value: Algorithm; label: string; blurb: string }[] = [
  { value: "token_bucket", label: "Token Bucket", blurb: "Refills tokens at a set rate; absorbs short bursts." },
  { value: "fixed_window", label: "Fixed Window", blurb: "Counter resets on fixed epoch windows." },
  { value: "sliding_window_counter", label: "Sliding Window Counter", blurb: "Blends previous + current window for smooth rates." },
  { value: "sliding_window_log", label: "Sliding Window Log", blurb: "Exact per-request timestamps. Highest accuracy, more memory." },
  { value: "leaky_bucket", label: "Leaky Bucket", blurb: "Drips queued requests at a constant interval." },
]

export const ALGO_LABEL: Record<Algorithm, string> = Object.fromEntries(
  ALGORITHMS.map((a) => [a.value, a.label]),
) as Record<Algorithm, string>

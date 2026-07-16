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
  avatar_url?: string
}

export interface Project {
  id: string
  user_id: string
  name: string
  description: string
  created_at: string
  updated_at: string
  role: "owner" | "admin" | "member"
}

export interface ProjectInvite {
  id: string
  project_id: string
  email: string
  role: "admin" | "member"
  status: "pending" | "accepted" | "revoked" | "expired"
  expires_at: string
  created_at: string
}

export interface AcceptInviteResponse {
  project_id: string
  project_name: string
  role: "admin" | "member"
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
  key_strategy: string
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

export interface Organization {
  id: string
  name: string
  slug: string
  description: string
  owner_id: string
  plan: string
  created_at: string
}

export interface OrganizationMember {
  id: string
  organization_id: string
  user_id: string
  role: string
  joined_at: string
}

export interface OrganizationGroup {
  id: string
  organization_id: string
  name: string
  description: string
}

export interface ApprovalWorkflow {
  id: string
  organization_id: string
  name: string
  action_type: string
  min_approvers: number
  enabled: boolean
}

export interface ApprovalRequest {
  id: string
  workflow_id: string
  requested_by: string
  status: "pending" | "approved" | "rejected"
  reason: string
  approved_by: string[]
  created_at: string
}

export interface NotificationPreferences {
  id: string
  user_id: string
  project_id: string
  email_notifications: boolean
  slack_notifications: boolean
  slack_webhook_url: string
  rate_limit_alerts: boolean
  member_join_alerts: boolean
  key_rotation_alerts: boolean
  weekly_digest: boolean
}

export interface Quota {
  id: string
  project_id: string
  per_minute: number
  per_hour: number
  per_day: number
  per_month: number
}

export interface TenantConfig {
  id: string
  project_id: string
  tenant_id: string
  customer_id: string
  max_req: number
  window_ms: number
  enabled: boolean
}

export interface SavedAnalyticsView {
  id: string
  project_id: string
  name: string
  config: string
  is_shared: boolean
  created_at: string
}

export interface AnomalyDetectionConfig {
  id: string
  project_id: string
  enabled: boolean
  sensitivity: number
  lookback_minutes: number
  alert_on_spike: boolean
  alert_on_drop: boolean
  slack_webhook_url: string
}

export interface Passkey {
  id: string
  nickname: string
  created_at: string
  last_used_at: string | null
}

export interface ImmutableAuditLog {
  id: string
  user_id: string
  project_id: string
  action: string
  resource: string
  details: string
  ip_address: string
  checksum: string
  prev_hash: string
  created_at: string
}

export interface Invoice {
  id: string
  organization_id: string
  amount: number
  currency: string
  status: "pending" | "paid"
  period_start: string
  period_end: string
  paid_at: string | null
  created_at: string
}

export interface SLAConfig {
  id: string
  organization_id: string
  uptime_sla: number
  response_time_p99: number
  support_level: string
  support_contact: string
}

export interface RegionConfig {
  id: string
  organization_id: string
  region: string
  gateway_url: string
  data_residency: boolean
  enabled: boolean
}

export interface EmailTemplate {
  id: string
  organization_id: string
  name: string
  subject: string
  html_body: string
}

export interface UsageRecord {
  id: string
  project_id: string
  request_count: number
  blocked_count: number
  period_start: string
  period_end: string
  tier: string
}

export interface MaintenanceStatus {
  enabled: boolean
  message: string
}

export interface SAMLConfig {
  idp_entity_id: string
  idp_sso_url: string
  sp_entity_id: string
  sp_acs_url: string
  enabled: boolean
}

export interface OIDCConfig {
  issuer_url: string
  client_id: string
  redirect_url: string
  enabled: boolean
  scopes: string
}

export interface AnalyticsData {
  total_requests: number
  blocked_requests: number
  avg_latency_ms: number
  p95: number
  p99: number
  top_routes: { route: string; count: number }[]
  top_api_keys: { api_key_id: string; count: number }[]
  time_series: { time: string; total: number; blocked: number }[]
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

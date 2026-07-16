"use client"

import * as React from "react"
import { Check, CreditCard, Sparkles, Zap, Building2, Save, FileText, Activity, Shield } from "lucide-react"

import {
  Panel,
  PanelHeader,
  BrutalButton,
  Spinner,
  InlineError,
  StatusBadge,
  Label,
  Field,
  SubmitButton,
} from "@/components/dashboard/kit"
import { cn } from "@/lib/utils"
import { api, ApiError } from "@/lib/api"
import { useAuth } from "@/lib/auth"
import { CHECKOUT_URL } from "@/lib/config"
import type { PlanID, Subscription, SLAConfig, Invoice, UsageRecord } from "@/lib/types"

const PLANS: {
  id: PlanID
  name: string
  price: string
  icon: typeof Zap
  features: string[]
}[] = [
  {
    id: "free",
    name: "Free",
    price: "$0",
    icon: Zap,
    features: ["1 project", "Token & fixed window", "10k requests / min", "24h analytics retention"],
  },
  {
    id: "pro",
    name: "Pro",
    price: "$29",
    icon: Sparkles,
    features: ["Unlimited projects", "All 5 algorithms", "1M requests / min", "30d analytics retention"],
  },
  {
    id: "enterprise",
    name: "Enterprise",
    price: "Custom",
    icon: Building2,
    features: ["Dedicated edge", "Custom algorithms", "Unlimited throughput", "90d+ retention & SLA"],
  },
]

export default function BillingPage() {
  const { user } = useAuth()
  const [sub, setSub] = React.useState<Subscription | null>(null)
  const [loading, setLoading] = React.useState(true)
  const [error, setError] = React.useState<string | null>(null)
  const [upgrading, setUpgrading] = React.useState<PlanID | null>(null)
  const [webhooks, setWebhooks] = React.useState<any[]>([])
  const [webhooksLoading, setWebhooksLoading] = React.useState(true)

  const [billingTab, setBillingTab] = React.useState("plan")

  // SLA
  const [sla, setSla] = React.useState<SLAConfig | null>(null)
  const [uptimeSla, setUptimeSla] = React.useState("99.9")
  const [responseTimeP99, setResponseTimeP99] = React.useState("200")
  const [supportLevel, setSupportLevel] = React.useState("standard")
  const [supportContact, setSupportContact] = React.useState("")
  const [savingSla, setSavingSla] = React.useState(false)

  // Invoices
  const [invoices, setInvoices] = React.useState<Invoice[]>([])
  const [invoicesLoading, setInvoicesLoading] = React.useState(true)

  // Usage
  const [usage, setUsage] = React.useState<UsageRecord[]>([])
  const [usageLoading, setUsageLoading] = React.useState(true)

  const load = React.useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const s = await api.get<Subscription>("/subscription")
      setSub(s)
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load subscription")
    } finally {
      setLoading(false)
    }
  }, [])

  const loadWebhooks = React.useCallback(async () => {
    setWebhooksLoading(true)
    try {
      const w = await api.get<any[]>("/billing/webhooks")
      setWebhooks(w ?? [])
    } catch {
      // Ignore background load error
    } finally {
      setWebhooksLoading(false)
    }
  }, [])

  const loadSla = React.useCallback(async () => {
    try {
      const orgs = await api.get<any[]>("/organizations")
      if (orgs.length > 0) {
        const s = await api.get<SLAConfig>(`/organizations/${orgs[0].id}/sla-config`).catch(() => null)
        setSla(s)
        if (s) {
          setUptimeSla(String(s.uptime_sla))
          setResponseTimeP99(String(s.response_time_p99))
          setSupportLevel(s.support_level)
          setSupportContact(s.support_contact)
        }
      }
    } catch {
      // no org
    }
  }, [])

  const loadInvoices = React.useCallback(async () => {
    setInvoicesLoading(true)
    try {
      const projs = await api.get<any[]>("/projects")
      if (projs.length > 0) {
        const inv = await api.get<Invoice[]>(`/projects/${projs[0].id}/invoices`)
        setInvoices(inv ?? [])
      }
    } catch {
      setInvoices([])
    } finally {
      setInvoicesLoading(false)
    }
  }, [])

  const loadUsage = React.useCallback(async () => {
    setUsageLoading(true)
    try {
      const projs = await api.get<any[]>("/projects")
      if (projs.length > 0) {
        const u = await api.get<UsageRecord[]>(`/projects/${projs[0].id}/usage`)
        setUsage(u ?? [])
      }
    } catch {
      setUsage([])
    } finally {
      setUsageLoading(false)
    }
  }, [])

  React.useEffect(() => {
    load()
    loadWebhooks()
    loadSla()
    loadInvoices()
    loadUsage()
  }, [load, loadWebhooks, loadSla, loadInvoices, loadUsage])

  const handleSaveSla = async () => {
    setSavingSla(true)
    setError(null)
    try {
      const orgs = await api.get<any[]>("/organizations")
      if (orgs.length > 0) {
        await api.put(`/organizations/${orgs[0].id}/sla-config`, {
          uptime_sla: parseFloat(uptimeSla) || 99.9,
          response_time_p99: parseInt(responseTimeP99) || 200,
          support_level: supportLevel,
          support_contact: supportContact,
        })
        loadSla()
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to save SLA")
    } finally {
      setSavingSla(false)
    }
  }

  // Pro upgrades go through Lemon Squeezy hosted checkout. After payment,
  // Lemon Squeezy fires the `subscription_created` webhook → POST /billing/webhook,
  // which flips the plan to pro (matched by the checkout email).
  const upgrade = (plan: PlanID) => {
    setError(null)
    if (plan !== "pro") return
    if (!CHECKOUT_URL) {
      setError("Checkout not configured — set NEXT_PUBLIC_LEMONSQUEEZY_CHECKOUT_URL in .env.local")
      return
    }
    setUpgrading(plan)
    const url = new URL(CHECKOUT_URL)
    // Prefill + lock the email so the webhook matches this account.
    if (user?.email) url.searchParams.set("checkout[email]", user.email)
    window.location.href = url.toString()
  }

  const currentPlan = sub?.plan_id ?? "free"

  const billingTabs = [
    { id: "plan", label: "Plan", icon: CreditCard },
    { id: "sla", label: "SLA", icon: Shield },
    { id: "invoices", label: "Invoices", icon: FileText },
    { id: "usage", label: "Usage", icon: Activity },
    { id: "webhooks", label: "Webhooks", icon: Sparkles },
  ]

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h1 className="text-lg font-bold uppercase tracking-widest">Billing</h1>
        <p className="mt-1 text-[10px] uppercase tracking-wider text-muted-foreground">
          Manage your plan and throughput limits
        </p>
      </div>

      <div className="flex flex-wrap items-center gap-2 border-b-2 border-foreground pb-2 select-none">
        {billingTabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setBillingTab(tab.id)}
            className={`flex items-center gap-2 px-4 py-1.5 text-[10px] font-bold uppercase transition-all cursor-pointer border-2 ${
              billingTab === tab.id
                ? "bg-foreground text-background border-foreground shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] translate-x-[-1px] translate-y-[-1px]"
                : "hover:bg-muted/10 text-foreground border-transparent"
            }`}
          >
            <tab.icon size={14} />
            {tab.label}
          </button>
        ))}
      </div>

      <InlineError message={error} />

      {billingTab === "plan" && (
        <>
          <Panel accent>
            <PanelHeader title="Current Subscription" icon={CreditCard} />
            {loading ? (
              <Spinner label="LOADING PLAN" />
            ) : (
              <div className="grid grid-cols-2 gap-px bg-foreground/10 md:grid-cols-4">
                <div className="bg-background p-4">
                  <Label>Plan</Label>
                  <div className="mt-1 text-sm font-bold uppercase text-[#ea580c]">{currentPlan}</div>
                </div>
                <div className="bg-background p-4">
                  <Label>Status</Label>
                  <div className="mt-1">
                    <StatusBadge status={sub?.status ?? "unknown"} />
                  </div>
                </div>
                <div className="bg-background p-4">
                  <Label>Started</Label>
                  <div className="mt-1 text-sm">
                    {sub?.starts_at ? new Date(sub.starts_at).toLocaleDateString() : "—"}
                  </div>
                </div>
                <div className="bg-background p-4">
                  <Label>Renews / Expires</Label>
                  <div className="mt-1 text-sm">
                    {sub?.expires_at ? new Date(sub.expires_at).toLocaleDateString() : "—"}
                  </div>
                </div>
              </div>
            )}
          </Panel>

          <div className="grid grid-cols-1 gap-6 md:grid-cols-3">
            {PLANS.map((plan) => {
              const isCurrent = plan.id === currentPlan
              return (
                <Panel key={plan.id} accent={plan.id === "pro"} className="flex flex-col">
                  <div className="border-b-2 border-foreground p-4">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <plan.icon size={16} className="text-[#ea580c]" />
                        <h3 className="text-sm font-bold uppercase tracking-widest">{plan.name}</h3>
                      </div>
                      {isCurrent && <StatusBadge status="active" />}
                    </div>
                    <div className="mt-3 flex items-baseline gap-1">
                      <span className="text-2xl font-bold">{plan.price}</span>
                      {plan.price !== "Custom" && (
                        <span className="text-[10px] uppercase text-muted-foreground">/mo</span>
                      )}
                    </div>
                  </div>
                  <ul className="flex flex-1 flex-col gap-2 p-4">
                    {plan.features.map((f) => (
                      <li key={f} className="flex items-center gap-2 text-xs">
                        <Check size={13} className="shrink-0 text-green-500" />
                        {f}
                      </li>
                    ))}
                  </ul>
                  <div className="p-4 pt-0">
                    <BrutalButton
                      variant={isCurrent ? "outline" : plan.id === "pro" ? "primary" : "outline"}
                      className={cn("w-full justify-center")}
                      disabled={isCurrent || plan.id === "enterprise" || (plan.id === "free" && !isCurrent)}
                      loading={upgrading === plan.id}
                      onClick={() => upgrade(plan.id)}
                    >
                      {isCurrent ? "Current Plan" : plan.id === "enterprise" ? "Contact Sales" : plan.id === "pro" ? "Upgrade via Lemon Squeezy" : "Free"}
                    </BrutalButton>
                  </div>
                </Panel>
              )
            })}
          </div>

          <p className="text-[10px] uppercase tracking-wider text-muted-foreground">
            Payments are processed via Lemon Squeezy. Subscription changes are reconciled through the billing webhook.
          </p>
        </>
      )}

      {billingTab === "sla" && (
        <Panel>
          <PanelHeader icon={Shield} title="Service Level Agreement" action={
            sla ? <StatusBadge status="configured" /> : <StatusBadge status="not set" />
          } />
          <div className="space-y-4">
            <Field label="Uptime SLA (%)" sub="Target uptime percentage">
              <input className="w-full rounded-base border-2 border-foreground bg-main px-3 py-2 font-mono text-sm" type="number" step="0.1" value={uptimeSla} onChange={(e) => setUptimeSla(e.target.value)} />
            </Field>
            <Field label="P99 Response Time (ms)" sub="Maximum acceptable P99 latency">
              <input className="w-full rounded-base border-2 border-foreground bg-main px-3 py-2 font-mono text-sm" type="number" value={responseTimeP99} onChange={(e) => setResponseTimeP99(e.target.value)} />
            </Field>
            <Field label="Support Level">
              <select className="w-full rounded-base border-2 border-foreground bg-main px-3 py-2 font-mono text-sm" value={supportLevel} onChange={(e) => setSupportLevel(e.target.value)}>
                <option value="standard">Standard</option>
                <option value="premium">Premium</option>
                <option value="enterprise">Enterprise</option>
              </select>
            </Field>
            <Field label="Support Contact" sub="Email or Slack channel for support">
              <input className="w-full rounded-base border-2 border-foreground bg-main px-3 py-2 font-mono text-sm" value={supportContact} onChange={(e) => setSupportContact(e.target.value)} />
            </Field>
            <SubmitButton onClick={handleSaveSla} loading={savingSla} icon={Save}>SAVE SLA</SubmitButton>
          </div>
        </Panel>
      )}

      {billingTab === "invoices" && (
        <Panel>
          <PanelHeader icon={FileText} title="Invoices" />
          {invoicesLoading ? (
            <Spinner label="LOADING INVOICES" />
          ) : invoices.length === 0 ? (
            <p className="text-sm text-muted-foreground">No invoices yet</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-left border-collapse font-mono">
                <thead>
                  <tr className="border-b-2 border-foreground bg-muted/10 text-[9px] uppercase tracking-wider text-muted-foreground">
                    <th className="p-3 font-bold">Period</th>
                    <th className="p-3 font-bold">Amount</th>
                    <th className="p-3 font-bold">Status</th>
                    <th className="p-3 font-bold">Paid At</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-foreground/10 text-[11px]">
                  {invoices.map((inv) => (
                    <tr key={inv.id} className="hover:bg-muted/5">
                      <td className="p-3">{new Date(inv.period_start).toLocaleDateString()} – {new Date(inv.period_end).toLocaleDateString()}</td>
                      <td className="p-3 font-bold">${(inv.amount / 100).toFixed(2)} {inv.currency.toUpperCase()}</td>
                      <td className="p-3"><StatusBadge status={inv.status} /></td>
                      <td className="p-3">{inv.paid_at ? new Date(inv.paid_at).toLocaleDateString() : "—"}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </Panel>
      )}

      {billingTab === "usage" && (
        <Panel>
          <PanelHeader icon={Activity} title="Usage Records" />
          {usageLoading ? (
            <Spinner label="LOADING USAGE" />
          ) : usage.length === 0 ? (
            <p className="text-sm text-muted-foreground">No usage records yet</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-left border-collapse font-mono">
                <thead>
                  <tr className="border-b-2 border-foreground bg-muted/10 text-[9px] uppercase tracking-wider text-muted-foreground">
                    <th className="p-3 font-bold">Period</th>
                    <th className="p-3 font-bold">Requests</th>
                    <th className="p-3 font-bold">Blocked</th>
                    <th className="p-3 font-bold">Tier</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-foreground/10 text-[11px]">
                  {usage.map((u) => (
                    <tr key={u.id} className="hover:bg-muted/5">
                      <td className="p-3">{new Date(u.period_start).toLocaleDateString()} – {new Date(u.period_end).toLocaleDateString()}</td>
                      <td className="p-3 tabular-nums">{u.request_count.toLocaleString()}</td>
                      <td className="p-3 tabular-nums">{u.blocked_count.toLocaleString()}</td>
                      <td className="p-3 font-bold uppercase">{u.tier}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </Panel>
      )}

      {billingTab === "webhooks" && (
        <Panel>
          <PanelHeader title="Billing Webhook Logs" icon={Sparkles} />
          {webhooksLoading ? (
            <Spinner label="LOADING AUDIT TRAIL" />
          ) : webhooks.length === 0 ? (
            <p className="p-4 text-[11px] uppercase text-muted-foreground">
              No billing webhooks received yet. Try starting checkout or triggering checkout callback.
            </p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-left border-collapse font-mono">
                <thead>
                  <tr className="border-b-2 border-foreground bg-muted/10 text-[9px] uppercase tracking-wider text-muted-foreground">
                    <th className="p-3 font-bold">Event Name</th>
                    <th className="p-3 font-bold">Received At</th>
                    <th className="p-3 font-bold">Verified</th>
                    <th className="p-3 font-bold">Audit Status</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-foreground/10 text-[11px]">
                  {webhooks.map((w) => (
                    <tr key={w.id} className="hover:bg-muted/5">
                      <td className="p-3 text-[#ea580c] font-bold">{w.event_name}</td>
                      <td className="p-3 tabular-nums">{new Date(w.received_at).toLocaleString()}</td>
                      <td className="p-3">
                        <span className={w.verified ? "text-green-500 font-bold" : "text-red-500 font-bold"}>
                          {w.verified ? "VERIFIED" : "UNVERIFIED"}
                        </span>
                      </td>
                      <td className="p-3 flex items-center gap-2">
                        <StatusBadge status={w.status === "processed" ? "active" : "inactive"} />
                        <span className="font-bold uppercase">{w.status}</span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </Panel>
      )}
    </div>
  )
}

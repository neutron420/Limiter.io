"use client"

import * as React from "react"
import { Check, CreditCard, Sparkles, Zap, Building2 } from "lucide-react"

import {
  Panel,
  PanelHeader,
  BrutalButton,
  Spinner,
  InlineError,
  StatusBadge,
  Label,
} from "@/components/dashboard/kit"
import { cn } from "@/lib/utils"
import { api } from "@/lib/api"
import { useAuth } from "@/lib/auth"
import { CHECKOUT_URL } from "@/lib/config"
import type { PlanID, Subscription } from "@/lib/types"

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

  React.useEffect(() => {
    load()
  }, [load])

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

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h1 className="text-lg font-bold uppercase tracking-widest">Billing</h1>
        <p className="mt-1 text-[10px] uppercase tracking-wider text-muted-foreground">
          Manage your plan and throughput limits
        </p>
      </div>

      <InlineError message={error} />

      {/* Current subscription */}
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

      {/* Plans */}
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
                  {isCurrent
                    ? "Current Plan"
                    : plan.id === "enterprise"
                      ? "Contact Sales"
                      : plan.id === "pro"
                        ? "Upgrade via Lemon Squeezy"
                        : "Free"}
                </BrutalButton>
              </div>
            </Panel>
          )
        })}
      </div>

      <p className="text-[10px] uppercase tracking-wider text-muted-foreground">
        Payments are processed via Lemon Squeezy. Subscription changes are reconciled through the billing
        webhook.
      </p>
    </div>
  )
}

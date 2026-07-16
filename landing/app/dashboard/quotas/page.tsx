"use client"

import * as React from "react"
import { motion } from "framer-motion"
import { Save, RotateCcw, Sliders, Gauge } from "lucide-react"

import {
  Panel,
  PanelHeader,
  BrutalButton,
  Field,
  InlineError,
  Spinner,
  StatusBadge,
  SubmitButton,
} from "@/components/dashboard/kit"
import { RequireProject } from "@/components/dashboard/require-project"
import { api, ApiError } from "@/lib/api"
import { useProject } from "@/lib/project-context"
import { canWrite } from "@/lib/rbac"
import type { Quota } from "@/lib/types"

export default function QuotasPage() {
  const { current, role } = useProject()
  const canWriteProject = canWrite(role)

  const [quota, setQuota] = React.useState<Quota | null>(null)
  const [loading, setLoading] = React.useState(true)
  const [saving, setSaving] = React.useState(false)
  const [error, setError] = React.useState("")
  const [success, setSuccess] = React.useState("")

  const [perMinute, setPerMinute] = React.useState("1000")
  const [perHour, setPerHour] = React.useState("10000")
  const [perDay, setPerDay] = React.useState("100000")
  const [perMonth, setPerMonth] = React.useState("1000000")

  const fetchQuota = React.useCallback(async () => {
    if (!current?.id) return
    setLoading(true)
    try {
      const data = await api.get<Quota>(`/projects/${current.id}/quotas`)
      setQuota(data)
      setPerMinute(String(data.per_minute))
      setPerHour(String(data.per_hour))
      setPerDay(String(data.per_day))
      setPerMonth(String(data.per_month))
    } catch {
      // 404 = no quota set yet, defaults apply
    } finally {
      setLoading(false)
    }
  }, [current?.id])

  React.useEffect(() => { fetchQuota() }, [fetchQuota])

  const handleSave = async () => {
    if (!current?.id) return
    setSaving(true)
    setError("")
    setSuccess("")
    try {
      const data = await api.put<Quota>(`/projects/${current.id}/quotas`, {
        per_minute: parseInt(perMinute) || 0,
        per_hour: parseInt(perHour) || 0,
        per_day: parseInt(perDay) || 0,
        per_month: parseInt(perMonth) || 0,
      })
      setQuota(data)
      setSuccess("Quota limits saved")
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Failed to save quotas")
    } finally {
      setSaving(false)
    }
  }

  return (
    <RequireProject>
      <motion.div initial={{ opacity: 0, y: 12 }} animate={{ opacity: 1, y: 0 }} className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold font-heading tracking-tight">QUOTAS</h1>
            <p className="text-sm text-muted-foreground mt-1">Per‑project rate limits by time window</p>
          </div>
        </div>

        {loading ? (
          <Spinner label="LOADING QUOTAS" />
        ) : (
          <Panel>
            <PanelHeader icon={Gauge} title="Request Quotas" />
            <div className="space-y-4">
              <Field label="Per Minute" sub="Maximum requests allowed each minute">
                <input
                  className="w-full rounded-base border-2 border-foreground bg-main px-3 py-2 font-mono text-sm focus:outline-none"
                  type="number"
                  min="0"
                  value={perMinute}
                  onChange={(e) => setPerMinute(e.target.value)}
                  disabled={!canWriteProject}
                />
              </Field>
              <Field label="Per Hour" sub="Maximum requests allowed each hour">
                <input
                  className="w-full rounded-base border-2 border-foreground bg-main px-3 py-2 font-mono text-sm focus:outline-none"
                  type="number"
                  min="0"
                  value={perHour}
                  onChange={(e) => setPerHour(e.target.value)}
                  disabled={!canWriteProject}
                />
              </Field>
              <Field label="Per Day" sub="Maximum requests allowed each day">
                <input
                  className="w-full rounded-base border-2 border-foreground bg-main px-3 py-2 font-mono text-sm focus:outline-none"
                  type="number"
                  min="0"
                  value={perDay}
                  onChange={(e) => setPerDay(e.target.value)}
                  disabled={!canWriteProject}
                />
              </Field>
              <Field label="Per Month" sub="Maximum requests allowed each month">
                <input
                  className="w-full rounded-base border-2 border-foreground bg-main px-3 py-2 font-mono text-sm focus:outline-none"
                  type="number"
                  min="0"
                  value={perMonth}
                  onChange={(e) => setPerMonth(e.target.value)}
                  disabled={!canWriteProject}
                />
              </Field>
            </div>

            {error && <InlineError message={error} />}
            {success && <p className="text-sm text-green-600 font-mono">{success}</p>}

            {canWriteProject && (
              <div className="flex gap-2 mt-4">
                <SubmitButton onClick={handleSave} loading={saving} icon={Save}>
                  SAVE QUOTAS
                </SubmitButton>
                <BrutalButton onClick={fetchQuota} variant="secondary">
                  <RotateCcw className="size-4" />
                  RESET
                </BrutalButton>
              </div>
            )}
          </Panel>
        )}
      </motion.div>
    </RequireProject>
  )
}

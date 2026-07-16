"use client"

import * as React from "react"
import { motion } from "framer-motion"
import { Plus, Save, Trash2, RotateCcw, Users, ToggleLeft, ToggleRight } from "lucide-react"

import {
  Panel,
  PanelHeader,
  BrutalButton,
  Field,
  Modal,
  InlineError,
  Spinner,
  EmptyState,
  StatusBadge,
  SubmitButton,
} from "@/components/dashboard/kit"
import { RequireProject } from "@/components/dashboard/require-project"
import { api, ApiError } from "@/lib/api"
import { useProject } from "@/lib/project-context"
import { canWrite } from "@/lib/rbac"
import type { TenantConfig } from "@/lib/types"

export default function TenantsPage() {
  const { current, role } = useProject()
  const canWriteProject = canWrite(role)

  const [tenants, setTenants] = React.useState<TenantConfig[]>([])
  const [loading, setLoading] = React.useState(true)
  const [error, setError] = React.useState("")
  const [showModal, setShowModal] = React.useState(false)
  const [saving, setSaving] = React.useState(false)

  const [tenantId, setTenantId] = React.useState("")
  const [customerId, setCustomerId] = React.useState("")
  const [maxReq, setMaxReq] = React.useState("100")
  const [windowMs, setWindowMs] = React.useState("60000")

  const fetchTenants = React.useCallback(async () => {
    if (!current?.id) return
    setLoading(true)
    try {
      const data = await api.get<TenantConfig[]>(`/projects/${current.id}/tenants`)
      setTenants(data)
    } catch {
      setTenants([])
    } finally {
      setLoading(false)
    }
  }, [current?.id])

  React.useEffect(() => { fetchTenants() }, [fetchTenants])

  const handleCreate = async () => {
    if (!current?.id) return
    setSaving(true)
    setError("")
    try {
      await api.post<TenantConfig>(`/projects/${current.id}/tenants`, {
        tenant_id: tenantId,
        customer_id: customerId,
        max_req: parseInt(maxReq) || 100,
        window_ms: parseInt(windowMs) || 60000,
      })
      setShowModal(false)
      setTenantId("")
      setCustomerId("")
      setMaxReq("100")
      setWindowMs("60000")
      fetchTenants()
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Failed to create tenant")
    } finally {
      setSaving(false)
    }
  }

  const handleToggle = async (t: TenantConfig) => {
    try {
      await api.put(`/projects/${current?.id}/tenants/${t.id}`, { enabled: !t.enabled })
      fetchTenants()
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Failed to update tenant")
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await api.del(`/projects/${current?.id}/tenants/${id}`)
      fetchTenants()
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Failed to delete tenant")
    }
  }

  return (
    <RequireProject>
      <motion.div initial={{ opacity: 0, y: 12 }} animate={{ opacity: 1, y: 0 }} className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold font-heading tracking-tight">TENANTS</h1>
            <p className="text-sm text-muted-foreground mt-1">Per‑customer rate limit overrides</p>
          </div>
          {canWriteProject && (
            <BrutalButton onClick={() => setShowModal(true)} icon={Plus}>ADD TENANT</BrutalButton>
          )}
        </div>

        {loading ? (
          <Spinner label="LOADING TENANTS" />
        ) : tenants.length === 0 ? (
          <EmptyState icon={Users} title="No tenants configured" />
        ) : (
          <div className="space-y-3">
            {tenants.map((t) => (
              <Panel key={t.id}>
                <div className="flex items-center justify-between">
                  <div className="space-y-1">
                    <p className="font-mono text-sm font-bold">{t.tenant_id}</p>
                    <p className="text-xs text-muted-foreground">Customer: {t.customer_id}</p>
                    <p className="text-xs text-muted-foreground">
                      {t.max_req} req / {t.window_ms}ms
                    </p>
                  </div>
                  <div className="flex items-center gap-2">
                    <StatusBadge status={t.enabled ? "active" : "disabled"} />
                    {canWriteProject && (
                      <>
                        <BrutalButton onClick={() => handleToggle(t)} variant="secondary" size="sm">
                          {t.enabled ? <ToggleRight className="size-4" /> : <ToggleLeft className="size-4" />}
                        </BrutalButton>
                        <BrutalButton onClick={() => handleDelete(t.id)} variant="destructive" size="sm">
                          <Trash2 className="size-4" />
                        </BrutalButton>
                      </>
                    )}
                  </div>
                </div>
              </Panel>
            ))}
          </div>
        )}

        {error && <InlineError message={error} />}

        <Modal open={showModal} onClose={() => setShowModal(false)} title="New Tenant">
          <div className="space-y-4">
            <Field label="Tenant ID" sub="Unique identifier for this tenant">
              <input className="w-full rounded-base border-2 border-foreground bg-main px-3 py-2 font-mono text-sm" value={tenantId} onChange={(e) => setTenantId(e.target.value)} />
            </Field>
            <Field label="Customer ID" sub="Your internal customer reference">
              <input className="w-full rounded-base border-2 border-foreground bg-main px-3 py-2 font-mono text-sm" value={customerId} onChange={(e) => setCustomerId(e.target.value)} />
            </Field>
            <Field label="Max Requests" sub="Maximum requests in the window">
              <input className="w-full rounded-base border-2 border-foreground bg-main px-3 py-2 font-mono text-sm" type="number" value={maxReq} onChange={(e) => setMaxReq(e.target.value)} />
            </Field>
            <Field label="Window (ms)" sub="Time window in milliseconds">
              <input className="w-full rounded-base border-2 border-foreground bg-main px-3 py-2 font-mono text-sm" type="number" value={windowMs} onChange={(e) => setWindowMs(e.target.value)} />
            </Field>
            <SubmitButton onClick={handleCreate} loading={saving} icon={Save}>CREATE</SubmitButton>
          </div>
        </Modal>
      </motion.div>
    </RequireProject>
  )
}

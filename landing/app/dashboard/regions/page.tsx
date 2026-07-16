"use client"

import * as React from "react"
import { motion } from "framer-motion"
import { Plus, Save, Trash2, Globe, ToggleLeft, ToggleRight } from "lucide-react"

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
import { api, ApiError } from "@/lib/api"
import type { RegionConfig } from "@/lib/types"

export default function RegionsPage() {
  const [orgId, setOrgId] = React.useState<string | null>(null)
  const [regions, setRegions] = React.useState<RegionConfig[]>([])
  const [loading, setLoading] = React.useState(true)
  const [error, setError] = React.useState("")
  const [saving, setSaving] = React.useState(false)
  const [showModal, setShowModal] = React.useState(false)
  const [region, setRegion] = React.useState("")
  const [gatewayUrl, setGatewayUrl] = React.useState("")
  const [dataResidency, setDataResidency] = React.useState(false)

  const fetchData = React.useCallback(async () => {
    setLoading(true)
    try {
      const orgs = await api.get<any[]>("/organizations")
      if (orgs.length > 0) {
        setOrgId(orgs[0].id)
        const r = await api.get<RegionConfig[]>(`/organizations/${orgs[0].id}/region-configs`)
        setRegions(r)
      }
    } catch {
      // no org
    } finally {
      setLoading(false)
    }
  }, [])

  React.useEffect(() => { fetchData() }, [fetchData])

  const handleCreate = async () => {
    if (!orgId) return
    setSaving(true)
    setError("")
    try {
      await api.post(`/organizations/${orgId}/region-configs`, {
        region, gateway_url: gatewayUrl, data_residency: dataResidency,
      })
      setShowModal(false)
      setRegion("")
      setGatewayUrl("")
      setDataResidency(false)
      fetchData()
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Failed to create region config")
    } finally {
      setSaving(false)
    }
  }

  const handleToggle = async (r: RegionConfig) => {
    if (!orgId) return
    try {
      await api.put(`/organizations/${orgId}/region-configs/${r.id}`, { enabled: !r.enabled })
      fetchData()
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Failed to update region")
    }
  }

  const handleDelete = async (id: string) => {
    if (!orgId) return
    try {
      await api.del(`/organizations/${orgId}/region-configs/${id}`)
      fetchData()
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Failed to delete region")
    }
  }

  return (
    <motion.div initial={{ opacity: 0, y: 12 }} animate={{ opacity: 1, y: 0 }} className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold font-heading tracking-tight">REGIONS</h1>
          <p className="text-sm text-muted-foreground mt-1">Multi‑region gateway configuration</p>
        </div>
        {orgId && <BrutalButton onClick={() => setShowModal(true)} icon={Plus}>ADD REGION</BrutalButton>}
      </div>

      {loading ? (
        <Spinner label="LOADING REGIONS" />
      ) : !orgId ? (
        <EmptyState icon={Globe} title="No organization" hint="Create an organization first" />
      ) : regions.length === 0 ? (
        <EmptyState icon={Globe} title="No regions configured" hint="Add your first region" />
      ) : (
        <div className="space-y-3">
          {regions.map((r) => (
            <Panel key={r.id}>
              <div className="flex items-center justify-between">
                <div className="space-y-1">
                  <p className="font-mono text-sm font-bold">{r.region}</p>
                  <p className="text-xs text-muted-foreground">{r.gateway_url}</p>
                  <p className="text-xs text-muted-foreground">
                    {r.data_residency ? "Data residency enforced" : "No data residency"}
                  </p>
                </div>
                <div className="flex items-center gap-2">
                  <StatusBadge status={r.enabled ? "active" : "disabled"} />
                  <BrutalButton onClick={() => handleToggle(r)} variant="secondary" size="sm">
                    {r.enabled ? <ToggleRight className="size-4" /> : <ToggleLeft className="size-4" />}
                  </BrutalButton>
                  <BrutalButton onClick={() => handleDelete(r.id)} variant="destructive" size="sm">
                    <Trash2 className="size-4" />
                  </BrutalButton>
                </div>
              </div>
            </Panel>
          ))}
        </div>
      )}

      {error && <InlineError message={error} />}

      <Modal open={showModal} onClose={() => setShowModal(false)} title="Add Region">
        <div className="space-y-4">
          <Field label="Region" sub="e.g. us-east-1, eu-west-1">
            <input className="w-full rounded-base border-2 border-foreground bg-main px-3 py-2 font-mono text-sm" value={region} onChange={(e) => setRegion(e.target.value)} />
          </Field>
          <Field label="Gateway URL">
            <input className="w-full rounded-base border-2 border-foreground bg-main px-3 py-2 font-mono text-sm" value={gatewayUrl} onChange={(e) => setGatewayUrl(e.target.value)} />
          </Field>
          <label className="flex items-center gap-2 text-sm font-mono cursor-pointer">
            <input type="checkbox" checked={dataResidency} onChange={(e) => setDataResidency(e.target.checked)} className="size-4" />
            Enforce data residency
          </label>
          <SubmitButton onClick={handleCreate} loading={saving} icon={Save}>CREATE</SubmitButton>
        </div>
      </Modal>
    </motion.div>
  )
}

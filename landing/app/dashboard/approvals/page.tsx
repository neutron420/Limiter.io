"use client"

import * as React from "react"
import { motion } from "framer-motion"
import { Plus, Save, Check, X, FileWarning, ThumbsUp, ThumbsDown, Clock } from "lucide-react"

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
import { useAuth } from "@/lib/auth"
import type { ApprovalWorkflow, ApprovalRequest } from "@/lib/types"

export default function ApprovalsPage() {
  const { user } = useAuth()
  const [orgId, setOrgId] = React.useState<string | null>(null)
  const [workflows, setWorkflows] = React.useState<ApprovalWorkflow[]>([])
  const [requests, setRequests] = React.useState<ApprovalRequest[]>([])
  const [loading, setLoading] = React.useState(true)
  const [error, setError] = React.useState("")
  const [saving, setSaving] = React.useState(false)
  const [showModal, setShowModal] = React.useState(false)
  const [name, setName] = React.useState("")
  const [actionType, setActionType] = React.useState("policy_change")
  const [minApprovers, setMinApprovers] = React.useState("2")

  const fetchData = React.useCallback(async () => {
    setLoading(true)
    try {
      const orgs = await api.get<any[]>("/organizations")
      if (orgs.length > 0) {
        setOrgId(orgs[0].id)
        const w = await api.get<ApprovalWorkflow[]>(`/organizations/${orgs[0].id}/approval-workflows`)
        setWorkflows(w)
        const r = await api.get<ApprovalRequest[]>(`/organizations/${orgs[0].id}/approval-requests`)
        setRequests(r)
      }
    } catch {
      // no org
    } finally {
      setLoading(false)
    }
  }, [])

  React.useEffect(() => { fetchData() }, [fetchData])

  const handleCreateWorkflow = async () => {
    if (!orgId) return
    setSaving(true)
    setError("")
    try {
      await api.post(`/organizations/${orgId}/approval-workflows`, {
        name, action_type: actionType, min_approvers: parseInt(minApprovers) || 2,
      })
      setName("")
      setActionType("policy_change")
      setMinApprovers("2")
      setShowModal(false)
      fetchData()
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Failed to create workflow")
    } finally {
      setSaving(false)
    }
  }

  const handleApprove = async (id: string) => {
    if (!orgId) return
    try {
      await api.post(`/approval-requests/${id}/approve`)
      fetchData()
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Failed to approve")
    }
  }

  const handleReject = async (id: string) => {
    if (!orgId) return
    try {
      await api.post(`/approval-requests/${id}/reject`)
      fetchData()
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Failed to reject")
    }
  }

  const statusIcon = (s: string) => {
    switch (s) {
      case "approved": return <ThumbsUp className="size-4 text-green-600" />
      case "rejected": return <ThumbsDown className="size-4 text-red-600" />
      default: return <Clock className="size-4 text-amber-600" />
    }
  }

  return (
    <motion.div initial={{ opacity: 0, y: 12 }} animate={{ opacity: 1, y: 0 }} className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold font-heading tracking-tight">APPROVALS</h1>
          <p className="text-sm text-muted-foreground mt-1">Multi‑approver workflow for sensitive changes</p>
        </div>
        {orgId && <BrutalButton onClick={() => setShowModal(true)} icon={Plus}>NEW WORKFLOW</BrutalButton>}
      </div>

      {loading ? (
        <Spinner label="LOADING APPROVALS" />
      ) : !orgId ? (
        <EmptyState icon={FileWarning} title="No organization" hint="Create an organization first" />
      ) : (
        <>
          <Panel>
            <PanelHeader icon={Check} title="Workflows" />
            {workflows.length === 0 ? (
              <p className="text-sm text-muted-foreground">No workflows defined</p>
            ) : (
              <div className="space-y-2">
                {workflows.map((w) => (
                  <div key={w.id} className="flex items-center justify-between rounded-base border-2 border-foreground p-3">
                    <div>
                      <p className="text-sm font-mono font-bold">{w.name}</p>
                      <p className="text-xs text-muted-foreground">
                        {w.action_type} · min {w.min_approvers} approvers
                      </p>
                    </div>
                    <StatusBadge status={w.enabled ? "active" : "disabled"} />
                  </div>
                ))}
              </div>
            )}
          </Panel>

          <Panel>
            <PanelHeader icon={FileWarning} title="Pending Requests" />
            {requests.filter(r => r.status === "pending").length === 0 ? (
              <p className="text-sm text-muted-foreground">No pending requests</p>
            ) : (
              <div className="space-y-2">
                {requests.filter(r => r.status === "pending").map((r) => (
                  <div key={r.id} className="flex items-center justify-between rounded-base border-2 border-foreground p-3">
                    <div>
                      <p className="text-sm font-mono font-bold">{r.reason}</p>
                      <p className="text-xs text-muted-foreground">by {r.requested_by}</p>
                    </div>
                    <div className="flex items-center gap-2">
                      <BrutalButton onClick={() => handleApprove(r.id)} variant="outline">
                        <ThumbsUp className="size-4" />
                      </BrutalButton>
                      <BrutalButton onClick={() => handleReject(r.id)} variant="danger">
                        <ThumbsDown className="size-4" />
                      </BrutalButton>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </Panel>

          <Panel>
            <PanelHeader icon={Clock} title="History" />
            {requests.filter(r => r.status !== "pending").length === 0 ? (
              <p className="text-sm text-muted-foreground">No history</p>
            ) : (
              <div className="space-y-2">
                {requests.filter(r => r.status !== "pending").map((r) => (
                  <div key={r.id} className="flex items-center justify-between rounded-base border-2 border-foreground p-3">
                    <div className="flex items-center gap-3">
                      {statusIcon(r.status)}
                      <div>
                        <p className="text-sm font-mono font-bold">{r.reason}</p>
                        <p className="text-xs text-muted-foreground">by {r.requested_by}</p>
                      </div>
                    </div>
                    <StatusBadge status={r.status} />
                  </div>
                ))}
              </div>
            )}
          </Panel>
        </>
      )}

      {error && <InlineError message={error} />}

      <Modal open={showModal} onClose={() => setShowModal(false)} title="New Approval Workflow">
        <div className="space-y-4">
          <Field label="Name">
            <input className="w-full rounded-base border-2 border-foreground bg-main px-3 py-2 font-mono text-sm" value={name} onChange={(e) => setName(e.target.value)} />
          </Field>
          <Field label="Action Type">
            <select className="w-full rounded-base border-2 border-foreground bg-main px-3 py-2 font-mono text-sm" value={actionType} onChange={(e) => setActionType(e.target.value)}>
              <option value="policy_change">Policy Change</option>
              <option value="key_rotation">Key Rotation</option>
              <option value="billing_change">Billing Change</option>
              <option value="member_remove">Member Removal</option>
            </select>
          </Field>
          <Field label="Min Approvers">
            <input className="w-full rounded-base border-2 border-foreground bg-main px-3 py-2 font-mono text-sm" type="number" min="1" value={minApprovers} onChange={(e) => setMinApprovers(e.target.value)} />
          </Field>
          <BrutalButton onClick={handleCreateWorkflow} loading={saving} icon={Save}>CREATE</BrutalButton>
        </div>
      </Modal>
    </motion.div>
  )
}

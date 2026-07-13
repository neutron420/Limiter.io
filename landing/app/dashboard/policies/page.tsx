"use client"

import * as React from "react"
import { motion } from "framer-motion"
import { Plus, ShieldAlert, Pencil, Trash2, Power } from "lucide-react"

import {
  Panel,
  PanelHeader,
  BrutalButton,
  Field,
  SelectField,
  Modal,
  InlineError,
  Spinner,
  EmptyState,
  StatusBadge,
  Label,
  SubmitButton,
} from "@/components/dashboard/kit"
import { RequireProject } from "@/components/dashboard/require-project"
import { api, ApiError } from "@/lib/api"
import { ALGORITHMS, ALGO_LABEL, type Algorithm, type Project, type Rule } from "@/lib/types"

interface FormState {
  name: string
  route_pattern: string
  algorithm: Algorithm
  key_strategy: string
  limit: string
  period: string
  burst: string
}

const emptyForm: FormState = {
  name: "",
  route_pattern: "",
  algorithm: "token_bucket",
  key_strategy: "api_key",
  limit: "100",
  period: "60",
  burst: "0",
}

function PoliciesInner({ project }: { project: Project }) {
  const pid = project.id
  const [rules, setRules] = React.useState<Rule[]>([])
  const [loading, setLoading] = React.useState(true)
  const [error, setError] = React.useState<string | null>(null)

  const [editing, setEditing] = React.useState<Rule | null>(null)
  const [formOpen, setFormOpen] = React.useState(false)
  const [form, setForm] = React.useState<FormState>(emptyForm)
  const [saving, setSaving] = React.useState(false)
  const [formError, setFormError] = React.useState<string | null>(null)

  const [toDelete, setToDelete] = React.useState<Rule | null>(null)
  const [deleting, setDeleting] = React.useState(false)

  const load = React.useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const list = await api.get<Rule[]>(`/projects/${pid}/rules`)
      setRules(list ?? [])
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load rules")
    } finally {
      setLoading(false)
    }
  }, [pid])

  React.useEffect(() => {
    load()
  }, [load])

  const openCreate = () => {
    setEditing(null)
    setForm(emptyForm)
    setFormError(null)
    setFormOpen(true)
  }

  const openEdit = (r: Rule) => {
    setEditing(r)
    setForm({
      name: r.name,
      route_pattern: r.route_pattern,
      algorithm: r.algorithm,
      key_strategy: r.key_strategy || "api_key",
      limit: String(r.limit),
      period: String(r.period),
      burst: String(r.burst),
    })
    setFormError(null)
    setFormOpen(true)
  }

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    setFormError(null)
    const payload = {
      name: form.name,
      route_pattern: form.route_pattern,
      algorithm: form.algorithm,
      key_strategy: form.key_strategy,
      limit: Number(form.limit),
      period: Number(form.period),
      burst: Number(form.burst) || 0,
    }
    if (payload.name.length < 3) return setFormError("Name must be at least 3 characters")
    if (!payload.route_pattern) return setFormError("Route pattern is required")
    if (payload.limit <= 0) return setFormError("Limit must be greater than 0")
    if (payload.period <= 0) return setFormError("Period must be greater than 0")

    setSaving(true)
    try {
      if (editing) await api.put(`/projects/${pid}/rules/${editing.id}`, payload)
      else await api.post(`/projects/${pid}/rules`, payload)
      setFormOpen(false)
      await load()
    } catch (err) {
      setFormError(err instanceof ApiError ? err.message : "Failed to save rule")
    } finally {
      setSaving(false)
    }
  }

  const toggleActive = async (r: Rule) => {
    // optimistic
    setRules((prev) => prev.map((x) => (x.id === r.id ? { ...x, is_active: !x.is_active } : x)))
    try {
      await api.put(`/projects/${pid}/rules/${r.id}`, { is_active: !r.is_active })
    } catch {
      load() // revert on failure
    }
  }

  const handleDelete = async () => {
    if (!toDelete) return
    setDeleting(true)
    try {
      await api.del(`/projects/${pid}/rules/${toDelete.id}`)
      setToDelete(null)
      await load()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Failed to delete rule")
    } finally {
      setDeleting(false)
    }
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-lg font-bold uppercase tracking-widest">Rate Policies</h1>
          <p className="mt-1 text-[10px] uppercase tracking-wider text-muted-foreground">
            Rules matched by route pattern, evaluated atomically in Redis
          </p>
        </div>
        <BrutalButton variant="primary" icon={Plus} onClick={openCreate}>
          New Rule
        </BrutalButton>
      </div>

      <InlineError message={error} />

      {loading ? (
        <Spinner label="LOADING RULES" />
      ) : rules.length === 0 ? (
        <EmptyState
          icon={ShieldAlert}
          title="No rate rules yet"
          hint="Define your first policy — pick an algorithm, a route pattern, and a limit per period."
          action={
            <BrutalButton variant="primary" icon={Plus} onClick={openCreate}>
              Create a rule
            </BrutalButton>
          }
        />
      ) : (
        <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
          {rules.map((r, i) => (
            <motion.div
              key={r.id}
              initial={{ opacity: 0, y: 12 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.25, delay: i * 0.04 }}
            >
              <Panel className="flex h-full flex-col">
                <PanelHeader
                  title={r.name}
                  subtitle={r.route_pattern}
                  icon={ShieldAlert}
                  action={<StatusBadge status={r.is_active ? "active" : "inactive"} />}
                />
                <div className="grid grid-cols-4 gap-px bg-foreground/10">
                  <div className="bg-background p-3">
                    <Label>Algorithm</Label>
                    <div className="mt-1 text-xs font-bold text-[#ea580c]">{ALGO_LABEL[r.algorithm]}</div>
                  </div>
                  <div className="bg-background p-3">
                    <Label>Limit / Period</Label>
                    <div className="mt-1 text-xs font-bold tabular-nums">
                      {r.limit} / {r.period}s
                    </div>
                  </div>
                  <div className="bg-background p-3">
                    <Label>Burst</Label>
                    <div className="mt-1 text-xs font-bold tabular-nums">{r.burst}</div>
                  </div>
                  <div className="bg-background p-3">
                    <Label>Granularity</Label>
                    <div className="mt-1 text-xs font-bold text-foreground">
                      {r.key_strategy === "ip"
                        ? "Client IP"
                        : r.key_strategy?.startsWith("header:")
                          ? `Header (${r.key_strategy.substring(7)})`
                          : "API Key"}
                    </div>
                  </div>
                </div>
                <div className="mt-auto flex items-center gap-2 border-t-2 border-foreground p-3">
                  <BrutalButton variant="outline" icon={Power} onClick={() => toggleActive(r)}>
                    {r.is_active ? "Disable" : "Enable"}
                  </BrutalButton>
                  <BrutalButton variant="outline" icon={Pencil} onClick={() => openEdit(r)}>
                    Edit
                  </BrutalButton>
                  <BrutalButton
                    variant="danger"
                    icon={Trash2}
                    className="ml-auto"
                    aria-label="Delete rule"
                    onClick={() => setToDelete(r)}
                  />
                </div>
              </Panel>
            </motion.div>
          ))}
        </div>
      )}

      {/* Create / edit modal */}
      <Modal open={formOpen} onClose={() => setFormOpen(false)} title={editing ? "Edit Rule" : "Create Rule"}>
        <form onSubmit={submit} className="flex flex-col gap-4">
          <InlineError message={formError} />
          <Field
            label="Rule Name"
            value={form.name}
            onChange={(e) => setForm({ ...form, name: e.target.value })}
            placeholder="Public API throttle"
            autoFocus
          />
          <Field
            label="Route Pattern"
            value={form.route_pattern}
            onChange={(e) => setForm({ ...form, route_pattern: e.target.value })}
            placeholder="/users/*"
            hint="Matched against the gateway sub-path."
          />
          <SelectField
            label="Algorithm"
            value={form.algorithm}
            onChange={(e) => setForm({ ...form, algorithm: e.target.value as Algorithm })}
            hint={ALGORITHMS.find((a) => a.value === form.algorithm)?.blurb}
          >
            {ALGORITHMS.map((a) => (
              <option key={a.value} value={a.value}>
                {a.label}
              </option>
            ))}
          </SelectField>
          
          <SelectField
            label="Limit Strategy"
            value={form.key_strategy === "ip" ? "ip" : form.key_strategy.startsWith("header:") ? "header" : "api_key"}
            onChange={(e) => {
              const val = e.target.value
              if (val === "header") {
                setForm({ ...form, key_strategy: "header:X-Client-ID" })
              } else {
                setForm({ ...form, key_strategy: val })
              }
            }}
            hint="Determine how client requests are bucketed together."
          >
            <option value="api_key">API Key (one counter per key)</option>
            <option value="ip">Client IP (one counter per IP address)</option>
            <option value="header">Custom HTTP Header</option>
          </SelectField>

          {form.key_strategy.startsWith("header:") && (
            <Field
              label="HTTP Header Name"
              value={form.key_strategy.substring(7)}
              onChange={(e) => setForm({ ...form, key_strategy: `header:${e.target.value}` })}
              placeholder="X-Client-ID"
              hint="Clients must pass this header to be limited. Falls back to API Key if missing."
            />
          )}
          <div className="grid grid-cols-3 gap-3">
            <Field
              label="Limit"
              type="number"
              min={1}
              value={form.limit}
              onChange={(e) => setForm({ ...form, limit: e.target.value })}
            />
            <Field
              label="Period (s)"
              type="number"
              min={1}
              value={form.period}
              onChange={(e) => setForm({ ...form, period: e.target.value })}
            />
            <Field
              label="Burst"
              type="number"
              min={0}
              value={form.burst}
              onChange={(e) => setForm({ ...form, burst: e.target.value })}
            />
          </div>
          <SubmitButton loading={saving}>{editing ? "SAVE CHANGES" : "CREATE RULE"}</SubmitButton>
        </form>
      </Modal>

      {/* Delete confirm */}
      <Modal open={!!toDelete} onClose={() => setToDelete(null)} title="Delete Rule">
        <div className="flex flex-col gap-4">
          <p className="text-xs uppercase tracking-wider text-muted-foreground">
            Delete <span className="font-bold text-foreground">{toDelete?.name}</span>? Traffic matching
            this route will no longer be throttled by it.
          </p>
          <div className="flex gap-2">
            <BrutalButton variant="outline" className="flex-1" onClick={() => setToDelete(null)}>
              Cancel
            </BrutalButton>
            <BrutalButton variant="danger" className="flex-1" loading={deleting} onClick={handleDelete}>
              Delete
            </BrutalButton>
          </div>
        </div>
      </Modal>
    </div>
  )
}

export default function PoliciesPage() {
  return <RequireProject>{(project) => <PoliciesInner project={project} />}</RequireProject>
}

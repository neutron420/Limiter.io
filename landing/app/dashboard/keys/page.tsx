"use client"

import * as React from "react"
import { motion } from "framer-motion"
import { Plus, Key, Copy, Check, RefreshCw, Ban, Trash2, ShieldCheck } from "lucide-react"

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
  Label,
  SubmitButton,
} from "@/components/dashboard/kit"
import { RequireProject } from "@/components/dashboard/require-project"
import { api, ApiError } from "@/lib/api"
import type { ApiKey, Project } from "@/lib/types"

function keyState(k: ApiKey): string {
  if (k.revoked_at) return "revoked"
  if (k.expires_at && new Date(k.expires_at) < new Date()) return "expired"
  return "active"
}

function CopyRow({ value }: { value: string }) {
  const [copied, setCopied] = React.useState(false)
  return (
    <div className="flex items-stretch border-2 border-foreground">
      <code className="flex-1 overflow-x-auto whitespace-nowrap bg-muted/10 p-3 text-xs">{value}</code>
      <button
        type="button"
        onClick={() => {
          navigator.clipboard.writeText(value)
          setCopied(true)
          setTimeout(() => setCopied(false), 1500)
        }}
        className="flex items-center gap-1 border-l-2 border-foreground bg-[#ea580c] px-3 text-xs font-bold uppercase text-white"
      >
        {copied ? <Check size={13} /> : <Copy size={13} />}
        {copied ? "Copied" : "Copy"}
      </button>
    </div>
  )
}

function KeysInner({ project }: { project: Project }) {
  const pid = project.id
  const [keys, setKeys] = React.useState<ApiKey[]>([])
  const [loading, setLoading] = React.useState(true)
  const [error, setError] = React.useState<string | null>(null)

  const [createOpen, setCreateOpen] = React.useState(false)
  const [name, setName] = React.useState("")
  const [saving, setSaving] = React.useState(false)
  const [formError, setFormError] = React.useState<string | null>(null)

  // Secret shown once after create/rotate.
  const [revealed, setRevealed] = React.useState<{ name: string; key: string } | null>(null)
  const [confirm, setConfirm] = React.useState<{ key: ApiKey; action: "revoke" | "delete" } | null>(null)
  const [busy, setBusy] = React.useState(false)

  const load = React.useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const list = await api.get<ApiKey[]>(`/projects/${pid}/keys`)
      setKeys(list ?? [])
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load keys")
    } finally {
      setLoading(false)
    }
  }, [pid])

  React.useEffect(() => {
    load()
  }, [load])

  const create = async (e: React.FormEvent) => {
    e.preventDefault()
    setFormError(null)
    if (name.trim().length < 3) return setFormError("Name must be at least 3 characters")
    setSaving(true)
    try {
      const created = await api.post<ApiKey>(`/projects/${pid}/keys`, { name })
      setCreateOpen(false)
      setName("")
      if (created?.key) setRevealed({ name: created.name, key: created.key })
      await load()
    } catch (err) {
      setFormError(err instanceof ApiError ? err.message : "Failed to create key")
    } finally {
      setSaving(false)
    }
  }

  const rotate = async (k: ApiKey) => {
    setBusy(true)
    try {
      const res = await api.post<ApiKey>(`/projects/${pid}/keys/${k.id}/rotate`)
      if (res?.key) setRevealed({ name: res.name ?? k.name, key: res.key })
      await load()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Failed to rotate key")
    } finally {
      setBusy(false)
    }
  }

  const doConfirm = async () => {
    if (!confirm) return
    setBusy(true)
    try {
      if (confirm.action === "revoke") await api.post(`/projects/${pid}/keys/${confirm.key.id}/revoke`)
      else await api.del(`/projects/${pid}/keys/${confirm.key.id}`)
      setConfirm(null)
      await load()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Action failed")
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-lg font-bold uppercase tracking-widest">API Keys</h1>
          <p className="mt-1 text-[10px] uppercase tracking-wider text-muted-foreground">
            Secrets developers send as X-API-Key to the gateway
          </p>
        </div>
        <BrutalButton variant="primary" icon={Plus} onClick={() => setCreateOpen(true)}>
          New Key
        </BrutalButton>
      </div>

      <InlineError message={error} />

      {loading ? (
        <Spinner label="LOADING KEYS" />
      ) : keys.length === 0 ? (
        <EmptyState
          icon={Key}
          title="No API keys yet"
          hint="Create a key to authenticate requests to the rate-limiter gateway."
          action={
            <BrutalButton variant="primary" icon={Plus} onClick={() => setCreateOpen(true)}>
              Create a key
            </BrutalButton>
          }
        />
      ) : (
        <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
          {keys.map((k, i) => {
            const st = keyState(k)
            const disabled = st !== "active"
            return (
              <motion.div
                key={k.id}
                initial={{ opacity: 0, y: 12 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.25, delay: i * 0.04 }}
              >
                <Panel className="flex h-full flex-col">
                  <PanelHeader title={k.name} icon={Key} action={<StatusBadge status={st} />} />
                  <div className="flex flex-col gap-3 p-4">
                    <div>
                      <Label>Key Prefix</Label>
                      <code className="mt-1 block text-xs font-bold text-[#ea580c]">{k.prefix}••••••••</code>
                    </div>
                    <div className="grid grid-cols-2 gap-3">
                      <div>
                        <Label>Last Used</Label>
                        <div className="mt-1 text-xs">
                          {k.last_used_at ? new Date(k.last_used_at).toLocaleString() : "Never"}
                        </div>
                      </div>
                      <div>
                        <Label>Expires</Label>
                        <div className="mt-1 text-xs">
                          {k.expires_at ? new Date(k.expires_at).toLocaleDateString() : "Never"}
                        </div>
                      </div>
                    </div>
                  </div>
                  <div className="mt-auto flex items-center gap-2 border-t-2 border-foreground p-3">
                    <BrutalButton
                      variant="outline"
                      icon={RefreshCw}
                      disabled={disabled || busy}
                      onClick={() => rotate(k)}
                    >
                      Rotate
                    </BrutalButton>
                    {!k.revoked_at && (
                      <BrutalButton
                        variant="outline"
                        icon={Ban}
                        disabled={busy}
                        onClick={() => setConfirm({ key: k, action: "revoke" })}
                      >
                        Revoke
                      </BrutalButton>
                    )}
                    <BrutalButton
                      variant="danger"
                      icon={Trash2}
                      className="ml-auto"
                      aria-label="Delete key"
                      disabled={busy}
                      onClick={() => setConfirm({ key: k, action: "delete" })}
                    />
                  </div>
                </Panel>
              </motion.div>
            )
          })}
        </div>
      )}

      {/* Create */}
      <Modal open={createOpen} onClose={() => setCreateOpen(false)} title="Create API Key">
        <form onSubmit={create} className="flex flex-col gap-4">
          <InlineError message={formError} />
          <Field
            label="Key Name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Backend service"
            hint="A label to recognise this key. 3–100 chars."
            autoFocus
          />
          <SubmitButton loading={saving}>GENERATE KEY</SubmitButton>
        </form>
      </Modal>

      {/* Reveal secret once */}
      <Modal open={!!revealed} onClose={() => setRevealed(null)} title="Copy Your API Key">
        <div className="flex flex-col gap-4">
          <div className="flex items-center gap-2 text-xs font-bold uppercase tracking-wider text-[#ea580c]">
            <ShieldCheck size={15} />
            Shown only once
          </div>
          <p className="text-[11px] uppercase tracking-wider text-muted-foreground">
            Store <span className="text-foreground">{revealed?.name}</span> now — you won&apos;t be able to
            see the full secret again.
          </p>
          {revealed && <CopyRow value={revealed.key} />}
          <BrutalButton variant="primary" className="w-full justify-center" onClick={() => setRevealed(null)}>
            I&apos;ve saved it
          </BrutalButton>
        </div>
      </Modal>

      {/* Revoke / delete confirm */}
      <Modal open={!!confirm} onClose={() => setConfirm(null)} title={confirm?.action === "revoke" ? "Revoke Key" : "Delete Key"}>
        <div className="flex flex-col gap-4">
          <p className="text-xs uppercase tracking-wider text-muted-foreground">
            {confirm?.action === "revoke"
              ? "Revoking immediately blocks all requests using this key. This cannot be undone."
              : "Permanently delete this key. Any client using it will be rejected."}
          </p>
          <div className="flex gap-2">
            <BrutalButton variant="outline" className="flex-1" onClick={() => setConfirm(null)}>
              Cancel
            </BrutalButton>
            <BrutalButton variant="danger" className="flex-1" loading={busy} onClick={doConfirm}>
              {confirm?.action === "revoke" ? "Revoke" : "Delete"}
            </BrutalButton>
          </div>
        </div>
      </Modal>
    </div>
  )
}

export default function KeysPage() {
  return <RequireProject>{(project) => <KeysInner project={project} />}</RequireProject>
}

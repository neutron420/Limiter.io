"use client"

import * as React from "react"
import { motion } from "framer-motion"
import { Save, Plus, Trash2, BadgeCheck, Users, UserPlus, Shield } from "lucide-react"

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
import type { Organization, OrganizationMember } from "@/lib/types"

export default function OrganizationPage() {
  const { user } = useAuth()

  const [org, setOrg] = React.useState<Organization | null>(null)
  const [members, setMembers] = React.useState<OrganizationMember[]>([])
  const [loading, setLoading] = React.useState(true)
  const [error, setError] = React.useState("")
  const [saving, setSaving] = React.useState(false)
  const [showCreate, setShowCreate] = React.useState(false)
  const [showInvite, setShowInvite] = React.useState(false)

  const [name, setName] = React.useState("")
  const [slug, setSlug] = React.useState("")
  const [description, setDescription] = React.useState("")
  const [inviteEmail, setInviteEmail] = React.useState("")
  const [inviteRole, setInviteRole] = React.useState("member")

  const fetchOrg = React.useCallback(async () => {
    setLoading(true)
    try {
      const data = await api.get<Organization[]>("/organizations")
      if (data.length > 0) {
        setOrg(data[0])
        const m = await api.get<OrganizationMember[]>(`/organizations/${data[0].id}/members`)
        setMembers(m)
      }
    } catch {
      // no org yet
    } finally {
      setLoading(false)
    }
  }, [])

  React.useEffect(() => { fetchOrg() }, [fetchOrg])

  const handleCreate = async () => {
    setSaving(true)
    setError("")
    try {
      const data = await api.post<Organization>("/organizations", { name, slug, description })
      setOrg(data)
      setShowCreate(false)
      fetchOrg()
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Failed to create organization")
    } finally {
      setSaving(false)
    }
  }

  const handleInvite = async () => {
    if (!org) return
    setSaving(true)
    setError("")
    try {
      await api.post(`/organizations/${org.id}/members`, { user_id: inviteEmail, role: inviteRole })
      setShowInvite(false)
      setInviteEmail("")
      fetchOrg()
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Failed to invite member")
    } finally {
      setSaving(false)
    }
  }

  const handleRemoveMember = async (memberId: string) => {
    if (!org) return
    try {
      await api.del(`/organizations/${org.id}/members/${memberId}`)
      fetchOrg()
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Failed to remove member")
    }
  }

  return (
    <motion.div initial={{ opacity: 0, y: 12 }} animate={{ opacity: 1, y: 0 }} className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold font-heading tracking-tight">ORGANIZATION</h1>
          <p className="text-sm text-muted-foreground mt-1">Manage your team and workspace</p>
        </div>
        {!org && (
          <BrutalButton onClick={() => setShowCreate(true)} icon={Plus}>CREATE ORGANIZATION</BrutalButton>
        )}
      </div>

      {loading ? (
        <Spinner label="LOADING ORGANIZATION" />
      ) : !org ? (
        <EmptyState icon={BadgeCheck} title="No organization" hint="Create an organization to collaborate with your team" />
      ) : (
        <>
          <Panel>
            <PanelHeader icon={BadgeCheck} title={org.name} />
            <div className="space-y-2 text-sm font-mono">
              <p><span className="text-muted-foreground">Slug:</span> {org.slug}</p>
              <p><span className="text-muted-foreground">Plan:</span> {org.plan}</p>
              {org.description && <p><span className="text-muted-foreground">Description:</span> {org.description}</p>}
            </div>
          </Panel>

          <Panel>
            <PanelHeader icon={Users} title="Members" action={
              <BrutalButton onClick={() => setShowInvite(true)} icon={UserPlus}>INVITE</BrutalButton>
            } />
            {members.length === 0 ? (
              <p className="text-sm text-muted-foreground">No members yet</p>
            ) : (
              <div className="space-y-2">
                {members.map((m) => (
                  <div key={m.id} className="flex items-center justify-between rounded-base border-2 border-foreground p-3">
                    <div className="flex items-center gap-3">
                      <Shield className="size-4 text-muted-foreground" />
                      <div>
                        <p className="text-sm font-mono font-bold">{m.user_id}</p>
                        <p className="text-xs text-muted-foreground">{m.role}</p>
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      <StatusBadge status={m.role} />
                      {m.role !== "owner" && (
                        <BrutalButton onClick={() => handleRemoveMember(m.id)} variant="danger">
                          <Trash2 className="size-4" />
                        </BrutalButton>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </Panel>
        </>
      )}

      {error && <InlineError message={error} />}

      <Modal open={showCreate} onClose={() => setShowCreate(false)} title="Create Organization">
        <div className="space-y-4">
          <Field label="Name">
            <input className="w-full rounded-base border-2 border-foreground bg-main px-3 py-2 font-mono text-sm" value={name} onChange={(e) => setName(e.target.value)} />
          </Field>
          <Field label="Slug">
            <input className="w-full rounded-base border-2 border-foreground bg-main px-3 py-2 font-mono text-sm" value={slug} onChange={(e) => setSlug(e.target.value)} />
          </Field>
          <Field label="Description">
            <textarea className="w-full rounded-base border-2 border-foreground bg-main px-3 py-2 font-mono text-sm" value={description} onChange={(e) => setDescription(e.target.value)} />
          </Field>
          <BrutalButton onClick={handleCreate} loading={saving} icon={Save}>CREATE</BrutalButton>
        </div>
      </Modal>

      <Modal open={showInvite} onClose={() => setShowInvite(false)} title="Invite Member">
        <div className="space-y-4">
          <Field label="Email">
            <input className="w-full rounded-base border-2 border-foreground bg-main px-3 py-2 font-mono text-sm" type="email" value={inviteEmail} onChange={(e) => setInviteEmail(e.target.value)} />
          </Field>
          <Field label="Role">
            <select className="w-full rounded-base border-2 border-foreground bg-main px-3 py-2 font-mono text-sm" value={inviteRole} onChange={(e) => setInviteRole(e.target.value)}>
              <option value="member">Member</option>
              <option value="admin">Admin</option>
            </select>
          </Field>
          <BrutalButton onClick={handleInvite} loading={saving} icon={UserPlus}>INVITE</BrutalButton>
        </div>
      </Modal>
    </motion.div>
  )
}

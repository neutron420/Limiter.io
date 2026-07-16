"use client"

import * as React from "react"
import { Activity, Check, KeyRound, LogOut, RefreshCw, Send, Trash2, UserCog, Users } from "lucide-react"

import {
  Panel,
  PanelHeader,
  BrutalButton,
  Field,
  SelectField,
  InlineError,
  Label,
  SubmitButton,
  Spinner,
} from "@/components/dashboard/kit"
import { api, ApiError } from "@/lib/api"
import { useAuth } from "@/lib/auth"
import { useProject } from "@/lib/project-context"
import { canWrite } from "@/lib/rbac"

type Member = {
  id: string
  user_id: string
  email: string
  role: "admin" | "member"
  created_at: string
}

type Invite = {
  id: string
  email: string
  role: "admin" | "member"
  expires_at: string
}

type AuditEvent = {
  id: string
  action: string
  actor_id: string
  target_type: string
  target_id: string
  metadata?: Record<string, unknown>
  created_at: string
}

function actionLabel(action: string) {
  return action.replace(/[._]/g, " ").toUpperCase()
}

function upsertInvite(list: Invite[], invite: Invite) {
  const email = invite.email.toLowerCase()
  return [invite, ...list.filter((item) => item.id !== invite.id && item.email.toLowerCase() !== email)]
}

export default function SettingsPage() {
  const { user, logout } = useAuth()
  const { current, role } = useProject()
  const canWriteProject = canWrite(role)

  // Password change state
  const [oldPassword, setOldPassword] = React.useState("")
  const [newPassword, setNewPassword] = React.useState("")
  const [confirm, setConfirm] = React.useState("")
  const [saving, setSaving] = React.useState(false)
  const [error, setError] = React.useState<string | null>(null)
  const [done, setDone] = React.useState(false)

  // Team sharing state
  const [members, setMembers] = React.useState<Member[]>([])
  const [membersLoading, setMembersLoading] = React.useState(false)
  const [updatingMemberId, setUpdatingMemberId] = React.useState<string | null>(null)
  const [invites, setInvites] = React.useState<Invite[]>([])
  const [invitesLoading, setInvitesLoading] = React.useState(false)
  const [resendingInviteId, setResendingInviteId] = React.useState<string | null>(null)
  const [inviteEmail, setInviteEmail] = React.useState("")
  const [inviteRole, setInviteRole] = React.useState<"admin" | "member">("member")
  const [inviteError, setInviteError] = React.useState<string | null>(null)
  const [inviting, setInviting] = React.useState(false)

  // Team activity state
  const [auditEvents, setAuditEvents] = React.useState<AuditEvent[]>([])
  const [auditLoading, setAuditLoading] = React.useState(false)

  const loadMembers = React.useCallback(async () => {
    if (!current) return
    setMembersLoading(true)
    try {
      const list = await api.get<Member[]>(`/projects/${current.id}/members`)
      setMembers(list ?? [])
    } catch {
      // ignore
    } finally {
      setMembersLoading(false)
    }
  }, [current])

  const loadInvites = React.useCallback(async () => {
    if (!current || !canWriteProject) return
    setInvitesLoading(true)
    try {
      const list = await api.get<Invite[]>(`/projects/${current.id}/invites`)
      setInvites(list ?? [])
    } catch {
      // ignore
    } finally {
      setInvitesLoading(false)
    }
  }, [current, canWriteProject])

  const loadAuditEvents = React.useCallback(async () => {
    if (!current || !canWriteProject) return
    setAuditLoading(true)
    try {
      const list = await api.get<AuditEvent[]>(`/projects/${current.id}/audit-events`)
      setAuditEvents(list ?? [])
    } catch {
      // ignore
    } finally {
      setAuditLoading(false)
    }
  }, [current, canWriteProject])

  React.useEffect(() => {
    loadMembers()
    loadInvites()
    loadAuditEvents()
  }, [loadMembers, loadInvites, loadAuditEvents])

  const changePassword = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    setDone(false)
    if (newPassword.length < 8) return setError("New password must be at least 8 characters")
    if (newPassword !== confirm) return setError("New passwords do not match")
    setSaving(true)
    try {
      await api.post("/auth/change-password", { old_password: oldPassword, new_password: newPassword })
      setDone(true)
      setOldPassword("")
      setNewPassword("")
      setConfirm("")
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Failed to change password")
    } finally {
      setSaving(false)
    }
  }

  const inviteMember = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!current) return
    setInviteError(null)
    setInviting(true)
    try {
      const invite = await api.post<Invite>(`/projects/${current.id}/invites`, {
        email: inviteEmail,
        role: inviteRole,
      })
      setInvites((prev) => upsertInvite(prev, invite))
      setInviteEmail("")
      loadAuditEvents()
    } catch (err) {
      setInviteError(err instanceof ApiError ? err.message : "Failed to invite member")
    } finally {
      setInviting(false)
    }
  }

  const resendInvite = async (invite: Invite) => {
    if (!current) return
    setResendingInviteId(invite.id)
    try {
      const nextInvite = await api.post<Invite>(`/projects/${current.id}/invites`, {
        email: invite.email,
        role: invite.role,
      })
      setInvites((prev) => upsertInvite(prev.filter((item) => item.id !== invite.id), nextInvite))
      loadAuditEvents()
    } catch (err) {
      alert(err instanceof ApiError ? err.message : "Failed to resend invite")
    } finally {
      setResendingInviteId(null)
    }
  }

  const revokeInvite = async (inviteId: string) => {
    if (!current) return
    if (typeof window !== "undefined" && !window.confirm("Are you sure you want to revoke this invitation?")) return
    try {
      await api.del(`/projects/${current.id}/invites/${inviteId}`)
      setInvites((prev) => prev.filter((i) => i.id !== inviteId))
      loadAuditEvents()
    } catch (err) {
      alert(err instanceof ApiError ? err.message : "Failed to revoke invite")
    }
  }

  const updateMemberRole = async (memberId: string, nextRole: "admin" | "member") => {
    if (!current) return
    setUpdatingMemberId(memberId)
    try {
      const updated = await api.patch<Member>(`/projects/${current.id}/members/${memberId}`, { role: nextRole })
      setMembers((prev) => prev.map((member) => (member.id === memberId ? updated : member)))
      loadAuditEvents()
    } catch (err) {
      alert(err instanceof ApiError ? err.message : "Failed to update member role")
    } finally {
      setUpdatingMemberId(null)
    }
  }

  const removeMember = async (memberId: string) => {
    if (!current) return
    if (typeof window !== "undefined" && !window.confirm("Are you sure you want to remove this member?")) return
    try {
      await api.del(`/projects/${current.id}/members/${memberId}`)
      setMembers((prev) => prev.filter((m) => m.id !== memberId))
      loadAuditEvents()
    } catch (err) {
      alert(err instanceof ApiError ? err.message : "Failed to remove member")
    }
  }

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h1 className="text-lg font-bold uppercase tracking-widest">Settings</h1>
        <p className="mt-1 text-[10px] uppercase tracking-wider text-muted-foreground">
          Manage your operator account
        </p>
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <div className="flex flex-col gap-6">
          {/* Account */}
          <Panel>
            <PanelHeader title="Account" icon={UserCog} />
            <div className="flex flex-col gap-4 p-4">
              <div>
                <Label>Email</Label>
                <div className="mt-1 text-sm font-bold">{user?.email}</div>
              </div>
              <div>
                <Label>Operator ID</Label>
                <code className="mt-1 block break-all text-xs text-muted-foreground">{user?.userId}</code>
              </div>
              <div className="pt-2">
                <BrutalButton variant="danger" icon={LogOut} onClick={() => logout()}>
                  Log out
                </BrutalButton>
              </div>
            </div>
          </Panel>

          {/* Change password */}
          <Panel>
            <PanelHeader title="Change Password" icon={KeyRound} />
            <form onSubmit={changePassword} className="flex flex-col gap-4 p-4">
              <InlineError message={error} />
              {done && (
                <div className="flex items-center gap-2 border-2 border-green-500 bg-green-500/10 px-3 py-2 text-xs font-bold uppercase tracking-wider text-green-500">
                  <Check size={14} /> Password updated
                </div>
              )}
              <Field
                label="Current Password"
                type="password"
                autoComplete="current-password"
                value={oldPassword}
                onChange={(e) => setOldPassword(e.target.value)}
              />
              <Field
                label="New Password"
                type="password"
                autoComplete="new-password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                hint="Minimum 8 characters."
              />
              <Field
                label="Confirm New Password"
                type="password"
                autoComplete="new-password"
                value={confirm}
                onChange={(e) => setConfirm(e.target.value)}
              />
              <SubmitButton loading={saving}>UPDATE PASSWORD</SubmitButton>
            </form>
          </Panel>
        </div>

        <div className="flex flex-col gap-6">
          {/* Project Team */}
          {current ? (
            <Panel>
              <PanelHeader title="Project Team Access" icon={Users} />
              <div className="flex flex-col gap-4 p-4">
                {/* Members */}
                <div>
                  <Label className="font-bold text-foreground mb-2 block">Members</Label>
                  <div className="flex flex-col gap-2 max-h-[180px] overflow-y-auto pr-1">
                    {membersLoading ? (
                      <Spinner label="LOADING TEAM MEMBERS" />
                    ) : members.length === 0 ? (
                      <p className="text-[11px] text-muted-foreground uppercase py-2">
                        No collaborators added yet. Only the owner has access.
                      </p>
                    ) : (
                      members.map((m) => (
                        <div key={m.id} className="flex justify-between gap-3 border-2 border-foreground p-2 bg-background">
                          <div className="min-w-0 flex-1">
                            <p className="text-xs font-bold truncate">{m.email}</p>
                            <p className="text-[9px] text-[#ea580c] uppercase font-bold">{m.role}</p>
                          </div>
                          {m.user_id !== user?.userId && canWriteProject && (
                            <div className="flex shrink-0 items-center gap-2">
                              <select
                                value={m.role}
                                disabled={updatingMemberId === m.id}
                                onChange={(e) => updateMemberRole(m.id, e.target.value as "admin" | "member")}
                                className="h-8 border-2 border-foreground bg-background px-2 font-mono text-[10px] font-bold uppercase focus:border-[#ea580c] focus:outline-none"
                                aria-label={`Role for ${m.email}`}
                              >
                                <option value="member">MEMBER</option>
                                <option value="admin">ADMIN</option>
                              </select>
                              <BrutalButton
                                variant="danger"
                                className="h-8 w-8 p-0 cursor-pointer"
                                onClick={() => removeMember(m.id)}
                                aria-label={`Remove ${m.email}`}
                              >
                                <Trash2 size={12} />
                              </BrutalButton>
                            </div>
                          )}
                        </div>
                      ))
                    )}
                  </div>
                </div>

                {/* Pending Invites */}
                <div className="border-t-2 border-foreground pt-4">
                  <Label className="font-bold text-foreground mb-2 block">Pending Invites</Label>
                  <div className="flex flex-col gap-2 max-h-[180px] overflow-y-auto pr-1">
                    {invitesLoading ? (
                      <Spinner label="LOADING INVITES" />
                    ) : invites.length === 0 ? (
                      <p className="text-[11px] text-muted-foreground uppercase py-2">No pending invitations.</p>
                    ) : (
                      invites.map((inv) => (
                        <div key={inv.id} className="flex justify-between gap-3 border-2 border-yellow-500 bg-yellow-500/10 p-2">
                          <div className="min-w-0 flex-1">
                            <p className="text-xs font-bold truncate">{inv.email}</p>
                            <div className="flex items-center gap-2">
                              <p className="text-[9px] text-[#ea580c] uppercase font-bold">{inv.role}</p>
                              <span className="border-2 border-yellow-500 text-yellow-600 bg-yellow-500/10 text-[8px] uppercase font-bold px-1.5 py-0.5">
                                PENDING
                              </span>
                            </div>
                            <p className="text-[8px] text-muted-foreground">
                              Expires: {new Date(inv.expires_at).toLocaleDateString()}
                            </p>
                          </div>
                          {canWriteProject && (
                            <div className="flex shrink-0 items-center gap-2">
                              <BrutalButton
                                className="h-8 w-8 p-0 cursor-pointer"
                                loading={resendingInviteId === inv.id}
                                onClick={() => resendInvite(inv)}
                                aria-label={`Resend invite to ${inv.email}`}
                              >
                                <Send size={12} />
                              </BrutalButton>
                              <BrutalButton
                                variant="danger"
                                className="h-8 w-8 p-0 cursor-pointer"
                                onClick={() => revokeInvite(inv.id)}
                                aria-label={`Revoke invite to ${inv.email}`}
                              >
                                <Trash2 size={12} />
                              </BrutalButton>
                            </div>
                          )}
                        </div>
                      ))
                    )}
                  </div>
                </div>

                {/* Invite Form */}
                {canWriteProject && (
                  <div className="border-t-2 border-foreground pt-4 mt-2">
                    <Label className="font-bold text-foreground">Invite Collaborator</Label>
                    <form onSubmit={inviteMember} className="flex flex-col gap-3 mt-2">
                      <InlineError message={inviteError} />
                      <Field
                        label="Collaborator Email"
                        type="email"
                        required
                        value={inviteEmail}
                        onChange={(e) => setInviteEmail(e.target.value)}
                        placeholder="teammate@company.com"
                      />
                      <SelectField
                        label="Role"
                        value={inviteRole}
                        onChange={(e) => setInviteRole(e.target.value as "admin" | "member")}
                      >
                        <option value="member">MEMBER (READ-ONLY)</option>
                        <option value="admin">ADMIN (READ-WRITE)</option>
                      </SelectField>
                      <SubmitButton loading={inviting}>INVITE MEMBER</SubmitButton>
                    </form>
                  </div>
                )}
              </div>
            </Panel>
          ) : (
            <Panel className="border-dashed border-foreground/30 flex items-center justify-center p-8 text-center min-h-[300px]">
              <p className="text-xs uppercase tracking-widest text-muted-foreground animate-pulse">
                Select a project to manage team access...
              </p>
            </Panel>
          )}

          {current && canWriteProject && (
            <Panel>
              <PanelHeader
                title="Team Activity"
                icon={Activity}
                action={
                  <BrutalButton className="h-8 px-2" onClick={() => loadAuditEvents()} aria-label="Refresh team activity">
                    <RefreshCw size={12} />
                  </BrutalButton>
                }
              />
              <div className="flex flex-col gap-2 max-h-[260px] overflow-y-auto p-4">
                {auditLoading ? (
                  <Spinner label="LOADING ACTIVITY" />
                ) : auditEvents.length === 0 ? (
                  <p className="text-[11px] text-muted-foreground uppercase py-2">No team activity recorded yet.</p>
                ) : (
                  auditEvents.map((event) => (
                    <div key={event.id} className="border-2 border-foreground bg-background p-2">
                      <div className="flex items-center justify-between gap-3">
                        <p className="text-[11px] font-bold uppercase tracking-wider">{actionLabel(event.action)}</p>
                        <p className="shrink-0 text-[8px] uppercase text-muted-foreground">
                          {new Date(event.created_at).toLocaleString()}
                        </p>
                      </div>
                      <p className="mt-1 text-[9px] uppercase text-muted-foreground break-all">
                        Actor: {event.actor_id}
                      </p>
                      {event.metadata && (
                        <p className="mt-1 text-[9px] uppercase text-[#ea580c]">
                          {[event.metadata.email, event.metadata.role].filter(Boolean).join(" / ")}
                        </p>
                      )}
                    </div>
                  ))
                )}
              </div>
            </Panel>
          )}
        </div>
      </div>
    </div>
  )
}
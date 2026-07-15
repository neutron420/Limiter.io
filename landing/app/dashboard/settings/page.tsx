"use client"

import * as React from "react"
import { KeyRound, LogOut, UserCog, Check, Users, Trash2 } from "lucide-react"

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
  const [members, setMembers] = React.useState<any[]>([])
  const [membersLoading, setMembersLoading] = React.useState(false)
  const [invites, setInvites] = React.useState<any[]>([])
  const [invitesLoading, setInvitesLoading] = React.useState(false)
  const [inviteEmail, setInviteEmail] = React.useState("")
  const [inviteRole, setInviteRole] = React.useState("member")
  const [inviteError, setInviteError] = React.useState<string | null>(null)
  const [inviting, setInviting] = React.useState(false)

  const loadMembers = React.useCallback(async () => {
    if (!current) return
    setMembersLoading(true)
    try {
      const list = await api.get<any[]>(`/projects/${current.id}/members`)
      setMembers(list ?? [])
    } catch {
      // ignore
    } finally {
      setMembersLoading(false)
    }
  }, [current])

  const loadInvites = React.useCallback(async () => {
    if (!current) return
    setInvitesLoading(true)
    try {
      const list = await api.get<any[]>(`/projects/${current.id}/invites`)
      setInvites(list ?? [])
    } catch {
      // ignore
    } finally {
      setInvitesLoading(false)
    }
  }, [current])

  React.useEffect(() => {
    loadMembers()
    loadInvites()
  }, [loadMembers, loadInvites])

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
      const invite = await api.post<any>(`/projects/${current.id}/invites`, {
        email: inviteEmail,
        role: inviteRole,
      })
      setInvites((prev) => [...prev, invite])
      setInviteEmail("")
    } catch (err) {
      setInviteError(err instanceof ApiError ? err.message : "Failed to invite member")
    } finally {
      setInviting(false)
    }
  }

  const revokeInvite = async (inviteId: string) => {
    if (!current) return
    if (typeof window !== "undefined" && !window.confirm("Are you sure you want to revoke this invitation?")) return
    try {
      await api.del(`/projects/${current.id}/invites/${inviteId}`)
      setInvites((prev) => prev.filter((i) => i.id !== inviteId))
    } catch (err) {
      alert(err instanceof ApiError ? err.message : "Failed to revoke invite")
    }
  }

  const removeMember = async (memberId: string) => {
    if (!current) return
    if (typeof window !== "undefined" && !window.confirm("Are you sure you want to remove this member?")) return
    try {
      await api.del(`/projects/${current.id}/members/${memberId}`)
      setMembers((prev) => prev.filter((m) => m.id !== memberId))
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

        <div>
          {/* Project Team */}
          {current ? (
            <Panel className="h-full">
              <PanelHeader title="Project Team Access" icon={Users} />
              <div className="flex flex-col gap-4 p-4">
                {/* Members */}
                <div>
                  <Label className="font-bold text-foreground mb-2 block">Members</Label>
                  <div className="flex flex-col gap-2 max-h-[150px] overflow-y-auto pr-1">
                    {membersLoading ? (
                      <Spinner label="LOADING TEAM MEMBERS" />
                    ) : members.length === 0 ? (
                      <p className="text-[11px] text-muted-foreground uppercase py-2">
                        No collaborators added yet. Only the owner has access.
                      </p>
                    ) : (
                      members.map((m) => (
                        <div key={m.id} className="flex justify-between items-center border-2 border-foreground p-2 bg-background">
                          <div className="min-w-0">
                            <p className="text-xs font-bold truncate">{m.email}</p>
                            <p className="text-[9px] text-[#ea580c] uppercase font-bold">{m.role}</p>
                          </div>
                          {m.user_id !== user?.userId && canWriteProject && (
                            <BrutalButton
                              variant="danger"
                              className="p-1.5 cursor-pointer"
                              onClick={() => removeMember(m.id)}
                            >
                              <Trash2 size={12} />
                            </BrutalButton>
                          )}
                        </div>
                      ))
                    )}
                  </div>
                </div>

                {/* Pending Invites */}
                <div className="border-t-2 border-foreground pt-4">
                  <Label className="font-bold text-foreground mb-2 block">Pending Invites</Label>
                  <div className="flex flex-col gap-2 max-h-[150px] overflow-y-auto pr-1">
                    {invitesLoading ? (
                      <Spinner label="LOADING INVITES" />
                    ) : invites.length === 0 ? (
                      <p className="text-[11px] text-muted-foreground uppercase py-2">
                        No pending invitations.
                      </p>
                    ) : (
                      invites.map((inv) => (
                        <div key={inv.id} className="flex justify-between items-center border-2 border-yellow-500 bg-yellow-500/10 p-2">
                          <div className="min-w-0">
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
                            <BrutalButton
                              variant="danger"
                              className="p-1.5 cursor-pointer"
                              onClick={() => revokeInvite(inv.id)}
                            >
                              <Trash2 size={12} />
                            </BrutalButton>
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
                        onChange={(e) => setInviteRole(e.target.value)}
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
            <Panel className="border-dashed border-foreground/30 flex items-center justify-center p-8 text-center min-h-[300px] h-full">
              <p className="text-xs uppercase tracking-widest text-muted-foreground animate-pulse">
                Select a project to manage team access...
              </p>
            </Panel>
          )}
        </div>
      </div>
    </div>
  )
}

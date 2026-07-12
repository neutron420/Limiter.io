"use client"

import * as React from "react"
import { KeyRound, LogOut, UserCog, Check } from "lucide-react"

import {
  Panel,
  PanelHeader,
  BrutalButton,
  Field,
  InlineError,
  Label,
  SubmitButton,
} from "@/components/dashboard/kit"
import { api, ApiError } from "@/lib/api"
import { useAuth } from "@/lib/auth"

export default function SettingsPage() {
  const { user, logout } = useAuth()

  const [oldPassword, setOldPassword] = React.useState("")
  const [newPassword, setNewPassword] = React.useState("")
  const [confirm, setConfirm] = React.useState("")
  const [saving, setSaving] = React.useState(false)
  const [error, setError] = React.useState<string | null>(null)
  const [done, setDone] = React.useState(false)

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

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h1 className="text-lg font-bold uppercase tracking-widest">Settings</h1>
        <p className="mt-1 text-[10px] uppercase tracking-wider text-muted-foreground">
          Manage your operator account
        </p>
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
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
    </div>
  )
}

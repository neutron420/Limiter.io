"use client"

import * as React from "react"
import { motion } from "framer-motion"
import { Save, Bell, Mail, Slack, RotateCcw } from "lucide-react"

import {
  Panel,
  PanelHeader,
  BrutalButton,
  InlineError,
  Spinner,
  SubmitButton,
} from "@/components/dashboard/kit"
import { RequireProject } from "@/components/dashboard/require-project"
import { api, ApiError } from "@/lib/api"
import { useProject } from "@/lib/project-context"
import { canWrite } from "@/lib/rbac"
import type { NotificationPreferences } from "@/lib/types"

export default function NotificationsPage() {
  const { current, role } = useProject()
  const canWriteProject = canWrite(role)

  const [prefs, setPrefs] = React.useState<NotificationPreferences | null>(null)
  const [loading, setLoading] = React.useState(true)
  const [saving, setSaving] = React.useState(false)
  const [error, setError] = React.useState("")
  const [success, setSuccess] = React.useState("")

  const [emailNotif, setEmailNotif] = React.useState(true)
  const [slackNotif, setSlackNotif] = React.useState(false)
  const [slackWebhook, setSlackWebhook] = React.useState("")
  const [rateLimitAlerts, setRateLimitAlerts] = React.useState(true)
  const [memberAlerts, setMemberAlerts] = React.useState(true)
  const [keyRotationAlerts, setKeyRotationAlerts] = React.useState(true)
  const [weeklyDigest, setWeeklyDigest] = React.useState(false)

  const fetchPrefs = React.useCallback(async () => {
    if (!current?.id) return
    setLoading(true)
    try {
      const data = await api.get<NotificationPreferences>(`/projects/${current.id}/notification-preferences`)
      setPrefs(data)
      setEmailNotif(data.email_notifications)
      setSlackNotif(data.slack_notifications)
      setSlackWebhook(data.slack_webhook_url || "")
      setRateLimitAlerts(data.rate_limit_alerts)
      setMemberAlerts(data.member_join_alerts)
      setKeyRotationAlerts(data.key_rotation_alerts)
      setWeeklyDigest(data.weekly_digest)
    } catch {
      // defaults
    } finally {
      setLoading(false)
    }
  }, [current?.id])

  React.useEffect(() => { fetchPrefs() }, [fetchPrefs])

  const handleSave = async () => {
    if (!current?.id) return
    setSaving(true)
    setError("")
    setSuccess("")
    try {
      await api.put(`/projects/${current.id}/notification-preferences`, {
        email_notifications: emailNotif,
        slack_notifications: slackNotif,
        slack_webhook_url: slackWebhook,
        rate_limit_alerts: rateLimitAlerts,
        member_join_alerts: memberAlerts,
        key_rotation_alerts: keyRotationAlerts,
        weekly_digest: weeklyDigest,
      })
      setSuccess("Notification preferences saved")
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Failed to save preferences")
    } finally {
      setSaving(false)
    }
  }

  const Toggle = ({ label, sub, checked, onChange, disabled }: { label: string; sub?: string; checked: boolean; onChange: (v: boolean) => void; disabled?: boolean }) => (
    <label className={`flex items-center justify-between rounded-base border-2 border-foreground p-3 cursor-pointer ${disabled ? "opacity-50" : ""}`}>
      <div>
        <p className="text-sm font-mono font-bold">{label}</p>
        {sub && <p className="text-xs text-muted-foreground">{sub}</p>}
      </div>
      <input type="checkbox" checked={checked} onChange={(e) => onChange(e.target.checked)} disabled={disabled} className="size-5 accent-foreground" />
    </label>
  )

  return (
    <RequireProject>
      <motion.div initial={{ opacity: 0, y: 12 }} animate={{ opacity: 1, y: 0 }} className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold font-heading tracking-tight">NOTIFICATIONS</h1>
            <p className="text-sm text-muted-foreground mt-1">Configure alert channels and preferences</p>
          </div>
        </div>

        {loading ? (
          <Spinner label="LOADING PREFERENCES" />
        ) : (
          <>
            <Panel>
              <PanelHeader icon={Bell} title="Alert Channels" />
              <div className="space-y-3">
                <Toggle label="Email Notifications" sub="Receive alerts via email" checked={emailNotif} onChange={setEmailNotif} disabled={!canWriteProject} />
                <Toggle label="Slack Notifications" sub="Receive alerts in Slack" checked={slackNotif} onChange={setSlackNotif} disabled={!canWriteProject} />
                {slackNotif && (
                  <input
                    className="w-full rounded-base border-2 border-foreground bg-main px-3 py-2 font-mono text-sm"
                    placeholder="https://hooks.slack.com/services/..."
                    value={slackWebhook}
                    onChange={(e) => setSlackWebhook(e.target.value)}
                    disabled={!canWriteProject}
                  />
                )}
              </div>
            </Panel>

            <Panel>
              <PanelHeader icon={Mail} title="Alert Types" />
              <div className="space-y-3">
                <Toggle label="Rate Limit Alerts" sub="When a project approaches its limit" checked={rateLimitAlerts} onChange={setRateLimitAlerts} disabled={!canWriteProject} />
                <Toggle label="Member Join Alerts" sub="When a new member joins" checked={memberAlerts} onChange={setMemberAlerts} disabled={!canWriteProject} />
                <Toggle label="Key Rotation Reminders" sub="When API keys need rotation" checked={keyRotationAlerts} onChange={setKeyRotationAlerts} disabled={!canWriteProject} />
                <Toggle label="Weekly Digest" sub="Weekly summary of usage" checked={weeklyDigest} onChange={setWeeklyDigest} disabled={!canWriteProject} />
              </div>
            </Panel>

            {error && <InlineError message={error} />}
            {success && <p className="text-sm text-green-600 font-mono">{success}</p>}

            {canWriteProject && (
              <div className="flex gap-2">
                <SubmitButton onClick={handleSave} loading={saving} icon={Save}>SAVE PREFERENCES</SubmitButton>
                <BrutalButton onClick={fetchPrefs} variant="secondary">
                  <RotateCcw className="size-4" />
                  RESET
                </BrutalButton>
              </div>
            )}
          </>
        )}
      </motion.div>
    </RequireProject>
  )
}

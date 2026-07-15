"use client"

import * as React from "react"
import { useRouter, useSearchParams } from "next/navigation"
import { Check, X, AlertCircle } from "lucide-react"

import { Panel, BrutalButton, Spinner } from "@/components/dashboard/kit"
import { api, ApiError } from "@/lib/api"
import { useAuth } from "@/lib/auth"
import { useProject } from "@/lib/project-context"
import type { AcceptInviteResponse } from "@/lib/types"

export default function AcceptInvitePage() {
  return (
    <React.Suspense
      fallback={
        <div className="min-h-screen flex items-center justify-center bg-background">
          <Spinner label="LOADING..." />
        </div>
      }
    >
      <AcceptInviteContent />
    </React.Suspense>
  )
}

function AcceptInviteContent() {
  const { user, ready } = useAuth()
  const { select, refresh } = useProject()
  const router = useRouter()
  const searchParams = useSearchParams()
  const token = searchParams.get("token")

  const [loading, setLoading] = React.useState(false)
  const [error, setError] = React.useState<string | null>(null)
  const [success, setSuccess] = React.useState(false)
  const [inviteData, setInviteData] = React.useState<AcceptInviteResponse | null>(null)

  // Redirect to login if not authenticated
  React.useEffect(() => {
    if (ready && !user) {
      const next = encodeURIComponent(`/accept-invite?token=${token || ""}`)
      router.push(`/login?next=${next}`)
    }
  }, [ready, user, token, router])

  // Redirect if no token
  React.useEffect(() => {
    if (ready && !token) {
      setError("No invite token provided. Please check your invitation link.")
    }
  }, [ready, token])

  const acceptInvite = async () => {
    if (!token) return
    setLoading(true)
    setError(null)
    try {
      const resp = await api.post<AcceptInviteResponse>("/invites/accept", { token })
      setInviteData(resp)
      setSuccess(true)

      // Refresh projects and switch to the new project
      await refresh()
      if (resp?.project_id) {
        select(resp.project_id)
      }

      // Redirect to dashboard after a short delay
      setTimeout(() => {
        router.push(resp.role === "member" ? "/member-dashboard" : "/dashboard")
      }, 1500)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Failed to accept invite")
    } finally {
      setLoading(false)
    }
  }

  if (!ready || !user) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <Spinner label="LOADING..." />
      </div>
    )
  }

  if (!token) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background p-4">
        <Panel className="max-w-md w-full">
          <div className="flex items-center gap-3 text-red-500 mb-4">
            <X size={24} />
            <h1 className="text-lg font-bold uppercase tracking-widest">Invalid Invite Link</h1>
          </div>
          <p className="text-sm text-muted-foreground mb-6">
            No invite token was provided. Please check your invitation link and try again.
          </p>
          <BrutalButton onClick={() => router.push("/dashboard")} className="w-full">
            Go to Dashboard
          </BrutalButton>
        </Panel>
      </div>
    )
  }

  if (success && inviteData) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background p-4">
        <Panel className="max-w-md w-full">
          <div className="flex items-center gap-3 text-green-500 mb-4">
            <Check size={24} />
            <h1 className="text-lg font-bold uppercase tracking-widest">Invite Accepted</h1>
          </div>
          <p className="text-sm mb-2">
            You have successfully joined <strong>{inviteData.project_name}</strong> as a{" "}
            <strong className="text-[#ea580c]">{inviteData.role}</strong>.
          </p>
          <p className="text-xs text-muted-foreground mb-6">Redirecting to dashboard...</p>
          <Spinner label="LOADING..." />
        </Panel>
      </div>
    )
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-background p-4">
      <Panel className="max-w-md w-full">
        <div className="flex items-center gap-3 text-[#ea580c] mb-4">
          <AlertCircle size={24} />
          <h1 className="text-lg font-bold uppercase tracking-widest">Project Invitation</h1>
        </div>

        {error && (
          <div className="mb-4 border-2 border-red-500 bg-red-500/10 p-3">
            <p className="text-xs font-bold text-red-500 uppercase tracking-wider">{error}</p>
          </div>
        )}

        <p className="text-sm mb-6">
          You have been invited to join a project. Click the button below to accept the invitation.
        </p>

        <div className="flex gap-3">
          <BrutalButton
            onClick={acceptInvite}
            loading={loading}
            disabled={!!error}
            className="flex-1"
          >
            Accept Invite
          </BrutalButton>
          <BrutalButton
            variant="outline"
            onClick={() => router.push("/dashboard")}
            disabled={loading}
          >
            Cancel
          </BrutalButton>
        </div>
      </Panel>
    </div>
  )
}

"use client"

import * as React from "react"
import { motion } from "framer-motion"
import { Plus, Trash2, Key, Smartphone, RefreshCw } from "lucide-react"

import {
  Panel,
  PanelHeader,
  BrutalButton,
  InlineError,
  Spinner,
  EmptyState,
  StatusBadge,
} from "@/components/dashboard/kit"
import { api, ApiError } from "@/lib/api"
import type { Passkey } from "@/lib/types"

export default function PasskeysPage() {
  const [passkeys, setPasskeys] = React.useState<Passkey[]>([])
  const [loading, setLoading] = React.useState(true)
  const [error, setError] = React.useState("")
  const [registering, setRegistering] = React.useState(false)

  const fetchPasskeys = React.useCallback(async () => {
    setLoading(true)
    try {
      const data = await api.get<Passkey[]>("/auth/passkeys")
      setPasskeys(data)
    } catch {
      setPasskeys([])
    } finally {
      setLoading(false)
    }
  }, [])

  React.useEffect(() => { fetchPasskeys() }, [fetchPasskeys])

  const handleRegister = async () => {
    setRegistering(true)
    setError("")
    try {
      const opts = await api.post<any>("/auth/passkeys/register/begin")
      const cred = await navigator.credentials.create({ publicKey: opts })
      if (!cred) { throw new Error("Passkey creation cancelled") }
      await api.post("/auth/passkeys/register/complete", { credential: cred })
      fetchPasskeys()
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Failed to register passkey")
    } finally {
      setRegistering(false)
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await api.del(`/auth/passkeys/${id}`)
      fetchPasskeys()
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Failed to delete passkey")
    }
  }

  return (
    <motion.div initial={{ opacity: 0, y: 12 }} animate={{ opacity: 1, y: 0 }} className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold font-heading tracking-tight">PASSKEYS</h1>
          <p className="text-sm text-muted-foreground mt-1">Passwordless authentication with WebAuthn</p>
        </div>
        <BrutalButton onClick={handleRegister} loading={registering} icon={Plus} disabled={registering}>
          REGISTER PASSKEY
        </BrutalButton>
      </div>

      {loading ? (
        <Spinner label="LOADING PASSKEYS" />
      ) : passkeys.length === 0 ? (
        <EmptyState icon={Key} title="No passkeys registered" hint="Register a passkey for passwordless login" />
      ) : (
        <div className="space-y-3">
          {passkeys.map((pk) => (
            <Panel key={pk.id}>
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <Smartphone className="size-5 text-muted-foreground" />
                  <div>
                    <p className="font-mono text-sm font-bold">{pk.nickname}</p>
                    <p className="text-xs text-muted-foreground">
                      Created: {new Date(pk.created_at).toLocaleDateString()}
                      {pk.last_used_at ? ` · Last used: ${new Date(pk.last_used_at).toLocaleDateString()}` : " · Never used"}
                    </p>
                  </div>
                </div>
                <BrutalButton onClick={() => handleDelete(pk.id)} variant="danger">
                  <Trash2 className="size-4" />
                </BrutalButton>
              </div>
            </Panel>
          ))}
        </div>
      )}

      {error && <InlineError message={error} />}
    </motion.div>
  )
}

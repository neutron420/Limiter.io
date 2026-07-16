"use client"

import * as React from "react"
import { motion } from "framer-motion"
import { ShieldCheck, RefreshCw, Check, X } from "lucide-react"

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
import type { ImmutableAuditLog } from "@/lib/types"

export default function AuditLogsPage() {
  const [logs, setLogs] = React.useState<ImmutableAuditLog[]>([])
  const [loading, setLoading] = React.useState(true)
  const [error, setError] = React.useState("")
  const [verifyResult, setVerifyResult] = React.useState<{ valid: boolean; count: number } | null>(null)

  const fetchLogs = React.useCallback(async () => {
    setLoading(true)
    try {
      const data = await api.get<ImmutableAuditLog[]>("/audit-logs")
      setLogs(data)
    } catch {
      setLogs([])
    } finally {
      setLoading(false)
    }
  }, [])

  React.useEffect(() => { fetchLogs() }, [fetchLogs])

  const handleVerify = async () => {
    setVerifyResult(null)
    try {
      const data = await api.get<{ valid: boolean; count: number }>("/audit-logs/verify-chain")
      setVerifyResult(data)
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Verification failed")
    }
  }

  return (
    <motion.div initial={{ opacity: 0, y: 12 }} animate={{ opacity: 1, y: 0 }} className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold font-heading tracking-tight">AUDIT LOGS</h1>
          <p className="text-sm text-muted-foreground mt-1">Immutable tamper‑proof audit trail</p>
        </div>
        <div className="flex gap-2">
          <BrutalButton onClick={handleVerify} variant={"secondary" as any} icon={Check}>VERIFY CHAIN</BrutalButton>
          <BrutalButton onClick={fetchLogs} variant={"secondary" as any} icon={RefreshCw}>REFRESH</BrutalButton>
        </div>
      </div>

      {verifyResult && (
        <Panel>
          <PanelHeader
            icon={verifyResult.valid ? Check : X}
            title={verifyResult.valid ? "Chain Verified ✓" : "Chain Integrity Error ✗"}
          />
          <p className="text-sm font-mono">{verifyResult.count} log entries checked</p>
        </Panel>
      )}

      {loading ? (
        <Spinner label="LOADING AUDIT LOGS" />
      ) : logs.length === 0 ? (
        <EmptyState icon={ShieldCheck} title="No audit logs" hint="Audit events will appear here" />
      ) : (
        <div className="space-y-2">
          {logs.map((log) => (
            <Panel key={log.id}>
              <div className="grid grid-cols-[auto_1fr_auto] gap-4 items-start">
                <span className="inline-block border border-foreground/30 bg-muted/10 text-muted-foreground px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider">{log.action.replace(/[._]/g, " ").toUpperCase()}</span>
                <div className="space-y-1 min-w-0">
                  <p className="text-sm font-mono font-bold truncate">{log.resource}</p>
                  <p className="text-xs text-muted-foreground truncate">{log.details}</p>
                  <p className="text-xs text-muted-foreground font-mono">
                    IP: {log.ip_address} · {new Date(log.created_at).toLocaleString()}
                  </p>
                </div>
                <p className="text-[10px] font-mono text-muted-foreground break-all max-w-30 text-right">
                  {log.checksum.slice(0, 16)}…
                </p>
              </div>
            </Panel>
          ))}
        </div>
      )}

      {error && <InlineError message={error} />}
    </motion.div>
  )
}
